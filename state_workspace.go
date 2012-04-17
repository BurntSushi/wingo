package main

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xrect"

	"github.com/BurntSushi/wingo/logger"
)

const (
	workspaceFloating = iota
	workspaceTiling
)

type workspaces []*workspace

type workspace struct {
	id          int    // unique across workspaces; must index into workspaces!
	name        string // note that this does not have to be unique
	head        int    // the most recent physical head this workspace was on
	active      bool
	promptStore map[string]*window
	state       int // the default placement policy of this workspace
	floaters    []layout
	tilers      []layout
}

func newWorkspace(id int) *workspace {
	wrk := &workspace{
		id:          id,
		name:        fmt.Sprintf("Default workspace %d", id+1),
		head:        -1,
		active:      false,
		promptStore: make(map[string]*window),
		state:       workspaceFloating,
	}
	wrk.floaters = []layout{newFloating(wrk)}
	wrk.tilers = []layout{newTileVertical(wrk)}

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
	if wrkNew.active || wrkNew.id < 0 {
		return
	}

	wrkActive := WM.wrkActive()

	if !wrkNew.visible() {
		wrkActive.hide()

		wrkActiveHead := wrkActive.head
		wrkActive.headSet(wrkNew.head)
		wrkNew.headSet(wrkActiveHead)

		wrkNew.show()
	} else if greedy {
		wrkActive.hide()
		wrkNew.hide()

		wrkActiveHead := wrkActive.head
		wrkActive.headSet(wrkNew.head)
		wrkNew.headSet(wrkActiveHead)

		wrkActive.show()
		wrkNew.show()
	}

	wrkActive.activeSet(false)
	wrkNew.activeSet(true)

	if fallback {
		WM.fallback()
	}
}

func (wrk *workspace) activeSet(active bool) {
	wrk.active = active
	if active {
		ewmh.CurrentDesktopSet(X, wrk.id)
	}
}

func (wrk *workspace) add(c *client) {
	// Don't forget to add transients if this isn't the client's first workspace
	if c.workspace != nil {
		for _, c2 := range WM.clients {
			if c.transient(c2) && c2.workspace != nil &&
				c2.workspace.id == c.workspace.id {

				wrk.addSingle(c2)
			}
		}
	}

	wrk.addSingle(c)
}

func (wrk *workspace) addSingle(c *client) {
	// Resist change if we don't need it.
	if c.workspace != nil && c.workspace.id == wrk.id {
		return
	}

	// Look at the client's current workspace, and if it's valid, remove it.
	if c.workspace != nil {
		c.workspace.remove(c)
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
	c.layoutSet()
	wrk.tile()
}

func (wrk *workspace) remove(c *client) {
	wrk.tilersRemove(c)
	c.workspace = nil // better be trashing this client or updating this soon!
	wrk.tile()
}

func (wrk *workspace) tilersAdd(c *client) {
	for _, ly := range wrk.tilers {
		ly.add(c)
	}
}

func (wrk *workspace) tilersRemove(c *client) {
	for _, ly := range wrk.tilers {
		ly.remove(c)
	}
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
	return wrk.head > -1 || wrk.id < 0
}

func (wrk *workspace) nameSet(name string) {
	wrk.name = name
	wrk.promptUpdateName()
	WM.ewmhUpdateDesktopNames()
}

func (wrk *workspace) headGeom() xrect.Rect {
	if wrk.id < 0 {
		logger.Error.Println(
			"The sticky workspace does not have a specific head geometry " +
				"associated with it. Assuming such things is a BUG.")
		panic("")
	}
	if wrk.head < 0 || wrk.head >= len(WM.heads) {
		logger.Error.Printf("'%d' is not a valid head number.", wrk.head)
		logger.Error.Printf(
			"Perhaps you're trying to access the geometry of a " +
				"hidden workspace? Bad mistake. :-/")
		panic("")
	}

	return WM.heads[wrk.head]
}

func (wrk *workspace) headSet(headNum int) {
	wrk.head = headNum
}

func (wrk *workspace) layout() layout {
	if wrk.tiling() {
		return wrk.tilers[0]
	}
	return wrk.floaters[0]
}

func (wrk *workspace) tile() {
	if wrk.tiling() {
		wrk.layout().place()
	}
}

func (wrk *workspace) untile() {
	if wrk.tiling() {
		wrk.layout().unplace()
	}
}

func (wrk *workspace) tiling() bool {
	return wrk.id >= 0 && wrk.state == workspaceTiling
}

func (wrk *workspace) tilingSet(enable bool) {
	if wrk.id < 0 {
		return
	}

	if enable {
		wrk.state = workspaceTiling
		wrk.tile()
	} else {
		wrk.untile()
		wrk.state = workspaceFloating
	}
}

func (wrk *workspace) hide() {
	for _, c := range WM.clients {
		if c.workspace.id == wrk.id {
			if c.layout().floating() {
				c.saveGeom("workspace_switch")
			}
			c.Unmap()
		}
	}
}

func (wrk *workspace) show() {
	wrk.tile()
	for _, c := range WM.stack {
		if c.workspace.id == wrk.id {
			if c.layout().floating() {
				c.loadGeom("workspace_switch")
			}
			c.Map()
		}
	}
}

func (wrk *workspace) String() string {
	return wrk.name
}
