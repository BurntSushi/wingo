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

func (f *frameBorders) newPieceWindow(cursor xgb.Id, direction uint32) *window {
    mask := uint32(xgb.CWBackPixmap | xgb.CWEventMask | xgb.CWCursor)
    vals := []uint32{xgb.BackPixmapParentRelative,
                     xgb.EventMaskButtonPress | xgb.EventMaskButtonRelease,
                     uint32(cursor)}
    wid := createWindow(f.ParentId(), mask, vals)

    // don't forget to attach some event handlers

    return wid
}

func (f *frameBorders) newTopSide() framePiece {
    imgA := xgraphics.Border(xgraphics.BorderTop, 0x00ff00, 0x3366ff, 1, BW)
    imgIa := xgraphics.Border(xgraphics.BorderTop, 0, 0, 1, BW)
    return framePiece{
        w: f.newPieceWindow(cursorTopSide, ewmh.SizeTop),
        imgActive: xgraphics.CreatePixmap(X, imgA),
        imgInactive: xgraphics.CreatePixmap(X, imgIa),
        xoff: CRNR,
        yoff: 0,
        woff: CRNR * 2,
        hoff: BW,
    }
}

