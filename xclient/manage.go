package xclient

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/focus"
	"github.com/BurntSushi/wingo/frame"
	"github.com/BurntSushi/wingo/heads"
	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/stack"
	"github.com/BurntSushi/wingo/wm"
)

func New(id xproto.Window) *Client {
	wm.X.Grab()
	defer wm.X.Ungrab()

	if client := wm.FindManagedClient(id); client != nil {
		logger.Message.Printf("Already managing client: %s", client)
		return nil
	}

	win := xwindow.New(wm.X, id)
	if _, err := win.Geometry(); err != nil {
		logger.Warning.Printf("Could not manage client %d because: %s", id, err)
		return nil
	}

	c := &Client{
		win:         win,
		name:        "N/A",
		state:       frame.Inactive,
		layer:       stack.LayerDefault,
		maximized:   false,
		iconified:   false,
		unmapIgnore: 0,
		floating:    false,
	}

	c.manage()
	if !c.iconified {
		c.Map()
		if c.primaryType == clientTypeNormal {
			focus.Focus(c)
		}
	}

	return c
}

func (c *Client) manage() {
	c.refreshName()
	logger.Message.Printf("Managing new client: %s", c)

	c.fetchXProperties()
	c.setPrimaryType()
	c.setInitialLayer()

	// Determine whether the client should start iconified or not.
	c.iconified = c.nhints.Flags&icccm.HintState > 0 &&
		c.hints.InitialState == icccm.StateIconic

	// newClientFrames sets c.frame.
	c.frames = c.newClientFrames()
	c.states = c.newClientStates()
	c.prompts = c.newClientPrompts()

	// Before adding the client into our data structures, we should first
	// make sure it's located on the right head. We do this by finding where
	// it *is* place and convert it into the coordinate space of where it
	// *should* be placed.
	oughtHeadGeom := wm.Workspace().Geom()
	cgeom := c.frame.Geom()
	if wrk := wm.Heads.FindMostOverlap(cgeom); wrk != nil {
		isHeadGeom := wrk.Geom()
		ngeom := heads.Convert(cgeom, isHeadGeom, oughtHeadGeom)
		c.MoveResize(true, ngeom.X(), ngeom.Y(), ngeom.Width(), ngeom.Height())
	} else {
		c.Move(oughtHeadGeom.X(), oughtHeadGeom.Y())
	}

	c.maybeInitPlace()
	wm.AddClient(c)
	focus.InitialAdd(c)
	stack.Raise(c)
	wm.Workspace().Add(c)
	c.attachEventCallbacks()
}

func (c *Client) maybeInitPlace() {
	// Any client that isn't normal doesn't get placed.
	// Let it do what it do, baby.
	if c.primaryType != clientTypeNormal {
		return
	}

	// Transients never get placed.
	if c.transientFor != nil {
		return
	}

	// If a user/program position is specified, do not place.
	if c.nhints.Flags&icccm.SizeHintUSPosition > 0 ||
		c.nhints.Flags&icccm.SizeHintPPosition > 0 {

		return
	}

	// We're good, do a placement.
	wm.Workspace().LayoutFloater().InitialPlacement(wm.Workspace().Geom(), c)
}

func (c *Client) fetchXProperties() {
	var err error

	c.hints, err = icccm.WmHintsGet(wm.X, c.Id())
	if err != nil {
		logger.Warning.Println(err)
		logger.Message.Printf("Using reasonable defaults for WM_HINTS for %X",
			c.Id())
		c.hints = &icccm.Hints{
			Flags:        icccm.HintInput | icccm.HintState,
			Input:        1,
			InitialState: icccm.StateNormal,
		}
	}

	c.nhints, err = icccm.WmNormalHintsGet(wm.X, c.Id())
	if err != nil {
		logger.Warning.Println(err)
		logger.Message.Printf("Using reasonable defaults for WM_NORMAL_HINTS "+
			"for %X", c.Id())
		c.nhints = &icccm.NormalHints{}
	}

	c.protocols, err = icccm.WmProtocolsGet(wm.X, c.Id())
	if err != nil {
		logger.Warning.Printf(
			"Window %X does not have WM_PROTOCOLS set.", c.Id())
	}

	c.winTypes, err = ewmh.WmWindowTypeGet(wm.X, c.Id())
	if err != nil {
		logger.Warning.Printf("Could not find window type for window %X, "+
			"using 'normal'.", c.Id())
		c.winTypes = []string{"_NET_WM_WINDOW_TYPE_NORMAL"}
	}

	trans, _ := icccm.WmTransientForGet(wm.X, c.Id())
	if trans == 0 {
		for _, c2_ := range wm.Clients {
			c2 := c2_.(*Client)
			if c2.transient(c) {
				c.transientFor = c2
				break
			}
		}
	} else if transCli := wm.FindManagedClient(trans); transCli != nil {
		c.transientFor = transCli.(*Client)
	}
}

func (c *Client) setPrimaryType() {
	switch {
	case c.hasType("_NET_WM_WINDOW_TYPE_DESKTOP"):
		c.primaryType = clientTypeDesktop
	case c.hasType("_NET_WM_WINDOW_TYPE_DOCK"):
		c.primaryType = clientTypeDock
	default:
		c.primaryType = clientTypeNormal
	}
}

func (c *Client) setInitialLayer() {
	switch c.primaryType {
	case clientTypeDesktop:
		c.layer = stack.LayerDesktop
	case clientTypeDock:
		c.layer = stack.LayerDock
	case clientTypeNormal:
		c.layer = stack.LayerDefault
	default:
		panic("Unimplemented client type.")
	}
}
