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
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"

	"github.com/BurntSushi/wingo/logger"
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
			kcmd.attach(commandFun(kcmd.keyStr, kcmd.cmd, kcmd.args...))
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
		mods, _, _ := keybind.ParseString(X, kcmd.keyStr)
		if mods == 0 {
			logger.Warning.Printf("Sorry but the key binding '%s' for the %s "+
				"command is invalid. It must have a modifier "+
				"to work properly. i.e., Mod1-tab where 'Mod1' "+
				"is the modifier.", kcmd.keyStr, kcmd.cmd)
			return
		}

		keybind.KeyPressFun(
			func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
				run()
			}).Connect(X, wingo.root.Id, kcmd.keyStr, true)
		keybind.KeyPressFun(
			func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
				run()
			}).Connect(X, X.Dummy(), kcmd.keyStr, true)
	} else {
		if kcmd.down {
			keybind.KeyPressFun(
				func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
					run()
				}).Connect(X, wingo.root.Id, kcmd.keyStr, true)
		} else {
			keybind.KeyReleaseFun(
				func(X *xgbutil.XUtil, ev xevent.KeyReleaseEvent) {
					run()
				}).Connect(X, wingo.root.Id, kcmd.keyStr, true)
		}
	}
}
