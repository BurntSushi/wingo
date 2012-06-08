package main

import (
	"strings"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/wingo/focus"
	"github.com/BurntSushi/wingo/frame"
	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/stack"
)

type state struct {
	clients []*client
	heads   *heads.Heads
}

func newState() *state {
	wingo := &state{
		clients: make([]*client, 50),
		heads:   nil,
	}
	wingo.heads = heads.NewHeads(X, wingo.clients, "Numero Uno")

	return wingo
}

func (wingo *state) add(c *client) {
	if cliIndex(c, wingo.clients) == -1 {
		logger.Message.Println("Managing new client:", c)
		wingo.clients = append(wingo.clients, c)
	} else {
		logger.Message.Println("Already managing client:", c)
	}
}

func (wingo *state) remove(c *client) {
	if i := cliIndex(c, wingo.clients); i > -1 {
		logger.Message.Println("Unmanaging client:", c)
		wingo.clients = append(wm.clients[:i], wm.clients[i+1:]...)
		focus.Remove(c)
		stack.Remove(c)
	}
}

func (wingo *state) focusFallback() {
	wrk := wingo.heads.ActiveWorkspace()
	for i := len(focus.Clients) - 1; i >= 0; i-- {
		switch client := focus.Clients.(type) {
		case *client:
			if !client.Iconified() && client.Workspace() == wrk {
				focus.Focus(client)
			}
		default:
			fmt.Printf("Unsupported client type: %T", client)
			panic("Not implemented.")
		}
	}
	focus.Root()
}
