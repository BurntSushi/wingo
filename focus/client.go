package focus

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/xwindow"
)

type Client interface {
	Id() xproto.Window
	Win() *xwindow.Window
	Focused()
	Unfocused()
	SetFocus()
	CanFocus() bool
	SendFocusNotify() bool
	MakeViewable()
}
