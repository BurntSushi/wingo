package stack

import (
	"github.com/BurntSushi/xgb/xproto"
)

type Client interface {
	Id() xproto.Window
	TopLevelId() xproto.Window
	Layer() int
	StackSibling(siblingWin xproto.Window, stackMode byte)
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
