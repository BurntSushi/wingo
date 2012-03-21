// A set of functions that are key-bindable
package main

func cmd_close_active() {
    focused := WM.focused()
    if focused != nil {
        focused.Close()
    }
}

func cmd_active_frame_nada() {
    focused := WM.focused()
    if focused != nil {
        focused.FrameNada()
    }
}

func cmd_active_frame_slim() {
    focused := WM.focused()
    if focused != nil {
        focused.FrameSlim()
    }
}

func cmd_active_frame_borders() {
    focused := WM.focused()
    if focused != nil {
        focused.FrameBorders()
    }
}

func cmd_active_frame_full() {
    focused := WM.focused()
    if focused != nil {
        focused.FrameFull()
    }
}

