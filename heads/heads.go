package heads

import (
	"fmt"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xinerama"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/workspace"
)

type Heads struct {
	X        *xgbutil.XUtil
	workarea xinerama.Heads // Slice of heads with struts applied.
	geom     xinerama.Heads // Raw geometry of heads.
	active   int            // Index in workarea/geom/visibles of active head.

	workspaces *workspace.Workspaces  // Slice of all available workspaces.
	visibles   []*workspace.Workspace // Slice of all visible workspaces.
}

func NewHeads(X *xgbutil.XUtil, clients workspace.Clients,
	workspaceNames ...string) *Heads {

	heads := &Heads{
		X:      X,
		active: -1, // Initialization value.
	}
	works := workspace.NewWorkspaces(X, heads, workspaceNames...)
	heads.workspaces = works

	// Now workarea, geom, active and visible will be set.
	// Indeed, they are always set in keeping with the invariants when Load
	// is called.
	heads.Load(clients)

	return heads
}

func (hds *Heads) Load(clients workspace.Clients) {
	hds.geom = query(hds.X)

	// If the number of workspaces is less than the number of heads,
	// add workspaces until number of workspaces equals number of heads.
	for i := len(hds.workspaces.Wrks); i < len(hds.geom); i++ {
		hds.workspaces.Add(fmt.Sprintf("Workspace %d", i))
	}

	// To make things simple, set the first workspace to be the
	// active workspace, and setup the visibles slice as the first N workspaces
	// where N is the number of heads.
	// TODO: There may be a saner way of orienting workspaces when the
	// phyiscal heads change. Implement it!
	hds.ActivateWorkspace(clients, hds.workspaces.Wrks[0])
	hds.visibles = make([]*workspace.Workspace, len(hds.geom))
	for i := 0; i < len(hds.geom); i++ {
		hds.visibles[i] = hds.workspaces.Wrks[i]
	}

	// Apply the struts set by clients to the workarea geometries.
	// This will fill in the hds.workarea slice.
	hds.ApplyStruts(clients)

	// Now show only the visibles and hide everything else.
	for _, wrk := range hds.workspaces.Wrks {
		if wrk.IsVisible() {
			wrk.Show(clients)
		} else {
			wrk.Hide(clients)
		}
	}
}

func (hds *Heads) ApplyStruts(clients workspace.Clients) {
	hds.workarea = make(xinerama.Heads, len(hds.geom))
	for i, hd := range hds.geom {
		hds.workarea[i] = xrect.New(hd.X(), hd.Y(), hd.Width(), hd.Height())
	}

	rgeom := xwindow.RootGeometry(hds.X)
	for i := 0; i < clients.Len(); i++ {
		c := clients.Get(i)

		strut, _ := ewmh.WmStrutPartialGet(hds.X, c.Id())
		if strut == nil {
			continue
		}
		xrect.ApplyStrut(hds.workarea, rgeom.Width(), rgeom.Height(),
			strut.Left, strut.Right, strut.Top, strut.Bottom,
			strut.LeftStartY, strut.LeftEndY,
			strut.RightStartY, strut.RightEndY,
			strut.TopStartX, strut.TopEndX,
			strut.BottomStartX, strut.BottomEndX)
	}
	for _, wrk := range hds.workspaces.Wrks {
		wrk.Place()
	}
}

func query(X *xgbutil.XUtil) xinerama.Heads {
	if X.ExtInitialized("XINERAMA") {
		heads, err := xinerama.PhysicalHeads(X)
		if err != nil || len(heads) == 0 {
			if err == nil {
				logger.Warning.Printf("Could not find any physical heads " +
					"with the Xinerama extension.")
			} else {
				logger.Warning.Printf("Could not load physical heads via "+
					"Xinerama: %s", err)
			}
			logger.Warning.Printf("Assuming one head with size equivalent " +
				"to the root window.")
		} else {
			return heads
		}
	}

	// If we're here, then something went wrong or the Xinerama extension
	// isn't available. So query the root window for its geometry and use that.
	rgeom := xwindow.RootGeometry(X)
	return xinerama.Heads{
		xrect.New(rgeom.X(), rgeom.Y(), rgeom.Width(), rgeom.Height()),
	}
}
