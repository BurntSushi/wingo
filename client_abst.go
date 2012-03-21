package main

import (
    "fmt"
)

import "code.google.com/p/jamslam-x-go-binding/xgb"

import (
    "github.com/BurntSushi/xgbutil"
    "github.com/BurntSushi/xgbutil/ewmh"
    "github.com/BurntSushi/xgbutil/icccm"
    "github.com/BurntSushi/xgbutil/mousebind"
    "github.com/BurntSushi/xgbutil/xevent"
    "github.com/BurntSushi/xgbutil/xprop"
    "github.com/BurntSushi/xgbutil/xrect"
    "github.com/BurntSushi/xgbutil/xwindow"
)

// An "abstClient" is a type that is never directly used. (i.e., abstract)
// It is only embedded. It provides a common set of methods and attributes
// to all clients. Some of its methods may be "overridden".
type abstClient struct {
    window *window
    layer int
    name string
    vname string
    isMapped bool
    initialMap bool
    lastTime uint32
    unmapIgnore int
    hints *icccm.Hints
    nhints *icccm.NormalHints
    protocols []string

    frame Frame
    frameNada *frameNada
    frameSlim *frameSlim
    frameBorders *frameBorders
    frameFull *frameFull
}

func newAbstractClient(id xgb.Id) (*abstClient, error) {
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

    nhints, err := icccm.WmNormalHintsGet(X, id)
    if err != nil {
        logWarning.Println(err)
        logMessage.Printf("Using reasonable defaults for WM_NORMAL_HINTS " +
                          "for %X", id)
        nhints = icccm.NormalHints{}
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

    vname, err := ewmh.WmVisibleNameGet(X, id)
    if err != nil {
        vname = ""
        logWarning.Printf("Could not find visible name for window %X, " +
                          "using 'N/A'.", id)
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

    return &abstClient{
        window: newWindow(id),
        layer: layer,
        name: name,
        vname: vname,
        isMapped: false,
        initialMap: false,
        lastTime: 0,
        unmapIgnore: 0,
        hints: &hints,
        nhints: &nhints,
        protocols: protocols,

        frame: nil,
        frameNada: nil,
        frameSlim: nil,
    }, nil
}

// manage sets everything up to bring a client window into window management.
// It is still possible for us to bail.
func (c *abstClient) manage() {
    _, err := c.Win().geometry()
    if err != nil {
        logWarning.Printf("Could not manage window %X because: %s",
                          c.window.id, err)
        return
    }

    // time for reparenting/decorating
    c.frameInit()
    c.FrameSlim()

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
            // Don't honor configure requests when we're moving or resizing
            if c.frame.Moving() || c.frame.Resizing() {
                return
            }
            c.frame.ConfigureClient(ev.ValueMask, ev.X, ev.Y,
                                    ev.Width, ev.Height,
                                    ev.Sibling, ev.StackMode, false)
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

    c.setupMoveDrag(c.frame.ParentId(), "Mod4-1")
    c.setupResizeDrag(c.frame.ParentId(), "Mod4-3", ewmh.Infer)

    // If the initial state isn't iconic or is absent, then we can map
    if c.hints.Flags & icccm.HintState == 0 ||
       c.hints.InitialState != icccm.StateIconic {
        c.Map()
    }
}

func (c *abstClient) frameInit() {
    // We want one parent window for all frames.
    parent := newParent(c)

    c.frameNada = newFrameNada(parent, c)
    c.frameSlim = newFrameSlim(parent, c)
    c.frameBorders = newFrameBorders(parent, c)
    c.frameFull = newFrameFull(parent, c)
}

func (c *abstClient) frameSet(f Frame) {
    if f == c.Frame() { // no need to change...
        return
    }
    if c.Frame() != nil {
        c.Frame().Off()
    }
    c.frame = f
    c.Frame().On()
    c.Frame().Reset()
}

// setupMoveDrag does the boiler plate for registering this client's
// "move" drag.
func (c *abstClient) setupMoveDrag(dragWin xgb.Id, buttonStr string) {
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
func (c *abstClient) setupResizeDrag(dragWin xgb.Id, buttonStr string,
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

func (c *abstClient) unmanage() {
    if c.isMapped {
        c.unmapped()
    }

    c.frame.Destroy()
    c.setWmState(icccm.StateWithdrawn)
    xevent.Detach(X, c.window.id)
    WM.stackRemove(c)
    WM.focusRemove(c)
    WM.clientRemove(c)

    WM.updateEwmhStacking()
}

func (c *abstClient) Map() {
    c.window.map_()
    c.frame.Map()
    c.Focus()
    c.isMapped = true
    c.setWmState(icccm.StateNormal)
}

func (c *abstClient) unmapped() {
    c.setWmState(icccm.StateIconic)
    focused := WM.focused()
    c.frame.Unmap()
    c.isMapped = false

    if focused != nil && focused.Id() == c.Id() {
        WM.fallback()
    }
}

func (c *abstClient) setWmState(state uint32) {
    if !c.TrulyAlive() {
        return
    }

    err := icccm.WmStateSet(X, c.window.id, icccm.WmState{State: state})
    if err != nil {
        var stateStr string
        switch state {
        case icccm.StateNormal: stateStr = "Normal"
        case icccm.StateIconic: stateStr = "Iconic"
        case icccm.StateWithdrawn: stateStr = "Withdrawn"
        default: stateStr = "Unknown"
        }
        logWarning.Printf("Could not set window state to %s on %s " +
                          "because: %v", stateStr, c, err)
    }
}

func (c *abstClient) Close() {
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

// Alive retrieves all X events up until the point of calling that have been
// sent. It then peeks at those events to see if there is an UnmapNotify
// for client c. If there is one, and if the 'unmapIgnore' at 0, then this
// client is marked for deletion and should be considered dead.
// (unmapIgnore is incremented whenever Wingo unmaps a window. When Wingo
// unmaps a window, we *don't* want to delete it, just hide it.)
func (c *abstClient) Alive() bool {
    X.Flush() // fills up the XGB event queue with ready events
    xevent.Read(X, false) // fills up the xgbutil event queue without blocking

    // we only consider a client marked for deletion when 'ignore' reaches 0
    ignore := c.unmapIgnore
    for _, ev := range X.QueuePeek() {
        if unmap, ok := ev.(xgb.UnmapNotifyEvent);
           ok && unmap.Window == c.Win().id {
            if ignore <= 0 {
                return false
            }
            ignore -= 1
        }
    }
    return true
}

// TrulyAlive is useful in scenarios when Alive doesn't help.
// Namely, when we know the window has been unmapped but are not sure
// if it is still an X resource.
func (c *abstClient) TrulyAlive() bool {
    _, err := xwindow.RawGeometry(X, c.window.id)
    if err != nil {
        return false
    }
    return true
}

func (c *abstClient) Focus() {
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

func (c *abstClient) focused() {
    WM.focusAdd(c)
}

func (c *abstClient) updateProperty(ev xevent.PropertyNotifyEvent) {
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
    case "_NET_WM_VISIBLE_NAME":
        newName, err := ewmh.WmVisibleNameGet(X, c.window.id)
        showVals(c.vname, newName)
        if err == nil {
            c.vname = newName
        }
    case "_NET_WM_USER_TIME":
        newTime, err := ewmh.WmUserTimeGet(X, c.window.id)
        showVals(c.lastTime, newTime)
        if err == nil {
            c.lastTime = newTime
        }
    }
}

func (c *abstClient) GravitizeX(x int16) int16 {
    return x
}

func (c *abstClient) GravitizeY(y int16) int16 {
    return y
}

func (c *abstClient) ValidateHeight(height uint16) uint16 {
    return c.validateSize(height,
                          uint16(c.nhints.HeightInc),
                          uint16(c.nhints.BaseHeight),
                          uint16(c.nhints.MinHeight),
                          uint16(c.nhints.MaxHeight))
}

func (c *abstClient) ValidateWidth(width uint16) uint16 {
    return c.validateSize(width,
                          uint16(c.nhints.WidthInc),
                          uint16(c.nhints.BaseWidth),
                          uint16(c.nhints.MinWidth),
                          uint16(c.nhints.MaxWidth))
}

func (c *abstClient) validateSize(size uint16, inc, base,
                                  min, max uint16) uint16 {
    if int16(size) < int16(min) && c.nhints.Flags & icccm.SizeHintPMinSize > 0 {
        return min
    }
    if int16(size) < 1 {
        return 1
    }
    if size > max && c.nhints.Flags & icccm.SizeHintPMaxSize > 0 {
        return max
    }
    if inc > 1 && c.nhints.Flags & icccm.SizeHintPResizeInc > 0 {
        var whichb uint16
        if base > 0 {
            whichb = base
        } else {
            whichb = min
        }
        size = whichb +
               (uint16(round(float64(size - whichb) / float64(inc))) * inc)
    }

    return size
}


//
// Accessors to satisfy the Client interface
//

func (c *abstClient) Frame() Frame {
    return c.frame
}

func (c *abstClient) FrameNada() {
    c.frameSet(c.frameNada)
}

func (c *abstClient) FrameSlim() {
    c.frameSet(c.frameSlim)
}

func (c *abstClient) FrameBorders() {
    c.frameSet(c.frameBorders)
}

func (c *abstClient) FrameFull() {
    c.frameSet(c.frameFull)
}

func (c *abstClient) Geom() xrect.Rect {
    return c.window.geom
}

func (c *abstClient) Id() xgb.Id {
    return c.window.id
}

func (c *abstClient) Layer() int {
    return c.layer
}

func (c *abstClient) Mapped() bool {
    return c.isMapped
}

func (c *abstClient) Name() string {
    if len(c.vname) > 0 {
        return c.vname
    }
    return c.name
}

func (c *abstClient) Win() *window {
    return c.window
}

func (c *abstClient) String() string {
    return fmt.Sprintf("%s (%X)", c.Name(), c.window.id)
}

