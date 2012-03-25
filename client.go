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

// I originally had this Client interface because my plan was to have
// several different kinds of clients that implement this interface.
// i.e., a normal client, a dock client, a desktop client, etc.
// However, I think the similarity between each client is far too great to
// be worth paying for code duplication. (Code gets duplicated when a function
// has to add the receiver as an implementation of Client.)
// We'll have to settle for things like
// "if window type is weird { do weird stuff } else { do normal stuff }"
// for now. I may change this at some point, or I may trash the interface
// idea completely. I just don't quite know my requirements yet.
type Client interface {
    Alive() bool
    Close()
    Focus()
    Focused()
    Frame() Frame
    FrameNada()
    FrameSlim()
    FrameBorders()
    FrameFull()
    Geom() xrect.Rect
    GravitizeX(x int, gravity int) int
    GravitizeY(y int, gravity int) int
    Id() xgb.Id
    Layer() int
    Map()
    Mapped() bool
    Raise()
    SetupFocus(win xgb.Id, buttonStr string, grab bool)
    SetupMoveDrag(parent xgb.Id, buttonStr string, grab bool)
    SetupResizeDrag(parent xgb.Id, buttonStr string, grab bool,
                    direction uint32)
    String() string
    TrulyAlive() bool
    Unfocused()
    ValidateHeight(height int) int
    ValidateWidth(width int) int
    Win() *window
}

func clientMapRequest(X *xgbutil.XUtil, ev xevent.MapRequestEvent) {
    X.Grab()
    defer X.Ungrab()

    client, err := newClient(ev.Window)
    if err != nil {
        logWarning.Printf("Could not manage window %X because: %v\n",
                          ev.Window, err)
        return
    }

    client.manage()
}

