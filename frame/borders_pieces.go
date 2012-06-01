package frame

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/cursors"
	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/render"
)

func (f *Borders) newPieceWindow(ident string,
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

	f.client.FramePieceMouseConfig("borders_"+ident, win.Id)

	return win
}

func (f *Borders) pieceImages(borderTypes, gradientType, gradientDir,
	width, height int) (xproto.Pixmap, xproto.Pixmap) {

	imgA := render.NewBorder(f.X, borderTypes,
		f.theme.Borders.AThinColor, f.theme.Borders.ABorderColor,
		width, height, gradientType, gradientDir)
	imgI := render.NewBorder(f.X, borderTypes,
		f.theme.Borders.IThinColor, f.theme.Borders.IBorderColor,
		width, height, gradientType, gradientDir)

	imgA.CreatePixmap()
	imgI.CreatePixmap()
	return imgA.Pixmap, imgI.Pixmap
}

func (f *Borders) cornerImages(borderTypes,
	diagonal int) (xproto.Pixmap, xproto.Pixmap) {

	imgA := render.NewCorner(f.X, borderTypes,
		f.theme.Borders.AThinColor, f.theme.Borders.ABorderColor,
		f.theme.Borders.BorderSize, f.theme.Borders.BorderSize,
		diagonal)
	imgI := render.NewCorner(f.X, borderTypes,
		f.theme.Borders.IThinColor, f.theme.Borders.IBorderColor,
		f.theme.Borders.BorderSize, f.theme.Borders.BorderSize,
		diagonal)

	imgA.CreatePixmap()
	imgI.CreatePixmap()
	return imgA.Pixmap, imgI.Pixmap
}

func (f *Borders) newTopSide() piece {
	pixA, pixI := f.pieceImages(render.BorderTop,
		render.GradientVert, render.GradientRegular,
		1, f.theme.Borders.BorderSize)
	win := f.newPieceWindow("top", cursors.TopSide)
	win.MROpt(fX|fY|fH, f.theme.Borders.BorderSize, 0,
		0, f.theme.Borders.BorderSize)
	return newPiece(win, pixA, pixI)
}

func (f *Borders) newBottomSide() piece {
	pixA, pixI := f.pieceImages(render.BorderBottom,
		render.GradientVert, render.GradientReverse,
		1, f.theme.Borders.BorderSize)
	win := f.newPieceWindow("bottom", cursors.BottomSide)
	win.MROpt(fX|fH,
		f.theme.Borders.BorderSize, 0, 0, f.theme.Borders.BorderSize)
	return newPiece(win, pixA, pixI)
}

func (f *Borders) newLeftSide() piece {
	pixA, pixI := f.pieceImages(render.BorderLeft,
		render.GradientHorz, render.GradientRegular,
		f.theme.Borders.BorderSize, 1)
	win := f.newPieceWindow("left", cursors.LeftSide)
	win.MROpt(fX|fY|fW,
		0, f.theme.Borders.BorderSize, f.theme.Borders.BorderSize, 0)
	return newPiece(win, pixA, pixI)
}

func (f *Borders) newRightSide() piece {
	pixA, pixI := f.pieceImages(render.BorderRight,
		render.GradientHorz, render.GradientReverse,
		f.theme.Borders.BorderSize, 1)
	win := f.newPieceWindow("right", cursors.RightSide)
	win.MROpt(fY|fW,
		0, f.theme.Borders.BorderSize, f.theme.Borders.BorderSize, 0)
	return newPiece(win, pixA, pixI)
}

func (f *Borders) newTopLeft() piece {
	pixA, pixI := f.cornerImages(render.BorderTop|render.BorderLeft,
		render.DiagTopLeft)
	win := f.newPieceWindow("topleft", cursors.TopLeftCorner)
	win.MROpt(fX|fY|fW|fH,
		0, 0, f.theme.Borders.BorderSize, f.theme.Borders.BorderSize)
	return newPiece(win, pixA, pixI)
}

func (f *Borders) newTopRight() piece {
	pixA, pixI := f.cornerImages(render.BorderTop|render.BorderRight,
		render.DiagTopRight)
	win := f.newPieceWindow("topright", cursors.TopRightCorner)
	win.MROpt(fY|fW|fH,
		0, 0, f.theme.Borders.BorderSize, f.theme.Borders.BorderSize)
	return newPiece(win, pixA, pixI)
}

func (f *Borders) newBottomLeft() piece {
	pixA, pixI := f.cornerImages(render.BorderBottom|render.BorderLeft,
		render.DiagBottomLeft)
	win := f.newPieceWindow("bottomleft", cursors.BottomLeftCorner)
	win.MROpt(fX|fW|fH,
		0, 0, f.theme.Borders.BorderSize, f.theme.Borders.BorderSize)
	return newPiece(win, pixA, pixI)
}

func (f *Borders) newBottomRight() piece {
	pixA, pixI := f.cornerImages(render.BorderBottom|render.BorderRight,
		render.DiagBottomRight)
	win := f.newPieceWindow("bottomright", cursors.BottomRightCorner)
	win.MROpt(fW|fH,
		0, 0, f.theme.Borders.BorderSize, f.theme.Borders.BorderSize)
	return newPiece(win, pixA, pixI)
}
