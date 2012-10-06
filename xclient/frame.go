package xclient

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/motif"
	"github.com/BurntSushi/xgbutil/xrect"

	"github.com/BurntSushi/wingo/frame"
	"github.com/BurntSushi/wingo/layout"
	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/wm"
)

// shouldDecor returns false if the client has requested no frames or
// has a type that implies it shouldn't be decorated.
func (c *Client) shouldDecor() bool {
	if c.primaryType != clientTypeNormal {
		return false
	}
	if c.hasType("_NET_WM_WINDOW_TYPE_SPLASH") {
		return false
	}

	mh, err := motif.WmHintsGet(wm.X, c.Id())
	if err == nil && !motif.Decor(mh) {
		return false
	}

	return true
}

// refreshExtents sets the _NET_FRAME_EXTENTS property whenever the frame
// of a client changes. (Or on maximization.)
func (c *Client) refreshExtents() {
	exts := ewmh.FrameExtents{
		Left:   c.frame.Left(),
		Right:  c.frame.Right(),
		Top:    c.frame.Top(),
		Bottom: c.frame.Bottom(),
	}
	ewmh.FrameExtentsSet(wm.X, c.Id(), &exts)
}

// FrameFull switches this client's frame to the 'Full' frame.
func (c *Client) FrameFull() {
	c.frames.set(c.frames.full)
}

// FrameBorders switches this client's frame to the 'Borders' frame.
func (c *Client) FrameBorders() {
	c.frames.set(c.frames.borders)
}

// FrameSlim switches this client's frame to the 'Slim' frame.
func (c *Client) FrameSlim() {
	c.frames.set(c.frames.slim)
}

// FrameNada switches this client's frame to the 'Nada' frame.
func (c *Client) FrameNada() {
	c.frames.set(c.frames.nada)
}

// Frame returns the current frame in use by the client.
func (c *Client) Frame() frame.Frame {
	return c.frame
}

// Geom returns the geometry of the client window (not the frame window).
func (c *Client) ClientGeom() xrect.Rect {
	return c.win.Geom
}

// EnsureUnmax makes sure the client is not in a maximized state.
// It's useful when a particular operation that doesn't work in maximized mode
// overrides a client's maximized state. (Like issuing a tiling request.)
func (c *Client) EnsureUnmax() {
	c.unmaximize()
}

func (c *Client) MaximizeToggle() {
	if c.Maximized() {
		c.Unmaximize()
	} else {
		c.Maximize()
	}
}

func (c *Client) Maximize() {
	if !c.canMaxUnmax() {
		return
	}
	if !c.Maximized() {
		c.SaveState("before-maximize")
		c.maximize()
	}
}

func (c *Client) Unmaximize() {
	if !c.canMaxUnmax() {
		return
	}
	if c.Maximized() {
		c.unmaximize()
		c.LoadState("before-maximize")
	}
}

func (c *Client) maximize() {
	if !c.canMaxUnmax() {
		return
	}

	c.maximized = true
	c.addState("_NET_WM_STATE_MAXIMIZE_HORZ")
	c.addState("_NET_WM_STATE_MAXIMIZE_VERT")

	c.frames.maximize()

	// Resize outside of the constraints of a layout.
	g := c.Workspace().Geom()
	c.MoveResize(false, g.X(), g.Y(), g.Width(), g.Height())

	// Since we moved outside of the layout, we have to save the last
	// floating state our selves.
	c.SaveState("last-floating")
}

func (c *Client) unmaximize() {
	if c.Workspace() == nil || !c.Workspace().IsVisible() {
		return
	}
	if c.maximized {
		c.maximized = false
		c.removeState("_NET_WM_STATE_MAXIMIZE_HORZ")
		c.removeState("_NET_WM_STATE_MAXIMIZE_VERT")
		c.frames.unmaximize()
	}
}

func (c *Client) canMaxUnmax() bool {
	if c.Workspace() == nil || !c.Workspace().IsVisible() {
		return false
	}
	if _, ok := c.Layout().(layout.Floater); !ok {
		return false
	}
	if c.fullscreen {
		return false
	}
	return true
}

