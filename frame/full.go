package frame

import (
	"image"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xgraphics"

	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/misc"
	"github.com/BurntSushi/wingo/render"
	"github.com/BurntSushi/wingo/theme"
)

type Full struct {
	*frame

	titleBar, titleText, icon                   piece
	buttonMinimize, buttonMaximize, buttonClose piece
	topSide, bottomSide, leftSide, rightSide    piece
	topLeft, topRight, bottomLeft, bottomRight  piece
	titleBottom                                 piece
}

func NewFull(X *xgbutil.XUtil,
	t *theme.Theme, p *Parent, c client) (*Full, error) {

	f, err := newFrame(X, t, p, c)
	if err != nil {
		return nil, err
	}

	ff := &Full{frame: f}

	ff.titleBar = ff.newTitleBar()
	ff.titleText = ff.newTitleText()
	ff.buttonClose = ff.newButtonClose()
	ff.buttonMaximize = ff.newButtonMaximize()
	ff.buttonMinimize = ff.newButtonMinimize()
	ff.icon = ff.newIcon()

	if f.theme.Full.BorderSize > 0 {
		ff.topSide = ff.newTopSide()
		ff.bottomSide = ff.newBottomSide()
		ff.leftSide = ff.newLeftSide()
		ff.rightSide = ff.newRightSide()
		ff.titleBottom = ff.newTitleBottom()

		ff.topLeft = ff.newTopLeft()
		ff.topRight = ff.newTopRight()
		ff.bottomLeft = ff.newBottomLeft()
		ff.bottomRight = ff.newBottomRight()
	}

	ff.UpdateTitle()
	ff.UpdateIcon()

	return ff, nil
}

func (f *Full) Current() bool {
	return f.client.Frame() == f
}

func (f *Full) Destroy() {
	if f.theme.Full.BorderSize > 0 {
		f.topSide.Destroy()
		f.bottomSide.Destroy()
		f.leftSide.Destroy()
		f.rightSide.Destroy()
		f.titleBottom.Destroy()

		f.topLeft.Destroy()
		f.topRight.Destroy()
		f.bottomLeft.Destroy()
		f.bottomRight.Destroy()
	}

	f.titleBar.Destroy()
	f.titleText.Destroy()
	f.icon.Destroy()
	f.buttonClose.Destroy()
	f.buttonMaximize.Destroy()
	f.buttonMinimize.Destroy()

	f.frame.Destroy()
}

func (f *Full) Off() {
	if f.theme.Full.BorderSize > 0 {
		f.topSide.Unmap()
		f.bottomSide.Unmap()
		f.leftSide.Unmap()
		f.rightSide.Unmap()
		f.titleBottom.Unmap()

		f.topLeft.Unmap()
		f.topRight.Unmap()
		f.bottomLeft.Unmap()
		f.bottomRight.Unmap()
	}

	f.titleBar.Unmap()
	f.titleText.Unmap()
	f.icon.Unmap()
	f.buttonClose.Unmap()
	f.buttonMaximize.Unmap()
	f.buttonMinimize.Unmap()
}

func (f *Full) On() {
	Reset(f)

	// Make sure the current state is properly shown
	if f.client.State() == Active {
		f.Active()
	} else {
		f.Inactive()
	}

	if f.theme.Full.BorderSize > 0 {
		f.titleBottom.Map()

		if !f.client.Maximized() {
			f.topSide.Map()
			f.bottomSide.Map()
			f.leftSide.Map()
			f.rightSide.Map()

			f.topLeft.Map()
			f.topRight.Map()
			f.bottomLeft.Map()
			f.bottomRight.Map()
		}
	}

	f.titleBar.Map()
	f.titleText.Map()
	f.icon.Map()
	f.buttonClose.Map()
	f.buttonMaximize.Map()
	f.buttonMinimize.Map()
}

func (f *Full) Active() {
	if f.theme.Full.BorderSize > 0 {
		f.topSide.Active()
		f.bottomSide.Active()
		f.leftSide.Active()
		f.rightSide.Active()
		f.titleBottom.Active()

		f.topLeft.Active()
		f.topRight.Active()
		f.bottomLeft.Active()
		f.bottomRight.Active()
	}

	f.titleBar.Active()
	f.titleText.Active()
	f.icon.Active()
	f.buttonClose.Active()
	f.buttonMaximize.Active()
	f.buttonMinimize.Active()

	f.parent.Change(xproto.CwBackPixel, uint32(0xffffff))
	f.parent.ClearAll()
}

