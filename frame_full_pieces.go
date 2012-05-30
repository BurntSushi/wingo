package main

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xwindow"
)

func (f *frameFull) newPieceWindow(ident string,
	cursor xproto.Cursor) *xwindow.Window {

	mask := xproto.CwBackPixmap | xproto.CwEventMask | xproto.CwCursor
	vals := []uint32{xproto.BackPixmapParentRelative,
		xproto.EventMaskButtonPress | xproto.EventMaskButtonRelease |
			xproto.EventMaskButtonMotion | xproto.EventMaskPointerMotion,
		uint32(cursor)}
	win := createWindow(f.ParentId(), mask, vals...)

	f.Client().framePieceMouseConfig("full_"+ident, win.id)

	return win
}

func (f *frameFull) newButtonClose() framePiece {
	imgA := renderBorder(0, 0, THEME.full.aTitleColor,
		THEME.full.titleSize, THEME.full.titleSize,
		renderGradientVert, renderGradientRegular)
	imgI := renderBorder(0, 0, THEME.full.iTitleColor,
		THEME.full.titleSize, THEME.full.titleSize,
		renderGradientVert, renderGradientRegular)

	xgraphics.BlendOld(imgA, THEME.full.aCloseButton, nil, 100, 0, 0)
	xgraphics.BlendOld(imgI, THEME.full.iCloseButton, nil, 100, 0, 0)

	win := f.newPieceWindow("close", 0)
	win.moveresize(DoY|DoW|DoH,
		0, THEME.full.borderSize,
		THEME.full.titleSize, THEME.full.titleSize)
	return newFramePiece(win, xgraphics.CreatePixmap(X, imgA),
		xgraphics.CreatePixmap(X, imgI))
}

func (f *frameFull) newButtonMaximize() framePiece {
	imgA := renderBorder(0, 0, THEME.full.aTitleColor,
		THEME.full.titleSize, THEME.full.titleSize,
		renderGradientVert, renderGradientRegular)
	imgI := renderBorder(0, 0, THEME.full.iTitleColor,
		THEME.full.titleSize, THEME.full.titleSize,
		renderGradientVert, renderGradientRegular)

	xgraphics.BlendOld(imgA, THEME.full.aMaximizeButton, nil, 100, 0, 0)
	xgraphics.BlendOld(imgI, THEME.full.iMaximizeButton, nil, 100, 0, 0)

	win := f.newPieceWindow("maximize", 0)
	win.moveresize(DoY|DoW|DoH,
		0, THEME.full.borderSize,
		THEME.full.titleSize, THEME.full.titleSize)
	return newFramePiece(win, xgraphics.CreatePixmap(X, imgA),
		xgraphics.CreatePixmap(X, imgI))
}

func (f *frameFull) newButtonMinimize() framePiece {
	imgA := renderBorder(0, 0, THEME.full.aTitleColor,
		THEME.full.titleSize, THEME.full.titleSize,
		renderGradientVert, renderGradientRegular)
	imgI := renderBorder(0, 0, THEME.full.iTitleColor,
		THEME.full.titleSize, THEME.full.titleSize,
		renderGradientVert, renderGradientRegular)

	xgraphics.BlendOld(imgA, THEME.full.aMinimizeButton, nil, 100, 0, 0)
	xgraphics.BlendOld(imgI, THEME.full.iMinimizeButton, nil, 100, 0, 0)

	win := f.newPieceWindow("minimize", 0)
	win.moveresize(DoY|DoW|DoH,
		0, THEME.full.borderSize,
		THEME.full.titleSize, THEME.full.titleSize)
	return newFramePiece(win, xgraphics.CreatePixmap(X, imgA),
		xgraphics.CreatePixmap(X, imgI))
}

func (f *frameFull) newTitleBar() framePiece {
	imgA := renderBorder(0, 0, THEME.full.aTitleColor,
		1, THEME.full.titleSize,
		renderGradientVert, renderGradientRegular)
	imgI := renderBorder(0, 0, THEME.full.iTitleColor,
		1, THEME.full.titleSize,
		renderGradientVert, renderGradientRegular)

	win := f.newPieceWindow("titlebar", 0)
	win.moveresize(DoX|DoY|DoH,
		THEME.full.borderSize,
		THEME.full.borderSize,
		0, THEME.full.titleSize)
	return newFramePiece(win, xgraphics.CreatePixmap(X, imgA),
		xgraphics.CreatePixmap(X, imgI))
}

func (f *frameFull) newTitleText() framePiece {
	win := f.newPieceWindow("titletext", 0)
	win.moveresize(DoX|DoY|DoH,
		THEME.full.borderSize+THEME.full.titleSize,
		THEME.full.borderSize,
		0, THEME.full.titleSize)
	return newFramePiece(win, 0, 0)
}

