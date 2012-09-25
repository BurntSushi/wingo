package heads

import (
	"github.com/BurntSushi/xgbutil/xrect"

	"github.com/BurntSushi/wingo/workspace"
)

func (hds *Heads) Workspaces() []*workspace.Workspace {
	return hds.workspaces.Wrks
}

// ActivateWorkspace will "focus" or "activate" the workspace provided.
// This only works when "wk" is visible.
// To activate a hidden workspace, please use SwitchWorkspaces.
func (hds *Heads) ActivateWorkspace(wk *workspace.Workspace) {
	wkvi := hds.visibleIndex(wk)
	if wkvi > -1 {
		hds.active = wkvi
	}
}

func (hds *Heads) SwitchWorkspaces(wk1, wk2 *workspace.Workspace) {
	v1, v2 := hds.visibleIndex(wk1), hds.visibleIndex(wk2)
	switch {
	case v1 > -1 && v2 > -1:
		hds.visibles[v1], hds.visibles[v2] = hds.visibles[v2], hds.visibles[v1]
		wk1.Place()
		wk2.Place()
	case v1 > -1 && v2 == -1:
		hds.visibles[v1] = wk2
		wk1.Hide()
		wk2.Show()
	case v1 == -1 && v2 > -1:
		hds.visibles[v2] = wk1
		wk2.Hide()
		wk1.Show()
	case v1 == -1 && v2 == -1:
		// Meaningless
	default:
		panic("unreachable")
	}
}

func (hds *Heads) NewWorkspace(name string) *workspace.Workspace {
	return hds.workspaces.NewWorkspace(name)
}

func (hds *Heads) AddWorkspace(wk *workspace.Workspace) {
	hds.workspaces.Add(wk)
}

func (hds *Heads) RemoveWorkspace(wk *workspace.Workspace) {
	// Don't allow it if this would result in fewer workspaces than there
	// are active physical heads.
	if len(hds.geom) == len(hds.workspaces.Wrks) {
		return
	}

	// There's a bit of complexity in choosing where to move the clients to.
	// Namely, if we're removing a hidden workspace, it's a simple matter of
	// moving the clients. However, if we're removing a visible workspace,
	// we have to make sure to make another workspace that is hidden take
	// its place. (Such a workspace is guaranteed to exist because we have at
	// least one more workspace than there are active physical heads.)
	if !wk.IsVisible() {
		moveClientsTo := hds.workspaces.Wrks[len(hds.workspaces.Wrks)-1]
		if moveClientsTo == wk {
			moveClientsTo = hds.workspaces.Wrks[len(hds.workspaces.Wrks)-2]
		}
		wk.RemoveAllAndAdd(moveClientsTo)
	} else {
		// Find the last-most hidden workspace that is not itself.
		for i := len(hds.workspaces.Wrks) - 1; i >= 0; i-- {
			work := hds.workspaces.Wrks[i]
			if work != wk && !work.IsVisible() {
				hds.SwitchWorkspaces(wk, work)
				wk.RemoveAllAndAdd(work)
				break
			}
		}
	}
	hds.workspaces.Remove(wk)
}

func (hds *Heads) ActiveWorkspace() *workspace.Workspace {
	return hds.visibles[hds.active]
}

func (hds *Heads) VisibleWorkspaces() []*workspace.Workspace {
	return hds.visibles
}

func (hds *Heads) IsActive(wrk *workspace.Workspace) bool {
	return hds.visibles[hds.active] == wrk
}

func (hds *Heads) Geom(wrk *workspace.Workspace) xrect.Rect {
	vi := hds.visibleIndex(wrk)
	if vi >= 0 {
		return hds.workarea[vi]
	}
	return nil
}

func (hds *Heads) visibleIndex(wk *workspace.Workspace) int {
	for i, vwk := range hds.visibles {
		if vwk == wk {
			return i
		}
	}
	return -1
}
