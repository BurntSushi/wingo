package main

import (
    "burntsushi.net/go/xgbutil"
    "burntsushi.net/go/xgbutil/xevent"
    "burntsushi.net/go/xgbutil/xinerama"
    "burntsushi.net/go/xgbutil/xrect"
)

func rootGeometryChange(X *xgbutil.XUtil, ev xevent.ConfigureNotifyEvent) {
    WM.headsLoad()
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
    wm.heads = heads
}

// fillWorkspaces is used when there are more heads than there are workspaces.
// This may be due to bad configuration OR if a head has been added with
// too few workspaces already existing.
func (wm *state) fillWorkspaces(heads xinerama.Heads) {
    logWarning.Println("There were not enough workspaces found." +
                       "Namely, there must be at least " +
                       "as many workspaces as there are phyiscal heads. " +
                       "We are forcefully making some and " +
                       "moving on. Please report this as a bug if you " +
                       "think you're configuration is correct.")

    for i := len(wm.workspaces); i < len(heads); i++ {
        wm.workspaces = append(wm.workspaces, newDefaultWorkspace(i))
    }
}

// stateHeadsGet does the plumbing to get the physical head info from Xinerama.
// Remember, Xinerama may be dated, but the extension doesn't have to be
// explicitly used for it to be useful. Namely, both RandR and TwinView report
// information via Xinerama. The only real down-side here is that we have
// to listen to geometry changes on the root window, rather than using RandR
// to listen to OutputChange events.
func stateHeadsGet() xinerama.Heads {
    heads, err := xinerama.PhysicalHeads(X)
    if err != nil || len(heads) == 0 {
        if err == nil {
            logWarning.Printf("Could not find any physical heads with the " +
                              "Xinerama extension.")
        } else {
            logWarning.Printf("Could not load physical heads via Xinerama: %s",
                              err)
        }
        logWarning.Printf("Assuming one head with size equivalent to the " +
                          "root window.")

        heads = xinerama.Heads{
            xrect.Make(ROOT.geom.X(), ROOT.geom.Y(),
                       ROOT.geom.Width(), ROOT.geom.Height()),
        }
    }
    return heads
}

