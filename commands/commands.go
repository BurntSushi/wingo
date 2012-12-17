package commands

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/BurntSushi/gribble"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xrect"

	"github.com/cshapeshifter/wingo/focus"
	"github.com/cshapeshifter/wingo/logger"
	"github.com/cshapeshifter/wingo/misc"
	"github.com/cshapeshifter/wingo/wm"
	"github.com/cshapeshifter/wingo/workspace"
	"github.com/cshapeshifter/wingo/xclient"
)

// Env declares all available commands. Any command not in
// this list cannot be executed.
var Env = gribble.New([]gribble.Command{
	&AddWorkspace{},
	&Close{},
	&Dale{},
	&Float{},
	&Focus{},
	&FocusRaise{},
	&FrameBorders{},
	&FrameFull{},
	&FrameNada{},
	&FrameSlim{},
	&HeadFocus{},
	&HeadFocusWithClient{},
	&ToggleFloating{},
	&ToggleIconify{},
	&ToggleMaximize{},
	&ToggleStackAbove{},
	&ToggleStackBelow{},
	&ToggleSticky{},
	&Maximize{},
	&MouseMove{},
	&MouseResize{},
	&Move{},
	&MoveRelative{},
	&MovePointer{},
	&MovePointerRelative{},
	&Raise{},
	&RemoveWorkspace{},
	&RenameWorkspace{},
	&Resize{},
	&Restart{},
	&Quit{},
	&SetLayout{},
	&SetOpacity{},
	&Shell{},
	&Unfloat{},
	&Unmaximize{},
	&WingoExec{},
	&WingoHelp{},
	&Workspace{},
	&WorkspaceGreedy{},
	&WorkspaceSendClient{},
	&WorkspaceToHead{},
	&WorkspaceWithClient{},
	&WorkspaceGreedyWithClient{},

	&AutoTile{},
	&AutoUntile{},
	&AutoCycle{},
	&AutoResizeMaster{},
	&AutoResizeWindow{},
	&AutoNext{},
	&AutoPrev{},
	&AutoSwitchNext{},
	&AutoSwitchPrev{},
	&AutoMaster{},
	&AutoMakeMaster{},
	&AutoMastersMore{},
	&AutoMastersFewer{},

	&CycleClientChoose{},
	&CycleClientHide{},
	&CycleClientNext{},
	&CycleClientPrev{},
	&Input{},
	&Message{},
	&SelectClient{},
	&SelectWorkspace{},

	&GetActive{},
	&GetClientList{},
	&GetClientName{},
	&GetClientType{},
	&GetClientWorkspace{},
	&GetHead{},
	&GetHeadWorkspace{},
	&GetLayout{},
	&GetWorkspace{},
	&GetWorkspaceId{},
	&GetWorkspaceList{},
	&GetWorkspaceNext{},
	&GetWorkspacePrefix{},
	&GetWorkspacePrev{},

	&True{},
	&False{},
	&MatchClientClass{},
	&MatchClientInstance{},
	&MatchClientIsTransient{},
	&MatchClientName{},
	&MatchClientType{},
	&Not{},
	&And{},
	&Or{},
})

var (
	// SafeExec is a channel through which a Gribble command execution is
	// sent and executed synchronously with respect to the X main event loop.
	// This is necessary to allow asynchronous prompts to run and return
	// values without locking up the rest of the window manager.
	SafeExec   = make(chan func() gribble.Value, 1)

	// SafeReturn is the means through which a return value from a Gribble
	// command is synchronously returned with respext to the X main event loop.
	// See SafeExec.
	SafeReturn = make(chan gribble.Value, 0)
)

func init() {
	// This should be false in general for logging purposes.
	// When a command is executed via IPC, we temporarily turn it on so we
	// can give the user better error messages.
	Env.Verbose = false
}

// syncRun should wrap the execution of most Gribble commands to ensure
// synchronous execution with respect to the main X event loop.
func syncRun(f func() gribble.Value) gribble.Value {
	SafeExec <- f
	return <-SafeReturn
}

type AddWorkspace struct {
	Name string `param:"1"`
	Help string `
Adds a new workspace to Wingo with a name Name. Note that a workspace name
must be unique with respect to other workspaces and must have non-zero length.

The name of the workspace that was added is returned.
`
}

