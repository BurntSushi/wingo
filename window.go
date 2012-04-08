package main

import "image"

import "code.google.com/p/jamslam-x-go-binding/xgb"

import (
    "github.com/BurntSushi/xgbutil"
    "github.com/BurntSushi/xgbutil/xevent"
    "github.com/BurntSushi/xgbutil/xgraphics"
    "github.com/BurntSushi/xgbutil/xrect"
    "github.com/BurntSushi/xgbutil/xwindow"
)

type window struct {
    id xgb.Id
    geom xrect.Rect
}

const (
    DoX = xgb.ConfigWindowX
    DoY = xgb.ConfigWindowY
    DoW = xgb.ConfigWindowWidth
    DoH = xgb.ConfigWindowHeight
    DoBorder = xgb.ConfigWindowBorderWidth
    DoSibling = xgb.ConfigWindowSibling
    DoStack = xgb.ConfigWindowStackMode
)

func newWindow(id xgb.Id) *window {
    return &window{
        id: id,
        geom: xrect.Make(0, 0, 1, 1),
    }
}

func createWindow(parent xgb.Id, masks int, vals... uint32) *window {
    wid := X.Conn().NewId()
    scrn := X.Screen()

    X.Conn().CreateWindow(scrn.RootDepth, wid, parent, 0, 0, 1, 1, 0,
                          xgb.WindowClassInputOutput, scrn.RootVisual,
                          uint32(masks), vals)

    return newWindow(wid)
}

func createImageWindow(parent xgb.Id, img image.Image,
                       masks int, vals... uint32) *window {
    newWin := createWindow(parent, masks, vals...)

    width, height := xgraphics.GetDim(img)
    newWin.moveresize(DoW | DoH, 0, 0, width, height)

    xgraphics.PaintImg(X, newWin.id, img)

    return newWin
}

func (w *window) listen(masks int) {
    xwindow.Listen(X, w.id, masks)
}

func (w *window) map_() {
    X.Conn().MapWindow(w.id)
}

func (w *window) unmap() {
    X.Conn().UnmapWindow(w.id)
}

func (w *window) change(masks int, vals... uint32) {
    X.Conn().ChangeWindowAttributes(w.id, uint32(masks), vals)
}

func (w *window) clear() {
    X.Conn().ClearArea(false, w.id, 0, 0, 0, 0)
}

func (w *window) geometry() (xrect.Rect, error) {
    var err error
    w.geom, err = xwindow.RawGeometry(X, w.id)
    if err != nil {
        return nil, err
    }
    return w.geom, nil
}

func (w *window) kill() {
    X.Conn().KillClient(uint32(w.id))
}

func (w *window) destroy() {
    X.Conn().DestroyWindow(w.id)
    xevent.Detach(X, w.id)
}

func (w *window) focus() {
    X.Conn().SetInputFocus(xgb.InputFocusPointerRoot, w.id, 0)
}

// moveresize is a wrapper around configure that only accepts parameters
// related to size and position.
func (win *window) moveresize(flags, x, y, w, h int) {
    // Kill any hopes of stacking
    flags = (flags & ^DoSibling) & ^DoStack
    win.configure(flags, x, y, w, h, xgb.Id(0), 0)
}

// configure is the method version of 'configure'.
// It is duplicated because we need to update our idea of the window's
// geometry. (We don't want another set of 'if' statements because it
// needs to be as efficient as possible.)
func (win *window) configure(flags, x, y, w, h int,
                             sibling xgb.Id, stackMode byte) {
    vals := []uint32{}

    if DoX & flags > 0 {
        vals = append(vals, uint32(x))
        win.geom.XSet(x)
    }
    if DoY & flags > 0 {
        vals = append(vals, uint32(y))
        win.geom.YSet(y)
    }
    if DoW & flags > 0 {
        if int16(w) <= 0 {
            w = 1
        }
        vals = append(vals, uint32(w))
        win.geom.WidthSet(w)
    }
    if DoH & flags > 0 {
        if int16(h) <= 0 {
            h = 1
        }
        vals = append(vals, uint32(h))
        win.geom.HeightSet(h)
    }
    if DoSibling & flags > 0 {
        vals = append(vals, uint32(sibling))
    }
    if DoStack & flags > 0 {
        vals = append(vals, uint32(stackMode))
    }

    // Don't send anything if we have nothing to send
    if len(vals) == 0 {
        return
    }

    X.Conn().ConfigureWindow(win.id, uint16(flags), vals)
}

// configure is a nice wrapper around ConfigureWindow.
// We purposefully omit 'BorderWidth' because I don't think it's ever used
// any more.
func configure(window xgb.Id, flags, x, y, w, h int,
               sibling xgb.Id, stackMode byte) {
    vals := []uint32{}

    if DoX & flags > 0 {
        vals = append(vals, uint32(x))
    }
    if DoY & flags > 0 {
        vals = append(vals, uint32(y))
    }
    if DoW & flags > 0 {
        if int16(w) <= 0 {
            w = 1
        }
        vals = append(vals, uint32(w))
    }
    if DoH & flags > 0 {
        if int16(h) <= 0 {
            h = 1
        }
        vals = append(vals, uint32(h))
    }
    if DoSibling & flags > 0 {
        vals = append(vals, uint32(sibling))
    }
    if DoStack & flags > 0 {
        vals = append(vals, uint32(stackMode))
    }

    X.Conn().ConfigureWindow(window, uint16(flags), vals)
}

// configureRequest responds to generic configure requests from windows that
// we don't manage.
func configureRequest(X *xgbutil.XUtil, ev xevent.ConfigureRequestEvent) {
    configure(ev.Window, int(ev.ValueMask) & ^int(DoStack) & ^int(DoSibling),
              int(ev.X), int(ev.Y), int(ev.Width), int(ev.Height),
              ev.Sibling, ev.StackMode)
}

