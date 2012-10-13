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

	Workspaces *workspace.Workspaces  // Slice of all available workspaces.
	visibles   []*workspace.Workspace // Slice of all visible workspaces.
}

func NewHeads(X *xgbutil.XUtil) *Heads {
	hds := &Heads{
		X:      X,
		active: 0,
	}
	hds.Workspaces = workspace.NewWorkspaces(X, hds)
	return hds
}

func (hds *Heads) Initialize(clients Clients) {
	hds.geom = query(hds.X)

	// Check if the number of workspaces is less than the number of heads.
	if len(hds.Workspaces.Wrks) < len(hds.geom) {
		logger.Error.Fatalf(
			"There must be at least %d workspaces (one for each head).",
			len(hds.geom))
	}

	// To make things simple, set the first workspace to be the
	// active workspace, and setup the visibles slice as the first N workspaces
	// where N is the number of heads.
	// TODO: There may be a saner way of orienting workspaces when the
	// phyiscal heads change. Implement it!
	hds.ActivateWorkspace(hds.Workspaces.Wrks[0])
	hds.visibles = make([]*workspace.Workspace, len(hds.geom))
	for i := 0; i < len(hds.geom); i++ {
		hds.visibles[i] = hds.Workspaces.Wrks[i]
	}

	// Apply the struts set by clients to the workarea geometries.
	// This will fill in the hds.workarea slice.
	hds.ApplyStruts(clients)

	// Now show only the visibles and hide everything else.
	for _, wrk := range hds.Workspaces.Wrks {
		if wrk.IsVisible() {
			wrk.Show()
		} else {
			wrk.Hide()
		}
	}
}

func (hds *Heads) Reload(clients Clients) {
	hds.geom = query(hds.X)

	// Check if the number of workspaces is less than the number of heads.
	if len(hds.Workspaces.Wrks) < len(hds.geom) {
		logger.Error.Fatalf(
			"There must be at least %d workspaces (one for each head).",
			len(hds.geom))
	}

	logger.Message.Printf("Root window geometry had changed. Mirgrating "+
		"from %d heads to %d heads.", len(hds.visibles), len(hds.geom))

	// Here comes the tricky part. We may have more, less or the same number
	// of heads. But we'd like there to be as much of an overlap as possible
	// between the heads that were visible before and the heads that will
	// be visible. If we have the same number of heads as before, then we
	// don't much care about this.
	if len(hds.visibles) < len(hds.geom) {
		// We have more heads than we had before. So let's just expand our
		// visibles with some new workspaces. Remember, we're guaranteed to
		// have at least as many workspaces as heads.
		// We also leave the currently active workspace alone.
		for i := len(hds.visibles); i < len(hds.geom); i++ {
			// Find an available (i.e., hidden) workspace.
			for _, wrk := range hds.Workspaces.Wrks {
				if hds.visibleIndex(wrk) == -1 {
					hds.visibles = append(hds.visibles, wrk)
					break
				}
			}
		}
	} else if len(hds.visibles) > len(hds.geom) {
		// We now have fewer heads than we had before, so we'll reconstruct
		// our list of visibles, with care to keep the same ordering and to
		// keep the currently workspace still visible. (I believe this behavior
		// to be the least surprising to the user.)
		oldActive := hds.visibles[hds.active]
		oldvis := hds.visibles
		newvis := make([]*workspace.Workspace, len(hds.geom))
		newActive := -1
		newi := 0
		for oldi := 0; oldi < len(oldvis) && newi < len(newvis); oldi++ {
			// We always add this workspace, UNLESS we have only one spot left
			// and haven't added the active workspace yet. (We reserve that
			// last spot for the active workspace.)
			wrk := oldvis[oldi]
			if newActive == -1 && newi == len(newvis)-1 && !wrk.IsActive() {
				continue
			}

			newvis[newi] = wrk
			if wrk.IsActive() {
				newActive = newi
			}
			newi++
		}

		// Now that we've collected our new visibles list, we need to hide
		// all of the workspaces. (We'll show them later.) This is so that
		// they get properly refreshed into the right locations on the screen.
		for _, wrk := range hds.Workspaces.Wrks {
			wrk.Hide()
		}
		hds.visibles = newvis
		hds.ActivateWorkspace(hds.visibles[newActive])

		if oldActive != hds.visibles[hds.active] {
			panic(fmt.Sprintf("BUG: Old active workspace %s is not the same "+
				"as the new active workspace.",
				oldActive, hds.visibles[hds.active]))
		}
	}

	// Protect my sanity...
	if len(hds.visibles) != len(hds.geom) {
		panic(fmt.Sprintf("BUG: length of visibles (%d) != length of "+
			"geometry (%d)", len(hds.visibles), len(hds.geom)))
	}

	// Apply the struts set by clients to the workarea geometries.
	// This will fill in the hds.workarea slice.
	hds.ApplyStruts(clients)

	// Now show only the visibles and hide everything else.
	for _, wrk := range hds.Workspaces.Wrks {
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
	for _, wrk := range hds.Workspaces.Wrks {
		wrk.Place()
	}
	for i := 0; i < clients.Len(); i++ {
		c := clients.Get(i)
		if c.IsMaximized() {
			c.Remaximize()
		}
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

// NumHeads returns the current number of heads that Wingo is using.
func (hds *Heads) NumHeads() int {
	return len(hds.geom)
}

// NumConnected pings the Xinerama extension for a fresh tally of the number
// of heads currently active.
func (hds *Heads) NumConnected() int {
	return len(query(hds.X))
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