func (cmd AddWorkspace) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		if err := wm.AddWorkspace(cmd.Name); err != nil {
			wm.PopupError("Could not add workspace '%s': %s", cmd.Name, err)
			return ""
		}
		return cmd.Name
	})
}

type Close struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help string `
Closes the window specified by Client.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd Close) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			c.Close()
		})
		return nil
	})
}

type Dale struct {
	Help string `
Make sure "audio_play_cmd" is set to a program that can play wav files.
`
}

func (cmd Dale) Run() gribble.Value {
	go func() {
		var stderr bytes.Buffer

		program := wm.Config.AudioProgram

		c := exec.Command(program)
		c.Stderr = &stderr
		c.Stdin = bytes.NewReader(misc.WingoWav)
		if err := c.Run(); err != nil {
			if stderr.Len() > 0 {
				logger.Warning.Printf("%s failed: %s", program, stderr.String())
			}
			logger.Warning.Printf("Error running %s: %s", program, err)
		}
	}()
	return nil
}

type Float struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help string `
Floats the window specified by Client. If the window is already floating,
this command has no effect.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd Float) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			c.Float()
		})
		return nil
	})
}

type Focus struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help string `
Focuses the window specified by Client.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd Focus) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		return withClient(cmd.Client, func(c *xclient.Client) {
			if c == nil {
				focus.Root()

				// Use the mouse coordinates to find which workspace it was
				// clicked in. If a workspace can be found (i.e., no clicks in
				// dead areas), then activate it.
				xc, rw := wm.X.Conn(), wm.X.RootWin()
				qp, err := xproto.QueryPointer(xc, rw).Reply()
				if err != nil {
					logger.Warning.Printf("Could not query pointer: %s", err)
					return
				}

				geom := xrect.New(int(qp.RootX), int(qp.RootY), 1, 1)
				if wrk := wm.Heads.FindMostOverlap(geom); wrk != nil {
					wm.SetWorkspace(wrk, false)
				}
			} else {
				c.Focus()
				xevent.ReplayPointer(wm.X)
			}
		})
	})
}

type FocusRaise struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help string `
Focuses and raises the window specified by Client.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd FocusRaise) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		return withClient(cmd.Client, func(c *xclient.Client) {
			c.Focus()
			c.Raise()
			xevent.ReplayPointer(wm.X)
		})
	})
}

type FrameBorders struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help string `
Set the decorations of the window specified by Client to the "Borders" frame.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd FrameBorders) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			c.FrameBorders()
		})
		return nil
	})
}

type FrameFull struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help string `
Set the decorations of the window specified by Client to the "Full" frame.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd FrameFull) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			c.FrameFull()
		})
		return nil
	})
}

type FrameNada struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help string `
Set the decorations of the window specified by Client to the "Nada" frame.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd FrameNada) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			c.FrameNada()
		})
		return nil
	})
}

type FrameSlim struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help string `
Set the decorations of the window specified by Client to the "Slim" frame.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd FrameSlim) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			c.FrameSlim()
		})
		return nil
	})
}

type HeadFocus struct {
	Head int `param:"1"`
	Help string `
Focuses the head indexed at Head. Indexing starts at 0. Heads are ordered
by their physical position: left to right and then top to bottom.
`
}

func (cmd HeadFocus) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		wm.Heads.WithVisibleWorkspace(cmd.Head,
			func(wrk *workspace.Workspace) {
				wm.SetWorkspace(wrk, false)
			})
		wm.FocusFallback()
		return nil
	})
}

type HeadFocusWithClient struct {
	Head   int         `param:"1"`
	Client gribble.Any `param:"2" types:"int,string"`
	Help string `
Focuses the head indexed at Head, and move the Client specified by client to
that head. Indexing of heads starts at 0. Heads are ordered by their physical 
position: left to right and then top to bottom.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd HeadFocusWithClient) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			wm.Heads.WithVisibleWorkspace(cmd.Head,
				func(wrk *workspace.Workspace) {
					wm.SetWorkspace(wrk, false)
					wrk.Add(c)
					c.Raise()
				})
		})
		return nil
	})
}

type ToggleFloating struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help string `
Toggles whether the window specified by Client should be forced into the
floating layout. A window forced into the floating layout CANNOT be tiled.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd ToggleFloating) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			c.FloatingToggle()
		})
		return nil
	})
}

