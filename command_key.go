/*
   command_key.go is responsible for setting up *all* key bindings found
   in the key.wini config file.

   It isn't quite the same as command_mouse.go because they operate under
   two different assumptions: key bindings are global in nature (i.e.,
   they are bound to the root window) while mouse bindings are window
   specific in nature (i.e., bound to each specific window).

   This actually makes command_key.go simpler than command_mouse.go, because
   we don't need to provide an interface for each client to bind keys
   separately. We can just bind them to the root window and let the commands
   infer state and act appropriately.
*/
package main

import (
	"fmt"
	"strconv"
	"strings"
)

import (
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
)

type keyCommand struct {
	cmd    string
	args   []string
	down   bool // 'up' when false
	keyStr string
}

func attachAllKeys() {
	for _, kcmds := range CONF.key {
		for _, kcmd := range kcmds {
			kcmd.attach(kcmd.commandFun())
		}
	}
}

func (kcmd keyCommand) attach(run func()) {
	if run == nil {
		return
	}

	if kcmd.cmd == "PromptCyclePrev" || kcmd.cmd == "PromptCycleNext" {
		// We've got to parse the key string first and make sure
		// there are some modifiers; otherwise this utterly fails!
		mods, _ := keybind.ParseString(X, kcmd.keyStr)
		if mods == 0 {
			logWarning.Printf("Sorry but the key binding '%s' for the %s "+
				"command is invalid. It must have a modifier "+
				"to work properly. i.e., Mod1-tab where 'Mod1' "+
				"is the modifier.", kcmd.keyStr, kcmd.cmd)
			return
		}

		keybind.KeyPressFun(
			func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
				run()
			}).Connect(X, ROOT.id, kcmd.keyStr)
		keybind.KeyPressFun(
			func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
				run()
			}).Connect(X, X.Dummy(), kcmd.keyStr)
	} else {
		if kcmd.down {
			keybind.KeyPressFun(
				func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
					run()
				}).Connect(X, ROOT.id, kcmd.keyStr)
		} else {
			keybind.KeyReleaseFun(
				func(X *xgbutil.XUtil, ev xevent.KeyReleaseEvent) {
					run()
				}).Connect(X, ROOT.id, kcmd.keyStr)
		}
	}
}

func (kcmd keyCommand) String() string {
	if len(kcmd.args) > 0 {
		return fmt.Sprintf("%s %s", kcmd.cmd, kcmd.args)
	}
	return fmt.Sprintf("%s", kcmd.cmd)
}

func (kcmd keyCommand) commandFun() func() {
	tryShellFun := commandShellFun(kcmd.cmd)
	if tryShellFun != nil {
		return tryShellFun
	}

	switch kcmd.cmd {
	case "Close":
		return cmdClose()
	case "FrameBorders":
		return cmdFrameBorders()
	case "FrameFull":
		return cmdFrameFull()
	case "FrameNada":
		return cmdFrameNada()
	case "FrameSlim":
		return cmdFrameSlim()
	case "HeadFocus":
		return cmdHeadFocus(kcmd.args...)
	case "MaximizeToggle":
		return cmdMaximizeToggle()
	case "PromptCycleNext":
		return cmdPromptCycleNext(kcmd.keyStr, kcmd.args...)
	case "PromptCyclePrev":
		return cmdPromptCyclePrev(kcmd.keyStr, kcmd.args...)
	case "PromptSelect":
		return cmdPromptSelect(kcmd.args...)
	case "Quit":
		return cmdQuit()
	case "Workspace":
		return cmdWorkspace(kcmd.args...)
	case "WorkspacePrefix":
		return cmdWorkspacePrefix(kcmd.args...)
	case "WorkspaceLeft":
		return cmdWorkspaceLeft()
	case "WorkspaceLeftWithClient":
		return cmdWorkspaceLeftWithClient()
	case "WorkspaceRight":
		return cmdWorkspaceRight()
	case "WorkspaceRightWithClient":
		return cmdWorkspaceRightWithClient()
	}

	logWarning.Printf("Unrecognized key command: %s", kcmd)

	return nil
}

