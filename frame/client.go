package frame

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"
)

type Client interface {
	State() int
	Frame() Frame
	IsMaximized() bool
	Icon(width, height int) *xgraphics.Image
	Name() string
	ClientGeom() xrect.Rect
	ValidateHeight(height int) int
	ValidateWidth(width int) int
	GravitizeX(x, gravity int) int
	GravitizeY(y, gravity int) int
	Win() *xwindow.Window
	Id() xproto.Window
	EnsureUnmax()
	FramePieceMouseSetup(ident string, wid xproto.Window)
	String() string
	HeadGeom() xrect.Rect
}
