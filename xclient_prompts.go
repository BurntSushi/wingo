package main

import (
	"github.com/BurntSushi/xgbutil/xgraphics"

	"github.com/BurntSushi/wingo/focus"
	"github.com/BurntSushi/wingo/prompt"
	"github.com/BurntSushi/wingo/stack"
)

type clientPrompts struct {
	client *client
	cycle  *prompt.CycleItem
	slct   *prompt.SelectItem
}

func (c *client) newClientPrompts() clientPrompts {
	return clientPrompts{
		client: c,
		cycle:  wingo.prompts.cycle.AddChoice(c),
		slct:   wingo.prompts.slct.AddChoice(c),
	}
}

func (p *clientPrompts) destroy() {
	p.cycle.Destroy()
	p.slct.Destroy()
}

func (p *clientPrompts) updateIcon() {
	p.cycle.UpdateImage()
}

func (p *clientPrompts) updateName() {
	if p.cycle == nil || p.slct == nil {
		return
	}
	p.cycle.UpdateText()
	p.slct.UpdateText()
}

// Satisfy the prompt.CycleChoice interface.

func (c *client) CycleIsActive() bool {
	return !c.iconified
}

func (c *client) CycleImage() *xgraphics.Image {
	theme := wingo.theme.prompt.CycleTheme()
	return c.Icon(theme.IconSize, theme.IconSize)
}

func (c *client) CycleText() string {
	return c.String()
}

func (c *client) CycleSelected() {
	if c.iconified {
		c.workspace.IconifyToggle(c)
	}
	focus.Focus(c)
	stack.Raise(c)
}

func (c *client) CycleHighlighted() {
}

// Satisfy the prompt.SelectChoice interface.

func (c *client) SelectText() string {
	return c.String()
}

func (c *client) SelectSelected() {
	focus.Focus(c)
	stack.Raise(c)
}

func (c *client) SelectHighlighted() {
}
