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
	workspace           *workspace
	layer               int
	name, vname, wmname string
	isMapped            bool
	initMap             bool
	state               int
	normal              bool
	forceFloating       bool
	maximized           bool
	iconified           bool
	initialMap          bool
	lastTime            int
	unmapIgnore         int
	hasStruts           bool

	types        []string
	hints        *icccm.Hints
	nhints       *icccm.NormalHints
	protocols    []string
	transientFor xgb.Id
	wmclass      *icccm.WmClass

	geomStore   map[string]*clientGeom
	promptStore map[string]*window

	frame        Frame
	frameNada    *frameNada
	frameSlim    *frameSlim
	frameBorders *frameBorders
	frameFull    *frameFull
}

func newClient(id xgb.Id) *client {
	return &client{
		window:        newWindow(id),
		workspace:     nil,
		layer:         stackDefault,
		name:          "",
		vname:         "",
		wmname:        "",
		isMapped:      false,
		initMap:       false,
		state:         StateInactive,
		normal:        true,
		forceFloating: false,
		maximized:     false,
		iconified:     false,
		initialMap:    false,
		lastTime:      0,
		unmapIgnore:   0,
		hints:         nil,
		nhints:        nil,
		protocols:     nil,
		transientFor:  0,
		wmclass:       nil,
		geomStore:     make(map[string]*clientGeom),
		promptStore:   make(map[string]*window),
		frame:         nil,
		frameNada:     nil,
		frameSlim:     nil,
		frameBorders:  nil,
		frameFull:     nil,
	}
}

func (c *client) unmanage() {
	if c.Mapped() {
		c.unmappedFallback()
	}
	c.workspace.remove(c)
	c.frame.Destroy()
	c.setWmState(icccm.StateWithdrawn)

	xevent.Detach(X, c.window.id)
	c.promptRemove()
	WM.stackRemove(c)
	WM.clientRemove(c)

	if c.normal {
		WM.focusRemove(c)
	}
	if c.hasStruts {
		WM.headsApplyStruts()
	}

	WM.updateEwmhStacking()
}

func (c *client) focusRaise() {
	if !c.normal {
		return
	}
	WM.focusAdd(c)
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
	if !c.Mapped() {
		return
	}
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
	if WM.wrkActive().id != c.workspace.id {
		c.workspace.activate(false, false)
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
	c.focusRaise()
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

func (c *client) Raise() {
	WM.stackRaise(c, false)

	// Also raise its transients if they are in the same layer...
	toRaise := []*client{c}
	for i := len(WM.stack) - 1; i >= 0; i-- {
		if c.transient(WM.stack[i]) {
			toRaise = append(toRaise, WM.stack[i])
		}
	}

	for _, c2 := range toRaise {
		WM.stackRaise(c2, false)
	}
	WM.stackUpdate(toRaise)
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
	case "WM_TRANSIENT_FOR":
		transientFor, err := icccm.WmTransientForGet(X, c.Id())
		if err == nil {
			c.transientFor = transientFor
		}
	case "_NET_WM_USER_TIME":
		newTime, err := ewmh.WmUserTimeGet(X, c.window.id)
		showVals(c.lastTime, newTime)
		if err == nil {
			c.lastTime = newTime
		}
	case "_NET_WM_STRUT_PARTIAL":
		WM.headsApplyStruts()
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

func (c *client) Frame() Frame {
	return c.frame
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
