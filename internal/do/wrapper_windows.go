package do

import (
	"errors"
	"fmt"
	"log/slog"
	"unsafe"

	"golang.org/x/sys/windows"
)

// This package will borrow a lot from https://github.com/eyeonus/Trade-Dangerous/blob/release/v1/tradedangerous/mfd/saitek/directoutput.py

// type (
//
//	BOOL          uint32
//	BOOLEAN       byte
//	BYTE          byte
//	DWORD         uint32
//	DWORD64       uint64
//	HANDLE        uintptr
//	HLOCAL        uintptr
//	LARGE_INTEGER int64
//	LONG          int32
//	LPVOID        uintptr
//	SIZE_T        uintptr
//	UINT          uint32
//	ULONG_PTR     uintptr
//	ULONGLONG     uint64
//	WORD          uint16
//	WCHAR_T       uint16 // wchar_t, A 16-bit Unicode character
//
// )
const (
	s_ok             uintptr = 0
	e_handle         uintptr = 0x80070006
	e_notimpl        uintptr = 0x80004001
	e_invalidarg     uintptr = 0x80070057
	e_pagenotacticve uintptr = 0xFF040001
)

var (
	directOutputDll = windows.NewLazyDLL("DirectOutput.dll")

	procDirectOutput_Initialize                = directOutputDll.NewProc("DirectOutput_Initialize")
	procDirectOutput_Deinitialize              = directOutputDll.NewProc("DirectOutput_Deinitialize")
	procDirectOutput_Enumerate                 = directOutputDll.NewProc("DirectOutput_Enumerate")
	proceDirectOutput_RegisterDeviceCallback   = directOutputDll.NewProc("DirectOutput_RegisterDeviceCallback")
	procDirectOutput_RegisterPageCallback      = directOutputDll.NewProc("DirectOutput_RegisterPageCallback")
	procDirectOuput_RegisterSoftButtonCallback = directOutputDll.NewProc("DirectOutput_RegisterSoftButtonCallback")

	procDirectOutput_AddPage   = directOutputDll.NewProc("DirectOutput_AddPage")
	procDirectOutput_SetString = directOutputDll.NewProc("DirectOutput_SetString")
)

func wchar_t(s string) *uint16 {
	ptr, err := windows.UTF16PtrFromString(s)
	if err != nil {
		panic(err)
	}
	return ptr
}

func failed(r uintptr) bool {
	return r != s_ok
}

func asError(r uintptr) error {
	switch r {
	case s_ok:
		return nil
	case e_handle:
		return errors.New("is not a valid handle")
	case e_notimpl:
		return errors.New("not implemented")
	case e_invalidarg:
		return errors.New("invalid argument")
	case e_pagenotacticve:
		return errors.New("page not active")
	default:
		return fmt.Errorf("unknown error: 0x%08X", r)
	}
}

type DirectOutput struct {
	log *slog.Logger
}

func New() *DirectOutput {

	d := &DirectOutput{
		log: slog.Default().With("module", "DirectOutput"),
	}

	return d
}
func (d *DirectOutput) Initialize(appName string) error {
	r, _, _ := procDirectOutput_Initialize.Call(uintptr(unsafe.Pointer(wchar_t(appName))))
	if failed(r) {
		return asError(r)
	}
	return nil
}

func (d *DirectOutput) Deinitialize() error {
	r, _, _ := procDirectOutput_Deinitialize.Call()
	d.log.Info("Deinitialize", "r", r)
	if failed(r) {
		return asError(r)
	}
	return nil
}

// DeviceChangeHandler typedef void (__stdcall *Pfn_DirectOutput_DeviceChange)(void* hDevice, bool bAdded, void* pCtxt);
type DeviceChangeHandler func(hDevice uintptr, bAdded bool, pCtxt uintptr) uintptr

func (d *DirectOutput) RegisterDeviceCallback(fn DeviceChangeHandler) error {
	callback := func(hDevice uintptr, bAdded uint32, pCtxt uintptr) uintptr {
		fn(hDevice, bAdded != 0, pCtxt)
		return 0
	}
	r, _, _ := proceDirectOutput_RegisterDeviceCallback.Call(windows.NewCallback(callback))
	if failed(r) {
		return asError(r)
	}
	return nil
}

type EnumerateHandler func(hDevice, pCtxt uintptr) uintptr

func (d *DirectOutput) Enumerate(fn EnumerateHandler) error {
	r, _, _ := procDirectOutput_Enumerate.Call(windows.NewCallback(fn), 0)
	if failed(r) {
		return asError(r)
	}
	return nil
}

type PageChangeHandler func(hDevice uintptr, page uint32, bActivated bool, pCtxt uintptr) uintptr

func (d *DirectOutput) RegisterPageCallback(hDevice uintptr, fn PageChangeHandler) error {
	r, _, _ := procDirectOutput_RegisterPageCallback.Call(hDevice, windows.NewCallback(fn), 0)
	if failed(r) {
		return asError(r)
	}
	return nil
}

type SoftButtonHandler func(hDevice uintptr, dwButtons uint32, pCtxt uintptr) uintptr

func (d *DirectOutput) RegisterSoftButtonCallback(hDevice uintptr, fn SoftButtonHandler) error {
	if r, _, _ := procDirectOuput_RegisterSoftButtonCallback.Call(hDevice, windows.NewCallback(fn)); failed(r) {
		return windows.GetLastError()
	}
	return nil
}

func (d *DirectOutput) AddPage(hDevice uintptr, id uint32, name string, SetActive bool) error {
	var flags uint32
	if SetActive {
		flags = 1
	}
	r, _, lastErr := procDirectOutput_AddPage.Call(
		hDevice,
		uintptr(id),
		uintptr(unsafe.Pointer(wchar_t(name))),
		uintptr(flags),
	)
	d.log.Info("AddPage", "r", r, "id", id, "name", name, "SetActive", SetActive, "flags", flags, "lastErr", lastErr)
	if failed(r) {
		return asError(r)
	}

	return nil
}

func (d *DirectOutput) SetString(hDevice uintptr, page uint32, line uint32, text string) error {
	count := uint32(len(text))
	r, _, lastErr := procDirectOutput_SetString.Call(
		hDevice,
		uintptr(unsafe.Pointer(&page)),
		uintptr(unsafe.Pointer(&line)),
		uintptr(unsafe.Pointer(&count)),
		uintptr(unsafe.Pointer(wchar_t(text))),
	)
	d.log.Info("SetString", "r", r, "page", page, "line", line, "text", text, "count", count, "lastErr", lastErr)
	if failed(r) {
		return asError(r)
	}
	return nil
}
