package cursors

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xcursor"

	"github.com/BurntSushi/wingo/logger"
)

var (
	LeftPtr           xproto.Cursor
	Fleur             xproto.Cursor
	Watch             xproto.Cursor
	TopSide           xproto.Cursor
	TopRightCorner    xproto.Cursor
	RightSide         xproto.Cursor
	BottomRightCorner xproto.Cursor
	BottomSide        xproto.Cursor
	BottomLeftCorner  xproto.Cursor
	LeftSide          xproto.Cursor
	TopLeftCorner     xproto.Cursor
)

func SetupCursors(X *xgbutil.XUtil) {
	// lazy...
	cc := func(cursor uint16) xproto.Cursor {
		cid, err := xcursor.CreateCursor(X, cursor)
		if err != nil {
			logger.Warning.Printf("Could not load cursor '%d'.", cursor)
			return 0
		}
		return cid
	}

	LeftPtr = cc(xcursor.LeftPtr)
	Fleur = cc(xcursor.Fleur)
	Watch = cc(xcursor.Watch)
	TopSide = cc(xcursor.TopSide)
	TopRightCorner = cc(xcursor.TopRightCorner)
	RightSide = cc(xcursor.RightSide)
	BottomRightCorner = cc(xcursor.BottomRightCorner)
	BottomSide = cc(xcursor.BottomSide)
	BottomLeftCorner = cc(xcursor.BottomLeftCorner)
	LeftSide = cc(xcursor.LeftSide)
	TopLeftCorner = cc(xcursor.TopLeftCorner)
}
