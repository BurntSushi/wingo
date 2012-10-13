package wm

import (
	"fmt"
	"time"

	"github.com/BurntSushi/wingo/focus"
	"github.com/BurntSushi/wingo/prompt"
)

type AllPrompts struct {
	Cycle   *prompt.Cycle
	Slct    *prompt.Select
	Input   *prompt.Input
	Message *prompt.Message

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
	inputConfig := prompt.InputConfig{
		CancelKey:    Config.CancelKey,
		BackspaceKey: Config.BackspaceKey,
		ConfirmKey:   Config.ConfirmKey,
	}
	msgConfig := prompt.MessageConfig{
		CancelKey:  Config.CancelKey,
		ConfirmKey: Config.ConfirmKey,
	}
	ps := AllPrompts{
		Cycle:   prompt.NewCycle(X, Theme.Prompt.CycleTheme(), cycleConfig),
		Slct:    prompt.NewSelect(X, Theme.Prompt.SelectTheme(), selectConfig),
		Input:   prompt.NewInput(X, Theme.Prompt.InputTheme(), inputConfig),
		Message: prompt.NewMessage(X, Theme.Prompt.MessageTheme(), msgConfig),
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

func PopupError(format string, vals ...interface{}) {
	if !Config.ShowErrors {
		return
	}

	nada := func(msg *prompt.Message) {}
	Prompts.Message.Hide()

	msg := fmt.Sprintf(format, vals...)
	Prompts.Message.Show(Workspace().Geom(), msg, 0, nada)
}

func FYI(format string, vals ...interface{}) {
	if !Config.ShowFyi {
		return
	}

	nada := func(msg *prompt.Message) {}
	Prompts.Message.Hide()

	timeout := time.Duration(Config.PopupTime) * time.Millisecond
	msg := fmt.Sprintf(format, vals...)
	Prompts.Message.Show(Workspace().Geom(), msg, timeout, nada)
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

func ShowSelectClient(tabComp int, activeWrk, visible, iconified bool,
	data interface{}) {

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

	Prompts.Slct.Show(Workspace().Geom(), tabComp, groups, data)
}

func ShowSelectWorkspace(tabComp int, data interface{}) {
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
