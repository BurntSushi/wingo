package frame

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/xgraphics"
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

	f.client.FramePieceMouseSetup("borders_"+ident, win.Id)

	return win
}

func (f *Borders) pieceImages(borderTypes, gradientType, gradientDir,
	width, height int) (*xgraphics.Image, *xgraphics.Image) {

	imgA := render.NewBorder(f.X, borderTypes,
		f.theme.AThinColor, f.theme.ABorderColor,
		width, height, gradientType, gradientDir)
	imgI := render.NewBorder(f.X, borderTypes,
		f.theme.IThinColor, f.theme.IBorderColor,
		width, height, gradientType, gradientDir)
	return imgA.Image, imgI.Image
}

func (f *Borders) cornerImages(borderTypes,
	diagonal int) (*xgraphics.Image, *xgraphics.Image) {

	imgA := render.NewCorner(f.X, borderTypes,
		f.theme.AThinColor, f.theme.ABorderColor,
		f.theme.BorderSize, f.theme.BorderSize,
		diagonal)
	imgI := render.NewCorner(f.X, borderTypes,
		f.theme.IThinColor, f.theme.IBorderColor,
		f.theme.BorderSize, f.theme.BorderSize,
		diagonal)
	return imgA.Image, imgI.Image
}

func (f *Borders) newTopSide() *piece {
	if f.theme.BorderSize == 0 {
		return newEmptyPiece()
	}

	pixA, pixI := f.pieceImages(render.BorderTop,
		render.GradientVert, render.GradientRegular,
		1, f.theme.BorderSize)
	win := f.newPieceWindow("top", cursors.TopSide)
	win.MROpt(fX|fY|fH, f.theme.BorderSize, 0,
		0, f.theme.BorderSize)
	return newPiece(win, pixA, pixI)
}

func (f *Borders) newBottomSide() *piece {
	if f.theme.BorderSize == 0 {
		return newEmptyPiece()
	}

	pixA, pixI := f.pieceImages(render.BorderBottom,
		render.GradientVert, render.GradientReverse,
		1, f.theme.BorderSize)
	win := f.newPieceWindow("bottom", cursors.BottomSide)
	win.MROpt(fX|fH,
		f.theme.BorderSize, 0, 0, f.theme.BorderSize)
	return newPiece(win, pixA, pixI)
}

func (f *Borders) newLeftSide() *piece {
	if f.theme.BorderSize == 0 {
		return newEmptyPiece()
	}

	pixA, pixI := f.pieceImages(render.BorderLeft,
		render.GradientHorz, render.GradientRegular,
		f.theme.BorderSize, 1)
	win := f.newPieceWindow("left", cursors.LeftSide)
	win.MROpt(fX|fY|fW,
		0, f.theme.BorderSize, f.theme.BorderSize, 0)
	return newPiece(win, pixA, pixI)
}

func (f *Borders) newRightSide() *piece {
	if f.theme.BorderSize == 0 {
		return newEmptyPiece()
	}

	pixA, pixI := f.pieceImages(render.BorderRight,
		render.GradientHorz, render.GradientReverse,
		f.theme.BorderSize, 1)
	win := f.newPieceWindow("right", cursors.RightSide)
	win.MROpt(fY|fW,
		0, f.theme.BorderSize, f.theme.BorderSize, 0)
	return newPiece(win, pixA, pixI)
}

func (f *Borders) newTopLeft() *piece {
	if f.theme.BorderSize == 0 {
		return newEmptyPiece()
	}

	pixA, pixI := f.cornerImages(render.BorderTop|render.BorderLeft,
		render.DiagTopLeft)
	win := f.newPieceWindow("topleft", cursors.TopLeftCorner)
	win.MROpt(fX|fY|fW|fH,
		0, 0, f.theme.BorderSize, f.theme.BorderSize)
	return newPiece(win, pixA, pixI)
}

func (f *Borders) newTopRight() *piece {
	if f.theme.BorderSize == 0 {
		return newEmptyPiece()
	}

	pixA, pixI := f.cornerImages(render.BorderTop|render.BorderRight,
		render.DiagTopRight)
	win := f.newPieceWindow("topright", cursors.TopRightCorner)
	win.MROpt(fY|fW|fH,
		0, 0, f.theme.BorderSize, f.theme.BorderSize)
	return newPiece(win, pixA, pixI)
}

func (f *Borders) newBottomLeft() *piece {
	if f.theme.BorderSize == 0 {
		return newEmptyPiece()
	}

	pixA, pixI := f.cornerImages(render.BorderBottom|render.BorderLeft,
		render.DiagBottomLeft)
	win := f.newPieceWindow("bottomleft", cursors.BottomLeftCorner)
	win.MROpt(fX|fW|fH,
		0, 0, f.theme.BorderSize, f.theme.BorderSize)
	return newPiece(win, pixA, pixI)
}

func (f *Borders) newBottomRight() *piece {
	if f.theme.BorderSize == 0 {
		return newEmptyPiece()
	}

	pixA, pixI := f.cornerImages(render.BorderBottom|render.BorderRight,
		render.DiagBottomRight)
	win := f.newPieceWindow("bottomright", cursors.BottomRightCorner)
	win.MROpt(fW|fH,
		0, 0, f.theme.BorderSize, f.theme.BorderSize)
	return newPiece(win, pixA, pixI)
}
