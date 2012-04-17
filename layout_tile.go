package main

import (
	"fmt"
	"strings"

	// "github.com/BurntSushi/wingo/logger" 
)

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

func (ly *tileVertical) floating() bool {
	return false
}

func (ly *tileVertical) place() {
	// If this layout isn't visible, don't do anything.
	if !ly.workspace.visible() {
		return
	}

	headGeom := ly.workspace.headGeom()
	msize, ssize := len(ly.store.masters), len(ly.store.slaves)

	mx, mw := headGeom.X(), int(float64(headGeom.Width())*ly.proportion)
	sx, sw := mx+mw, headGeom.Width()-mw

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
			item.client.saveGeomNoClobber("layout_before_tiling")
			item.client.FrameBorders()
			item.client.moveresizeNoValid(mx, headGeom.Y()+i*mh, mw, mh)
		}
	}
	if ssize > 0 {
		if msize == 0 {
			sx, sw = headGeom.X(), headGeom.Width()
		}
		sy := headGeom.Y()
		for _, item := range ly.store.slaves {
			sh := int(float64(headGeom.Height()) * item.proportion)
			item.client.saveGeomNoClobber("layout_before_tiling")
			item.client.FrameBorders()
			item.client.moveresizeNoValid(sx, sy, sw, sh)
			sy += sh
		}
	}

	// logger.Debug.Println(ly) 
}

func (ly *tileVertical) unplace() {
	// Instead of just hopping through the masters and slaves directly,
	// we start "unplacing" from the top of the stack. This makes untiling
	// with lots of windows appear quicker :-)
	for _, c := range WM.stack {
		if i := ly.store.mFindClient(c); i > -1 {
			ly.store.masters[i].client.loadGeom("layout_before_tiling")
		}
		if i := ly.store.sFindClient(c); i > -1 {
			ly.store.slaves[i].client.loadGeom("layout_before_tiling")
		}
	}
}

func (ly *tileVertical) add(c *client) {
	if added := ly.store.add(c); added {
	}
}

func (ly *tileVertical) remove(c *client) {
	if removed := ly.store.remove(c); removed {
	}
}

func (ly *tileVertical) maximizable() bool {
	return false
}

func (ly *tileVertical) move(c *client, x, y int) {
	c.moveNoValid(x, y)
}

func (ly *tileVertical) resize(c *client, w, h int) {
	c.resizeNoValid(w, h)
}

func (ly *tileVertical) moveresize(c *client, x, y, w, h int) {
	c.moveresizeNoValid(x, y, w, h)
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
	return fmt.Sprintf("\n%sTile Vertical on workspace '%s':\n"+
		"MASTERS:\n\t%s\nSLAVES:\n\t%s\n%s",
		sep, ly.workspace, strings.Join(masters, "\n"),
		strings.Join(slaves, "\n\t"), sep)
}
