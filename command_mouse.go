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
	"github.com/BurntSushi/gribble"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/mousebind"
	"github.com/BurntSushi/xgbutil/xevent"

	"github.com/BurntSushi/wingo/cursors"
)

// mouseClientClicked is a terrible hack to inject state into commands.
// Basically, if a command is given with "::mouse::" as the argument for
// a client parameter, this variable will be checked and its value will
// be used.
var mouseClientClicked *client

type mouseCommand struct {
	cmd       gribble.Command
	cmdName   string
	down      bool // 'up' when false
	buttonStr string
}

func (mcmd mouseCommand) setup(c *client, wid xproto.Window) {
	// Check if this command is a drag... If it is, it needs special attention.
	if mcmd.cmdName == "MouseMove" {
		c.setupMoveDrag(wid, mcmd.buttonStr, true)
		return
	}
	if mcmd.cmdName == "MouseResize" {
		direction := strToDirection(mcmd.cmd.(*CmdMouseResize).Direction)
		c.setupResizeDrag(wid, mcmd.buttonStr, true, direction)
		return
	}

	// If we're putting this on the client or frame window, we need to propagate
	// the events (i.e., grab synchronously).
	// Otherwise, we don't need to grab at all!
	run := func() {
		mouseClientClicked = c
		mcmd.cmd.Run()
		mouseClientClicked = nil
	}
	if wid == c.Id() || (c.Frame() != nil && wid == c.Frame().Parent().Id) {
		if mcmd.down {
			mcmd.attach(wid, run, true, true)
		} else { // we have to handle release grabs specially!
			mcmd.attachGrabRelease(wid, run)
		}
	} else {
		mcmd.attach(wid, run, false, false)
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

// attach sets up the event handlers for a mouse button press OR release.
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

// attachGrabRelease is a special case of 'attach' that is necessary when
// attaching a mouse release event to either the client or frame window.
//
// TODO: Recall and document *why* this is needed.
func (mcmd mouseCommand) attachGrabRelease(wid xproto.Window, run func()) {
	mousebind.ButtonPressFun(
		func(X *xgbutil.XUtil, ev xevent.ButtonPressEvent) {
			// empty
		}).Connect(X, wid, mcmd.buttonStr, false, true)
	mousebind.ButtonReleaseFun(
		func(X *xgbutil.XUtil, ev xevent.ButtonReleaseEvent) {
			run()
		}).Connect(X, wid, mcmd.buttonStr, false, false)
}

func rootMouseConfig() {
	for _, mcmd := range wingo.conf.mouse["root"] {
		mcmd.attach(wingo.root.Id, func() { mcmd.cmd.Run() }, false, false)
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
