/*
    Creates resources for the different cursors that Wingo uses.
*/
package main

import "burntsushi.net/go/x-go-binding/xgb"

import "burntsushi.net/go/xgbutil/xcursor"

var (
    cursorLeftPtr xgb.Id
    cursorFleur xgb.Id
    cursorWatch xgb.Id
    cursorTopSide xgb.Id
    cursorTopRightCorner xgb.Id
    cursorRightSide xgb.Id
    cursorBottomRightCorner xgb.Id
    cursorBottomSide xgb.Id
    cursorBottomLeftCorner xgb.Id
    cursorLeftSide xgb.Id
    cursorTopLeftCorner xgb.Id
)

func setupCursors() {
    // lazy...
    cc := func(cursor uint16) xgb.Id {
        return xcursor.CreateCursor(X, cursor)
    }

    cursorLeftPtr = cc(xcursor.LeftPtr)
    cursorFleur = cc(xcursor.Fleur)
    cursorWatch = cc(xcursor.Watch)
    cursorTopSide = cc(xcursor.TopSide)
    cursorTopRightCorner = cc(xcursor.TopRightCorner)
    cursorRightSide = cc(xcursor.RightSide)
    cursorBottomRightCorner = cc(xcursor.BottomRightCorner)
    cursorBottomSide = cc(xcursor.BottomSide)
    cursorBottomLeftCorner = cc(xcursor.BottomLeftCorner)
    cursorLeftSide = cc(xcursor.LeftSide)
    cursorTopLeftCorner = cc(xcursor.TopLeftCorner)
}

