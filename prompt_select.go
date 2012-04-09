package main

/*
	prompt_select.go handles creation of a window prompt that can list
	arbitrary groups of items. The items are then searchable with tab
	completion.

	Arbitrary actions, represented by 'func()', can be run on selection.

	I am OK with the design, but I think there is room for simplification.
	I suspect interfaces could make things clearer, but I find it difficult
	to envision the proper level of abstraction when it only supports
	displaying two kinds of lists: clients and workspaces.

	Perhaps when I add layouts to the mix, things will be clearer.

	The best thing would be to separate prompts out into their own separate
	package, but there is too much coupling going on right now. Particular
	with functions in 'window.go'.

	The real cool thing would be to abstract this enough so that it could
	handle truly arbitrary lists. Like, directory listings. But I don't have
	the bandwidth for that at the moment.
*/

import (
	"strings"

	"code.google.com/p/jamslam-x-go-binding/xgb"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xgraphics"
)

const (
	textPadding  = 4  // spacing between input box borders and input text
	labelSpacing = 20 // spacing before each label heading
	labelVisible = "Visible"
	labelHidden  = "Hidden"
)

// promptSelectListFun is the type of function required to produce a list
// of items to be shown by the selection prompt.
type promptSelectListFun func() []*promptSelectGroup

// promptSelectGroup represents a list of items to display in a selection
// prompt under a specific heading.
type promptSelectGroup struct {
	label string
	win   *window
	items []*promptSelectItem
}

// newPromptSelectGroup creates a new group.
// If 'win' is nil, then the group has no visual effect, but can still be used
// to order items. Or, more practically, a singleton group with no label.
func newPromptSelectGroup(label string, win *window,
	items []*promptSelectItem) *promptSelectGroup {

	return &promptSelectGroup{
		label: label,
		win:   win,
		items: items,
	}
}

// promptSelectItem represents a single *selectable* item in a prompt.
type promptSelectItem struct {
	text     string  // visually displayed and used for tab completion
	action   func()  // performed when this item is selected
	active   *window // window w/ image when selected
	inactive *window // window w/ image when not selected
}

func newPromptSelectItem(text string, action func(),
	active, inactive *window) *promptSelectItem {

	return &promptSelectItem{
		text:     text,
		action:   action,
		active:   active,
		inactive: inactive,
	}
}

// promptSelect encapsulates information related to the visual display
// and function of a selection prompt. Most of the information in promptSelect
// is refreshed on each viewing, but the 'top', 'input', and 'b*' fields
// stay static throughout Wingo's lifespan.
type promptSelect struct {
	showing      bool                // whether the prompt is showing.
	selected     int                 // the selected item. -1 for no selection
	listFun      promptSelectListFun // list generator function
	prefixSearch bool                // whether prefix or substring search
	groups       []*promptSelectGroup
	itemsShowing []*promptSelectItem
	top          *window
	input        *textInput
	labVisible   *window
	labHidden    *window
	bInp         *window
	bTop, bBot   *window
	bLft, bRht   *window
}

// Id returns the parent window of this prompt.
func (ps *promptSelect) Id() xgb.Id {
	return ps.top.id
}

