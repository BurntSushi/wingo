package main

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/wingo/logger"
)

type workspaces []*workspace

type workspace struct {
	id          int    // unique across workspaces; must index into workspaces!
	name        string // note that this does not have to be unique
	head        int    // the most recent physical head this workspace was on
	active      bool
	promptStore map[string]*window
}

func newWorkspace(id int) *workspace {
	wrk := &workspace{
		id:          id,
		name:        fmt.Sprintf("Default workspace %d", id+1),
		head:        -1,
		active:      false,
		promptStore: make(map[string]*window),
	}

	wrk.promptAdd()

	return wrk
}

func (wm *state) WrkActive() *workspace {
	for _, wrk := range wm.workspaces {
		if wrk.active {
			return wrk
		}
	}

	logger.Error.Printf("Could not find an active workspace in: %v",
		wm.workspaces)
	panic("Wingo *must* have an active workspace at all times. This is a bug!")
}

func (wm *state) WrkActiveInd() int {
	for i, wrk := range wm.workspaces {
		if wrk.active {
			return i
		}
	}

	logger.Error.Printf("Could not find an active workspace index in: %v",
		wm.workspaces)
	panic("Wingo *must* have an active workspace at all times. This is a bug!")
}

func (wm *state) WrkHead(head int) *workspace {
	for _, wrk := range wm.workspaces {
		if wrk.head == head {
			return wrk
		}
	}

	logger.Error.Printf("Could not find a workspace on head %d in: %v",
		head, wm.workspaces)
	panic("Wingo *must* have a workspace on each monitor at all times. " +
		"This is a bug!")
}

func (wm *state) WrkFind(name string) *workspace {
	for _, wrk := range wm.workspaces {
		if strings.ToLower(name) == strings.ToLower(wrk.name) {
			return wrk
		}
	}
	return nil
}

func (wm *state) WrkSet(wrk int, fallback bool, greedy bool) {
	if wrk > len(wm.workspaces) || wm.workspaces[wrk].active {
		return
	}

	wrkActive := wm.WrkActive()
	wrkNew := wm.workspaces[wrk]

	if !wrkNew.visible() {
		wrkActiveHead := wrkActive.head
		wrkActive.headSet(wrkNew.head)
		wrkNew.headSet(wrkActiveHead)

		wrkActive.hide()
		wrkNew.show()
	} else if greedy {
		wrkActiveHead := wrkActive.head
		wrkActive.headSet(wrkNew.head)
		wrkNew.headSet(wrkActiveHead)

		wrkActive.hide()
		wrkNew.hide()
		wrkActive.show()
		wrkNew.show()
	}

	wrkActive.active = false
	wrkNew.active = true

	if fallback {
		WM.fallback()
	}
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
		if wrk.visible() {
			c.Map()
		} else {
			c.Unmap()
		}
	}
}

func (wrk *workspace) visible() bool {
	return wrk.head > -1
}

func (wrk *workspace) nameSet(name string) {
	wrk.name = name
	wrk.promptUpdateName()
}

func (wrk *workspace) headSet(headNum int) {
	wrk.head = headNum
	if wrk.visible() {
		wrk.relayout()
	}
}

func (wrk *workspace) relayout() {
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
