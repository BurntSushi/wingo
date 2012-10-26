// Example input shows how to create a window that reads text typed by the user
// and displays it.
//
// Note that while this boiler plate will get you up and running, the true
// complexity that makes this example work is in xgbutil/keybind. Basically,
// this is where the translation from a (modifier, keycode) tuple to a single
// character is done. It is by no means comprehensive and most definitely will
// not work in all environments (particularly with other languages).
package main

import (
	"bytes"
	"image/color"
	"log"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/misc"
	"github.com/BurntSushi/wingo/render"
	"github.com/BurntSushi/wingo/text"
)

var (
	// A near guaranteed font. If parsing fails, MustFont wll panic.
	font = xgraphics.MustFont(xgraphics.ParseFont(
		bytes.NewBuffer(misc.DataFile("DejaVuSans.ttf"))))
	fontSize = 30.0
	fontColor = render.NewImageColor(color.RGBA{0x0, 0x0, 0x0, 0xff})
	bgColor = render.NewImageColor(color.RGBA{0xff, 0xff, 0xff, 0xff})
	width = 800
	padding = 10
)

func fatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	X, err := xgbutil.NewConn()
	fatal(err)

	// Remember to always initialize the keybind package before usage.
	keybind.Initialize(X)

	// We create a benign parent window. Its only value in this example is
	// instructional. In particular, it shows how to have a sub-window get
	// focus using the ICCCM WM_TAKE_FOCUS protocol. (Since by default,
	// the window manager will assign focus to your top-level window.)
	// The real magic happens below with the WMTakeFocus method call.
	parentWin, err := xwindow.Create(X, X.RootWin())
	fatal(err)

	// NewInput creates a new input window and handles all of the text drawing
	// for you. It can be any width, but the height will be automatically
	// determined by the height extents of the font size chosen.
	input := text.NewInput(X, parentWin.Id, width, padding, font, fontSize,
		fontColor, bgColor)
	parentWin.Resize(input.Geom.Width(), input.Geom.Height())

	// Make sure X reports KeyPress events when this window is focused.
	input.Listen(xproto.EventMaskKeyPress)

	// Listen to KeyPress events. If it's a BackSpace, remove the last
	// character in the input box. If it's "Return" quit. If it's "Escape",
	// clear the input box.
	// Otherwise, try to add the key pressed to the input box.
	// (Not all key presses correspond to a single character that is added.)
	xevent.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			mods, kc := ev.State, ev.Detail
			switch {
			case keybind.KeyMatch(X, "BackSpace", mods, kc):
				input.Remove()
			case keybind.KeyMatch(X, "Return", mods, kc):
				log.Println("Return has been pressed.")
				log.Printf("The current text is: %s", string(input.Text))
				log.Println("Quitting...")
				xevent.Quit(X)
			case keybind.KeyMatch(X, "Escape", mods, kc):
				input.Reset()
			default:
				input.Add(mods, kc)
			}
		}).Connect(X, input.Id)

	// Implement the WM_DELETE_WINDOW protocol.
	parentWin.WMGracefulClose(func(w *xwindow.Window) {
		xevent.Quit(X)
	})

	// Implement the WM_TAKE_FOCUS protocol. The callback function provided
	// is executed when a valid WM_TAKE_FOCUS ClientMessage event has been
	// received from the window manager.
	// According to ICCCM Section 4.2.7, this is one of the three valid ways
	// of setting input focus to a sub-window. (It's also easiest since it
	// doesn't require us to monitor FocusChange events. EW.)
	// If you have multiple sub-windows that can be focused, this callback
	// function is where the logic would go to pick which sub-window should
	// be focused upon receipt of a WM_TAKE_FOCUS message.
	parentWin.WMTakeFocus(func(w *xwindow.Window, tstamp xproto.Timestamp) {
		input.FocusParent(tstamp)
	})

	// Map the window and start the main X event loop.
	input.Map()
	parentWin.Map()
	xevent.Main(X)
}
