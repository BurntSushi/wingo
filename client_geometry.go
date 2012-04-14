package main

import (
	"code.google.com/p/jamslam-x-go-binding/xgb"

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

type clientGeom struct {
	xrect.Rect
	maximized bool
	frame     Frame
}

func (c *client) newClientGeom() *clientGeom {
	return &clientGeom{
		Rect:      xrect.New(xrect.Pieces(c.Frame().Geom())),
		maximized: c.maximized,
		frame:     c.frame,
	}
}

func (c *client) saveGeom(key string) {
	c.geomStore[key] = c.newClientGeom()
}

func (c *client) saveGeomNoClobber(key string) {
	if _, ok := c.geomStore[key]; !ok {
		c.saveGeom(key)
	}
}

func (c *client) loadGeom(key string) {
	if cgeom, ok := c.geomStore[key]; ok {
		c.frameSet(cgeom.frame)
		if cgeom.maximized {
			// We don't save because we don't want to clobber the existing
			// "restore" state the client has saved.
			c.maximizeRaw()
		} else {
			c.Frame().ConfigureFrame(
				DoX|DoY|DoW|DoH,
				cgeom.X(), cgeom.Y(), cgeom.Width(), cgeom.Height(),
				0, 0, false, true)
		}
		delete(c.geomStore, key)
	}
}
