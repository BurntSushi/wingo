package heads

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/xgbutil/xrect"

	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/misc"
	"github.com/BurntSushi/wingo/workspace"
)

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
		wk1.Hide()
		wk2.Hide()
		hds.visibles[v1], hds.visibles[v2] = hds.visibles[v2], hds.visibles[v1]
		// wk1.Place() 
		// wk2.Place() 
		wk1.Show()
		wk2.Show()
	case v1 > -1 && v2 == -1:
		wk1.Hide()
		hds.visibles[v1] = wk2
		wk2.Show()
	case v1 == -1 && v2 > -1:
		wk2.Hide()
		hds.visibles[v2] = wk1
		wk1.Show()
	case v1 == -1 && v2 == -1:
		// Meaningless
	default:
		panic("unreachable")
	}
}

func (hds *Heads) NewWorkspace(name string) *workspace.Workspace {
	return hds.Workspaces.NewWorkspace(name)
}

func (hds *Heads) AddWorkspace(wk *workspace.Workspace) {
	hds.Workspaces.Add(wk)
}

func (hds *Heads) RemoveWorkspace(wk *workspace.Workspace) {
	// Don't allow it if this would result in fewer workspaces than there
	// are active physical heads.
	if len(hds.geom) == len(hds.Workspaces.Wrks) {
		panic("Cannot have fewer workspaces than active monitors.")
	}

	// A non-empty workspace cannot be removed.
	if len(wk.Clients) > 0 {
		panic(fmt.Sprintf("Non-empty workspace '%s' cannot be removed.", wk))
	}

	if wk.IsVisible() {
		// Find the last-most hidden workspace that is not itself and switch.
		for i := len(hds.Workspaces.Wrks) - 1; i >= 0; i-- {
			work := hds.Workspaces.Wrks[i]
			if work != wk && !work.IsVisible() {
				hds.SwitchWorkspaces(wk, work)
				break
			}
		}
	}
	hds.Workspaces.Remove(wk)
}

func (hds *Heads) ActiveWorkspace() *workspace.Workspace {
	return hds.visibles[hds.active]
}

func (hds *Heads) VisibleWorkspaces() []*workspace.Workspace {
	return hds.visibles
}

// WithVisibleWorkspace takes a head number and a closure and executes
// the closure safely with the workspace corresponding to head number i.
//
// This approach is necessary for safety, since the user can send commands
// with arbitrary head numbers. We need to make sure we don't crash if we
// get an invalid head number.
func (hds *Heads) WithVisibleWorkspace(i int, f func(w *workspace.Workspace)) {
	if i < 0 || i >= len(hds.visibles) {
		headNums := make([]string, len(hds.visibles))
		for j := range headNums {
			headNums[j] = fmt.Sprintf("%d", j)
		}
		logger.Warning.Printf("Head index %d is not valid. "+
			"Valid heads are: [%s].", i, strings.Join(headNums, ", "))
		return
	}
	f(hds.visibles[i])
}

func (hds *Heads) FindMostOverlap(needle xrect.Rect) *workspace.Workspace {
	haystack := make([]xrect.Rect, len(hds.geom))
	for i := range haystack {
		haystack[i] = hds.geom[i]
	}

	index := xrect.LargestOverlap(needle, haystack)
	if index == -1 {
		return nil
	}
	return hds.visibles[index]
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

func (hds *Heads) HeadGeom(wrk *workspace.Workspace) xrect.Rect {
	vi := hds.visibleIndex(wrk)
	if vi >= 0 {
		return hds.geom[vi]
	}
	return nil
}

func (hds *Heads) NextWorkspace() *workspace.Workspace {
	if cur := hds.GlobalIndex(hds.ActiveWorkspace()); cur > -1 {
		next := (cur + 1) % len(hds.Workspaces.Wrks)
		return hds.Workspaces.Get(next)
	}
	panic("bug")
}

func (hds *Heads) PrevWorkspace() *workspace.Workspace {
	if cur := hds.GlobalIndex(hds.ActiveWorkspace()); cur > -1 {
		// I fucking hate Go's modulo operator. WTF.
		prev := misc.Mod(cur-1, len(hds.Workspaces.Wrks))
		return hds.Workspaces.Get(prev)
	}
	panic("bug")
}

func (hds *Heads) visibleIndex(wk *workspace.Workspace) int {
	for i, vwk := range hds.visibles {
		if vwk == wk {
			return i
		}
	}
	return -1
}

func (hds *Heads) GlobalIndex(wkNeedle *workspace.Workspace) int {
	for i, wk := range hds.Workspaces.Wrks {
		if wk == wkNeedle {
			return i
		}
	}
	return -1
}
