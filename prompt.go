package main

import (
	"github.com/BurntSushi/wingo/focus"
	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/prompt"
	"github.com/BurntSushi/wingo/workspace"
)

type prompts struct {
	cycle *prompt.Cycle
	slct  *prompt.Select

	slctVisible, slctHidden *prompt.SelectGroupItem
}

func newPrompts() prompts {
	cycleConfig := prompt.CycleConfig{
		Grab:       true,
		CancelKey:  wingo.conf.cancelKey,
		ConfirmKey: wingo.conf.confirmKey,
	}
	selectConfig := prompt.SelectConfig{
		CancelKey:    wingo.conf.cancelKey,
		BackspaceKey: wingo.conf.backspaceKey,
		ConfirmKey:   wingo.conf.confirmKey,
		TabKey:       wingo.conf.tabKey,
	}
	ps := prompts{
		cycle: prompt.NewCycle(X, wingo.theme.prompt.CycleTheme(),
			cycleConfig),
		slct: prompt.NewSelect(X, wingo.theme.prompt.SelectTheme(),
			selectConfig),
	}
	ps.slctVisible = ps.slct.AddGroup(ps.slct.NewStaticGroup("Visible"))
	ps.slctHidden = ps.slct.AddGroup(ps.slct.NewStaticGroup("Hidden"))
	return ps
}

func filterClient(client *client, activeWrk, visible, iconified bool) bool {
	if activeWrk && !client.workspace.IsActive() {
		return false
	}
	if visible && !client.workspace.IsVisible() {
		return false
	}
	if !iconified && client.iconified {
		return false
	}
	return true
}

func showCycleClient(keyStr string, activeWrk, visible, iconified bool) {
	items := make([]*prompt.CycleItem, 0, len(focus.Clients))
	for i := len(focus.Clients) - 1; i >= 0; i-- {
		switch client := focus.Clients[i].(type) {
		case *client:
			if filterClient(client, activeWrk, visible, iconified) {
				items = append(items, client.prompts.cycle)
			}
		default:
			logger.Error.Panicf(
				"Client type %T not supported for cycle prompt.", client)
		}
	}
	wingo.prompts.cycle.Show(wingo.workspace().Geom(), keyStr, items)
}

func showSelectClient(tabComp int, activeWrk, visible, iconified bool) {
	allWrks := wingo.heads.Workspaces()

	groups := make([]*prompt.SelectShowGroup, len(allWrks))
	for i, wrk := range allWrks {
		items := make([]*prompt.SelectItem, 0, len(focus.Clients))
		for i := len(focus.Clients) - 1; i >= 0; i-- {
			switch client := focus.Clients[i].(type) {
			case *client:
				if client.workspace != wrk {
					continue
				}
				if !filterClient(client, activeWrk, visible, iconified) {
					continue
				}
				items = append(items, client.prompts.slct)
			default:
				logger.Error.Panicf(
					"Client type %T not supported for select prompt.", client)
			}
		}
		groups[i] = wrk.PromptSlctGroup.ShowGroup(items)
	}

	wingo.prompts.slct.Show(wingo.workspace().Geom(), tabComp, groups, nil)
}

func showSelectWorkspace(tabComp int, data workspace.SelectData) {
	allWrks := wingo.heads.Workspaces()
	visibles := wingo.heads.VisibleWorkspaces()

	wrksVisible := make([]*prompt.SelectItem, 0, len(allWrks))
	wrksHidden := make([]*prompt.SelectItem, 0, len(allWrks))
	for _, wrk := range visibles {
		wrksVisible = append(wrksVisible, wrk.PromptSlctItem)
	}
	for _, wrk := range allWrks {
		if !wrk.IsVisible() {
			wrksHidden = append(wrksHidden, wrk.PromptSlctItem)
		}
	}

	groups := []*prompt.SelectShowGroup{
		wingo.prompts.slctVisible.ShowGroup(wrksVisible),
		wingo.prompts.slctHidden.ShowGroup(wrksHidden),
	}
	wingo.prompts.slct.Show(wingo.workspace().Geom(), tabComp, groups, data)
}
