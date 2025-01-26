package x52pro

import (
	"errors"

	"github.com/kmpm/go-x52pro/internal/do"
)

type Page struct {
	id     int
	name   string
	lines  [3]string
	device *do.DirectOutputDevice
	active bool
}

func newPage(d *do.DirectOutputDevice, id int, name string, active bool) *Page {
	p := &Page{
		device: d,
		id:     id,
		name:   name,
		active: active,
	}
	p.device.AddPage(id, name, active)
	return p
}

func (p *Page) SetLine(line int, text string) error {
	if line < 0 || line > 2 {
		return errors.New("line out of range")
	}
	p.lines[line] = text
	return p.device.SetString(p.id, line, text)
}

func (p *Page) GetLine(line int) string {
	if line < 0 || line > 2 {
		return ""
	}
	return p.lines[line]
}

func (p *Page) Refresh() {
	if p.active {
		p.device.AddPage(p.id, p.name, true)
		for i, line := range p.lines {
			p.device.SetString(p.id, i, line)
		}
	}
}
