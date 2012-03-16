package main

import (
    "fmt"
    // "time" 
)

import "code.google.com/p/jamslam-x-go-binding/xgb"

import (
    "github.com/BurntSushi/xgbutil"
    "github.com/BurntSushi/xgbutil/ewmh"
    "github.com/BurntSushi/xgbutil/icccm"
    "github.com/BurntSushi/xgbutil/xevent"
    "github.com/BurntSushi/xgbutil/xprop"
    // "github.com/BurntSushi/xgbutil/xwindow" 
)

type client interface {
    close_()
    focus()
    frame() frame
    id() xgb.Id
    manage()
    map_()
    mapped() bool

    String() string
}

// An "abstractClient" is a type that is never directly used.
// It is only embedded. It provides a common set of methods and attributes
// to all clients.
type abstractClient struct {
    window *window
    frm frame
    name string
    isMapped bool
    initialMap bool
    lastTime uint32
    unmapIgnore int
    hints *icccm.Hints
    protocols []string
}

type normalClient struct {
    *abstractClient
}

func clientMapRequest(X *xgbutil.XUtil, ev xevent.MapRequestEvent) {
    X.Grab()
    defer X.Ungrab()

    client, err := newNormalClient(ev.Window)
    if err != nil {
        logWarning.Printf("Could not manage window %X because: %v\n",
                          ev.Window, err)
        return
    }

    client.manage()
}

func newNormalClient(id xgb.Id) (*normalClient, error) {
    absCli, err := newAbstractClient(id)
    if err != nil {
        return nil, err
    }

    return &normalClient{
        abstractClient: absCli,
    }, nil
}

func newAbstractClient(id xgb.Id) (*abstractClient, error) {
    hints, err := icccm.WmHintsGet(X, id)
    if err != nil {
        return nil, err
    }

    protocols, err := icccm.WmProtocolsGet(X, id)
    if err != nil {
        return nil, err
    }

    name, err := ewmh.WmNameGet(X, id)
    if err != nil {
        name = "N/A"
        logWarning.Printf("Could not find name for window %X, using 'N/A'.",
                          id)
    }

    return &abstractClient{
        window: newWindow(id),
        frm: nil,
        name: name,
        isMapped: false,
        initialMap: false,
        lastTime: 0,
        unmapIgnore: 0,
        hints: &hints,
        protocols: protocols,
    }, nil
}

// manage sets everything up to bring a client window into window management.
// It is still possible for us to bail.
func (c *abstractClient) manage() {
    // time for reparenting
    var err error
    c.frm, err = newFrameNada(c.window)
    if err != nil {
        logWarning.Printf("Could not manage window %X because we could not " +
                          "get its geometry. The reason given: %s",
                          c.window.id, err)
        return
    }

    // We're committed now...

    // time to add the client to the WM state
    WM.clientAdd(c)
    WM.focusAdd(c)

    c.window.listen(xgb.EventMaskPropertyChange |
                    xgb.EventMaskStructureNotify)

    // attach some event handlers
    xevent.PropertyNotifyFun(
        func(X *xgbutil.XUtil, ev xevent.PropertyNotifyEvent) {
            c.updateProperty(ev)
    }).Connect(X, c.window.id)
    xevent.UnmapNotifyFun(
        func(X *xgbutil.XUtil, ev xevent.UnmapNotifyEvent) {
            if !c.isMapped {
                return
            }

            if c.unmapIgnore > 0 {
                c.unmapIgnore -= 1
                return
            }

            c.unmapped()
            c.unmanage()
    }).Connect(X, c.window.id)
    xevent.DestroyNotifyFun(
        func(X *xgbutil.XUtil, ev xevent.DestroyNotifyEvent) {
            c.unmanage()
    }).Connect(X, c.window.id)

    // If the initial state isn't iconic or is absent, then we can map
    if c.hints.Flags & icccm.HintState == 0 ||
       c.hints.InitialState != icccm.StateIconic {
        c.map_()
    }
}

func (c *abstractClient) unmanage() {
    if c.isMapped {
        c.unmapped()
    }

    c.frm.destroy()
    WM.focusRemove(c)
    xevent.Detach(X, c.window.id)
    WM.clientRemove(c)
}

func (c *abstractClient) map_() {
    c.window.map_()
    c.frm.map_()
    c.focus()
    c.isMapped = true
}

func (c *abstractClient) unmapped() {
    focused := WM.focused()
    c.frm.unmap()
    c.isMapped = false

    if focused != nil && focused.id() == c.id() {
        WM.fallback()
    }
}

func (c *abstractClient) close_() {
    if strIndex("WM_DELETE_WINDOW", c.protocols) > -1 {
        wm_protocols, err := xprop.Atm(X, "WM_PROTOCOLS")
        if err != nil {
            logWarning.Println(err)
            return
        }

        wm_del_win, err := xprop.Atm(X, "WM_DELETE_WINDOW")
        if err != nil {
            logWarning.Println(err)
            return
        }

        cm, err := xevent.NewClientMessage(32, c.window.id, wm_protocols,
                                           uint32(wm_del_win))
        if err != nil {
            logWarning.Println(err)
            return
        }

        X.Conn().SendEvent(false, c.window.id, 0, cm.Bytes())
    } else {
        c.window.kill()
    }

    c.unmanage()
}

func (c *abstractClient) focus() {
    if c.hints.Flags & icccm.HintInput > 0 && c.hints.Input == 1 {
        c.window.focus()
        c.focused()
    } else if strIndex("WM_TAKE_FOCUS", c.protocols) > -1 {
        wm_protocols, err := xprop.Atm(X, "WM_PROTOCOLS")
        if err != nil {
            logWarning.Println(err)
            return
        }

        wm_take_focus, err := xprop.Atm(X, "WM_TAKE_FOCUS")
        if err != nil {
            logWarning.Println(err)
            return
        }

        cm, err := xevent.NewClientMessage(32, c.window.id,
                                           wm_protocols,
                                           uint32(wm_take_focus),
                                           uint32(X.GetTime()))
        if err != nil {
            logWarning.Println(err)
            return
        }

        X.Conn().SendEvent(false, c.window.id, 0, cm.Bytes())

        c.focused()
    }
}

func (c *abstractClient) focused() {
    // focusAbove
}

func (c *abstractClient) updateProperty(ev xevent.PropertyNotifyEvent) {
    name, err := xprop.AtomName(X, ev.Atom)
    if err != nil {
        logWarning.Println("Could not get property atom name for", ev.Atom)
        return
    }

    logLots.Printf("Updating property %s with state %v on window %s",
                   name, ev.State, c)

    // helper function to log property vals
    showVals := func(o, n interface{}) {
        logLots.Printf("\tOld value: '%s', new value: '%s'", o, n)
    }

    // Start the arduous process of updating properties...
    switch name {
    case "_NET_WM_NAME":
        newName, err := ewmh.WmNameGet(X, c.window.id)
        showVals(c.name, newName)
        if err == nil {
            c.name = newName
        }
    case "_NET_WM_USER_TIME":
        newTime, err := ewmh.WmUserTimeGet(X, c.window.id)
        showVals(c.lastTime, newTime)
        if err == nil {
            c.lastTime = newTime
        }
    }
}

func (c *abstractClient) frame() frame {
    return c.frm
}

func (c *abstractClient) id() xgb.Id {
    return c.window.id
}

func (c *abstractClient) mapped() bool {
    return c.isMapped
}

func (c *abstractClient) String() string {
    return fmt.Sprintf("%s (%X)", c.name, c.window.id)
}

