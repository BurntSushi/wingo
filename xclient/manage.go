package xclient

import (
	"github.com/BurntSushi/xgb/shape"
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/event"
	"github.com/BurntSushi/wingo/focus"
	"github.com/BurntSushi/wingo/frame"
	"github.com/BurntSushi/wingo/heads"
	"github.com/BurntSushi/wingo/hook"
	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/stack"
	"github.com/BurntSushi/wingo/wm"
	"github.com/BurntSushi/wingo/workspace"
)

func New(id xproto.Window) *Client {
	wm.X.Grab()
	defer wm.X.Ungrab()

	// If this is an override redirect, skip...
	attrs, err := xproto.GetWindowAttributes(wm.X.Conn(), id).Reply()
	if err != nil {
		logger.Warning.Printf("Could not get window attributes for '%d': %s",
			id, err)
	} else {
		if attrs.OverrideRedirect {
			logger.Message.Printf(
				"Not managing override redirect window %d", id)
			return nil
		}
	}

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

	// We don't fire the managed hook on startup since it can lead to
	// unintuitive state changes.
	// If someone really wants it, we can add a new "startup_managed" hook
	// or something.
	if !wm.Startup {
		event.Notify(event.ManagedClient{c.Id()})
		c.FireHook(hook.Managed)
	}
	if !c.iconified {
		c.Map()
		if !wm.Startup && c.PrimaryType() == TypeNormal {
			if !wm.Config.Ffm || wm.Config.FfmStartupFocus {
				c.Focus()
			}
		}
	}

	return c
}

