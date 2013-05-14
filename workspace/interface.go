package workspace

import (
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo-conc/layout"
)

type Workspacer interface {
	String() string
	LayoutName() string
	Geom() xrect.Rect
	HeadGeom() xrect.Rect
	IsActive() bool
	IsVisible() bool
	Add(c Client)
	Remove(c Client)
	IconifyToggle(c Client)
	Layout(c Client) layout.Layout
}

type Sticky struct {
	X       *xgbutil.XUtil
	floater layout.Floater
}

func (wrks *Workspaces) NewSticky() *Sticky {
	return &Sticky{wrks.X, layout.NewFloating()}
}

func (wrk *Sticky) String() string {
	return "Sticky"
}

func (wrk *Sticky) LayoutName() string {
	return wrk.floater.Name()
}

func (wrk *Sticky) Geom() xrect.Rect {
	return xwindow.RootGeometry(wrk.X)
}

func (wrk *Sticky) HeadGeom() xrect.Rect {
	return xwindow.RootGeometry(wrk.X)
}

func (wrk *Sticky) IsActive() bool {
	return true
}

func (wrk *Sticky) IsVisible() bool {
	return true
}

func (wrk *Sticky) Add(c Client) {}

func (wrk *Sticky) Remove(c Client) {}

func (wrk *Sticky) IconifyToggle(c Client) {
	if c.Iconified() {
		c.LoadState("before-iconify")
		c.IconifiedSet(false)
		c.Map()
	} else {
		c.SaveState("before-iconify")
		c.IconifiedSet(true)
		c.Unmap()
	}
}

func (wrk *Sticky) Layout(c Client) layout.Layout {
	return wrk.floater
}