func (c *Client) HeadGeom() xrect.Rect {
	return c.workspace.Geom()
}

func (c *Client) FramePieceMouseSetup(piece string, pieceid xproto.Window) {
	wm.FramePieceMouseSetup(c, piece, pieceid)
}

// GravitizeX adjusts the x coordinate of a window's position using the gravity
// value set. Gravity refers to the way (x, y) coordinates are interpreted with
// respect to a client's decorations. See Section 4.1.2.3 of the ICCCM for more
// details.
func (c *Client) GravitizeX(x int, gravity int) int {
	// Don't do anything if there's no gravity options set and we're
	// trying to infer gravity.
	// This is equivalent to setting NorthWest gravity
	if gravity > -1 && c.nhints.Flags&icccm.SizeHintPWinGravity == 0 {
		return x
	}

	// Otherwise, we're either inferring gravity (from normal hints), or
	// using some forced notion of gravity (probably from EWMH stuff)
	var g int
	if gravity > -1 {
		g = gravity
	} else {
		g = int(c.nhints.WinGravity)
	}

	f := c.Frame()
	switch {
	case g == xproto.GravityStatic || g == xproto.GravityBitForget:
		x -= f.Left()
	case g == xproto.GravityNorth || g == xproto.GravitySouth ||
		g == xproto.GravityCenter:
		x -= abs(f.Left()-f.Right()) / 2
	case g == xproto.GravityNorthEast || g == xproto.GravityEast ||
		g == xproto.GravitySouthEast:
		x -= f.Left() + f.Right()
	}

	return x
}

// GravitizeY adjusts the y coordinate of a window's position using the gravity
// value set. Gravity refers to the way (x, y) coordinates are interpreted with
// respect to a client's decorations. See Section 4.1.2.3 of the ICCCM for more
// details.
func (c *Client) GravitizeY(y int, gravity int) int {
	// Don't do anything if there's no gravity options set and we're
	// trying to infer gravity.
	// This is equivalent to setting NorthWest gravity
	if gravity > -1 && c.nhints.Flags&icccm.SizeHintPWinGravity == 0 {
		return y
	}

	// Otherwise, we're either inferring gravity (from normal hints), or
	// using some forced notion of gravity (probably from EWMH stuff)
	var g int
	if gravity > -1 {
		g = gravity
	} else {
		g = int(c.nhints.WinGravity)
	}

	f := c.Frame()
	switch {
	case g == xproto.GravityStatic || g == xproto.GravityBitForget:
		y -= f.Top()
	case g == xproto.GravityEast || g == xproto.GravityWest ||
		g == xproto.GravityCenter:
		y -= abs(f.Top()-f.Bottom()) / 2
	case g == xproto.GravitySouthEast || g == xproto.GravitySouth ||
		g == xproto.GravitySouthWest:
		y -= f.Top() + f.Bottom()
	}

	return y
}

// ValidateHeight use's a clients min/max height and height increment values
// from the WM_NORMAL_HINTS property to change 'height' to a valid height.
// See Section 4.1.2.3 of the ICCCM for more details.
func (c *Client) ValidateHeight(height int) int {
	return c.validateSize(height, c.nhints.HeightInc, c.nhints.BaseHeight,
		c.nhints.MinHeight, c.nhints.MaxHeight)
}

// ValidateWidth use's a clients min/max width and width increment values
// from the WM_NORMAL_HINTS property to change 'width' to a valid width.
// See Section 4.1.2.3 of the ICCCM for more details.
func (c *Client) ValidateWidth(width int) int {
	return c.validateSize(width, c.nhints.WidthInc, c.nhints.BaseWidth,
		c.nhints.MinWidth, c.nhints.MaxWidth)
}

