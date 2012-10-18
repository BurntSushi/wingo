package frame

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"

	"github.com/BurntSushi/wingo/render"
)

type Borders struct {
	*frame
	theme *BordersTheme

	topSide, bottomSide, leftSide, rightSide   *piece
	topLeft, topRight, bottomLeft, bottomRight *piece
}

func NewBorders(X *xgbutil.XUtil,
	t *BordersTheme, p *Parent, c Client) (*Borders, error) {

	f, err := newFrame(X, p, c)
	if err != nil {
		return nil, err
	}

	bf := &Borders{frame: f, theme: t}

	bf.topSide = bf.newTopSide()
	bf.bottomSide = bf.newBottomSide()
	bf.leftSide = bf.newLeftSide()
	bf.rightSide = bf.newRightSide()

	bf.topLeft = bf.newTopLeft()
	bf.topRight = bf.newTopRight()
	bf.bottomLeft = bf.newBottomLeft()
	bf.bottomRight = bf.newBottomRight()

	return bf, nil
}

func (f *Borders) Current() bool {
	return f.client.Frame() == f
}

func (f *Borders) Destroy() {
	f.topSide.Destroy()
	f.bottomSide.Destroy()
	f.leftSide.Destroy()
	f.rightSide.Destroy()

	f.topLeft.Destroy()
	f.topRight.Destroy()
	f.bottomLeft.Destroy()
	f.bottomRight.Destroy()

	f.frame.Destroy()
}

func (f *Borders) Off() {
	f.topSide.Unmap()
	f.bottomSide.Unmap()
	f.leftSide.Unmap()
	f.rightSide.Unmap()

	f.topLeft.Unmap()
	f.topRight.Unmap()
	f.bottomLeft.Unmap()
	f.bottomRight.Unmap()
}

func (f *Borders) On() {
	Reset(f)

	if f.client.State() == Active {
		f.Active()
	} else {
		f.Inactive()
	}

	if !f.client.IsMaximized() {
		f.topSide.Map()
		f.bottomSide.Map()
		f.leftSide.Map()
		f.rightSide.Map()

		f.topLeft.Map()
		f.topRight.Map()
		f.bottomLeft.Map()
		f.bottomRight.Map()
	}
}

func (f *Borders) Active() {
	f.State = Active

	f.topSide.Active()
	f.bottomSide.Active()
	f.leftSide.Active()
	f.rightSide.Active()

	f.topLeft.Active()
	f.topRight.Active()
	f.bottomLeft.Active()
	f.bottomRight.Active()

	f.parent.Change(xproto.CwBackPixel, uint32(0xffffff))
	f.parent.ClearAll()
}

func (f *Borders) Inactive() {
	f.State = Inactive

	f.topSide.Inactive()
	f.bottomSide.Inactive()
	f.leftSide.Inactive()
	f.rightSide.Inactive()

	f.topLeft.Inactive()
	f.topRight.Inactive()
	f.bottomLeft.Inactive()
	f.bottomRight.Inactive()

	f.parent.Change(xproto.CwBackPixel, uint32(0xffffff))
	f.parent.ClearAll()
}

func (f *Borders) Maximize() {
	if f.theme.BorderSize > 0 && f.Current() {
		f.topSide.Unmap()
		f.bottomSide.Unmap()
		f.leftSide.Unmap()
		f.rightSide.Unmap()

		f.topLeft.Unmap()
		f.topRight.Unmap()
		f.bottomLeft.Unmap()
		f.bottomRight.Unmap()

		Reset(f)
	}
}

func (f *Borders) Unmaximize() {
	if f.theme.BorderSize > 0 && f.Current() {
		f.topSide.Map()
		f.bottomSide.Map()
		f.leftSide.Map()
		f.rightSide.Map()

		f.topLeft.Map()
		f.topRight.Map()
		f.bottomLeft.Map()
		f.bottomRight.Map()

		Reset(f)
	}
}

func (f *Borders) Top() int {
	if f.client.IsMaximized() {
		return 0
	}
	return f.theme.BorderSize
}

func (f *Borders) Bottom() int {
	if f.client.IsMaximized() {
		return 0
	}
	return f.theme.BorderSize
}

func (f *Borders) Left() int {
	if f.client.IsMaximized() {
		return 0
	}
	return f.theme.BorderSize
}

func (f *Borders) Right() int {
	if f.client.IsMaximized() {
		return 0
	}
	return f.theme.BorderSize
}

func (f *Borders) moveresizePieces() {
	fg := f.Geom()

	f.topSide.MROpt(fW, 0, 0, fg.Width()-f.topLeft.w()-f.topRight.w(), 0)
	f.bottomSide.MROpt(fY|fW, 0, fg.Height()-f.bottomSide.h(), f.topSide.w(), 0)
	f.leftSide.MROpt(fH, 0, 0, 0, fg.Height()-f.topLeft.h()-f.bottomLeft.h())
	f.rightSide.MROpt(fX|fH, fg.Width()-f.rightSide.w(), 0, 0, f.leftSide.h())

	f.topRight.MROpt(fX, f.topLeft.w()+f.topSide.w(), 0, 0, 0)
	f.bottomLeft.MROpt(fY, 0, f.bottomSide.y(), 0, 0)
	f.bottomRight.MROpt(fX|fY,
		f.bottomLeft.w()+f.bottomSide.w(), f.bottomSide.y(), 0, 0)
}

func (f *Borders) MROpt(validate bool, flags, x, y, w, h int) {
	mropt(f, validate, flags, x, y, w, h)
	f.moveresizePieces()
}

func (f *Borders) MoveResize(validate bool, x, y, w, h int) {
	moveresize(f, validate, x, y, w, h)
	f.moveresizePieces()
}

func (f *Borders) Move(x, y int) {
	move(f, x, y)
}

func (f *Borders) Resize(validate bool, w, h int) {
	resize(f, validate, w, h)
	f.moveresizePieces()
}

type BordersTheme struct {
	BorderSize                 int
	AThinColor, IThinColor     render.Color
	ABorderColor, IBorderColor render.Color
}

func DefaultBordersTheme() *BordersTheme {
	return &BordersTheme{
		BorderSize:   10,
		AThinColor:   render.NewColor(0x0),
		IThinColor:   render.NewColor(0x0),
		ABorderColor: render.NewColor(0x3366ff),
		IBorderColor: render.NewColor(0xdfdcdf),
	}
}