// newPromptSelect is run at Wingo startup and creates all of the windows
// necessary for basic construction. It also does as much positioning/resizing
// as it can without knowing the kind of list it will show.
// It also sets up a KeyPress handler that takes care of:
// 1) Text input
// 2) Canceling the prompt
// 3) Tabing through items that match search
// 4) Initiating the selection's action
func newPromptSelect() *promptSelect {
	top := createWindow(ROOT.id, 0)
	input := renderTextInputCreate(
		top, THEME.prompt.bgColor, THEME.prompt.font, THEME.prompt.fontSize,
		THEME.prompt.fontColor, 1000)
	bInp := createWindow(top.id, 0)
	bTop, bBot := createWindow(top.id, 0), createWindow(top.id, 0)
	bLft, bRht := createWindow(top.id, 0), createWindow(top.id, 0)

	bs := THEME.prompt.borderSize
	input.win.moveresize(DoX|DoY, bs+textPadding, bs+textPadding, 0, 0)
	bInp.moveresize(DoX|DoY|DoH,
		bs, bs+(2*textPadding)+input.win.geom.Height(), 0, bs)
	bTop.moveresize(DoX|DoY|DoH, 0, 0, 0, bs)
	bBot.moveresize(DoX|DoH, 0, 0, 0, bs)
	bLft.moveresize(DoX|DoY|DoW, 0, 0, bs, 0)
	bRht.moveresize(DoY|DoW, 0, 0, bs, 0)

	top.change(xgb.CWBackPixel, uint32(THEME.prompt.bgColor))
	bInp.change(xgb.CWBackPixel, uint32(THEME.prompt.borderColor))
	bTop.change(xgb.CWBackPixel, uint32(THEME.prompt.borderColor))
	bBot.change(xgb.CWBackPixel, uint32(THEME.prompt.borderColor))
	bLft.change(xgb.CWBackPixel, uint32(THEME.prompt.borderColor))
	bRht.change(xgb.CWBackPixel, uint32(THEME.prompt.borderColor))

	// actual mapping doesn't happen until top is mapped
	bInp.map_()
	bTop.map_()
	bBot.map_()
	bLft.map_()
	bRht.map_()
	input.win.map_()

	ps := &promptSelect{
		showing:      false,
		selected:     -1,
		listFun:      nil,
		prefixSearch: true,
		groups:       nil,
		itemsShowing: nil,
		bInp:         bInp,
		top:          top,
		input:        input,
		labVisible:   nil,
		labHidden:    nil,
		bTop:         bTop,
		bBot:         bBot,
		bLft:         bLft,
		bRht:         bRht,
	}
	ps.createWorkspaceLabels()

	// I love my xgbutil library. It provides a nice key binding interface
	// via the keybind package, but xevent still lets us handle raw key
	// press events :D
	xevent.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			if !ps.showing {
				return
			}

			beforeLen := len(ps.input.text)
			mods, kc := keybind.DeduceKeyInfo(ev.State, ev.Detail)

			s := keybind.LookupString(X, mods, kc)
			if len(s) == 1 {
				ps.input.add(rune(s[0]))
			}

			switch {
			case keyMatch(CONF.backspaceKey, mods, kc):
				ps.input.remove()
			case keyMatch(CONF.cancelKey, mods, kc):
				ps.hide()
				return
			case keyMatch(CONF.confirmKey, mods, kc):
				if ps.selected >= 0 && ps.selected < len(ps.itemsShowing) {
					ps.itemsShowing[ps.selected].action()
					ps.hide()
				} else if len(ps.itemsShowing) == 1 {
					ps.itemsShowing[0].action()
					ps.hide()
				}
				return
			case keyMatch(CONF.tabKey, mods, kc) ||
				keyMatch(CONF.revTabKey, mods, kc):

				if len(ps.itemsShowing) == 0 {
					break
				}
				if mods&xgb.ModMaskShift > 0 {
					if ps.selected == -1 {
						ps.selected++
					}
					ps.selected = mod(ps.selected-1, len(ps.itemsShowing))
				} else {
					ps.selected = mod(ps.selected+1, len(ps.itemsShowing))
				}
				ps.highlight()
			}

			// If the length of the input changed, then re-evaluate completion
			if beforeLen != len(ps.input.text) {
				ps.showItems(string(ps.input.text))
				ps.selected = -1
			}
		}).Connect(X, X.Dummy())

	return ps
}

// highlight highlights the item indicated by ps.selected, and unhighlights
// everything else. Note that ps.selected is an index into the items currently
// showing, not all possible items.
func (ps *promptSelect) highlight() {
	for i, item := range ps.itemsShowing {
		if i == ps.selected {
			item.inactive.unmap()
			item.active.map_()
		} else {
			item.active.unmap()
			item.inactive.map_()
		}
	}
}

