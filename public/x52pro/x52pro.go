// SPDX-FileCopyrightText: 2025 Peter Magnusson <me@kmpm.se>
//
// SPDX-License-Identifier: MPL-2.0

package x52pro

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/kmpm/go-x52pro/internal/sdk"
)

type X52Pro struct {
	pages       map[string]*Page
	device      *sdk.DirectOutputDevice
	pageCounter int
	mu          sync.Mutex
	log         *slog.Logger
}

func New() (*X52Pro, error) {
	device, err := sdk.NewDevice()
	if err != nil {
		return nil, err
	}

	x := &X52Pro{
		pages:  make(map[string]*Page),
		device: device,
		log:    slog.Default().With("module", "X52Pro"),
	}
	device.SetPageChangeHandler(x.onPageChange)

	return x, nil
}

func (x *X52Pro) Close() {
	x.device.Close()
	x.device = nil
}

func (x *X52Pro) onPageChange(page int, activated bool) {
	x.mu.Lock()
	defer x.mu.Unlock()
	x.log.Debug("onPageChange", "page", page, "activated", activated)
	for _, p := range x.pages {
		if p.id == page {
			p.SetActivation(activated)
			break
		}
	}
}

// AddPage adds or overwrites a new page to the X52Pro device.
func (x *X52Pro) AddPage(name string, setActive bool) (*Page, error) {
	x.mu.Lock()
	defer x.mu.Unlock()
	x.log.Debug("AddPage", "name", name, "setActive", setActive)
	x.pageCounter++
	p, err := newPage(x.device, x.pageCounter, name, setActive)
	if err != nil {
		return nil, err
	}
	x.pages[name] = p
	//TODO: possible refresh here
	return p, nil
}

func (x *X52Pro) RemovePage(name string) error {
	if _, ok := x.pages[name]; !ok {
		return errors.New("page does not exist")
	}
	delete(x.pages, name)
	return nil
}

func (x *X52Pro) Page(name string) (*Page, error) {
	p, ok := x.pages[name]
	if !ok {
		return nil, errors.New("page does not exist")
	}
	return p, nil
}

func (x *X52Pro) SetString(pgName string, line int, text string) error {
	if p, err := x.Page(pgName); err != nil {
		return err
	} else {
		return p.SetLine(line, text)
	}
}

func (x *X52Pro) GetType() string {
	if t, err := x.device.GetDeviceType(); err != nil {
		return "unknown"
	} else {
		return t.String()
	}
}
