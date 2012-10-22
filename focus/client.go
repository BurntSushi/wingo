package focus

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/xwindow"
)

// Client is the minimal implementation necessary for any particular window
// to be tracked by the focus package.
type Client interface {
	Id() xproto.Window
	Win() *xwindow.Window

	// Called whenever the client should represent themselves as having
	// the input focus.
	Focused()

	// Called whenever the client should represent themselves as NOT
	// having the input focus.
	Unfocused()

	// Returns true if the client should be considered for focus.
	CanFocus() bool

	// Returns true if the client participates in the WM_TAKE_FOCUS
	// protocol as specified by the ICCCM.
	SendFocusNotify() bool

	// Whatever action needs to occur before the client can accept input
	// focus. (Usually showing its workspace or deiconifying.)
	PrepareForFocus()

	// Whether the client believes it has input focus or not.
	IsActive() bool
}
