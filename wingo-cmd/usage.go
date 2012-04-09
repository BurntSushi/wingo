package main

import (
	"fmt"

	"github.com/BurntSushi/wingo/logger"
)

var usage = map[string]string{
	"Close": "[win:window-id]",
	"FrameBorders": "[win:window-id]",
	"FrameFull": "[win:window-id]",
	"FrameNada": "[win:window-id]",
	"FrameSlim": "[win:window-id]",
	"HeadFocus": "head-number",
	"HeadFocusWithClient": "head-number [win:window-id]",
	"MaximizeToggle": "[win:window-id]",
	"Minimize": "[win:window-id]",
	"PromptCycleNext": "client-list-name",
	"PromptCyclePrev": "client-list-name",
	"PromptSelect": "list-name [Prefix | Substring]",
	"Quit": "",
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

