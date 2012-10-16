package xclient

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xprop"

	"github.com/BurntSushi/wingo/focus"
	"github.com/BurntSushi/wingo/frame"
	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/stack"
	"github.com/BurntSushi/wingo/wm"
)

func (c *Client) handleClientMessage(name string, data []uint32) {
	switch name {
	case "WM_CHANGE_STATE":
		if data[0] == icccm.StateIconic && !c.iconified {
			c.IconifyToggle()
		}
	case "_NET_ACTIVE_WINDOW":
		focus.Focus(c)
		stack.Raise(c)
	case "_NET_CLOSE_WINDOW":
		c.Close()
	case "_NET_MOVERESIZE_WINDOW":
		// The data[0] element contains bit-packed information. See
		// EWMH _NET_MOVERESIZE_WINDOW for the deets.
		gravity := int(data[0] & 0xff)
		xflags := int((data[0] >> 8) & 0xf)
		x, y, w, h := frame.ClientToFrame(c.frame, gravity,
			int(data[1]), int(data[2]), int(data[3]), int(data[4]))
		c.LayoutMROpt(xflags, x, y, w, h)
	case "_NET_RESTACK_WINDOW":
		// We basically treat this as a request to stack the window.
		// We ignore the sibling. Maybe someday we can support that, but eh...
		stack.Raise(c)
	case "_NET_WM_DESKTOP":
		if data[0] == 0xFFFFFFFF {
			c.stick()
			return
		}
		if wrk := wm.Heads.Workspaces.Get(int(data[0])); wrk != nil {
			wrk.Add(c)
		} else {
			logger.Warning.Printf(
				"_NET_WM_DESKTOP ClientMessage: No workspace indexed at '%d' "+
					"exists.", data[0])
		}
	case "_NET_WM_STATE":
		prop1, _ := xprop.AtomName(wm.X, xproto.Atom(data[1]))
		prop2, _ := xprop.AtomName(wm.X, xproto.Atom(data[2]))
		switch data[0] {
		case 0:
			c.updateStates("remove", prop1, prop2)
		case 1:
			c.updateStates("add", prop1, prop2)
		case 2:
			c.updateStates("toggle", prop1, prop2)
		default:
			logger.Warning.Printf(
				"_NET_WM_STATE: Unknown action '%d'.", data[0])
		}
	default:
		logger.Warning.Printf("Unknown ClientMessage for '%s': %s.", c, name)
	}
}

func (c *Client) updateStates(action, prop1, prop2 string) {
	// Since we don't support vertical XOR horizontal states and only both,
	// check if prop1 and prop2 are vert and horz and treat it as a maximize
	// request. Otherwise, process prop1 and prop2 independently.
	if (prop1 == "_NET_WM_STATE_MAXIMIZED_VERT" &&
		prop2 == "_NET_WM_STATE_MAXIMIZED_HORZ") ||
		(prop1 == "_NET_WM_STATE_MAXIMIZED_HORZ" &&
			prop2 == "_NET_WM_STATE_MAXIMIZED_VERT") {

		c.updateState(action, "_NET_WM_STATE_MAXIMIZED")
	} else {
		if len(prop1) > 0 {
			c.updateState(action, prop1)
		}
		if len(prop2) > 0 {
			c.updateState(action, prop2)
		}
	}
}