// show, given a function the generates a list of groups of items and the
// kind of tab completion search to perform, will do the initialization steps
// of showing a selection prompt. Namely, initiating a grab and computing
// the geometry given the groups generated by listFun.
func (ps *promptSelect) show(listFun promptSelectListFun,
	prefixSearch bool) bool {

	// Note that DummyGrab is smart and avoids races. Check it out
	// in xgbutil/keybind.go if you're interested.
	// This makes it impossible to press and release alt-tab too quickly
	// to have it not register.
	if err := keybind.DummyGrab(X); err != nil {
		logWarning.Println("Could not grab keyboard for prompt select: %v", err)
		return false
	}

	// Set some config options
	ps.listFun = listFun
	ps.prefixSearch = prefixSearch

	// Some short aliases
	bs := THEME.prompt.borderSize
	padding := THEME.prompt.padding

	// Reset the input box and load all items
	ps.input.reset()
	ps.groups = ps.listFun()

	// Bail if there are no items to load
	if len(ps.groups) == 0 {
		ps.hide()
		return false
	}

	// get our screen geometry so we can position ourselves
	headGeom := WM.headActive()
	maxWidth := int(float64(headGeom.Width()) * 0.8)

	// Draw the items without restriction
	ps.showItems("")

	inputHeight := ps.input.win.geom.Height()
	height := 2*(bs+padding) + inputHeight + (2 * textPadding) + bs
	maxFontWidth := 0
	addedLabelSpacing := false
	for _, group := range ps.groups {
		if group.win != nil {
			maxFontWidth = max(maxFontWidth, group.win.geom.Width())
			height += group.win.geom.Height() + labelSpacing
			addedLabelSpacing = true
		}
		for _, item := range group.items {
			maxFontWidth = max(maxFontWidth, item.inactive.geom.Width())
			height += item.inactive.geom.Height()
		}
	}

	if addedLabelSpacing {
		height -= labelSpacing // we only want spacing for N - 1 groups
	}
	width := min(maxWidth, maxFontWidth+2*(bs+padding))

	// position the damn window based on its width/height (i.e., center it)
	posx := headGeom.Width()/2 - width/2
	posy := headGeom.Height()/2 - height/2

	// Issue the configure requests. We also need to adjust the borders.
	ps.top.moveresize(DoX|DoY|DoW|DoH, posx, posy, width, height)
	ps.bInp.moveresize(DoW, 0, 0, width, 0)
	ps.bTop.moveresize(DoW, 0, 0, width, 0)
	ps.bBot.moveresize(DoY|DoW, 0, height-bs, width, 0)
	ps.bLft.moveresize(DoH, 0, 0, 0, height)
	ps.bRht.moveresize(DoX|DoH, width-bs, 0, 0, height)

	// To the top!
	if len(WM.stack) > 0 {
		ps.top.configure(DoSibling|DoStack, 0, 0, 0, 0,
			WM.stack[0].Frame().ParentId(), xgb.StackModeAbove)
	}

	ps.showing = true
	ps.selected = -1
	ps.top.map_()

	return true
}

