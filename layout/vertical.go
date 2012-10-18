package layout

import (
	"fmt"

	"github.com/BurntSushi/xgbutil/xrect"
)

type verthorz struct {
	store                 *tree
	root, masters, slaves splitter
	allowedMasters        int
	geom                  xrect.Rect
}

type Vertical struct {
	verthorz
}

type Horizontal struct {
	verthorz
}

func NewVertical() *Vertical {
	t := newTree()
	lay := &Vertical{verthorz{
		store:          t,
		root:           newHSplit(nil),
		masters:        newVSplit(t.child),
		slaves:         newVSplit(t.child),
		allowedMasters: 1,
	}}

	lay.root.SetProportion(fullPortion)
	lay.store.setChild(lay.root)
	lay.root.AddNode(lay.masters, true)
	lay.root.AddNode(lay.slaves, true)

	return lay
}

func NewHorizontal() *Horizontal {
	t := newTree()
	lay := &Horizontal{verthorz{
		store:          t,
		root:           newVSplit(nil),
		masters:        newHSplit(t.child),
		slaves:         newHSplit(t.child),
		allowedMasters: 1,
	}}

	lay.root.SetProportion(fullPortion)
	lay.store.setChild(lay.root)
	lay.root.AddNode(lay.masters, true)
	lay.root.AddNode(lay.slaves, true)

	return lay
}

func (v *Vertical) Name() string {
	return "Vertical"
}

func (v *Horizontal) Name() string {
	return "Horizontal"
}

func (lay verthorz) Destroy() {
}

func (lay *verthorz) SetGeom(geom xrect.Rect) {
	lay.geom = geom
}

func (lay verthorz) Place() {
	lay.store.place(lay.geom)
}

func (lay verthorz) Unplace() {}

func (lay verthorz) Exists(c Client) bool {
	return lay.store.findLeaf(c) != nil
}

func (lay verthorz) ResizeMaster(amount float64) {
	if lay.root.Size() == 2 {
		lay.root.PropsSave()

		newProp := lay.masters.Proportion() + proportion(amount)
		lay.root.SetChildProportion(lay.masters, newProp)

		if lay.store.place(lay.geom) {
			lay.root.PropsClear()
		} else {
			lay.root.PropsRollback()
		}
	}
}

func (lay verthorz) ResizeWindow(amount float64) {
	if lf := lay.leafCurrent(); lf != nil && lf.parent.Size() > 1 {
		lf.parent.PropsSave()

		newProp := lf.Proportion() + proportion(amount)
		lf.parent.SetChildProportion(lf, newProp)

		if lay.store.place(lay.geom) {
			lf.parent.PropsClear()
		} else {
			lf.parent.PropsRollback()
		}
	}
}

func (lay verthorz) Add(c Client) {
	lay.slaves.AddNode(newLeaf(lay.slaves, c), true)
	lay.adjustMasters()
	lay.adjustSplits()
}

func (lay verthorz) Remove(c Client) {
	if leaf := lay.store.findLeaf(c); leaf != nil {
		switch {
		case leaf.parent == lay.masters:
			lay.masters.RemoveNode(leaf)
		case leaf.parent == lay.slaves:
			lay.slaves.RemoveNode(leaf)
		default:
			panic(fmt.Sprintf("Client '%s' not in masters or slaves.", c))
		}
		lay.adjustMasters()
		lay.adjustSplits()
	}
}

func (lay verthorz) Next() {
	if lf := lay.leafCurrent(); lf != nil {
		next := lay.leafNext(lf).client
		next.Focus()
		next.Raise()
	}
}

func (lay verthorz) Prev() {
	if lf := lay.leafCurrent(); lf != nil {
		prev := lay.leafPrev(lf).client
		prev.Focus()
		prev.Raise()
	}
}

func (lay Horizontal) Next() {
	lay.verthorz.Prev()
}

func (lay Horizontal) Prev() {
	lay.verthorz.Next()
}

func (lay verthorz) SwitchNext() {
	if lf := lay.leafCurrent(); lf != nil {
		next := lay.leafNext(lf)
		lay.store.switchClients(lf, next)
		lay.Place()
	}
}

func (lay verthorz) SwitchPrev() {
	if lf := lay.leafCurrent(); lf != nil {
		next := lay.leafPrev(lf)
		lay.store.switchClients(lf, next)
		lay.Place()
	}
}

func (lay Horizontal) SwitchNext() {
	lay.verthorz.SwitchPrev()
}

func (lay Horizontal) SwitchPrev() {
	lay.verthorz.SwitchNext()
}

func (lay verthorz) FocusMaster() {
	if lay.masters.Size() > 0 {
		c := lay.masters.Child(0).(*leaf).client
		c.Focus()
		c.Raise()
	}
}

func (lay verthorz) MakeMaster() {
	if lf := lay.leafCurrent(); lf != nil && lay.masters.Size() > 0 {
		masterLeaf := lay.masters.Child(0).(*leaf)
		lay.store.switchClients(lf, masterLeaf)
		lay.Place()
	}
}

func (lay *verthorz) MastersMore() {
	lay.allowedMasters += 1
	lay.adjustMasters()
	lay.adjustSplits()
	lay.Place()
}

