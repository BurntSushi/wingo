package main

type workspaces []*workspace

type workspace struct {
    id int // unique across all workspaces; must also be index into workspaces!
    name string // note that this does not have to be unique
    head int // the most recent physical head this workspace was on
    active bool
    visible bool
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

func (wm *state) WrkSet(wrk int) {
    if wrk > len(wm.workspaces) || wm.workspaces[wrk].active {
        return
    }

    wrkActive := wm.WrkActive()

    wrkActive.Hide()
    wm.workspaces[wrk].Show()

    wrkActive.active = false
    wm.workspaces[wrk].active = true

    WM.fallback()
}

func (wrk *workspace) Add(c *client, checkVisible bool) {
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
        if wrk.visible {
            c.Map()
        } else {
            c.Unmap()
        }
    }
}

func (wrk *workspace) Hide() {
    for _, c := range WM.clients {
        if c.workspace == wrk.id {
            c.Unmap()
        }
    }

    wrk.visible = false
}

func (wrk *workspace) Show() {
    for _, c := range WM.clients {
        if c.workspace == wrk.id {
            c.Map()
        }
    }

    wrk.visible = true
}

func (wrk *workspace) Activate() {

}

