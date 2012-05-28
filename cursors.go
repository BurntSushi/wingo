/*
   Creates resources for the different cursors that Wingo uses.
*/
package main

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/xcursor"

	"github.com/BurntSushi/wingo/logger"
)

var (
	cursorLeftPtr           xproto.Cursor
	cursorFleur             xproto.Cursor
	cursorWatch             xproto.Cursor
	cursorTopSide           xproto.Cursor
	cursorTopRightCorner    xproto.Cursor
	cursorRightSide         xproto.Cursor
	cursorBottomRightCorner xproto.Cursor
	cursorBottomSide        xproto.Cursor
	cursorBottomLeftCorner  xproto.Cursor
	cursorLeftSide          xproto.Cursor
	cursorTopLeftCorner     xproto.Cursor
)

func setupCursors() {
	// lazy...
	cc := func(cursor uint16) xproto.Cursor {
		cid, err := xcursor.CreateCursor(X, cursor)
		if err != nil {
			logger.Warning.Printf("Could not load cursor '%d'.", cursor)
			return 0
		}
		return cid
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