func (f *frameFull) newIcon() framePiece {
	win := f.newPieceWindow("icon", 0)
	win.moveresize(DoX|DoY|DoW|DoH,
		THEME.full.borderSize, THEME.full.borderSize,
		THEME.full.titleSize, THEME.full.titleSize)
	return newFramePiece(win, 0, 0)
}

//
// What follows is a simplified version of 'frame_borders_pieces.go'.
// The major simplifying difference is that we don't support gradients
// on the borders of a 'full' frame.
//

func (f *frameFull) borderImages(
	width, height int) (xproto.Pixmap, xproto.Pixmap) {

	imgA := renderBorder(0, 0, newThemeColor(THEME.full.aBorderColor),
		width, height, 0, 0)
	imgI := renderBorder(0, 0, newThemeColor(THEME.full.iBorderColor),
		width, height, 0, 0)
	return xgraphics.CreatePixmap(X, imgA), xgraphics.CreatePixmap(X, imgI)
}

func (f *frameFull) newTopSide() framePiece {
	pixA, pixI := f.borderImages(1, THEME.full.borderSize)
	win := f.newPieceWindow("top", cursorTopSide)
	win.moveresize(DoX|DoY|DoH,
		THEME.full.borderSize, 0, 0, THEME.full.borderSize)
	return newFramePiece(win, pixA, pixI)
}

func (f *frameFull) newBottomSide() framePiece {
	pixA, pixI := f.borderImages(1, THEME.full.borderSize)
	win := f.newPieceWindow("bottom", cursorBottomSide)
	win.moveresize(DoX|DoH,
		THEME.full.borderSize, 0, 0, THEME.full.borderSize)
	return newFramePiece(win, pixA, pixI)
}

func (f *frameFull) newLeftSide() framePiece {
	pixA, pixI := f.borderImages(THEME.full.borderSize, 1)
	win := f.newPieceWindow("left", cursorLeftSide)
	win.moveresize(DoX|DoY|DoW,
		0, THEME.full.borderSize, THEME.full.borderSize, 0)
	return newFramePiece(win, pixA, pixI)
}

func (f *frameFull) newRightSide() framePiece {
	pixA, pixI := f.borderImages(THEME.full.borderSize, 1)
	win := f.newPieceWindow("right", cursorRightSide)
	win.moveresize(DoY|DoW,
		0, THEME.full.borderSize, THEME.full.borderSize, 0)
	return newFramePiece(win, pixA, pixI)
}

func (f *frameFull) newTitleBottom() framePiece {
	pixA, pixI := f.borderImages(1, THEME.full.borderSize)
	win := f.newPieceWindow("titlebottom", 0)
	win.moveresize(DoX|DoY|DoH,
		THEME.full.borderSize,
		THEME.full.borderSize+THEME.full.titleSize,
		0, THEME.full.borderSize)
	return newFramePiece(win, pixA, pixI)
}

func (f *frameFull) newTopLeft() framePiece {
	pixA, pixI := f.borderImages(THEME.full.borderSize, THEME.full.borderSize)
	win := f.newPieceWindow("topleft", cursorTopLeftCorner)
	win.moveresize(DoX|DoY|DoW|DoH,
		0, 0,
		THEME.full.borderSize, THEME.full.borderSize)
	return newFramePiece(win, pixA, pixI)
}

func (f *frameFull) newTopRight() framePiece {
	pixA, pixI := f.borderImages(THEME.full.borderSize, THEME.full.borderSize)
	win := f.newPieceWindow("topright", cursorTopRightCorner)
	win.moveresize(DoY|DoW|DoH,
		0, 0,
		THEME.full.borderSize, THEME.full.borderSize)
	return newFramePiece(win, pixA, pixI)
}

func (f *frameFull) newBottomLeft() framePiece {
	pixA, pixI := f.borderImages(THEME.full.borderSize, THEME.full.borderSize)
	win := f.newPieceWindow("bottomleft", cursorBottomLeftCorner)
	win.moveresize(DoX|DoW|DoH,
		0, 0,
		THEME.full.borderSize, THEME.full.borderSize)
	return newFramePiece(win, pixA, pixI)
}

func (f *frameFull) newBottomRight() framePiece {
	pixA, pixI := f.borderImages(THEME.full.borderSize, THEME.full.borderSize)
	win := f.newPieceWindow("bottomright", cursorBottomRightCorner)
	win.moveresize(DoW|DoH,
		0, 0,
		THEME.full.borderSize, THEME.full.borderSize)
	return newFramePiece(win, pixA, pixI)
}
