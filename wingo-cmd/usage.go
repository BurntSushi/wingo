package main

import (
	"fmt"

	"github.com/BurntSushi/wingo/logger"
)

var usage = map[string]string{
	"Close": "",
	"FrameBorders": "",
	"FrameFull": "",
	"FrameNada": "",
	"FrameSlim": "",
	"HeadFocus": "head-number",
	"HeadFocusWithClient": "head-number",
	"MaximizeToggle": "",
	"Minimize": "",
	"PromptCycleNext": "client-list-name",
	"PromptCyclePrev": "client-list-name",
	"PromptSelect": "list-name [Prefix | Substring]",
	"Quit": "",
	"Workspace": "workspace-name [Greedy]",
	"WorkspacePrefix": "workspace-name-prefix [Greedy]",
	"WorkspacePrefixWithClient": "workspace-name-prefix [Greedy]",
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

