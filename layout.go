package main

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/wingo/logger"
)

type layout interface {
	place()
	add(c *client)
	remove(c *client)
}

type tileVertical struct {
	workspace  *workspace
	store      *layoutStore
	proportion float64 // 0 < proportion <= 1
}

func newTileVertical(wrk *workspace) *tileVertical {
	return &tileVertical{
		workspace:  wrk,
		store:      newLayoutStorage(),
		proportion: 0.5,
	}
}

func (ly *tileVertical) add(c *client) {
	if added := ly.store.add(c); added {
		logger.Debug.Println("ADDED NEW CLIENT:", c)
	}
}

func (ly *tileVertical) remove(c *client) {
	if removed := ly.store.remove(c); removed {
		logger.Debug.Println("REMOVED CLIENT:", c)
	}
}

func (ly *tileVertical) place() {
	headGeom := ly.workspace.headGeom()
	msize, ssize := len(ly.store.masters), len(ly.store.slaves)

	mx, mw := headGeom.X(), int(float64(headGeom.Width()) * ly.proportion)
	sx, sw := mx + mw, headGeom.Width() - mw

	// If we have zero widths, then we don't place.
	if mw <= 0 || mw > headGeom.Width() || sw <= 0 || sw > headGeom.Width() {
		return
	}

	if msize > 0 {
		mh := headGeom.Height() / msize
		if ssize == 0 {
			mw = headGeom.Width()
		}
		for i, item := range ly.store.masters {
			item.client.moveresize(mx, headGeom.Y() + i * mh, mw, mh)
		}
	}
	if ssize > 0 {
		if msize == 0 {
			sx, sw = headGeom.X(), headGeom.Width()
		}
		sy := headGeom.Y()
		for _, item := range ly.store.slaves {
			sh := int(float64(headGeom.Height()) * item.proportion)
			item.client.moveresize(sx, sy, sw, sh)
			sy += sh
		}
	}

	logger.Debug.Println(ly)
}

// in determines whether the client should be in this layout.
// Note that 'in' either returns true for every layout in a particular
// workspace or false for every layout in a particular workspace.
// It's up to the workspace to handle which layout is controlling placement.
func (ly *tileVertical) in(c *client) bool {
	return !c.floating &&
		ly.workspace.tiling() &&
		c.workspace.id == ly.workspace.id
}

// For debugging
func (ly *tileVertical) String() string {
	masters := make([]string, 0, len(ly.store.masters))
	slaves := make([]string, 0, len(ly.store.slaves))
	for _, item := range ly.store.masters {
		masters = append(masters, item.String())
	}
	for _, item := range ly.store.slaves {
		slaves = append(slaves, item.String())
	}
	sep := "--------------------------------------\n"
	return fmt.Sprintf("\n%sTile Vertical on workspace '%s':\n" +
		"MASTERS:\n\t%s\nSLAVES:\n\t%s\n%s",
		sep, ly.workspace, strings.Join(masters, "\n"),
		strings.Join(slaves, "\n\t"), sep)
}
