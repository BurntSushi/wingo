package prompt

import (
	"image/color"

	"code.google.com/p/freetype-go/freetype/truetype"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/misc"
)

type Cycle struct {
	X *xgbutil.XUtil
	theme CycleTheme
	config CycleConfig

	items []*CycleItem
	showing bool
	selected int
	grabMods uint16
	fontHeight int

	win *xwindow.Window
	bTop, bBot, bLft, bRht *xwindow.Window
}

type CycleTheme struct {
	BorderSize int
	BgColor color.RGBA
	BorderColor color.RGBA
	Padding int

	Font *truetype.Font
	FontSize float64
	FontColor color.RGBA

	IconSize int
	IconBorderSize int
	IconTransparency int
}

type CycleConfig struct {
	GrabWin xproto.Window
	CancelKey string
}

func NewCycle(X *xgbutil.XUtil, theme CycleTheme, config CycleConfig) *Cycle {
	cycle := &Cycle{
		X: X,
		theme: theme,
		config: config,
		showing: false,
		selected: 0,
		grabMods: 0,
	}

	// Create all windows used for the base of the cycle prompt.
	// This obviously doesn't include the windows representing the items.
	cwin := func(p xproto.Window) *xwindow.Window {
		return xwindow.Must(xwindow.Create(X, p))
	}
	cycle.win = cwin(X.RootWin())
	cycle.bTop, cycle.bBot = cwin(cycle.win.Id), cwin(cycle.win.Id)
	cycle.bLft, cycle.bRht = cwin(cycle.win.Id), cwin(cycle.win.Id)

	// Set the colors of each window.
	cclr := func(w *xwindow.Window, clr color.RGBA) {
		w.Change(xproto.CwBackPixel, uint32(misc.IntFromColor(clr)))
	}
	cclr(cycle.win, cycle.theme.BgColor)
	cclr(cycle.bTop, cycle.theme.BorderColor)
	cclr(cycle.bBot, cycle.theme.BorderColor)
	cclr(cycle.bLft, cycle.theme.BorderColor)
	cclr(cycle.bRht, cycle.theme.BorderColor)

	// Map the sub-windows once. (Real mapping only happens when
	// cycle.win is mapped.)
	cycle.bTop.Map()
	cycle.bBot.Map()
	cycle.bLft.Map()
	cycle.bRht.Map()

	// Connect the key response handler (i.e., the alt-tab'ing, canceling, etc.)
	cycle.keyResponse().Connect(X, cycle.config.GrabWin)

	// Guess the maximum font height.
	_, cycle.fontHeight = xgraphics.TextMaxExtents(
		cycle.theme.Font, cycle.theme.FontSize, "A")
	cycle.fontHeight += misc.TextBreathe

	return cycle
}

func (cycle *Cycle) Id() xproto.Window {
	return cycle.win.Id
}

// keyResponse translates key board input into two different actions: canceling
// the current prompt and making a choice in the cycle prompt.
// Canceling a prompt corresponds to the "CycleConfig.CancelKey" being pressed.
// Making a choice in the cycle prompt corresponds to releasing all of the
// modifiers used to initiate showing the prompt (when "CycleConfig.AutoChoose"
// is true).
// If CancelKey is empty, then no cancel key functionality is provided.
// If AutoChoose is false, then releasing the modifiers will have no effect.
// 
// For thos interested in the X details:
// The prompt cycle dialog needs to choose the selection when the 
// modifiers (i.e., "alt" in "alt-tab") are released.
// The only way to do this (generally) is to check the raw KeyRelease event.
// Namely, if the keycode *released* is a modifier, we have to and-out
// that modifier from the key release event data. If the modifiers
// remaining aren't up to snuff with the original grabbed modifiers,
// then we can finally "choose" the selection.
// TL;DR - This is how we "exit" the prompt cycle dialog.
func (cycle *Cycle) keyResponse() xevent.KeyReleaseFun {
	f := func(X *xgbutil.XUtil, ev xevent.KeyReleaseEvent) {
		if !cycle.showing {
			return
		}

		mods, kc := keybind.DeduceKeyInfo(ev.State, ev.Detail)

		// If the key release is the cancel key, quit the prompt and
		// don't do anything.
		if misc.KeyMatch(X, cycle.config.CancelKey, mods, kc) {
			cycle.Hide()
			return
		}

		mods &= ^keybind.ModGet(X, ev.Detail)
		if cycle.grabMods > 0 && mods&cycle.grabMods == 0 {

			cycle.Choose()
		}
	}
	return xevent.KeyReleaseFun(f)
}


