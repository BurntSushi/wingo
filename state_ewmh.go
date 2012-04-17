package main

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/xgbutil/ewmh"
)

func (wm *state) ewmhUpdateDesktopNames() {
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
