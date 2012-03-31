// A set of functions that are key-bindable
package main

import "time"

// Shortcut for executing Client interface functions that have no parameters
// and no return values on the currently focused window.
func withFocused(f func(c *client)) {
    focused := WM.focused()
    if focused != nil {
        f(focused)
    }
}

func cmd_workspace_prev() {
    wrkAct := WM.WrkActive()
    WM.WrkSet(mod(wrkAct.id - 1, len(WM.workspaces)))
}

func cmd_workspace_next() {
    wrkAct := WM.WrkActive()
    WM.WrkSet(mod(wrkAct.id + 1, len(WM.workspaces)))
}

func cmd_active_workspace_prev() {
    withFocused(func(c *client) {
        wrkAct := WM.WrkActive()
        wrkPrev := WM.workspaces[mod(wrkAct.id - 1, len(WM.workspaces))]
        wrkPrev.Add(c, false)
        WM.WrkSet(wrkPrev.id)

        c.Raise()
    })
}

func cmd_active_workspace_next() {
    withFocused(func(c *client) {
        wrkAct := WM.WrkActive()
        wrkNext := WM.workspaces[mod(wrkAct.id + 1, len(WM.workspaces))]
        wrkNext.Add(c, false)
        WM.WrkSet(wrkNext.id)

        c.Raise()
    })
}

func cmd_active_test1() {
    withFocused(func(c *client) {
        FrameMR(c.Frame(), DoX | DoY, 0, 0, 0, 0, false)
    })
}

func cmd_active_close() {
    withFocused(func(c *client) {
        c.Close()
    })
}

func cmd_active_maximize_toggle() {
    withFocused(func(c *client) {
        c.MaximizeToggle()
    })
}

func cmd_active_frame_nada() {
    withFocused(func(c *client) {
        c.FrameNada()
    })
}

func cmd_active_frame_slim() {
    withFocused(func(c *client) {
        c.FrameSlim()
    })
}

func cmd_active_frame_borders() {
    withFocused(func(c *client) {
        c.FrameBorders()
    })
}

func cmd_active_frame_full() {
    withFocused(func(c *client) {
        c.FrameFull()
    })
}

// This is a start, but it is not quite ready for prime-time yet.
// 1. If the window is destroyed while the go routine is still running,
// we're in big trouble.
// 2. This has no way to stop from some external event (like focus).
// Basically, both of these things can be solved by using channels to tell
// the goroutine to quit. Not difficult but also not worth my time atm.
func cmd_active_flash() {
    focused := WM.focused()

    if focused != nil {
        go func(c *client) {
            for i := 0; i < 10; i++ {
                if c.Frame().State() == StateActive {
                    c.Frame().Inactive()
                } else {
                    c.Frame().Active()
                }

                time.Sleep(time.Second)
            }
        }(focused)
    }
}

