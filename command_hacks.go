package main

import (
	"fmt"

	"github.com/BurntSushi/gribble"

	"github.com/BurntSushi/wingo/commands"
	"github.com/BurntSushi/wingo/wm"
)

func newHacks() wm.CommandHacks {
	return wm.CommandHacks{
		MouseResizeDirection:     mouseResizeDirection,
		CycleClientRunWithKeyStr: cycleClientRunWithKeyStr,
	}
}

func mouseResizeDirection(cmd gribble.Command) string {
	return cmd.(*commands.MouseResize).Direction
}

func cycleClientRunWithKeyStr(keyStr string, cmd gribble.Command) func() {
	var run func() = nil
	switch t := cmd.(type) {
	case *commands.CycleClientNext:
		run = func() { t.RunWithKeyStr(keyStr) }
	case *commands.CycleClientPrev:
		run = func() { t.RunWithKeyStr(keyStr) }
	default:
		panic(fmt.Sprintf("bug: unknown type %T", t))
	}
	return run
}
