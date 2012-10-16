package xclient

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"

	"github.com/BurntSushi/wingo/focus"
	"github.com/BurntSushi/wingo/frame"
	"github.com/BurntSushi/wingo/layout"
	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/wm"
)

func (c *Client) attachEventCallbacks() {
	c.win.Listen(xproto.EventMaskPropertyChange |
		xproto.EventMaskStructureNotify)

	pid := c.Frame().Parent().Id
	attrs, err := xproto.GetWindowAttributes(wm.X.Conn(), pid).Reply()
	if err == nil {
		masks := int(attrs.YourEventMask)
		if wm.Config.Ffm {
			masks |= xproto.EventMaskEnterWindow
		}
		c.Frame().Parent().Listen(masks)
	}

	c.cbMapNotify().Connect(wm.X, c.Id())
	c.cbUnmapNotify().Connect(wm.X, c.Id())
	c.cbDestroyNotify().Connect(wm.X, c.Id())
	c.cbConfigureRequest().Connect(wm.X, c.Id())
	c.cbPropertyNotify().Connect(wm.X, c.Id())
	c.cbClientMessage().Connect(wm.X, c.Id())

	// Focus follows mouse?
	if wm.Config.Ffm {
		c.cbEnterNotify().Connect(wm.X, c.Frame().Parent().Id)
	}
	c.handleFocusIn().Connect(wm.X, c.Frame().Parent().Id)
	c.handleFocusOut().Connect(wm.X, c.Frame().Parent().Id)

	wm.ClientMouseSetup(c)
	wm.FrameMouseSetup(c, c.frame.Parent().Id)
}

func (c *Client) cbMapNotify() xevent.MapNotifyFun {
	f := func(X *xgbutil.XUtil, ev xevent.MapNotifyEvent) {
		if c.IsMapped() {
			return
		}

		// I don't know what to do here if the client wasn't iconified.
		// It *should* have been unmanaged if the client issues an Unmap
		// on its own...
		if !c.iconified {
			logger.Warning.Printf("POSSIBLE BUG: Client '%s' is trying to map "+
				"itself, but Wingo doesn't think it was iconified.", c)
			return
		}

		c.IconifyToggle()
	}
	return xevent.MapNotifyFun(f)
}

func (c *Client) cbDestroyNotify() xevent.DestroyNotifyFun {
	f := func(X *xgbutil.XUtil, ev xevent.DestroyNotifyEvent) {
		c.unmanage()
	}
	return xevent.DestroyNotifyFun(f)
}

func (c *Client) cbUnmapNotify() xevent.UnmapNotifyFun {
	f := func(X *xgbutil.XUtil, ev xevent.UnmapNotifyEvent) {
		// When a client issues an Unmap request, the window manager should
		// unmanage it. However, when wingo unmaps the window, we shouldn't
		// unmanage it. Thus, every time wingo unmaps the window, the
		// unmapIgnore counter is incremented. Only when it is zero does it mean
		// that we should unmanage the client (i.e., the unmap request came
		// from somewhere other than Wingo.)
		if c.unmapIgnore > 0 {
			c.unmapIgnore--
			return
		}
		c.unmanage()
	}
	return xevent.UnmapNotifyFun(f)
}

func (c *Client) cbConfigureRequest() xevent.ConfigureRequestFun {
	f := func(X *xgbutil.XUtil, ev xevent.ConfigureRequestEvent) {
		if c.frame.Moving() || c.frame.Resizing() || c.maximized {
			logger.Lots.Printf("Denying ConfigureRequest from client because " +
				"the client is in the processing of moving/resizing, or is " +
				"maximized.")

			// As per ICCCM 4.1.5, a window that has not been moved or resized
			// must receive a synthetic ConfigureNotify event.
			c.sendConfigureNotify()
			return
		}
		if _, ok := c.Layout().(layout.Floater); !ok {
			logger.Lots.Printf("Denying ConfigureRequest from client because " +
				"the client is not in a floating layout.")

			// As per ICCCM 4.1.5, a window that has not been moved or resized
			// must receive a synthetic ConfigureNotify event.
			c.sendConfigureNotify()
			return
		}

		flags := int(ev.ValueMask) &
			^int(xproto.ConfigWindowStackMode) &
			^int(xproto.ConfigWindowSibling)
		x, y, w, h := frame.ClientToFrame(c.frame, -1,
			int(ev.X), int(ev.Y), int(ev.Width), int(ev.Height))
		c.LayoutMROpt(flags, x, y, w, h)
	}
	return xevent.ConfigureRequestFun(f)
}

