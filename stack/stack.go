package stack

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
)

const (
	LayerDesktop = iota
	LayerBelow
	LayerDefault
	LayerAbove
	LayerDock
	LayerFullscreen
)

const (
	below = xproto.StackModeBelow
	above = xproto.StackModeAbove
)

var (
	X       *xgbutil.XUtil
	Clients []Client
)

func Raise(client Client) {
	raise(client)

	// A slice of clients to physically update. The idea here is to do all of
	// the stacking state changes, and then apply them in one swoop. This allows
	// us to avoid flashing or redundantly stacking windows.
	// TODO: Find a more elegant way to do this.
	updateClients := make([]Client, 0, 4)
	updateClients = append(updateClients, client)
	for i := len(Clients) - 1; i >= 0; i-- {
		if client.Transient(Clients[i]) {
			updateClients = append(updateClients, Clients[i])
		}
	}
	for _, client2 := range updateClients {
		raise(client2)
	}
	realize(updateClients)
}

func raise(client Client) {
	Remove(client)
	if len(Clients) == 0 {
		Clients = []Client{client}
		return
	}
	for i, client2 := range Clients {
		if client.Id() == client2.Id() {
			continue
		}
		if client2.Layer() <= client.Layer() {
			Clients = append(Clients[:i],
				append([]Client{client}, Clients[i:]...)...)
			return
		}
	}
	Clients = append(Clients, client)
}

func realize(updateClients []Client) {
	if len(Clients) <= 1 {
		return
	}
	for i := len(Clients) - 1; i >= 0; i-- {
		if clientIndex(Clients[i], updateClients) > -1 {
			if i == len(Clients)-1 {
				Clients[i].StackSibling(Clients[i-1].TopLevelId(), below)
			} else {
				Clients[i].StackSibling(Clients[i+1].TopLevelId(), above)
			}
		}
	}
}

func Remove(client Client) {
	if i := clientIndex(client, Clients); i > -1 {
		Clients = append(Clients[:i], Clients[i+1:]...)
	}
}
