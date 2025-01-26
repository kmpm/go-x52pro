// SPDX-FileCopyrightText: 2025 Peter Magnusson <me@kmpm.se>
//
// SPDX-License-Identifier: MPL-2.0

package do

import (
	"fmt"
	"log/slog"

	"github.com/kmpm/go-x52pro/internal/helper"
	"golang.org/x/sys/windows"
)

type DevicePageChangeHandler func(page int, activated bool)

type DirectOutputDevice struct {
	deviceHandle uintptr
	wrapper      *DirectOutput
	log          *slog.Logger
	pageChange   DevicePageChangeHandler
	debug        bool
}

func NewDevice() (dev *DirectOutputDevice, err error) {
	wrapper := New()
	log := slog.Default().With("module", "DirectOutputDevice")
	defer func() {
		if err != nil {
			log.Warn("Failed to initialize DirectOutputDevice", "error", err)
			if err = wrapper.Deinitialize(); err != nil {
				log.Error("Failed to deinitialize DirectOutput", "error", err)
			}
		}
	}()

	if err = wrapper.Initialize("go-x52pro"); err != nil {
		return nil, fmt.Errorf("failed to initialize DirectOutput: %w", err)
	}

	dev = &DirectOutputDevice{
		wrapper: wrapper,
		log:     log,
		debug:   helper.HasDebug("DirectOutputDevice"),
	}

	if err = wrapper.RegisterDeviceCallback(dev.onDeviceChange); err != nil {
		return nil, fmt.Errorf("failed to register device callback: %w", err)
	}

	if err = wrapper.Enumerate(dev.onEnumerate); err != nil {
		return nil, fmt.Errorf("failed to enumerate devices: %w", err)
	}
	return
}

func (d *DirectOutputDevice) Close() error {
	return d.wrapper.Deinitialize()
}

func (d *DirectOutputDevice) SetPageChangeHandler(h DevicePageChangeHandler) {
	d.pageChange = h
}

func (d *DirectOutputDevice) onDeviceChange(hDevice uintptr, bAdded bool, pCtxt uintptr) uintptr {
	if d.debug {
		d.log.Info("Device change", "hDevice", hDevice, "bAdded", bAdded, "pCtxt", pCtxt)
	}
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
	if d.debug {
		d.log.Info("Enumerate", "device", hDevice, "context", pCtxt)
	}

	return d.onDeviceChange(hDevice, true, pCtxt)
}

func (d *DirectOutputDevice) onPageChange(hDevice uintptr, page uint32, bActivated bool, pCtxt uintptr) uintptr {
	if d.debug {
		d.log.Info("onPageChange", "hDevice", hDevice, "page", page, "activated", bActivated, "pCtxt", pCtxt)
	}

	defer func() {
		if d.pageChange != nil {
			d.pageChange(int(page), bActivated)
		}
	}()
	return s_ok
}

func (d *DirectOutputDevice) onSoftButtonChange(hDevice uintptr, dwButtons uint32, pCtxt uintptr) uintptr {
	// d.log.Info("Soft button change", "device", hDevice, "buttons", dwButtons, "context", pCtxt)
	return s_ok
}

func (d *DirectOutputDevice) AddPage(id int, name string, setActive bool) error {
	d.log.Info("AddPage", "id", id, "name", name, "setActive", setActive)
	f := uint32(0)
	if setActive {
		f = f | flagSetAsActive
	}
	return d.wrapper.AddPage(d.deviceHandle, uint32(id), name, f)
}

func (d *DirectOutputDevice) SetString(page int, line int, text string) error {
	return d.wrapper.SetString(d.deviceHandle, uint32(page), uint32(line), text)
}

func (d *DirectOutputDevice) GetDeviceType() (*windows.GUID, error) {
	if t, err := d.wrapper.GetDeviceType(d.deviceHandle); err != nil {
		return nil, err

	} else {
		return &t, nil
	}
}
