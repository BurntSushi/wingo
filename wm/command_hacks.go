package wm

import "github.com/BurntSushi/gribble"

type CommandHacks struct {
	MouseResizeDirection     func(cmd gribble.Command) string
	CycleClientRunWithKeyStr func(keyStr string, cmd gribble.Command) func()
}
