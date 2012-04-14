package main

import (
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xinerama"
	"github.com/BurntSushi/xgbutil/xrect"

	"github.com/BurntSushi/wingo/logger"
)

func rootGeometryChange(X *xgbutil.XUtil, ev xevent.ConfigureNotifyEvent) {
	WM.headsLoad()
}

func (wm *state) headActive() xrect.Rect {
	return wm.headGeom(wm.wrkActive().head)
}

func (wm *state) headGeom(i int) xrect.Rect {
	if i < 0 || i >= len(wm.heads) {
		return nil
	}
	return wm.heads[i]
}

// headChoose *only* looks at the client's geometry, determines which head
// it overlaps with the most, and updates that client's workspace appropriately.
// If the workspace is updated, headChoose returns true. Otherwise, false.
func (wm *state) headChoose(c *client, newGeom xrect.Rect) bool {
	// If this client isn't mapped, don't do anything.
	if !c.Mapped() {
		return false
	}

	mostOverlap := xrect.LargestOverlap(newGeom, WM.headsRaw)

	// If there's no overlap or no change, just leave the client where it is.
	if mostOverlap < 0 || mostOverlap == c.workspace.head {
		return false
	}

	// mostOverlap is a different monitor.
	// So switch this client to its new workspace.
	wrk := WM.wrkHead(mostOverlap)
	wrk.add(c)

	// If this is an active client, then update the active workspace too!
	if c.state == StateActive {
		wrk.activate(false, false)
	}
	return true
}

// headConvert takes a source and a destination rect, along with a rect
// in the source's rectangle, and returns a new rect translated into the
// destination rect.
func (wm *state) headConvert(rect, src, dest xrect.Rect) xrect.Rect {
	nx, ny, nw, nh := xrect.Pieces(rect)

	rectRatio := func(r xrect.Rect) float64 {
		return float64(r.Width()) / float64(r.Height())
	}
	ratio := rectRatio(dest) / rectRatio(src)

	nx = int(ratio*float64(nx-src.X())) + dest.X()
	ny = int(ratio*float64(ny-src.Y())) + dest.Y()

	// XXX: Allow window scaling as a config option.

	return xrect.New(nx, ny, nw, nh)
}

func (wm *state) headsLoad() {
	heads := stateHeadsGet()

	if len(wm.workspaces) < len(heads) {
		wm.fillWorkspaces(heads)
	}

	// Is this the first time loading heads?
	firstTime := true
	for _, wrk := range wm.workspaces {
		if wrk.active {
			firstTime = false
			break
		}
	}

	// Make the first one active and the first N workspaces visible,
	// where N is the number of heads
	if firstTime {
		wm.workspaces[0].active = true
		for i := 0; i < len(heads); i++ {
			wm.workspaces[i].headSet(i)
		}
	} else {
		// make sure we have no workspaces with bad heads
		activeHidden := false
		for _, wrk := range wm.workspaces {
			if wrk.head >= len(heads) {
				wrk.hide()
				wrk.headSet(-1)
				if wrk.active {
					wrk.active = false
					activeHidden = true
				}
			}
		}

		// now make sure we have one workspace attached to every head
		for i, _ := range heads {
			attached := false
			for _, wrk := range wm.workspaces {
				if wrk.head == i {
					attached = true
					break
				}
			}
			if !attached {
				for _, wrk := range wm.workspaces {
					if !wrk.visible() {
						wrk.headSet(i)
						wrk.show()
						break
					}
				}
			}
		}

		// finally, if we've hidden the active workspace, give up
		// and activate the first visible workspace
		if activeHidden {
			for _, wrk := range wm.workspaces {
				if wrk.visible() {
					wrk.active = true
					break
				}
			}
		}

		// totally. we may have hidden the active workspace, so we'll need
		// to update the focus!
		wm.fallback()
	}

	// update the state of the world
	wm.headsRaw = heads

	// apply struts!
	wm.headsApplyStruts()
}

// headsApplyStruts looks for struts set on all clients, and applies them
// to the current set of heads.
func (wm *state) headsApplyStruts() {
	// reset the current heads
	wm.headsReload()

	// now go through each client, find any struts and apply them. easy peasy!
	for _, c := range wm.clients {
		strut, _ := ewmh.WmStrutPartialGet(X, c.Id())
		if strut == nil {
			continue
		}
		logger.Debug.Println(c)
		xrect.ApplyStrut(wm.heads, ROOT.geom.Width(), ROOT.geom.Height(),
			strut.Left, strut.Right, strut.Top, strut.Bottom,
			strut.LeftStartY, strut.LeftEndY,
			strut.RightStartY, strut.RightEndY,
			strut.TopStartX, strut.TopEndX,
			strut.BottomStartX, strut.BottomEndX)
	}

	// Make currently visible and maximized clients fix themselves.
	for _, c := range wm.clients {
		if c.isMapped && c.maximized {
			c.maximize()
		}
	}
}

// fillWorkspaces is used when there are more heads than there are workspaces.
// This may be due to bad configuration OR if a head has been added with
// too few workspaces already existing.
func (wm *state) fillWorkspaces(heads xinerama.Heads) {
	logger.Warning.Println("There were not enough workspaces found." +
		"Namely, there must be at least " +
		"as many workspaces as there are phyiscal heads. " +
		"We are forcefully making some and " +
		"moving on. Please report this as a bug if you " +
		"think you're configuration is correct.")

	for i := len(wm.workspaces); i < len(heads); i++ {
		wm.workspaces = append(wm.workspaces, newWorkspace(i))
	}
}

// stateHeadsGet does the plumbing to get the physical head info from Xinerama.
// Remember, Xinerama may be dated, but the extension doesn't have to be
// explicitly used for it to be useful. Namely, both RandR and TwinView report
// information via Xinerama. The only real down-side here is that we have
// to listen to geometry changes on the root window, rather than using RandR
// to listen to OutputChange events.
func stateHeadsGet() xinerama.Heads {
	rawHeads, err := xinerama.PhysicalHeads(X)
	if err != nil || len(rawHeads) == 0 {
		if err == nil {
			logger.Warning.Printf("Could not find any physical heads with " +
				"the Xinerama extension.")
		} else {
			logger.Warning.Printf("Could not load physical heads via "+
				"Xinerama: %s", err)
		}
		logger.Warning.Printf("Assuming one head with size equivalent to the " +
			"root window.")

		rawHeads = xinerama.Heads{
			xrect.New(ROOT.geom.X(), ROOT.geom.Y(),
				ROOT.geom.Width(), ROOT.geom.Height()),
		}
	}
	return rawHeads
}

// headsReload puts the raw monitor geometry back into wm.heads
func (wm *state) headsReload() {
	wm.heads = make(xinerama.Heads, len(wm.headsRaw))
	for i, h := range wm.headsRaw {
		wm.heads[i] = xrect.New(h.X(), h.Y(), h.Width(), h.Height())
	}
}
