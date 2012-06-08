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
	return Clients[len(Clients)-1]
}

func Remove(c Client) {
	for i, c2 := range Clients {
		if c.Id() == c2.Id() {
			Clients = append(Clients[:i], Clients[i+1:]...)
			break
		}
	}
}

func Focus(c Client) {
	Remove(c)
	Clients = append(Clients, c)

	if c.CanFocus() || c.SendFocusNotify() {
		c.MakeViewable()
		c.Focused()
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
	rwin := xwindow.New(X, X.RootWin())
	for _, c := range Clients {
		c.Unfocused()
	}
	rwin.Focus()
}

func UnfocusExcept(c Client) {
	id := c.Id()
	for i := len(Clients) - 1; i >= 0; i-- {
		if Clients[i].Id() != id {
			Clients[i].Unfocused()
		}
	}
}
