package main

import (
	"code.google.com/p/jamslam-x-go-binding/xgb"

	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xrect"

	"github.com/BurntSushi/wingo/logger"
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
	case g == xgb.GravityStatic || g == xgb.GravityBitForget:
		x -= f.Left()
	case g == xgb.GravityNorth || g == xgb.GravitySouth ||
		g == xgb.GravityCenter:
		x -= abs(f.Left()-f.Right()) / 2
	case g == xgb.GravityNorthEast || g == xgb.GravityEast ||
		g == xgb.GravitySouthEast:
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
	case g == xgb.GravityStatic || g == xgb.GravityBitForget:
		y -= f.Top()
	case g == xgb.GravityEast || g == xgb.GravityWest ||
		g == xgb.GravityCenter:
		y -= abs(f.Top()-f.Bottom()) / 2
	case g == xgb.GravitySouthEast || g == xgb.GravitySouth ||
		g == xgb.GravitySouthWest:
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
		c.saveGeom("unmaximized")
	}

	c.maximizeRaw()
}

func (c *client) unmaximize() {
	if !c.workspace.visible() {
		return
	}

	c.unmaximizeRaw()
	c.loadGeom("unmaximized")
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

type clientGeom struct {
	xrect.Rect
	maximized bool
	frame     Frame
	headGeom  xrect.Rect
}

func (c *client) newClientGeom() *clientGeom {
	var headGeom xrect.Rect = nil
	if c.workspace.visible() {
		headGeom = xrect.New(xrect.Pieces(c.workspace.headGeom()))
	}

	return &clientGeom{
		Rect:      xrect.New(xrect.Pieces(c.Frame().Geom())),
		maximized: c.maximized,
		frame:     c.frame,
		headGeom:  headGeom,
	}
}

func (c *client) saveGeom(key string) {
	c.geomStore[key] = c.newClientGeom()
}

func (c *client) saveGeomTransients(key string) {
	for _, c2 := range WM.clients {
		if c.transient(c2) && c2.workspace != nil &&
			c2.workspace.id == c.workspace.id {

			c2.saveGeom(key)
		}
	}
	c.saveGeom(key)
}

func (c *client) saveGeomNoClobber(key string) {
	if _, ok := c.geomStore[key]; !ok {
		c.saveGeom(key)
	}
}

func (c *client) copyGeom(src, dest string) {
	c.geomStore[dest] = c.geomStore[src]
}

func (c *client) copyGeomTransients(src, dest string) {
	for _, c2 := range WM.clients {
		if c.transient(c2) && c2.workspace != nil &&
			c2.workspace.id == c.workspace.id {

			c2.copyGeom(src, dest)
		}
	}
	c.copyGeom(src, dest)
}

func (c *client) loadGeom(key string) {
	if cgeom, ok := c.geomStore[key]; ok {
		c.frameSet(cgeom.frame)
		newGeom := cgeom.Rect

		// let's convert head geometry if need be.
		if c.workspace.visible() && cgeom.headGeom != nil &&
			c.workspace.headGeom() != cgeom.headGeom {

			logger.Debug.Println(c, cgeom.headGeom, c.workspace.headGeom())
			newGeom = WM.headConvert(cgeom, cgeom.headGeom,
				c.workspace.headGeom())
		}

		if cgeom.maximized {
			// We don't save because we don't want to clobber the existing
			// "restore" state the client has saved.
			c.maximizeRaw()
		} else {
			c.Frame().ConfigureFrame(
				DoX|DoY|DoW|DoH,
				newGeom.X(), newGeom.Y(), newGeom.Width(), newGeom.Height(),
				0, 0, false, true)
		}
		delete(c.geomStore, key)
	}
}

func (c *client) loadGeomTransients(key string) {
	for _, c2 := range WM.clients {
		if c.transient(c2) && c2.workspace != nil &&
			c2.workspace.id == c.workspace.id {

			c2.loadGeom(key)
		}
	}
	c.loadGeom(key)
}

func (c *client) deleteGeom(key string) {
	if _, ok := c.geomStore[key]; ok {
		delete(c.geomStore, key)
	}
}

func (c *client) deleteGeomTransients(key string) {
	for _, c2 := range WM.clients {
		if c.transient(c2) && c2.workspace != nil &&
			c2.workspace.id == c.workspace.id {

			c2.deleteGeom(key)
		}
	}
	c.deleteGeom(key)
}
