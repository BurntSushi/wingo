package main

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/frame"
	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/workspace"
)

const (
	clientTypeNormal = iota
	clientTypeDesktop
	clientTypeDock
)

type client struct {
	X *xgbutil.XUtil

	win       *xwindow.Window
	frame     frame.Frame
	workspace *workspace.Workspace

	frames  clientFrames
	states  map[string]clientState
	prompts clientPrompts

	name      string
	state     int // One of frame.Active or frame.Inactive.
	layer     int // From constants in stack package.
	maximized bool
	iconified bool

	primaryType  int // one of clientType[...]
	winTypes     []string
	hints        *icccm.Hints
	nhints       *icccm.NormalHints
	protocols    []string
	class        *icccm.WmClass
	transientFor *client
	time         xproto.Timestamp

	// unmapIgnore is the number of UnmapNotify events to ignore.
	// When 0, an UnmapNotify event causes a client to be unmanaged.
	unmapIgnore int

	// floating, when true, this client will *always* be in the floating layer.
	floating bool
}

func (c *client) Map() {
	if c.frame.IsMapped() {
		return
	}
	c.win.Map()
	c.frame.Map()
	icccm.WmStateSet(c.X, c.Id(), &icccm.WmState{State: icccm.StateNormal})
}

func (c *client) Unmap() {
	if !c.frame.IsMapped() {
		return
	}
	c.unmapIgnore++
	c.frame.Unmap()
	c.win.Unmap()
	icccm.WmStateSet(c.X, c.Id(), &icccm.WmState{State: icccm.StateIconic})
}

func (c *client) Close() {
	if strIndex("WM_DELETE_WINDOW", c.protocols) > -1 {
		wm_protocols, err := xprop.Atm(X, "WM_PROTOCOLS")
		if err != nil {
			logger.Warning.Println(err)
			return
		}

		wm_del_win, err := xprop.Atm(X, "WM_DELETE_WINDOW")
		if err != nil {
			logger.Warning.Println(err)
			return
		}

		cm, err := xevent.NewClientMessage(32, c.Id(), wm_protocols,
			int(wm_del_win))
		if err != nil {
			logger.Warning.Println(err)
			return
		}

		err = xproto.SendEventChecked(X.Conn(), false, c.Id(), 0,
			string(cm.Bytes())).Check()
		if err != nil {
			logger.Message.Printf("Could not send WM_DELETE_WINDOW "+
				"ClientMessage because: %s", err)
		}
	} else {
		c.win.Kill() // HULK SMASH!
	}
}

func (c *client) refreshName() {
	defer func() {
		c.frames.full.UpdateTitle()
		c.prompts.updateName()
	}()

	c.name, _ = ewmh.WmVisibleNameGet(c.X, c.Id())
	if len(c.name) > 0 {
		return
	}

	c.name, _ = ewmh.WmNameGet(c.X, c.Id())
	if len(c.name) > 0 {
		return
	}

	c.name, _ = icccm.WmNameGet(c.X, c.Id())
	if len(c.name) > 0 {
		return
	}

	c.name = "Unnamed Window"
}

func (c *client) hasType(atom string) bool {
	return strIndex(atom, c.winTypes) > -1
}

func (c *client) String() string {
	return c.name
}

func (c *client) Id() xproto.Window {
	return c.win.Id
}

func (c *client) Win() *xwindow.Window {
	return c.win
}

func (c *client) TopWin() *xwindow.Window {
	return c.frame.Parent().Window
}

func (c *client) Layer() int {
	return c.layer
}

func (c *client) Maximized() bool {
	return c.maximized
}

func (c *client) Name() string {
	return c.String()
}

func (c *client) State() int {
	return c.state
}