type client struct {
    window *window
    layer int
    name, vname, wmname string
    isMapped bool
    initialMap bool
    lastTime int
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

func newClient(id xgb.Id) (*client, error) {
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
        name = ""
        logWarning.Printf("Could not find name for window %X.", id)
    }

    vname, err := ewmh.WmVisibleNameGet(X, id)
    if err != nil {
        vname = ""
    }
    wmname, err := icccm.WmNameGet(X, id)
    if err != nil {
        wmname = ""
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

    return &client{
        window: newWindow(id),
        layer: layer,
        name: name,
        vname: vname,
        wmname: wmname,
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
func (c *client) manage() {
    _, err := c.Win().geometry()
    if err != nil {
        logWarning.Printf("Could not manage window %X because: %s",
                          c.window.id, err)
        return
    }

    // time for reparenting/decorating
    c.frameInit()
    c.FrameFull()

    // We're committed now...

    // time to add the client to the WM state
    WM.clientAdd(c)
    WM.focusAdd(c)
    c.Raise()

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
            c.frame.ConfigureClient(int(ev.ValueMask), int(ev.X), int(ev.Y),
                                    int(ev.Width), int(ev.Height),
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

    c.clientMouseConfig()
    c.frameMouseConfig()

    // c.SetupMoveDrag(c.frame.ParentId(), "Mod4-1", true) 
    // c.SetupResizeDrag(c.frame.ParentId(), "Mod4-3", true, ewmh.Infer) 

    // If the initial state isn't iconic or is absent, then we can map
    if c.hints.Flags & icccm.HintState == 0 ||
       c.hints.InitialState != icccm.StateIconic {
        c.Map()
    }
}

func (c *client) frameInit() {
    // We want one parent window for all frames.
    parent := newParent(c)

    c.frameNada = newFrameNada(parent, c)
    c.frameSlim = newFrameSlim(parent, c)
    c.frameBorders = newFrameBorders(parent, c)
    c.frameFull = newFrameFull(parent, c)
}

func (c *client) frameSet(f Frame) {
    if f == c.Frame() { // no need to change...
        return
    }
    if c.Frame() != nil {
        c.Frame().Off()
    }
    c.frame = f
    c.Frame().On()
    FrameReset(c.Frame())
}

// SetupFocus is a useful function to setup a callback when you want a
// client to have focus. Particularly if, in the future, we want to allow
// a new focus model (like follows-mouse).
// This is not used in the 'Manage' method because we have to do some special
// stuff when attaching a button press to an actual client window.
func (c *client) SetupFocus(win xgb.Id, buttonStr string, grab bool) {
    mousebind.ButtonPressFun(
        func(X *xgbutil.XUtil, ev xevent.ButtonPressEvent) {
            c.Focus()
            c.Raise()
    }).Connect(X, win, buttonStr, false, grab)
}

// setupMoveDrag does the boiler plate for registering this client's
// "move" drag.
func (c *client) SetupMoveDrag(dragWin xgb.Id, buttonStr string, grab bool) {
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
func (c *client) SetupResizeDrag(dragWin xgb.Id, buttonStr string, grab bool,
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

func (c *client) unmanage() {
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

func (c *client) Map() {
    c.window.map_()
    c.frame.Map()
    c.Focus()
    c.isMapped = true
    c.setWmState(icccm.StateNormal)
}

func (c *client) unmapped() {
    c.setWmState(icccm.StateIconic)
    focused := WM.focused()
    c.frame.Unmap()
    c.isMapped = false

    if focused != nil && focused.Id() == c.Id() {
        WM.fallback()
    }
}

func (c *client) setWmState(state int) {
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

func (c *client) Close() {
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
                                           int(wm_del_win))
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
func (c *client) Alive() bool {
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
func (c *client) TrulyAlive() bool {
    _, err := xwindow.RawGeometry(X, c.window.id)
    if err != nil {
        return false
    }
    return true
}

func (c *client) Focus() {
    if c.hints.Flags & icccm.HintInput > 0 && c.hints.Input == 1 {
        c.window.focus()
        c.Focused()
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
                                           int(wm_take_focus),
                                           int(X.GetTime()))
        if err != nil {
            logWarning.Println(err)
            return
        }

        X.Conn().SendEvent(false, c.window.id, 0, cm.Bytes())

        c.Focused()
    }
}

func (c *client) Focused() {
    WM.focusAdd(c)
    c.Frame().StateActive()

    // Forcefully unfocus all other clients
    WM.unfocusExcept(c.Id())
}

func (c *client) Unfocused() {
    c.Frame().StateInactive()
}

func (c *client) updateProperty(ev xevent.PropertyNotifyEvent) {
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

func (c *client) GravitizeX(x int, gravity int) int {
    // Don't do anything if there's no gravity options set and we're
    // trying to infer gravity.
    // This is equivalent to setting NorthWest gravity
    if gravity > -1 && c.nhints.Flags & icccm.SizeHintPWinGravity == 0 {
        return x
    }

    // Otherwise, we're either inferring gravity (from normal hints), or
    // using some forced notion of gravity (probably from EWMH stuff)
    var g int
    if gravity > -1 {
        g = gravity
    } else {
        g = int(c.nhints.WinGravity)
    }

    f := c.Frame()
    switch {
    case g == xgb.GravityStatic || g == xgb.GravityBitForget:
        x -= f.Left()
    case g == xgb.GravityNorth || g == xgb.GravitySouth ||
         g == xgb.GravityCenter:
        x -= abs(f.Left() - f.Right()) / 2
    case g == xgb.GravityNorthEast || g == xgb.GravityEast ||
         g == xgb.GravitySouthEast:
        x -= f.Left() + f.Right()
    }

    return x
}

func (c *client) GravitizeY(y int, gravity int) int {
    // Don't do anything if there's no gravity options set and we're
    // trying to infer gravity.
    // This is equivalent to setting NorthWest gravity
    if gravity > -1 && c.nhints.Flags & icccm.SizeHintPWinGravity == 0 {
        return y
    }

    // Otherwise, we're either inferring gravity (from normal hints), or
    // using some forced notion of gravity (probably from EWMH stuff)
    var g int
    if gravity > -1 {
        g = gravity
    } else {
        g = int(c.nhints.WinGravity)
    }

    f := c.Frame()
    switch {
    case g == xgb.GravityStatic || g == xgb.GravityBitForget:
        y -= f.Top()
    case g == xgb.GravityEast || g == xgb.GravityWest ||
         g == xgb.GravityCenter:
        y -= abs(f.Top() - f.Bottom()) / 2
    case g == xgb.GravitySouthEast || g == xgb.GravitySouth ||
         g == xgb.GravitySouthWest:
        y -= f.Top() + f.Bottom()
    }

    return y
}

func (c *client) ValidateHeight(height int) int {
    return c.validateSize(height, c.nhints.HeightInc, c.nhints.BaseHeight,
                          c.nhints.MinHeight, c.nhints.MaxHeight)
}

func (c *client) ValidateWidth(width int) int {
    return c.validateSize(width, c.nhints.WidthInc, c.nhints.BaseWidth,
                          c.nhints.MinWidth, c.nhints.MaxWidth)
}

func (c *client) validateSize(size, inc, base, min, max int) int {
    if size < min && c.nhints.Flags & icccm.SizeHintPMinSize > 0 {
        return min
    }
    if size < 1 {
        return 1
    }
    if size > max && c.nhints.Flags & icccm.SizeHintPMaxSize > 0 {
        return max
    }
    if inc > 1 && c.nhints.Flags & icccm.SizeHintPResizeInc > 0 {
        var whichb int
        if base > 0 {
            whichb = base
        } else {
            whichb = min
        }
        size = whichb +
               (int(round(float64(size - whichb) / float64(inc))) * inc)
    }

    return size
}


//
// Accessors to satisfy the Client interface
//

func (c *client) Frame() Frame {
    return c.frame
}

func (c *client) FrameNada() {
    c.frameSet(c.frameNada)
}

func (c *client) FrameSlim() {
    c.frameSet(c.frameSlim)
}

func (c *client) FrameBorders() {
    c.frameSet(c.frameBorders)
}

func (c *client) FrameFull() {
    c.frameSet(c.frameFull)
}

func (c *client) Geom() xrect.Rect {
    return c.window.geom
}

func (c *client) Id() xgb.Id {
    return c.window.id
}

func (c *client) Layer() int {
    return c.layer
}

func (c *client) Mapped() bool {
    return c.isMapped
}

func (c *client) Name() string {
    if len(c.vname) > 0 {
        return c.vname
    }
    if len(c.name) > 0 {
        return c.name
    }
    if len(c.wmname) > 0 {
        return c.wmname
    }
    return "N/A"
}

func (c *client) Raise() {
    WM.stackRaise(c, true)
}

func (c *client) Win() *window {
    return c.window
}

func (c *client) String() string {
    return fmt.Sprintf("%s (%X)", c.Name(), c.window.id)
}

