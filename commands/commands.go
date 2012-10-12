package commands

import (
	"bytes"
	"os/exec"
	"strings"
	"time"
	"unicode"

	"github.com/BurntSushi/gribble"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xrect"

	"github.com/BurntSushi/wingo/focus"
	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/stack"
	"github.com/BurntSushi/wingo/wm"
	"github.com/BurntSushi/wingo/workspace"
	"github.com/BurntSushi/wingo/xclient"
)

// Env declares all available commands. Any command not in
// this list cannot be executed.
var Env = gribble.New([]gribble.Command{
	&AddWorkspace{},
	&Close{},
	&CycleClientNext{},
	&CycleClientPrev{},
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
	&MovePointerAbsolute{},
	&MovePointerRelative{},
	&Raise{},
	&RemoveWorkspace{},
	&Resize{},
	&Quit{},
	&Shell{},
	&TileStart{},
	&TileStop{},
	&Unmaximize{},
	&Workspace{},
	&WorkspaceSendClient{},
	&WorkspaceWithClient{},

	&Input{},
	&SelectClient{},
	&SelectWorkspace{},

	&GetWorkspace{},
	&GetWorkspaceNext{},
	&GetWorkspacePrev{},
})

var (
	SafeExec   = make(chan func() gribble.Value, 1)
	SafeReturn = make(chan gribble.Value, 0)
)

func syncRun(f func() gribble.Value) gribble.Value {
	SafeExec <- f
	return <-SafeReturn
}

type AddWorkspace struct {
	Name string `param:"1"`
}

func (cmd AddWorkspace) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		if err := wm.AddWorkspace(cmd.Name); err != nil {
			logger.Warning.Printf(
				"Could not add workspace '%s': %s", cmd.Name, err)
			return ""
		}
		return cmd.Name
	})
}

type Close struct {
	Client gribble.Any `param:"1" types:"int,string"`
}

func (cmd Close) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			c.Close()
		})
		return nil
	})
}

type CycleClientChoose struct{}

func (cmd CycleClientChoose) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		wm.Prompts.Cycle.Choose()
		return nil
	})
}

type CycleClientHide struct{}

func (cmd CycleClientHide) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		wm.Prompts.Cycle.Hide()
		return nil
	})
}

type CycleClientNext struct {
	OnlyActiveWorkspace string `param:"1"`
	OnlyVisible         string `param:"2"`
	ShowIconified       string `param:"3"`
}

func (cmd CycleClientNext) Run() gribble.Value {
	cmd.RunWithKeyStr("")
	return nil
}

func (cmd CycleClientNext) RunWithKeyStr(keyStr string) {
	syncRun(func() gribble.Value {
		wm.ShowCycleClient(keyStr,
			stringBool(cmd.OnlyActiveWorkspace),
			stringBool(cmd.OnlyVisible),
			stringBool(cmd.ShowIconified))
		wm.Prompts.Cycle.Next()
		return nil
	})
}

type CycleClientPrev struct {
	OnlyActiveWorkspace string `param:"1"`
	OnlyVisible         string `param:"2"`
	ShowIconified       string `param:"3"`
}

func (cmd CycleClientPrev) Run() gribble.Value {
	cmd.RunWithKeyStr("")
	return nil
}

func (cmd CycleClientPrev) RunWithKeyStr(keyStr string) {
	syncRun(func() gribble.Value {
		wm.ShowCycleClient(keyStr,
			stringBool(cmd.OnlyActiveWorkspace),
			stringBool(cmd.OnlyVisible),
			stringBool(cmd.ShowIconified))
		wm.Prompts.Cycle.Prev()
		return nil
	})
}

type Focus struct {
	Client gribble.Any `param:"1" types:"int,string"`
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
					wrk.Activate(false)
				}
			} else {
				focus.Focus(c)
				xevent.ReplayPointer(wm.X)
			}
		})
	})
}

type FocusRaise struct {
	Client gribble.Any `param:"1" types:"int,string"`
}

func (cmd FocusRaise) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		return withClient(cmd.Client, func(c *xclient.Client) {
			focus.Focus(c)
			stack.Raise(c)
			xevent.ReplayPointer(wm.X)
		})
	})
}

type FrameBorders struct {
	Client gribble.Any `param:"1" types:"int,string"`
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
}

func (cmd HeadFocus) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		wm.Heads.WithVisibleWorkspace(cmd.Head,
			func(wrk *workspace.Workspace) {
				wrk.Activate(false)
			})
		wm.FocusFallback()
		return nil
	})
}

type HeadFocusWithClient struct {
	Head   int         `param:"1"`
	Client gribble.Any `param:"2" types:"int,string"`
}

func (cmd HeadFocusWithClient) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			wm.Heads.WithVisibleWorkspace(cmd.Head,
				func(wrk *workspace.Workspace) {
					wrk.Activate(false)
					wrk.Add(c)
					stack.Raise(c)
				})
		})
		return nil
	})
}

type ToggleFloating struct {
	Client gribble.Any `param:"1" types:"int,string"`
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
}

func (cmd Maximize) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			c.Maximize()
		})
		return nil
	})
}

type MouseMove struct{}

func (cmd MouseMove) Run() gribble.Value { return nil }

type MouseResize struct {
	Direction string `param:"1"`
}

func (cmd MouseResize) Run() gribble.Value { return nil }

type Raise struct {
	Client gribble.Any `param:"1" types:"int,string"`
}

func (cmd Raise) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		return withClient(cmd.Client, func(c *xclient.Client) {
			stack.Raise(c)
			xevent.ReplayPointer(wm.X)
		})
	})
}

