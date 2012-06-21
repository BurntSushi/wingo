package stack

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/xwindow"
)

type Client interface {
	Id() xproto.Window
	Win() *xwindow.Window
	TopWin() *xwindow.Window
	Layer() int
	Transient(client Client) bool
}

func clientIndex(needle Client, haystack []Client) int {
	for i, client := range haystack {
		if client.Id() == needle.Id() {
			return i
		}
	}
	return -1
}