// Show will map and show the slice of items provided.
//
// 'workarea' is the rectangle to position the prompt window in. (i.e.,
// typically the rectangle of the monitor to place it on.)
//
// 'keyStr' is an optional parameter. If this prompt is shown in
// response to a keybinding, then keyStr should be the keybinding used.
// If there are modifiers used in the keyStr, the prompt will automatically
// close if all of the modifiers are released. (This is the "alt-tab"
// functionality.)
// Note that if you don't want this auto-closing feature, simply leave keyStr
// blank, even if the prompt is shown in response to a key binding.
//
// Show returns false if the prompt cannot be shown for some reason.
func (cycle *Cycle) Show(workarea xrect.Rect,
	keyStr string, items []*CycleItem) bool {

	// If there are no items, obviously quit.
	if len(items) == 0 {
		return false
	}

	// Note that SmartGrab is smart and avoids races. Check it out
	// in xgbutil/keybind.go if you're interested.
	// This makes it impossible to press and release alt-tab too quickly
	// to have it not register.
	if err := keybind.SmartGrab(cycle.X, cycle.config.GrabWin); err != nil {
		logger.Warning.Println("Could not grab keyboard for prompt cycle: %s",
			err)
		return false
	}

	// Save the list of cycle items (this how we know when to cycle between
	// them). Namely, cycle.selected is an index to this list.
	cycle.items = items

	// Save the modifiers used, if any.
	cycle.grabMods, _, _ = keybind.ParseString(cycle.X, keyStr)

	// Put the prompt window on top of the window stack.
	cycle.win.Stack(xproto.StackModeAbove)

	// Create some short aliases and start computing the geometry of the
	// cycle window.
	bs := cycle.theme.BorderSize
	cbs := cycle.theme.IconBorderSize
	is := cycle.theme.IconSize
	pad := cycle.theme.Padding

	maxWidth := int(float64(workarea.Width()) * 0.8)
	x, y := bs+pad, bs+pad+cbs+cycle.fontHeight
	width := 2 * (bs + pad)
	height := (2 * (bs + pad + cbs)) + is + cbs + cycle.fontHeight
	maxFontWidth := 0

	widthStatic := false // when true, we stop increasing width
	for _, item := range items {
		maxFontWidth = misc.Max(maxFontWidth, item.text.Geom.Width())

		// Check if we should move on to the next row.
		if x+(is+(2*cbs))+pad+bs > maxWidth {
			x = bs + pad
			y += is + (2 * cbs)
			height += is + (2 * cbs)
			widthStatic = true
		}

		// Position the icon window and map its active version or its
		// inactive version if it's iconified.
		item.show(x, y)

		// Only increase the width if we're still adding icons to the first row.
		if !widthStatic {
			width += is + (2 * cbs)
		}
		x += is + (2 * cbs)
	}

	// If the computed width is less than the max font width, then increase
	// the width of the prompt to fit the longest window title.
	// Forcefully cap it as the maxWidth, though.
	if maxFontWidth+2*(pad+bs) > width {
		width = misc.Min(maxWidth, maxFontWidth+2*(pad+bs))
	}

	// position the damn window based on its width/height (i.e., center it)
	posx := workarea.X() + workarea.Width()/2 - width/2
	posy := workarea.Y() + workarea.Height()/2 - height/2

	// Issue the configure requests. We also need to adjust the borders.
	cycle.win.MoveResize(posx, posy, width, height)
	cycle.bTop.Resize(width, bs)
	cycle.bBot.MoveResize(0, height-bs, width, bs)
	cycle.bLft.Resize(bs, height)
	cycle.bRht.MoveResize(width-bs, 0, bs, height)

	cycle.showing = true
	cycle.selected = -1
	cycle.win.Map()

	return true
}

func (cycle *Cycle) Next() {
	if !cycle.showing {
		return
	}

	if cycle.selected == -1 {
		if len(cycle.items) > 1 {
			cycle.selected = 1
		} else {
			cycle.selected = 0
		}
	} else {
		cycle.selected++
	}

	cycle.selected = misc.Mod(cycle.selected, len(cycle.items))
	cycle.highlight()
}

func (cycle *Cycle) Prev() {
	if !cycle.showing {
		return
	}

	if cycle.selected == -1 {
		cycle.selected = len(cycle.items) - 1
	} else {
		cycle.selected--
	}

	cycle.selected = misc.Mod(cycle.selected, len(cycle.items))
	cycle.highlight()
}

func (cycle *Cycle) AddItem(choice CycleChoice) *CycleItem {
	return newCycleItem(cycle, choice)
}

func (cycle *Cycle) Hide() {
	cycle.win.Unmap()
	keybind.SmartUngrab(cycle.X)

	cycle.showing = false
	cycle.selected = -1
	cycle.grabMods = 0
	cycle.items = nil
}

func (cycle *Cycle) Choose() {
	if !cycle.showing ||
		len(cycle.items) == 0 ||
		cycle.selected < 0 ||
		cycle.selected >= len(cycle.items) {

		return
	}

	cycle.items[cycle.selected].choose()
	cycle.Hide()
}

func (cycle *Cycle) highlight() {
	for i, item := range cycle.items {
		if i == cycle.selected {
			item.highlight()
		} else {
			item.unhighlight()
		}
	}
}