func (c *Client) updateState(action, prop string) {
	switch prop {
	case "_NET_WM_STATE_STICKY":
		switch action {
		case "remove":
			c.unstick()
		case "add":
			c.stick()
		case "toggle":
			c.StickyToggle()
		}
	case "_NET_WM_STATE_MAXIMIZED":
		switch action {
		case "remove":
			c.Unmaximize()
		case "add":
			c.Maximize()
		case "toggle":
			c.MaximizeToggle()
		}
	case "_NET_WM_STATE_SKIP_TASKBAR":
		switch action {
		case "remove":
			c.skipTaskbar = false
			c.removeState("_NET_WM_STATE_SKIP_TASKBAR")
		case "add":
			c.skipTaskbar = true
			c.addState("_NET_WM_STATE_SKIP_TASKBAR")
		case "toggle":
			c.skipTaskbar = !c.skipTaskbar
			if c.skipTaskbar {
				c.addState("_NET_WM_STATE_SKIP_TASKBAR")
			} else {
				c.removeState("_NET_WM_STATE_SKIP_TASKBAR")
			}
		}
	case "_NET_WM_STATE_SKIP_PAGER":
		switch action {
		case "remove":
			c.skipPager = false
			c.removeState("_NET_WM_STATE_SKIP_PAGER")
		case "add":
			c.skipPager = true
			c.addState("_NET_WM_STATE_SKIP_PAGER")
		case "toggle":
			c.skipPager = !c.skipPager
			if c.skipPager {
				c.addState("_NET_WM_STATE_SKIP_PAGER")
			} else {
				c.removeState("_NET_WM_STATE_SKIP_PAGER")
			}
		}
	case "_NET_WM_STATE_HIDDEN":
		switch action {
		case "remove":
			if c.Iconified() {
				c.IconifyToggle()
			}
		case "add":
			if !c.Iconified() {
				c.IconifyToggle()
			}
		case "toggle":
			c.IconifyToggle()
		}
	case "_NET_WM_STATE_FULLSCREEN":
		switch action {
		case "remove":
			c.unfullscreened()
		case "add":
			c.fullscreened()
		case "toggle":
			if c.fullscreen {
				c.unfullscreened()
			} else {
				c.fullscreened()
			}
		}
	case "_NET_WM_STATE_ABOVE":
		switch action {
		case "remove":
			c.unstackAbove()
		case "add":
			c.stackAbove()
		case "toggle":
			c.StackAboveToggle()
		}
	case "_NET_WM_STATE_BELOW":
		switch action {
		case "remove":
			c.unstackBelow()
		case "add":
			c.stackBelow()
		case "toggle":
			c.StackBelowToggle()
		}
	case "_NET_WM_STATE_DEMANDS_ATTENTION":
		switch action {
		case "remove":
			c.attnStop()
		case "add":
			c.attnStart()
		case "toggle":
			if c.demanding {
				c.attnStop()
			} else {
				c.attnStart()
			}
		}
	default:
		logger.Warning.Printf("_NET_WM_STATE: Unsupported state '%s'.", prop)
	}
}

func (c *Client) addState(name string) {
	if strIndex(name, c.winStates) == -1 {
		c.winStates = append(c.winStates, name)
		ewmh.WmStateSet(wm.X, c.Id(), c.winStates)
	}
}

func (c *Client) removeState(name string) {
	if i := strIndex(name, c.winStates); i > -1 {
		c.winStates = append(c.winStates[:i], c.winStates[i+1:]...)
		ewmh.WmStateSet(wm.X, c.Id(), c.winStates)
	}
}

func (c *Client) refreshState() {
	atoms := make([]string, 0, 4)

	// ignoring _NET_WM_STATE_MODAL
	if c.sticky {
		atoms = append(atoms, "_NET_WM_STATE_STICKY")
	}
	if c.maximized {
		atoms = append(atoms, "_NET_WM_STATE_MAXIMIZED_VERT")
		atoms = append(atoms, "_NET_WM_STATE_MAXIMIZED_HORZ")
	}
	// ignoring _NET_WM_STATE_SHADED
	if c.skipTaskbar {
		atoms = append(atoms, "_NET_WM_STATE_SKIP_TASKBAR")
	}
	if c.skipPager {
		atoms = append(atoms, "_NET_WM_STATE_SKIP_PAGER")
	}
	if c.Iconified() {
		atoms = append(atoms, "_NET_WM_STATE_HIDDEN")
	}
	if c.fullscreen {
		atoms = append(atoms, "_NET_WM_STATE_FULLSCREEN")
	}
	switch c.layer {
	case stack.LayerAbove:
		atoms = append(atoms, "_NET_WM_STATE_ABOVE")
	case stack.LayerBelow:
		atoms = append(atoms, "_NET_WM_STATE_BELOW")
	}
	// ignoring _NET_WM_STATE_DEMANDS_ATTENTION
	// ignoring _NET_WM_STATE_FOCUSED

	ewmh.WmStateSet(wm.X, c.Id(), atoms)
}
