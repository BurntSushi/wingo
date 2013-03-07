package layout

import (
	"container/list"
	"fmt"

	"github.com/BurntSushi/xgbutil/xrect"
)

type Maximized struct {
	clients	*list.List
	geom	xrect.Rect
}

func NewMaximized() *Maximized {
	return &Maximized{
		clients:	list.New(),
	}
}

func (m *Maximized) Name() string {
	return "Maximized"
}

func (m *Maximized) SetGeom(geom xrect.Rect) {
	m.geom = geom
}

func (m *Maximized) Place() {
	e := m.clients.Front()
	if e != nil {
		c := e.Value.(Client)
		x, y, w, h := m.geom.X(), m.geom.Y(), m.geom.Width(), m.geom.Height()
		c.FrameTile()
		c.MoveResize(x, y, w, h)
	}
}

func (m *Maximized) Unplace() {
}

func (m *Maximized) Add(c Client) {
	if !m.Exists(c) {
		m.clients.PushFront(c)
	}
}

func (m *Maximized) Remove(c Client) {
}

func (m *Maximized) Destroy() {
}

func (m *Maximized) Exists(c Client) bool {
	for e := m.clients.Front(); e != nil; e = e.Next() {
		if e.Value.(Client) == c {
			return true
		}
	}
	return false
}

func (m *Maximized) ResizeMaster(amount float64) {
}

func (m *Maximized) ResizeWindow(amount float64) {
}

func (m *Maximized) Next() {
	if f := m.clients.Front(); f != nil {
		m.clients.MoveToBack(f)
		c := m.clients.Front().Value.(Client)
		c.Focus()
		c.Raise()
		m.Place()
	}
}

func (m *Maximized) Prev() {
	if b := m.clients.Back(); b != nil {
		m.clients.MoveToFront(b)
		c := m.clients.Front().Value.(Client)
		c.Focus()
		c.Raise()
		m.Place()
	}
}

// This is useful, but can be implemented later
func (m *Maximized) SwitchNext() {
}

// This is useful, but can be implemented later
func (m *Maximized) SwitchPrev() {
}

func (m *Maximized) FocusMaster() {
}

func (m *Maximized) MakeMaster() {
}

func (m *Maximized) MastersMore() {
}

func (m *Maximized) MastersFewer() {
}

func (m *Maximized) MROpt(c Client, flags, x, y, width, height int) {}

func (m *Maximized) MoveResize(c Client, x, y, width, height int) {}

func (m *Maximized) Move(c Client, x, y int) {}

func (m *Maximized) Resize(c Client, width, height int) {}