type ToggleIconify struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help string `
Iconifies (minimizes) or deiconifies (unminimizes) the window specified by 
Client.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd ToggleIconify) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			c.IconifyToggle()
		})
		return nil
	})
}

type ToggleMaximize struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help string `
Maximizes or restores the window specified by Client.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd ToggleMaximize) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			c.MaximizeToggle()
		})
		return nil
	})
}

type ToggleStackAbove struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help string `
Toggles the layer of the window specified by Client from normal to above. When
a window is in the "above" layer, it will always be above other (normal)
clients.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd ToggleStackAbove) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			c.StackAboveToggle()
		})
		return nil
	})
}

type ToggleStackBelow struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help string `
Toggles the layer of the window specified by Client from normal to below. When
a window is in the "below" layer, it will always be below other (normal)
clients.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd ToggleStackBelow) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			c.StackBelowToggle()
		})
		return nil
	})
}

type ToggleSticky struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help string `
Toggles the sticky status of the window specified by Client. When a window is
sticky, it will always be visible unless iconified. (i.e., it does not belong
to any particular workspace.)

Client may be the window id or a substring that matches a window name.
`
}

func (cmd ToggleSticky) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			c.StickyToggle()
		})
		return nil
	})
}

type Maximize struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help string `
Maximizes the window specified by Client. If the window is already maximized,
this command has no effect.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd Maximize) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			c.Maximize()
		})
		return nil
	})
}

type MouseMove struct{
	Help string `
Initiates a drag that allows a window to be moved with the mouse.

This is a special command that can only be assigned in Wingo's mouse
configuration file. Invoking this command in any other way has no effect.
`
}

func (cmd MouseMove) Run() gribble.Value {
	logger.Warning.Printf("The MouseMove command can only be invoked from "+
		"the Wingo mouse configuration file.")
	return nil
}

type MouseResize struct {
	Direction string `param:"1"`
	Help string `
Initiates a drag that allows a window to be resized with the mouse.

Direction specifies how the window should be resized, and what the pointer
should look like. For example, if Direction is set to "BottomRight", then only
the width and height of the window can change---but not the x or y position.

Valid values for Direction are: Infer, Top, Bottom, Left, Right, TopLeft,
TopRight, BottomLeft and BottomRight. When "Infer" is used, the direction
is determined based on where the pointer is on the window when the drag is
initiated.

This is a special command that can only be assigned in Wingo's mouse
configuration file. Invoking this command in any other way has no effect.
`
}

func (cmd MouseResize) Run() gribble.Value {
	logger.Warning.Printf("The MouseResize command can only be invoked from "+
		"the Wingo mouse configuration file.")
	return nil
}

type Raise struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help string `
Raises the window specified by Client to the top of its layer.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd Raise) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		return withClient(cmd.Client, func(c *xclient.Client) {
			c.Raise()
			xevent.ReplayPointer(wm.X)
		})
	})
}

type Move struct {
	Client gribble.Any `param:"1" types:"int,string"`
	X      gribble.Any `param:"2" types:"int,float"`
	Y      gribble.Any `param:"3" types:"int,float"`
	Help string `
Moves the window specified by Client to the x and y position specified by
X and Y. Note that the origin is located in the top left corner.

X and Y may either be pixels (integers) or ratios in the range 0.0 to
1.0 (specifically, (0.0, 1.0]). Ratios are measured with respect to the
window's workspace's geometry.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd Move) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		x, xok := parsePos(wm.Workspace().Geom(), cmd.X, false)
		y, yok := parsePos(wm.Workspace().Geom(), cmd.Y, true)
		if !xok || !yok {
			return nil
		}
		withClient(cmd.Client, func(c *xclient.Client) {
			c.EnsureUnmax()
			c.LayoutMove(x, y)
		})
		return nil
	})
}

type MoveRelative struct {
	Client gribble.Any `param:"1" types:"int,string"`
	X      gribble.Any `param:"2" types:"int,float"`
	Y      gribble.Any `param:"3" types:"int,float"`
	Help string `
Moves the window specified by Client to the x and y position specified by
X and Y, relative to its workspace. Note that the origin is located in the top
left corner of the client's workspace.

X and Y may either be pixels (integers) or ratios in the range 0.0 to
1.0 (specifically, (0.0, 1.0]). Ratios are measured with respect to the
window's workspace's geometry.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd MoveRelative) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		geom := wm.Workspace().Geom()
		x, xok := parsePos(geom, cmd.X, false)
		y, yok := parsePos(geom, cmd.Y, true)
		if !xok || !yok {
			return nil
		}
		withClient(cmd.Client, func(c *xclient.Client) {
			c.EnsureUnmax()
			c.LayoutMove(geom.X() + x, geom.Y() + y)
		})
		return nil
	})
}

