package main

import (
	"image/color"
	"log"
	"os"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xgraphics"

	"github.com/BurntSushi/wingo/prompt"
)

var (
	theme = prompt.CycleTheme{
		BorderSize: 5,
		BgColor: color.RGBA{0xff, 0xff, 0xff, 0xff},
		BorderColor: color.RGBA{0x33, 0x66, 0xff, 0xff},
		Padding: 10,
		Font: nil, // set later; have to read the font file
		FontSize: 20.0,
		FontColor: color.RGBA{0x0, 0x0, 0x0, 0xff},
		IconSize: 100,
		IconBorderSize: 5,
	}

	config = prompt.CycleConfig{
		CancelKey: "Escape",
	}

	fontPath = "/usr/share/fonts/TTF/FreeMonoBold.ttf"
)

type window struct {
	X *xgbutil.XUtil
	Id xproto.Window
}

func newWindow(X *xgbutil.XUtil, parent, id xproto.Window) *window {
	return &window{
		X: X,
		Id: id,
	}
}

func (w *window) CycleIsActive() bool {
	return true
}

func (w *window) CycleImage() *xgraphics.Image {
	ximg, err := xgraphics.FindIcon(w.X, w.Id, theme.IconSize, theme.IconSize)
	fatal(err)

	return ximg
}

func (w *window) CycleText() string {
	name, err := ewmh.WmNameGet(w.X, w.Id)
	if err != nil {
		return "N/A"
	}
	return name
}

func (w *window) CycleHighlighted() {
	println("highlighted")
}

func (w *window) CycleSelected() {
	println("selected")
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

	fontReader, err := os.Open(fontPath)
	fatal(err)

	font, err := xgraphics.ParseFont(fontReader)
	fatal(err)

	theme.Font = font
	config.GrabWin = X.Dummy()
	cycle := prompt.NewCycle(X, theme, config)

	clients, err := ewmh.ClientListGet(X)
	fatal(err)
	items := make([]*prompt.CycleItem, 0)
	for _, cid := range clients {
		item := cycle.AddItem(newWindow(X, cycle.Id(), cid))
		items = append(items, item)
	}

	keybind.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			shown := cycle.Show(xrect.New(0, 0, 1920, 1080), "Mod4-tab", items)
			if !shown {
				log.Fatal("Did not show cycle prompt.")
			}
			cycle.Next()
		}).Connect(X, X.RootWin(), "Mod4-tab", true)
	keybind.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			cycle.Next()
		}).Connect(X, X.Dummy(), "Mod4-tab", true)

	keybind.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			shown := cycle.Show(xrect.New(0, 0, 1920, 1080), "Mod4-Shift-tab", items)
			if !shown {
				log.Fatal("Did not show cycle prompt.")
			}
			cycle.Prev()
		}).Connect(X, X.RootWin(), "Mod4-Shift-tab", true)
	keybind.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			cycle.Prev()
		}).Connect(X, X.Dummy(), "Mod4-Shift-tab", true)

	println("Loaded...")
	xevent.Main(X)
}
