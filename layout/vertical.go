package layout

import (
	"fmt"

	"github.com/BurntSushi/xgbutil/xrect"
)

type Vertical struct {
	store           *tree
	root            *hsplit
	masters, slaves *vsplit
	allowedMasters  int
}

func NewVertical() *Vertical {
	t := newTree()
	lay := &Vertical{
		store:          t,
		root:           newHSplit(nil),
		masters:        newVSplit(t.child),
		slaves:         newVSplit(t.child),
		allowedMasters: 1,
	}

	lay.root.prop = fullPortion
	lay.store.setChild(lay.root)
	lay.root.addNode(lay.masters, true)
	lay.root.addNode(lay.slaves, true)

	return lay
}

func (v *Vertical) AutoTiler() {}

func (v *Vertical) Place(geom xrect.Rect) {
	if geom == nil {
		return
	}
	v.store.place(geom)
}

func (v *Vertical) Unplace(geom xrect.Rect) {}

func (v *Vertical) Exists(c Client) bool {
	return v.store.findLeaf(c) != nil
}

func (v *Vertical) Add(c Client) {
	v.slaves.addNode(newLeaf(v.slaves, c), true)
	v.adjustMasters()
	v.adjustSplits()
}

func (v *Vertical) Remove(c Client) {
	if leaf := v.store.findLeaf(c); leaf != nil {
		switch {
		case leaf.parent == v.root:
			v.root.removeNode(leaf)
		case leaf.parent == v.masters:
			v.masters.removeNode(leaf)
		case leaf.parent == v.slaves:
			v.slaves.removeNode(leaf)
		}
		v.adjustMasters()
		v.adjustSplits()
	}
}

func (v *Vertical) adjustMasters() {
	// promote?
	if v.allowedMasters > 0 &&
		len(v.masters.children) < v.allowedMasters &&
		len(v.slaves.children) > 0 {

		// Just remove the first slave window and add it to the end of masters.
		n := v.slaves.children[0]
		v.slaves.removeNode(n)
		n.SetParent(v.masters)
		v.masters.addNode(n, true)
	}

	// demote?
	if len(v.masters.children) > v.allowedMasters {
		// Just remove the last master window and add it to the start of slaves.
		n := v.masters.children[len(v.masters.children)-1]
		v.masters.removeNode(n)
		n.SetParent(v.slaves)
		v.slaves.addNode(n, true)
	}
}

// This is responsible for adding or removing the master or slave splits.
func (v *Vertical) adjustSplits() {
	switch {
	case len(v.root.children) == 2:
		// We have both the master and slave splits. So make sure we have
		// some slave windows, otherwise we toss the slave split.
		if len(v.slaves.children) == 0 {
			v.root.removeNode(v.slaves)
		}
	case len(v.root.children) == 0:
		// *Either* the masters or slaves splits could have a child now.
		if len(v.masters.children) > 0 {
			v.root.addNode(v.masters, true)
		} else if len(v.slaves.children) > 0 {
			v.root.addNode(v.slaves, true)
		}
	case v.root.children[0] == v.masters:
		// Only need to check if the masters is empty or slaves is non-empty.
		if len(v.masters.children) == 0 {
			v.root.removeNode(v.masters)
		}
		if len(v.slaves.children) > 0 {
			v.root.addNode(v.slaves, true)
		}
	case v.root.children[0] == v.slaves:
		// Only need to check if the slaves is empty of masters is non-empty.
		if len(v.slaves.children) == 0 {
			v.root.removeNode(v.slaves)
		}
		if len(v.masters.children) > 0 {
			v.root.addNode(v.masters, false)
		}
	default:
		panic(fmt.Sprintf("Unknown state. len(masters) = %d, len(slaves) = %d",
			len(v.masters.children), len(v.slaves.children)))
	}
}

func (v *Vertical) MROpt(c Client, flags, x, y, width, height int) {}

func (v *Vertical) MoveResize(c Client, x, y, width, height int) {}

func (v *Vertical) Move(c Client, x, y int) {}

func (v *Vertical) Resize(c Client, width, height int) {}