// validateSize is does the math for ValidateWidth and ValidateHeight.
func (c *Client) validateSize(size, inc, base, min, max int) int {
	if size < min && c.nhints.Flags&icccm.SizeHintPMinSize > 0 {
		return min
	}
	if size < 1 {
		return 1
	}
	if size > max && c.nhints.Flags&icccm.SizeHintPMaxSize > 0 {
		return max
	}
	if inc > 1 && c.nhints.Flags&icccm.SizeHintPResizeInc > 0 {
		var whichb int
		if base > 0 {
			whichb = base
		} else {
			whichb = min
		}
		size = whichb +
			(int(round(float64(size-whichb)/float64(inc))) * inc)
	}

	return size
}

// clientFrames represents the group of all possible frames that the client
// can switch to at any point in time.
type clientFrames struct {
	client  *Client
	full    *frame.Full
	borders *frame.Borders
	slim    *frame.Slim
	nada    *frame.Nada
}

// newClientFrames constructs a clientFrames value, initializes all possible
// frames for this client, and sets up and activates the initial frame.
func (c *Client) newClientFrames() clientFrames {
	// When reparenting, an UnmapNotify is generated. We must ignore it!
	c.unmapIgnore++
	cf := createFrames(c)

	if c.shouldDecor() {
		c.frame = cf.full
	} else {
		c.frame = cf.nada
	}

	x, y, w, h := frame.ClientToFrame(c.frame,
		c.win.Geom.X(), c.win.Geom.Y(), c.win.Geom.Width(), c.win.Geom.Height())
	x, y = max(0, x), max(0, y)
	c.frame.MoveResize(true, x, y, w, h)
	c.frame.On()

	c.refreshExtents()

	return cf
}

// createFrames constructs each individual frame for a clientFrames value.
// At present, Wingo will die if there are any errors.
func createFrames(c *Client) clientFrames {
	var err error
	errHandle := func(err error) {
		if err != nil {
			logger.Error.Fatalln(err)
		}
	}
	cf := clientFrames{client: c}

	cf.nada, err = frame.NewNada(wm.X, nil, c)
	errHandle(err)

	cf.slim, err = frame.NewSlim(wm.X, wm.Theme.Slim.FrameTheme(),
		cf.nada.Parent(), c)
	errHandle(err)

	cf.borders, err = frame.NewBorders(wm.X, wm.Theme.Borders.FrameTheme(),
		cf.nada.Parent(), c)
	errHandle(err)

	cf.full, err = frame.NewFull(wm.X, wm.Theme.Full.FrameTheme(),
		cf.nada.Parent(), c)
	errHandle(err)

	return cf
}

// set will switch the current frame of the client to the frame provided.
// It is preferrable to use 'Frame[FrameType]' instead.
func (cf clientFrames) set(f frame.Frame) {
	current := cf.client.Frame()
	if current == f {
		return
	}
	cf.client.frame.Off()
	cf.client.frame = f
	cf.client.frame.On()
	frame.Reset(cf.client.frame)

	cf.client.refreshExtents()
}

// destroy will destroy all resources associated with any frames created for
// this client.
func (cf clientFrames) destroy() {
	cf.nada.Destroy()
	cf.slim.Destroy()
	cf.borders.Destroy()
	cf.full.Destroy()

	// Since a single parent window is shared between all frames, we only need
	// to pick a parent window from one of the frames, and destroy that.
	cf.full.Parent().Destroy()
}

func (cf clientFrames) maximize() {
	cf.full.Maximize()
	cf.borders.Maximize()
	cf.slim.Maximize()
	cf.nada.Maximize()

	cf.client.refreshExtents()
}

func (cf clientFrames) unmaximize() {
	cf.full.Unmaximize()
	cf.borders.Unmaximize()
	cf.slim.Unmaximize()
	cf.nada.Unmaximize()

	cf.client.refreshExtents()
}

// updateIcon updates any frames that use a client's icon.
func (cf clientFrames) updateIcon() {
	cf.full.UpdateIcon()
}

// updateName updates any frames that use a client's name.
func (cf clientFrames) updateName() {
	cf.full.UpdateTitle()
}
