package main

import (
	"fmt"

	"code.google.com/p/jamslam-x-go-binding/xgb"

	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/logger"
)

type client struct {
	window              *window
	workspace           int
	layer               int
	name, vname, wmname string
	isMapped            bool
	initMap             bool
	state               int
	maximized           bool
	iconified           bool
	initialMap          bool
	lastTime            int
	unmapIgnore         int

	types        []string
	hints        *icccm.Hints
	nhints       *icccm.NormalHints
	protocols    []string
	transientFor xgb.Id

	geomStore   map[string]xrect.Rect
	promptStore map[string]*window

	frame        Frame
	frameNada    *frameNada
	frameSlim    *frameSlim
	frameBorders *frameBorders
	frameFull    *frameFull
}

func newClient(id xgb.Id) *client {
	return &client{
		window:       newWindow(id),
		workspace:    -1,
		layer:        StackDefault,
		name:         "",
		vname:        "",
		wmname:       "",
		isMapped:     false,
		initMap:      false,
		state:        StateInactive,
		maximized:    false,
		iconified:    false,
		initialMap:   false,
		lastTime:     0,
		unmapIgnore:  0,
		hints:        nil,
		nhints:       nil,
		protocols:    nil,
		transientFor: 0,
		geomStore:    make(map[string]xrect.Rect),
		promptStore:  make(map[string]*window),
		frame:        nil,
		frameNada:    nil,
		frameSlim:    nil,
		frameBorders: nil,
		frameFull:    nil,
	}
}

func (c *client) frameSet(f Frame) {
	if f == c.Frame() { // no need to change...
		return
	}
	if c.Frame() != nil {
		c.Frame().Off()
	}
	c.frame = f
	c.Frame().On()
	FrameReset(c.Frame())
}

func (c *client) unmanage() {
	if c.Mapped() {
		c.unmappedFallback()
	}
	c.frame.Destroy()
	c.setWmState(icccm.StateWithdrawn)

	xevent.Detach(X, c.window.id)
	c.promptRemove()
	WM.stackRemove(c)
	WM.focusRemove(c)
	WM.clientRemove(c)

	WM.updateEwmhStacking()
}

func (c *client) Map() {
	if c.Mapped() {
		return
	}
	c.window.map_()
	c.frame.Map()
	c.isMapped = true
	c.initMap = true
	c.setWmState(icccm.StateNormal)
}

func (c *client) Unmap() {
	if !c.Mapped() {
		return
	}
	c.unmapIgnore++
	c.unmapped()
}

func (c *client) UnmapFallback() {
	c.unmapIgnore++
	c.unmappedFallback()
}

func (c *client) unmapped() {
	c.frame.Unmap()
	c.setWmState(icccm.StateIconic)
	c.isMapped = false
}

func (c *client) unmappedFallback() {
	focused := WM.focused()
	c.unmapped()
	if focused != nil && focused.Id() == c.Id() {
		WM.fallback()
	}
}

func (c *client) IconifyToggle() {
	if c.iconified {
		c.Map()
	} else {
		c.UnmapFallback()
	}
	c.iconified = !c.iconified
}

func (c *client) setWmState(state int) {
	if !c.TrulyAlive() {
		return
	}

	err := icccm.WmStateSet(X, c.window.id, &icccm.WmState{State: state})
	if err != nil {
		var stateStr string
		switch state {
		case icccm.StateNormal:
			stateStr = "Normal"
		case icccm.StateIconic:
			stateStr = "Iconic"
		case icccm.StateWithdrawn:
			stateStr = "Withdrawn"
		default:
			stateStr = "Unknown"
		}
		logger.Warning.Printf("Could not set window state to %s on %s "+
			"because: %v", stateStr, c, err)
	}
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

		cm, err := xevent.NewClientMessage(32, c.window.id, wm_protocols,
			int(wm_del_win))
		if err != nil {
			logger.Warning.Println(err)
			return
		}

		X.Conn().SendEvent(false, c.window.id, 0, cm.Bytes())
	} else {
		c.window.kill()
	}
}

