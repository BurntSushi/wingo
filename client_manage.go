package main

import (
	"code.google.com/p/jamslam-x-go-binding/xgb"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xevent"

	"github.com/BurntSushi/wingo/logger"
)

// manage sets everything up to bring a client window into window management.
// It is still possible for us to bail.
func (c *client) manage() error {
	// Before we bring a client into window management, we need to populate
	// some information first. Sometimes this process results in us *not*
	// managing the client!
	_, err := c.Win().geometry()
	if err != nil {
		return err
	}

	err = c.initPopulate()
	if err != nil {
		return err
	}

	// time for reparenting/decorating
	c.initFrame()
	c.frame = c.frameFull
	FrameClientReset(c.Frame())
	c.Frame().On()

	// Reparent's sends an unmap, we need to ignore it!
	c.unmapIgnore++

	// time to add the client to the WM state
	WM.clientAdd(c)
	WM.focusAdd(c)
	c.Raise()

	// Some prompts need to do some heavy-lifting ONE time for each client.
	// (i.e., creating images.)
	// These images are added to the "prompt" map in each client.
	c.promptAdd()

	// Listen to the client for property and structure changes.
	c.window.listen(xgb.EventMaskPropertyChange |
		xgb.EventMaskStructureNotify)

	// attach some event handlers
	xevent.PropertyNotifyFun(
		func(X *xgbutil.XUtil, ev xevent.PropertyNotifyEvent) {
			c.updateProperty(ev)
		}).Connect(X, c.window.id)
	xevent.ConfigureRequestFun(
		func(X *xgbutil.XUtil, ev xevent.ConfigureRequestEvent) {
			// Don't honor configure requests when we're moving or resizing
			// Or if we're maximized. They need to oblige EWMH for that!
			if c.frame.Moving() || c.frame.Resizing() || c.maximized {
				return
			}

			flags := int(ev.ValueMask) & ^int(DoStack) & ^int(DoSibling)
			c.frame.ConfigureClient(flags, int(ev.X), int(ev.Y),
				int(ev.Width), int(ev.Height),
				ev.Sibling, ev.StackMode, false)
		}).Connect(X, c.window.id)
	xevent.UnmapNotifyFun(
		func(X *xgbutil.XUtil, ev xevent.UnmapNotifyEvent) {
			if !c.Mapped() {
				return
			}

			if c.unmapIgnore > 0 {
				c.unmapIgnore -= 1
				return
			}

			c.unmappedFallback()
			c.unmanage()
		}).Connect(X, c.window.id)
	xevent.DestroyNotifyFun(
		func(X *xgbutil.XUtil, ev xevent.DestroyNotifyEvent) {
			c.unmanage()
		}).Connect(X, c.window.id)

	c.clientMouseConfig()
	c.frameMouseConfig()

	// Find the current workspace and attach this client
	WM.WrkActive().Add(c, false)

	// If the initial state isn't iconic or is absent, then we can map
	if c.hints.Flags&icccm.HintState == 0 ||
		c.hints.InitialState != icccm.StateIconic {
		c.Map()
		c.Focus()
	}

	return nil
}

func (c *client) initFrame() {
	// We want one parent window for all frames.
	parent := newParent(c)

	c.frameNada = newFrameNada(parent, c)
	c.frameSlim = newFrameSlim(parent, c)
	c.frameBorders = newFrameBorders(parent, c)
	c.frameFull = newFrameFull(parent, c)
}

func (c *client) initPopulate() error {
	var err error

	c.hints, err = icccm.WmHintsGet(X, c.Id())
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

	c.nhints, err = icccm.WmNormalHintsGet(X, c.Id())
	if err != nil {
		logger.Warning.Println(err)
		logger.Message.Printf("Using reasonable defaults for WM_NORMAL_HINTS "+
			"for %X", c.Id())
		c.nhints = &icccm.NormalHints{}
	}

	c.protocols, err = icccm.WmProtocolsGet(X, c.Id())
	if err != nil {
		logger.Warning.Printf(
			"Window %X does not have WM_PROTOCOLS set.", c.Id())
		c.protocols = []string{}
	}

	c.name, err = ewmh.WmNameGet(X, c.Id())
	if err != nil {
		c.name = ""
		logger.Warning.Printf("Could not find name for window %X.", c.Id())
	}

	c.vname, _ = ewmh.WmVisibleNameGet(X, c.Id())
	c.wmname, _ = icccm.WmNameGet(X, c.Id())
	c.transientFor, _ = icccm.WmTransientForGet(X, c.Id())

	c.types, err = ewmh.WmWindowTypeGet(X, c.Id())
	if err != nil {
		logger.Warning.Printf("Could not find window type for window %X, "+
			"using 'normal'.", c.Id())
		c.types = []string{"_NET_WM_WINDOW_TYPE_NORMAL"}
	}
	return nil
}

func clientMapRequest(X *xgbutil.XUtil, ev xevent.MapRequestEvent) {
	X.Grab()
	defer X.Ungrab()

	// whoa whoa... what if we're already managing this window?
	for _, c := range WM.clients {
		if ev.Window == c.Id() {
			logger.Warning.Printf("Could not manage window %X because we are "+
				"already managing %s.", ev.Window, c)
			return
		}
	}

	client := newClient(ev.Window)

	err := client.manage()
	if err != nil {
		logger.Warning.Printf("Could not manage window %X because: %v\n",
			client, err)
		return
	}
}
