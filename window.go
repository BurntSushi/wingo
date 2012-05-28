package main

import (
	"image"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/logger"
)

type window struct {
	id   xproto.Window
	geom xrect.Rect
}

const (
	DoX       = xproto.ConfigWindowX
	DoY       = xproto.ConfigWindowY
	DoW       = xproto.ConfigWindowWidth
	DoH       = xproto.ConfigWindowHeight
	DoBorder  = xproto.ConfigWindowBorderWidth
	DoSibling = xproto.ConfigWindowSibling
	DoStack   = xproto.ConfigWindowStackMode
)

func newWindow(id xproto.Window) *window {
	return &window{
		id:   id,
		geom: xrect.New(0, 0, 1, 1),
	}
}

func createWindow(parent xproto.Window, masks int, vals ...uint32) *window {
	wid, err := xproto.NewWindowId(X.Conn())
	if err != nil {
		logger.Error.Printf("Could not create window: %s", err)
		return nil
	}
	scrn := X.Screen()

	xproto.CreateWindow(X.Conn(), scrn.RootDepth, wid, parent, 0, 0, 1, 1, 0,
		xproto.WindowClassInputOutput, scrn.RootVisual,
		uint32(masks), vals)

	return newWindow(wid)
}

func createImageWindow(parent xproto.Window, img image.Image,
	masks int, vals ...uint32) *window {
	newWin := createWindow(parent, masks, vals...)

	width, height := xgraphics.GetDim(img)
	newWin.moveresize(DoW|DoH, 0, 0, width, height)

	xgraphics.PaintImg(X, newWin.id, img)

	return newWin
}

func (w *window) listen(masks int) {
	xproto.ChangeWindowAttributes(X.Conn(), w.id,
		xproto.CwEventMask, []uint32{uint32(masks)})
}

func (w *window) map_() {
	xproto.MapWindow(X.Conn(), w.id)
}

func (w *window) unmap() {
	xproto.UnmapWindow(X.Conn(), w.id)
}

func (w *window) change(masks int, vals ...uint32) {
	xproto.ChangeWindowAttributes(X.Conn(), w.id, uint32(masks), vals)
}

func (w *window) clear() {
	xproto.ClearArea(X.Conn(), false, w.id, 0, 0, 0, 0)
}

func (w *window) geometry() (xrect.Rect, error) {
	var err error
	w.geom, err = xwindow.RawGeometry(X, xproto.Drawable(w.id))
	if err != nil {
		return nil, err
	}
	return w.geom, nil
}

func (w *window) kill() {
	xproto.KillClient(X.Conn(), uint32(w.id))
}

func (w *window) destroy() {
	xproto.DestroyWindow(X.Conn(), w.id)
	xevent.Detach(X, w.id)
}

func (w *window) focus() {
	xproto.SetInputFocus(X.Conn(), xproto.InputFocusPointerRoot, w.id, 0)
}

// moveresize is a wrapper around configure that only accepts parameters
// related to size and position.
func (win *window) moveresize(flags, x, y, w, h int) {
	// Kill any hopes of stacking
	flags = (flags & ^DoSibling) & ^DoStack
	win.configure(flags, x, y, w, h, 0, 0)
}

// configure is the method version of 'configure'.
// It is duplicated because we need to update our idea of the window's
// geometry. (We don't want another set of 'if' statements because it
// needs to be as efficient as possible.)
func (win *window) configure(flags, x, y, w, h int,
	sibling xproto.Window, stackMode byte) {

	vals := []uint32{}

	if DoX&flags > 0 {
		vals = append(vals, uint32(x))
		win.geom.XSet(x)
	}
	if DoY&flags > 0 {
		vals = append(vals, uint32(y))
		win.geom.YSet(y)
	}
	if DoW&flags > 0 {
		if int16(w) <= 0 {
			w = 1
		}
		vals = append(vals, uint32(w))
		win.geom.WidthSet(w)
	}
	if DoH&flags > 0 {
		if int16(h) <= 0 {
			h = 1
		}
		vals = append(vals, uint32(h))
		win.geom.HeightSet(h)
	}
	if DoSibling&flags > 0 {
		vals = append(vals, uint32(sibling))
	}
	if DoStack&flags > 0 {
		vals = append(vals, uint32(stackMode))
	}

	// Don't send anything if we have nothing to send
	if len(vals) == 0 {
		return
	}

	xproto.ConfigureWindow(X.Conn(), win.id, uint16(flags), vals)
}

// configure is a nice wrapper around ConfigureWindow.
// We purposefully omit 'BorderWidth' because I don't think it's ever used
// any more.
func configure(window xproto.Window, flags, x, y, w, h int,
	sibling xproto.Window, stackMode byte) {

	vals := []uint32{}

	if DoX&flags > 0 {
		vals = append(vals, uint32(x))
	}
	if DoY&flags > 0 {
		vals = append(vals, uint32(y))
	}
	if DoW&flags > 0 {
		if int16(w) <= 0 {
			w = 1
		}
		vals = append(vals, uint32(w))
	}
	if DoH&flags > 0 {
		if int16(h) <= 0 {
			h = 1
		}
		vals = append(vals, uint32(h))
	}
	if DoSibling&flags > 0 {
		vals = append(vals, uint32(sibling))
	}
	if DoStack&flags > 0 {
		vals = append(vals, uint32(stackMode))
	}

	xproto.ConfigureWindow(X.Conn(), window, uint16(flags), vals)
}

// configureRequest responds to generic configure requests from windows that
// we don't manage.
func configureRequest(X *xgbutil.XUtil, ev xevent.ConfigureRequestEvent) {
	configure(ev.Window, int(ev.ValueMask) & ^int(DoStack) & ^int(DoSibling),
		int(ev.X), int(ev.Y), int(ev.Width), int(ev.Height),
		ev.Sibling, ev.StackMode)
}
