package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

import (
    "github.com/BurntSushi/xgbutil/ewmh"
    "github.com/BurntSushi/xgbutil/xgraphics"
)

func (f *frameBorders) newPieceWindow(ident string, cursor xgb.Id,
                                      direction uint32) *window {
    mask := uint32(xgb.CWBackPixmap | xgb.CWEventMask | xgb.CWCursor)
    vals := []uint32{xgb.BackPixmapParentRelative,
                     xgb.EventMaskButtonPress | xgb.EventMaskButtonRelease |
                     xgb.EventMaskButtonMotion | xgb.EventMaskPointerMotion,
                     uint32(cursor)}
    win := createWindow(f.ParentId(), mask, vals)

    f.Client().framePieceMouseConfig("borders_" + ident, win.id)

    return win
}

func (f *frameBorders) pieceImages(borderTypes int,
                                   width, height int) (xgb.Id, xgb.Id) {
    imgA := xgraphics.Border(borderTypes,
                             THEME.borders.aThinColor,
                             THEME.borders.aBorderColor,
                             width, height)
    imgI := xgraphics.Border(borderTypes,
                             THEME.borders.iThinColor,
                             THEME.borders.iBorderColor,
                             width, height)
    return xgraphics.CreatePixmap(X, imgA), xgraphics.CreatePixmap(X, imgI)
}

func (f *frameBorders) newTopSide() framePiece {
    pixA, pixI := f.pieceImages(xgraphics.BorderTop,
                                1, THEME.borders.borderSize)
    return framePiece{
        win: f.newPieceWindow("topside", cursorTopSide, ewmh.SizeTop),
        imgActive: pixA,
        imgInactive: pixI,
        xoff: THEME.borders.cornerSize,
        yoff: 0,
        woff: THEME.borders.cornerSize * 2,
        hoff: THEME.borders.borderSize,
    }
}

func (f *frameBorders) newTopLeft() framePiece {
    pixA, pixI := f.pieceImages(xgraphics.BorderTop | xgraphics.BorderLeft,
                                THEME.borders.cornerSize,
                                THEME.borders.borderSize)
    return framePiece{
        win: f.newPieceWindow("topleft", cursorTopLeftCorner, ewmh.SizeTopLeft),
        imgActive: pixA,
        imgInactive: pixI,
        xoff: 0,
        yoff: 0,
        woff: THEME.borders.cornerSize,
        hoff: THEME.borders.borderSize,
    }
}

func (f *frameBorders) newTopRight() framePiece {
    pixA, pixI := f.pieceImages(xgraphics.BorderTop | xgraphics.BorderRight,
                                THEME.borders.cornerSize,
                                THEME.borders.borderSize)
    return framePiece{
        win: f.newPieceWindow("topright", cursorTopRightCorner,
                              ewmh.SizeTopRight),
        imgActive: pixA,
        imgInactive: pixI,
        xoff: THEME.borders.cornerSize,
        yoff: 0,
        woff: THEME.borders.cornerSize,
        hoff: THEME.borders.borderSize,
    }
}

func (f *frameBorders) newLeftSide() framePiece {
    pixA, pixI := f.pieceImages(xgraphics.BorderLeft,
                                THEME.borders.borderSize,
                                1)
    return framePiece{
        win: f.newPieceWindow("left", cursorLeftSide, ewmh.SizeLeft),
        imgActive: pixA,
        imgInactive: pixI,
        xoff: 0,
        yoff: THEME.borders.cornerSize,
        woff: THEME.borders.borderSize,
        hoff: THEME.borders.cornerSize * 2,
    }
}

func (f *frameBorders) newLeftTop() framePiece {
    pixA, pixI := f.pieceImages(xgraphics.BorderTop | xgraphics.BorderLeft,
                                THEME.borders.borderSize,
                                THEME.borders.cornerSize)
    return framePiece{
        win: f.newPieceWindow("lefttop", cursorTopLeftCorner, ewmh.SizeTopLeft),
        imgActive: pixA,
        imgInactive: pixI,
        xoff: 0,
        yoff: 0,
        woff: THEME.borders.borderSize,
        hoff: THEME.borders.cornerSize,
    }
}

