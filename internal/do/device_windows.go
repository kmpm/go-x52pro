package do

import (
	"fmt"
	"log/slog"
)

type DirectOutputDevice struct {
	deviceHandle uintptr
	wrapper      *DirectOutput
	log          *slog.Logger
}

func NewDevice() (dev *DirectOutputDevice, err error) {
	wrapper := New()
	log := slog.Default().WithGroup("DirectOutputDevice")
	defer func() {
		if err != nil {
			log.Warn("Failed to initialize DirectOutputDevice", "error", err)
			wrapper.Deinitialize()
		}
	}()

	if err = wrapper.Initialize("go-x52pro"); err != nil {
		return nil, fmt.Errorf("failed to initialize DirectOutput: %w", err)
	}

	dev = &DirectOutputDevice{
		wrapper: wrapper,
		log:     log,
	}

	if err = wrapper.RegisterDeviceCallback(dev.onDeviceChange); err != nil {
		return nil, fmt.Errorf("failed to register device callback: %w", err)
	}

	if err = wrapper.Enumerate(dev.onEnumerate); err != nil {
		return nil, fmt.Errorf("failed to enumerate devices: %w", err)
	}
	return
}

func (d *DirectOutputDevice) Close() {
	d.wrapper.Deinitialize()
}

func (d *DirectOutputDevice) onDeviceChange(hDevice uintptr, bAdded bool, pCtxt uintptr) uintptr {
	d.log.Info("Device change", "hDevice", hDevice, "bAdded", bAdded, "pCtxt", pCtxt)
	if !bAdded {
		d.log.Error("device removal not supported", "hDevice", hDevice)
		return e_notimpl
	}
	if d.deviceHandle != 0 && d.deviceHandle != hDevice {
		d.log.Error("multiple devices not supported", "hDevice", hDevice)
		return e_handle
	}
	d.deviceHandle = hDevice

	if err := d.wrapper.RegisterPageCallback(hDevice, d.onPageChange); err != nil {
		// return nil, fmt.Errorf("failed to register page callback: %w", err)
		d.log.Error("failed to register page callback", "error", err)
		panic(err)
	}

	if err := d.wrapper.RegisterSoftButtonCallback(hDevice, d.onSoftButtonChange); err != nil {
		// return nil, fmt.Errorf("failed to register soft button callback: %w", err)
		d.log.Error("failed to register soft button callback", "error", err)
		panic(err)
	}
	return s_ok
}

func (d *DirectOutputDevice) onEnumerate(hDevice, pCtxt uintptr) uintptr {
	d.log.Info("Enumerate", "device", hDevice, "context", pCtxt)
	return d.onDeviceChange(hDevice, true, pCtxt)
}

func (d *DirectOutputDevice) onPageChange(hDevice uintptr, dwPage uint32, bActivated bool, pCtxt uintptr) uintptr {
	d.log.Info("Page change", "device", hDevice, "page", dwPage, "activated", bActivated, "context", pCtxt)
	return s_ok
}

func (d *DirectOutputDevice) onSoftButtonChange(hDevice uintptr, dwButtons uint32, pCtxt uintptr) uintptr {
	d.log.Info("Soft button change", "device", hDevice, "buttons", dwButtons, "context", pCtxt)
	return s_ok
}

func (d *DirectOutputDevice) AddPage(id int, name string, setActive bool) error {
	return d.wrapper.AddPage(d.deviceHandle, uint32(id), name, setActive)
}

func (d *DirectOutputDevice) SetString(page int, line int, text string) error {
	return d.wrapper.SetString(d.deviceHandle, uint32(page), uint32(line), text)
}