func (f *Full) Inactive() {
	if f.theme.Full.BorderSize > 0 {
		f.topSide.Inactive()
		f.bottomSide.Inactive()
		f.leftSide.Inactive()
		f.rightSide.Inactive()
		f.titleBottom.Inactive()

		f.topLeft.Inactive()
		f.topRight.Inactive()
		f.bottomLeft.Inactive()
		f.bottomRight.Inactive()
	}

	f.titleBar.Inactive()
	f.titleText.Inactive()
	f.icon.Inactive()
	f.buttonClose.Inactive()
	f.buttonMaximize.Inactive()
	f.buttonMinimize.Inactive()

	f.parent.Change(xproto.CwBackPixel, uint32(0xffffff))
	f.parent.ClearAll()
}

func (f *Full) Maximize() {
	f.buttonClose.MROpt(fY, 0, 0, 0, 0)
	f.buttonMaximize.MROpt(fY, 0, 0, 0, 0)
	f.buttonMinimize.MROpt(fY, 0, 0, 0, 0)
	f.titleBar.MROpt(fX|fY, 0, 0, 0, 0)
	f.titleText.MROpt(fX|fY, f.theme.Full.TitleSize, 0, 0, 0)
	f.icon.MROpt(fX|fY, 0, 0, 0, 0)
	f.titleBottom.MROpt(fX|fY, 0, f.theme.Full.TitleSize, 0, 0)

	if f.theme.Full.BorderSize > 0 && f.Current() {
		f.topSide.Unmap()
		f.bottomSide.Unmap()
		f.leftSide.Unmap()
		f.rightSide.Unmap()

		f.topLeft.Unmap()
		f.topRight.Unmap()
		f.bottomLeft.Unmap()
		f.bottomRight.Unmap()

		Reset(f)
	}
}

func (f *Full) Unmaximize() {
	f.buttonClose.MROpt(fY, 0, f.theme.Full.BorderSize, 0, 0)
	f.buttonMaximize.MROpt(fY, 0, f.theme.Full.BorderSize, 0, 0)
	f.buttonMinimize.MROpt(fY, 0, f.theme.Full.BorderSize, 0, 0)
	f.titleBar.MROpt(fX|fY, f.theme.Full.BorderSize,
		f.theme.Full.BorderSize, 0, 0)
	f.titleText.MROpt(fX|fY,
		f.theme.Full.BorderSize+f.theme.Full.TitleSize,
		f.theme.Full.BorderSize, 0, 0)
	f.icon.MROpt(fX|fY, f.theme.Full.BorderSize,
		f.theme.Full.BorderSize, 0, 0)
	f.titleBottom.MROpt(fX|fY, f.theme.Full.BorderSize,
		f.theme.Full.BorderSize+f.theme.Full.TitleSize,
		0, 0)

	if f.theme.Full.BorderSize > 0 && f.Current() {
		f.topSide.Map()
		f.bottomSide.Map()
		f.leftSide.Map()
		f.rightSide.Map()

		f.topLeft.Map()
		f.topRight.Map()
		f.bottomLeft.Map()
		f.bottomRight.Map()

		Reset(f)
	}
}

func (f *Full) Top() int {
	if f.client.Maximized() {
		return f.theme.Full.BorderSize + f.theme.Full.TitleSize
	}
	return (f.theme.Full.BorderSize * 2) + f.theme.Full.TitleSize
}

func (f *Full) Bottom() int {
	if f.client.Maximized() {
		return 0
	}
	return f.theme.Full.BorderSize
}

func (f *Full) Left() int {
	if f.client.Maximized() {
		return 0
	}
	return f.theme.Full.BorderSize
}

func (f *Full) Right() int {
	if f.client.Maximized() {
		return 0
	}
	return f.theme.Full.BorderSize
}

