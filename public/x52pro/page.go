// SPDX-FileCopyrightText: 2025 Peter Magnusson <me@kmpm.se>
//
// SPDX-License-Identifier: MPL-2.0

package x52pro

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/kmpm/go-x52pro/internal/do"
)

type Page struct {
	id     int
	name   string
	lines  [3]string
	device *do.DirectOutputDevice
	active bool
	log    *slog.Logger
	mu     sync.Mutex
}

func newPage(d *do.DirectOutputDevice, id int, name string, active bool) (*Page, error) {

	p := &Page{
		device: d,
		id:     id,
		name:   name,
		active: active,
		log:    slog.Default().With("module", "Page", slog.Group("Page", "id", id, "name", name)),
	}
	p.log.Info("newPage", "active", active)
	err := p.device.AddPage(id, name, active)
	if err != nil {
		return nil, err
	}
	return p, nil
}
func (p *Page) SetActivation(active bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.active = active
	if active {
		p.Refresh()
	}
}

func (p *Page) SetLine(line int, text string) (err error) {
	if line < 0 || line > 2 {
		return errors.New("line out of range")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	// p.log.Info("SetLine", "line", line, "text", text)
	p.lines[line] = text
	if p.active {
		err = p.device.SetString(p.id, line, text)
	}
	return
}

func (p *Page) GetLine(line int) string {
	if line < 0 || line > 2 {
		return ""
	}
	return p.lines[line]
}

func (p *Page) Refresh() {
	var err error
	// p.log.Info("Refresh", "active", p.active)
	if p.active {
		// err := p.device.AddPage(p.id, p.name, true)
		// if err != nil {
		// 	p.log.Warn("Failed to add page", "error", err)
		// 	return
		// }
		for i, text := range p.lines {
			err = p.device.SetString(p.id, i, text)
			if err != nil {
				p.log.Warn("Failed to set string", "line", i, "error", err)
			}
		}
	}
}
