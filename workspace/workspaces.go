package workspace

import (
	"fmt"
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
	HeadGeom(wrk *Workspace) xrect.Rect

	ActivateWorkspace(wrk *Workspace)
	SwitchWorkspaces(wrk1, wrk2 *Workspace)
}

type Workspaces struct {
	X             *xgbutil.XUtil
	Wrks          []*Workspace
	heads         Heads
	defaultLayout string
}

func NewWorkspaces(
	X *xgbutil.XUtil, heads Heads, defaultLayout string) *Workspaces {

	return &Workspaces{
		X:             X,
		Wrks:          make([]*Workspace, 0, 1),
		heads:         heads,
		defaultLayout: defaultLayout,
	}
}

// Add adds a new workspace to the set. Add will panic if a workspace with
// the same case-insensitive name as wrk already exists. Add will also panic
// if the workspace has a zero-length name.
func (wrks *Workspaces) Add(wrk *Workspace) {
	if len(wrk.Name) == 0 {
		panic("Workspaces must have a name of length at least one.")
	}
	for _, w := range wrks.Wrks {
		if strings.ToLower(w.Name) == strings.ToLower(wrk.Name) {
			panic(fmt.Sprintf("A workspace with the name '%s' already exists.",
				wrk.Name))
		}
	}
	wrks.Wrks = append(wrks.Wrks, wrk)
}

func (wrks *Workspaces) Remove(wrk *Workspace) {
	for i, wrk2 := range wrks.Wrks {
		if wrk == wrk2 {
			wrks.Wrks = append(wrks.Wrks[:i], wrks.Wrks[i+1:]...)
			wrk.Destroy()
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
