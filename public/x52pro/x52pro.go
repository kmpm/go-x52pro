package x52pro

import (
	"errors"
	"sync"

	"github.com/kmpm/go-x52pro/internal/do"
)

type X52Pro struct {
	pages        map[string]*Page
	device       *do.DirectOutputDevice
	page_counter int
	mu           sync.Mutex
}

func New() (*X52Pro, error) {
	device, err := do.NewDevice()
	if err != nil {
		return nil, err
	}

	x := &X52Pro{
		pages:  make(map[string]*Page),
		device: device,
	}

	return x, nil
}

func (x *X52Pro) Close() {
	x.device.Close()
	x.device = nil
}

// AddPage adds or overwrites a new page to the X52Pro device.
func (x *X52Pro) AddPage(name string, active bool) *Page {
	x.mu.Lock()
	defer x.mu.Unlock()
	x.pages[name] = newPage(x.device, x.page_counter, name, active)
	x.page_counter++
	return x.pages[name]
}

func (x *X52Pro) RemovePage(name string) error {
	if _, ok := x.pages[name]; !ok {
		return errors.New("page does not exist")
	}
	delete(x.pages, name)
	return nil
}
