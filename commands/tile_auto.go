package commands

import (
	"github.com/BurntSushi/gribble"

	"github.com/BurntSushi/wingo-conc/workspace"
)

type AutoTile struct {
	Workspace gribble.Any `param:"1" types:"int,string"`
	Help      string      `
Initiates automatic tiling on the workspace specified by Workspace. If tiling
is already active, the layout will be re-placed.

Note that this command has no effect if the workspace is not visible.

Workspace may be a workspace index (integer) starting at 0, or a workspace
name.
`
}

func (cmd AutoTile) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withWorkspace(cmd.Workspace, func(wrk *workspace.Workspace) {
			wrk.LayoutStateSet(workspace.AutoTiling)
		})
		return nil
	})
}

type AutoUntile struct {
	Workspace gribble.Any `param:"1" types:"int,string"`
	Help      string      `
Stops automatic tiling on the workspace specified by Workspace, and restores
windows to their position and geometry before being tiled. If tiling is not
active on the specified workspace, this command has no effect.

Note that this command has no effect if the workspace is not visible.

Workspace may be a workspace index (integer) starting at 0, or a workspace
name.
`
}

func (cmd AutoUntile) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withWorkspace(cmd.Workspace, func(wrk *workspace.Workspace) {
			wrk.LayoutStateSet(workspace.Floating)
		})
		return nil
	})
}

type AutoCycle struct {
	Workspace gribble.Any `param:"1" types:"int,string"`
	Help      string      `
Cycles to the next automatic tiling layout in the workspace specified by
Workspace.

Note that this command has no effect if the workspace is not visible.

Workspace may be a workspace index (integer) starting at 0, or a workspace
name.
`
}

func (cmd AutoCycle) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withWorkspace(cmd.Workspace, func(wrk *workspace.Workspace) {
			wrk.AutoCycle()
		})
		return nil
	})
}

type AutoResizeMaster struct {
	Workspace gribble.Any `param:"1" types:"int,string"`
	Amount    float64     `param:"2"`
	Help      string      `
Increases or decreases the size of the master split by Amount in the layout on
the workspace specified by Workspace.

Amount should be a ratio between 0.0 and 1.0.

Workspace may be a workspace index (integer) starting at 0, or a workspace
name.
`
}

func (cmd AutoResizeMaster) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withWorkspace(cmd.Workspace, func(wrk *workspace.Workspace) {
			if wrk.State != workspace.AutoTiling {
				return
			}
			wrk.LayoutAutoTiler().ResizeMaster(cmd.Amount)
		})
		return nil
	})
}

type AutoResizeWindow struct {
	Workspace gribble.Any `param:"1" types:"int,string"`
	Amount    float64     `param:"2"`
	Help      string      `
Increases or decreases the size of the current window by Amount in the layout
on the workspace specified by Workspace.

Amount should be a ratio between 0.0 and 1.0.

Workspace may be a workspace index (integer) starting at 0, or a workspace
name.
`
}

func (cmd AutoResizeWindow) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withWorkspace(cmd.Workspace, func(wrk *workspace.Workspace) {
			if wrk.State != workspace.AutoTiling {
				return
			}
			wrk.LayoutAutoTiler().ResizeWindow(cmd.Amount)
		})
		return nil
	})
}

type AutoNext struct {
	Workspace gribble.Any `param:"1" types:"int,string"`
	Help      string      `
Moves focus to the next client in the layout.

Note that this command has no effect if the workspace is not visible.

Workspace may be a workspace index (integer) starting at 0, or a workspace
name.
`
}

func (cmd AutoNext) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withWorkspace(cmd.Workspace, func(wrk *workspace.Workspace) {
			if wrk.State != workspace.AutoTiling {
				return
			}
			wrk.LayoutAutoTiler().Next()
		})
		return nil
	})
}