func (lay *verthorz) MastersFewer() {
	if lay.allowedMasters == 0 {
		return
	}
	lay.allowedMasters -= 1
	lay.adjustMasters()
	lay.adjustSplits()
	lay.Place()
}

func (lay verthorz) leafCurrent() *leaf {
	var lf *leaf
	lay.store.child.VisitLeafNodes(func(visit *leaf) bool {
		if visit.client.IsActive() {
			lf = visit
			return false
		}
		return true
	})
	return lf
}

func (lay verthorz) leafNext(lf *leaf) *leaf {
	var next node
	switch {
	case lf.parent == lay.masters:
		ind := lay.masters.ChildIndex(lf)
		if ind > 0 {
			next = lay.masters.Child(ind - 1)
		} else {
			if lay.slaves.Size() > 0 {
				next = lay.slaves.Child(0)
			} else {
				next = lay.masters.Child(lay.masters.Size() - 1)
			}
		}
	case lf.parent == lay.slaves:
		ind := lay.slaves.ChildIndex(lf)
		if ind < lay.slaves.Size()-1 {
			next = lay.slaves.Child(ind + 1)
		} else {
			if lay.masters.Size() > 0 {
				next = lay.masters.Child(lay.masters.Size() - 1)
			} else {
				next = lay.slaves.Child(0)
			}
		}
	default:
		panic(fmt.Sprintf("Leaf with client '%s' is not in masters or slaves.",
			lf.client))
	}
	return next.(*leaf)
}

func (lay verthorz) leafPrev(lf *leaf) *leaf {
	var prev node
	switch {
	case lf.parent == lay.masters:
		ind := lay.masters.ChildIndex(lf)
		if ind < lay.masters.Size()-1 {
			prev = lay.masters.Child(ind + 1)
		} else {
			if lay.slaves.Size() > 0 {
				prev = lay.slaves.Child(lay.slaves.Size() - 1)
			} else {
				prev = lay.masters.Child(0)
			}
		}
	case lf.parent == lay.slaves:
		ind := lay.slaves.ChildIndex(lf)
		if ind > 0 {
			prev = lay.slaves.Child(ind - 1)
		} else {
			if lay.masters.Size() > 0 {
				prev = lay.masters.Child(0)
			} else {
				prev = lay.slaves.Child(lay.slaves.Size() - 1)
			}
		}
	default:
		panic(fmt.Sprintf("Leaf with client '%s' is not in masters or slaves.",
			lf.client))
	}
	return prev.(*leaf)
}

func (lay verthorz) adjustMasters() {
	// promote?
	if lay.allowedMasters > 0 &&
		lay.masters.Size() < lay.allowedMasters &&
		lay.slaves.Size() > 0 {

		// Just remove the first slave window and add it to the end of masters.
		n := lay.slaves.Child(0)
		lay.slaves.RemoveNode(n)
		n.SetParent(lay.masters)
		lay.masters.AddNode(n, true)
	}

	// demote?
	if lay.masters.Size() > lay.allowedMasters {
		// Just remove the last master window and add it to the start of slaves.
		n := lay.masters.Child(lay.masters.Size() - 1)
		lay.masters.RemoveNode(n)
		n.SetParent(lay.slaves)
		lay.slaves.AddNode(n, false)
	}
}

// This is responsible for adding or removing the master or slave splits.
func (lay verthorz) adjustSplits() {
	switch {
	case lay.root.Size() == 2:
		// We have both the master and slave splits. So make sure we have
		// some slave windows, otherwise we toss the slave split.
		// Same with masters.
		if lay.masters.Size() == 0 {
			lay.root.RemoveNode(lay.masters)
		}
		if lay.slaves.Size() == 0 {
			lay.root.RemoveNode(lay.slaves)
		}
	case lay.root.Size() == 0:
		// *Either* the masters or slaves splits could have a child now.
		if lay.masters.Size() > 0 {
			lay.root.AddNode(lay.masters, true)
		} else if lay.slaves.Size() > 0 {
			lay.root.AddNode(lay.slaves, true)
		}
	case lay.root.Child(0) == lay.masters:
		// Only need to check if the masters is empty or slaves is non-empty.
		if lay.masters.Size() == 0 {
			lay.root.RemoveNode(lay.masters)
		}
		if lay.slaves.Size() > 0 {
			lay.root.AddNode(lay.slaves, true)
		}
	case lay.root.Child(0) == lay.slaves:
		// Only need to check if the slaves is empty of masters is non-empty.
		if lay.slaves.Size() == 0 {
			lay.root.RemoveNode(lay.slaves)
		}
		if lay.masters.Size() > 0 {
			lay.root.AddNode(lay.masters, false)
		}
	default:
		panic(fmt.Sprintf("Unknown state. len(masters) = %d, len(slaves) = %d",
			lay.masters.Size(), lay.slaves.Size()))
	}
}

func (lay verthorz) MROpt(c Client, flags, x, y, width, height int) {}

func (lay verthorz) MoveResize(c Client, x, y, width, height int) {}

func (lay verthorz) Move(c Client, x, y int) {}

func (lay verthorz) Resize(c Client, width, height int) {}
