package frame

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/cursors"
	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/render"
	"github.com/BurntSushi/wingo/theme"
)

func (f *Full) newPieceWindow(ident string,
	cursor xproto.Cursor) *xwindow.Window {

	win, err := xwindow.Generate(f.X)
	if err != nil {
		logger.Error.Printf("Could not create a frame window for client "+
			"with id '%d' because: %s", f.client.Id(), err)
		logger.Error.Fatalf("In a state where no new windows can be created. " +
			"Unfortunately, we must exit.")
	}

	err = win.CreateChecked(f.parent.Id, 0, 0, 1, 1,
		xproto.CwBackPixmap|xproto.CwEventMask|xproto.CwCursor,
		xproto.BackPixmapParentRelative,
		xproto.EventMaskButtonPress|xproto.EventMaskButtonRelease|
			xproto.EventMaskButtonMotion|xproto.EventMaskPointerMotion,
		uint32(cursor))
	if err != nil {
		logger.Warning.Println(err)
	}

	f.client.FramePieceMouseConfig("full_"+ident, win.Id)

	return win
}

func (f *Full) newButtonClose() piece {
	imgA := render.NewBorder(f.X, 0, 0, f.theme.Full.ATitleColor,
		f.theme.Full.TitleSize, f.theme.Full.TitleSize,
		render.GradientVert, render.GradientRegular)
	imgI := render.NewBorder(f.X, 0, 0, f.theme.Full.ITitleColor,
		f.theme.Full.TitleSize, f.theme.Full.TitleSize,
		render.GradientVert, render.GradientRegular)

	xgraphics.BlendOld(imgA, f.theme.Full.ACloseButton, nil, 100, 0, 0)
	xgraphics.BlendOld(imgI, f.theme.Full.ICloseButton, nil, 100, 0, 0)

	win := f.newPieceWindow("close", 0)
	win.MROpt(fY|fW|fH,
		0, f.theme.Full.BorderSize,
		f.theme.Full.TitleSize, f.theme.Full.TitleSize)

	imgA.CreatePixmap()
	imgI.CreatePixmap()
	return newPiece(win, imgA.Pixmap, imgI.Pixmap)
}

func (f *Full) newButtonMaximize() piece {
	imgA := render.NewBorder(f.X, 0, 0, f.theme.Full.ATitleColor,
		f.theme.Full.TitleSize, f.theme.Full.TitleSize,
		render.GradientVert, render.GradientRegular)
	imgI := render.NewBorder(f.X, 0, 0, f.theme.Full.ITitleColor,
		f.theme.Full.TitleSize, f.theme.Full.TitleSize,
		render.GradientVert, render.GradientRegular)

	xgraphics.BlendOld(imgA, f.theme.Full.AMaximizeButton, nil, 100, 0, 0)
	xgraphics.BlendOld(imgI, f.theme.Full.IMaximizeButton, nil, 100, 0, 0)

	win := f.newPieceWindow("maximize", 0)
	win.MROpt(fY|fW|fH,
		0, f.theme.Full.BorderSize,
		f.theme.Full.TitleSize, f.theme.Full.TitleSize)

	imgA.CreatePixmap()
	imgI.CreatePixmap()
	return newPiece(win, imgA.Pixmap, imgI.Pixmap)
}

func (f *Full) newButtonMinimize() piece {
	imgA := render.NewBorder(f.X, 0, 0, f.theme.Full.ATitleColor,
		f.theme.Full.TitleSize, f.theme.Full.TitleSize,
		render.GradientVert, render.GradientRegular)
	imgI := render.NewBorder(f.X, 0, 0, f.theme.Full.ITitleColor,
		f.theme.Full.TitleSize, f.theme.Full.TitleSize,
		render.GradientVert, render.GradientRegular)

	xgraphics.BlendOld(imgA, f.theme.Full.AMinimizeButton, nil, 100, 0, 0)
	xgraphics.BlendOld(imgI, f.theme.Full.IMinimizeButton, nil, 100, 0, 0)

	win := f.newPieceWindow("minimize", 0)
	win.MROpt(fY|fW|fH,
		0, f.theme.Full.BorderSize,
		f.theme.Full.TitleSize, f.theme.Full.TitleSize)

	imgA.CreatePixmap()
	imgI.CreatePixmap()
	return newPiece(win, imgA.Pixmap, imgI.Pixmap)
}

func (f *Full) newTitleBar() piece {
	imgA := render.NewBorder(f.X, 0, 0, f.theme.Full.ATitleColor,
		1, f.theme.Full.TitleSize,
		render.GradientVert, render.GradientRegular)
	imgI := render.NewBorder(f.X, 0, 0, f.theme.Full.ITitleColor,
		1, f.theme.Full.TitleSize,
		render.GradientVert, render.GradientRegular)

	win := f.newPieceWindow("titlebar", 0)
	win.MROpt(fX|fY|fH,
		f.theme.Full.BorderSize, f.theme.Full.BorderSize,
		0, f.theme.Full.TitleSize)

	imgA.CreatePixmap()
	imgI.CreatePixmap()
	return newPiece(win, imgA.Pixmap, imgI.Pixmap)
}

