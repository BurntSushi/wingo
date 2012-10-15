package frame

import (
	"github.com/BurntSushi/xgbutil/xrect"
)

type Frame interface {
	Client() Client
	Parent() *Parent

	Map()
	Unmap()
	Destroy()
	Geom() xrect.Rect
	IsMapped() bool

	Move(x, y int)
	Resize(validate bool, width, height int)
	MoveResize(validate bool, x, y, width, height int)
	MROpt(validate bool, flags, x, y, width, height int)

	Moving() bool
	MovingState() *MoveState
	Resizing() bool
	ResizingState() *ResizeState

	Current() bool

	Top() int
	Bottom() int
	Left() int
	Right() int

	Off()
	On()
	Active()
	Inactive()
	Maximize()
	Unmaximize()
}

func Reset(f Frame) {
	geom := f.Geom()
	f.Resize(false, geom.Width(), geom.Height())
}

func ClientToFrame(f Frame, gravity, x, y, w, h int) (int, int, int, int) {
	return f.Client().GravitizeX(x, gravity),
		f.Client().GravitizeY(y, gravity),
		w + f.Left() + f.Right(),
		h + f.Top() + f.Bottom()
}

func validateWidthHeight(f Frame, validate bool,
	w, h int) (fw, fh, cw, ch int) {

	fw, fh, cw, ch = w, h, w, h

	cw -= f.Left() + f.Right()
	if validate {
		cw = f.Client().ValidateWidth(cw)
		fw = cw + f.Left() + f.Right()
	}

	ch -= f.Top() + f.Bottom()
	if validate {
		ch = f.Client().ValidateHeight(ch)
		fh = ch + f.Top() + f.Bottom()
	}

	return
}

func mropt(f Frame, validate bool, flags, x, y, w, h int) {
	fw, fh, cw, ch := validateWidthHeight(f, validate, w, h)

	f.Parent().MROpt(flags, x, y, fw, fh)
	f.Client().Win().MoveResize(f.Left(), f.Top(), cw, ch)
}

func moveresize(f Frame, validate bool, x, y, w, h int) {
	fw, fh, cw, ch := validateWidthHeight(f, validate, w, h)

	f.Parent().MoveResize(x, y, fw, fh)
	f.Client().Win().MoveResize(f.Left(), f.Top(), cw, ch)
}

func move(f Frame, x, y int) {
	f.Parent().Move(x, y)
}

func resize(f Frame, validate bool, w, h int) {
	fw, fh, cw, ch := validateWidthHeight(f, validate, w, h)

	f.Parent().Resize(fw, fh)
	f.Client().Win().MoveResize(f.Left(), f.Top(), cw, ch)
}

func Maximize(f Frame) {
	hg := xrect.New(xrect.Pieces(f.Client().HeadGeom()))
	f.MoveResize(false, hg.X(), hg.Y(), hg.Width(), hg.Height())
}
