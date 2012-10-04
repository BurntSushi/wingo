package wm

type CommandHacks struct {
	MouseResizeDirection     func(cmdStr string) (string, error)
	CycleClientRunWithKeyStr func(keyStr, cmdStr string) (func(), error)
}