type Move struct {
	Client gribble.Any `param:"1" types:"int,string"`
	X      gribble.Any `param:"2" types:"int,float"`
	Y      gribble.Any `param:"3" types:"int,float"`
}

func (cmd Move) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		x, xok := parsePos(cmd.X, false)
		y, yok := parsePos(cmd.Y, true)
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

type MovePointerAbsolute struct {
	X int `param:"1"`
	Y int `param:"2"`
}

func (cmd MovePointerAbsolute) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		xproto.WarpPointer(wm.X.Conn(), 0, wm.X.RootWin(), 0, 0, 0, 0,
			int16(cmd.X), int16(cmd.Y))
		return nil
	})
}

type MovePointerRelative struct {
	X int `param:"1"`
	Y int `param:"2"`
}

func (cmd MovePointerRelative) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		geom := wm.Workspace().Geom()
		xproto.WarpPointer(wm.X.Conn(), 0, wm.X.RootWin(), 0, 0, 0, 0,
			int16(geom.X()+cmd.X), int16(geom.Y()+cmd.Y))
		return nil
	})
}

type RemoveWorkspace struct {
	Name gribble.Any `param:"1" types:"int,string"`
}

func (cmd RemoveWorkspace) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withWorkspace(cmd.Name, func(wrk *workspace.Workspace) {
			if err := wm.RemoveWorkspace(wrk); err != nil {
				logger.Warning.Printf("Could not remove workspace '%s': %s",
					wrk, err)
				return
			}
			wm.FocusFallback()
		})
		return nil
	})
}

type Resize struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Width  gribble.Any `param:"2" types:"int,float"`
	Height gribble.Any `param:"3" types:"int,float"`
}

func (cmd Resize) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		w, wok := parseDim(cmd.Width, false)
		h, hok := parseDim(cmd.Height, true)
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

type Quit struct{}

func (cmd Quit) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		logger.Message.Println("The User has told us to quit.")
		xevent.Quit(wm.X)
		return nil
	})
}

// Shell takes a command specified in a configuration file and
// tries to parse it as an executable command. The parser currently has
// only very basic support for quoted values and should be considered
// fragile. This should NOT be considered as a suitable replacement for
// something like `xbindkeys`.
type Shell struct {
	Command string `param:"1"`
}

func (cmd Shell) Run() gribble.Value {
	var stderr bytes.Buffer

	if len(cmd.Command) == 0 {
		return nil
	}

	splitCmdName := strings.SplitN(cmd.Command, " ", 2)
	cmdName := splitCmdName[0]
	args := make([]string, 0)
	addArg := func(start, end int) {
		args = append(args, strings.TrimSpace(splitCmdName[1][start:end]))
	}

	if len(splitCmdName) > 1 {
		startArgPos := 0
		inQuote := false
		for i, char := range splitCmdName[1] {
			// Add arguments enclosed in quotes
			// Yes, this mixes up quotes.
			if char == '"' || char == '\'' {
				inQuote = !inQuote

				if !inQuote {
					addArg(startArgPos, i)
				}
				startArgPos = i + 1 // skip the end quote character
			}

			// Add arguments separated by spaces without quotes
			if !inQuote && unicode.IsSpace(char) {
				addArg(startArgPos, i)
				startArgPos = i
			}
		}

		// add anything that's left over
		addArg(startArgPos, len(splitCmdName[1]))
	}

	// XXX: This is very weird.
	// If I don't put this into its own go-routine and wait a small
	// amount of time, commands that start new X clients fail miserably.
	// And when I say miserably, I mean they take down X itself.
	// For some reason, this avoids that problem. For now...
	// (I thought the problem was the grab imposed by a key binding,
	// but ungrabbing the keyboard before running this command didn't
	// change behavior.)
	go func() {
		time.Sleep(time.Microsecond)
		shellCmd := exec.Command(cmdName, args...)
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

type TileStart struct{}

func (cmd TileStart) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		wm.Workspace().LayoutStateSet(workspace.AutoTiling)
		return nil
	})
}

type TileStop struct{}

func (cmd TileStop) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		wm.Workspace().LayoutStateSet(workspace.Floating)
		return nil
	})
}

type Unmaximize struct {
	Client gribble.Any `param:"1" types:"int,string"`
}

func (cmd Unmaximize) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withClient(cmd.Client, func(c *xclient.Client) {
			c.Unmaximize()
		})
		return nil
	})
}

type Workspace struct {
	Name gribble.Any `param:"1" types:"int,string"`
}

func (cmd Workspace) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withWorkspace(cmd.Name, func(wrk *workspace.Workspace) {
			wrk.Activate(false)
			wm.FocusFallback()
		})
		return nil
	})
}

type WorkspaceSendClient struct {
	Name   gribble.Any `param:"1" types:"int,string"`
	Client gribble.Any `param:"2" types:"int,string"`
}

func (cmd WorkspaceSendClient) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withWorkspace(cmd.Name, func(wrk *workspace.Workspace) {
			withClient(cmd.Client, func(c *xclient.Client) {
				wrk.Add(c)
			})
		})
		return nil
	})
}

type WorkspaceWithClient struct {
	Name   gribble.Any `param:"1" types:"int,string"`
	Client gribble.Any `param:"2" types:"int,string"`
}

func (cmd WorkspaceWithClient) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withWorkspace(cmd.Name, func(wrk *workspace.Workspace) {
			withClient(cmd.Client, func(c *xclient.Client) {
				stack.Raise(c)
				wrk.Add(c)
				wrk.Activate(false)
				wm.FocusFallback()
			})
		})
		return nil
	})
}