// showItems shows all items in ps.items with tab completion.
// It will also create "itemsShowing" so we can tab through visible items.
func (ps *promptSelect) showItems(search string) {
	// Some short aliases
	bs := THEME.prompt.borderSize
	padding := THEME.prompt.padding
	inputHeight := ps.input.win.geom.Height()

	// Initialize itemsShowing data
	ps.itemsShowing = make([]*promptSelectItem, 0)

	x, y := bs+padding, (2*bs)+padding+inputHeight+(2*textPadding)
	for _, group := range ps.groups {
		groupShown := false // true when at least 1 item in group is showing

		if group.win != nil {
			group.win.moveresize(DoX|DoY, x, y, 0, 0)
			y += group.win.geom.Height()
		}

		for _, item := range group.items {
			haystack := strings.ToLower(item.text)
			needle := strings.ToLower(search)
			if ps.prefixSearch && !strings.HasPrefix(haystack, needle) {
				item.inactive.unmap()
				item.active.unmap()
				continue
			}
			if !ps.prefixSearch && !strings.Contains(haystack, needle) {
				item.inactive.unmap()
				item.active.unmap()
				continue
			}

			item.active.moveresize(DoX|DoY, x, y, 0, 0)
			item.inactive.moveresize(DoX|DoY, x, y, 0, 0)
			item.inactive.map_()

			y += item.inactive.geom.Height()

			ps.itemsShowing = append(ps.itemsShowing, item)
			groupShown = true
		}

		// If we have a valid group label and it wasn't shown,
		// then subtract the addition we made to 'y' and make sure
		// the label is hidden. Otherwise, add some label spacing
		// and map the label.
		if group.win != nil {
			if !groupShown {
				y -= group.win.geom.Height()
				group.win.unmap()
			} else {
				group.win.map_()

				// space out the labels
				y += labelSpacing
			}
		}
	}
}

// hide stops the grab and hides the prompt.
func (ps *promptSelect) hide() {
	ps.top.unmap()
	keybind.DummyUngrab(X)
	ps.showing = false

	for _, group := range ps.groups {
		if group.win != nil {
			group.win.unmap()
		}
		for _, item := range group.items {
			item.active.unmap()
			item.inactive.unmap()
		}
	}
}

// promptSelectListWorkspaces generates two groups of all workspaces.
// The first group contains all visible workspaces.
// The second group contains the rest (hidden).
func promptSelectListWorkspaces(
	action func(*workspace) func()) []*promptSelectGroup {

	vWrks := make([]*promptSelectItem, 0, len(WM.heads))
	hWrks := make([]*promptSelectItem, 0, len(WM.workspaces)-len(WM.heads))

	for head := range WM.heads {
		wrk := WM.WrkHead(head)
		vWrks = append(vWrks,
			newPromptSelectItem(wrk.name, action(wrk),
				wrk.promptStore["select_active"],
				wrk.promptStore["select_inactive"]))
	}
	for _, wrk := range WM.workspaces {
		if !wrk.visible() {
			hWrks = append(hWrks,
				newPromptSelectItem(wrk.name, action(wrk),
					wrk.promptStore["select_active"],
					wrk.promptStore["select_inactive"]))
		}
	}
	return []*promptSelectGroup{
		newPromptSelectGroup(labelVisible, PROMPTS.slct.labVisible, vWrks),
		newPromptSelectGroup(labelHidden, PROMPTS.slct.labHidden, hWrks),
	}
}

// promptSelectAdd adds a workspace to the current prompt and sets up name
// images.
func (wrk *workspace) promptSelectAdd() {
	if PROMPTS.slct.showing {
		PROMPTS.slct.hide()
	}

	wrk.promptSelectUpdateName()
}

// promptSelectRemove removes a workspace from the current prompt.
func (wrk *workspace) promptSelectRemove() {
	if PROMPTS.slct.showing {
		PROMPTS.slct.hide()
	}

	if w, ok := wrk.promptStore["select_active"]; ok {
		w.unmap()
		w.destroy()
	}
	if w, ok := wrk.promptStore["select_inactive"]; ok {
		w.unmap()
		w.destroy()
	}
	if w, ok := wrk.promptStore["select_label"]; ok {
		w.unmap()
		w.destroy()
	}
}

