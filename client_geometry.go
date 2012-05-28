package main

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xrect"
)

func (c *client) GravitizeX(x int, gravity int) int {
	// Don't do anything if there's no gravity options set and we're
	// trying to infer gravity.
	// This is equivalent to setting NorthWest gravity
	if gravity > -1 && c.nhints.Flags&icccm.SizeHintPWinGravity == 0 {
		return x
	}

	// Otherwise, we're either inferring gravity (from normal hints), or
	// using some forced notion of gravity (probably from EWMH stuff)
	var g int
	if gravity > -1 {
		g = gravity
	} else {
		g = int(c.nhints.WinGravity)
	}

	f := c.Frame()
	switch {
	case g == xproto.GravityStatic || g == xproto.GravityBitForget:
		x -= f.Left()
	case g == xproto.GravityNorth || g == xproto.GravitySouth ||
		g == xproto.GravityCenter:
		x -= abs(f.Left()-f.Right()) / 2
	case g == xproto.GravityNorthEast || g == xproto.GravityEast ||
		g == xproto.GravitySouthEast:
		x -= f.Left() + f.Right()
	}

	return x
}

func (c *client) GravitizeY(y int, gravity int) int {
	// Don't do anything if there's no gravity options set and we're
	// trying to infer gravity.
	// This is equivalent to setting NorthWest gravity
	if gravity > -1 && c.nhints.Flags&icccm.SizeHintPWinGravity == 0 {
		return y
	}

	// Otherwise, we're either inferring gravity (from normal hints), or
	// using some forced notion of gravity (probably from EWMH stuff)
	var g int
	if gravity > -1 {
		g = gravity
	} else {
		g = int(c.nhints.WinGravity)
	}

	f := c.Frame()
	switch {
	case g == xproto.GravityStatic || g == xproto.GravityBitForget:
		y -= f.Top()
	case g == xproto.GravityEast || g == xproto.GravityWest ||
		g == xproto.GravityCenter:
		y -= abs(f.Top()-f.Bottom()) / 2
	case g == xproto.GravitySouthEast || g == xproto.GravitySouth ||
		g == xproto.GravitySouthWest:
		y -= f.Top() + f.Bottom()
	}

	return y
}

func (c *client) ValidateHeight(height int) int {
	return c.validateSize(height, c.nhints.HeightInc, c.nhints.BaseHeight,
		c.nhints.MinHeight, c.nhints.MaxHeight)
}

func (c *client) ValidateWidth(width int) int {
	return c.validateSize(width, c.nhints.WidthInc, c.nhints.BaseWidth,
		c.nhints.MinWidth, c.nhints.MaxWidth)
}

func (c *client) validateSize(size, inc, base, min, max int) int {
	if size < min && c.nhints.Flags&icccm.SizeHintPMinSize > 0 {
		return min
	}
	if size < 1 {
		return 1
	}
	if size > max && c.nhints.Flags&icccm.SizeHintPMaxSize > 0 {
		return max
	}
	if inc > 1 && c.nhints.Flags&icccm.SizeHintPResizeInc > 0 {
		var whichb int
		if base > 0 {
			whichb = base
		} else {
			whichb = min
		}
		size = whichb +
			(int(round(float64(size-whichb)/float64(inc))) * inc)
	}

	return size
}

func (c *client) MaximizeToggle() {
	// Don't do anything if a max size is specified.
	if c.nhints.Flags&icccm.SizeHintPMaxSize > 0 {
		return
	}

	if c.maximized {
		c.unmaximize()
	} else {
		c.maximize()
	}
}

func (c *client) maximizable() bool {
	return c.workspace.id >= 0 &&
		c.workspace.visible() &&
		c.layout().maximizable()
}

func (c *client) maximize() {
	if !c.maximizable() {
		return
	}
	if !c.maximized {
		c.saveState("unmaximized")
	}

	c.maximizeRaw()
}

func (c *client) unmaximize() {
	if !c.workspace.visible() {
		return
	}

	c.unmaximizeRaw()
	c.loadState("unmaximized")
}

func (c *client) maximizeRaw() {
	if c.workspace.id < 0 || !c.workspace.visible() {
		return
	}

	c.maximized = true
	c.frameNada.Maximize()
	c.frameSlim.Maximize()
	c.frameBorders.Maximize()
	c.frameFull.Maximize()
	frameMaximize(c.Frame())
}

func (c *client) unmaximizeRaw() {
	if !c.workspace.visible() {
		return
	}

	c.maximized = false
	c.frameNada.Unmaximize()
	c.frameSlim.Unmaximize()
	c.frameBorders.Unmaximize()
	c.frameFull.Unmaximize()
}

