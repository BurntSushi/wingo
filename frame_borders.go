package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

type frameBorders struct {
	*abstFrame

	// pieces
	topSide, bottomSide, leftSide, rightSide   framePiece
	topLeft, topRight, bottomLeft, bottomRight framePiece
}

func newFrameBorders(p *frameParent, c *client) *frameBorders {
	f := &frameBorders{abstFrame: newFrameAbst(p, c)}

	f.topSide = f.newTopSide()
	f.bottomSide = f.newBottomSide()
	f.leftSide = f.newLeftSide()
	f.rightSide = f.newRightSide()

	f.topLeft = f.newTopLeft()
	f.topRight = f.newTopRight()
	f.bottomLeft = f.newBottomLeft()
	f.bottomRight = f.newBottomRight()

	return f
}

func (f *frameBorders) Current() bool {
	return f.Client().Frame() == f
}

func (f *frameBorders) Destroy() {
	f.topSide.destroy()
	f.bottomSide.destroy()
	f.leftSide.destroy()
	f.rightSide.destroy()

	f.topLeft.destroy()
	f.topRight.destroy()
	f.bottomLeft.destroy()
	f.bottomRight.destroy()

	f.abstFrame.Destroy()
}

func (f *frameBorders) Off() {
	f.topSide.win.unmap()
	f.bottomSide.win.unmap()
	f.leftSide.win.unmap()
	f.rightSide.win.unmap()

	f.topLeft.win.unmap()
	f.topRight.win.unmap()
	f.bottomLeft.win.unmap()
	f.bottomRight.win.unmap()
}

func (f *frameBorders) On() {
	FrameReset(f)

	// Make sure the current state is properly shown
	if f.State() == StateActive {
		f.Active()
	} else {
		f.Inactive()
	}

	if !f.Client().maximized {
		f.topSide.win.map_()
		f.bottomSide.win.map_()
		f.leftSide.win.map_()
		f.rightSide.win.map_()

		f.topLeft.win.map_()
		f.topRight.win.map_()
		f.bottomLeft.win.map_()
		f.bottomRight.win.map_()
	}
}

func (f *frameBorders) Active() {
	f.topSide.active()
	f.bottomSide.active()
	f.leftSide.active()
	f.rightSide.active()

	f.topLeft.active()
	f.topRight.active()
	f.bottomLeft.active()
	f.bottomRight.active()

	f.ParentWin().change(xgb.CWBackPixel, uint32(0xff0000))
	f.ParentWin().clear()
}

func (f *frameBorders) Inactive() {
	f.topSide.inactive()
	f.bottomSide.inactive()
	f.leftSide.inactive()
	f.rightSide.inactive()

	f.topLeft.inactive()
	f.topRight.inactive()
	f.bottomLeft.inactive()
	f.bottomRight.inactive()

	f.ParentWin().change(xgb.CWBackPixel, uint32(0xff0000))
	f.ParentWin().clear()
}

func (f *frameBorders) Maximize() {
	if THEME.full.borderSize > 0 && f.Current() {
		f.topSide.win.unmap()
		f.bottomSide.win.unmap()
		f.leftSide.win.unmap()
		f.rightSide.win.unmap()

		f.topLeft.win.unmap()
		f.topRight.win.unmap()
		f.bottomLeft.win.unmap()
		f.bottomRight.win.unmap()

		FrameReset(f)
	}
}

func (f *frameBorders) Unmaximize() {
	if THEME.full.borderSize > 0 && f.Current() {
		f.topSide.win.map_()
		f.bottomSide.win.map_()
		f.leftSide.win.map_()
		f.rightSide.win.map_()

		f.topLeft.win.map_()
		f.topRight.win.map_()
		f.bottomLeft.win.map_()
		f.bottomRight.win.map_()

		FrameReset(f)
	}
}

func (f *frameBorders) Top() int {
	if f.Client().maximized {
		return 0
	}
	return THEME.borders.borderSize
}

func (f *frameBorders) Bottom() int {
	if f.Client().maximized {
		return 0
	}
	return THEME.borders.borderSize
}

func (f *frameBorders) Left() int {
	if f.Client().maximized {
		return 0
	}
	return THEME.borders.borderSize
}

func (f *frameBorders) Right() int {
	if f.Client().maximized {
		return 0
	}
	return THEME.borders.borderSize
}

func (f *frameBorders) ConfigureClient(flags, x, y, w, h int,
	sibling xgb.Id, stackMode byte, ignoreHints bool) {

	x, y, w, h = FrameConfigureClient(f, flags, x, y, w, h)
	f.ConfigureFrame(flags, x, y, w, h, sibling, stackMode, ignoreHints, true)
}

func (f *frameBorders) ConfigureFrame(flags, fx, fy, fw, fh int,
	sibling xgb.Id, stackMode byte, ignoreHints bool, sendNotify bool) {

	FrameConfigureFrame(f, flags, fx, fy, fw, fh, sibling, stackMode,
		ignoreHints, sendNotify)
	fg := f.Geom()

	f.topSide.win.moveresize(DoW, 0, 0,
		fg.Width()-f.topLeft.w()-f.topRight.w(), 0)
	f.bottomSide.win.moveresize(DoY|DoW,
		0, fg.Height()-f.bottomSide.h(),
		f.topSide.w(), 0)
	f.leftSide.win.moveresize(DoH, 0, 0,
		0, fg.Height()-f.topLeft.h()-f.bottomLeft.h())
	f.rightSide.win.moveresize(DoX|DoH,
		fg.Width()-f.rightSide.w(), 0,
		0, f.leftSide.h())

	f.topRight.win.moveresize(DoX, f.topLeft.w()+f.topSide.w(), 0, 0, 0)
	f.bottomLeft.win.moveresize(DoY, 0, f.bottomSide.y(), 0, 0)
	f.bottomRight.win.moveresize(DoX|DoY,
		f.bottomLeft.w()+f.bottomSide.w(),
		f.bottomSide.y(),
		0, 0)
}
