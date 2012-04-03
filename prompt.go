package main

import (
    "github.com/BurntSushi/xgbutil"
    "github.com/BurntSushi/xgbutil/xevent"
)

type prompts struct {
    cycle *promptCycle
}

func promptsInitialize() {
    PROMPTS = prompts{
        cycle: newPromptCycle(),
    }
}

func (c *client) promptAdd() {
    c.promptCycleAdd()
}

func rootTryStopGrab(X *xgbutil.XUtil, ev xevent.KeyReleaseEvent) {
    logDebug.Printf("State: %d", ev.State)
    logDebug.Printf("Detail: %d", ev.Detail)
    println("")
}

