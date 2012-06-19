package main

import (
	"bytes"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"

	"github.com/BurntSushi/wingo/cmdusage"
	"github.com/BurntSushi/wingo/focus"
	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/stack"
)

func commandFun(keyStr string, cmd string, args ...string) func() {
	tryShellFun := commandShellFun(cmd)
	if tryShellFun != nil {
		return tryShellFun
	}

	usage := func(a func()) func() {
		return cmdusage.MaybeUsage(cmd, a)
	}

	switch cmd {
	case "Close":
		return usage(cmdClose(args...))
	case "Focus":
		return usage(cmdFocus(args...))
	case "Move":
		return usage(cmdMove(args...))
	case "Quit":
		return usage(cmdQuit())
	case "Raise":
		return usage(cmdRaise(args...))
	case "Resize":
		return usage(cmdResize(args...))
	}

	if len(args) > 0 {
		logger.Warning.Printf("Unrecognized key command: %s %s", cmd, args)
	} else {
		logger.Warning.Printf("Unrecognized key command: %s", cmd)
	}

	return nil
}

// commandHandler responds to client message events that issue commands.
func commandHandler(X *xgbutil.XUtil, cm xevent.ClientMessageEvent) {
	typeName, err := xprop.AtomName(X, cm.Type)
	if err != nil {
		logger.Warning.Printf(
			"Could not get type of ClientMessage event: %s", cm)
		return
	}

	if typeName == "_WINGO_CMD" {
		cmd, err := cmdusage.CmdGet(X)
		if err != nil {
			logger.Warning.Printf("Could not get _WINGO_CMD value: %s", err)
			return
		}

		// Blank out the command
		cmdusage.CmdSet(X, "")

		// Parse the command
		cmdName, args := commandParse(cmd)
		cmdFun := commandFun("", cmdName, args...)
		if cmdFun != nil {
			cmdFun()
			cmdusage.StatusSet(X, true)
		} else {
			cmdusage.StatusSet(X, false)
		}
	}
}

// commandParse takes a single string and parses it into a
// (CommandName, [Arg1, Arg2, ...]) tuple.
func commandParse(command string) (cmd string, args []string) {
	pieces := strings.Split(command, " ")
	cmd = pieces[0]
	args = make([]string, len(pieces)-1)

	for i, arg := range pieces[1:] {
		args[i] = strings.ToLower(strings.TrimSpace(arg))
	}
	return
}

// commandArgsClient scans an argument list for a window id.
// A window id has the form 'win:WINDOW-ID_NUMBER'.
// Both 'win:0x0001' and 'win:1' are valid thanks to Go's ParseInt.
// Finally, if the window id corresponds to managed client, return that
// client. Otherwise, return nil and emit an error if we have an invalid ID.
// We also return a bool as a second argument which should be interpreted
// as whether or not to continue the current operation.
// i.e., not finding anything that looks like a window id is safe to ignore,
// but if we find something like an ID and error out, we should stop the
// command entirely.
func commandArgsClient(args []string) (*client, bool) {
	for _, arg := range args {
		if len(arg) < 5 || arg[0:4] != "win:" {
			continue
		}

		maybeId64, err := strconv.ParseInt(arg[4:], 0, 0)
		if err != nil {
			logger.Warning.Printf("'%s' is not a valid window id.", arg[4:])
			return nil, false
		}

		goodId := xproto.Window(maybeId64)
		for _, c := range wingo.clients {
			if c.Id() == goodId {
				return c, true
			}
		}

		logger.Warning.Printf(
			"'%s' is a valid window ID, but does not match any managed "+
				"window ID by Wingo.", arg[4:])
		return nil, false
	}
	return nil, true
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

func withFocusedOrArg(args []string, f func(c *client)) {
	client, ok := commandArgsClient(args)
	if !ok {
		return
	}

	if client == nil {
		withFocused(f)
	} else {
		f(client)
	}
}

func cmdClose(args ...string) func() {
	return func() {
		withFocusedOrArg(args, func(c *client) {
			c.Close()
		})
	}
}

func cmdFocus(args ...string) func() {
	return func() {
		withFocusedOrArg(args, func(c *client) {
			focus.Focus(c)
		})
	}
}

func cmdMove(args ...string) func() {
	if len(args) < 2 {
		return nil
	}

	return func() {
		x, xok := parsePos(args[0], false)
		y, yok := parsePos(args[1], true)
		if !xok || !yok {
			return
		}

		withFocusedOrArg(args, func(c *client) {
			c.LayoutMove(x, y)
		})
	}
}

func cmdQuit() func() {
	return func() {
		logger.Message.Println("The User has told us to quit.")
		xevent.Quit(X)
	}
}

func cmdRaise(args ...string) func() {
	return func() {
		withFocusedOrArg(args, func(c *client) {
			stack.Raise(c)
		})
	}
}

func cmdResize(args ...string) func() {
	if len(args) < 2 {
		return nil
	}

	return func() {
		w, wok := parseDim(args[0], false)
		h, hok := parseDim(args[1], true)
		if !wok || !hok {
			return
		}

		withFocusedOrArg(args, func(c *client) {
			c.LayoutResize(w, h)
		})
	}
}

// commandShellFun takes a command specified in a configuration file and
// tries to parse it as an executable command. The command must be wrapped
// in "`" and "`" (back-quotes). If it's not, we return nil. Otherwise, we
// return a function that will execute the command.
// This provides rudimentary support for quoted values in the command.
func commandShellFun(cmd string) func() {
	if cmd[0] != '`' || cmd[len(cmd)-1] != '`' {
		return nil
	}

	return func() {
		var stderr bytes.Buffer

		allCmd := cmd[1 : len(cmd)-1]

		splitCmdName := strings.SplitN(allCmd, " ", 2)
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
			cmd := exec.Command(cmdName, args...)
			cmd.Stderr = &stderr

			err := cmd.Run()
			if err != nil {
				logger.Warning.Printf("Error running '%s': %s", allCmd, err)
				if stderr.Len() > 0 {
					logger.Warning.Printf("Error running '%s': %s",
						allCmd, stderr.String())
				}
			}
		}()
	}
}
