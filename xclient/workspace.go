package xclient

import (
	"github.com/BurntSushi/wingo/workspace"
)

func (c *Client) ShouldForceFloating() bool {
	return c.floating ||
		c.transientFor != nil ||
		c.primaryType != clientTypeNormal
}

func (c *Client) Workspace() *workspace.Workspace {
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
