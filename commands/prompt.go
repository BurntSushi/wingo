package commands

import (
	"time"

	"github.com/BurntSushi/gribble"

	"github.com/BurntSushi/wingo/prompt"
	"github.com/BurntSushi/wingo/wm"
	"github.com/BurntSushi/wingo/workspace"
	"github.com/BurntSushi/wingo/xclient"
)

type CycleClientChoose struct{
	Help string `
Activates the current choice in a cycle prompt.
`
}

func (cmd CycleClientChoose) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		wm.Prompts.Cycle.Choose()
		return nil
	})
}

type CycleClientHide struct{
	Help string `
Hides (i.e., cancels) the current cycle prompt.
`
}

func (cmd CycleClientHide) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		wm.Prompts.Cycle.Hide()
		return nil
	})
}

type CycleClientNext struct {
	OnlyActiveWorkspace string `param:"1"`
	OnlyVisible         string `param:"2"`
	ShowIconified       string `param:"3"`
	Help string `
Shows the cycle prompt for clients and advances the selection to the next
client. If the cycle prompt is already visible, then the selection is advanced
to the next client.

OnlyActiveWorkspace specifies that only clients on the current workspace should
be listed. Valid values are "yes" or "no".

OnlyVisible specifies that only clients on visible workspaces should be listed.
Valid values are "yes" or "no".

ShowIconified specifies that iconified clients will be shown. Valid values are
"yes" or "no".
`
}

func (cmd CycleClientNext) Run() gribble.Value {
	cmd.RunWithKeyStr("")
	return nil
}

func (cmd CycleClientNext) RunWithKeyStr(keyStr string) {
	syncRun(func() gribble.Value {
		wm.ShowCycleClient(keyStr,
			stringBool(cmd.OnlyActiveWorkspace),
			stringBool(cmd.OnlyVisible),
			stringBool(cmd.ShowIconified))
		wm.Prompts.Cycle.Next()
		return nil
	})
}

type CycleClientPrev struct {
	OnlyActiveWorkspace string `param:"1"`
	OnlyVisible         string `param:"2"`
	ShowIconified       string `param:"3"`
	Help string `
Shows the cycle prompt for clients and advances the selection to the previous
client. If the cycle prompt is already visible, then the selection is advanced
to the previous client.

OnlyActiveWorkspace specifies that only clients on the current workspace should
be listed. Valid values are "yes" or "no".

OnlyVisible specifies that only clients on visible workspaces should be listed.
Valid values are "yes" or "no".

ShowIconified specifies that iconified clients will be shown. Valid values are
"yes" or "no".
`
}

func (cmd CycleClientPrev) Run() gribble.Value {
	cmd.RunWithKeyStr("")
	return nil
}

func (cmd CycleClientPrev) RunWithKeyStr(keyStr string) {
	syncRun(func() gribble.Value {
		wm.ShowCycleClient(keyStr,
			stringBool(cmd.OnlyActiveWorkspace),
			stringBool(cmd.OnlyVisible),
			stringBool(cmd.ShowIconified))
		wm.Prompts.Cycle.Prev()
		return nil
	})
}

type Input struct {
	Label string `param:"1"`
	Help string `
Shows a centered prompt window that allows the user to type in text. If the
user presses the Confirm Key (i.e., enter), then the text typed into the
input box will be returned.

Label will be shown next to the input box.

This command may be used as a sub-command to pass user provided arguments to
another command.
`
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

type Message struct {
	Text string `param:"1"`
	Help string `
Shows a centered prompt window with the text specified by Text. The message
will not disappear until it loses focus or when the confirm or cancel key
is pressed.
`
}

func (cmd Message) Run() gribble.Value {
	wm.PopupError("%s", cmd.Text)
	return nil
}

type SelectClient struct {
	TabCompletion       string `param:"1"`
	OnlyActiveWorkspace string `param:"2"`
	OnlyVisible         string `param:"3"`
	ShowIconified       string `param:"4"`
	Help string `
Shows a centered prompt window with a list of clients satisfying the arguments
provided.

OnlyActiveWorkspace specifies that only clients on the current workspace should
be listed. Valid values are "yes" or "no".

OnlyVisible specifies that only clients on visible workspaces should be listed.
Valid values are "yes" or "no".

ShowIconified specifies that iconified clients will be shown. Valid values are
"yes" or "no".

TabCompletetion can be set to either "Prefix", "Any" or "Multiple". When it's
set to "Prefix", the clients can be searched by a prefix matching string. When
it's set to "Any", the clients can be searched by a substring matching string.
When it's set to "Multiple", the clients can be searched by multiple space-
separated substring matching strings.

This command may be used as a sub-command to pass a particular client to
another command.
`
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
	Help string `
Shows a centered prompt window with a list of all workspaces.

TabCompletetion can be set to either "Prefix", "Any" or "Multiple". When it's
set to "Prefix", the clients can be searched by a prefix matching string. When
it's set to "Any", the clients can be searched by a substring matching string.
When it's set to "Multiple", the clients can be searched by multiple space-
separated substring matching strings.

This command may be used as a sub-command to pass a particular workspace to
another command.
`
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