// Alive retrieves all X events up until the point of calling that have been
// sent. It then peeks at those events to see if there is an UnmapNotify
// for client c. If there is one, and if the 'unmapIgnore' at 0, then this
// client is marked for deletion and should be considered dead.
// (unmapIgnore is incremented whenever Wingo unmaps a window. When Wingo
// unmaps a window, we *don't* want to delete it, just hide it.)
func (c *client) Alive() bool {
	X.Flush()             // fills up the XGB event queue with ready events
	xevent.Read(X, false) // fills up the xgbutil event queue without blocking

	// we only consider a client marked for deletion when 'ignore' reaches 0
	ignore := c.unmapIgnore
	for _, ev := range X.QueuePeek() {
		wid := c.Win().id
		if unmap, ok := ev.(xgb.UnmapNotifyEvent); ok && unmap.Window == wid {
			if ignore <= 0 {
				return false
			}
			ignore -= 1
		}
	}
	return true
}

// TrulyAlive is useful in scenarios when Alive doesn't help.
// Namely, when we know the window has been unmapped but are not sure
// if it is still an X resource.
func (c *client) TrulyAlive() bool {
	_, err := xwindow.RawGeometry(X, c.window.id)
	if err != nil {
		return false
	}
	return true
}

// ForceWorkspace makes the current workspace this client's workspace.
func (c *client) ForceWorkspace() {
	if WM.WrkActiveInd() != c.workspace {
		WM.WrkSet(c.workspace, false, false)
	}
}

func (c *client) Focus() {
	if c.hints.Flags&icccm.HintInput > 0 && c.hints.Input == 1 {
		c.ForceWorkspace()
		c.window.focus()
		c.Focused()
	} else if strIndex("WM_TAKE_FOCUS", c.protocols) > -1 {
		c.ForceWorkspace()

		wm_protocols, err := xprop.Atm(X, "WM_PROTOCOLS")
		if err != nil {
			logger.Warning.Println(err)
			return
		}

		wm_take_focus, err := xprop.Atm(X, "WM_TAKE_FOCUS")
		if err != nil {
			logger.Warning.Println(err)
			return
		}

		cm, err := xevent.NewClientMessage(32, c.window.id,
			wm_protocols,
			int(wm_take_focus),
			int(X.GetTime()))
		if err != nil {
			logger.Warning.Println(err)
			return
		}

		X.Conn().SendEvent(false, c.window.id, 0, cm.Bytes())

		c.Focused()
	}
}

func (c *client) Focused() {
	WM.focusAdd(c)
	c.state = StateActive
	c.Frame().Active()

	// Forcefully unfocus all other clients
	WM.unfocusExcept(c.Id())
	c.ForceWorkspace()
}

func (c *client) Unfocused() {
	c.state = StateInactive
	c.Frame().Inactive()
}

func (c *client) MaximizeToggle() {
	// Don't do anything if a max size is specified.
	if c.nhints.Flags&icccm.SizeHintPMaxSize > 0 {
		return
	}

	if c.maximized {
		c.maximized = false
		c.frameNada.Unmaximize()
		c.frameSlim.Unmaximize()
		c.frameBorders.Unmaximize()
		c.frameFull.Unmaximize()
		c.LoadGeom("unmaximized")
	} else {
		c.maximized = true
		c.SaveGeom("unmaximized")
		c.frameNada.Maximize()
		c.frameSlim.Maximize()
		c.frameBorders.Maximize()
		c.frameFull.Maximize()
		frameMaximize(c.Frame())
	}
}

func (c *client) EnsureUnmax() {
	if c.maximized {
		c.maximized = false
		c.frameNada.Unmaximize()
		c.frameSlim.Unmaximize()
		c.frameBorders.Unmaximize()
		c.frameFull.Unmaximize()
	}
}

func (c *client) SaveGeom(key string) {
	c.geomStore[key] = xrect.Make(xrect.Pieces(c.Frame().Geom()))
}

