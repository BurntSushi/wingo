package prompt

import (
	"bytes"
	"image/color"
	"strings"

	"code.google.com/p/freetype-go/freetype/truetype"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/bindata"
	"github.com/BurntSushi/wingo/misc"
	"github.com/BurntSushi/wingo/text"
)

const (
	TabCompletePrefix = iota
	TabCompleteAny
)

type Select struct {
	X      *xgbutil.XUtil
	theme  SelectTheme
	config SelectConfig

	showing     bool
	selected    int
	tabComplete int

	groups []*SelectShowGroup
	items  []*SelectItem

	input *text.Input

	win                          *xwindow.Window
	bInp, bTop, bBot, bLft, bRht *xwindow.Window
}

func NewSelect(X *xgbutil.XUtil,
	theme SelectTheme, config SelectConfig) *Select {

	slct := &Select{
		X:           X,
		theme:       theme,
		config:      config,
		showing:     false,
		selected:    -1,
		tabComplete: TabCompletePrefix,
	}

	// Create all windows used for the base of the select prompt. This
	// obviously doesn't include the windows representing the items/groups.
	cwin := func(p xproto.Window) *xwindow.Window {
		return xwindow.Must(xwindow.Create(X, p))
	}
	slct.win = cwin(X.RootWin())
	slct.bInp = cwin(slct.win.Id)
	slct.bTop, slct.bBot = cwin(slct.win.Id), cwin(slct.win.Id)
	slct.bLft, slct.bRht = cwin(slct.win.Id), cwin(slct.win.Id)

	// Make the top-level window override redirect so the window manager
	// doesn't mess with us.
	slct.win.Change(xproto.CwOverrideRedirect, 1)

	// Create the text input window.
	slct.input = text.NewInput(X, slct.win.Id, 1000, 10,
		slct.theme.Font, slct.theme.FontSize,
		slct.theme.FontColor, slct.theme.BgColor)

	// Colorize the windows.
	cclr := func(w *xwindow.Window, clr color.RGBA) {
		w.Change(xproto.CwBackPixel, uint32(misc.IntFromColor(clr)))
	}
	cclr(slct.win, slct.theme.BgColor)
	cclr(slct.bInp, slct.theme.BorderColor)
	cclr(slct.bTop, slct.theme.BorderColor)
	cclr(slct.bBot, slct.theme.BorderColor)
	cclr(slct.bLft, slct.theme.BorderColor)
	cclr(slct.bRht, slct.theme.BorderColor)

	// Map the sub-windows once. (Real mapping only happens when
	// cycle.win is mapped.)
	slct.bInp.Map()
	slct.bTop.Map()
	slct.bBot.Map()
	slct.bLft.Map()
	slct.bRht.Map()
	slct.input.Map()

	// Connect the key response handler.
	// The handler is responsible for tab completion and quitting if the
	// cancel key has been pressed.
	slct.keyResponse().Connect(X, slct.input.Id)

	return slct
}

func (slct *Select) GrabId() xproto.Window {
	return slct.X.Dummy()
}

func (slct *Select) Id() xproto.Window {
	return slct.win.Id
}

// NewStaticGroup returns a value implementing the SelectGroup interface with
// the label provided. This is useful for generating group labels that never
// change. (i.e., in Wingo, these would be the no-label, Visible and Hidden
// groups. While the groups defined by workspace have to implement the
// SelectGroup interface themselves.)
func (slct *Select) NewStaticGroup(label string) SelectGroup {
	return group(label)
}

func (slct *Select) keyResponse() xevent.KeyPressFun {
	f := func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
		if !slct.showing {
			return
		}

		beforeLen := len(slct.input.Text)
		mods, kc := keybind.DeduceKeyInfo(ev.State, ev.Detail)

		slct.input.Add(mods, kc)

		switch {
		case misc.KeyMatch(X, slct.config.BackspaceKey, mods, kc):
			slct.input.Remove()
		case misc.KeyMatch(X, slct.config.CancelKey, mods, kc):
			slct.Hide()
			return
		case misc.KeyMatch(X, slct.config.ConfirmKey, mods, kc):
			if slct.selected >= 0 && slct.selected < len(slct.items) {
				slct.items[slct.selected].choose()
				slct.Hide()
			} else if len(slct.items) == 1 {
				slct.items[0].choose()
				slct.Hide()
			}
			return
		case misc.KeyMatch(X, "Tab", mods, kc) ||
			misc.KeyMatch(X, "ISO_Left_Tab", mods, kc):

			if len(slct.items) == 0 {
				break
			}
			if mods&xproto.ModMaskShift > 0 {
				if slct.selected == -1 {
					slct.selected++
				}
				slct.selected = misc.Mod(slct.selected-1, len(slct.items))
			} else {
				slct.selected = misc.Mod(slct.selected+1, len(slct.items))
			}
			slct.highlight()
		}

		// If the length of the input changed, then re-evaluate completion
		if beforeLen != len(slct.input.Text) {
			slct.FilterItems(string(slct.input.Text))
			slct.selected = -1
		}
	}
	return xevent.KeyPressFun(f)
}