func (f *Full) newTitleText() piece {
	win := f.newPieceWindow("titletext", 0)
	win.MROpt(fX|fY|fH,
		f.theme.Full.BorderSize+f.theme.Full.TitleSize,
		f.theme.Full.BorderSize,
		0, f.theme.Full.TitleSize)

	return newPiece(win, 0, 0)
}

func (f *Full) newIcon() piece {
	win := f.newPieceWindow("icon", 0)
	win.MROpt(fX|fY|fW|fH,
		f.theme.Full.BorderSize, f.theme.Full.BorderSize,
		f.theme.Full.TitleSize, f.theme.Full.TitleSize)
	return newPiece(win, 0, 0)
}

//
// What follows is a simplified version of 'frame_borders_pieces.go'.
// The major simplifying difference is that we don't support gradients
// on the borders of a 'full' frame.
//

func (f *Full) borderImages(width, height int) (xproto.Pixmap, xproto.Pixmap) {
	imgA := render.NewBorder(f.X, 0, 0,
		theme.NewColor(f.theme.Full.ABorderColor), width, height, 0, 0)
	imgI := render.NewBorder(f.X, 0, 0,
		theme.NewColor(f.theme.Full.IBorderColor), width, height, 0, 0)

	imgA.CreatePixmap()
	imgI.CreatePixmap()
	return imgA.Pixmap, imgI.Pixmap
}

func (f *Full) newTopSide() piece {
	pixA, pixI := f.borderImages(1, f.theme.Full.BorderSize)
	win := f.newPieceWindow("top", cursors.TopSide)
	win.MROpt(fX|fY|fH, f.theme.Full.BorderSize, 0, 0, f.theme.Full.BorderSize)
	return newPiece(win, pixA, pixI)
}

func (f *Full) newBottomSide() piece {
	pixA, pixI := f.borderImages(1, f.theme.Full.BorderSize)
	win := f.newPieceWindow("bottom", cursors.BottomSide)
	win.MROpt(fX|fH, f.theme.Full.BorderSize, 0, 0, f.theme.Full.BorderSize)
	return newPiece(win, pixA, pixI)
}

func (f *Full) newLeftSide() piece {
	pixA, pixI := f.borderImages(f.theme.Full.BorderSize, 1)
	win := f.newPieceWindow("left", cursors.LeftSide)
	win.MROpt(fX|fY|fW, 0, f.theme.Full.BorderSize, f.theme.Full.BorderSize, 0)
	return newPiece(win, pixA, pixI)
}

func (f *Full) newRightSide() piece {
	pixA, pixI := f.borderImages(f.theme.Full.BorderSize, 1)
	win := f.newPieceWindow("right", cursors.RightSide)
	win.MROpt(fY|fW, 0, f.theme.Full.BorderSize, f.theme.Full.BorderSize, 0)
	return newPiece(win, pixA, pixI)
}

func (f *Full) newTitleBottom() piece {
	pixA, pixI := f.borderImages(1, f.theme.Full.BorderSize)
	win := f.newPieceWindow("titlebottom", 0)
	win.MROpt(fX|fY|fH,
		f.theme.Full.BorderSize, f.theme.Full.BorderSize+f.theme.Full.TitleSize,
		0, f.theme.Full.BorderSize)
	return newPiece(win, pixA, pixI)
}

func (f *Full) newTopLeft() piece {
	pixA, pixI := f.borderImages(f.theme.Full.BorderSize,
		f.theme.Full.BorderSize)
	win := f.newPieceWindow("topleft", cursors.TopLeftCorner)
	win.MROpt(fX|fY|fW|fH,
		0, 0, f.theme.Full.BorderSize, f.theme.Full.BorderSize)
	return newPiece(win, pixA, pixI)
}

func (f *Full) newTopRight() piece {
	pixA, pixI := f.borderImages(f.theme.Full.BorderSize,
		f.theme.Full.BorderSize)
	win := f.newPieceWindow("topright", cursors.TopRightCorner)
	win.MROpt(fY|fW|fH, 0, 0, f.theme.Full.BorderSize, f.theme.Full.BorderSize)
	return newPiece(win, pixA, pixI)
}

func (f *Full) newBottomLeft() piece {
	pixA, pixI := f.borderImages(f.theme.Full.BorderSize,
		f.theme.Full.BorderSize)
	win := f.newPieceWindow("bottomleft", cursors.BottomLeftCorner)
	win.MROpt(fX|fW|fH, 0, 0, f.theme.Full.BorderSize, f.theme.Full.BorderSize)
	return newPiece(win, pixA, pixI)
}

func (f *Full) newBottomRight() piece {
	pixA, pixI := f.borderImages(f.theme.Full.BorderSize,
		f.theme.Full.BorderSize)
	win := f.newPieceWindow("bottomright", cursors.BottomRightCorner)
	win.MROpt(fW|fH, 0, 0, f.theme.Full.BorderSize, f.theme.Full.BorderSize)
	return newPiece(win, pixA, pixI)
}
