package xclient

import (
	"github.com/BurntSushi/xgbutil/xgraphics"

	"github.com/BurntSushi/wingo/focus"
	"github.com/BurntSushi/wingo/prompt"
	"github.com/BurntSushi/wingo/stack"
	"github.com/BurntSushi/wingo/wm"
)

type clientPrompts struct {
	client *Client
	cycle  *prompt.CycleItem
	slct   *prompt.SelectItem
}

func (c *Client) CycleItem() *prompt.CycleItem {
	return c.prompts.cycle
}

func (c *Client) SelectItem() *prompt.SelectItem {
	return c.prompts.slct
}

func (c *Client) newClientPrompts() clientPrompts {
	return clientPrompts{
		client: c,
		cycle:  wm.Prompts.Cycle.AddChoice(c),
		slct:   wm.Prompts.Slct.AddChoice(c),
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

func (c *Client) CycleIsActive() bool {
	return !c.iconified
}

func (c *Client) CycleImage() *xgraphics.Image {
	theme := wm.Theme.Prompt.CycleTheme()
	return c.Icon(theme.IconSize, theme.IconSize)
}

func (c *Client) CycleText() string {
	return c.String()
}

func (c *Client) CycleSelected() {
	if c.iconified {
		c.workspace.IconifyToggle(c)
	}
	focus.Focus(c)
	stack.Raise(c)
}

func (c *Client) CycleHighlighted() {
}

// Satisfy the prompt.SelectChoice interface.

func (c *Client) SelectText() string {
	return c.String()
}

func (c *Client) SelectSelected(data interface{}) {
	focus.Focus(c)
	stack.Raise(c)
}

func (c *Client) SelectHighlighted(data interface{}) {
}
