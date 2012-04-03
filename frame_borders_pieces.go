package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

import (
    "github.com/BurntSushi/xgbutil/xgraphics"
)

func (f *frameBorders) newPieceWindow(ident string, cursor xgb.Id) *window {
    mask := xgb.CWBackPixmap | xgb.CWEventMask | xgb.CWCursor
    vals := []uint32{xgb.BackPixmapParentRelative,
                     xgb.EventMaskButtonPress | xgb.EventMaskButtonRelease |
                     xgb.EventMaskButtonMotion | xgb.EventMaskPointerMotion,
                     uint32(cursor)}
    win := createWindow(f.ParentId(), mask, vals...)

    f.Client().framePieceMouseConfig("borders_" + ident, win.id)

    return win
}

func (f *frameBorders) pieceImages(borderTypes, gradientType, gradientDir,
                                   width, height int) (xgb.Id, xgb.Id) {
    imgA := renderBorder(borderTypes,
                         THEME.borders.aThinColor, THEME.borders.aBorderColor,
                         width, height, gradientType, gradientDir)
    imgI := renderBorder(borderTypes,
                         THEME.borders.iThinColor, THEME.borders.iBorderColor,
                         width, height, gradientType, gradientDir)
    return xgraphics.CreatePixmap(X, imgA), xgraphics.CreatePixmap(X, imgI)
}

func (f *frameBorders) cornerImages(borderTypes,
                                    diagonal int) (xgb.Id, xgb.Id) {
    imgA := renderCorner(borderTypes,
                         THEME.borders.aThinColor, THEME.borders.aBorderColor,
                         THEME.borders.borderSize, THEME.borders.borderSize,
                         diagonal)
    imgI := renderCorner(borderTypes,
                         THEME.borders.iThinColor, THEME.borders.iBorderColor,
                         THEME.borders.borderSize, THEME.borders.borderSize,
                         diagonal)
    return xgraphics.CreatePixmap(X, imgA), xgraphics.CreatePixmap(X, imgI)
}

func (f *frameBorders) newTopSide() framePiece {
    pixA, pixI := f.pieceImages(renderBorderTop,
                                renderGradientVert, renderGradientRegular,
                                1, THEME.borders.borderSize)
    win := f.newPieceWindow("top", cursorTopSide)
    win.moveresize(DoX | DoY | DoH,
                   THEME.borders.borderSize, 0,
                   0, THEME.borders.borderSize)
    return newFramePiece(win, pixA, pixI)
}

func (f *frameBorders) newBottomSide() framePiece {
    pixA, pixI := f.pieceImages(renderBorderBottom,
                                renderGradientVert, renderGradientReverse,
                                1, THEME.borders.borderSize)
    win := f.newPieceWindow("bottom", cursorBottomSide)
    win.moveresize(DoX | DoH,
                   THEME.borders.borderSize, 0, 0, THEME.borders.borderSize)
    return newFramePiece(win, pixA, pixI)
}

func (f *frameBorders) newLeftSide() framePiece {
    pixA, pixI := f.pieceImages(renderBorderLeft,
                                renderGradientHorz, renderGradientRegular,
                                THEME.borders.borderSize, 1)
    win := f.newPieceWindow("left", cursorLeftSide)
    win.moveresize(DoX | DoY | DoW,
                   0, THEME.borders.borderSize,
                   THEME.borders.borderSize, 0)
    return newFramePiece(win, pixA, pixI)
}

func (f *frameBorders) newRightSide() framePiece {
    pixA, pixI := f.pieceImages(renderBorderRight,
                                renderGradientHorz, renderGradientReverse,
                                THEME.borders.borderSize, 1)
    win := f.newPieceWindow("right", cursorRightSide)
    win.moveresize(DoY | DoW,
                   0, THEME.borders.borderSize,
                   THEME.borders.borderSize, 0)
    return newFramePiece(win, pixA, pixI)
}

func (f *frameBorders) newTopLeft() framePiece {
    pixA, pixI := f.cornerImages(renderBorderTop | renderBorderLeft,
                                 renderDiagTopLeft)
    win := f.newPieceWindow("topleft", cursorTopLeftCorner)
    win.moveresize(DoX | DoY | DoW | DoH,
                   0, 0,
                   THEME.borders.borderSize, THEME.borders.borderSize)
    return newFramePiece(win, pixA, pixI)
}

func (f *frameBorders) newTopRight() framePiece {
    pixA, pixI := f.cornerImages(renderBorderTop | renderBorderRight,
                                 renderDiagTopRight)
    win := f.newPieceWindow("topright", cursorTopRightCorner)
    win.moveresize(DoY | DoW | DoH,
                   0, 0,
                   THEME.borders.borderSize, THEME.borders.borderSize)
    return newFramePiece(win, pixA, pixI)
}

func (f *frameBorders) newBottomLeft() framePiece {
    pixA, pixI := f.cornerImages(renderBorderBottom | renderBorderLeft,
                                 renderDiagBottomLeft)
    win := f.newPieceWindow("bottomleft", cursorBottomLeftCorner)
    win.moveresize(DoX | DoW | DoH,
                   0, 0,
                   THEME.borders.borderSize, THEME.borders.borderSize)
    return newFramePiece(win, pixA, pixI)
}

func (f *frameBorders) newBottomRight() framePiece {
    pixA, pixI := f.cornerImages(renderBorderBottom | renderBorderRight,
                                 renderDiagBottomRight)
    win := f.newPieceWindow("bottomright", cursorBottomRightCorner)
    win.moveresize(DoW | DoH,
                   0, 0,
                   THEME.borders.borderSize, THEME.borders.borderSize)
    return newFramePiece(win, pixA, pixI)
}

