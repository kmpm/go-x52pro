package x52pro

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/kmpm/go-x52pro/internal/do"
)

type X52Pro struct {
	pages        map[string]*Page
	device       *do.DirectOutputDevice
	page_counter int
	mu           sync.Mutex
	log          *slog.Logger
}

func New() (*X52Pro, error) {
	device, err := do.NewDevice()
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
	// x.log.Info("onPageChange", "page", page, "activated", activated)
	active := make([]string, len(x.pages))
	i := 0
	for _, p := range x.pages {
		if p.id == page {
			p.SetActivation(activated)
			// break
		}
		if p.active {
			active[i] = p.name
		} else {
			active[i] = "-"
		}
		i++
	}
	x.log.Info("post onPageChange", "active", active)
}

// AddPage adds or overwrites a new page to the X52Pro device.
func (x *X52Pro) AddPage(name string, setActive bool) *Page {
	x.mu.Lock()
	defer x.mu.Unlock()
	x.log.Info("AddPage", "name", name, "setActive", setActive)
	x.page_counter++
	p := newPage(x.device, x.page_counter, name, setActive)
	x.pages[name] = p
	//p.Refresh()
	return p
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