type MovePointer struct {
	X int `param:"1"`
	Y int `param:"2"`
	Help string `
Moves the pointer to the x and y position specified by X and Y. Note the the
origin is located in the top left corner.
`
}

func (cmd MovePointer) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		xproto.WarpPointer(wm.X.Conn(), 0, wm.X.RootWin(), 0, 0, 0, 0,
			int16(cmd.X), int16(cmd.Y))
		return nil
	})
}

type MovePointerRelative struct {
	X gribble.Any `param:"1" types:"int,float"`
	Y gribble.Any `param:"2" types:"int,float"`
	Help string `
Moves the pointer to the x and y position specified by X and Y relative to the
current workspace. Note the the origin is located in the top left corner of
the current workspace.

X and Y may either be pixels (integers) or ratios in the range 0.0 to
1.0 (specifically, (0.0, 1.0]). Ratios are measured with respect to the
workspace's geometry.
`
}

func (cmd MovePointerRelative) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		geom := wm.Workspace().Geom()
		x, xok := parsePos(geom, cmd.X, false)
		y, yok := parsePos(geom, cmd.Y, true)
		if !xok || !yok {
			return nil
		}
		xproto.WarpPointer(wm.X.Conn(), 0, wm.X.RootWin(), 0, 0, 0, 0,
			int16(geom.X()+x), int16(geom.Y()+y))
		return nil
	})
}

type Restart struct {
	Help string `
Restarts Wingo in place using exec. This should be used to reload Wingo
after you've made changes to its configuration.
`
}

func (cmd Restart) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		wm.Restart = true // who says globals are bad?
		xevent.Quit(wm.X)
		return nil
	})
}

type Quit struct {
	Help string `
Stops Wingo.
`
}

func (cmd Quit) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		logger.Message.Println("The User has told us to quit.")
		xevent.Quit(wm.X)
		return nil
	})
}

type SetLayout struct {
	Workspace gribble.Any `param:"1" types:"int,string"`
	Name string `param:"2"`
	Help string `
Sets the current layout of the workspace specified by Workspace to the layout
named by Name. If a layout with name Name does not exist, this command has
no effect.

Note that this command has no effect if the workspace is not visible.

Workspace may be a workspace index (integer) starting at 0, or a workspace name.
`
}

func (cmd SetLayout) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withWorkspace(cmd.Workspace, func(wrk *workspace.Workspace) {
			wrk.SetLayout(cmd.Name)
		})
		return nil
	})
}


type SetOpacity struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Opacity float64 `param:"2"`
	Help string `
Sets the opacity of the window specified by Client to the opacity level
specified by Opacity.

This command won't have any effect unless you're running a compositing manager
like compton or cairo-compmgr.

Client may be the window id or a substring that matches a window name.

Opacity should be a float in the range 0.0 to 1.0, inclusive, where 0.0 is
completely transparent and 1.0 is completely opaque.
`
}

func (cmd SetOpacity) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		if cmd.Opacity < 0.0 || cmd.Opacity > 1.0 {
			logger.Warning.Printf(
				"Opacity %f is not in the range [0, 1].", cmd.Opacity)
			return nil
		}
		withClient(cmd.Client, func(c *xclient.Client) {
			// Opacity is set on the top-most frame window of the client.
			ewmh.WmWindowOpacitySet(wm.X, c.Frame().Parent().Id, cmd.Opacity)
		})
		return nil
	})
}

type RemoveWorkspace struct {
	Workspace gribble.Any `param:"1" types:"int,string"`
	Help string `
Removes the workspace specified by Workspace. Note that a workspace can *only*
be removed if it is empty (i.e., does not contain any windows).

Workspace may be a workspace index (integer) starting at 0, or a workspace name.
`
}

func (cmd RemoveWorkspace) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withWorkspace(cmd.Workspace, func(wrk *workspace.Workspace) {
			if err := wm.RemoveWorkspace(wrk); err != nil {
				wm.PopupError("Could not remove workspace '%s': %s", wrk, err)
				return
			}

			wm.FYI("Workspace %s removed.", wrk)
			wm.FocusFallback()
		})
		return nil
	})
}

