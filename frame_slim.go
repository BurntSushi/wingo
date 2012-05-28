package main

import (
	"github.com/BurntSushi/xgb/xproto"
)

type frameSlim struct {
	*abstFrame
}

func newFrameSlim(p *frameParent, c *client) *frameSlim {
	return &frameSlim{newFrameAbst(p, c)}
}

func (f *frameSlim) Current() bool {
	return f.Client().Frame() == f
}

func (f *frameSlim) Off() {
}

func (f *frameSlim) On() {
	FrameReset(f)

	// Make sure the current state is properly shown
	if f.State() == StateActive {
		f.Active()
	} else {
		f.Inactive()
	}
}

func (f *frameSlim) Active() {
	f.ParentWin().change(xproto.CwBackPixel, uint32(THEME.slim.aBorderColor))
	f.ParentWin().clear()
}

func (f *frameSlim) Inactive() {
	f.ParentWin().change(xproto.CwBackPixel, uint32(THEME.slim.iBorderColor))
	f.ParentWin().clear()
}

func (f *frameSlim) Maximize() {
}

func (f *frameSlim) Unmaximize() {
}

func (f *frameSlim) Top() int {
	if f.Client().maximized {
		return 0
	}
	return THEME.slim.borderSize
}

func (f *frameSlim) Bottom() int {
	if f.Client().maximized {
		return 0
	}
	return THEME.slim.borderSize
}

func (f *frameSlim) Left() int {
	if f.Client().maximized {
		return 0
	}
	return THEME.slim.borderSize
}

func (f *frameSlim) Right() int {
	if f.Client().maximized {
		return 0
	}
	return THEME.slim.borderSize
}

func (f *frameSlim) ConfigureClient(flags, x, y, w, h int,
	sibling xproto.Window, stackMode byte, ignoreHints bool) {

	x, y, w, h = FrameConfigureClient(f, flags, x, y, w, h)
	f.ConfigureFrame(flags, x, y, w, h, sibling, stackMode, ignoreHints, true)
}

func (f *frameSlim) ConfigureFrame(flags, fx, fy, fw, fh int,
	sibling xproto.Window, stackMode byte, ignoreHints bool, sendNotify bool) {

	FrameConfigureFrame(f, flags, fx, fy, fw, fh, sibling, stackMode,
		ignoreHints, sendNotify)
}
