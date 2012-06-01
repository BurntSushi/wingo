package prompt

import (
	"image"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/misc"
)

// CycleChoice is any value capable of being shown in a prompt cycle.
type CycleChoice interface {
	CycleIsActive() bool
	CycleImage() *xgraphics.Image
	CycleText() string

	CycleSelected()
	CycleHighlighted()
}

// CycleItem is a representation of a CycleChoice that is amenable to being
// displayed in a cycle prompt. A CycleItem value is created and returned to
// the caller whenever (*Cycle).AddItem is called.
//
// CycleItem values are used as a parameter to Show to dictate which choices
// are displayed in a viewing of a prompt.
//
// Also, the Image and Text corresponding to this item can be updated.
type CycleItem struct {
	cycle  *Cycle
	choice CycleChoice

	win              *xwindow.Window
	active, inactive *xwindow.Window
	text             *xwindow.Window
}

func newCycleItem(cycle *Cycle, choice CycleChoice) *CycleItem {
	ci := &CycleItem{
		cycle:  cycle,
		choice: choice,
	}
	t := ci.cycle.theme

	ci.win = xwindow.Must(xwindow.Create(ci.cycle.X, ci.cycle.win.Id))
	ci.active = xwindow.Must(xwindow.Create(ci.cycle.X, ci.win.Id))
	ci.inactive = xwindow.Must(xwindow.Create(ci.cycle.X, ci.win.Id))
	ci.text = xwindow.Must(xwindow.Create(ci.cycle.X, ci.cycle.win.Id))

	ci.active.MoveResize(t.IconBorderSize, t.IconBorderSize,
		t.IconSize, t.IconSize)
	ci.inactive.MoveResize(t.IconBorderSize, t.IconBorderSize,
		t.IconSize, t.IconSize)
	ci.text.MoveResize(t.BorderSize+t.Padding, t.BorderSize+t.Padding,
		1, 1)

	// If the text overruns, make sure it's below the borders.
	ci.text.StackSibling(ci.cycle.bRht.Id, xproto.StackModeBelow)

	ci.UpdateImage()
	ci.UpdateText()

	ci.unhighlight()

	return ci
}

func (ci *CycleItem) show(x, y int) {
	if ci.choice.CycleIsActive() {
		ci.active.Map()
		ci.inactive.Unmap()
	} else {
		ci.inactive.Map()
		ci.active.Unmap()
	}

	is, ibs := ci.cycle.theme.IconSize, ci.cycle.theme.IconBorderSize
	ci.win.MoveResize(x, y, is+2*ibs, is+2*ibs)
	ci.win.Map()
}

func (ci *CycleItem) choose() {
	ci.choice.CycleSelected()
}

func (ci *CycleItem) highlight() {
	ci.choice.CycleHighlighted()
	ci.text.Map()
	ci.win.Change(xproto.CwBackPixel,
		uint32(misc.IntFromColor(ci.cycle.theme.BorderColor)))
	ci.win.ClearAll()
}

func (ci *CycleItem) unhighlight() {
	ci.text.Unmap()
	ci.win.Change(xproto.CwBackPixel,
		uint32(misc.IntFromColor(ci.cycle.theme.BgColor)))
	ci.win.ClearAll()
}

func (ci *CycleItem) UpdateImage() {
	active, inactive := ci.choice.CycleImage(), ci.choice.CycleImage()

	xgraphics.BlendBgColor(active, ci.cycle.theme.BgColor)
	xgraphics.BlendBgColor(inactive, ci.cycle.theme.BgColor)

	active.XSurfaceSet(ci.active.Id)
	active.XDraw()
	active.XPaint(ci.active.Id)
	active.Destroy()

	inactive.XSurfaceSet(ci.inactive.Id)
	inactive.XDraw()
	inactive.XPaint(ci.inactive.Id)
	inactive.Destroy()
}

func (ci *CycleItem) UpdateText() {
	t := ci.cycle.theme
	text := ci.choice.CycleText()

	ewidth, eheight := xgraphics.TextMaxExtents(t.Font, t.FontSize, text)
	eheight += misc.TextBreathe

	img := xgraphics.New(ci.cycle.X, image.Rect(0, 0, ewidth, eheight))
	xgraphics.BlendBgColor(img, t.BgColor)

	x, y, err := img.Text(0, 0, t.FontColor, t.FontSize, t.Font, text)
	if err != nil {
		logger.Warning.Printf("Could not draw text for prompt cycle "+
			"because: %s", err)
		return
	}

	// Use the x,y returned by img.Text to resize the window to the real
	// dimensions and to only draw the appropriate image contents.
	w, h := x, y+misc.TextBreathe
	ci.text.Resize(w, h)

	img.XSurfaceSet(ci.text.Id)
	subimg := img.SubImage(image.Rect(0, 0, w, h))
	subimg.XDraw()
	subimg.XPaint(ci.text.Id)
	img.Destroy()
}