type AutoPrev struct {
	Workspace gribble.Any `param:"1" types:"int,string"`
	Help      string      `
Moves focus to the next client in the layout.

Note that this command has no effect if the workspace is not visible.

Workspace may be a workspace index (integer) starting at 0, or a workspace
name.
`
}

func (cmd AutoPrev) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withWorkspace(cmd.Workspace, func(wrk *workspace.Workspace) {
			if wrk.State != workspace.AutoTiling {
				return
			}
			wrk.LayoutAutoTiler().Prev()
		})
		return nil
	})
}

type AutoSwitchNext struct {
	Workspace gribble.Any `param:"1" types:"int,string"`
	Help      string      `
Switches the current window with the next window in the layout.

Note that this command has no effect if the workspace is not visible.

Workspace may be a workspace index (integer) starting at 0, or a workspace
name.
`
}

func (cmd AutoSwitchNext) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withWorkspace(cmd.Workspace, func(wrk *workspace.Workspace) {
			if wrk.State != workspace.AutoTiling {
				return
			}
			wrk.LayoutAutoTiler().SwitchNext()
		})
		return nil
	})
}

type AutoSwitchPrev struct {
	Workspace gribble.Any `param:"1" types:"int,string"`
	Help      string      `
Switches the current window with the previous window in the layout.

Note that this command has no effect if the workspace is not visible.

Workspace may be a workspace index (integer) starting at 0, or a workspace
name.
`
}

func (cmd AutoSwitchPrev) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withWorkspace(cmd.Workspace, func(wrk *workspace.Workspace) {
			if wrk.State != workspace.AutoTiling {
				return
			}
			wrk.LayoutAutoTiler().SwitchPrev()
		})
		return nil
	})
}

type AutoMaster struct {
	Workspace gribble.Any `param:"1" types:"int,string"`
	Help      string      `
Focuses the (first) master window in the layout for the workspace specified
by Workspace.

Note that this command has no effect if the workspace is not visible.

Workspace may be a workspace index (integer) starting at 0, or a workspace
name.
`
}

func (cmd AutoMaster) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withWorkspace(cmd.Workspace, func(wrk *workspace.Workspace) {
			if wrk.State != workspace.AutoTiling {
				return
			}
			wrk.LayoutAutoTiler().FocusMaster()
		})
		return nil
	})
}

type AutoMakeMaster struct {
	Workspace gribble.Any `param:"1" types:"int,string"`
	Help      string      `
Switches the current window with the first master in the layout for the
workspace specified by Workspace.

Note that this command has no effect if the workspace is not visible.

Workspace may be a workspace index (integer) starting at 0, or a workspace
name.
`
}

func (cmd AutoMakeMaster) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withWorkspace(cmd.Workspace, func(wrk *workspace.Workspace) {
			if wrk.State != workspace.AutoTiling {
				return
			}
			wrk.LayoutAutoTiler().MakeMaster()
		})
		return nil
	})
}

type AutoMastersFewer struct {
	Workspace gribble.Any `param:"1" types:"int,string"`
	Help      string      `
Allows one fewer master window to fit into the master split.

Workspace may be a workspace index (integer) starting at 0, or a workspace
name.
`
}

func (cmd AutoMastersFewer) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withWorkspace(cmd.Workspace, func(wrk *workspace.Workspace) {
			if wrk.State != workspace.AutoTiling {
				return
			}
			wrk.LayoutAutoTiler().MastersFewer()
		})
		return nil
	})
}

type AutoMastersMore struct {
	Workspace gribble.Any `param:"1" types:"int,string"`
	Help      string      `
Allows one more master window to fit into the master split.

Workspace may be a workspace index (integer) starting at 0, or a workspace
name.
`
}

func (cmd AutoMastersMore) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		withWorkspace(cmd.Workspace, func(wrk *workspace.Workspace) {
			if wrk.State != workspace.AutoTiling {
				return
			}
			wrk.LayoutAutoTiler().MastersMore()
		})
		return nil
	})
}
