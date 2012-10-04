package wm

import (
	"github.com/BurntSushi/wingo/focus"
	"github.com/BurntSushi/wingo/prompt"
	"github.com/BurntSushi/wingo/workspace"
)

type AllPrompts struct {
	Cycle *prompt.Cycle
	Slct  *prompt.Select

	slctVisible, slctHidden *prompt.SelectGroupItem
}

func newPrompts() AllPrompts {
	cycleConfig := prompt.CycleConfig{
		Grab:       true,
		CancelKey:  Config.CancelKey,
		ConfirmKey: Config.ConfirmKey,
	}
	selectConfig := prompt.SelectConfig{
		CancelKey:    Config.CancelKey,
		BackspaceKey: Config.BackspaceKey,
		ConfirmKey:   Config.ConfirmKey,
		TabKey:       Config.TabKey,
	}
	ps := AllPrompts{
		Cycle: prompt.NewCycle(X, Theme.Prompt.CycleTheme(),
			cycleConfig),
		Slct: prompt.NewSelect(X, Theme.Prompt.SelectTheme(),
			selectConfig),
	}
	ps.slctVisible = ps.Slct.AddGroup(ps.Slct.NewStaticGroup("Visible"))
	ps.slctHidden = ps.Slct.AddGroup(ps.Slct.NewStaticGroup("Hidden"))
	return ps
}

func filterClient(client Client, activeWrk, visible, iconified bool) bool {
	if activeWrk && !client.Workspace().IsActive() {
		return false
	}
	if visible && !client.Workspace().IsVisible() {
		return false
	}
	if !iconified && client.Iconified() {
		return false
	}
	return true
}

func ShowCycleClient(keyStr string, activeWrk, visible, iconified bool) {
	items := make([]*prompt.CycleItem, 0, len(focus.Clients))
	for i := len(focus.Clients) - 1; i >= 0; i-- {
		client := focus.Clients[i].(Client)
		if !filterClient(client, activeWrk, visible, iconified) {
			continue
		}
		items = append(items, client.CycleItem())
	}
	Prompts.Cycle.Show(Workspace().Geom(), keyStr, items)
}

func ShowSelectClient(tabComp int, activeWrk, visible, iconified bool) {
	allWrks := Heads.Workspaces.Wrks

	groups := make([]*prompt.SelectShowGroup, len(allWrks))
	for i, wrk := range allWrks {
		items := make([]*prompt.SelectItem, 0, len(focus.Clients))
		for i := len(focus.Clients) - 1; i >= 0; i-- {
			client := focus.Clients[i].(Client)
			if client.Workspace() != wrk {
				continue
			}
			if !filterClient(client, activeWrk, visible, iconified) {
				continue
			}
			items = append(items, client.SelectItem())
		}
		groups[i] = wrk.PromptSlctGroup.ShowGroup(items)
	}

	Prompts.Slct.Show(Workspace().Geom(), tabComp, groups, nil)
}

func ShowSelectWorkspace(tabComp int, data workspace.SelectData) {
	allWrks := Heads.Workspaces.Wrks
	visibles := Heads.VisibleWorkspaces()

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
		Prompts.slctVisible.ShowGroup(wrksVisible),
		Prompts.slctHidden.ShowGroup(wrksHidden),
	}
	Prompts.Slct.Show(Workspace().Geom(), tabComp, groups, data)
}
