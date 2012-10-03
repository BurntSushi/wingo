package focus

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/workspace"
)

var (
	X       *xgbutil.XUtil
	Clients []Client
)

func Initialize(xu *xgbutil.XUtil) {
	X = xu
	Clients = make([]Client, 0, 100)
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

func Remove(c Client) {
	for i, c2 := range Clients {
		if c.Id() == c2.Id() {
			Clients = append(Clients[:i], Clients[i+1:]...)
			break
		}
	}
}

func InitialAdd(c Client) {
	Clients = append([]Client{c}, Clients...)
}

// SetFocus moves the given client to the top of the focus stack and does
// nothing else. This is a way to force the focus stack into a state that
// has been discovered via Focus{In,Out} events.
func SetFocus(c Client) {
	Remove(c)
	Clients = append(Clients, c)
}

func Focus(c Client) {
	Remove(c)

	if c.CanFocus() || c.SendFocusNotify() {
		Clients = append(Clients, c)
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

func Fallback(wrk *workspace.Workspace) {
	for i := len(Clients) - 1; i >= 0; i-- {
		c := Clients[i]
		if c.IsMapped() && c.Workspace() == wrk && !c.ImminentDestruction() {
			Focus(c)
			return
		}
	}
	Root()
}
