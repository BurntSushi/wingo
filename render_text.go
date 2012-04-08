package main

import (
	"image"
	"image/draw"

	"code.google.com/p/freetype-go/freetype/truetype"

	"github.com/BurntSushi/xgbutil/xgraphics"
)

// The extents appear to cut off some of the text.
// It may be a bad idea to hard code this value, but my knowledge
// of drawing fonts is pretty shaky. It seems to work.
// (This is always added to the width/height of text extents.)
const renderTextBreathe = 5

type textInput struct {
	win       *window
	img       *wImg
	text      []rune
	bgColor   int
	font      *truetype.Font
	fontSize  float64
	fontColor int
}

func renderTextInputCreate(parent *window, bgColor int,
	font *truetype.Font, fontSize float64, fontColor int,
	width int) *textInput {

	_, height, err := xgraphics.TextMaxExtents(font, fontSize, "M")
	if err != nil {
		return nil
	}

	img := renderSolid(bgColor, width, height+renderTextBreathe)
	win := createImageWindow(parent.id, img, 0)

	return &textInput{
		win:       win,
		img:       img,
		text:      make([]rune, 0),
		bgColor:   bgColor,
		font:      font,
		fontSize:  fontSize,
		fontColor: fontColor,
	}
}

// reset removes all characters from the input box and re-renders.
func (ti *textInput) reset() {
	ti.text = ti.text[:0]
	ti.render()
}

// Needs to be able to filter the stuff in "text"
// i.e., a-zA-Z0-9_- special chars, etc.
// Essentially, even though text is a string, we're only going to take the
// first character of that string iff len(text) == 1. Otherwise, we ignore it.
func (ti *textInput) add(char rune) {
	if char == 0 {
		return
	}
	ti.text = append(ti.text, char)
	ti.render()
}

func (ti *textInput) remove() {
	if len(ti.text) == 0 {
		return
	}
	ti.text = ti.text[:len(ti.text)-1]
	ti.render()
}

func (ti *textInput) render() {
	draw.Draw(ti.img, ti.img.Bounds(),
		image.NewUniform(colorFromInt(ti.bgColor)), image.ZP, draw.Src)
	renderText(ti.img, ti.bgColor, ti.font, ti.fontSize, ti.fontColor,
		string(ti.text))
	xgraphics.PaintImg(X, ti.win.id, ti.img)
}

// renderTextSolid does the plumbing for drawing text on a solid background.
// It returns the width and height of the TEXT, the image itself may be bigger.
func renderTextSolid(bgColor int, font *truetype.Font, fontSize float64,
	fontColor int, text string) (*wImg, int, int, error) {

	// ew and eh are the *max* text extents (since it assumes every character
	// is 1em in width). I don't know how to get accurate text extents without
	// actually drawing the text, so this will have to due for now. We'll end
	// up creating bigger images than we need, but we can resize the window
	// itself after we get the *real* extents when we draw the text.
	ew, eh, err := xgraphics.TextMaxExtents(font, fontSize, text)
	if err != nil {
		logWarning.Printf("Could not get text extents for text '%s' "+
			"because: %v", text, err)
		logWarning.Printf("Resorting to default with of 200.")
		ew = 200
	}

	// Draw the background for the text plus some breathing room
	textImg := renderSolid(bgColor,
		ew+renderTextBreathe, eh+renderTextBreathe)
	return renderText(textImg, bgColor, font, fontSize, fontColor, text)
}

// renderText draws text on a source background image with some breathing
// room. It returns the text extents (with the breathing room including),
// and logs an error if unsuccessful.
func renderText(src *wImg, bgColor int, font *truetype.Font, fontSize float64,
	fontColor int, text string) (*wImg, int, int, error) {

	// rew and reh are the real text extents (since we started at 0, 0)
	rew, reh, err := xgraphics.DrawText(src, 0, 0, colorFromInt(fontColor),
		fontSize, font, text)
	if err != nil {
		logWarning.Printf("Could not draw text '%s' because: %v", text, err)
		return nil, 0, 0, err
	}

	return src, rew + renderTextBreathe, reh + renderTextBreathe, nil
}
