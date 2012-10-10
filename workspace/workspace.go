package workspace

import (
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xrect"

	"github.com/BurntSushi/wingo/layout"
	"github.com/BurntSushi/wingo/prompt"
)

type Workspace struct {
	X   *xgbutil.XUtil
	all *Workspaces

	Name    string
	State   int
	Clients []Client

	floaters   []layout.Floater
	curFloater int

	autoTilers   []layout.AutoTiler
	curAutoTiler int

	PromptSlctGroup *prompt.SelectGroupItem
	PromptSlctItem  *prompt.SelectItem
}

func (wrks *Workspaces) NewWorkspace(name string) *Workspace {
	wrk := &Workspace{
		X:       wrks.X,
		all:     wrks,
		Name:    name,
		State:   Floating,
		Clients: make([]Client, 0, 40),

		curFloater:   0,
		curAutoTiler: 0,
	}

	// Layouts must be listed in the order in which their corresponding
	// constants are defined in the layout package.
	wrk.floaters = []layout.Floater{
		layout.NewFloating(),
	}
	wrk.autoTilers = []layout.AutoTiler{
		layout.NewVertical(),
	}

	return wrk
}

func (wrk *Workspace) String() string {
	return wrk.Name
}

func (wrk *Workspace) Geom() xrect.Rect {
	if !wrk.IsVisible() {
		panic("Cannot get geometry of a hidden workspace.")
	}
	return wrk.all.heads.Geom(wrk)
}

func (wrk *Workspace) HeadGeom() xrect.Rect {
	if !wrk.IsVisible() {
		panic("Cannot get head geometry of a hidden workspace.")
	}
	return wrk.all.heads.HeadGeom(wrk)
}

func (wrk *Workspace) IsActive() bool {
	return wrk.all.heads.IsActive(wrk)
}

func (wrk *Workspace) IsVisible() bool {
	return wrk.all.heads.Geom(wrk) != nil
}

func (wrk *Workspace) Activate(greedy bool) {
	if wrk.IsActive() {
		return
	}

	active := wrk.all.Active()
	if !wrk.IsVisible() || greedy {
		wrk.all.heads.SwitchWorkspaces(wrk, active)
	}
	wrk.all.heads.ActivateWorkspace(wrk)
}

func (wrk *Workspace) Add(c Client) {
	if c.IsSticky() {
		return
	}

	current := c.Workspace()
	if current == wrk {
		return
	}

	// When a client transitions from a workspace that is tiling to a workspace
	// that is floating, its last floating state needs to be refreshed.
	if current != nil {
		if _, ok := c.Layout().(layout.Floater); ok {
			c.SaveState("last-floating")
		}
	}

	c.WorkspaceSet(wrk)
	wrk.Clients = append(wrk.Clients, c)

	if current != nil {
		current.Remove(c)
	}

	if !c.Iconified() {
		wrk.addToFloaters(c)
		wrk.addToTilers(c)
		wrk.Place()
	}
	if _, ok := c.Layout().(layout.Floater); ok && wrk.IsVisible() {
		c.LoadState("last-floating")
	}

	// If the old and new workspace have never visibilities, adjust the
	// client appropriately.
	if current != nil {
		if current.IsVisible() && !wrk.IsVisible() {
			c.SaveState("workspace-switch")
			c.Unmap()
		} else if !current.IsVisible() && wrk.IsVisible() {
			c.LoadState("workspace-switch")
			c.Map()
		}
	}
}

func (wrk *Workspace) Remove(c Client) {
	for i, c2 := range wrk.Clients {
		if c2.Id() == c.Id() {
			wrk.Clients = append(wrk.Clients[:i], wrk.Clients[i+1:]...)
			break
		}
	}
	wrk.removeFromFloaters(c)
	wrk.removeFromTilers(c)
	wrk.Place()
}

func (wrk *Workspace) RemoveAllAndAdd(newWk *Workspace) {
	mapOrUnmap := func(c Client) {
		if newWk.IsVisible() && !wrk.IsVisible() {
			c.Map()
		} else if !newWk.IsVisible() && wrk.IsVisible() {
			c.Unmap()
		}
	}
	for _, c := range wrk.Clients {
		if c.Workspace() != wrk {
			continue
		}

		c.WorkspaceSet(newWk)
		wrk.removeFromFloaters(c)
		wrk.removeFromTilers(c)
		if !c.Iconified() {
			wrk.addToFloaters(c)
			wrk.addToTilers(c)
			mapOrUnmap(c)
		}
	}
	newWk.Place()
}

func (wrk *Workspace) Show() {
	wrk.Place()
	for _, c := range wrk.Clients {
		if c.Iconified() {
			continue
		}
		if c.Workspace() == wrk {
			if _, ok := wrk.Layout(c).(layout.Floater); ok {
				c.LoadState("workspace-switch")
			}
			c.Map()
		}
	}
}

func (wrk *Workspace) Hide() {
	for _, c := range wrk.Clients {
		if c.Workspace() == wrk {
			c.SaveState("workspace-switch")
			c.Unmap()
		}
	}
}

