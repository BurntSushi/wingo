package main

import (
	"github.com/BurntSushi/wingo/workspace"
)

func (c *client) ShouldForceFloating() bool {
	return c.floating ||
		c.transientFor != nil ||
		c.primaryType != clientTypeNormal
}

func (c *client) Workspace() *workspace.Workspace {
	return c.workspace
}

func (c *client) WorkspaceSet(workspace *workspace.Workspace) {
	c.workspace = workspace
}

func (c *client) Iconified() bool {
	return c.iconified
}

func (c *client) IconifiedSet(iconified bool) {
	c.iconified = iconified
}
