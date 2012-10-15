package main

import (
	"fmt"
	"time"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/logger"
)

// This file is responsible for implementing all the crud around getting
// and giving up the window manager's selection ownership. This is effectively
// how one window manager is able to "replace" another one. Aren't we so
// cooperative?

// own requests ownership over the role of window manager in the current
// X environment. It can fail if it does not successfully get ownership.
//
// When 'replace' is true, Wingo will attempt to replace an window manager
// that is currently running. Otherwise, Wingo will quit if a window manager
// is running.
func own(X *xgbutil.XUtil, replace bool) error {
	otherWmRunning := false
	otherWmName := ""
	otherWmOwner := xproto.Window(xproto.WindowNone)

	xTime, err := currentTime(X)
	if err != nil {
		return err
	}

	selAtom, err := managerAtom(X)
	if err != nil {
		return err
	}

	// Check to see if we need to replace. If so, determine whether to
	// continue based on `replace`.
	reply, err := xproto.GetSelectionOwner(X.Conn(), selAtom).Reply()
	if err != nil {
		return err
	}
	if reply.Owner != xproto.WindowNone {
		otherWmRunning = true
		otherWmOwner = reply.Owner

		// Look for the window manager's name for a nicer error message.
		otherWmName, err = ewmh.GetEwmhWM(X)
		if err != nil || len(otherWmName) == 0 {
			otherWmName = "Unknown"
		}

		// We need to listen for DestroyNotify events on the selection
		// owner in case we need to replace the WM.
		owner := xwindow.New(X, reply.Owner)
		if err = owner.Listen(xproto.EventMaskStructureNotify); err != nil {
			return err
		}
	}
	if otherWmRunning {
		if !replace {
			return fmt.Errorf(
				"Another window manager (%s) is already running. Please use "+
					"the '--replace' option to replace the current window "+
					"manager with Wingo.", otherWmName)
		} else {
			logger.Message.Printf(
				"Waiting for %s to shutdown and transfer ownership to us.",
				otherWmName)
		}
	}

	err = xproto.SetSelectionOwnerChecked(
		X.Conn(), X.Dummy(), selAtom, xTime).Check()
	if err != nil {
		return err
	}

	// Now we've got to make sure that we *actually* got ownership.
	reply, err = xproto.GetSelectionOwner(X.Conn(), selAtom).Reply()
	if err != nil {
		return err
	}
	if reply.Owner != X.Dummy() {
		return fmt.Errorf(
			"Could not acquire ownership with SetSelectionOwner. "+
				"GetSelectionOwner claims that '%d' is the owner, but '%d' "+
				"needs to be.", reply.Owner, X.Dummy())
	}

	// While X now acknowledges us as the selection owner, it's possible
	// that the window manager is misbehaving. ICCCM 2.8 calls for the previous
	// manager to destroy the selection owner when it's OK for us to take
	// over. Otherwise, listening to SubstructureRedirect on the root window
	// might fail if we move too quickly.
	timeout := time.After(3 * time.Second)
	if otherWmRunning {
	OTHER_WM_SHUTDOWN:
		for {
			select {
			case <-timeout:
				return fmt.Errorf(
					"Wingo failed to replace the currently running window "+
						"manager (%s). Namely, Wingo was not able to detect "+
						"that the current window manager had shut down.",
					otherWmName)
			default:
				ev, err := X.Conn().PollForEvent()
				if err != nil {
					continue
				}
				if destNotify, ok := ev.(xproto.DestroyNotifyEvent); ok {
					if destNotify.Window == otherWmOwner {
						break OTHER_WM_SHUTDOWN
					}
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
	}

	logger.Message.Println("Wingo has window manager ownership!")
	announce(X)

	// Listen for SelectionClear events. When we get one of these, then we
	// know a window manager is trying to replace us.
	xevent.SelectionClearFun(disown).Connect(X, X.Dummy())
	return nil
}

// disown responds to SelectionClear events on our dummy window.
// This means another window manager is trying to run, so quit.
func disown(X *xgbutil.XUtil, ev xevent.SelectionClearEvent) {
	logger.Message.Println(
		"Some other window manager is replacing us. Exiting...")
	xevent.Quit(X)
}

// announce sends a ClientMessage event to the root window to let everyone
// know that Wingo is the boss. (As per ICCCM 2.8.)
func announce(X *xgbutil.XUtil) {
	typAtom, err := xprop.Atm(X, "MANAGER")
	if err != nil {
		logger.Warning.Println(err)
		return
	}
	manSelAtom, err := managerAtom(X)
	if err != nil {
		logger.Warning.Println(err)
		return
	}
	cm, err := xevent.NewClientMessage(32, X.RootWin(), typAtom,
		int(X.TimeGet()), int(manSelAtom), int(X.Dummy()))
	xproto.SendEvent(X.Conn(), false, X.RootWin(),
		xproto.EventMaskStructureNotify, string(cm.Bytes()))
}

// managerAtom returns an xproto.Atom of the manager selection atom.
// Usually it's "WM_S0", where "0" is the screen number.
func managerAtom(X *xgbutil.XUtil) (xproto.Atom, error) {
	name := fmt.Sprintf("WM_S%d", X.Conn().DefaultScreen)
	return xprop.Atm(X, name)
}

// currentTime forcefully causes a PropertyNotify event to fire on the root
// window, then scans the event queue and picks up the time.
//
// It is NOT SAFE to call this function in a place other than Wingo's
// initialization. Namely, this function subverts xevent's queue and reads
// events directly from X.
func currentTime(X *xgbutil.XUtil) (xproto.Timestamp, error) {
	wmClassAtom, err := xprop.Atm(X, "WM_CLASS")
	if err != nil {
		return 0, err
	}

	stringAtom, err := xprop.Atm(X, "STRING")
	if err != nil {
		return 0, err
	}

	// Make sure we're listening to PropertyChange events on the root window.
	err = xwindow.New(X, X.RootWin()).Listen(xproto.EventMaskPropertyChange)
	if err != nil {
		return 0, fmt.Errorf(
			"Could not listen to Root window events (PropertyChange): %s", err)
	}

	// Do a zero-length append on a property as suggested by ICCCM 2.1.
	err = xproto.ChangePropertyChecked(
		X.Conn(), xproto.PropModeAppend, X.RootWin(),
		wmClassAtom, stringAtom, 8, 0, nil).Check()
	if err != nil {
		return 0, err
	}

	// Now look for the PropertyNotify generated by that zero-length append
	// and return the timestamp attached to that event.
	// Note that we do this outside of xgbutil/xevent, since ownership
	// is literally the first thing we do after connecting to X.
	// (i.e., we don't have our event handling system initialized yet.)
	timeout := time.After(3 * time.Second)
	for {
		select {
		case <-timeout:
			return 0, fmt.Errorf(
				"Expected a PropertyNotify event to get a valid timestamp, " +
					"but never received one.")
		default:
			ev, err := X.Conn().PollForEvent()
			if err != nil {
				continue
			}
			if propNotify, ok := ev.(xproto.PropertyNotifyEvent); ok {
				X.TimeSet(propNotify.Time) // why not?
				return propNotify.Time, nil
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
	panic("unreachable")
}
