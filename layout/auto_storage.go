package layout

/*
Layout storage is meant to serve as an abstraction of clients tiled
on the screen. It is particularly useful when a layout uses the concept
of a 'masters' and 'slaves' section.

Layout storage is by no means necessary, particularly if the layout
itself doesn't need to be stateful.
*/

import (
	"fmt"
)

type store struct {
	masterNum int
	masters   section
	slaves    section
}

func newStore() *store {
	return &store{
		masterNum: 1,
		masters:   make([]*item, 0),
		slaves:    make([]*item, 0),
	}
}

func (ls *store) add(c Client) bool {
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

func (ls *store) remove(c Client) bool {
	removed := false
	if mi := ls.mFindClient(c); mi >= 0 {
		(&ls.masters).remove(mi)
		removed = true
	}
	if si := ls.sFindClient(c); si >= 0 {
		(&ls.slaves).remove(si)
		removed = true
	}
	return removed
}

func (ls *store) mFindClient(c Client) int {
	for i, item := range ls.masters {
		if c.Id() == item.client.Id() {
			return i
		}
	}
	return -1
}

func (ls *store) sFindClient(c Client) int {
	for i, item := range ls.slaves {
		if c.Id() == item.client.Id() {
			return i
		}
	}
	return -1
}

type section []*item

func (s *section) add(item *item) {
	// Let's make sure we set the proper proportion.
	// To do this, we give it its equal share and take what we need
	// from each item in this section.
	item.proportion = 1.0 / float64(len(*s)+1)
	for _, item2 := range *s {
		item2.proportion -= item2.proportion / float64(len(*s)+1)
	}
	*s = append(*s, item)
}

func (s *section) remove(i int) {
	// First take the item out of the section slice.
	*s = append((*s)[:i], (*s)[i+1:]...)

	// Now add some proportion back to the remaining items in this section.
	for _, item := range *s {
		item.proportion += item.proportion / float64(len(*s))
	}
}

type item struct {
	proportion float64 // 0 <= proportion <= 1
	client     Client
}

func newLayoutItem(c Client) *item {
	return &item{
		proportion: 0,
		client:     c,
	}
}

func (li *item) String() string {
	return fmt.Sprintf("%s [%f]", li.client, li.proportion)
}