type RenameWorkspace struct {
	Workspace gribble.Any `param:"1" types:"int,string"`
	NewName string `param:"2"`
	Help string `
Renames the workspace specified by Workspace to the name in NewName.

Workspace may be a workspace index (integer) starting at 0, or a workspace name.
NewName can only be a string.
`
}

func (cmd RenameWorkspace) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withWorkspace(cmd.Workspace, func(wrk *workspace.Workspace) {
			oldName := wrk.String()
			if err := wm.RenameWorkspace(wrk, cmd.NewName); err != nil {
				wm.PopupError("Could not rename workspace '%s': %s", wrk, err)
				return
			}

			wm.FYI("Workspace %s renamed to %s.", oldName, cmd.NewName)
		})
		return nil
	})
}

type Resize struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Width  gribble.Any `param:"2" types:"int,float"`
	Height gribble.Any `param:"3" types:"int,float"`
	Help string `
Resizes the window specified by Client to some width and height specified by 
Width and Height.

Width and Height may either be pixels (integers) or ratios in the range 0.0 to
1.0 (specifically, (0.0, 1.0]). Ratios are measured with respect to the
window's workspace's geometry.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd Resize) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		w, wok := parseDim(wm.Workspace().Geom(), cmd.Width, false)
		h, hok := parseDim(wm.Workspace().Geom(), cmd.Height, true)
		if !wok || !hok {
			return nil
		}
		withClient(cmd.Client, func(c *xclient.Client) {
			c.EnsureUnmax()
			c.LayoutResize(w, h)
		})
		return nil
	})
}

type Shell struct {
	Command string `param:"1"`
	Help string `
Attempts to execute the shell command specified by Command. If an error occurs,
it will be logged to Wingo's stderr.

Note that the parser for translating shell commands to something acceptable
for Go's os/exec package is fairly primitive. Therefore, this should not be
considered as a suitable replacement for similar utilities (like gmrun or
xbindkeys).
`
}

func (cmd Shell) Run() gribble.Value {
	// XXX: This is very weird.
	// If I don't put this into its own go-routine and wait a small
	// amount of time, commands that start new X clients fail miserably.
	// And when I say miserably, I mean they take down X itself.
	// For some reason, this avoids that problem. For now...
	// (I thought the problem was the grab imposed by a key binding,
	// but ungrabbing the keyboard before running this command didn't
	// change behavior.)
	go func() {
		var stderr bytes.Buffer

		// For some reason, Go's text/scanner doesn't unescape escaped quotes
		// in strings. So we try to be nice and do it here.
		cmd.Command = strings.Replace(cmd.Command, "\\\"", "\"", -1)

		// BUG(burntsushi): I think there is a bug in text/scanner where if
		// a string ends with an escaped quote, the quote is cutoff and the
		// backslash is left intact.
		if cmd.Command[len(cmd.Command)-1] == '\\' {
			cmd.Command = fmt.Sprintf("%s\"", cmd.Command[0:len(cmd.Command)-1])
		}

		time.Sleep(time.Microsecond)
		logger.Message.Printf("bash -c [%s]", cmd.Command)
		shellCmd := exec.Command("bash", "-c", cmd.Command)
		shellCmd.Stderr = &stderr

		err := shellCmd.Run()
		if err != nil {
			logger.Warning.Printf("Error running '%s': %s", cmd.Command, err)
			if stderr.Len() > 0 {
				logger.Warning.Printf("Error running '%s': %s",
					cmd.Command, stderr.String())
			}
		}
	}()
	return nil
}

type Unfloat struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help string `
Unfloats the window specified by Client. If the window is not floating,
this command has no effect.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd Unfloat) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			c.Unfloat()
		})
		return nil
	})
}

type Unmaximize struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help string `
Unmaximizes the window specified by Client. If the window is not maximized,
this command has no effect.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd Unmaximize) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			c.Unmaximize()
		})
		return nil
	})
}

type WingoExec struct {
	Commands string `param:"1"`
	Help string `
Executes a series of Wingo commands specified by Commands. If an error occurs
while executing the command, it will be shown in a popup message.
`
}

func (cmd WingoExec) Run() gribble.Value {
	Env.Verbose = true
	_, err := Env.RunMany(cmd.Commands)
	Env.Verbose = false
	if len(cmd.Commands) > 0 && err != nil {
		wm.PopupError("%s", err)
	}
	return nil
}

type WingoHelp struct {
	CommandName string `param:"1"`
	Help string `