func (c *Client) manage() {
	c.refreshName()
	logger.Message.Printf("Managing new client: %s", c)

	promptDone := make(chan struct{}, 0)
	go func() {
		c.prompts = c.newClientPrompts()
		promptDone <- struct{}{}
	}()

	c.fetchXProperties()
	c.setPrimaryType()
	c.setInitialLayer()

	// Determine whether the client should start iconified or not.
	c.iconified = c.nhints.Flags&icccm.HintState > 0 &&
		c.hints.InitialState == icccm.StateIconic

	// newClientFrames sets c.frame.
	c.frames = c.newClientFrames()
	c.states = c.newClientStates()

	presumedWorkspace := c.findPresumedWorkspace()

	c.moveToProperHead(presumedWorkspace)
	c.maybeInitPlace(presumedWorkspace)
	wm.AddClient(c)
	c.maybeAddToFocusStack()
	c.Raise()
	c.attachEventCallbacks()
	c.maybeApplyStruts()

	if _, ok := presumedWorkspace.(*workspace.Sticky); ok {
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

	<-promptDone
}

func (c *Client) maybeInitPlace(presumedWorkspace workspace.Workspacer) {
	// This is a hack. Before a client gets sucked into some layout, we
	// always want to have some floating state to fall back on to. However,
	// by the time we're "allowed" to save the client's state, it will have
	// already been placed in the hands of some layout---which may or may
	// not be floating. So we inject our own state forcefully here.
	defer func() {
		wrk := presumedWorkspace
		if wrk.IsVisible() {
			c.states["last-floating"] = clientState{
				geom:      xrect.New(xrect.Pieces(c.frame.Geom())),
				headGeom:  xrect.New(xrect.Pieces(wrk.HeadGeom())),
				frame:     c.frame,
				maximized: c.maximized,
			}
		} else if wm.Startup {
			// This is a bit tricky. If the window manager is starting up and
			// has to manage existing clients, then we need to find which
			// head the client is on and save its state. This is so future
			// workspace switches will be able to place the client
			// appropriately.
			// (This is most common on a Wingo restart.)
			// We refer to detected workspace as "fake" because the client
			// isn't on a visible workspace (see above), and therefore the
			// visible workspace returned by FindMostOverlap *cannot* contain
			// this client. Therefore, we're only using the fake workspace
			// to get the geometry.
			// (This would make more sense if FindMostOverlap returned a head
			// geometry, but it turns out that a workspace geometry is more
			// useful.)
			cgeom := c.frame.Geom()
			if fakeWrk := wm.Heads.FindMostOverlap(cgeom); fakeWrk != nil {
				c.states["last-floating"] = clientState{
					geom:      xrect.New(xrect.Pieces(c.frame.Geom())),
					headGeom:  xrect.New(xrect.Pieces(fakeWrk.HeadGeom())),
					frame:     c.frame,
					maximized: c.maximized,
				}
			}
		}
	}()

	// Any client that isn't normal doesn't get placed.
	// Let it do what it do, baby.
	if c.PrimaryType() != TypeNormal {
		return
	}

	// If it's sticky, let it do what it do.
	if _, ok := presumedWorkspace.(*workspace.Sticky); ok {
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
	if !presumedWorkspace.IsVisible() || !c.isAttrsUnmapped() {
		return
	}
	w := presumedWorkspace.(*workspace.Workspace)
	w.LayoutFloater().InitialPlacement(c)
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

	c.winStates, err = ewmh.WmStateGet(wm.X, c.Id())
	if err != nil {
		c.winStates = []string{}
		ewmh.WmStateSet(wm.X, c.Id(), c.winStates)
	}

	c.class, err = icccm.WmClassGet(wm.X, c.Id())
	if err != nil {
		logger.Warning.Printf("Could not find window class for window %X: %s",
			c.Id(), err)
		c.class = &icccm.WmClass{
			Instance: "",
			Class:    "",
		}
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

	tmp, err := xprop.PropValNum(
		xprop.GetProperty(wm.X, c.Id(), "_GTK_HIDE_TITLEBAR_WHEN_MAXIMIZED"))
	if err == nil {
		c.gtkMaximizeNada = (tmp == 1)
	} else {
		c.gtkMaximizeNada = false
	}

	c.setShaped()
}

func (c *Client) setPrimaryType() {
	switch {
	case c.hasType("_NET_WM_WINDOW_TYPE_DESKTOP"):
		c.primaryType = TypeDesktop
	case c.hasType("_NET_WM_WINDOW_TYPE_DOCK"):
		c.primaryType = TypeDock
	default:
		c.primaryType = TypeNormal
	}
}

func (c *Client) PrimaryType() int {
	return c.primaryType
}

func (c *Client) PrimaryTypeString() string {
	switch c.PrimaryType() {
	case TypeNormal:
		return "normal"
	case TypeDesktop:
		return "desktop"
	case TypeDock:
		return "dock"
	}
	logger.Error.Fatalf("BUG: Unknown client type %d", c.PrimaryType())
	panic("unreachable")
}

func (c *Client) maybeAddToFocusStack() {
	if c.PrimaryType() == TypeDesktop ||
		c.PrimaryType() == TypeDock {

		return
	}
	focus.InitialAdd(c)
}

func (c *Client) setInitialLayer() {
	switch c.PrimaryType() {
	case TypeDesktop:
		c.layer = stack.LayerDesktop
	case TypeDock:
		c.layer = stack.LayerDock
	case TypeNormal:
		c.layer = stack.LayerDefault
	default:
		panic("Unimplemented client type.")
	}
}

func (c *Client) updateInitStates() {
	// Keep a copy of the states since we change things as we go along.
	copied := make([]string, len(c.winStates))
	copy(copied, c.winStates)

	// Handle the weird maximize cases first.
	if strIndex("_NET_WM_STATE_MAXIMIZED_VERT", copied) > -1 &&
		strIndex("_NET_WM_STATE_MAXIMIZED_HORZ", copied) > -1 {

		c.updateState("add", "_NET_WM_STATE_MAXIMIZED")
	}

	for _, state := range copied {
		if state == "_NET_WM_STATE_MAXIMIZED_VERT" ||
			state == "_NET_WM_STATE_MAXIMIZED_HORZ" {

			continue
		}
		c.updateState("add", state)
	}
}

// findPresumedWorkspace inspects a client before it is fully managed to
// see which workspace it should go to. Basically, if _NET_WM_DESKTOP is
// to a valid workspace number, then we grant the request. Otherwise, we use
// the current workspace.
func (c *Client) findPresumedWorkspace() workspace.Workspacer {
	d, err := ewmh.WmDesktopGet(wm.X, c.Id())
	if err != nil {
		return wm.Workspace()
	}
	if d == 0xFFFFFFFF {
		return wm.StickyWrk
	}
	if d < 0 || d >= uint(len(wm.Heads.Workspaces.Wrks)) {
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
func (c *Client) moveToProperHead(presumedWorkspace workspace.Workspacer) {
	if c.PrimaryType() != TypeNormal {
		return
	}
	if _, ok := presumedWorkspace.(*workspace.Sticky); ok {
		return
	}
	if !presumedWorkspace.IsVisible() {
		return
	}

	oughtHeadGeom := presumedWorkspace.HeadGeom()
	cgeom := c.frame.Geom()
	if wrk := wm.Heads.FindMostOverlap(cgeom); wrk != nil {
		if wrk != presumedWorkspace {
			isHeadGeom := wrk.HeadGeom()
			ngeom := heads.Convert(cgeom, isHeadGeom, oughtHeadGeom)
			c.MoveResizeValid(
				ngeom.X(), ngeom.Y(), ngeom.Width(), ngeom.Height())
		}
	} else {
		// If we're here, that means the client *ought* to belong to a visible
		// workspace but it could not be found to overlap with *any* visible
		// workspace. Therefore, just use a hammer and move it to the root
		// coordinates of the presumed workspace.
		geom := presumedWorkspace.Geom()
		c.Move(geom.X(), geom.Y())
	}
}

func (c *Client) setShaped() {
	c.shaped = false
	if wm.ShapeExt {
		err := shape.SelectInputChecked(wm.X.Conn(), c.Id(), true).Check()
		if err != nil {
			logger.Warning.Printf("Could not select Shape events for '%s': %s",
				c, err)
			return
		}

		extents, err := shape.QueryExtents(wm.X.Conn(), c.Id()).Reply()
		if err != nil {
			logger.Warning.Printf("X Shape QueryExtents failed on '%s': %s",
				c, err)
			return
		}
		c.shaped = extents.BoundingShaped
	}
}