func (slct *Select) Show(workarea xrect.Rect, tabCompleteType int,
	groups []*SelectShowGroup) bool {

	// if there aren't any groups, we obviously don't need to show anything.
	if len(groups) == 0 {
		return false
	}

	slct.groups = groups
	slct.tabComplete = tabCompleteType

	slct.win.Stack(xproto.StackModeAbove)
	slct.input.Reset()

	// Position the initial list of items with no filter.
	slct.FilterItems("")

	// Create some short aliases and start computing the geometry of the
	// prompt window.
	bs := slct.theme.BorderSize
	pad := slct.theme.Padding

	maxWidth := int(float64(workarea.Width()) * 0.8)
	inpHeight := slct.input.Geom.Height()
	height := 2*(bs+pad) + inpHeight + (2 * misc.TextBreathe) + bs
	maxFontWidth := 0
	didLabelSpacing := false
	for _, group := range slct.groups {
		if group.hasLabel() {
			maxFontWidth = misc.Max(maxFontWidth, group.win.Geom.Width())
			height += group.win.Geom.Height() + slct.theme.LabelSpacing
			didLabelSpacing = true
		}
		for _, item := range group.items {
			maxFontWidth = misc.Max(maxFontWidth, item.regular.Geom.Width())
			height += item.regular.Geom.Height()
		}
	}

	if didLabelSpacing {
		height -= slct.theme.LabelSpacing
	}
	width := misc.Min(maxWidth, maxFontWidth+2*(bs+pad))

	// position the damn window based on its width/height (i.e., center it)
	posx := workarea.X() + workarea.Width()/2 - width/2
	posy := workarea.Y() + workarea.Height()/2 - height/2

	// Issue the configure requests. We also need to adjust the borders.
	slct.win.MoveResize(posx, posy, width, height)
	slct.bInp.Resize(width, bs)
	slct.bTop.Resize(width, bs)
	slct.bBot.MoveResize(0, height-bs, width, bs)
	slct.bLft.Resize(bs, height)
	slct.bRht.MoveResize(width-bs, 0, bs, height)

	slct.showing = true
	slct.selected = -1
	slct.win.Map()

	return true
}

func (slct *Select) FilterItems(search string) {
	bs := slct.theme.BorderSize
	pad := slct.theme.Padding
	inpHeight := slct.input.Geom.Height()
	needle := strings.ToLower(search)

	slct.items = make([]*SelectItem, 0)

	x, y := bs+pad, (2*bs)+pad+inpHeight+(2*misc.TextBreathe)
	for _, group := range slct.groups {
		shown := false // true when at least 1 item is showing

		if group.hasLabel() {
			y += group.win.Geom.Height()
		}
		for _, item := range group.items {
			haystack := strings.ToLower(item.text)
			switch slct.tabComplete {
			case TabCompleteAny:
				if !strings.Contains(haystack, needle) {
					continue
				}
			default:
				if !strings.HasPrefix(haystack, needle) {
					continue
				}
			}

			item.show(x, y)
			y += item.regular.Geom.Height()
			slct.items = append(slct.items, item)
			shown = true
		}
		if group.hasLabel() {
			if shown {
				group.show(x, y)
				y += slct.theme.LabelSpacing
			} else {
				y -= group.win.Geom.Height()
			}
		}
	}
}

func (slct *Select) Hide() {
	if !slct.showing {
		return
	}

	slct.win.Unmap()

	for _, group := range slct.groups {
		group.hide()
	}
	slct.showing = false
	slct.selected = -1
	slct.groups = nil
	slct.items = nil
	slct.tabComplete = TabCompletePrefix
}

func (slct *Select) highlight() {
}

type SelectTheme struct {
	BorderSize  int
	BgColor     color.RGBA
	BorderColor color.RGBA
	Padding     int

	Font      *truetype.Font
	FontSize  float64
	FontColor color.RGBA

	ActiveBgColor   color.RGBA
	ActiveFontColor color.RGBA

	LabelBgColor   color.RGBA
	LabelFont      *truetype.Font
	LabelFontSize  float64
	LabelFontColor color.RGBA
	LabelSpacing   int
}

var DefaultSelectTheme = SelectTheme{
	BorderSize:  10,
	BgColor:     color.RGBA{0xff, 0xff, 0xff, 0xff},
	BorderColor: color.RGBA{0x0, 0x0, 0x0, 0xff},
	Padding:     10,

	Font: xgraphics.MustFont(xgraphics.ParseFont(
		bytes.NewBuffer(bindata.DejavusansTtf()))),
	FontSize:  20.0,
	FontColor: color.RGBA{0x0, 0x0, 0x0, 0xff},

	ActiveBgColor:   color.RGBA{0x0, 0x0, 0x0, 0xff},
	ActiveFontColor: color.RGBA{0xff, 0xff, 0xff, 0xff},

	LabelBgColor: color.RGBA{0xff, 0xff, 0xff, 0xff},
	LabelFont: xgraphics.MustFont(xgraphics.ParseFont(
		bytes.NewBuffer(bindata.DejavusansTtf()))),
	LabelFontSize:  30.0,
	LabelFontColor: color.RGBA{0x0, 0x0, 0x0, 0xff},
	LabelSpacing:   20,
}

type SelectConfig struct {
	CancelKey    string
	BackspaceKey string
	ConfirmKey   string
	TabKey       string
}

var DefaultSelectConfig = SelectConfig{
	CancelKey:    "Escape",
	BackspaceKey: "BackSpace",
	ConfirmKey:   "Return",
}
