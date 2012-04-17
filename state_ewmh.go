package main

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/xgbutil/ewmh"

	"github.com/BurntSushi/wingo/logger"
)

func (wm *state) ewmhSupportingWmCheck() {
	supportingWin := createWindow(ROOT.id, 0)
	ewmh.SupportingWmCheckSet(X, ROOT.id, supportingWin.id)
	ewmh.SupportingWmCheckSet(X, supportingWin.id, supportingWin.id)
	ewmh.WmNameSet(X, supportingWin.id, "Wingo")
}

func (wm *state) ewmhDesktopNames() {
	if wm == nil || wm.workspaces == nil {
		return // still starting up
	}

	names := make([]string, len(wm.workspaces))
	for i, wrk := range wm.workspaces {
		if len(strings.TrimSpace(wrk.name)) == 0 {
			names[i] = fmt.Sprintf("Default workspace %d", i)
		} else {
			names[i] = wrk.name
		}
	}
	ewmh.DesktopNamesSet(X, names)
}

// ewmhWorkarea is responsible for syncing _NET_WORKAREA with the current
// workspace state.
// Since multiple workspaces can be viewable at one time, this property
// doesn't make much sense. So I'm not going to implement it until it's obvious
// that I have to.
func (wm *state) ewmhWorkarea() {
}

// ewmhDesktopGeometry is another totally useless property. Christ.
func (wm *state) ewmhDesktopGeometry() {
	rootGeom, err := ROOT.geometry()
	if err != nil {
		logger.Error.Printf("Could not get ROOT window geometry: %s", err)
		panic("")
	}

	ewmh.DesktopGeometrySet(X,
		&ewmh.DesktopGeometry{
			Width:  rootGeom.Width(),
			Height: rootGeom.Height(),
		})
}
