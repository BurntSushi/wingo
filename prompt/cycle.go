package prompt

import (
	"bytes"
	"image/color"

	"github.com/BurntSushi/freetype-go/freetype/truetype"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/misc"
	"github.com/BurntSushi/wingo/render"
)

// Cycle represents a single cycle prompt. A new cycle prompt can be created by:
//
//	cycle := prompt.NewCycle(XUtilValue, cycleThemeValue, cycleConfigValue)
//
// And it can be displayed using:
//
//	shown := cycle.Show(geometry, "", cycleItemsSlice)
//
// Where the cycle item slice is constructed from *CycleItem values that are
// created using the (*Cycle).AddItem method.
type Cycle struct {
	X      *xgbutil.XUtil // exported for no reason
	theme  *CycleTheme
	config CycleConfig

	items      []*CycleItem
	showing    bool
	selected   int
	grabMods   uint16
	fontHeight int

	win                    *xwindow.Window
	bTop, bBot, bLft, bRht *xwindow.Window
}

// NewCycle creates a new prompt. As many prompts as you want can be created,
// and they could even technically be shown simultaneously so long as at most
// one of them is using a grab. (The grab will fail for the others and they
// will not be shown.)
//
// CycleTheme and CycleConfig values can either use DefaultCycle{Theme,Config}
// values found in this package, or custom ones can be created using
// composite literals.
func NewCycle(X *xgbutil.XUtil, theme *CycleTheme, config CycleConfig) *Cycle {
	cycle := &Cycle{
		X:        X,
		theme:    theme,
		config:   config,
		showing:  false,
		selected: -1,
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

	// Make the top-level window override redirect so the window manager
	// doesn't mess with us.
	cycle.win.Change(xproto.CwOverrideRedirect, 1)

	// Set the colors of each window.
	cclr := func(w *xwindow.Window, clr render.Color) {
		w.Change(xproto.CwBackPixel, uint32(clr.Int()))
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
	cycle.keyResponse().Connect(X, X.Dummy())

	// Guess the maximum font height.
	_, cycle.fontHeight = xgraphics.Extents(
		cycle.theme.Font, cycle.theme.FontSize, "A")

	return cycle
}

func (cycle *Cycle) Destroy() {
	cycle.bTop.Destroy()
	cycle.bBot.Destroy()
	cycle.bLft.Destroy()
	cycle.bRht.Destroy()
	cycle.win.Destroy()
}

// GrabId returns the window id that the grab is set on. This is useful if you
// need to attach any Key{Press,Release} handlers.
func (cycle *Cycle) GrabId() xproto.Window {
	return cycle.X.Dummy()
}

// Id returns the window id of the top-level window of the cycle prompt.
// I'm not sure why you might need it.
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
		if keybind.KeyMatch(X, cycle.config.CancelKey, mods, kc) {
			cycle.Hide()
			return
		}

		mods &= ^keybind.ModGet(X, ev.Detail)
		if cycle.grabMods > 0 {
			if mods&cycle.grabMods == 0 {
				cycle.Choose()
			}
		} else {
			if keybind.KeyMatch(X, cycle.config.ConfirmKey, mods, kc) {
				cycle.Choose()
			}
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

	if cycle.showing {
		return false
	}

	// If there are no items, obviously quit.
	if len(items) == 0 {
		return false
	}

	// Note that SmartGrab is smart and avoids races. Check it out
	// in xgbutil/keybind.go if you're interested.
	// This makes it impossible to press and release alt-tab too quickly
	// to have it not register.
	if cycle.config.Grab {
		if err := keybind.SmartGrab(cycle.X, cycle.X.Dummy()); err != nil {
			logger.Warning.Printf(
				"Could not grab keyboard for prompt cycle: %s", err)
			return false
		}
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

// Hide will hide the cycle prompt and reset any relevant state information.
// The keyboard grab will also be released if one was made.
func (cycle *Cycle) Hide() {
	if !cycle.showing {
		return
	}

	if cycle.config.Grab {
		keybind.SmartUngrab(cycle.X)
	}
	cycle.win.Unmap()

	for _, item := range cycle.items {
		item.hide()
	}
	cycle.showing = false
	cycle.selected = -1
	cycle.grabMods = 0
	cycle.items = nil
}

// Next will highlight the next choice in the dialog.
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

// Prev will highlight the previous choice in the dialog.
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

// AddItem should be thought of as a *CycleItem constructor. Its main role is
// to adapt a CycleChoice value to a value that is suitable for the cycle
// prompt to paint onto its window. The resulting CycleItem value can be used
// to update the image/text.
//
// The CycleItem value must be destroyed by calling (*CycleItem).Destroy when
// it is no longer used. (This frees the X window resources associated with
// the *CycleItem.)
func (cycle *Cycle) AddChoice(choice CycleChoice) *CycleItem {
	return newCycleItem(cycle, choice)
}

// Choose "selects" the currently highlighted choice.
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

// highlight highlights the current choice and unhighlights the rest.
func (cycle *Cycle) highlight() {
	for i, item := range cycle.items {
		if i == cycle.selected {
			item.highlight()
		} else {
			item.unhighlight()
		}
	}
}

// CycleTheme values can be used to create prompts with different colors,
// padding, border sizes, icon sizes, fonts, etc. You may use DefaultCycleTheme
// for a reasonable default theme if you don't care about the particulars.
type CycleTheme struct {
	BorderSize  int
	BgColor     render.Color
	BorderColor render.Color
	Padding     int

	Font      *truetype.Font
	FontSize  float64
	FontColor render.Color

	IconSize         int
	IconBorderSize   int
	IconTransparency int
}

var DefaultCycleTheme = &CycleTheme{
	BorderSize:  10,
	BgColor:     render.NewImageColor(color.RGBA{0xff, 0xff, 0xff, 0xff}),
	BorderColor: render.NewImageColor(color.RGBA{0x0, 0x0, 0x0, 0xff}),
	Padding:     10,
	Font: xgraphics.MustFont(xgraphics.ParseFont(
		bytes.NewBuffer(misc.DataFile("DejaVuSans.ttf")))),
	FontSize:         20.0,
	FontColor:        render.NewImageColor(color.RGBA{0x0, 0x0, 0x0, 0xff}),
	IconSize:         100,
	IconBorderSize:   5,
	IconTransparency: 50,
}

// CycleConfig values can be used to create prompts with different
// configurations. As of right now, the only configuration options supported
// is whether to issue a keyboard grab and the key to
// use to "cancel" the prompt. (If empty, no cancel key feature will be used
// automatically.)
// For a reasonable default configuration, use DefaultCycleConfig. It will
// set "Escape" as the cancel key and issue a grab.
type CycleConfig struct {
	Grab       bool
	CancelKey  string
	ConfirmKey string
}

var DefaultCycleConfig = CycleConfig{
	Grab:       true,
	CancelKey:  "Escape",
	ConfirmKey: "Return",
}
