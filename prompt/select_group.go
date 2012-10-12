package prompt

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/text"
)

type SelectGroup interface {
	SelectGroupText() string
}

// group is used internally to generate values implementing SelectGroup using
// (*Select).NewStaticGroup.
type group string

func (g group) SelectGroupText() string {
	return string(g)
}

type SelectGroupItem struct {
	slct  *Select
	group SelectGroup

	win *xwindow.Window
}

func newSelectGroupItem(slct *Select, group SelectGroup) *SelectGroupItem {
	si := &SelectGroupItem{
		slct:  slct,
		group: group,
	}

	si.win = xwindow.Must(xwindow.Create(si.slct.X, si.slct.win.Id))
	si.win.Change(xproto.CwBackPixel, si.slct.theme.BgColor.Uint32())

	// If the text overruns, make sure it's below the borders.
	si.win.StackSibling(si.slct.bRht.Id, xproto.StackModeBelow)

	si.UpdateText()

	return si
}

func (si *SelectGroupItem) hasGroup() bool {
	return !(si.win.Geom.Width() == 1 && si.win.Geom.Height() == 1)
}

func (si *SelectGroupItem) Destroy() {
	si.win.Destroy()
}

func (si *SelectGroupItem) UpdateText() {
	t := si.slct.theme
	txt := si.group.SelectGroupText()

	// Create a one pixel window and exit if there's no text.
	if len(txt) == 0 {
		si.win.Resize(1, 1)
		return
	}

	err := text.DrawText(si.win, t.GroupFont, t.GroupFontSize,
		t.GroupFontColor, t.GroupBgColor, txt)
	if err != nil {
		logger.Warning.Printf("(*SelectGroupItem).UpdateText: "+
			"Could not render text: %s", err)
	}
}

type SelectShowGroup struct {
	*SelectGroupItem
	items []*SelectItem
}

func (si *SelectGroupItem) ShowGroup(items []*SelectItem) *SelectShowGroup {
	return &SelectShowGroup{
		SelectGroupItem: si,
		items:           items,
	}
}

func (si *SelectShowGroup) show(x, y int) {
	si.win.Move(x, y)
	si.win.Map()
}

func (si *SelectShowGroup) hide() {
	si.win.Unmap()
	for _, item := range si.items {
		item.hide()
	}
}