// promptSelectUpdateName for workspaces updates the name in three different
// places: the header label for the client list prompts and the active/inactive
// labels for the workspace list prompt.
func (wrk *workspace) promptSelectUpdateName() {
	text := wrk.name

	aImg, aew, aeh, err := renderTextSolid(
		THEME.prompt.selectActiveBgColor, THEME.prompt.font,
		THEME.prompt.fontSize, THEME.prompt.selectActiveColor, text)
	if err != nil {
		return
	}

	iImg, iew, ieh, err := renderTextSolid(
		THEME.prompt.bgColor, THEME.prompt.font, THEME.prompt.fontSize,
		THEME.prompt.fontColor, text)
	if err != nil {
		return
	}

	labImg, lew, leh, err := renderTextSolid(
		THEME.prompt.bgColor, THEME.prompt.font,
		THEME.prompt.selectLabelFontSize, THEME.prompt.selectLabelColor, text)
	if err != nil {
		return
	}

	// For each image, either paint the new updated image
	// or create the image window necessary for display.

	if w, ok := wrk.promptStore["select_active"]; ok {
		xgraphics.PaintImg(X, w.id, aImg)
	} else {
		wrk.promptStore["select_active"] = createImageWindow(PROMPTS.slct.Id(),
			aImg, 0)
	}
	if w, ok := wrk.promptStore["select_inactive"]; ok {
		xgraphics.PaintImg(X, w.id, iImg)
	} else {
		wrk.promptStore["select_inactive"] = createImageWindow(
			PROMPTS.slct.Id(), iImg, 0)
	}
	if w, ok := wrk.promptStore["select_label"]; ok {
		xgraphics.PaintImg(X, w.id, labImg)
		w.moveresize(DoW|DoH, 0, 0, lew, leh)
	} else {
		wrk.promptStore["select_label"] = createImageWindow(PROMPTS.slct.Id(),
			labImg, 0)
	}

	// fit the text!
	wrk.promptStore["select_active"].moveresize(DoW|DoH, 0, 0, aew, aeh)
	wrk.promptStore["select_inactive"].moveresize(DoW|DoH, 0, 0, iew, ieh)
	wrk.promptStore["select_label"].moveresize(DoW|DoH, 0, 0, lew, leh)

	// Don't let text overlap borders.
	wrk.promptStore["select_active"].configure(
		DoSibling|DoStack, 0, 0, 0, 0,
		PROMPTS.slct.bRht.id, xgb.StackModeBelow)
	wrk.promptStore["select_inactive"].configure(
		DoSibling|DoStack, 0, 0, 0, 0,
		PROMPTS.slct.bRht.id, xgb.StackModeBelow)
	wrk.promptStore["select_label"].configure(
		DoSibling|DoStack, 0, 0, 0, 0,
		PROMPTS.slct.bRht.id, xgb.StackModeBelow)
}

// createWorkspaceLabels creates the "Visible" and "Hidden" labels used
// in the workspace list prompt.
func (ps *promptSelect) createWorkspaceLabels() {
	vimg, vew, veh, err := renderTextSolid(
		THEME.prompt.bgColor, THEME.prompt.font,
		THEME.prompt.selectLabelFontSize, THEME.prompt.selectLabelColor,
		labelVisible)
	if err != nil {
		return
	}

	himg, hew, heh, err := renderTextSolid(
		THEME.prompt.bgColor, THEME.prompt.font,
		THEME.prompt.selectLabelFontSize, THEME.prompt.selectLabelColor,
		labelHidden)
	if err != nil {
		return
	}

	ps.labVisible = createImageWindow(ps.Id(), vimg, 0)
	ps.labHidden = createImageWindow(ps.Id(), himg, 0)

	// fit the text!
	ps.labVisible.moveresize(DoW|DoH, 0, 0, vew, veh)
	ps.labHidden.moveresize(DoW|DoH, 0, 0, hew, heh)

	// Don't let text overlap borders.
	ps.labVisible.configure(
		DoSibling|DoStack, 0, 0, 0, 0,
		ps.bRht.id, xgb.StackModeBelow)
	ps.labHidden.configure(
		DoSibling|DoStack, 0, 0, 0, 0,
		ps.bRht.id, xgb.StackModeBelow)
}

