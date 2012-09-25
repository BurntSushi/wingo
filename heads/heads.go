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

func NewHeads(X *xgbutil.XUtil) *Heads {
	hds := &Heads{
		X:      X,
		active: 0,
	}
	hds.workspaces = workspace.NewWorkspaces(X, hds)
	return hds
}

func (hds *Heads) Initialize(clients Clients) {
	// Now workarea, geom, active and visible will be set.
	// Indeed, they are always set in keeping with the invariants when Load
	// is called.
	hds.Load(clients)
}

func (hds *Heads) Load(clients Clients) {
	hds.geom = query(hds.X)

	// Check if the number of workspaces is less than the number of heads.
	if len(hds.workspaces.Wrks) < len(hds.geom) {
		panic(fmt.Sprintf(
			"There must be at least %d workspaces (one for each head.",
			len(hds.geom)))
	}

	// To make things simple, set the first workspace to be the
	// active workspace, and setup the visibles slice as the first N workspaces
	// where N is the number of heads.
	// TODO: There may be a saner way of orienting workspaces when the
	// phyiscal heads change. Implement it!
	hds.ActivateWorkspace(hds.workspaces.Wrks[0])
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
			wrk.Show()
		} else {
			wrk.Hide()
		}
	}
}

func (hds *Heads) ApplyStruts(clients Clients) {
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

// Convert takes a source and a destination rect, along with a rect
// in the source's rectangle, and returns a new rect translated into the
// destination rect.
func Convert(rect, src, dest xrect.Rect) xrect.Rect {
	nx, ny, nw, nh := xrect.Pieces(rect)

	rectRatio := func(r xrect.Rect) float64 {
		return float64(r.Width()) / float64(r.Height())
	}
	ratio := rectRatio(dest) / rectRatio(src)

	nx = int(ratio*float64(nx-src.X())) + dest.X()
	ny = int(ratio*float64(ny-src.Y())) + dest.Y()

	// XXX: Allow window scaling as a config option.

	return xrect.New(nx, ny, nw, nh)
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
