package main

import (
	"github.com/BurntSushi/xgbutil/xgraphics"

	"github.com/BurntSushi/wingo/prompt"
)

type clientPrompts struct {
	client *client
	cycle  *prompt.CycleItem
	slct   *prompt.SelectItem
}

func newClientPrompts(c *client) *clientPrompts {
	return &clientPrompts{
		client: c,
		cycle:  PROMPTS.cycle.AddChoice(c),
		slct:   PROMPTS.slct.AddChoice(c),
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
	p.cycle.UpdateText()
	p.slct.UpdateText()
}

// Satisfy the prompt.CycleChoice interface.

func (c *client) CycleIsActive() bool {
	return !c.iconified
}

func (c *client) CycleImage() *xgraphics.Image {
	return c.Icon(100, 100)
}

func (c *client) CycleText() string {
	return c.String()
}

func (c *client) CycleSelected() {
	if c.iconified {
		c.IconifyToggle()
	}
	c.Focus()
	c.Raise()
}

func (c *client) CycleHighlighted() {
}

// Satisfy the prompt.SelectChoice interface.

func (c *client) SelectText() string {
	return c.String()
}

func (c *client) SelectSelected() {
	c.Focus()
	c.Raise()
}

func (c *client) SelectHighlighted() {
}
