// SPDX-FileCopyrightText: 2025 Peter Magnusson <me@kmpm.se>
//
// SPDX-License-Identifier: MPL-2.0
//lint:file-ignore ST1003 keep likenes to the original code

package sdk

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
	flagSetAsActive  uint32  = 0x00000001
	nullContext      uintptr = 0
)

var (
	directOutputDll = windows.NewLazyDLL("DirectOutput.dll")

	procDirectOutput_Initialize                = directOutputDll.NewProc("DirectOutput_Initialize")
	procDirectOutput_Deinitialize              = directOutputDll.NewProc("DirectOutput_Deinitialize")
	procDirectOutput_Enumerate                 = directOutputDll.NewProc("DirectOutput_Enumerate")
	proceDirectOutput_RegisterDeviceCallback   = directOutputDll.NewProc("DirectOutput_RegisterDeviceCallback")
	procDirectOutput_RegisterPageCallback      = directOutputDll.NewProc("DirectOutput_RegisterPageCallback")
	procDirectOuput_RegisterSoftButtonCallback = directOutputDll.NewProc("DirectOutput_RegisterSoftButtonCallback")

	procDirectOutput_AddPage       = directOutputDll.NewProc("DirectOutput_AddPage")
	procDirectOutput_SetString     = directOutputDll.NewProc("DirectOutput_SetString")
	procDirectOutput_GetDeviceType = directOutputDll.NewProc("DirectOutput_GetDeviceType")
)

func wcharT(s string) *uint16 {
	if ptr, err := windows.UTF16PtrFromString(s); err != nil {
		panic(err)
	} else {
		return ptr
	}
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
	r, _, _ := procDirectOutput_Initialize.Call(uintptr(unsafe.Pointer(wcharT(appName))))
	if failed(r) {
		return asError(r)
	}
	return nil
}

func (d *DirectOutput) Deinitialize() error {
	r, _, _ := procDirectOutput_Deinitialize.Call()
	d.log.Debug("Deinitialize", "r", r)
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
	r, _, _ := procDirectOutput_Enumerate.Call(windows.NewCallback(fn), nullContext)
	if failed(r) {
		return asError(r)
	}
	return nil
}

type PageChangeHandler func(hDevice uintptr, page uint32, bActivated bool, pCtxt uintptr) uintptr

func (d *DirectOutput) RegisterPageCallback(hDevice uintptr, fn PageChangeHandler) error {
	r, _, _ := procDirectOutput_RegisterPageCallback.Call(hDevice, windows.NewCallback(fn), nullContext)
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

func (d *DirectOutput) AddPage(hDevice uintptr, id uint32, name string, flags uint32) error {
	ptr, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return err
	}
	r, _, lastErr := procDirectOutput_AddPage.Call(
		hDevice,
		uintptr(id),
		uintptr(unsafe.Pointer(ptr)),
		uintptr(flags),
	)

	d.log.Debug("AddPage", "r", r, "id", id, "name", name, "flags", flags, "lastErr", lastErr)

	if failed(r) {
		return asError(r)
	}

	return nil
}

func (d *DirectOutput) SetString(hDevice uintptr, page uint32, line uint32, text string) error {
	ptr, _ := windows.UTF16PtrFromString(text)
	count := uint32(len(text))
	r, _, lastErr := procDirectOutput_SetString.Call(
		hDevice,
		uintptr(page),
		uintptr(line),
		uintptr(count),
		uintptr(unsafe.Pointer(ptr)),
	)

	if failed(r) {
		d.log.Debug("SetString failed", "r", r, "lastErr", lastErr, "page", page, "line", line, "text", text, "count", count)
		return asError(r)
	}
	return nil
}

func (d *DirectOutput) GetDeviceType(hDevice uintptr) (windows.GUID, error) {

	deviceType := windows.GUID{}

	r, _, lastErr := procDirectOutput_GetDeviceType.Call(
		hDevice,
		uintptr(unsafe.Pointer(&deviceType)),
	)

	d.log.Debug("GetDeviceType", "r", r, "deviceType", deviceType, "lastErr", lastErr)

	if failed(r) {
		return windows.GUID{}, asError(r)
	}
	return deviceType, nil
}
