package xclient

import (
	"fmt"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/frame"
	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/wm"
	"github.com/BurntSushi/wingo/workspace"
)

const (
	clientTypeNormal = iota
	clientTypeDesktop
	clientTypeDock
)

type Client struct {
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
	transientFor *Client
	time         xproto.Timestamp

	// unmapIgnore is the number of UnmapNotify events to ignore.
	// When 0, an UnmapNotify event causes a client to be unmanaged.
	unmapIgnore int

	// floating, when true, this client will *always* be in the floating layer.
	floating bool
}

func (c *Client) IsMapped() bool {
	return c.frame.IsMapped()
}

func (c *Client) Map() {
	if c.IsMapped() {
		return
	}
	c.win.Map()
	c.frame.Map()
	icccm.WmStateSet(wm.X, c.Id(), &icccm.WmState{State: icccm.StateNormal})
}

func (c *Client) Unmap() {
	if !c.IsMapped() {
		return
	}
	c.unmapIgnore++
	c.frame.Unmap()
	c.win.Unmap()
	icccm.WmStateSet(wm.X, c.Id(), &icccm.WmState{State: icccm.StateIconic})
}

func (c *Client) Close() {
	if strIndex("WM_DELETE_WINDOW", c.protocols) > -1 {
		wm_protocols, err := xprop.Atm(wm.X, "WM_PROTOCOLS")
		if err != nil {
			logger.Warning.Println(err)
			return
		}

		wm_del_win, err := xprop.Atm(wm.X, "WM_DELETE_WINDOW")
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

		err = xproto.SendEventChecked(wm.X.Conn(), false, c.Id(), 0,
			string(cm.Bytes())).Check()
		if err != nil {
			logger.Message.Printf("Could not send WM_DELETE_WINDOW "+
				"ClientMessage because: %s", err)
		}
	} else {
		c.win.Kill() // HULK SMASH!
	}
}

func (c *Client) refreshName() {
	defer func() {
		c.frames.full.UpdateTitle()
		c.prompts.updateName()
	}()

	c.name, _ = ewmh.WmVisibleNameGet(wm.X, c.Id())
	if len(c.name) > 0 {
		return
	}

	c.name, _ = ewmh.WmNameGet(wm.X, c.Id())
	if len(c.name) > 0 {
		return
	}

	c.name, _ = icccm.WmNameGet(wm.X, c.Id())
	if len(c.name) > 0 {
		return
	}

	c.name = "Unnamed Window"
}

func (c *Client) hasType(atom string) bool {
	return strIndex(atom, c.winTypes) > -1
}

func (c *Client) String() string {
	// return c.name 
	return fmt.Sprintf("%d :: %s", c.Id(), c.name)
}

func (c *Client) Id() xproto.Window {
	return c.win.Id
}

func (c *Client) Win() *xwindow.Window {
	return c.win
}

func (c *Client) TopWin() *xwindow.Window {
	return c.frame.Parent().Window
}

func (c *Client) Layer() int {
	return c.layer
}

func (c *Client) Maximized() bool {
	return c.maximized
}

func (c *Client) Name() string {
	return c.String()
}

func (c *Client) State() int {
	return c.state
}
