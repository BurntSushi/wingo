package workspace

import (
	"strings"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xrect"
)

const (
	Floating = iota
	AutoTiling
	ManualTiling
)

type Heads interface {
	ActiveWorkspace() *Workspace
	VisibleWorkspaces() []*Workspace
	IsActive(wrk *Workspace) bool
	Geom(wrk *Workspace) xrect.Rect

	ActivateWorkspace(wrk *Workspace)
	SwitchWorkspaces(wrk1, wrk2 *Workspace)
}

type Workspaces struct {
	X     *xgbutil.XUtil
	Wrks  []*Workspace
	heads Heads
}

func NewWorkspaces(X *xgbutil.XUtil, heads Heads, names ...string) *Workspaces {
	if len(names) < 1 {
		panic("NewWorkspaces requires at least one name to create " +
			"the first workspace.")
	}

	workspaces := &Workspaces{
		X:     X,
		Wrks:  make([]*Workspace, 0, len(names)),
		heads: heads,
	}
	for _, name := range names {
		workspaces.Add(name)
	}
	return workspaces
}

func (wrks *Workspaces) Add(name string) {
	wrks.Wrks = append(wrks.Wrks, wrks.newWorkspace(name))
}

func (wrks *Workspaces) Remove(wrk *Workspace) {
	for i, wrk2 := range wrks.Wrks {
		if wrk == wrk2 {
			wrks.Wrks = append(wrks.Wrks[:i], wrks.Wrks[i+1:]...)
			return
		}
	}
}

func (wrks *Workspaces) Active() *Workspace {
	return wrks.heads.ActiveWorkspace()
}

func (wrks *Workspaces) Visibles() []*Workspace {
	return wrks.heads.VisibleWorkspaces()
}

func (wrks *Workspaces) Find(name string) *Workspace {
	name = strings.ToLower(name)
	for _, wrk := range wrks.Wrks {
		if name == strings.ToLower(wrk.Name) {
			return wrk
		}
	}
	return nil
}

func (wrks *Workspaces) Get(i int) *Workspace {
	if i < 0 || i >= len(wrks.Wrks) {
		return nil
	}
	return wrks.Wrks[i]
}