func (c *client) LoadGeom(key string) {
	if geom, ok := c.geomStore[key]; ok {
		c.Frame().ConfigureFrame(
			DoX|DoY|DoW|DoH,
			geom.X(), geom.Y(), geom.Width(), geom.Height(),
			0, 0, false, true)
	}
}

func (c *client) Raise() {
	WM.stackRaise(c, false)

	// Also raise its transients...
	toRaise := make([]*client, 0, 2)
	for i := len(WM.stack) - 1; i >= 0; i-- {
		if c.transient(WM.stack[i]) {
			toRaise = append(toRaise, WM.stack[i])
		}
	}

	for _, c2 := range toRaise {
		WM.stackRaise(c2, false)
	}
	WM.stackRefresh(len(toRaise) + 1)
}

func (c *client) updateProperty(ev xevent.PropertyNotifyEvent) {
	name, err := xprop.AtomName(X, ev.Atom)
	if err != nil {
		logger.Warning.Println("Could not get property atom name for", ev.Atom)
		return
	}

	logger.Lots.Printf("Updating property %s with state %v on window %s",
		name, ev.State, c)

	// helper function to log property vals
	showVals := func(o, n interface{}) {
		logger.Lots.Printf("\tOld value: '%s', new value: '%s'", o, n)
	}

	// Start the arduous process of updating properties...
	switch name {
	case "_NET_WM_NAME":
		fallthrough
	case "_NET_WM_VISIBLE_NAME":
		fallthrough
	case "WM_NAME":
		c.updateName()
	case "_NET_WM_ICON":
		c.frameFull.updateIcon()
		c.promptUpdateIcon()
	case "WM_HINTS":
		hints, err := icccm.WmHintsGet(X, c.Id())
		if err == nil {
			c.hints = hints
			c.frameFull.updateIcon()
		}
	case "WM_NORMAL_HINTS":
		nhints, err := icccm.WmNormalHintsGet(X, c.Id())
		if err == nil {
			c.nhints = nhints
		}
	case "_NET_WM_USER_TIME":
		newTime, err := ewmh.WmUserTimeGet(X, c.window.id)
		showVals(c.lastTime, newTime)
		if err == nil {
			c.lastTime = newTime
		}
	}
}

func (c *client) updateName() {
	// helper function to log property vals
	showVals := func(o, n interface{}) {
		logger.Lots.Printf("\tOld value: '%s', new value: '%s'", o, n)
	}

	var name string
	var err error

	name, err = ewmh.WmVisibleNameGet(X, c.Id())
	showVals(c.vname, name)
	if err == nil {
		c.vname = name
	}

	name, err = ewmh.WmNameGet(X, c.Id())
	showVals(c.name, name)
	if err == nil {
		c.name = name
	}

	// Only look for the old style name if we don't have one
	if name == "" {
		name, err = icccm.WmNameGet(X, c.Id())
		showVals(c.name, name)
		if err == nil {
			c.name = name
		}
	}

	c.frameFull.updateTitle()
	c.promptUpdateName()
}

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

func (c *client) Frame() Frame {
	return c.frame
}

func (c *client) FrameNada() {
	c.frameSet(c.frameNada)
}

func (c *client) FrameSlim() {
	c.frameSet(c.frameSlim)
}

func (c *client) FrameBorders() {
	c.frameSet(c.frameBorders)
}

func (c *client) FrameFull() {
	c.frameSet(c.frameFull)
}

func (c *client) Geom() xrect.Rect {
	return c.window.geom
}

func (c *client) Id() xgb.Id {
	return c.window.id
}

func (c *client) Layer() int {
	return c.layer
}

func (c *client) Mapped() bool {
	return c.isMapped
}

func (c *client) Name() string {
	if len(c.vname) > 0 {
		return c.vname
	}
	if len(c.name) > 0 {
		return c.name
	}
	if len(c.wmname) > 0 {
		return c.wmname
	}
	return "N/A"
}

func (c *client) Win() *window {
	return c.window
}

func (c *client) String() string {
	return fmt.Sprintf("%s (%X)", c.Name(), c.window.id)
}