// Start implementation of all command functions

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

func cmdHeadFocus(args ...string) func() {
	if len(args) < 1 {
		logWarning.Printf("Improper use of HeadFocus command.")
		logWarning.Printf("Usage: HeadFocus head_number [Greedy]")
		return nil
	}

	headNum64, err := strconv.ParseInt(args[0], 0, 0)
	if err != nil {
		logWarning.Printf("Improper use of HeadFocus command.")
		logWarning.Printf("'%s' is not a valid head number.", args[0])
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
			logWarning.Printf("The 'HeadFocus' command could not find "+
				"head number '%d'.", headNum)
			return
		}

		greedy := false
		if len(args) > 1 && args[1] == "greedy" {
			greedy = true
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

func cmdPromptCycleNext(keyStr string, args ...string) func() {
	activeWrk, visible, iconified := false, false, true
	if len(args) > 0 {
		switch args[0] {
		case "clientsworkspace":
			activeWrk = true
		case "clientsmonitors":
			visible = true
		default:
			logWarning.Printf("Unrecognized argument '%s' for PromptCycleNext "+
				"command", args[0])
			logWarning.Print("Usage: PromptCycleNext ClientListName")
			return nil
		}
	}
	return func() {
		PROMPTS.cycle.next(keyStr, activeWrk, visible, iconified)
	}
}

func cmdPromptCyclePrev(keyStr string, args ...string) func() {
	activeWrk, visible, iconified := false, false, true
	if len(args) > 0 {
		switch args[0] {
		case "clientsworkspace":
			activeWrk = true
		case "clientsmonitors":
			visible = true
		default:
			logWarning.Printf("Unrecognized argument '%s' for PromptCyclePrev "+
				"command", args[0])
			logWarning.Print("Usage: PromptCyclePrev ClientListName")
			return nil
		}
	}

	return func() {
		PROMPTS.cycle.prev(keyStr, activeWrk, visible, iconified)
	}
}

func cmdPromptSelect(args ...string) func() {
	usage := "Usage: PromptSelect ListName [Substring | Prefix]"

	if len(args) < 1 {
		logWarning.Printf("Improper use of PromptSelect command.")
		logWarning.Print(usage)
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
		f = promptSelectListWorkspaces
	default:
		logWarning.Printf("Unrecognized argument '%s' for PromptSelect "+
			"command", args[0])
		logWarning.Print(usage)
		return nil
	}

	return func() {
		PROMPTS.slct.show(f, prefixSearch)
	}
}

func cmdQuit() func() {
	return func() {
		logMessage.Println("The User has told us to quit.")
		X.Quit()
	}
}

func cmdWorkspace(args ...string) func() {
	if len(args) < 1 {
		logWarning.Printf("Improper use of Workspace command.")
		logWarning.Printf("Usage: Workspace name [Greedy]")
		return nil
	}

	return func() {
		wrk := WM.WrkFind(args[0])
		if wrk == nil {
			logWarning.Printf("The 'Workspace' command could not find "+
				"workspace '%s'.", args[0])
			return
		}

		greedy := false
		if len(args) > 1 && args[1] == "greedy" {
			greedy = true
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
func cmdWorkspacePrefix(args ...string) func() {
	if len(args) < 1 {
		logWarning.Printf("Improper use of WorkspacePrefix command.")
		logWarning.Printf("Usage: WorkspacePrefix prefix [Greedy]")
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
			logWarning.Printf("The 'WorkspacePrefix' command could not find "+
				"a non-visible workspace with prefix '%s'.",
				args[0])
			return
		}

		greedy := false
		if len(args) > 1 && args[1] == "greedy" {
			greedy = true
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
