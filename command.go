package main

import (
	"bytes"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unicode"

	"code.google.com/p/jamslam-x-go-binding/xgb"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"

	"github.com/BurntSushi/wingo/logger"
	command "github.com/BurntSushi/wingo/wingo-cmd"
)

func commandFun(keyStr string, cmd string, args ...string) func() {
	tryShellFun := commandShellFun(cmd)
	if tryShellFun != nil {
		return tryShellFun
	}

	usage := func(a func()) func() {
		return command.MaybeUsage(cmd, a)
	}

	switch cmd {
	case "Close":
		return usage(cmdClose(args...))
	case "FrameBorders":
		return usage(cmdFrameBorders(args...))
	case "FrameFull":
		return usage(cmdFrameFull(args...))
	case "FrameNada":
		return usage(cmdFrameNada(args...))
	case "FrameSlim":
		return usage(cmdFrameSlim(args...))
	case "HeadFocus":
		return usage(cmdHeadFocus(false, args...))
	case "HeadFocusWithClient":
		return usage(cmdHeadFocus(true, args...))
	case "MaximizeToggle":
		return usage(cmdMaximizeToggle(args...))
	case "Minimize":
		return usage(cmdMinimize(args...))
	case "PromptCycleNext":
		if len(keyStr) == 0 {
			return nil
		}
		return usage(cmdPromptCycleNext(keyStr, args...))
	case "PromptCyclePrev":
		if len(keyStr) == 0 {
			return nil
		}
		return usage(cmdPromptCyclePrev(keyStr, args...))
	case "PromptSelect":
		return usage(cmdPromptSelect(args...))
	case "Quit":
		return usage(cmdQuit())
	case "Workspace":
		return usage(cmdWorkspace(false, false, args...))
	case "WorkspacePrefix":
		return usage(cmdWorkspacePrefix(false, args...))
	case "WorkspaceSetClient":
		return usage(cmdWorkspace(false, true, args...))
	case "WorkspaceWithClient":
		return usage(cmdWorkspace(true, false, args...))
	case "WorkspacePrefixWithClient":
		return usage(cmdWorkspacePrefix(true, args...))
	case "WorkspaceLeft":
		return usage(cmdWorkspaceLeft())
	case "WorkspaceLeftWithClient":
		return usage(cmdWorkspaceLeftWithClient())
	case "WorkspaceRight":
		return usage(cmdWorkspaceRight())
	case "WorkspaceRightWithClient":
		return usage(cmdWorkspaceRightWithClient())
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
		cmd, err := command.Get(X)
		if err != nil {
			logger.Warning.Printf("Could not get _WINGO_CMD value: %s", err)
			return
		}

		// Blank out the command
		command.Set(X, "")

		// Parse the command
		cmdName, args := commandParse(cmd)
		cmdFun := commandFun("", cmdName, args...)
		if cmdFun != nil {
			cmdFun()
			command.StatusSet(X, true)
		} else {
			command.StatusSet(X, false)
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

		goodId := xgb.Id(maybeId64)
		for _, c := range WM.clients {
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
	focused := WM.focused()
	if focused != nil {
		f(focused)
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

func cmdFrameBorders(args ...string) func() {
	return func() {
		withFocusedOrArg(args, func(c *client) {
			c.FrameBorders()
		})
	}
}

func cmdFrameFull(args ...string) func() {
	return func() {
		withFocusedOrArg(args, func(c *client) {
			c.FrameFull()
		})
	}
}

func cmdFrameNada(args ...string) func() {
	return func() {
		withFocusedOrArg(args, func(c *client) {
			c.FrameNada()
		})
	}
}

func cmdFrameSlim(args ...string) func() {
	return func() {
		withFocusedOrArg(args, func(c *client) {
			c.FrameSlim()
		})
	}
}

func cmdHeadFocus(withClient bool, args ...string) func() {
	if len(args) < 1 {
		logger.Warning.Printf("Improper use of HeadFocus command.")
		logger.Warning.Printf("Usage: HeadFocus head_number [Greedy]")
		return nil
	}

	headNum64, err := strconv.ParseInt(args[0], 0, 0)
	if err != nil {
		logger.Warning.Printf("Improper use of HeadFocus command.")
		logger.Warning.Printf("'%s' is not a valid head number.", args[0])
		return nil
	}
	headNum := int(headNum64)

	return func() {
		var wrk *workspace
		for _, wrk2 := range WM.workspaces {
			if wrk2.head == headNum {
				wrk = wrk2
				break
			}
		}

		if wrk == nil {
			logger.Warning.Printf("The 'HeadFocus' command could not find "+
				"head number '%d'.", headNum)
			return
		}

		greedy := false
		if len(args) > 1 && args[1] == "greedy" {
			greedy = true
		}

		if withClient {
			withFocusedOrArg(args, func(c *client) {
				c.Raise()
				wrk.Add(c, false)
			})
		}
		WM.WrkSet(wrk.id, true, greedy)
	}
}

func cmdMaximizeToggle(args ...string) func() {
	return func() {
		withFocusedOrArg(args, func(c *client) {
			c.MaximizeToggle()
		})
	}
}

func cmdMinimize(args ...string) func() {
	return func() {
		withFocusedOrArg(args, func(c *client) {
			c.IconifyToggle()
		})
	}
}

func cmdPromptCycleNext(keyStr string, args ...string) func() {
	if len(args) < 1 {
		return nil
	}

	activeWrk, visible, iconified := false, false, true
	switch args[0] {
	case "clientsworkspace":
		activeWrk = true
	case "clientsmonitors":
		visible = true
	default:
		logger.Warning.Printf("Unrecognized argument '%s' for PromptCycleNext "+
			"command", args[0])
		return nil
	}

	return func() {
		PROMPTS.cycle.next(keyStr, activeWrk, visible, iconified)
	}
}

func cmdPromptCyclePrev(keyStr string, args ...string) func() {
	if len(args) < 1 {
		return nil
	}

	activeWrk, visible, iconified := false, false, true
	switch args[0] {
	case "clientsworkspace":
		activeWrk = true
	case "clientsmonitors":
		visible = true
	default:
		logger.Warning.Printf("Unrecognized argument '%s' for PromptCyclePrev "+
			"command", args[0])
		return nil
	}

	return func() {
		PROMPTS.cycle.prev(keyStr, activeWrk, visible, iconified)
	}
}

func cmdPromptSelect(args ...string) func() {
	if len(args) < 1 {
		return nil
	}

	prefixSearch := true
	if len(args) > 1 && args[1] == "substring" {
		prefixSearch = false
	}

	var f func() []*promptSelectGroup

	switch args[0] {
	case "clientsall":
		f = func() []*promptSelectGroup {
			return promptSelectListClients(false, false, true)
		}
	case "clientsworkspace":
		f = func() []*promptSelectGroup {
			return promptSelectListClients(true, false, true)
		}
	case "clientsmonitors":
		f = func() []*promptSelectGroup {
			return promptSelectListClients(false, true, true)
		}
	case "workspaces":
		f = func() []*promptSelectGroup {
			action := func(wrk *workspace) func() {
				return func() {
					WM.WrkSet(wrk.id, true, true)
				}
			}
			return promptSelectListWorkspaces(action)
		}
	case "workspaceswithclient":
		f = func() []*promptSelectGroup {
			action := func(wrk *workspace) func() {
				return func() {
					withFocused(func(c *client) {
						c.Raise()
						wrk.Add(c, false)
						WM.WrkSet(wrk.id, true, true)
					})
				}
			}
			return promptSelectListWorkspaces(action)
		}
	default:
		logger.Warning.Printf("Unrecognized argument '%s' for PromptSelect "+
			"command", args[0])
		return nil
	}

	return func() {
		PROMPTS.slct.show(f, prefixSearch)
	}
}

func cmdQuit() func() {
	return func() {
		logger.Message.Println("The User has told us to quit.")
		X.Quit()
	}
}

func cmdWorkspace(withClient bool, setClient bool, args ...string) func() {
	if withClient && setClient {
		panic("Bug in cmdWorkspace. 'withClient' and 'setClient' cannot " +
			"both be true.")
	}
	if len(args) < 1 {
		return nil
	}

	return func() {
		wrk := WM.WrkFind(args[0])
		if wrk == nil {
			logger.Warning.Printf("The 'Workspace' command could not find "+
				"workspace '%s'.", args[0])
			return
		}

		greedy := false
		if len(args) > 1 && args[1] == "greedy" {
			greedy = true
		}

		if withClient {
			withFocused(func(c *client) {
				c.Raise()
				wrk.Add(c, false)
			})
		}

		if setClient {
			withFocusedOrArg(args, func(c *client) {
				wrk.Add(c, true)
			})
		} else {
			WM.WrkSet(wrk.id, true, greedy)
		}
	}
}

// cmdWorkspacePrefix is actually a bit tricky.
// The naive solution is to simply search the list of workspaces, and if any
// of them have a prefix matching the argument, use that workspace.
// However, this results in only being able to switch between TWO workspaces
// with the same prefix (namely, the first two in the workspaces list).
// We fix this by making sure we don't start our prefix search until we have
// seen the active workspace. Also, we never choose a workspace that is
// visible.
// If that search turns up nothing, we run the search again, but take the
// first non-visible workspace matching the prefix.
// We don't choose visible workspaces, because the idea is that those can
// be switched to more explicitly with the 'Workspace' command or with one
// of the 'HeadFocus' commands.
func cmdWorkspacePrefix(withClient bool, args ...string) func() {
	if len(args) < 1 {
		return nil
	}

	return func() {
		var wrk *workspace

		// visibles := 0 
		pastActive := false
		prefix := strings.ToLower(args[0])
		for _, wrk2 := range WM.workspaces {
			if wrk2.active {
				pastActive = true
				continue
			}
			if !pastActive {
				continue
			}
			if wrk2.visible() {
				continue
			}
			if strings.HasPrefix(strings.ToLower(wrk2.name), prefix) {
				wrk = wrk2
				break
			}
		}

		// Try the search again, but use the first non-visible workspace
		// with a matching prefix.
		if wrk == nil {
			for _, wrk2 := range WM.workspaces {
				if !wrk2.visible() &&
					strings.HasPrefix(strings.ToLower(wrk2.name), prefix) {
					wrk = wrk2
					break
				}
			}
		}

		if wrk == nil {
			logger.Warning.Printf(
				"The 'WorkspacePrefix' command could not find "+
					"a non-visible workspace with prefix '%s'.", args[0])
			return
		}

		greedy := false
		if len(args) > 1 && args[1] == "greedy" {
			greedy = true
		}

		if withClient {
			withFocused(func(c *client) {
				c.Raise()
				wrk.Add(c, false)
			})
		}
		WM.WrkSet(wrk.id, true, greedy)
	}
}

func cmdWorkspaceLeft() func() {
	return func() {
		wrkAct := WM.WrkActive()
		WM.WrkSet(mod(wrkAct.id-1, len(WM.workspaces)), true, false)
	}
}

func cmdWorkspaceLeftWithClient() func() {
	return func() {
		withFocused(func(c *client) {
			c.Raise()
			wrkAct := WM.WrkActive()
			wrkPrev := WM.workspaces[mod(wrkAct.id-1, len(WM.workspaces))]
			wrkPrev.Add(c, false)
			WM.WrkSet(wrkPrev.id, true, false)
		})
	}
}

func cmdWorkspaceRight() func() {
	return func() {
		wrkAct := WM.WrkActive()
		WM.WrkSet(mod(wrkAct.id+1, len(WM.workspaces)), true, false)
	}
}

func cmdWorkspaceRightWithClient() func() {
	return func() {
		withFocused(func(c *client) {
			c.Raise()
			wrkAct := WM.WrkActive()
			wrkNext := WM.workspaces[mod(wrkAct.id+1, len(WM.workspaces))]
			wrkNext.Add(c, false)
			WM.WrkSet(wrkNext.id, true, false)
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
	}
}

// This is a start, but it is not quite ready for prime-time yet.
// 1. If the window is destroyed while the go routine is still running,
// we're in big trouble.
// 2. This has no way to stop from some external event (like focus).
// Basically, both of these things can be solved by using channels to tell
// the goroutine to quit. Not difficult but also not worth my time atm.
func cmd_active_flash() {
	focused := WM.focused()

	if focused != nil {
		go func(c *client) {
			for i := 0; i < 10; i++ {
				if c.Frame().State() == StateActive {
					c.Frame().Inactive()
				} else {
					c.Frame().Active()
				}

				time.Sleep(time.Second)
			}
		}(focused)
	}
}
