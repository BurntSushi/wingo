package prompt

import (
	"bytes"
	"image/color"

	"code.google.com/p/freetype-go/freetype/truetype"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/misc"
	"github.com/BurntSushi/wingo/render"
	"github.com/BurntSushi/wingo/text"
)

type Input struct {
	X      *xgbutil.XUtil
	theme  *InputTheme
	config InputConfig

	history      []string
	historyIndex int

	showing  bool
	do       func(inp *Input, text string)
	canceled func(inp *Input)

	win                          *xwindow.Window
	label                        *xwindow.Window
	input                        *text.Input
	bInp, bTop, bBot, bLft, bRht *xwindow.Window
}

func NewInput(X *xgbutil.XUtil, theme *InputTheme, config InputConfig) *Input {
	input := &Input{
		X:       X,
		theme:   theme,
		config:  config,
		showing: false,
		do:      nil,
		history: make([]string, 0, 100),
	}

	// Create all windows used for the base of the input prompt.
	cwin := func(p xproto.Window) *xwindow.Window {
		return xwindow.Must(xwindow.Create(X, p))
	}
	input.win = cwin(X.RootWin())
	input.label = cwin(input.win.Id)
	input.input = text.NewInput(X, input.win.Id, 1000, theme.Padding,
		theme.Font, theme.FontSize, theme.FontColor, theme.BgColor)
	input.bInp = cwin(input.win.Id)
	input.bTop, input.bBot = cwin(input.win.Id), cwin(input.win.Id)
	input.bLft, input.bRht = cwin(input.win.Id), cwin(input.win.Id)

	// Make the top-level window override redirect so the window manager
	// doesn't mess with us.
	input.win.Change(xproto.CwOverrideRedirect, 1)
	input.win.Listen(xproto.EventMaskFocusChange)
	input.input.Listen(xproto.EventMaskKeyPress)

	// Colorize the windows.
	cclr := func(w *xwindow.Window, clr render.Color) {
		w.Change(xproto.CwBackPixel, clr.Uint32())
	}
	cclr(input.win, input.theme.BgColor)
	cclr(input.bInp, input.theme.BorderColor)
	cclr(input.bTop, input.theme.BorderColor)
	cclr(input.bBot, input.theme.BorderColor)
	cclr(input.bLft, input.theme.BorderColor)
	cclr(input.bRht, input.theme.BorderColor)

	// Map the sub-windows once.
	input.label.Map()
	input.input.Map()
	input.bInp.Map()
	input.bTop.Map()
	input.bBot.Map()
	input.bLft.Map()
	input.bRht.Map()

	// Connect the key response handler.
	// The handler is responsible for tab completion and quitting if the
	// cancel key has been pressed.
	input.keyResponse().Connect(X, input.input.Id)
	input.focusResponse().Connect(X, input.win.Id)

	return input
}

func (inp *Input) Showing() bool {
	return inp.showing
}

func (inp *Input) Destroy() {
	inp.input.Destroy()
	inp.label.Destroy()
	inp.bInp.Destroy()
	inp.bTop.Destroy()
	inp.bBot.Destroy()
	inp.bLft.Destroy()
	inp.bRht.Destroy()
	inp.win.Destroy()
}

func (inp *Input) Id() xproto.Window {
	return inp.win.Id
}

