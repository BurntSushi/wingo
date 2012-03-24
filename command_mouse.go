package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

import (
    "github.com/BurntSushi/xgbutil"
    "github.com/BurntSushi/xgbutil/mousebind"
    "github.com/BurntSushi/xgbutil/xevent"
)

type mouseCommand struct {
    cmd string
    down bool // 'up' when false
    buttonStr string
    direction uint32 // only used by Resize command
}

func (mcmd mouseCommand) setup(c *client, wid xgb.Id) {
    // Check if this command is a drag... If it is, it needs special attention.
    if mcmd.cmd == "Move" {
        c.SetupMoveDrag(wid, mcmd.buttonStr, true)
        return
    }
    if mcmd.cmd == "Resize" {
        c.SetupResizeDrag(wid, mcmd.buttonStr, true, mcmd.direction)
        return
    }

    // Now check if it's a *real* command, like, in the shell
    var run func(c *client)
    if mcmd.cmd[0] == '`' && mcmd.cmd[len(mcmd.cmd) - 1] == '`' {
        // XXX TODO
        run = func(c *client) { }
    } else {
        run = getMouseCommand(mcmd.cmd)
        if run == nil {
            logWarning.Printf("Undefined mouse command: '%s'", mcmd.cmd)
            return
        }
    }

    // If we're putting this on the client window, we need to propagate
    // the events (i.e., grab synchronously).
    // Otherwise, we don't need to grab at all!
    if wid == c.window.id {
        mcmd.down = true // X dies otherwise, WEIRD!
        mcmd.attach(wid, func() { run(c) }, true, true)
    } else {
        mcmd.attach(wid, func() { run(c) }, false, false)
    }
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

func getMouseCommand(cmd string) func(c *client) {
    switch cmd {
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
    }

    return nil
}

