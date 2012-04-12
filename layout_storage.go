package main

/*
	Layout storage is meant to serve as an abstraction of clients tiled
	on the screen. It is particularly useful when a layout uses the concept
	of a 'masters' and 'slaves' section.

	Layout storage is by no means necessary, particularly if the layout
	itself doesn't need to be stateful.

	Therefore, layout storage must be handled at the layout implementation
	level.
*/

import (
	"fmt"
)

type layoutSection []*layoutItem

func (s *layoutSection) add(item *layoutItem) {
	// Let's make sure we set the proper proportion.
	// To do this, we give it its equal share and take what we need
	// from each item in this section.
	item.proportion = 1.0 / float64(len(*s) + 1)
	for _, item2 := range *s {
		item2.proportion -= item2.proportion / float64(len(*s) + 1)
	}
	*s = append(*s, item)
}

func (s *layoutSection) remove(i int) {
	// First take the item out of the section slice.
	*s = append((*s)[:i], (*s)[i+1:]...)

	// Now add some proportion back to the remaining items in this section.
	for _, item := range *s {
		item.proportion += item.proportion / float64(len(*s))
	}
}

type layoutStore struct {
	masterNum int
	masters   layoutSection
	slaves    layoutSection
}

func newLayoutStorage() *layoutStore {
	return &layoutStore{
		masterNum: 1,
		masters:   make([]*layoutItem, 0),
		slaves:    make([]*layoutItem, 0),
	}
}

func (ls *layoutStore) add(c *client) bool {
	mi, si := ls.mFindClient(c), ls.sFindClient(c)
	if mi >= 0 || si >= 0 {
		return false
	}

	// Now we know this client isn't in storage, so create an item for it.
	item := newLayoutItem(c)

	// If we're short on masters, add it to the masters slice. Otherwise,
	// it goes to slaves.
	if len(ls.masters) < ls.masterNum {
		(&ls.masters).add(item)
	} else {
		(&ls.slaves).add(item)
	}

	return true
}

func (ls *layoutStore) remove(c *client) bool {
	if mi := ls.mFindClient(c); mi >= 0 {
		(&ls.masters).remove(mi)
		return true
	}
	if si := ls.sFindClient(c); si >= 0 {
		(&ls.slaves).remove(si)
		return true
	}
	return false
}

func (ls *layoutStore) mFindClient(c *client) int {
	for i, item := range ls.masters {
		if c.Id() == item.client.Id() {
			return i
		}
	}
	return -1
}

func (ls *layoutStore) sFindClient(c *client) int {
	for i, item := range ls.slaves {
		if c.Id() == item.client.Id() {
			return i
		}
	}
	return -1
}

type layoutItem struct {
	proportion float64 // 0 <= proportion <= 1
	client     *client
}

func newLayoutItem(c *client) *layoutItem {
	return &layoutItem{
		proportion: 0,
		client:     c,
	}
}

func (li *layoutItem) String() string {
	return fmt.Sprintf("%s [%f]", li.client, li.proportion)
}