Shows the usage information for a particular command specified by CommandName.
`
}

func (cmd WingoHelp) Run() gribble.Value {
	if len(strings.TrimSpace(cmd.CommandName)) == 0 {
		return nil
	}
	usage := Env.UsageTypes(cmd.CommandName)
	help := Env.Help(cmd.CommandName)
	wm.PopupError("%s\n%s\n%s", usage, strings.Repeat("-", len(usage)), help)
	return nil
}

type Workspace struct {
	Workspace gribble.Any `param:"1" types:"int,string"`
	Help string `
Sets the current workspace to the one specified by Workspace.

Workspace may be a workspace index (integer) starting at 0, or a workspace name.
`
}

func (cmd Workspace) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withWorkspace(cmd.Workspace, func(wrk *workspace.Workspace) {
			wm.SetWorkspace(wrk, false)
			wm.FocusFallback()
		})
		return nil
	})
}

type WorkspaceGreedy struct {
	Workspace gribble.Any `param:"1" types:"int,string"`
	Help string `
Sets the current workspace to the one specified by Workspace in a greedy
fashion.

A greedy switch *always* brings the specified workspace to the
currently focused head. (N.B. Greedy is only different when switching between
two visible workspaces.)

Workspace may be a workspace index (integer) starting at 0, or a workspace name.
`
}

func (cmd WorkspaceGreedy) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withWorkspace(cmd.Workspace, func(wrk *workspace.Workspace) {
			wm.SetWorkspace(wrk, true)
			wm.FocusFallback()
		})
		return nil
	})
}

type WorkspaceSendClient struct {
	Workspace   gribble.Any `param:"1" types:"int,string"`
	Client gribble.Any `param:"2" types:"int,string"`
	Help string `
Sends the window specified by Client to the workspace specified by Workspace.

Workspace may be a workspace index (integer) starting at 0, or a workspace name.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd WorkspaceSendClient) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withWorkspace(cmd.Workspace, func(wrk *workspace.Workspace) {
			withClient(cmd.Client, func(c *xclient.Client) {
				wrk.Add(c)
			})
		})
		return nil
	})
}

type WorkspaceToHead struct {
	Head int `param:"1"`
	Workspace gribble.Any `param:"2" types:"int,string"`
	Help string `
Sets the workspace specified by Workspace to appear on the head specified by
the Head index.

Workspace may be a workspace index (integer) starting at 0, or a workspace name.

Head indexing starts at 0. Heads are ordered by their physical position: left 
to right and then top to bottom.
`
}

func (cmd WorkspaceToHead) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withWorkspace(cmd.Workspace, func(wrk *workspace.Workspace) {
			wm.WorkspaceToHead(cmd.Head, wrk)
			wm.FocusFallback()
		})
		return nil
	})
}

type WorkspaceWithClient struct {
	Workspace   gribble.Any `param:"1" types:"int,string"`
	Client gribble.Any `param:"2" types:"int,string"`
	Help string `
Sets the current workspace to the workspace specified by Workspace, and moves
the window specified by Client to that workspace.

Workspace may be a workspace index (integer) starting at 0, or a workspace name.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd WorkspaceWithClient) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withWorkspace(cmd.Workspace, func(wrk *workspace.Workspace) {
			withClient(cmd.Client, func(c *xclient.Client) {
				c.Raise()
				wrk.Add(c)
				wm.SetWorkspace(wrk, false)
				wm.FocusFallback()
			})
		})
		return nil
	})
}

type WorkspaceGreedyWithClient struct {
	Workspace   gribble.Any `param:"1" types:"int,string"`
	Client gribble.Any `param:"2" types:"int,string"`
	Help string `
Sets the current workspace to the workspace specified by Workspace in a greedy
fashion, and moves the window specified by Client to that workspace.

A greedy switch *always* brings the specified workspace to the
currently focused head. (N.B. Greedy is only different when switching between
two visible workspaces.)

Workspace may be a workspace index (integer) starting at 0, or a workspace name.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd WorkspaceGreedyWithClient) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withWorkspace(cmd.Workspace, func(wrk *workspace.Workspace) {
			withClient(cmd.Client, func(c *xclient.Client) {
				c.Raise()
				wrk.Add(c)
				wm.SetWorkspace(wrk, true)
				wm.FocusFallback()
			})
		})
		return nil
	})
}
