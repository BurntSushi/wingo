package main

func (c *client) toggleForceFloating() {
	c.forceFloating = !c.forceFloating
	c.layoutSet()
	c.workspace.tile()
}

func (c *client) floating() bool {
	return c.forceFloating || c.transientFor != 0 || !c.normal
}

// layoutSet determines whether client MUST be floating or not.
// If a client doesn't have to be floating, then it is *always* stored
// in the tilers' state. The presumption is, if a client doesn't have to be
// floating, then when a tiler is active, it is being tiled.
// If a client MUST be floating, then the tilers should not know about it.
// DO NOT INVOKE A TILE COMMAND HERE. Try it. I dare you.
func (c *client) layoutSet() {
	if c.floating() {
		c.workspace.tilersRemove(c)
	} else {
		c.workspace.tilersAdd(c)
	}
}

func (c client) layout() layout {
	if c.floating() {
		return c.workspace.floaters[0]
	}
	return c.workspace.layout()
}
