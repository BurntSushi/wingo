package xclient

import (
	"github.com/BurntSushi/xgbutil/xgraphics"

	"github.com/cshapeshifter/wingo/prompt"
	"github.com/cshapeshifter/wingo/wm"
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
	if c.Iconified() {
		c.IconifyToggle()
	}
	c.Focus()
	c.Raise()
}

func (c *Client) CycleHighlighted() {
}

// Satisfy the prompt.SelectChoice interface.

type SelectData struct {
	Selected    func(c *Client)
	Highlighted func(c *Client)
}

func (c *Client) SelectText() string {
	return c.String()
}

func (c *Client) SelectSelected(data interface{}) {
	if f := data.(SelectData).Selected; f != nil {
		f(c)
	}
}

func (c *Client) SelectHighlighted(data interface{}) {
	if f := data.(SelectData).Highlighted; f != nil {
		f(c)
	}
}