func (c *client) EnsureUnmax() {
	if c.maximized {
		c.unmaximizeRaw()
	}
}

func (c *client) geomChange(flags, x, y, w, h int) {
	c.EnsureUnmax()
	c.Frame().ConfigureFrame(flags, x, y, w, h, 0, 0, false, true)
}

func (c *client) geomChangeNoValid(flags, x, y, w, h int) {
	c.EnsureUnmax()
	c.Frame().ConfigureFrame(flags, x, y, w, h, 0, 0, true, true)
}

func (c *client) move(x, y int) {
	c.geomChange(DoX|DoY, x, y, 0, 0)
}

func (c *client) moveNoValid(x, y int) {
	c.geomChangeNoValid(DoX|DoY, x, y, 0, 0)
}

func (c *client) resize(w, h int) {
	c.geomChange(DoW|DoH, 0, 0, w, h)
}

func (c *client) resizeNoValid(w, h int) {
	c.geomChangeNoValid(DoW|DoH, 0, 0, w, h)
}

func (c *client) moveresize(x, y, w, h int) {
	c.geomChange(DoX|DoY|DoW|DoH, x, y, w, h)
}

func (c *client) moveresizeNoValid(x, y, w, h int) {
	c.geomChangeNoValid(DoX|DoY|DoW|DoH, x, y, w, h)
}

const (
	clientStateGeom = 1 << iota
	clientStateFrame
	clientStateHead
)

var clientStateAll = clientStateGeom | clientStateFrame | clientStateHead

type clientState struct {
	xrect.Rect
	maximized bool
	frame     Frame
	headGeom  xrect.Rect
}

func (c *client) newClientState() *clientState {
	var headGeom xrect.Rect = nil
	if c.workspace.visible() {
		headGeom = xrect.New(xrect.Pieces(c.workspace.headGeom()))
	}

	return &clientState{
		Rect:      xrect.New(xrect.Pieces(c.Frame().Geom())),
		maximized: c.maximized,
		frame:     c.frame,
		headGeom:  headGeom,
	}
}

func (c *client) saveState(key string) {
	c.stateStore[key] = c.newClientState()
}

func (c *client) saveStateTransients(key string) {
	for _, c2 := range WM.clients {
		if c.transient(c2) && c2.workspace != nil &&
			c2.workspace.id == c.workspace.id {

			c2.saveState(key)
		}
	}
	c.saveState(key)
}

func (c *client) saveStateNoClobber(key string) {
	if _, ok := c.stateStore[key]; !ok {
		c.saveState(key)
	}
}

func (c *client) copyState(src, dest string) {
	c.stateStore[dest] = c.stateStore[src]
}

func (c *client) copyStateTransients(src, dest string) {
	for _, c2 := range WM.clients {
		if c.transient(c2) && c2.workspace != nil &&
			c2.workspace.id == c.workspace.id {

			c2.copyState(src, dest)
		}
	}
	c.copyState(src, dest)
}

func (c *client) loadState(key string) {
	if cgeom, ok := c.stateStore[key]; ok {
		newGeom := cgeom.Rect

		if c.workspace.visible() && cgeom.headGeom != nil &&
			c.workspace.headGeom() != cgeom.headGeom {

			newGeom = WM.headConvert(cgeom, cgeom.headGeom,
				c.workspace.headGeom())
		}

		if cgeom.maximized {
			// We don't save because we don't want to clobber the existing
			// "restore" state the client has saved.
			c.maximizeRaw()
		} else {
			// Only reset this geometry if it isn't finishing a move/resize
			if !c.frame.Moving() && !c.frame.Resizing() {
				c.Frame().ConfigureFrame(
					DoX|DoY|DoW|DoH,
					newGeom.X(), newGeom.Y(), newGeom.Width(), newGeom.Height(),
					0, 0, false, true)
			}
		}

		// This comes last, otherwise we might be inspecting the wrong frame
		// for information (like whether the client is moving/resizing).
		c.frameSet(cgeom.frame)

		delete(c.stateStore, key)
	}
}

func (c *client) loadStateTransients(key string, flags int) {
	for _, c2 := range WM.clients {
		if c.transient(c2) && c2.workspace != nil &&
			c2.workspace.id == c.workspace.id {

			c2.loadState(key)
		}
	}
	c.loadState(key)
}

func (c *client) deleteState(key string) {
	if _, ok := c.stateStore[key]; ok {
		delete(c.stateStore, key)
	}
}

func (c *client) deleteStateTransients(key string) {
	for _, c2 := range WM.clients {
		if c.transient(c2) && c2.workspace != nil &&
			c2.workspace.id == c.workspace.id {

			c2.deleteState(key)
		}
	}
	c.deleteState(key)
}
