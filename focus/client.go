package focus

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/workspace"
)

type Client interface {
	Id() xproto.Window
	Win() *xwindow.Window
	IsMapped() bool
	ImminentDestruction() bool
	Focused()
	Unfocused()
	CanFocus() bool
	SendFocusNotify() bool
	PrepareForFocus()
	IsActive() bool
	Workspace() *workspace.Workspace
}
