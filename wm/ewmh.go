package wm

import (
	"fmt"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo-conc/event"
	"github.com/BurntSushi/wingo-conc/workspace"
)

func ewmhClientList() {
	ids := make([]xproto.Window, len(Clients))
	for i, client := range Clients {
		ids[i] = client.Id()
	}
	ewmh.ClientListSet(X, ids)
}

func ewmhNumberOfDesktops() {
	ewmh.NumberOfDesktopsSet(X, uint(len(Heads.Workspaces.Wrks)))
}

func ewmhCurrentDesktop() {
	ewmh.CurrentDesktopSet(X, uint(workspaceIndex(Workspace())))
	event.Notify(event.ChangedWorkspace{})
}

func ewmhVisibleDesktops() {
	visibles := Heads.VisibleWorkspaces()
	desks := make([]uint, len(visibles))
	for i, wrk := range visibles {
		desks[i] = uint(workspaceIndex(wrk))
	}
	ewmh.VisibleDesktopsSet(X, desks)

	event.Notify(event.ChangedVisibleWorkspace{})
}

func ewmhDesktopNames() {
	names := make([]string, len(Heads.Workspaces.Wrks))
	for i, wrk := range Heads.Workspaces.Wrks {
		names[i] = wrk.Name
	}
	ewmh.DesktopNamesSet(X, names)

	event.Notify(event.ChangedWorkspaceNames{})
}

func ewmhDesktopGeometry() {
	rgeom := xwindow.RootGeometry(X)

	ewmh.DesktopGeometrySet(X,
		&ewmh.DesktopGeometry{
			Width:  rgeom.Width(),
			Height: rgeom.Height(),
		})
}

func workspaceIndex(needle *workspace.Workspace) int {
	index := -1
	for i, wrk := range Heads.Workspaces.Wrks {
		if wrk == needle {
			index = i
			break
		}
	}
	if index == -1 {
		panic(fmt.Sprintf(
			"BUG: Could not determine index of workspace: %s", needle))
	}
	return index
}
