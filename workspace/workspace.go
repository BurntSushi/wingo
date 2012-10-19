package workspace

import (
	"fmt"
	"strings"

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
		layout.NewHorizontal(),
	}

	return wrk
}

func (wrk *Workspace) Destroy() {
	for _, lay := range wrk.floaters {
		lay.Destroy()
	}
	for _, lay := range wrk.autoTilers {
		lay.Destroy()
	}
	wrk.PromptSlctGroup.Destroy()
	wrk.PromptSlctItem.Destroy()
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
	if _, ok := c.Layout().(layout.Floater); ok {
		// If the workspace is visible, reload a state now.
		// Otherwise, we get a little hacky and copy the state into
		// workspace-switch, which will be invoked then the workspace
		// becomes visible...
		if wrk.IsVisible() {
			c.LoadState("last-floating")
		} else {
			c.CopyState("last-floating", "workspace-switch")
		}
	}

	// If the old and new workspace have different visibilities, adjust the
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
	for _, c := range wrk.Clients {
		if c.Workspace() != wrk {
			continue
		}
		newWk.Add(c)
	}
	newWk.Place()
}

func (wrk *Workspace) setGeom(geom xrect.Rect) {
	for _, lay := range wrk.floaters {
		lay.SetGeom(geom)
	}
	for _, lay := range wrk.autoTilers {
		lay.SetGeom(geom)
	}
}

func (wrk *Workspace) Show() {
	wrk.setGeom(wrk.Geom())
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
	wrk.setGeom(nil)
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

	// Refresh the geometry, since this is called right after struts change
	// or when monitor configuration changes.
	wrk.setGeom(wrk.Geom())

	// Floater layouts always get placed.
	wrk.LayoutFloater().Place()

	// Tiling layouts are only "placed" when the workspace is in the
	// appropriate layout mode.
	switch wrk.State {
	case Floating:
		// Nada nada limonada
	case AutoTiling:
		wrk.LayoutAutoTiler().Place()
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
		if _, ok := c.Layout().(layout.Floater); ok {
			c.LoadState("last-floating")
		}
		wrk.addToFloaters(c)
		wrk.addToTilers(c)
		c.IconifiedSet(false)

		wrk.Place()
		c.Map()
	} else {
		if _, ok := c.Layout().(layout.Floater); ok {
			c.SaveState("last-floating")
		}
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

func (wrk *Workspace) AutoCycle() {
	if wrk.State == AutoTiling {
		wrk.curAutoTiler = (wrk.curAutoTiler + 1) % len(wrk.autoTilers)
		wrk.LayoutAutoTiler().Place()
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

func (wrk *Workspace) SetLayout(name string) {
	var use layout.Layout = nil
	var index int

	name = strings.ToLower(name)
	for i, lay := range wrk.floaters {
		if name == strings.ToLower(lay.Name()) {
			use = lay
			index = i
			break
		}
	}
	if use == nil {
		for i, lay := range wrk.autoTilers {
			if name == strings.ToLower(lay.Name()) {
				use = lay
				index = i
				break
			}
		}
		if use == nil {
			return
		}
	}

	if _, ok := use.(layout.Floater); ok {
		wrk.curFloater = index
		wrk.LayoutStateSet(Floating)
	} else if _, ok := use.(layout.AutoTiler); ok {
		wrk.curAutoTiler = index
		wrk.LayoutStateSet(AutoTiling)
	} else {
		panic(fmt.Sprintf("Unknown layout type: %T", use))
	}
}

func (wrk *Workspace) LayoutName() string {
	switch wrk.State {
	case Floating:
		return wrk.LayoutFloater().Name()
	case AutoTiling:
		return wrk.LayoutAutoTiler().Name()
	}
	panic(fmt.Sprintf("Unknown workspace layout state: %d", wrk.State))
}

func (wrk *Workspace) LayoutStateSet(state int) {
	if !wrk.IsVisible() {
		return
	}

	if state == wrk.State {
		// If it's an AutoTiler, then just call Place again.
		if wrk.State == AutoTiling {
			wrk.LayoutAutoTiler().Place()
		}
		return
	}

	// First undo the current layout.
	switch wrk.State {
	case Floating:
		wrk.LayoutFloater().Save()
		wrk.LayoutFloater().Unplace()
	case AutoTiling:
		wrk.LayoutAutoTiler().Unplace()
	default:
		panic("Layout state not implemented.")
	}

	// Now apply the new layout.
	switch state {
	case Floating:
		wrk.State = state
		wrk.LayoutFloater().Place()
		wrk.LayoutFloater().Reposition()
	case AutoTiling:
		wrk.State = state
		wrk.LayoutAutoTiler().Place()
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