func (c *Client) sendConfigureNotify() {
	geom := c.Frame().Geom()
	ev := xproto.ConfigureNotifyEvent{
		Event:            c.Id(),
		Window:           c.Id(),
		AboveSibling:     0,
		X:                int16(geom.X()),
		Y:                int16(geom.Y()),
		Width:            uint16(c.win.Geom.Width()),
		Height:           uint16(c.win.Geom.Height()),
		BorderWidth:      0,
		OverrideRedirect: false,
	}
	xproto.SendEvent(wm.X.Conn(), false, c.Id(),
		xproto.EventMaskStructureNotify, string(ev.Bytes()))
}

func (c *Client) cbPropertyNotify() xevent.PropertyNotifyFun {
	f := func(X *xgbutil.XUtil, ev xevent.PropertyNotifyEvent) {
		name, err := xprop.AtomName(wm.X, ev.Atom)
		if err != nil {
			logger.Warning.Printf("Could not get property atom name for '%s' "+
				"because: %s.", ev, err)
			return
		}

		logger.Lots.Printf("Updating property %s with state %v on window %s",
			name, ev.State, c)
		c.handleProperty(name)
	}
	return xevent.PropertyNotifyFun(f)
}

func (c *Client) cbClientMessage() xevent.ClientMessageFun {
	f := func(X *xgbutil.XUtil, ev xevent.ClientMessageEvent) {
		name, err := xprop.AtomName(wm.X, ev.Type)
		if err != nil {
			logger.Warning.Printf("Could not get property atom name for "+
				"ClientMessage event on '%s': %s.", c, err)
			return
		}

		logger.Lots.Printf(
			"Handling ClientMessage '%s' on client '%s'.", name, c)
		c.handleClientMessage(name, ev.Data.Data32)
	}
	return xevent.ClientMessageFun(f)
}

func ignoreFocus(modeByte, detailByte byte) bool {
	mode, detail := focus.Modes[modeByte], focus.Details[detailByte]

	if mode == "NotifyGrab" || mode == "NotifyUngrab" {
		return true
	}
	if detail == "NotifyAncestor" ||
		detail == "NotifyInferior" ||
		detail == "NotifyNonlinear" ||
		detail == "NotifyPointer" ||
		detail == "NotifyPointerRoot" ||
		detail == "NotifyNone" {

		return true
	}
	// Only accept modes: NotifyNormal and NotifyWhileGrabbed
	// Only accept details: NotifyVirtual, NotifyNonlinearVirtual
	return false
}

func (c *Client) cbEnterNotify() xevent.EnterNotifyFun {
	f := func(X *xgbutil.XUtil, ev xevent.EnterNotifyEvent) {
		focus.Focus(c)
	}
	return xevent.EnterNotifyFun(f)
}

func (c *Client) handleFocusIn() xevent.FocusInFun {
	f := func(X *xgbutil.XUtil, ev xevent.FocusInEvent) {
		if ignoreFocus(ev.Mode, ev.Detail) {
			return
		}

		c.Focused()
		// logger.Debug.Println("---------------------------------------------")
		// logger.Debug.Println("Focus In") 
		// logger.Debug.Printf("Window: %s", c.Name()) 
		// logger.Debug.Printf("Mode: %s", modes[ev.Mode]) 
		// logger.Debug.Printf("Detail: %s", details[ev.Detail]) 
		// logger.Debug.Println("---------------------------------------------")
	}
	return xevent.FocusInFun(f)
}

func (c *Client) handleFocusOut() xevent.FocusOutFun {
	f := func(X *xgbutil.XUtil, ev xevent.FocusOutEvent) {
		if ignoreFocus(ev.Mode, ev.Detail) {
			return
		}
		c.Unfocused()

		// logger.Debug.Println("---------------------------------------------")
		// logger.Debug.Println("Focus Out") 
		// logger.Debug.Printf("Window: %s", c.Name()) 
		// logger.Debug.Printf("Mode: %s", modes[ev.Mode]) 
		// logger.Debug.Printf("Detail: %s", details[ev.Detail]) 
		// logger.Debug.Println("---------------------------------------------")
	}
	return xevent.FocusOutFun(f)
}
