package prompt

import (
	"bytes"
	"image/color"
	"time"

	"code.google.com/p/freetype-go/freetype/truetype"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/bindata"
	"github.com/BurntSushi/wingo/render"
	"github.com/BurntSushi/wingo/text"
)

type Message struct {
	X      *xgbutil.XUtil
	theme  *MessageTheme
	config MessageConfig

	showing  bool
	lastShow time.Time
	duration time.Duration
	hidden   func(msg *Message)

	cancelTimeout chan struct{}

	win                    *xwindow.Window
	text                   *xwindow.Window
	bTop, bBot, bLft, bRht *xwindow.Window
}

func NewMessage(X *xgbutil.XUtil,
	theme *MessageTheme, config MessageConfig) *Message {

	msg := &Message{
		X:             X,
		theme:         theme,
		config:        config,
		showing:       false,
		duration:      0,
		hidden:        nil,
		cancelTimeout: make(chan struct{}, 0),
	}

	// Create all windows used for the base of the message prompt.
	cwin := func(p xproto.Window) *xwindow.Window {
		return xwindow.Must(xwindow.Create(X, p))
	}
	msg.win = cwin(X.RootWin())
	msg.text = cwin(msg.win.Id)
	msg.bTop, msg.bBot = cwin(msg.win.Id), cwin(msg.win.Id)
	msg.bLft, msg.bRht = cwin(msg.win.Id), cwin(msg.win.Id)

	// Make the top-level window override redirect so the window manager
	// doesn't mess with us.
	msg.win.Change(xproto.CwOverrideRedirect, 1)
	msg.win.Listen(xproto.EventMaskFocusChange)
	msg.text.Listen(xproto.EventMaskKeyPress)

	// Colorize the windows.
	cclr := func(w *xwindow.Window, clr render.Color) {
		w.Change(xproto.CwBackPixel, clr.Uint32())
	}
	cclr(msg.win, msg.theme.BgColor)
	cclr(msg.bTop, msg.theme.BorderColor)
	cclr(msg.bBot, msg.theme.BorderColor)
	cclr(msg.bLft, msg.theme.BorderColor)
	cclr(msg.bRht, msg.theme.BorderColor)

	// Map the sub-windows once.
	msg.text.Map()
	msg.bTop.Map()
	msg.bBot.Map()
	msg.bLft.Map()
	msg.bRht.Map()

	msg.keyResponse().Connect(X, msg.text.Id)
	msg.focusResponse().Connect(X, msg.win.Id)

	return msg
}

func (msg *Message) Showing() bool {
	return msg.showing
}

func (msg *Message) Destroy() {
	msg.text.Destroy()
	msg.bTop.Destroy()
	msg.bBot.Destroy()
	msg.bLft.Destroy()
	msg.bRht.Destroy()
	msg.win.Destroy()
}

func (msg *Message) Id() xproto.Window {
	return msg.win.Id
}

func (msg *Message) Show(workarea xrect.Rect, message string,
	duration time.Duration, hidden func(msg *Message)) bool {

	if msg.showing {
		return false
	}

	msg.win.Stack(xproto.StackModeAbove)

	text.DrawText(msg.text, msg.theme.Font, msg.theme.FontSize,
		msg.theme.FontColor, msg.theme.BgColor, message)

	pad, bs := msg.theme.Padding, msg.theme.BorderSize
	width := (pad * 2) + (bs * 2) + msg.text.Geom.Width()
	height := (pad * 2) + (bs * 2) + msg.text.Geom.Height()

	// position the damn window based on its width/height (i.e., center it)
	posx := workarea.X() + workarea.Width()/2 - width/2
	posy := workarea.Y() + workarea.Height()/2 - height/2

	msg.win.MoveResize(posx, posy, width, height)
	msg.text.Move(bs+pad, pad+bs)
	msg.bTop.Resize(width, bs)
	msg.bBot.MoveResize(0, height-bs, width, bs)
	msg.bLft.Resize(bs, height)
	msg.bRht.MoveResize(width-bs, 0, bs, height)

	msg.showing = true
	msg.duration = duration
	msg.hidden = hidden
	msg.win.Map()
	msg.lastShow = time.Now()

	// If the duration is non-zero, then wait for that amount of time and
	// automatically hide the popup. Otherwise, focus the window and wait
	// for user interaction.
	if duration == 0 {
		msg.text.Focus()
	} else {
		go func() {
			// If `Hide` is called before the timeout expires, we'll
			// cancel the timeout.
			select {
			case <-time.After(duration):
				msg.Hide()
			case <-msg.cancelTimeout:
			}
		}()
	}

	return true
}

func (msg *Message) Hide() {
	if !msg.showing {
		return
	}

	// If there is a timeout in progress, can it.
	select {
	case msg.cancelTimeout <- struct{}{}:
	default:
	}

	msg.hidden(msg)
	msg.win.Unmap()

	msg.showing = false
	msg.duration = 0
	msg.hidden = nil
}

func (msg *Message) focusResponse() xevent.FocusOutFun {
	f := func(X *xgbutil.XUtil, ev xevent.FocusOutEvent) {
		if !ignoreFocus(ev.Mode, ev.Detail) {
			// We only want to lose focus if enough time has elapsed since
			// the message was shown. Otherwise, we might get some weird
			// events that cause us to hide prematurely...
			elapsed := time.Duration(time.Since(msg.lastShow).Nanoseconds())
			if elapsed/time.Millisecond >= 100 {
				msg.Hide()
			} else {
				// Otherwise, we need to reacquire focus.
				msg.text.Focus()
			}
		}
	}
	return xevent.FocusOutFun(f)
}

func (msg *Message) keyResponse() xevent.KeyPressFun {
	f := func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
		if !msg.showing {
			return
		}

		mods, kc := keybind.DeduceKeyInfo(ev.State, ev.Detail)
		switch {
		case keybind.KeyMatch(X, msg.config.ConfirmKey, mods, kc):
			fallthrough
		case keybind.KeyMatch(X, msg.config.CancelKey, mods, kc):
			msg.Hide()
		}
	}
	return xevent.KeyPressFun(f)
}

type MessageTheme struct {
	BorderSize  int
	BgColor     render.Color
	BorderColor render.Color
	Padding     int

	Font      *truetype.Font
	FontSize  float64
	FontColor render.Color
}

var DefaultMessageTheme = &MessageTheme{
	BorderSize:  5,
	BgColor:     render.NewImageColor(color.RGBA{0xff, 0xff, 0xff, 0xff}),
	BorderColor: render.NewImageColor(color.RGBA{0x0, 0x0, 0x0, 0xff}),
	Padding:     10,

	Font: xgraphics.MustFont(xgraphics.ParseFont(
		bytes.NewBuffer(bindata.DejavusansTtf()))),
	FontSize:  20.0,
	FontColor: render.NewImageColor(color.RGBA{0x0, 0x0, 0x0, 0xff}),
}

type MessageConfig struct {
	CancelKey  string
	ConfirmKey string
}

var DefaultMessageConfig = MessageConfig{
	CancelKey:  "Escape",
	ConfirmKey: "Return",
}
