package commands

import (
	"time"

	"github.com/BurntSushi/gribble"

	"github.com/BurntSushi/wingo/prompt"
	"github.com/BurntSushi/wingo/wm"
	"github.com/BurntSushi/wingo/workspace"
	"github.com/BurntSushi/wingo/xclient"
)

type Input struct {
	Label string `param:"1"`
}

func (cmd Input) Run() gribble.Value {
	inputted := make(chan string, 0)

	response := func(inp *prompt.Input, text string) {
		inputted <- text
		inp.Hide()
	}
	canceled := func(inp *prompt.Input) {
		inputted <- ""
	}
	geom := wm.Workspace().Geom()
	if !wm.Prompts.Input.Show(geom, cmd.Label, response, canceled) {
		return ""
	}

	return <-inputted
}

type SelectClient struct {
	TabCompletion       string `param:"1"`
	OnlyActiveWorkspace string `param:"2"`
	OnlyVisible         string `param:"3"`
	ShowIconified       string `param:"4"`
}

func (cmd SelectClient) Run() gribble.Value {
	selected := make(chan int, 1)

	data := xclient.SelectData{
		Selected: func(c *xclient.Client) {
			selected <- int(c.Id())
		},
		Highlighted: nil,
	}
	wm.ShowSelectClient(
		stringTabComp(cmd.TabCompletion),
		stringBool(cmd.OnlyActiveWorkspace),
		stringBool(cmd.OnlyVisible),
		stringBool(cmd.ShowIconified),
		data)

	for {
		select {
		case clientId := <-selected:
			return clientId
		case <-time.After(10 * time.Second):
			if !wm.Prompts.Slct.Showing() {
				return ":void:"
			}
		}
	}
	panic("unreachable")
}

type SelectWorkspace struct {
	TabCompletion string `param:"1"`
}

func (cmd SelectWorkspace) Run() gribble.Value {
	selected := make(chan string, 1)

	data := workspace.SelectData{
		Selected: func(wrk *workspace.Workspace) {
			selected <- wrk.Name
		},
		Highlighted: nil,
	}
	wm.ShowSelectWorkspace(stringTabComp(cmd.TabCompletion), data)

	for {
		select {
		case wrkName := <-selected:
			return wrkName
		case <-time.After(10 * time.Second):
			if !wm.Prompts.Slct.Showing() {
				return ""
			}
		}
	}
	panic("unreachable")
}