func (wrk *Workspace) Place() {
	if !wrk.IsVisible() {
		return
	}

	// Floater layouts always get placed.
	wrk.LayoutFloater().Place(wrk.Geom())

	// Tiling layouts are only "placed" when the workspace is in the
	// appropriate layout mode.
	switch wrk.State {
	case Floating:
		// Nada nada limonada
	case AutoTiling:
		wrk.LayoutAutoTiler().Place(wrk.Geom())
	default:
		panic("Layout mode not implemented.")
	}
}

func (wrk *Workspace) IconifyToggle(c Client) {
	// If it's not the current workspace, a window cannot toggle iconification.
	if wrk != wrk.all.Active() {
		return
	}
	if c.Iconified() {
		if _, ok := wrk.Layout(c).(layout.Floater); ok {
			c.LoadState("before-iconify")
		} else {
			c.DeleteState("before-iconify")
		}
		wrk.addToFloaters(c)
		wrk.addToTilers(c)
		c.IconifiedSet(false)

		wrk.Place()
		c.Map()
	} else {
		c.SaveState("before-iconify")
		wrk.removeFromFloaters(c)
		wrk.removeFromTilers(c)
		c.IconifiedSet(true)

		c.Unmap()
		wrk.Place()
	}
}

// CheckFloatingStatus queries the Floating method of a client, and if it's
// different than what the workspace believes it should be, the proper state
// transition will be invoked.
// Namely, if Client.Floating is true, but workspace thinks it's false, the
// workspace will remove it from its list of tilable clients and re-tile.
// Otherwise, if Client.Floating is false, but workspace thinks it's true,
// the workspace will add the client to its list of tilable clients and re-tile.
func (wrk *Workspace) CheckFloatingStatus(c Client) {
	// If it's in one tiler, it's in them all.
	tilable := wrk.LayoutAutoTiler().Exists(c)
	if tilable && c.ShouldForceFloating() {
		wrk.removeFromTilers(c)
		if wrk.State != Floating {
			c.LoadState("last-floating")
		}
		wrk.Place()
	} else if !tilable && !c.ShouldForceFloating() {
		wrk.addToTilers(c)
		wrk.Place()
	}
}

func (wrk *Workspace) LayoutFloater() layout.Floater {
	return wrk.floaters[wrk.curFloater]
}

func (wrk *Workspace) LayoutAutoTiler() layout.AutoTiler {
	return wrk.autoTilers[wrk.curAutoTiler]
}

func (wrk *Workspace) addToFloaters(c Client) {
	for _, floater := range wrk.floaters {
		floater.Add(c)
	}
}

func (wrk *Workspace) removeFromFloaters(c Client) {
	for _, floater := range wrk.floaters {
		floater.Remove(c)
	}
}

func (wrk *Workspace) addToTilers(c Client) {
	if c.ShouldForceFloating() {
		return
	}
	for _, autoTiler := range wrk.autoTilers {
		autoTiler.Add(c)
	}
}

func (wrk *Workspace) removeFromTilers(c Client) {
	for _, autoTiler := range wrk.autoTilers {
		autoTiler.Remove(c)
	}
}

func (wrk *Workspace) Layout(c Client) layout.Layout {
	switch {
	case wrk.State == Floating || c.ShouldForceFloating():
		return wrk.LayoutFloater()
	case wrk.State == AutoTiling:
		return wrk.LayoutAutoTiler()
	default:
		panic("Layout state not implemented.")
	}
	panic("unreachable")
}

func (wrk *Workspace) LayoutStateSet(state int) {
	if state == wrk.State {
		// If it's an AutoTiler, then just call Place again.
		if wrk.State == AutoTiling {
			wrk.LayoutAutoTiler().Place(wrk.Geom())
		}
		return
	}

	// First undo the current layout.
	switch wrk.State {
	case Floating:
		wrk.LayoutFloater().Save()
		wrk.LayoutFloater().Unplace(wrk.Geom())
	case AutoTiling:
		wrk.LayoutAutoTiler().Unplace(wrk.Geom())
	default:
		panic("Layout state not implemented.")
	}

	// Now apply the new layout.
	switch state {
	case Floating:
		wrk.State = state
		wrk.LayoutFloater().Place(wrk.Geom())
		wrk.LayoutFloater().Reposition(wrk.Geom())
	case AutoTiling:
		wrk.State = state
		wrk.LayoutAutoTiler().Place(wrk.Geom())
	default:
		panic("Layout state not implemented.")
	}
}

func (wrk *Workspace) SelectGroupText() string {
	return wrk.String()
}

func (wrk *Workspace) SelectText() string {
	return wrk.String()
}

type SelectData struct {
	Selected    func(wrk *Workspace)
	Highlighted func(wrk *Workspace)
}

func (wrk *Workspace) SelectSelected(data interface{}) {
	if f := data.(SelectData).Selected; f != nil {
		f(wrk)
	}
}

func (wrk *Workspace) SelectHighlighted(data interface{}) {
	if f := data.(SelectData).Highlighted; f != nil {
		f(wrk)
	}
}
