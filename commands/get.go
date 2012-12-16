package commands

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/gribble"

	"github.com/cshapeshifter/wingo/logger"
	"github.com/cshapeshifter/wingo/wm"
	"github.com/cshapeshifter/wingo/workspace"
	"github.com/cshapeshifter/wingo/xclient"
)

type GetActive struct {
	Help string `
Returns the id of the currently active window. If there is no active window,
0 is returned.
`
}

func (cmd GetActive) Run() gribble.Value {
	reply, err := xproto.GetInputFocus(wm.X.Conn()).Reply()
	if err != nil {
		logger.Warning.Printf("Could not get input focus: %s", err)
		return 0
	}

	// If our dummy window has focus, then it's equivalent to having root
	// window focus.
	// XXX: This may not be right if we're in a DE with desktop windows...
	if reply.Focus == wm.X.Dummy() {
		return 0
	}
	if focused := wm.LastFocused(); focused != nil {
		client := focused.(*xclient.Client)
		return int(client.Id())
	}
	return 0
}

type GetClientList struct {
	Workspace   gribble.Any `param:"1" types:"int,string"`
	Help string `
Returns a list of client ids separated by new lines on the workspace specified
by Workspace. Clients are listed in their focus orderering, from most recently
focused to least recently focused.

Workspace may be a workspace index (integer) starting at 0, or a workspace name.
`
}

func (cmd GetClientList) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		cids := make([]string, 0)
		withWorkspace(cmd.Workspace, func(wrk *workspace.Workspace) {
			for _, client := range wrk.Clients {
				cids = append(cids, fmt.Sprintf("%d", client.Id()))
			}
		})
		return strings.Join(cids, "\n")
	})
}

type GetClientName struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help string `
Returns the name of the window specified by Client active window.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd GetClientName) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		name := ""
		withClient(cmd.Client, func(c *xclient.Client) {
			name = c.Name()
		})
		return name
	})
}

type GetClientType struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help string `
Returns the type of the window specified by Client active window. A window
type will either be "desktop", "dock" or "normal".

Client may be the window id or a substring that matches a window name.
`
}

func (cmd GetClientType) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		typ := ""
		withClient(cmd.Client, func(c *xclient.Client) {
			typ = c.PrimaryTypeString()
		})
		return typ
	})
}

type GetClientWorkspace struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help string `
Returns the workspace of the window specified by Client active window.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd GetClientWorkspace) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		var wrk workspace.Workspacer = nil
		withClient(cmd.Client, func(c *xclient.Client) {
			wrk = c.Workspace()
		})
		if wrk == nil {
			return ""
		}
		return wrk.String()
	})
}

type GetHead struct {
	Help string `
Returns the index of the current head. Indexing starts at 0. Heads are ordered 
by their physical position: left to right and then top to bottom.
`
}

func (cmd GetHead) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		return wm.Heads.VisibleIndex(wm.Workspace())
	})
}

type GetHeadWorkspace struct {
	Head int `param:"1"`
	Help string `
Returns the name of the workspace currently visible on the monitor indexed by
Head. Indexing starts at 0. Heads are ordered by their physical position:
left to right and then top to bottom.
`
}

func (cmd GetHeadWorkspace) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		name := ""
		wm.Heads.WithVisibleWorkspace(cmd.Head,
			func(wrk *workspace.Workspace) {
				name = wrk.String()
			})
		return name
	})
}

type GetLayout struct {
	Workspace gribble.Any `param:"1" types:"int,string"`
	Help string `
Returns the name of the currently active (or "default") layout on the workspace
specified by Workspace. Note that when a workspace is set to a tiling layout,
it is still possible for clients to be floating.

Workspace may be a workspace index (integer) starting at 0, or a workspace name.
`
}

func (cmd GetLayout) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		var w workspace.Workspacer = nil
		withWorkspace(cmd.Workspace, func(wrk *workspace.Workspace) {
			w = wrk
		})
		if w == nil {
			return ""
		}
		return w.LayoutName()
	})
}


type GetWorkspace struct {
	Help string `
Returns the name of the current workspace.
`
}

func (cmd GetWorkspace) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		return wm.Workspace().Name
	})
}

type GetWorkspaceId struct {
	Workspace   gribble.Any `param:"1" types:"int,string"`
	Help string `
Returns the id (the index) of the workspace specified by Workspace.

Workspace may be a workspace index (integer) starting at 0, or a workspace name.
`
}

func (cmd GetWorkspaceId) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		ind := -1
		withWorkspace(cmd.Workspace, func(wrk *workspace.Workspace) {
			ind = wm.Heads.GlobalIndex(wrk)
		})
		return ind
	})
}

type GetWorkspaceList struct {
	Help string `
Returns a list of all workspaces, in the order that they were added.

The special "Sticky" workspace is not included.
`
}

func (cmd GetWorkspaceList) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		wrks := make([]string, len(wm.Heads.Workspaces.Wrks))
		for i, wrk := range wm.Heads.Workspaces.Wrks {
			wrks[i] = wrk.Name
		}
		return strings.Join(wrks, "\n")
	})
}

type GetWorkspaceNext struct {
	Help string `
Returns the name of the "next" workspace. The ordering of workspaces is
the order in which they were added. This might cause confusing behavior in
multi-head setups, since multiple workspaces can be viewable at one time.
`
}

func (cmd GetWorkspaceNext) Run() gribble.Value {
	return wm.Heads.NextWorkspace().Name
}

type GetWorkspacePrefix struct {
	Prefix string `param:"1"`
	Help   string `
Returns the first non-visible workspace starting with Prefix. If the current 
workspace starts with Prefix, then the first workspace *after* the current 
workspace starting with Prefix will be returned.
`
}

func (cmd GetWorkspacePrefix) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		hasPre := func(wrk *workspace.Workspace, prefix string) bool {
			return strings.HasPrefix(strings.ToLower(wrk.Name), prefix)
		}
		preAndHidden := func(wrk *workspace.Workspace, prefix string) bool {
			return !wrk.IsVisible() && hasPre(wrk, prefix)
		}

		needle := strings.ToLower(cmd.Prefix)
		cur := wm.Workspace()
		if hasPre(cur, needle) {
			past := false
			for _, wrk := range wm.Heads.Workspaces.Wrks {
				if past {
					if preAndHidden(wrk, needle) {
						return wrk.Name
					}
					continue
				}
				if wrk == cur {
					past = true
				}
			}

			// Nothing? Now look for one before 'cur'...
			for _, wrk := range wm.Heads.Workspaces.Wrks {
				if wrk == cur { // we've gone too far...
					return ""
				}
				if preAndHidden(wrk, needle) {
					return wrk.Name
				}
			}
		} else {
			for _, wrk := range wm.Heads.Workspaces.Wrks {
				if preAndHidden(wrk, needle) {
					return wrk.Name
				}
			}
		}
		return ""
	})
}

type GetWorkspacePrev struct {
	Help string `
Returns the name of the "previous" workspace. The ordering of workspaces is
the order in which they were added. This might cause confusing behavior in
multi-head setups, since multiple workspaces can be viewable at one time.
`
}

func (cmd GetWorkspacePrev) Run() gribble.Value {
	return wm.Heads.PrevWorkspace().Name
}
