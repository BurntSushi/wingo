package main

import (
	"github.com/BurntSushi/wingo/prompt"
)

type prompts struct {
	cycle *prompt.Cycle
	slct  *prompt.Select
}

func newPrompts() prompts {
	return prompts{
		cycle: prompt.NewCycle(X, prompt.DefaultCycleTheme,
			prompt.DefaultCycleConfig),
		slct: prompt.NewSelect(X, prompt.DefaultSelectTheme,
			prompt.DefaultSelectConfig),
	}
}

func showPromptCycle(keyStr string, activeWrk, visible, iconified bool) {
	items := make([]*prompt.CycleItem, 0, len(WM.focus))
	for i := len(WM.focus) - 1; i >= 0; i-- {
		client := WM.focus[i]
		if activeWrk && !client.workspace.active {
			continue
		}
		if visible && !client.workspace.visible() {
			continue
		}
		if !iconified && client.iconified {
			continue
		}
		items = append(items, client.prompts.cycle)
	}

	PROMPTS.cycle.Show(WM.headActive(), keyStr, items)
}