// promptSelectListClients generates a list of clients grouped by workspace.
// They are filtered by the three booleans "activeWrk", "visible" and
// "iconified." The three booleans roughly correspond to the different
// client lists available.
func promptSelectListClients(activeWrk, visible,
	iconified bool) []*promptSelectGroup {

	addItems := func(wrk *workspace) *promptSelectGroup {
		items := make([]*promptSelectItem, 0)
		for i := len(WM.focus) - 1; i >= 0; i-- {
			c := WM.focus[i]

			if c.workspace != wrk.id {
				continue
			}

			_, actok := c.promptStore["select_active"]
			_, inactok := c.promptStore["select_inactive"]
			if !actok || !inactok {
				continue
			}

			w := WM.workspaces[c.workspace]
			if activeWrk && !w.active {
				continue
			}
			if visible && !w.visible() {
				continue
			}
			if !iconified && c.iconified {
				continue
			}

			focusRaise := func(c *client) func() {
				return func() {
					if c.iconified {
						c.IconifyToggle()
					}
					c.Focus()
					c.Raise()
				}
			}(c)
			items = append(items,
				newPromptSelectItem(c.Name(), focusRaise,
					c.promptStore["select_active"],
					c.promptStore["select_inactive"]))
		}
		if len(items) > 0 {
			return newPromptSelectGroup(wrk.name,
				wrk.promptStore["select_label"], items)
		}
		return nil
	}

	groups := make([]*promptSelectGroup, 0)

	// If we're only getting visible clients (i.e., clients on a monitor),
	// then lets order the groups by monitor.
	if visible {
		for head := range WM.heads {
			if newGroup := addItems(WM.WrkHead(head)); newGroup != nil {
				groups = append(groups, newGroup)
			}
		}
	} else {
		for _, wrk := range WM.workspaces {
			if newGroup := addItems(wrk); newGroup != nil {
				groups = append(groups, newGroup)
			}
		}
	}

	return groups
}

// promptSelectAdd adds the client to the prompt select dialog.
func (c *client) promptSelectAdd() {
	if PROMPTS.slct.showing {
		PROMPTS.slct.hide()
	}

	c.promptSelectUpdateName()
}

// promptSelectRemove removes the client from the prompt select dialog.
func (c *client) promptSelectRemove() {
	if PROMPTS.slct.showing {
		PROMPTS.slct.hide()
	}

	if w, ok := c.promptStore["select_active"]; ok {
		w.unmap()
		w.destroy()
	}
	if w, ok := c.promptStore["select_inactive"]; ok {
		w.unmap()
		w.destroy()
	}
}

// promptSelectUpdateName updates the name text images for a client.
func (c *client) promptSelectUpdateName() {
	text := c.Name()

	aImg, aew, aeh, err := renderTextSolid(
		THEME.prompt.selectActiveBgColor, THEME.prompt.font,
		THEME.prompt.fontSize, THEME.prompt.selectActiveColor, text)
	if err != nil {
		return
	}

	iImg, iew, ieh, err := renderTextSolid(
		THEME.prompt.bgColor, THEME.prompt.font, THEME.prompt.fontSize,
		THEME.prompt.fontColor, text)
	if err != nil {
		return
	}

	if w, ok := c.promptStore["select_active"]; ok {
		xgraphics.PaintImg(X, w.id, aImg)
	} else {
		c.promptStore["select_active"] = createImageWindow(PROMPTS.slct.Id(),
			aImg, 0)
	}

	if w, ok := c.promptStore["select_inactive"]; ok {
		xgraphics.PaintImg(X, w.id, iImg)
	} else {
		c.promptStore["select_inactive"] = createImageWindow(PROMPTS.slct.Id(),
			iImg, 0)
	}

	// fit the text!
	c.promptStore["select_active"].moveresize(DoW|DoH, 0, 0, aew, aeh)
	c.promptStore["select_inactive"].moveresize(DoW|DoH, 0, 0, iew, ieh)

	// Don't let text overlap borders.
	c.promptStore["select_active"].configure(
		DoSibling|DoStack, 0, 0, 0, 0,
		PROMPTS.slct.bRht.id, xgb.StackModeBelow)
	c.promptStore["select_inactive"].configure(
		DoSibling|DoStack, 0, 0, 0, 0,
		PROMPTS.slct.bRht.id, xgb.StackModeBelow)
}
