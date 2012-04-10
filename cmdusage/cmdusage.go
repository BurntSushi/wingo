package cmdusage

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xprop"

	"github.com/BurntSushi/wingo/logger"
)

var usage = map[string]string{
	"Close": "[win:window-id]",
	"Focus": "[win:window-id]",
	"FrameBorders": "[win:window-id]",
	"FrameFull": "[win:window-id]",
	"FrameNada": "[win:window-id]",
	"FrameSlim": "[win:window-id]",
	"HeadFocus": "head-number",
	"HeadFocusWithClient": "head-number [win:window-id]",
	"MaximizeToggle": "[win:window-id]",
	"Minimize": "[win:window-id]",
	"Move": "x y [win:window-id]",
	"PromptCycleNext": "client-list-name",
	"PromptCyclePrev": "client-list-name",
	"PromptSelect": "list-name [Prefix | Substring]",
	"Quit": "",
	"Raise": "[win:window-id]",
	"Resize": "width height [win:window-id]",
	"Workspace": "workspace-name [Greedy]",
	"WorkspacePrefix": "workspace-name-prefix [Greedy]",
	"WorkspacePrefixWithClient": "workspace-name-prefix [Greedy]",
	"WorkspaceSetClient": "workspace-name [win:window-id]",
	"WorkspaceWithClient": "workspace-name [Greedy]",
	"WorkspaceLeft": "",
	"WorkspaceLeftWithClient": "",
	"WorkspaceRight": "",
	"WorkspaceRightWithClient": "",
}

func MaybeUsage(cmd string, action func()) func() {
	if action == nil {
		ShowUsage(cmd)
	}
	return action
}

func ShowUsage(cmd string) {
	argUsage, ok := usage[cmd]
	if !ok {
		return
	}

	var s string
	if len(argUsage) > 0 {
		s = fmt.Sprintf("%s %s", cmd, argUsage)
	} else {
		s = fmt.Sprintf("%s", cmd)
	}

	logger.Warning.Printf("Error: Invalid arguments for '%s' command.\n", cmd)
	logger.Warning.Printf("Usage: %s\n", s)
}

func CmdGet(X *xgbutil.XUtil) (string, error) {
	return xprop.PropValStr(xprop.GetProperty(X, X.RootWin(), "_WINGO_CMD"))
}

func CmdSet(X *xgbutil.XUtil, cmd string) error {
	return xprop.ChangeProp(X, X.RootWin(), 8,
		"_WINGO_CMD", "UTF8_STRING", []byte(cmd))
}

func StatusGet(X *xgbutil.XUtil) bool {
	status, err := xprop.PropValStr(xprop.GetProperty(X, X.RootWin(),
		"_WINGO_CMD_STATUS"))

	return err == nil && strings.ToLower(status) == "success"
}

func StatusSet(X *xgbutil.XUtil, status bool) {
	var statusStr string
	if status {
		statusStr = "Success"
	} else {
		statusStr = "Error"
	}

	// throw away the error
	xprop.ChangeProp(X, X.RootWin(), 8, "_WINGO_CMD_STATUS", "UTF8_STRING",
		[]byte(statusStr))
}
