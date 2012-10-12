package prompt

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/text"
)

type SelectChoice interface {
	SelectText() string
	SelectSelected(data interface{})
	SelectHighlighted(data interface{})
}

type SelectItem struct {
	slct   *Select
	choice SelectChoice

	text                 string
	regular, highlighted *xwindow.Window
}

func newSelectItem(slct *Select, choice SelectChoice) *SelectItem {
	si := &SelectItem{
		slct:   slct,
		choice: choice,
	}

	si.regular = xwindow.Must(xwindow.Create(si.slct.X, si.slct.win.Id))
	si.highlighted = xwindow.Must(xwindow.Create(si.slct.X, si.slct.win.Id))

	// If the text overruns, make sure it's below the borders.
	si.regular.StackSibling(si.slct.bRht.Id, xproto.StackModeBelow)
	si.highlighted.StackSibling(si.slct.bRht.Id, xproto.StackModeBelow)

	si.UpdateText()

	return si
}

func (si *SelectItem) show(x, y int) {
	si.regular.Move(x, y)
	si.highlighted.Move(x, y)
	si.regular.Map()
}

func (si *SelectItem) hide() {
	si.regular.Unmap()
	si.highlighted.Unmap()
}

func (si *SelectItem) choose() {
	si.choice.SelectSelected(si.slct.data)
}

func (si *SelectItem) highlight() {
	si.choice.SelectHighlighted(si.slct.data)
	si.highlighted.Map()
	si.regular.Unmap()
}

func (si *SelectItem) unhighlight() {
	si.regular.Map()
	si.highlighted.Unmap()
}

func (si *SelectItem) Destroy() {
	si.regular.Destroy()
	si.highlighted.Destroy()
}

func (si *SelectItem) UpdateText() {
	t := si.slct.theme
	si.text = si.choice.SelectText()

	// Always have some text.
	if len(si.text) == 0 {
		si.text = "N/A"
	}

	err := text.DrawText(si.regular, t.Font, t.FontSize,
		t.FontColor, t.BgColor, si.text)
	if err != nil {
		logger.Warning.Printf("(*SelectItem).UpdateText: "+
			"Could not render text: %s", err)
	}

	err = text.DrawText(si.highlighted, t.Font, t.FontSize,
		t.ActiveFontColor, t.ActiveBgColor, si.text)
	if err != nil {
		logger.Warning.Printf("(*SelectItem).UpdateText: "+
			"Could not render text: %s", err)
	}
}
