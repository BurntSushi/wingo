package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

import (
    "github.com/BurntSushi/xgbutil/ewmh"
    "github.com/BurntSushi/xgbutil/xgraphics"
)

const (
    BW = 5
    CRNR = BW + 24
)

func (f *frameBorders) newPieceWindow(ident string, cursor xgb.Id,
                                      direction uint32) *window {
    mask := uint32(xgb.CWBackPixmap | xgb.CWEventMask | xgb.CWCursor)
    vals := []uint32{xgb.BackPixmapParentRelative,
                     xgb.EventMaskButtonPress | xgb.EventMaskButtonRelease,
                     uint32(cursor)}
    win := createWindow(f.ParentId(), mask, vals)

    f.Client().framePieceMouseConfig("borders_" + ident, win.id)

    return win
}

func (f *frameBorders) newTopSide() framePiece {
    imgA := xgraphics.Border(xgraphics.BorderTop, 0x0, 0x3366ff, 1, BW)
    imgI := xgraphics.Border(xgraphics.BorderTop, 0, 0, 1, BW)

    return framePiece{
        win: f.newPieceWindow("topside", cursorTopSide, ewmh.SizeTop),
        imgActive: xgraphics.CreatePixmap(X, imgA),
        imgInactive: xgraphics.CreatePixmap(X, imgI),
        xoff: CRNR,
        yoff: 0,
        woff: CRNR * 2,
        hoff: BW,
    }
}

func (f *frameBorders) newTopLeft() framePiece {
    imgA := xgraphics.Border(xgraphics.BorderTop | xgraphics.BorderLeft,
                             0x0, 0x3366ff, CRNR, BW)
    imgI := xgraphics.Border(xgraphics.BorderTop, 0, 0, CRNR, BW)
    return framePiece{
        win: f.newPieceWindow("topleft", cursorTopLeftCorner, ewmh.SizeTopLeft),
        imgActive: xgraphics.CreatePixmap(X, imgA),
        imgInactive: xgraphics.CreatePixmap(X, imgI),
        xoff: 0,
        yoff: 0,
        woff: CRNR,
        hoff: BW,
    }
}

func (f *frameBorders) newTopRight() framePiece {
    imgA := xgraphics.Border(xgraphics.BorderTop | xgraphics.BorderRight,
                             0x0, 0x3366ff, CRNR, BW)
    imgI := xgraphics.Border(xgraphics.BorderTop, 0, 0, CRNR, BW)
    return framePiece{
        win: f.newPieceWindow("topright", cursorTopRightCorner,
                              ewmh.SizeTopRight),
        imgActive: xgraphics.CreatePixmap(X, imgA),
        imgInactive: xgraphics.CreatePixmap(X, imgI),
        xoff: CRNR,
        yoff: 0,
        woff: CRNR,
        hoff: BW,
    }
}

