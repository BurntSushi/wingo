package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

import (
    "github.com/BurntSushi/xgbutil/xevent"
    "github.com/BurntSushi/xgbutil/xrect"
)

type abstFrame struct {
    parent *frameParent
    clientOffset clientOffset
    moving *moveState
    resizing *resizeState
}

func newFrameAbst(p *frameParent, c *client, cp clientOffset) *abstFrame {
    f := &abstFrame {
        parent: p,
        clientOffset: cp,
        moving: &moveState{},
        resizing: &resizeState{},
    }

    return f
}

func (f *abstFrame) Destroy() {
    if f.Client().TrulyAlive() {
        X.Conn().ReparentWindow(f.Client().Id(), ROOT.id, 0, 0)
    }
    f.parent.window.destroy()
}

func (f *abstFrame) State() int {
    return f.Client().state
}

func (f *abstFrame) Map() {
    f.parent.window.map_()
}

func (f *abstFrame) Unmap() {
    f.parent.window.unmap()
}

func (f *abstFrame) Client() *client {
    return f.parent.client
}

// Configure is from the perspective of the client.
// Namely, the width and height specified here will be precisely the width
// and height that the client itself ends up with, assuming it passes
// validation. (Therefore, the actual window itself will be bigger, because
// of decorations.)
// Moreover, the x and y coordinates are gravitized. Yuck.
func (f *abstFrame) configureClient(flags, x, y, w, h int) (int, int,
                                                            int, int) {
    // Defy gravity!
    if DoX & flags > 0 {
        x = f.Client().GravitizeX(x, -1)
    }
    if DoY & flags > 0 {
        y = f.Client().GravitizeY(y, -1)
    }

    // This will change with other frames
    if DoW & flags > 0 {
        w += f.clientOffset.w
    }
    if DoH & flags > 0 {
        h += f.clientOffset.h
    }

    return x, y, w, h
}

// ConfigureFrame is from the perspective of the frame.
// The fw and fh specify the width of the entire window, so that the client
// will end up slightly smaller than the width/height specified here.
// Also, the fx and fy coordinates are interpreted plainly as root window
// coordinates. (No gravitization.)
func (f *abstFrame) configureFrame(flags, fx, fy, fw, fh int,
                                   sibling xgb.Id, stackMode byte,
                                   ignoreHints bool, sendNotify bool) {
    cw, ch := fw, fh
    framex, framey, _, _ := xrect.RectPieces(f.Geom())
    _, _, clientw, clienth := xrect.RectPieces(f.Client().Geom())

    if DoX & flags > 0 {
        framex = fx
    }
    if DoY & flags > 0 {
        framey = fy
    }
    if DoW & flags > 0 {
        cw -= f.clientOffset.w
        if !ignoreHints {
            cw = f.Client().ValidateWidth(cw)
            fw = cw + f.clientOffset.w
        }
        clientw = cw
    }
    if DoH & flags > 0 {
        ch -= f.clientOffset.h
        if !ignoreHints {
            ch = f.Client().ValidateHeight(ch)
            fh = ch + f.clientOffset.h
        }
        clienth = ch
    }

    if sendNotify {
        configNotify := xevent.NewConfigureNotify(f.Client().Id(),
                                                  f.Client().Id(),
                                                  0, framex, framey,
                                                  clientw, clienth, 0, false)
        X.Conn().SendEvent(false, f.Client().Id(), xgb.EventMaskStructureNotify,
                           configNotify.Bytes())
    }

    f.Parent().Win().configure(flags, fx, fy, fw, fh, sibling, stackMode)
    f.Client().Win().moveresize(flags | DoX | DoY,
                                f.clientOffset.x, f.clientOffset.y, cw, ch)
}

func (f *abstFrame) Geom() xrect.Rect {
    return f.parent.window.geom
}

func (f *abstFrame) Moving() bool {
    return f.moving.moving
}

func (f *abstFrame) MovingState() *moveState {
    return f.moving
}

func (f *abstFrame) Parent() *frameParent {
    return f.parent
}

func (f *abstFrame) ParentId() xgb.Id {
    return f.parent.window.id
}

func (f *abstFrame) ParentWin() *window {
    return f.parent.window
}

func (f *abstFrame) Resizing() bool {
    return f.resizing.resizing
}

func (f *abstFrame) ResizingState() *resizeState {
    return f.resizing
}


// ValidateHeight validates a height of a *frame*, which is equivalent
// to validating the height of a client.
func (f *abstFrame) ValidateHeight(height int) int {
    return f.Client().ValidateHeight(height - f.clientOffset.h) +
           f.clientOffset.h
}

// ValidateWidth validates a width of a *frame*, which is equivalent
// to validating the width of a client.
func (f *abstFrame) ValidateWidth(width int) int {
    return f.Client().ValidateWidth(width - f.clientOffset.w) +
           f.clientOffset.w
}

