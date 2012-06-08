package main

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/logger"
)

func (wm *state) ewmhSupportingWmCheck() {
	supportingWin := xwindow.Must(xwindow.Create(X, ROOT.Id))
	ewmh.SupportingWmCheckSet(X, ROOT.Id, supportingWin.Id)
	ewmh.SupportingWmCheckSet(X, supportingWin.Id, supportingWin.Id)
	ewmh.WmNameSet(X, supportingWin.Id, "Wingo")
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
	rootGeom, err := ROOT.Geometry()
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
