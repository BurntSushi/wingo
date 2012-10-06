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
	X         *xgbutil.XUtil
	Clients   []Client
	clientSet map[xproto.Window]bool // Constant time set of Clients
)

func Initialize(xu *xgbutil.XUtil) {
	X = xu
	Clients = make([]Client, 0, 100)
	clientSet = make(map[xproto.Window]bool, 100)
}

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

func LastCurrent() Client {
	if len(Clients) == 0 {
		return nil
	}
	return Clients[len(Clients)-1]
}

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

func InitialAdd(c Client) {
	Clients = append([]Client{c}, Clients...)
	clientSet[c.Id()] = true
}

// SetFocus moves the given client to the top of the focus stack and does
// nothing else. This is a way to force the focus stack into a state that
// has been discovered via Focus{In,Out} events.
func SetFocus(c Client) {
	Remove(c)
	add(c)
}

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

func Root() {
	for _, c := range Clients {
		c.Unfocused()
	}
	xwindow.New(X, X.Dummy()).Focus()
}

func Fallback(focusable func(c Client) bool) {
	if c := LastFocused(focusable); c != nil {
		Focus(c)
	} else {
		Root()
	}
}

func LastFocused(focusable func(c Client) bool) Client {
	for i := len(Clients) - 1; i >= 0; i-- {
		c := Clients[i]
		if clientSet[c.Id()] && focusable(c) {
			return c
		}
	}
	return nil
}
