package text

import (
	"image"

	"github.com/BurntSushi/freetype-go/freetype/truetype"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/render"
)

// Input encapsulates the information needed to construct and maintain
// an input window. The only exposed information is the Text field, in case
// you need to inspect it. Input values should *only* be made with the
// NewInput constructor.
type Input struct {
	Text []rune
	*xwindow.Window
	img *xgraphics.Image

	font      *truetype.Font
	fontSize  float64
	fontColor render.Color
	bgColor   render.Color
	padding   int
}

// NewInput constructs Input values. It needs an X connection, a parent window,
// the width of the input box, and theme information related for the font
// and background. Padding separating the text and the edges of the window
// may also be specified.
//
// While NewInput returns an *Input, a Input value also has an xwindow.Window
// value embedded into it. Thus, an Input can also be treated as a normal
// window on which you can assign callbacks, close, destroy, etc.
//
// As with all windows, the Input window should be destroyed when it is no
// longer in used.
func NewInput(X *xgbutil.XUtil, parent xproto.Window, width int, padding int,
	font *truetype.Font, fontSize float64,
	fontColor, bgColor render.Color) *Input {

	_, height := xgraphics.Extents(font, fontSize, "M")

	width, height = width+2*padding, height+2*padding

	img := xgraphics.New(X, image.Rect(0, 0, width, height))
	win := xwindow.Must(xwindow.Create(X, parent))
	win.Listen(xproto.EventMaskKeyPress)
	win.Resize(width, height)

	ti := &Input{
		Window:    win,
		img:       img,
		Text:      make([]rune, 0, 50),
		font:      font,
		fontSize:  fontSize,
		fontColor: fontColor,
		bgColor:   bgColor,
		padding:   padding,
	}

	ti.Render()
	return ti
}

// Render will redraw the background image, and write whatever text is in
// (*Input).Text to the image. No clean-up by the caller is necessary.
//
// Render probably should not be called, unless you are manipulating
// (*Input).Text manually. Otherwise, it is preferrable to use the Add, Remove
// and Reset methods.
func (ti *Input) Render() {
	r, g, b := ti.bgColor.RGB8()
	ti.img.ForExp(func(x, y int) (uint8, uint8, uint8, uint8) {
		return r, g, b, 0xff
	})

	ti.img.Text(ti.padding, ti.padding,
		ti.fontColor.ImageColor(), ti.fontSize, ti.font, string(ti.Text))

	ti.img.XSurfaceSet(ti.Id)
	ti.img.XDraw()
	ti.img.XPaint(ti.Id)
	ti.img.Destroy()
}

// Add will convert a (modifiers, keycode) tuple taken directly from a
// Key{Press,Release}Event to a single character string.
// Note that sometimes this conversion will fail. When it fails, a message will
// be logged and no text will be added.
//
// Note that sometimes the conversion should fail (like when the Shift key
// is pressed), and other times it will fail because the xgbutil/keybind
// package provides inadequate support for keyboard encodings.
//
// I suspect that languages other than English will completely fail here.
//
// If a work-around is desperately needed, use AddLetter.
func (ti *Input) Add(mods uint16, kc xproto.Keycode) {
	s := keybind.LookupString(ti.X, mods, kc)
	if len(s) != 1 {
		logger.Lots.Printf("(*Input).Add: Could not translate string '%s' "+
			"received from the keyboard to a single character.", s)
		return
	}
	ti.AddLetter(rune(s[0]))
}

// AddLetter will add a single character to the input and re-render the input
// box. Note that the Add method is quite convenient and should be used when
// reading Key{Press,Release} events. If input is coming from somewhere else,
// or if Add is not working, resort to AddLetter.
func (ti *Input) AddLetter(char rune) {
	if char == 0 {
		logger.Warning.Printf("(*Input).Add: Strange input: '%s'.", char)
		return
	}

	ti.Text = append(ti.Text, char)
	ti.Render()
}

// SetString will clear the input box and set it to the string provided.
func (ti *Input) SetString(s string) {
	ti.Text = make([]rune, 0, 50)
	for _, char := range s {
		ti.Text = append(ti.Text, char)
	}
	ti.Render()
}

// Remove will remove the last character in the input box and re-render.
// If there are no characters in the input, Remove has no effect.
func (ti *Input) Remove() {
	if len(ti.Text) == 0 {
		return
	}
	ti.Text = ti.Text[:len(ti.Text)-1]
	ti.Render()
}

// Reset will clear the entire input box and re-render.
func (ti *Input) Reset() {
	ti.Text = make([]rune, 0, 50)
	ti.Render()
}
