package prompt

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/text"
)

// CycleChoice is any value capable of being shown in a prompt cycle.
type CycleChoice interface {
	// CycleIsActive should return whether the particular choice is "active" or
	// not. This is called every time the cycle prompt is displayed. In the
	// typical "alt-tab" example, this returns false when the window is
	// iconified (minimized). When this is false, the an "inactive" image is
	// used instead. (Which is a image with transparency equal to the
	// IconTransparency CycleTheme option.)
	CycleIsActive() bool

	// CycleImage returns the image used for the choice. (Both the active and
	// inactive images are built from this value.)
	// Note that it is okay for this method to be slow. It is only called
	// when a CycleChoice is added to the cycle prompt or when
	// (*CycleItem).UpdateImage is called. (So no image operations take place
	// when the cycle prompt is actually shown.)
	CycleImage() *xgraphics.Image

	// CycleText returns the text representing this choice. It can be empty.
	CycleText() string

	// CycleSelected is a hook that is called when this choice is chosen in the
	// cycle prompt.
	CycleSelected()

	// CycleHighlighted is a hook that is called when this choice is highlighted
	// in the cycle prompt.
	CycleHighlighted()
}

// CycleItem is a representation of a CycleChoice that is amenable to being
// displayed in a cycle prompt. A CycleItem value is created and returned to
// the caller whenever (*Cycle).AddItem is called.
//
// CycleItem values are used as a parameter to Show to dictate which choices
// are displayed in a viewing of a prompt.
//
// Also, the Image and Text corresponding to this item can be updated using
// UpdateImage and UpdateText.
//
// Finally, when a CycleChoice (and by extension, a CycleItem) is no longer in
// use, (*CycleItem).Destroy should be called. (This will destroy all X windows
// associated with the CycleItem.) Forgetting to call Destroy will result in
// X resources (window identifiers) not being freed until your connection
// is closed.
type CycleItem struct {
	cycle  *Cycle
	choice CycleChoice

	win              *xwindow.Window
	active, inactive *xwindow.Window
	text             *xwindow.Window
}

// newCycleItem sets up the windows and images associated with a particular
// CycleChoice. This is the meat of (*Cycle).AddItem.
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

// show positions the CycleItem according to the parameters and maps either
// the active or inactive image depening upon the choice's CycleIsActive.
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

// hide simply unmaps the parent window for this item.
func (ci *CycleItem) hide() {
	ci.win.Unmap()
	ci.text.Unmap()
}

// choose selects this choice.
func (ci *CycleItem) choose() {
	ci.choice.CycleSelected()
}

// highlight highlights this choice.
func (ci *CycleItem) highlight() {
	ci.choice.CycleHighlighted()
	ci.text.Map()
	ci.win.Change(xproto.CwBackPixel, ci.cycle.theme.BorderColor.Uint32())
	ci.win.ClearAll()
}

// unhighlight cancels any highlight associated with this choice.
func (ci *CycleItem) unhighlight() {
	ci.text.Unmap()
	ci.win.Change(xproto.CwBackPixel, ci.cycle.theme.BgColor.Uint32())
	ci.win.ClearAll()
}

// Destroy destroys all windows associated with the CycleItem. This is necessary
// to free X resources and should be called whenever the CycleItem will no
// longer be used.
func (ci *CycleItem) Destroy() {
	ci.active.Destroy()
	ci.inactive.Destroy()
	ci.text.Destroy()
	ci.win.Destroy()
}

// UpdateImage will repaint the active and inactive images by calling
// CycleChoice.CycleImage. This is not called when the cycle prompt is shown;
// rather the burden is on the user to make sure the prompt has the most up
// to date image.
func (ci *CycleItem) UpdateImage() {
	active, inactive := ci.choice.CycleImage(), ci.choice.CycleImage()

	xgraphics.Alpha(inactive, ci.cycle.theme.IconTransparency)

	xgraphics.BlendBgColor(active, ci.cycle.theme.BgColor.ImageColor())
	xgraphics.BlendBgColor(inactive, ci.cycle.theme.BgColor.ImageColor())

	active.XSurfaceSet(ci.active.Id)
	active.XDraw()
	active.XPaint(ci.active.Id)
	active.Destroy()

	inactive.XSurfaceSet(ci.inactive.Id)
	inactive.XDraw()
	inactive.XPaint(ci.inactive.Id)
	inactive.Destroy()
}

// UpdateText repaints the text to an image associated with a particular
// CycleChoice. The text is retrieved by calling CycleChoice.CycleText.
func (ci *CycleItem) UpdateText() {
	t := ci.cycle.theme
	txt := ci.choice.CycleText()

	err := text.DrawText(ci.text, t.Font, t.FontSize, t.FontColor,
		t.BgColor, txt)
	if err != nil {
		logger.Warning.Printf("(*CycleItem).UpdateText: "+
			"Could not render text: %s", err)
	}
}
