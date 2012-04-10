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

func (wm *state) wrkActive() *workspace {
	for _, wrk := range wm.workspaces {
		if wrk.active {
			return wrk
		}
	}

	logger.Error.Printf("Could not find an active workspace in: %v",
		wm.workspaces)
	panic("Wingo *must* have an active workspace at all times. This is a bug!")
}

func (wm *state) wrkHead(head int) *workspace {
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

func (wm *state) wrkFind(name string) *workspace {
	for _, wrk := range wm.workspaces {
		if strings.ToLower(name) == strings.ToLower(wrk.name) {
			return wrk
		}
	}
	return nil
}

func (wrkNew *workspace) activate(fallback bool, greedy bool) {
	if wrkNew.active {
		return
	}

	wrkActive := WM.wrkActive()

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

func (wrk *workspace) add(c *client) {
	wrk.addSingle(c)

	// Don't forget to add transients
	for _, c2 := range WM.clients {
		if c.transient(c2) && c2.workspace != nil &&
			c2.workspace.id == c.workspace.id {

			wrk.addSingle(c2)
		}
	}
}

func (wrk *workspace) addSingle(c *client) {
	// Resist change if we don't need it.
	if c.workspace != nil && c.workspace.id == wrk.id {
		return
	}

	// Look at the client's current workspace, and if it's valid, remove it.
	if c.workspace != nil {
		// code for removing a client from a workspace
	}

	// This should be the *only* place this happens!!!
	c.workspace = wrk

	// We're going to have to do layout stuff here.
	// To determine which layout the client will be in, we must first
	// determine whether or not it must be floating. If it must be floating,
	// obviously the client will be in the floating layout.
	// Otherwise, we must look at the currently active layout for this
	// workspace, and use that.
	// The aforementioned logic should actually be encapsulated somewhere
	// else (with a workspace receiver).
}

func (wrk *workspace) mapOrUnmap(c *client) {
	if wrk.visible() {
		c.Map()
	} else {
		c.Unmap()
	}

	for _, c2 := range WM.clients {
		if c.transient(c2) && c2.workspace != nil &&
			c2.workspace.id == c.workspace.id {

			if wrk.visible() {
				c2.Map()
			} else {
				c2.Unmap()
			}
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
		if c.workspace.id == wrk.id {
			c.Unmap()
		}
	}
}

func (wrk *workspace) show() {
	for _, c := range WM.clients {
		if c.workspace.id == wrk.id {
			c.Map()
		}
	}
}
