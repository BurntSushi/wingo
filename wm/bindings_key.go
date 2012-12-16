package wm

import (
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"

	"github.com/cshapeshifter/wingo/logger"
)

type keyCommand struct {
	cmdStr  string
	cmdName string
	args    []string
	down    bool // 'up' when false
	keyStr  string
}

func keybindings() {
	for _, kcmds := range Config.key {
		for _, kcmd := range kcmds {
			kcmd.attach()
		}
	}
}

func (kcmd keyCommand) attach() {
	if kcmd.cmdName == "CycleClientPrev" || kcmd.cmdName == "CycleClientNext" {
		// We've got to parse the key string first and make sure
		// there are some modifiers; otherwise this utterly fails!
		mods, _, _ := keybind.ParseString(X, kcmd.keyStr)
		if mods == 0 {
			logger.Warning.Printf("Sorry but the key binding '%s' for the %s "+
				"command is invalid. It must have a modifier "+
				"to work properly. i.e., Mod1-tab where 'Mod1' "+
				"is the modifier.", kcmd.keyStr, kcmd.cmdName)
			return
		}

		run, err := cmdHacks.CycleClientRunWithKeyStr(kcmd.keyStr, kcmd.cmdStr)
		if err != nil {
			logger.Warning.Printf("Could not setup %s: %s", kcmd.cmdName, err)
			return
		}

		err = keybind.KeyPressFun(
			func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
				go run()
			}).Connect(X, Root.Id, kcmd.keyStr, true)
		if err != nil {
			logger.Warning.Printf("Could not bind '%s': %s", kcmd.keyStr, err)
		}

		err = keybind.KeyPressFun(
			func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
				go run()
			}).Connect(X, X.Dummy(), kcmd.keyStr, true)
		if err != nil {
			logger.Warning.Printf("Could not bind '%s': %s", kcmd.keyStr, err)
		}
	} else {
		run := func() {
			go func() {
				_, err := gribbleEnv.Run(kcmd.cmdStr)
				if err != nil {
					logger.Warning.Println(err)
				}
			}()
		}
		if kcmd.down {
			err := keybind.KeyPressFun(
				func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
					run()
				}).Connect(X, Root.Id, kcmd.keyStr, true)
			if err != nil {
				logger.Warning.Printf("Could not bind '%s': %s",
					kcmd.keyStr, err)
			}
		} else {
			err := keybind.KeyReleaseFun(
				func(X *xgbutil.XUtil, ev xevent.KeyReleaseEvent) {
					run()
				}).Connect(X, Root.Id, kcmd.keyStr, true)
			if err != nil {
				logger.Warning.Printf("Could not bind '%s': %s",
					kcmd.keyStr, err)
			}
		}
	}
}
