package xclient

import (
	"github.com/BurntSushi/xgbutil/xrect"

	"github.com/BurntSushi/wingo-conc/frame"
	"github.com/BurntSushi/wingo-conc/heads"
)

func (c *Client) newClientStates() map[string]clientState {
	return make(map[string]clientState, 10)
}

type clientState struct {
	geom      xrect.Rect
	headGeom  xrect.Rect
	frame     frame.Frame
	maximized bool
}

func (c *Client) newClientState() clientState {
	s := clientState{
		geom:      xrect.New(xrect.Pieces(c.frame.Geom())),
		headGeom:  nil,
		frame:     c.frame,
		maximized: c.maximized,
	}
	if c.workspace.IsVisible() {
		s.headGeom = xrect.New(xrect.Pieces(c.workspace.HeadGeom()))
	}
	return s
}

func (c *Client) HasState(name string) bool {
	_, ok := c.states[name]
	return ok
}

// Don't save when moving or resizing.
// Also don't save when client's workspace isn't visible.
func (c *Client) SaveState(name string) {
	if c.workspace == nil {
		return
	}
	if !c.workspace.IsVisible() {
		return
	}
	c.states[name] = c.newClientState()
}

func (c *Client) CopyState(src, dest string) {
	if from, ok := c.states[src]; ok {
		c.states[dest] = from
	}
}

// Don't revert to regular geometry when moving/resizing. We can still revert
// the frame or the maximized state, though.
// Also don't load *ever* when client's workspace isn't visible.
func (c *Client) LoadState(name string) {
	if !c.workspace.IsVisible() {
		return
	}

	s, ok := c.states[name]
	if !ok {
		return
	}

	// We're committed now to at least reverting frame.
	c.frames.set(s.frame)

	// Delete the state entry here. We do this because this state may be
	// re-added when maximizing or moving the window. (Like "last-floating".)
	delete(c.states, name)

	// If the state calls for maximization, maximize the client and be done.
	if s.maximized {
		c.maximize()
		return
	}

	// Finally, if we're here and the client isn't being moved/resized, then
	// we can revert to the geometry specified by the state, adjusted for the
	// head geometry used when capturing that state.
	if c.frame.Moving() || c.frame.Resizing() {
		return
	}

	// We also need to check if the client has just moved via its layout.
	// If so, don't continue.
	if c.moving || c.resizing {
		return
	}

	if s.headGeom != nil && c.workspace.HeadGeom() != s.headGeom {
		s.geom = heads.Convert(s.geom, s.headGeom, c.workspace.HeadGeom())
	}
	c.LayoutMoveResize(s.geom.X(), s.geom.Y(),
		s.geom.Width(), s.geom.Height())
}

func (c *Client) DeleteState(name string) {
	if _, ok := c.states[name]; ok {
		delete(c.states, name)
	}
}
