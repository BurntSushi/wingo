package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

import (
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/mousebind"
	"github.com/BurntSushi/xgbutil/xevent"
)

type mouseCommand struct {
	cmd       string
	down      bool // 'up' when false
	buttonStr string
	direction uint32 // only used by Resize command
}

func (mcmd mouseCommand) setup(c *client, wid xgb.Id) {
	// Check if this command is a drag... If it is, it needs special attention.
	if mcmd.cmd == "Move" {
		c.setupMoveDrag(wid, mcmd.buttonStr, true)
		return
	}
	if mcmd.cmd == "Resize" {
		c.setupResizeDrag(wid, mcmd.buttonStr, true, mcmd.direction)
		return
	}

	// If we're putting this on the client or frame window, we need to propagate
	// the events (i.e., grab synchronously).
	// Otherwise, we don't need to grab at all!
	run := mcmd.commandFun()
	if wid == c.Id() || (c.Frame() != nil && wid == c.Frame().ParentId()) {
		if mcmd.down {
			mcmd.attach(wid, func() { run(c) }, true, true)
		} else { // we have to handle release grabs specially!
			mcmd.attachClick(wid, func() { run(c) })
		}
	} else {
		mcmd.attach(wid, func() { run(c) }, false, false)
	}
}

// setupMoveDrag does the boiler plate for registering this client's
// "move" drag.
func (c *client) setupMoveDrag(dragWin xgb.Id, buttonStr string, grab bool) {
	dStart := xgbutil.MouseDragBeginFun(
		func(X *xgbutil.XUtil, rx, ry, ex, ey int) (bool, xgb.Id) {
			frameMoveBegin(c.Frame(), rx, ry, ex, ey)
			return true, cursorFleur
		})
	dStep := xgbutil.MouseDragFun(
		func(X *xgbutil.XUtil, rx, ry, ex, ey int) {
			frameMoveStep(c.Frame(), rx, ry, ex, ey)
		})
	dEnd := xgbutil.MouseDragFun(
		func(X *xgbutil.XUtil, rx, ry, ex, ey int) {
			frameMoveEnd(c.Frame(), rx, ry, ex, ey)
		})
	mousebind.Drag(X, dragWin, buttonStr, grab, dStart, dStep, dEnd)
}

// setupResizeDrag does the boiler plate for registering this client's
// "resize" drag.
func (c *client) setupResizeDrag(dragWin xgb.Id, buttonStr string, grab bool,
	direction uint32) {

	dStart := xgbutil.MouseDragBeginFun(
		func(X *xgbutil.XUtil, rx, ry, ex, ey int) (bool, xgb.Id) {
			return frameResizeBegin(c.Frame(), direction, rx, ry, ex, ey)
		})
	dStep := xgbutil.MouseDragFun(
		func(X *xgbutil.XUtil, rx, ry, ex, ey int) {
			frameResizeStep(c.Frame(), rx, ry, ex, ey)
		})
	dEnd := xgbutil.MouseDragFun(
		func(X *xgbutil.XUtil, rx, ry, ex, ey int) {
			frameResizeEnd(c.Frame(), rx, ry, ex, ey)
		})
	mousebind.Drag(X, dragWin, buttonStr, grab, dStart, dStep, dEnd)
}

func (mcmd mouseCommand) attachClick(wid xgb.Id, run func()) {
	mousebind.ButtonPressFun(
		func(X *xgbutil.XUtil, ev xevent.ButtonPressEvent) {
			// empty
		}).Connect(X, wid, mcmd.buttonStr, false, true)
	mousebind.ButtonReleaseFun(
		func(X *xgbutil.XUtil, ev xevent.ButtonReleaseEvent) {
			run()
		}).Connect(X, wid, mcmd.buttonStr, false, false)
}

func (mcmd mouseCommand) attach(wid xgb.Id, run func(), propagate, grab bool) {
	if mcmd.down {
		mousebind.ButtonPressFun(
			func(X *xgbutil.XUtil, ev xevent.ButtonPressEvent) {
				run()
			}).Connect(X, wid, mcmd.buttonStr, propagate, grab)
	} else {
		mousebind.ButtonReleaseFun(
			func(X *xgbutil.XUtil, ev xevent.ButtonReleaseEvent) {
				run()
			}).Connect(X, wid, mcmd.buttonStr, propagate, grab)
	}
}

func rootMouseConfig() {
	for _, mcmd := range CONF.mouse["root"] {
		run := getRootMouseCommand(mcmd.cmd)
		if run == nil {
			logWarning.Printf("Undefined root mouse command: '%s'", mcmd.cmd)
			continue
		}
		mcmd.attach(ROOT.id, run, false, false)
	}
}

func (c *client) clientMouseConfig() {
	for _, mcmd := range CONF.mouse["client"] {
		mcmd.setup(c, c.window.id)
	}
}

func (c *client) frameMouseConfig() {
	for _, mcmd := range CONF.mouse["frame"] {
		mcmd.setup(c, c.Frame().ParentId())
	}
}

func (c *client) framePieceMouseConfig(piece string, pieceid xgb.Id) {
	for _, mcmd := range CONF.mouse[piece] {
		mcmd.setup(c, pieceid)
	}
}

func (mcmd mouseCommand) commandFun() func(c *client) {
	tryShellFun := commandShellFun(mcmd.cmd)
	if tryShellFun != nil {
		return func(c *client) {
			tryShellFun()
			xevent.ReplayPointer(X)
		}
	}

	switch mcmd.cmd {
	case "FocusRaise":
		return func(c *client) {
			c.Focus()
			c.Raise()
			xevent.ReplayPointer(X)
		}
	case "Focus":
		return func(c *client) {
			c.Focus()
			xevent.ReplayPointer(X)
		}
	case "Raise":
		return func(c *client) {
			c.Raise()
			xevent.ReplayPointer(X)
		}
	case "Close":
		return func(c *client) {
			c.Close()
		}
	case "MaximizeToggle":
		return func(c *client) {
			c.MaximizeToggle()
		}
	case "Minimize":
		return func(c *client) {
			c.IconifyToggle()
		}
	}

	logWarning.Printf("Undefined mouse command: '%s'", mcmd.cmd)

	return nil
}

func getRootMouseCommand(cmd string) func() {
	switch cmd {
	case "Focus":
		return func() {
			ROOT.focus()
			WM.unfocusExcept(0)
		}
	}

	return nil
}
