package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

import (
    "github.com/BurntSushi/xgbutil/xgraphics"
)

func (f *frameFull) newPieceWindow(ident string, cursor xgb.Id) *window {
    mask := uint32(xgb.CWBackPixmap | xgb.CWEventMask | xgb.CWCursor)
    vals := []uint32{xgb.BackPixmapParentRelative,
                     xgb.EventMaskButtonPress | xgb.EventMaskButtonRelease |
                     xgb.EventMaskButtonMotion | xgb.EventMaskPointerMotion,
                     uint32(cursor)}
    win := createWindow(f.ParentId(), mask, vals)

    f.Client().framePieceMouseConfig("full_" + ident, win.id)

    return win
}

//
// What follows is a simplified version of 'frame_borders_pieces.go'.
// The major simplifying difference is that we don't support gradients
// on the borders of a 'full' frame.
//

func (f *frameFull) borderImages(width, height int) (xgb.Id, xgb.Id) {
    imgA := renderBorder(0, 0, newThemeColor(THEME.full.aBorderColor),
                         width, height, 0, 0)
    imgI := renderBorder(0, 0, newThemeColor(THEME.full.iBorderColor),
                         width, height, 0, 0)
    return xgraphics.CreatePixmap(X, imgA), xgraphics.CreatePixmap(X, imgI)
}

func (f *frameFull) newTopSide() framePiece {
    pixA, pixI := f.borderImages(1, THEME.full.borderSize)
    win := f.newPieceWindow("top", cursorTopSide)
    win.moveresize(DoX | DoY | DoH,
                   THEME.full.borderSize, 0, 0, THEME.full.borderSize)
    return newFramePiece(win, pixA, pixI)
}

func (f *frameFull) newBottomSide() framePiece {
    pixA, pixI := f.borderImages(1, THEME.full.borderSize)
    win := f.newPieceWindow("bottom", cursorBottomSide)
    win.moveresize(DoX | DoH,
                   THEME.full.borderSize, 0, 0, THEME.full.borderSize)
    return newFramePiece(win, pixA, pixI)
}

func (f *frameFull) newLeftSide() framePiece {
    pixA, pixI := f.borderImages(THEME.full.borderSize, 1)
    win := f.newPieceWindow("left", cursorLeftSide)
    win.moveresize(DoX | DoY | DoW,
                   0, THEME.full.borderSize, THEME.full.borderSize, 0)
    return newFramePiece(win, pixA, pixI)
}

func (f *frameFull) newRightSide() framePiece {
    pixA, pixI := f.borderImages(THEME.full.borderSize, 1)
    win := f.newPieceWindow("right", cursorRightSide)
    win.moveresize(DoY | DoW,
                   0, THEME.full.borderSize, THEME.full.borderSize, 0)
    return newFramePiece(win, pixA, pixI)
}

func (f *frameFull) newTopLeft() framePiece {
    pixA, pixI := f.borderImages(THEME.full.borderSize, THEME.full.borderSize)
    win := f.newPieceWindow("topleft", cursorTopLeftCorner)
    win.moveresize(DoX | DoY | DoW | DoH,
                   0, 0,
                   THEME.full.borderSize, THEME.full.borderSize)
    return newFramePiece(win, pixA, pixI)
}

func (f *frameFull) newTopRight() framePiece {
    pixA, pixI := f.borderImages(THEME.full.borderSize, THEME.full.borderSize)
    win := f.newPieceWindow("topright", cursorTopRightCorner)
    win.moveresize(DoY | DoW | DoH,
                   0, 0,
                   THEME.full.borderSize, THEME.full.borderSize)
    return newFramePiece(win, pixA, pixI)
}

func (f *frameFull) newBottomLeft() framePiece {
    pixA, pixI := f.borderImages(THEME.full.borderSize, THEME.full.borderSize)
    win := f.newPieceWindow("bottomleft", cursorBottomLeftCorner)
    win.moveresize(DoX | DoW | DoH,
                   0, 0,
                   THEME.full.borderSize, THEME.full.borderSize)
    return newFramePiece(win, pixA, pixI)
}

func (f *frameFull) newBottomRight() framePiece {
    pixA, pixI := f.borderImages(THEME.full.borderSize, THEME.full.borderSize)
    win := f.newPieceWindow("bottomright", cursorBottomRightCorner)
    win.moveresize(DoW | DoH,
                   0, 0,
                   THEME.full.borderSize, THEME.full.borderSize)
    return newFramePiece(win, pixA, pixI)
}

