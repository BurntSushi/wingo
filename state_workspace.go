package main

import "strings"

type workspaces []*workspace

type workspace struct {
    id int // unique across all workspaces; must also be index into workspaces!
    name string // note that this does not have to be unique
    head int // the most recent physical head this workspace was on
    active bool
}

func (wm *state) WrkActive() *workspace {
    for _, wrk := range wm.workspaces {
        if wrk.active {
            return wrk
        }
    }

    logError.Printf("Could not find an active workspace in: %v", wm.workspaces)
    panic("Wingo *must* have an active workspace at all times. This is a bug!")
}

func (wm *state) WrkActiveInd() int {
    for i, wrk := range wm.workspaces {
        if wrk.active {
            return i
        }
    }

    logError.Printf("Could not find an active workspace index in: %v",
                    wm.workspaces)
    panic("Wingo *must* have an active workspace at all times. This is a bug!")
}

func (wm *state) WrkFind(name string) *workspace {
    for _, wrk := range wm.workspaces {
        if strings.ToLower(name) == strings.ToLower(wrk.name) {
            return wrk
        }
    }
    return nil
}

func (wm *state) WrkSet(wrk int, greedy bool) {
    if wrk > len(wm.workspaces) || wm.workspaces[wrk].active {
        return
    }

    wrkActive := wm.WrkActive()
    wrkNew := wm.workspaces[wrk]

    if wrkNew.head <= -1 {
        wrkActive.head, wrkNew.head = wrkNew.head, wrkActive.head
        wrkActive.hide()
        wrkNew.show()
    } else if greedy {
        wrkActive.head, wrkNew.head = wrkNew.head, wrkActive.head

        wrkActive.hide()
        wrkNew.hide()
        wrkActive.show()
        wrkNew.show()
    }

    wrkActive.active = false
    wrkNew.active = true

    WM.fallback()
}

func (wrk *workspace) Add(c *client, checkVisible bool) {
    cwork := c.workspace
    wrk.add(c, checkVisible)

    // Don't forget to add transients
    for _, c2 := range WM.clients {
        if c.transient(c2) && c2.workspace == cwork {
            wrk.add(c2, checkVisible)
        }
    }
}

func (wrk *workspace) add(c *client, checkVisible bool) {
    // Resist change if we don't need it.
    if c.workspace == wrk.id {
        return
    }

    // Look at the client's current workspace, and if it's valid, remove it.
    if c.workspace >= 0 && c.workspace < len(WM.workspaces) {
        // code for removing a client from a workspace
    }

    // This should be the *only* place this happens!!!
    c.workspace = wrk.id

    // It's okay if the following map/unmap is redundant with the client's
    // current state. They will bail appropriately if so.
    if checkVisible {
        if wrk.head > -1 {
            c.Map()
        } else {
            c.Unmap()
        }
    }
}

func (wrk *workspace) hide() {
    for _, c := range WM.clients {
        if c.workspace == wrk.id {
            c.Unmap()
        }
    }
}

func (wrk *workspace) show() {
    for _, c := range WM.clients {
        if c.workspace == wrk.id {
            c.Map()
        }
    }
}

func (wrk *workspace) Activate() {

}

