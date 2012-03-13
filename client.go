package main

import (
    "log"
)

import "code.google.com/p/jamslam-x-go-binding/xgb"

import (
    "github.com/BurntSushi/xgbutil"
    "github.com/BurntSushi/xgbutil/icccm"
    "github.com/BurntSushi/xgbutil/xevent"
    "github.com/BurntSushi/xgbutil/xprop"
    // "github.com/BurntSushi/xgbutil/xwindow" 
)

type client struct {
    window *window

    hints *icccm.Hints
    protocols []string
}

func clientMapRequest(X *xgbutil.XUtil, ev xevent.MapRequestEvent) {
    client, err := newClient(ev.Window)
    if err != nil {
        log.Printf("Could not manage window %X because: %v\n", ev.Window, err)
        return
    }

    client.manage()
}

func newClient(id xgb.Id) (*client, error) {
    hints, err := icccm.WmHintsGet(X, id)
    if err != nil {
        return nil, err
    }

    protocols, err := icccm.WmProtocolsGet(X, id)
    if err != nil {
        return nil, err
    }

    return &client{
        window: newWindow(id),
        hints: &hints,
        protocols: protocols,
    }, nil
}

func (c *client) manage() {
    // If the initial state isn't iconic or is absent, then we can map
    if c.hints.Flags & icccm.HintState == 0 ||
       c.hints.InitialState != icccm.StateIconic {
        c.map_()
    }
}

func (c *client) map_() {
    c.window.map_()
    c.focus()
}

func (c *client) focus() {
    if c.hints.Flags & icccm.HintInput > 0 && c.hints.Input == 1 {
        c.window.focus()
    } else {
        for _, a := range c.protocols {
            if a == "WM_TAKE_FOCUS" {
                wm_protocols, err := xprop.Atm(X, "WM_PROTOCOLS")
                if err != nil {
                    log.Println(err)
                    break
                }

                wm_take_focus, err := xprop.Atm(X, "WM_TAKE_FOCUS")
                if err != nil {
                    log.Println(err)
                    break
                }

                cm, err := xevent.NewClientMessage(32, c.window.id,
                                                   wm_protocols,
                                                   uint32(wm_take_focus),
                                                   uint32(X.GetTime()))
                if err != nil {
                    log.Println(err)
                    break
                }

                X.Conn().SendEvent(false, c.window.id, 0, cm.Bytes())
            }
        }
    }
}

