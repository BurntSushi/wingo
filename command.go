package main

import (
	"bytes"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/BurntSushi/wingo/logger"
	cmdUsage "github.com/BurntSushi/wingo/wingo-cmd"
)

func commandFun(keyStr string, cmd string, args ...string) func() {
	tryShellFun := commandShellFun(cmd)
	if tryShellFun != nil {
		return tryShellFun
	}

	usage := func(a func()) func() {
		return cmdUsage.MaybeUsage(cmd, a)
	}

	switch cmd {
	case "Close":
		return usage(cmdClose())
	case "FrameBorders":
		return usage(cmdFrameBorders())
	case "FrameFull":
		return usage(cmdFrameFull())
	case "FrameNada":
		return usage(cmdFrameNada())
	case "FrameSlim":
		return usage(cmdFrameSlim())
	case "HeadFocus":
		return usage(cmdHeadFocus(false, args...))
	case "HeadFocusWithClient":
		return usage(cmdHeadFocus(true, args...))
	case "MaximizeToggle":
		return usage(cmdMaximizeToggle())
	case "Minimize":
		return usage(cmdMinimize())
	case "PromptCycleNext":
		return usage(cmdPromptCycleNext(keyStr, args...))
	case "PromptCyclePrev":
		return usage(cmdPromptCyclePrev(keyStr, args...))
	case "PromptSelect":
		return usage(cmdPromptSelect(args...))
	case "Quit":
		return usage(cmdQuit())
	case "Workspace":
		return usage(cmdWorkspace(false, args...))
	case "WorkspacePrefix":
		return usage(cmdWorkspacePrefix(false, args...))
	case "WorkspaceWithClient":
		return usage(cmdWorkspace(true, args...))
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

// Shortcut for executing Client interface functions that have no parameters
// and no return values on the currently focused window.
func withFocused(f func(c *client)) {
	focused := WM.focused()
	if focused != nil {
		f(focused)
	}
}

func cmdClose() func() {
	return func() {
		withFocused(func(c *client) {
			c.Close()
		})
	}
}

func cmdFrameBorders() func() {
	return func() {
		withFocused(func(c *client) {
			c.FrameBorders()
		})
	}
}

func cmdFrameFull() func() {
	return func() {
		withFocused(func(c *client) {
			c.FrameFull()
		})
	}
}

func cmdFrameNada() func() {
	return func() {
		withFocused(func(c *client) {
			c.FrameNada()
		})
	}
}

func cmdFrameSlim() func() {
	return func() {
		withFocused(func(c *client) {
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
			withFocused(func(c *client) {
				c.Raise()
				wrk.Add(c, false)
			})
		}
		WM.WrkSet(wrk.id, true, greedy)
	}
}

func cmdMaximizeToggle() func() {
	return func() {
		withFocused(func(c *client) {
			c.MaximizeToggle()
		})
	}
}

func cmdMinimize() func() {
	return func() {
		withFocused(func(c *client) {
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

func cmdWorkspace(withClient bool, args ...string) func() {
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
		WM.WrkSet(wrk.id, true, greedy)
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
			logger.Warning.Printf("The 'WorkspacePrefix' command could not find "+
				"a non-visible workspace with prefix '%s'.",
				args[0])
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
