package focus

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/logger"
)

var (
	// A global representing the X connection.
	X *xgbutil.XUtil

	// All of the clients tracked by the focus package.
	Clients []Client

	// Set of clients being tracked to avoid too many O(n) operations.
	clientSet map[xproto.Window]bool
)

// Initialize sets up the state for the focus package. It MUST be called
// before any other functions in the package are used.
func Initialize(xu *xgbutil.XUtil) {
	X = xu
	Clients = make([]Client, 0, 100)
	clientSet = make(map[xproto.Window]bool, 100)
}

// Returns the currently focused client, or nil if no client has focus.
func Current() Client {
	if len(Clients) == 0 {
		return nil
	}

	// It's technically possible for no client to have focus.
	possible := Clients[len(Clients)-1]
	if possible.IsActive() {
		return possible
	}
	return nil
}

// Remove removes the specified client from the focus stack.
func Remove(c Client) {
	for i, c2 := range Clients {
		if c.Id() == c2.Id() {
			Clients = append(Clients[:i], Clients[i+1:]...)
			delete(clientSet, c.Id())
			break
		}
	}
}

func add(c Client) {
	Clients = append(Clients, c)
	clientSet[c.Id()] = true
}

// InitialAdd should be used whenever a client is being entered into the
// focus stack for the first time. It does NOT focus the client.
func InitialAdd(c Client) {
	Clients = append([]Client{c}, Clients...)
	clientSet[c.Id()] = true
}

// SetFocus moves the given client to the top of the focus stack and does
// nothing else. This is a way to force the focus stack into a state that
// has been discovered via Focus{In,Out} events.
//
// TODO(burntsushi): This is a hack. Try to move more of the focus logic into
// this package.
func SetFocus(c Client) {
	Remove(c)
	add(c)
}

// Focus will speak the proper X mumbo jumbo to send input focus to the
// specified client. This is the appropriate function to call whenever you
// want to focus any particular client.
//
// Focus has no effect if this is called on a client that was not added to the 
// focus stack with InitialAdd. Focus also has no effect if the client cannot
// handle input focus (like `xclock` or `xeyes`).
func Focus(c Client) {
	if !clientSet[c.Id()] {
		return
	}

	Remove(c)
	if c.CanFocus() || c.SendFocusNotify() {
		add(c)
		c.PrepareForFocus()
	}
	if c.CanFocus() {
		c.Win().Focus()
	}
	if c.SendFocusNotify() {
		protsAtm, err := xprop.Atm(X, "WM_PROTOCOLS")
		if err != nil {
			logger.Warning.Println(err)
			return
		}

		takeFocusAtm, err := xprop.Atm(X, "WM_TAKE_FOCUS")
		if err != nil {
			logger.Warning.Println(err)
			return
		}

		cm, err := xevent.NewClientMessage(32, c.Id(), protsAtm,
			int(takeFocusAtm), int(X.TimeGet()))
		if err != nil {
			logger.Warning.Println(err)
			return
		}

		xproto.SendEvent(X.Conn(), false, c.Id(), 0, string(cm.Bytes()))
	}
}

// Root emulates focus of the root window.
//
// N.B. Technically, a special off-screen window maintained by Wingo gets
// focus, but you won't be able to tell the difference. (I hope.)
func Root() {
	for _, c := range Clients {
		c.Unfocused()
	}
	xwindow.New(X, X.Dummy()).Focus()
}

// Fallback determines which client in the focus stack should be focused, and
// asks for it to be focused. The list of possible clients to be focused is
// filtered by the predicate focusable.
//
// This should be called after state changes (like when the current workspace
// is changed).
//
// If no focusable client is found, Root() is called.
func Fallback(focusable func(c Client) bool) {
	if c := LastFocused(focusable); c != nil {
		Focus(c)
	} else {
		Root()
	}
}

// LastFocused returns the last client that was focused that satisfies the
// predicate focusable. This is only used in the commands package to retrieve
// the active window. (It's a hack to work around the fact that prompts can
// steal focus, which makes the GetActive command not work properly.)
func LastFocused(focusable func(c Client) bool) Client {
	for i := len(Clients) - 1; i >= 0; i-- {
		c := Clients[i]
		if clientSet[c.Id()] && focusable(c) {
			return c
		}
	}
	return nil
}
