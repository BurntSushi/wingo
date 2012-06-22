package main

import (
	"github.com/BurntSushi/wingo/focus"
	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/prompt"
)

type prompts struct {
	cycle *prompt.Cycle
	slct  *prompt.Select
}

func newPrompts() prompts {
	return prompts{
		cycle: prompt.NewCycle(X, wingo.theme.prompt.CycleTheme(),
			prompt.DefaultCycleConfig),
		slct: prompt.NewSelect(X, wingo.theme.prompt.SelectTheme(),
			prompt.DefaultSelectConfig),
	}
}

func showPromptCycle(keyStr string, activeWrk, visible, iconified bool) {
	items := make([]*prompt.CycleItem, 0, len(focus.Clients))
	for i := len(focus.Clients) - 1; i >= 0; i-- {
		switch client := focus.Clients[i].(type) {
		case *client:
			if activeWrk && !client.workspace.IsActive() {
				continue
			}
			if visible && !client.workspace.IsVisible() {
				continue
			}
			if !iconified && client.iconified {
				continue
			}
			items = append(items, client.prompts.cycle)
		default:
			logger.Error.Panicf("Client type %T not support for cycle prompt.",
				client)
		}
	}

	wingo.prompts.cycle.Show(wingo.workspace().Geom(), keyStr, items)
}