func (inp *Input) Show(workarea xrect.Rect, label string,
	do func(inp *Input, text string), canceled func(inp *Input)) bool {

	if inp.showing {
		return false
	}

	inp.win.Stack(xproto.StackModeAbove)
	inp.input.Reset()

	text.DrawText(inp.label, inp.theme.Font, inp.theme.FontSize,
		inp.theme.FontColor, inp.theme.BgColor, label)

	pad, bs := inp.theme.Padding, inp.theme.BorderSize
	width := (pad * 2) + (bs * 2) +
		inp.label.Geom.Width() + inp.theme.InputWidth
	height := (pad * 2) + (bs * 2) + inp.label.Geom.Height()

	// position the damn window based on its width/height (i.e., center it)
	posx := workarea.X() + workarea.Width()/2 - width/2
	posy := workarea.Y() + workarea.Height()/2 - height/2

	inp.win.MoveResize(posx, posy, width, height)
	inp.label.Move(bs+pad, pad+bs)
	inp.bInp.MoveResize(pad+inp.label.Geom.X()+inp.label.Geom.Width(), 0,
		bs, height)
	inp.bTop.Resize(width, bs)
	inp.bBot.MoveResize(0, height-bs, width, bs)
	inp.bLft.Resize(bs, height)
	inp.bRht.MoveResize(width-bs, 0, bs, height)
	inp.input.Move(inp.bInp.Geom.X()+inp.bInp.Geom.Width(), bs)

	inp.showing = true
	inp.do = do
	inp.canceled = canceled
	inp.win.Map()
	inp.input.Focus()
	inp.historyIndex = len(inp.history)

	return true
}

func (inp *Input) Hide() {
	if !inp.showing {
		return
	}

	inp.win.Unmap()
	inp.input.Reset()

	inp.showing = false
	inp.do = nil
	inp.canceled = nil
}

func (inp *Input) focusResponse() xevent.FocusOutFun {
	f := func(X *xgbutil.XUtil, ev xevent.FocusOutEvent) {
		if !ignoreFocus(ev.Mode, ev.Detail) {
			if inp.canceled != nil {
				inp.canceled(inp)
			}
			inp.Hide()
		}
	}
	return xevent.FocusOutFun(f)
}

func (inp *Input) keyResponse() xevent.KeyPressFun {
	f := func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
		if !inp.showing {
			return
		}

		mods, kc := keybind.DeduceKeyInfo(ev.State, ev.Detail)
		switch {
		case keybind.KeyMatch(X, "Up", mods, kc):
			if inp.historyIndex > 0 {
				inp.historyIndex--
				inp.input.SetString(inp.history[inp.historyIndex])
			}
		case keybind.KeyMatch(X, "Down", mods, kc):
			if inp.historyIndex < len(inp.history) {
				inp.historyIndex++
				if inp.historyIndex < len(inp.history) {
					inp.input.SetString(inp.history[inp.historyIndex])
				} else {
					inp.input.Reset()
				}
			}
		case keybind.KeyMatch(X, inp.config.BackspaceKey, mods, kc):
			inp.input.Remove()
		case keybind.KeyMatch(X, inp.config.ConfirmKey, mods, kc):
			if inp.do != nil {
				s := string(inp.input.Text)
				histLen := len(inp.history)

				// Don't added repeated entries.
				if histLen == 0 || s != inp.history[histLen-1] {
					inp.history = append(inp.history, s)
				}
				inp.do(inp, s)
			}
		case keybind.KeyMatch(X, inp.config.CancelKey, mods, kc):
			if inp.canceled != nil {
				inp.canceled(inp)
			}
			inp.Hide()
		default:
			inp.input.Add(mods, kc)
		}
	}
	return xevent.KeyPressFun(f)
}

type InputTheme struct {
	BorderSize  int
	BgColor     render.Color
	BorderColor render.Color
	Padding     int

	Font      *truetype.Font
	FontSize  float64
	FontColor render.Color

	InputWidth int
}

var DefaultInputTheme = &InputTheme{
	BorderSize:  5,
	BgColor:     render.NewImageColor(color.RGBA{0xff, 0xff, 0xff, 0xff}),
	BorderColor: render.NewImageColor(color.RGBA{0x0, 0x0, 0x0, 0xff}),
	Padding:     10,

	Font: xgraphics.MustFont(xgraphics.ParseFont(
		bytes.NewBuffer(misc.DataFile("DejaVuSans.ttf")))),
	FontSize:   20.0,
	FontColor:  render.NewImageColor(color.RGBA{0x0, 0x0, 0x0, 0xff}),
	InputWidth: 500,
}

type InputConfig struct {
	CancelKey    string
	BackspaceKey string
	ConfirmKey   string
}

var DefaultInputConfig = InputConfig{
	CancelKey:    "Escape",
	BackspaceKey: "BackSpace",
	ConfirmKey:   "Return",
}
