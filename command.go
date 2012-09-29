package main

import (
	"bytes"
	"os/exec"
	"strings"
	"time"
	"unicode"

	"github.com/BurntSushi/gribble"

	"github.com/BurntSushi/xgbutil/xevent"

	"github.com/BurntSushi/wingo/focus"
	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/prompt"
	"github.com/BurntSushi/wingo/stack"
)

// gribbleCommandEnv declares all available commands. Any command not in
// this list cannot be executed.
var gribbleCommandEnv = gribble.New([]gribble.Command{
	&CmdClose{},
	&CmdCycleClientNext{},
	&CmdCycleClientPrev{},
	&CmdFocus{},
	&CmdFocusRaise{},
	&CmdIconifyToggle{},
	&CmdMouseMove{},
	&CmdMouseResize{},
	&CmdMove{},
	&CmdRaise{},
	&CmdResize{},
	&CmdQuit{},
	&CmdSelectClient{},
	&CmdShell{},
})

type CmdClose struct {
	name   string      `Close`
	Client gribble.Any `param:"1" types:"int,string"`
}

func (cmd CmdClose) Run() gribble.Value {
	withClient(cmd.Client, func(c *client) {
		c.Close()
	})
	return nil
}

type CmdCycleClientChoose struct {
	name string `CycleClientChoose`
}

func (cmd CmdCycleClientChoose) Run() gribble.Value {
	wingo.prompts.cycle.Choose()
	return nil
}

type CmdCycleClientHide struct {
	name string `CycleClientHide`
}

func (cmd CmdCycleClientHide) Run() gribble.Value {
	wingo.prompts.cycle.Hide()
	return nil
}

type CmdCycleClientNext struct {
	name                string `CycleClientNext`
	OnlyActiveWorkspace string `param:"1"`
	OnlyVisible         string `param:"2"`
	ShowIconified       string `param:"3"`
}

func (cmd CmdCycleClientNext) Run() gribble.Value {
	cmd.RunWithKeyStr("")
	return nil
}

func (cmd CmdCycleClientNext) RunWithKeyStr(keyStr string) {
	showCycleClient(keyStr,
		stringBool(cmd.OnlyActiveWorkspace),
		stringBool(cmd.OnlyVisible),
		stringBool(cmd.ShowIconified))
	wingo.prompts.cycle.Next()
}

type CmdCycleClientPrev struct {
	name                string `CycleClientPrev`
	OnlyActiveWorkspace string `param:"1"`
	OnlyVisible         string `param:"2"`
	ShowIconified       string `param:"3"`
}

func (cmd CmdCycleClientPrev) Run() gribble.Value {
	cmd.RunWithKeyStr("")
	return nil
}

func (cmd CmdCycleClientPrev) RunWithKeyStr(keyStr string) {
	showCycleClient(keyStr,
		stringBool(cmd.OnlyActiveWorkspace),
		stringBool(cmd.OnlyVisible),
		stringBool(cmd.ShowIconified))
	wingo.prompts.cycle.Prev()
}

type CmdFocusRaise struct {
	name   string      `FocusRaise`
	Client gribble.Any `param:"1" types:"int,string"`
}

func (cmd CmdFocusRaise) Run() gribble.Value {
	withClient(cmd.Client, func(c *client) {
		focus.Focus(c)
		stack.Raise(c)
		xevent.ReplayPointer(X)
	})
	return nil
}

type CmdFocus struct {
	name   string      `Focus`
	Client gribble.Any `param:"1" types:"int,string"`
}

func (cmd CmdFocus) Run() gribble.Value {
	withClient(cmd.Client, func(c *client) {
		if c == nil {
			focus.Root()
		} else {
			focus.Focus(c)
			xevent.ReplayPointer(X)
		}
	})
	return nil
}

type CmdIconifyToggle struct {
	name   string      `IconifyToggle`
	Client gribble.Any `param:"1" types:"int,string"`
}

func (cmd CmdIconifyToggle) Run() gribble.Value {
	withClient(cmd.Client, func(c *client) {
		c.workspace.IconifyToggle(c)
	})
	return nil
}

type CmdMouseMove struct {
	name string `MouseMove`
}

func (cmd CmdMouseMove) Run() gribble.Value { return nil }

type CmdMouseResize struct {
	name      string `MouseResize`
	Direction string `param:"1"`
}

func (cmd CmdMouseResize) Run() gribble.Value { return nil }

type CmdRaise struct {
	name   string      `Raise`
	Client gribble.Any `param:"1" types:"int,string"`
}

func (cmd CmdRaise) Run() gribble.Value {
	withClient(cmd.Client, func(c *client) {
		stack.Raise(c)
		xevent.ReplayPointer(X)
	})
	return nil
}

type CmdMove struct {
	name   string      `Move`
	Client gribble.Any `param:"1" types:"int,string"`
	X      gribble.Any `param:"2" types:"int,float"`
	Y      gribble.Any `param:"3" types:"int,float"`
}

