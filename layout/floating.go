package layout

import (
	"github.com/BurntSushi/xgbutil/xrect"
)

type Floating struct {
	clients []Client
}

func NewFloating() *Floating {
	return &Floating{
		clients: make([]Client, 0),
	}
}

func (f *Floating) Place(geom xrect.Rect) {
	if geom == nil {
		return
	}
	for _, c := range f.clients {
		if _, ok := c.Layout().(*Floating); ok {
			c.LoadState("last-floating")
		}
	}
}

func (f *Floating) Unplace(geom xrect.Rect) {
	if geom == nil {
		return
	}
	for _, c := range f.clients {
		if _, ok := c.Layout().(*Floating); ok {
			c.SaveState("last-floating")
		}
	}
}

func (f *Floating) Exists(c Client) bool {
	for _, client := range f.clients {
		if client == c {
			return true
		}
	}
	return false
}

func (f *Floating) Add(c Client) {
	if !f.Exists(c) {
		f.clients = append(f.clients, c)
	}
}

func (f *Floating) Remove(c Client) {
	for i, client := range f.clients {
		if client == c {
			f.clients = append(f.clients[:i], f.clients[i+1:]...)
		}
	}
}

func (f *Floating) MoveResize(c Client, x, y, width, height int) {
	c.MoveResize(true, x, y, width, height)
}

func (f *Floating) Move(c Client, x, y int) {
	c.Move(x, y)
}

func (f *Floating) Resize(c Client, width, height int) {
	c.Resize(true, width, height)
}