func (f *frameBorders) newLeftBottom() framePiece {
    pixA, pixI := f.pieceImages(xgraphics.BorderBottom | xgraphics.BorderLeft,
                                THEME.borders.borderSize,
                                THEME.borders.cornerSize)
    return framePiece{
        win: f.newPieceWindow("leftbottom", cursorBottomLeftCorner,
                              ewmh.SizeBottomLeft),
        imgActive: pixA,
        imgInactive: pixI,
        xoff: 0,
        yoff: THEME.borders.cornerSize,
        woff: THEME.borders.borderSize,
        hoff: THEME.borders.cornerSize,
    }
}

func (f *frameBorders) newBottomSide() framePiece {
    pixA, pixI := f.pieceImages(xgraphics.BorderBottom,
                                1, THEME.borders.borderSize)
    return framePiece{
        win: f.newPieceWindow("bottomside", cursorBottomSide, ewmh.SizeBottom),
        imgActive: pixA,
        imgInactive: pixI,
        xoff: THEME.borders.cornerSize,
        yoff: THEME.borders.borderSize,
        woff: THEME.borders.cornerSize * 2,
        hoff: THEME.borders.borderSize,
    }
}

func (f *frameBorders) newBottomLeft() framePiece {
    pixA, pixI := f.pieceImages(xgraphics.BorderBottom | xgraphics.BorderLeft,
                                THEME.borders.cornerSize,
                                THEME.borders.borderSize)
    return framePiece{
        win: f.newPieceWindow("bottomleft", cursorBottomLeftCorner,
                              ewmh.SizeBottomLeft),
        imgActive: pixA,
        imgInactive: pixI,
        xoff: 0,
        yoff: THEME.borders.cornerSize,
        woff: THEME.borders.cornerSize,
        hoff: THEME.borders.borderSize,
    }
}

func (f *frameBorders) newBottomRight() framePiece {
    pixA, pixI := f.pieceImages(xgraphics.BorderBottom | xgraphics.BorderRight,
                                THEME.borders.cornerSize,
                                THEME.borders.borderSize)
    return framePiece{
        win: f.newPieceWindow("bottomright", cursorBottomRightCorner,
                              ewmh.SizeBottomRight),
        imgActive: pixA,
        imgInactive: pixI,
        xoff: THEME.borders.cornerSize,
        yoff: THEME.borders.cornerSize,
        woff: THEME.borders.cornerSize,
        hoff: THEME.borders.borderSize,
    }
}

func (f *frameBorders) newRightSide() framePiece {
    pixA, pixI := f.pieceImages(xgraphics.BorderRight,
                                THEME.borders.borderSize,
                                1)
    return framePiece{
        win: f.newPieceWindow("right", cursorRightSide, ewmh.SizeRight),
        imgActive: pixA,
        imgInactive: pixI,
        xoff: THEME.borders.borderSize,
        yoff: THEME.borders.cornerSize,
        woff: THEME.borders.borderSize,
        hoff: THEME.borders.cornerSize * 2,
    }
}

func (f *frameBorders) newRightTop() framePiece {
    pixA, pixI := f.pieceImages(xgraphics.BorderTop | xgraphics.BorderRight,
                                THEME.borders.borderSize,
                                THEME.borders.cornerSize)
    return framePiece{
        win: f.newPieceWindow("righttop", cursorTopRightCorner,
                              ewmh.SizeTopRight),
        imgActive: pixA,
        imgInactive: pixI,
        xoff: 0,
        yoff: 0,
        woff: THEME.borders.borderSize,
        hoff: THEME.borders.cornerSize,
    }
}

func (f *frameBorders) newRightBottom() framePiece {
    pixA, pixI := f.pieceImages(xgraphics.BorderBottom | xgraphics.BorderRight,
                                THEME.borders.borderSize,
                                THEME.borders.cornerSize)
    return framePiece{
        win: f.newPieceWindow("rightbottom", cursorBottomRightCorner,
                              ewmh.SizeBottomRight),
        imgActive: pixA,
        imgInactive: pixI,
        xoff: 0,
        yoff: THEME.borders.cornerSize,
        woff: THEME.borders.borderSize,
        hoff: THEME.borders.cornerSize,
    }
}

