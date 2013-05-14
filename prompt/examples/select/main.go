// Example select shows how to use a Select prompt from the prompt pacakge.
// Note that this example is rather messy and lacks documentation at the
// moment.
package main

import (
	"log"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xinerama"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo-conc/prompt"
)

var (
	// The key sequence to bring up the selection prompt.
	selectActivate = "Control-Mod4-Return"
)

type item struct {
	text string
	group int
	promptItem *prompt.SelectItem
}

func newItem(text string, group int) *item {
	return &item{
		text: text,
		group: group,
		promptItem: nil,
	}
}

func (item *item) SelectText() string {
	return item.text
}

func (item *item) SelectHighlighted(data interface{}) {
	log.Printf("highlighted: %s", item.text)
}

func (item *item) SelectSelected(data interface{}) {
	log.Printf("selected: %s", item.text)
}

func fatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	X, err := xgbutil.NewConn()
	fatal(err)

	keybind.Initialize(X)

	slct := prompt.NewSelect(X,
		prompt.DefaultSelectTheme, prompt.DefaultSelectConfig)

	// Create some artifical groups to use.
	artGroups := []prompt.SelectGroup{
		slct.NewStaticGroup("Group 1"),
		slct.NewStaticGroup("Group 2"),
		slct.NewStaticGroup("Group 3"),
		slct.NewStaticGroup("Group 4"),
		slct.NewStaticGroup("Group 5"),
	}

	// And now create some artificial items.
	items := []*item{
		newItem("andrew", 1), newItem("bruce", 2),
		newItem("kaitlyn", 3),
		newItem("cauchy", 4), newItem("plato", 1),
		newItem("platonic", 2),
		newItem("andrew gallant", 3),
		newItem("Andrew Gallant", 4), newItem("Andrew", 1),
		newItem("jim", 1), newItem("jimmy", 2),
		newItem("jimbo", 3),
	}

	groups := make([]*prompt.SelectGroupItem, len(artGroups))
	for i, artGroup := range artGroups {
		groups[i] = slct.AddGroup(artGroup)
	}
	for _, item := range items {
		item.promptItem = slct.AddChoice(item)
	}

	geom := headGeom(X)
	keybind.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			showGroups := newGroups(groups, items)
			slct.Show(geom, prompt.TabCompletePrefix, showGroups, nil)
		}).Connect(X, X.RootWin(), selectActivate, true)

	println("Loaded...")
	xevent.Main(X)
}

func newGroups(groups []*prompt.SelectGroupItem,
	items []*item) []*prompt.SelectShowGroup {

	showGroups := make([]*prompt.SelectShowGroup, 0)
	for i, group := range groups {
		showItems := make([]*prompt.SelectItem, 0)
		for _, item := range items {
			if item.group == i+1 {
				showItems = append(showItems, item.promptItem)
			}
		}
		showGroups = append(showGroups, group.ShowGroup(showItems))
	}
	return showGroups
}

func headGeom(X *xgbutil.XUtil) xrect.Rect {
	if X.ExtInitialized("XINERAMA") {
		heads, err := xinerama.PhysicalHeads(X)
		if err == nil {
			return heads[0]
		}
	}

	geom, err := xwindow.New(X, X.RootWin()).Geometry()
	fatal(err)
	return geom
}
