// Example cycle shows how to use the cycle prompt. It is by no means
// comprehensive and should not be considered functional. Namely, it polls
// for state once and never updates itself. The example is intended to
// half-heartedly emulate an alt-tab window cycling dialog.
//
// A proper usage of alt-tab would require listening for changes on the
// _NET_CLIENT_STACKING_LIST property and listening for changes on each
// individual window (like its name, icon and existence itself).
package main

import (
	"log"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xinerama"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo-conc/prompt"
)

var (
	// The key combinations to use to cycle forwards and backwards in the
	// cycle prompt.
	cyclePrev, cycleNext = "Mod4-Shift-tab", "Mod4-tab"
)

type window struct {
	X *xgbutil.XUtil
	id xproto.Window
	mapped bool
}

func newWindow(X *xgbutil.XUtil, parent, id xproto.Window) *window {
	w := &window{
		X: X,
		id: id,
	}
	w.setMapped()
	return w
}

// Using EWMH, ask the window manager whether the window is mapped or not.
func (w *window) setMapped() {
	states, err := ewmh.WmStateGet(w.X, w.id)
	fatal(err)
	for _, state := range states {
		if state == "_NET_WM_STATE_HIDDEN" {
			w.mapped = false
			return
		}
	}
	w.mapped = true
}

func (w *window) CycleIsActive() bool {
	return w.mapped
}

func (w *window) CycleImage() *xgraphics.Image {
	ximg, err := xgraphics.FindIcon(w.X, w.id,
		prompt.DefaultCycleTheme.IconSize, prompt.DefaultCycleTheme.IconSize)
	fatal(err)

	return ximg
}

func (w *window) CycleText() string {
	name, err := ewmh.WmNameGet(w.X, w.id)
	if err != nil {
		return "N/A"
	}
	return name
}

func (w *window) CycleHighlighted() {
	println("highlighted")
}

func (w *window) CycleSelected() {
	fatal(ewmh.ActiveWindowReq(w.X, w.id))
}

func fatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	X, err := xgbutil.NewConn()
	fatal(err)

	keybind.Initialize(X)

	cycle := prompt.NewCycle(X,
		prompt.DefaultCycleTheme, prompt.DefaultCycleConfig)

	clients, err := ewmh.ClientListStackingGet(X)
	fatal(err)
	items := make([]*prompt.CycleItem, 0)
	for i := len(clients) - 1; i >= 0; i-- {
		item := cycle.AddChoice(newWindow(X, cycle.Id(), clients[i]))
		items = append(items, item)
	}

	keyHandlers(X, cycle, items)

	println("Loaded...")
	xevent.Main(X)
}

func headGeom(X *xgbutil.XUtil) xrect.Rect {
	if X.ExtInitialized("XINERAMA") {
		heads, err := xinerama.PhysicalHeads(X)
		if err == nil {
			return heads[0]
		}
	}

	geom, err := xwindow.New(X, X.RootWin()).Geometry()
	fatal(err)
	return geom
}

func keyHandlers(X *xgbutil.XUtil,
	cycle *prompt.Cycle, items []*prompt.CycleItem) {

	geom := headGeom(X)

	keybind.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			shown := cycle.Show(geom, cycleNext, items)
			if !shown {
				log.Fatal("Did not show cycle prompt.")
			}
			cycle.Next()
		}).Connect(X, X.RootWin(), cycleNext, true)
	keybind.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			cycle.Next()
		}).Connect(X, cycle.GrabId(), cycleNext, true)

	keybind.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			shown := cycle.Show(geom, cyclePrev, items)
			if !shown {
				log.Fatal("Did not show cycle prompt.")
			}
			cycle.Prev()
		}).Connect(X, X.RootWin(), cyclePrev, true)
	keybind.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			cycle.Prev()
		}).Connect(X, cycle.GrabId(), cyclePrev, true)
}