func (f *Full) moveresizePieces() {
	fg := f.Geom()

	if f.theme.Full.BorderSize > 0 {
		f.topSide.MROpt(fW, 0, 0, fg.Width()-f.topLeft.w()-f.topRight.w(), 0)
		f.bottomSide.MROpt(
			fY|fW, 0, fg.Height()-f.bottomSide.h(), f.topSide.w(), 0)
		f.leftSide.MROpt(
			fH, 0, 0, 0, fg.Height()-f.topLeft.h()-f.bottomLeft.h())
		f.rightSide.MROpt(
			fX|fH, fg.Width()-f.rightSide.w(), 0, 0, f.leftSide.h())
		f.titleBottom.MROpt(fW, 0, 0, fg.Width()-f.Left()-f.Right(), 0)

		f.topRight.MROpt(fX, f.topLeft.w()+f.topSide.w(), 0, 0, 0)
		f.bottomLeft.MROpt(fY, 0, f.bottomSide.y(), 0, 0)
		f.bottomRight.MROpt(fX|fY,
			f.bottomLeft.w()+f.bottomSide.w(), f.bottomSide.y(), 0, 0)
	}

	f.titleBar.MROpt(fW, 0, 0, fg.Width()-f.Left()-f.Right(), 0)
	f.buttonClose.MROpt(fX, fg.Width()-f.Right()-f.buttonClose.w(), 0, 0, 0)
	f.buttonMaximize.MROpt(fX, f.buttonClose.x()-f.buttonMinimize.w(), 0, 0, 0)
	f.buttonMinimize.MROpt(fX,
		f.buttonMaximize.x()-f.buttonMinimize.w(), 0, 0, 0)
}

func (f *Full) MROpt(validate bool, flags, x, y, w, h int) {
	mropt(f, validate, flags, x, y, w, h)
	f.moveresizePieces()
}

func (f *Full) MoveResize(validate bool, x, y, w, h int) {
	moveresize(f, validate, x, y, w, h)
	f.moveresizePieces()
}

func (f *Full) Move(x, y int) {
	move(f, x, y)
}

func (f *Full) Resize(validate bool, w, h int) {
	resize(f, validate, w, h)
	f.moveresizePieces()
}

func (f *Full) UpdateIcon() {
	size := f.theme.Full.TitleSize
	imgA := render.NewBorder(f.X, 0, 0, f.theme.Full.ATitleColor,
		size, size, render.GradientVert, render.GradientRegular)
	imgI := render.NewBorder(f.X, 0, 0, f.theme.Full.ITitleColor,
		size, size, render.GradientVert, render.GradientRegular)

	img := f.client.Icon(size-4, size-4)

	sub := image.Rect(2, 2, size-2, size-2)
	xgraphics.Blend(imgA.SubImage(sub), img, image.ZP)
	xgraphics.Blend(imgI.SubImage(sub), img, image.ZP)

	if f.icon.active > 0 {
		xgraphics.FreePixmap(f.X, f.icon.active)
	}
	if f.icon.inactive > 0 {
		xgraphics.FreePixmap(f.X, f.icon.inactive)
	}

	imgA.CreatePixmap()
	imgI.CreatePixmap()
	f.icon.active, f.icon.inactive = imgA.Pixmap, imgI.Pixmap

	if f.client.State() == Active {
		f.icon.Active()
	} else {
		f.icon.Inactive()
	}
}

func (f *Full) UpdateTitle() {
	title := f.client.Name()
	font := f.theme.Full.Font
	fontSize := f.theme.Full.FontSize
	aFontColor := misc.ColorFromInt(f.theme.Full.AFontColor)
	iFontColor := misc.ColorFromInt(f.theme.Full.IFontColor)

	ew, eh := xgraphics.TextMaxExtents(font, fontSize, title)

	imgA := render.NewBorder(f.X, 0, 0, f.theme.Full.ATitleColor,
		ew, f.theme.Full.TitleSize,
		render.GradientVert, render.GradientRegular)
	imgI := render.NewBorder(f.X, 0, 0, f.theme.Full.ITitleColor,
		ew, f.theme.Full.TitleSize,
		render.GradientVert, render.GradientRegular)

	y := (f.theme.Full.TitleSize-eh)/2 - 1

	_, _, err := imgA.Text(0, y, aFontColor, fontSize, font, title)
	if err != nil {
		logger.Warning.Printf("Could not draw window title for window %s "+
			"because: %v", f.client, err)
	}

	_, _, err = imgI.Text(0, y, iFontColor, fontSize, font, title)
	if err != nil {
		logger.Warning.Printf("Could not draw window title for window %s "+
			"because: %v", f.client, err)
	}

	if f.titleText.active > 0 {
		xgraphics.FreePixmap(f.X, f.titleText.active)
	}
	if f.titleText.inactive > 0 {
		xgraphics.FreePixmap(f.X, f.titleText.inactive)
	}

	imgA.CreatePixmap()
	imgI.CreatePixmap()
	f.titleText.active, f.titleText.inactive = imgA.Pixmap, imgI.Pixmap

	f.titleText.MROpt(fW, 0, 0, ew, 0)
	if f.client.State() == Active {
		f.titleText.Active()
	} else {
		f.titleText.Inactive()
	}
}
