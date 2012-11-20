/*
Package workspace is responsible for maintaining the state for the workspaces
used in Wingo. It's responsible for showing or hiding a workspace, activating a
workspace, adding/removing workspaces, adding clients to workspaces, and
managing the state of all layouts for each workspace.

Heads

Most of the invariants surrounding attached physical heads are maintained in
the heads package. Chief among these invariants: 1) There must be at least as
many workspaces as there are heads. 2) There must be precisely N visible
workspaces where N is the number of active physical heads. 3) There must always
be one and only one active workspace.

The heads package maintains these invariants by 1) adding workspaces when there
aren't enough, 2) showing/hiding workspaces as appropriate the number of active
physical heads has changed, and 3) checking to make sure one workspace is
active after the number of physical heads has changed and being the only one
responsible for "switching" workspaces.

Namely, information about which workspace is active and which are
visible/hidden is maintained in the heads package.

Layout

Every layout is responsible for maintaining a list of clients that *may* use
that particular layout to be placed. In essence, this allows quick access to
all clients in a particular layout for a single workspace.

The list contains clients that *may* be in the particular layout because a
client's layout is determined by two factors: 1) if a client *must* not be
tiled, it will always use a floating layout and 2) the client's layout is the
default layout currently in use by the client's workspace.

Since a layout only maintains a list of potential clients, it is possible to
allow layouts to maintain state even when they are not active. That is, when a
workspace is in tiling mode, there could be clients in that workspace's
floating layout that are currently being tiled. So that if a floating layout
needs to access all of its clients, one could do:

	for _, client := range floatingLayout.clients {
		if _, ok := client.Layout().(layout.Floater); ok {
			// client is currently part of floatingLayout
		}
	}

For tiling layouts, this works a little differently since clients can only be
in a tiling layout if they aren't forced into a floating layout. Thus, clients
in a tiling layout's state can always be assumed to be using that tiler as a
layout.

This last point is only possible because layouts are not responsible for
knowing whether they are "active" or not. The workspace manages the activity of
layouts, and knows when to call "Place" on any particular layout.

(Note that iconified clients are removed from all layouts.)
*/
package workspace
