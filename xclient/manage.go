package xclient

import (
	"time"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/focus"
	"github.com/BurntSushi/wingo/frame"
	"github.com/BurntSushi/wingo/heads"
	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/stack"
	"github.com/BurntSushi/wingo/wm"
	"github.com/BurntSushi/wingo/workspace"
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
		fullscreen:  false,
		skipTaskbar: false,
		skipPager:   false,
		demanding:   false,
		attnQuit:    make(chan struct{}, 0),
	}

	c.manage()
	if !c.iconified {
		c.Map()
		if !wm.Startup && c.primaryType == clientTypeNormal {
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

	presumedWorkspace := c.findPresumedWorkspace()

	c.moveToProperHead(presumedWorkspace)
	c.maybeInitPlace(presumedWorkspace)
	wm.AddClient(c)
	c.maybeAddToFocusStack()
	stack.Raise(c)
	c.attachEventCallbacks()
	c.maybeApplyStruts()

	if d, _ := ewmh.WmDesktopGet(wm.X, c.Id()); int64(d) == 0xFFFFFFFF {
		c.stick()
	} else {
		presumedWorkspace.Add(c)
	}

	c.updateInitStates()
	ewmh.WmAllowedActionsSet(wm.X, c.Id(), allowedActions)

	err := xproto.ChangeSaveSetChecked(
		wm.X.Conn(), xproto.SetModeInsert, c.Id()).Check()
	if err != nil {
		logger.Warning.Printf(
			"Could not add client '%s' to SaveSet. This may be problematic "+
				"if you try to replace Wingo with another window manager: %s",
			c, err)
	}
}

func (c *Client) fullscreenToggle() {
	if c.fullscreen {
		c.fullscreened()
	} else {
		c.unfullscreened()
	}
}

func (c *Client) fullscreened() {
	if c.workspace == nil || !c.workspace.IsVisible() {
		return
	}
	if c.fullscreen {
		return
	}
	c.SaveState("before-fullscreen")
	c.fullscreen = true

	// Make sure the window has been forced into a floating layout.
	if wrk, ok := c.Workspace().(*workspace.Workspace); ok {
		wrk.CheckFloatingStatus(c)
	}

	// Resize outside of the constraints of a layout.
	g := c.Workspace().HeadGeom()
	c.FrameNada()
	c.MoveResize(false, g.X(), g.Y(), g.Width(), g.Height())

	// Since we moved outside of the layout, we have to save the last
	// floating state our selves.
	c.SaveState("last-floating")

	c.addState("_NET_WM_STATE_FULLSCREEN")
}

func (c *Client) unfullscreened() {
	if !c.fullscreen {
		return
	}
	c.fullscreen = false
	c.LoadState("before-fullscreen")

	c.removeState("_NET_WM_STATE_FULLSCREEN")
}

func (c *Client) IsSticky() bool {
	return c.sticky
}

func (c *Client) StickyToggle() {
	if c.sticky {
		c.unstick()
	} else {
		c.stick()
	}
}

func (c *Client) unstick() {
	c.sticky = false
	c.workspace = nil
	wm.Workspace().Add(c)

	c.removeState("_NET_WM_STATE_STICKY")
}

func (c *Client) stick() {
	if c.sticky {
		return
	}

	c.sticky = true
	if c.workspace != nil {
		c.workspace.(*workspace.Workspace).CheckFloatingStatus(c)
		c.workspace.Remove(c)
	}
	c.WorkspaceSet(wm.StickyWrk)

	c.addState("_NET_WM_STATE_STICKY")
}

func (c *Client) maybeInitPlace(presumedWorkspace *workspace.Workspace) {
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

	// We're good, do a placement unless we're already mapped or on a
	// hidden workspace..
	if presumedWorkspace.IsVisible() {
		if c.isAttrsUnmapped() {
			layFloater := presumedWorkspace.LayoutFloater()
			layFloater.InitialPlacement(presumedWorkspace.Geom(), c)
		}

		// This is a hack. Before a client gets sucked into some layout, we
		// always want to have some floating state to fall back on to. However,
		// by the time we're "allowed" to save the client's state, it will have
		// already been placed in the hands of some layout---which may or may
		// not be floating. So we inject our own state forcefully here.
		c.states["last-floating"] = clientState{
			geom:      xrect.New(xrect.Pieces(c.frame.Geom())),
			headGeom:  xrect.New(xrect.Pieces(presumedWorkspace.Geom())),
			frame:     c.frame,
			maximized: c.maximized,
		}
	}
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

func (c *Client) maybeAddToFocusStack() {
	if c.primaryType == clientTypeDesktop ||
		c.primaryType == clientTypeDock {

		return
	}
	focus.InitialAdd(c)
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

func (c *Client) updateInitStates() {
	var err error

	c.winStates, err = ewmh.WmStateGet(wm.X, c.Id())
	if err != nil {
		c.winStates = []string{}
		return
	}

	// Handle the weird maximize cases first.
	if strIndex("_NET_WM_STATE_MAXIMIZED_VERT", c.winStates) > -1 &&
		strIndex("_NET_WM_STATE_MAXIMIZED_HORZ", c.winStates) > -1 {

		c.updateState("add", "_NET_WM_STATE_MAXIMIZED")
	}
	for _, state := range c.winStates {
		if state == "_NET_WM_STATE_MAXIMIZED_VERT" ||
			state == "_NET_WM_STATE_MAXIMIZED_HORZ" {

			continue
		}
		c.updateState("add", state)
	}
}

func (c *Client) attnStart() {
	if c.demanding {
		return
	}

	c.demanding = true
	go func() {
		for {
			select {
			case <-time.After(500 * time.Millisecond):
				if c.State() == frame.Active {
					c.frame.Inactive()
					c.state = frame.Inactive
				} else {
					c.frame.Active()
					c.state = frame.Active
				}
			case <-c.attnQuit:
				return
			}
		}
	}()

	c.addState("_NET_WM_STATE_DEMANDS_ATTENTION")
}

func (c *Client) attnStop() {
	if !c.demanding {
		return
	}

	c.attnQuit <- struct{}{}
	c.demanding = false
	c.frame.Inactive()

	c.removeState("_NET_WM_STATE_DEMANDS_ATTENTION")
}

func (c *Client) isAttrsUnmapped() bool {
	attrs, err := xproto.GetWindowAttributes(wm.X.Conn(), c.Id()).Reply()
	if err != nil {
		logger.Warning.Printf(
			"Could not get window attributes for '%s': %s.", c, err)
	}
	return attrs.MapState == xproto.MapStateUnmapped
}

// findPresumedWorkspace inspects a client before it is fully managed to
// see which workspace it should go to. Basically, if _NET_WM_DESKTOP is
// to a valid workspace number, then we grant the request. Otherwise, we use
// the current workspace.
func (c *Client) findPresumedWorkspace() *workspace.Workspace {
	d, err := ewmh.WmDesktopGet(wm.X, c.Id())
	if err != nil || int64(d) == 0xFFFFFFFF {
		return wm.Workspace()
	}
	if d < 0 || d >= int64(len(wm.Heads.Workspaces.Wrks)) {
		return wm.Workspace()
	}
	return wm.Heads.Workspaces.Get(int(d))
}

// moveToProperHead is used to make sure a newly managed client is placed on
// the correct monitor.
//
// Before adding the client into our data structures, we should first
// make sure it's located on the right head. We do this by finding where
// it *is* placed and convert it into the coordinate space of where it
// *should* be placed.
//
// Note that presumedWorkspace MUST be visible.
func (c *Client) moveToProperHead(presumedWorkspace *workspace.Workspace) {
	if !presumedWorkspace.IsVisible() {
		return
	}

	oughtHeadGeom := presumedWorkspace.Geom()
	cgeom := c.frame.Geom()
	if wrk := wm.Heads.FindMostOverlap(cgeom); wrk != nil {
		if wrk != presumedWorkspace {
			isHeadGeom := wrk.Geom()
			ngeom := heads.Convert(cgeom, isHeadGeom, oughtHeadGeom)
			c.MoveResize(true,
				ngeom.X(), ngeom.Y(), ngeom.Width(), ngeom.Height())
		}
	} else {
		// If we're here, that means the client *ought* to belong to a visible
		// workspace but it could not be found to overlap with *any* visible
		// workspace. Therefore, just use a hammer and move it to the root
		// coordinates of the presumed workspace.
		c.Move(oughtHeadGeom.X(), oughtHeadGeom.Y())
	}
}
