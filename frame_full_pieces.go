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

func (f *frameFull) newTitleBar() framePiece {
    imgA := renderBorder(0, 0, THEME.full.aTitleColor,
                         1, THEME.full.titleSize,
                         renderGradientVert, renderGradientRegular)
    imgI := renderBorder(0, 0, THEME.full.iTitleColor,
                         1, THEME.full.titleSize,
                         renderGradientVert, renderGradientRegular)

    win := f.newPieceWindow("titlebar", 0)
    win.moveresize(DoX | DoY | DoH,
                   THEME.full.borderSize,
                   THEME.full.borderSize,
                   0, THEME.full.titleSize)
    return newFramePiece(win, xgraphics.CreatePixmap(X, imgA),
                         xgraphics.CreatePixmap(X, imgI))
}

func (f *frameFull) newTitleText() framePiece {
    title := f.Client().Name()
    font := THEME.full.font
    fontSize := THEME.full.fontSize
    fontColor := ColorFromInt(THEME.full.fontColor)

    ew, eh, err := xgraphics.TextExtents(font, fontSize, title)
    if err != nil {
        logWarning.Printf("Could not get text extents for name '%s' on " +
                          "window %s. Resorting to default width of 300.",
                          title, f.Client())
        ew = 300
    }

    // XXX: We still can't send images with more pixels than 256x256.
    // This is a point where that limitation is very easy to surpass if
    // we have long window titles. Do a sanity check here and bail on the
    // window title if X is going to stomp on us.
    if ew * THEME.full.titleSize > 255 * 255 {
        logWarning.Printf("The image containing the window title is just too " +
                          "big for XGB to handle. I really hope to fix this " +
                          "soon. Falling back to 'N/A' for now...")
        title = "N/A"
        ew, eh, err = xgraphics.TextExtents(font, fontSize, title)
        if err != nil {
            logWarning.Printf("Could not get text extents for name '%s' on " +
                              "window %s. Resorting to default width of 100.",
                              title, f.Client())
            ew = 100
        }
    }

    imgA := renderBorder(0, 0, THEME.full.aTitleColor,
                         ew, THEME.full.titleSize,
                         renderGradientVert, renderGradientRegular)
    imgI := renderBorder(0, 0, THEME.full.iTitleColor,
                         ew, THEME.full.titleSize,
                         renderGradientVert, renderGradientRegular)

    y := (THEME.full.titleSize - eh) / 2 - 1
    xgraphics.DrawText(imgA, 0, y, fontColor, fontSize, font, title)
    xgraphics.DrawText(imgI, 0, y, fontColor, fontSize, font, title)

    win := f.newPieceWindow("titletext", 0)
    win.moveresize(DoX | DoY | DoW | DoH,
                   THEME.full.borderSize + THEME.full.titleSize,
                   THEME.full.borderSize,
                   ew, THEME.full.titleSize)
    return newFramePiece(win, xgraphics.CreatePixmap(X, imgA),
                         xgraphics.CreatePixmap(X, imgI))
}

func (f *frameFull) newIcon() framePiece {
    imgA := renderBorder(0, 0, THEME.full.aTitleColor,
                         THEME.full.titleSize, THEME.full.titleSize,
                         renderGradientVert, renderGradientRegular)
    imgI := renderBorder(0, 0, THEME.full.iTitleColor,
                         THEME.full.titleSize, THEME.full.titleSize,
                         renderGradientVert, renderGradientRegular)

    img, msk := f.Client().iconImage(THEME.full.titleSize - 4,
                                     THEME.full.titleSize - 4)
    xgraphics.Blend(imgA, img, msk, 100, 2, 2)
    xgraphics.Blend(imgI, img, msk, 100, 2, 2)

    win := f.newPieceWindow("icon", 0)
    win.moveresize(DoX | DoY | DoW | DoH,
                   THEME.full.borderSize, THEME.full.borderSize,
                   THEME.full.titleSize, THEME.full.titleSize)
    return newFramePiece(win, xgraphics.CreatePixmap(X, imgA),
                         xgraphics.CreatePixmap(X, imgI))
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