func (cmd CmdMove) Run() gribble.Value {
	x, xok := parsePos(cmd.X, false)
	y, yok := parsePos(cmd.Y, true)
	if !xok || !yok {
		return nil
	}
	withClient(cmd.Client, func(c *client) {
		c.LayoutMove(x, y)
	})
	return nil
}

type CmdResize struct {
	name   string      `Resize`
	Client gribble.Any `param:"1" types:"int,string"`
	Width  gribble.Any `param:"2" types:"int,float"`
	Height gribble.Any `param:"3" types:"int,float"`
}

func (cmd CmdResize) Run() gribble.Value {
	w, wok := parseDim(cmd.Width, false)
	h, hok := parseDim(cmd.Height, true)
	if !wok || !hok {
		return nil
	}
	withClient(cmd.Client, func(c *client) {
		c.LayoutResize(w, h)
	})
	return nil
}

type CmdQuit struct {
	name string `Quit`
}

func (cmd CmdQuit) Run() gribble.Value {
	logger.Message.Println("The User has told us to quit.")
	xevent.Quit(X)
	return nil
}

type CmdSelectClient struct {
	name                string `SelectClient`
	TabCompletion       string `param:"1"`
	OnlyActiveWorkspace string `param:"2"`
	OnlyVisible         string `param:"3"`
	ShowIconified       string `param:"4"`
}

func (cmd CmdSelectClient) Run() gribble.Value {
	tabComp := prompt.TabCompletePrefix
	switch cmd.TabCompletion {
	case "Prefix":
		tabComp = prompt.TabCompletePrefix
	case "Any":
		tabComp = prompt.TabCompleteAny
	default:
		logger.Warning.Printf(
			"Tab completion mode '%s' not supported.", cmd.TabCompletion)
	}
	showSelectClient(
		tabComp,
		stringBool(cmd.OnlyActiveWorkspace),
		stringBool(cmd.OnlyVisible),
		stringBool(cmd.ShowIconified))
	return nil
}

// commandShellFun takes a command specified in a configuration file and
// tries to parse it as an executable command. The command must be wrapped
// in "`" and "`" (back-quotes). If it's not, we return nil. Otherwise, we
// return a function that will execute the command.
// This provides rudimentary support for quoted values in the command.
type CmdShell struct {
	name    string `Shell`
	Command string `param:"1"`
}

func (cmd CmdShell) Run() gribble.Value {
	var stderr bytes.Buffer

	splitCmdName := strings.SplitN(cmd.Command, " ", 2)
	cmdName := splitCmdName[0]
	args := make([]string, 0)
	addArg := func(start, end int) {
		args = append(args, strings.TrimSpace(splitCmdName[1][start:end]))
	}

	if len(splitCmdName) > 1 {
		startArgPos := 0
		inQuote := false
		for i, char := range splitCmdName[1] {
			// Add arguments enclosed in quotes
			// Yes, this mixes up quotes.
			if char == '"' || char == '\'' {
				inQuote = !inQuote

				if !inQuote {
					addArg(startArgPos, i)
				}
				startArgPos = i + 1 // skip the end quote character
			}

			// Add arguments separated by spaces without quotes
			if !inQuote && unicode.IsSpace(char) {
				addArg(startArgPos, i)
				startArgPos = i
			}
		}

		// add anything that's left over
		addArg(startArgPos, len(splitCmdName[1]))
	}

	// XXX: This is very weird.
	// If I don't put this into its own go-routine and wait a small
	// amount of time, commands that start new X clients fail miserably.
	// And when I say miserably, I mean they take down X itself.
	// For some reason, this avoids that problem. For now...
	// (I thought the problem was the grab imposed by a key binding,
	// but ungrabbing the keyboard before running this command didn't
	// change behavior.)
	go func() {
		time.Sleep(time.Microsecond)
		shellCmd := exec.Command(cmdName, args...)
		shellCmd.Stderr = &stderr

		err := shellCmd.Run()
		if err != nil {
			logger.Warning.Printf("Error running '%s': %s", cmd.Command, err)
			if stderr.Len() > 0 {
				logger.Warning.Printf("Error running '%s': %s",
					cmd.Command, stderr.String())
			}
		}
	}()

	return nil
}

// stringBool takes a string and returns true if the string corresponds
// to a "true" value. i.e., "Yes", "Y", "y", "YES", "yEs", etc.
func stringBool(s string) bool {
	sl := strings.ToLower(s)
	return sl == "yes" || sl == "y"
}

// Shortcut for executing Client interface functions that have no parameters
// and no return values on the currently focused window.
func withFocused(f func(c *client)) {
	focused := focus.Current()
	if focused != nil {
		if client, ok := focused.(*client); ok {
			f(client)
		}
	}
}

func withClient(clientArg gribble.Any, f func(c *client)) {
	switch c := clientArg.(type) {
	case int:
		if c == 0 {
			withFocused(f)
			return
		}
		for _, client := range wingo.clients {
			if int(client.win.Id) == c {
				f(client)
				return
			}
		}
		return
	case string:
		switch c {
		case ":mouse:":
			f(mouseClientClicked)
		case ":active:":
			withFocused(f)
		default:
			panic("Client name Not implemented: " + c)
		}
		return
	}
	panic("unreachable")
}
