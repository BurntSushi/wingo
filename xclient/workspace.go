package xclient

import (
	"github.com/BurntSushi/wingo/workspace"
)

func (c *Client) ShouldForceFloating() bool {
	return c.floating ||
		c.sticky ||
		c.transientFor != nil ||
		c.primaryType != clientTypeNormal
}

func (c *Client) FloatingToggle() {
	// Doesn't work on sticky windows. They are already floating.
	if wrk, ok := c.Workspace().(*workspace.Workspace); ok {
		c.floating = !c.floating
		wrk.CheckFloatingStatus(c)
	}
}

func (c *Client) Workspace() workspace.Workspacer {
	return c.workspace
}

func (c *Client) WorkspaceSet(workspace *workspace.Workspace) {
	c.workspace = workspace
}

func (c *Client) Iconified() bool {
	return c.iconified
}

func (c *Client) IconifiedSet(iconified bool) {
	c.iconified = iconified
}
