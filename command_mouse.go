package main

/*
	I haven't thought of a way to reconcile Mouse commands with the rest
	of the commands. It may be easier to leave them separate.

	Basically, the problem is that a mouse command operates on a client whereas
	key (or general) commands always operate on the current state of Wingo.
	The former requires parameterization over a client whereas the latter
	can simply rely on the current global state.

	This makes them fundamentally different.
*/

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/mousebind"
	"github.com/BurntSushi/xgbutil/xevent"

	"github.com/BurntSushi/wingo/cursors"
	"github.com/BurntSushi/wingo/focus"
	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/stack"
)

type mouseCommand struct {
	cmd       string
	down      bool // 'up' when false
	buttonStr string
	direction uint32 // only used by Resize command
}

func (mcmd mouseCommand) setup(c *client, wid xproto.Window) {
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
	if wid == c.Id() || (c.Frame() != nil && wid == c.Frame().Parent().Id) {
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
func (c *client) setupMoveDrag(dragWin xproto.Window,
	buttonStr string, grab bool) {

	dStart := xgbutil.MouseDragBeginFun(
		func(X *xgbutil.XUtil, rx, ry, ex, ey int) (bool, xproto.Cursor) {
			c.dragMoveBegin(rx, ry, ex, ey)
			return true, cursors.Fleur
		})
	dStep := xgbutil.MouseDragFun(
		func(X *xgbutil.XUtil, rx, ry, ex, ey int) {
			c.dragMoveStep(rx, ry, ex, ey)
		})
	dEnd := xgbutil.MouseDragFun(
		func(X *xgbutil.XUtil, rx, ry, ex, ey int) {
			c.dragMoveEnd(rx, ry, ex, ey)
		})
	mousebind.Drag(X, X.Dummy(), dragWin, buttonStr, grab, dStart, dStep, dEnd)
}

// setupResizeDrag does the boiler plate for registering this client's
// "resize" drag.
func (c *client) setupResizeDrag(dragWin xproto.Window,
	buttonStr string, grab bool, direction uint32) {

	dStart := xgbutil.MouseDragBeginFun(
		func(X *xgbutil.XUtil, rx, ry, ex, ey int) (bool, xproto.Cursor) {
			return c.dragResizeBegin(direction, rx, ry, ex, ey)
		})
	dStep := xgbutil.MouseDragFun(
		func(X *xgbutil.XUtil, rx, ry, ex, ey int) {
			c.dragResizeStep(rx, ry, ex, ey)
		})
	dEnd := xgbutil.MouseDragFun(
		func(X *xgbutil.XUtil, rx, ry, ex, ey int) {
			c.dragResizeEnd(rx, ry, ex, ey)
		})
	mousebind.Drag(X, X.Dummy(), dragWin, buttonStr, grab, dStart, dStep, dEnd)
}

func (mcmd mouseCommand) attachClick(wid xproto.Window, run func()) {
	mousebind.ButtonPressFun(
		func(X *xgbutil.XUtil, ev xevent.ButtonPressEvent) {
			// empty
		}).Connect(X, wid, mcmd.buttonStr, false, true)
	mousebind.ButtonReleaseFun(
		func(X *xgbutil.XUtil, ev xevent.ButtonReleaseEvent) {
			run()
		}).Connect(X, wid, mcmd.buttonStr, false, false)
}

func (mcmd mouseCommand) attach(wid xproto.Window, run func(),
	propagate, grab bool) {

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
	for _, mcmd := range wingo.conf.mouse["root"] {
		run := getRootMouseCommand(mcmd.cmd)
		if run == nil {
			logger.Warning.Printf(
				"Undefined root mouse command: '%s'", mcmd.cmd)
			continue
		}
		mcmd.attach(wingo.root.Id, run, false, false)
	}
}

func (c *client) clientMouseConfig() {
	for _, mcmd := range wingo.conf.mouse["client"] {
		mcmd.setup(c, c.Id())
	}
}

func (c *client) frameMouseConfig() {
	for _, mcmd := range wingo.conf.mouse["frame"] {
		mcmd.setup(c, c.Frame().Parent().Id)
	}
}

func (c *client) FramePieceMouseConfig(piece string, pieceid xproto.Window) {
	for _, mcmd := range wingo.conf.mouse[piece] {
		mcmd.setup(c, pieceid)
	}
}

func (mcmd mouseCommand) commandFun() func(c *client) {
	// tryShellFun := commandShellFun(mcmd.cmd) 
	// if tryShellFun != nil { 
	// return func(c *client) { 
	// tryShellFun() 
	// xevent.ReplayPointer(X) 
	// } 
	// } 

	switch mcmd.cmd {
	case "FocusRaise":
		return func(c *client) {
			focus.Focus(c)
			stack.Raise(c)
			xevent.ReplayPointer(X)
		}
	case "Focus":
		return func(c *client) {
			focus.Focus(c)
			xevent.ReplayPointer(X)
		}
	case "Raise":
		return func(c *client) {
			stack.Raise(c)
			xevent.ReplayPointer(X)
		}
	case "Close":
		return func(c *client) {
			c.Close()
		}
	case "MaximizeToggle":
		return func(c *client) {
			// c.MaximizeToggle() 
		}
	case "Minimize":
		return func(c *client) {
			c.workspace.IconifyToggle(c)
		}
	}

	logger.Warning.Printf("Undefined mouse command: '%s'", mcmd.cmd)

	return nil
}

func getRootMouseCommand(cmd string) func() {
	switch cmd {
	case "Focus":
		return func() {
			focus.Root()
		}
	}

	return nil
}
