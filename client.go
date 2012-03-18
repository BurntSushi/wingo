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
    "github.com/BurntSushi/xgbutil/mousebind"
    "github.com/BurntSushi/xgbutil/xevent"
    "github.com/BurntSushi/xgbutil/xprop"
    "github.com/BurntSushi/xgbutil/xwindow"
)

type Client interface {
    Alive() bool
    Close()
    Focus()
    Frame() Frame
    Id() xgb.Id
    Layer() int
    Manage()
    Map()
    Mapped() bool
    String() string
    Win() *window
}

// An "abstractClient" is a type that is never directly used.
// It is only embedded. It provides a common set of methods and attributes
// to all clients.
type abstractClient struct {
    window *window
    frame Frame
    layer int
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

    client.Manage()
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
        logWarning.Println(err)
        logMessage.Printf("Using reasonable defaults for WM_HINTS for %X", id)
        hints = icccm.Hints{
            Flags: icccm.HintInput | icccm.HintState,
            Input: 1,
            InitialState: icccm.StateNormal,
        }
    }

    protocols, err := icccm.WmProtocolsGet(X, id)
    if err != nil {
        logWarning.Printf("Window %X does not have WM_PROTOCOLS set.", id)
        protocols = []string{}
    }

    name, err := ewmh.WmNameGet(X, id)
    if err != nil {
        name = "N/A"
        logWarning.Printf("Could not find name for window %X, using 'N/A'.", id)
    }

    wintypes, err := ewmh.WmWindowTypeGet(X, id)
    layer := layerDefault
    if err != nil {
        logWarning.Printf("Could not find window type for window %X, " +
                          "using 'normal'.", id)
    } else {
        if strIndex("_NET_WM_WINDOW_TYPE_DIALOG", wintypes) > -1 {
            layer = layerAbove
        }
    }

    return &abstractClient{
        window: newWindow(id),
        frame: nil,
        layer: layer,
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
func (c *abstractClient) Manage() {
    // time for reparenting
    var err error
    c.frame, err = newFrameNada(c)
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
    WM.stackRaise(c, false)

    c.window.listen(xgb.EventMaskPropertyChange |
                    xgb.EventMaskStructureNotify)

    // attach some event handlers
    xevent.PropertyNotifyFun(
        func(X *xgbutil.XUtil, ev xevent.PropertyNotifyEvent) {
            c.updateProperty(ev)
    }).Connect(X, c.window.id)
    xevent.ConfigureRequestFun(
        func(X *xgbutil.XUtil, ev xevent.ConfigureRequestEvent) {
            c.frame.Configure(ev.ValueMask, ev.X, ev.Y, ev.Width, ev.Height,
                              ev.Sibling, ev.StackMode)
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
    mousebind.ButtonPressFun(
        func(X *xgbutil.XUtil, ev xevent.ButtonPressEvent) {
            c.Focus()
            WM.stackRaise(c, true)
            xevent.ReplayPointer(X)
    }).Connect(X, c.window.id, "1", true, true)

    c.setupMoveDrag(c.frame.Parent().window.id, "Mod4-1")
    c.setupResizeDrag(c.frame.Parent().window.id, "Mod4-3", ewmh.Infer)

    // If the initial state isn't iconic or is absent, then we can map
    if c.hints.Flags & icccm.HintState == 0 ||
       c.hints.InitialState != icccm.StateIconic {
        c.Map()
    }
}

// setupMoveDrag does the boiler plate for registering this client's
// "move" drag.
func (c *abstractClient) setupMoveDrag(dragWin xgb.Id, buttonStr string) {
    dStart := xgbutil.MouseDragBeginFun(
        func(X *xgbutil.XUtil, rx, ry, ex, ey int16) (bool, xgb.Id) {
            c.frame.moveBegin(rx, ry, ex, ey)
            return true, cursorFleur
    })
    dStep := xgbutil.MouseDragFun(
        func(X *xgbutil.XUtil, rx, ry, ex, ey int16) {
            c.frame.moveStep(rx, ry, ex, ey)
    })
    dEnd := xgbutil.MouseDragFun(
        func(X *xgbutil.XUtil, rx, ry, ex, ey int16) {
            c.frame.moveEnd(rx, ry, ex, ey)
    })
    mousebind.Drag(X, dragWin, buttonStr, dStart, dStep, dEnd)
}

// setupResizeDrag does the boiler plate for registering this client's
// "resize" drag.
func (c *abstractClient) setupResizeDrag(dragWin xgb.Id, buttonStr string,
                                         direction uint32) {
    dStart := xgbutil.MouseDragBeginFun(
        func(X *xgbutil.XUtil, rx, ry, ex, ey int16) (bool, xgb.Id) {
            return c.frame.resizeBegin(direction, rx, ry, ex, ey)
    })
    dStep := xgbutil.MouseDragFun(
        func(X *xgbutil.XUtil, rx, ry, ex, ey int16) {
            c.frame.resizeStep(rx, ry, ex, ey)
    })
    dEnd := xgbutil.MouseDragFun(
        func(X *xgbutil.XUtil, rx, ry, ex, ey int16) {
            c.frame.resizeEnd(rx, ry, ex, ey)
    })
    mousebind.Drag(X, dragWin, buttonStr, dStart, dStep, dEnd)
}

func (c *abstractClient) unmanage() {
    if c.isMapped {
        c.unmapped()
    }

    c.frame.Destroy()
    xevent.Detach(X, c.window.id)
    WM.stackRemove(c)
    WM.focusRemove(c)
    WM.clientRemove(c)

    WM.updateEwmhStacking()
}

func (c *abstractClient) Map() {
    c.window.map_()
    c.frame.Map()
    c.Focus()
    c.isMapped = true
}

func (c *abstractClient) unmapped() {
    focused := WM.focused()
    c.frame.Unmap()
    c.isMapped = false

    if focused != nil && focused.Id() == c.Id() {
        WM.fallback()
    }
}

func (c *abstractClient) Close() {
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

func (c *abstractClient) Alive() bool {
    _, err := xwindow.RawGeometry(X, c.window.id)
    if err != nil {
        return false
    }
    return true
}

func (c *abstractClient) Focus() {
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
    WM.focusAdd(c)
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

func (c *abstractClient) Frame() Frame {
    return c.frame
}

func (c *abstractClient) Id() xgb.Id {
    return c.window.id
}

func (c *abstractClient) Layer() int {
    return c.layer
}

func (c *abstractClient) Mapped() bool {
    return c.isMapped
}

func (c *abstractClient) Win() *window {
    return c.window
}

func (c *abstractClient) String() string {
    return fmt.Sprintf("%s (%X)", c.name, c.window.id)
}

