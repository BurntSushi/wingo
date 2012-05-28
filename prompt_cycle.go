package main

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xgraphics"

	"github.com/BurntSushi/wingo/logger"
)

type promptCycle struct {
	showing                bool
	selected               int
	grabbedMods            uint16
	clients                []*client
	fontHeight             int
	top                    *window
	bTop, bBot, bLft, bRht *window
	iconBorder             *window
}

func (pc *promptCycle) Id() xproto.Window {
	return pc.top.id
}

// promptCycleInitialize sets up the cycle prompt window.
// Note that this does not map it, it just creates all the resources so that
// mapping can be quick.
// XXX: Should also create resources for each already existing client by
// using promptCycleAdd.
func newPromptCycle() *promptCycle {
	top := createWindow(ROOT.id, 0)
	bTop, bBot := createWindow(top.id, 0), createWindow(top.id, 0)
	bLft, bRht := createWindow(top.id, 0), createWindow(top.id, 0)

	// Apply as much theme related crud as we can
	// We don't do width/height until we actually show the prompt.
	bs := THEME.prompt.borderSize
	bTop.moveresize(DoX|DoY|DoH, 0, 0, 0, bs)
	bBot.moveresize(DoX|DoH, 0, 0, 0, bs)
	bLft.moveresize(DoX|DoY|DoW, 0, 0, bs, 0)
	bRht.moveresize(DoY|DoW, 0, 0, bs, 0)

	top.change(xproto.CwBackPixel, uint32(THEME.prompt.bgColor))
	bTop.change(xproto.CwBackPixel, uint32(THEME.prompt.borderColor))
	bBot.change(xproto.CwBackPixel, uint32(THEME.prompt.borderColor))
	bLft.change(xproto.CwBackPixel, uint32(THEME.prompt.borderColor))
	bRht.change(xproto.CwBackPixel, uint32(THEME.prompt.borderColor))

	// actual mapping doesn't happen until top is mapped
	bTop.map_()
	bBot.map_()
	bLft.map_()
	bRht.map_()

	pc := &promptCycle{
		showing:     false,
		selected:    0,
		grabbedMods: 0,
		clients:     make([]*client, 0),
		top:         top,
		bTop:        bTop,
		bBot:        bBot,
		bLft:        bLft,
		bRht:        bRht,
	}

	// This bears explanation.
	// The prompt cycle dialog needs to choose the selection when the 
	// modifiers (i.e., "alt" in "alt-tab") are released.
	// The only way to do this (generally) is to check the raw KeyRelease event.
	// Namely, if the keycode *released* is a modifier, we have to and-out
	// that modifier from the key release event data. If the modifiers
	// remaining aren't up to snuff with the original grabbed modifiers,
	// then we can finally "choose" the selection.
	// TL;DR - This is how we "exit" the prompt cycle dialog.
	xevent.KeyReleaseFun(
		func(X *xgbutil.XUtil, ev xevent.KeyReleaseEvent) {
			if !pc.showing {
				return
			}

			mods, kc := keybind.DeduceKeyInfo(ev.State, ev.Detail)

			// Allow user to quit without effect
			if keyMatch(CONF.cancelKey, mods, kc) {
				pc.hide()
				return
			}

			mods &= ^keybind.ModGet(X, ev.Detail)
			if pc.grabbedMods > 0 && mods&pc.grabbedMods == 0 {
				pc.choose()
			}
		}).Connect(X, X.Dummy())

	return pc
}

func (pc *promptCycle) next(keyStr string, activeWrk, visible, iconified bool) {
	if !pc.showing && !pc.show(keyStr, activeWrk, visible, iconified) {
		return
	}

	// First time?
	if pc.selected == -1 {
		if len(pc.clients) > 1 {
			pc.selected = 1
		} else {
			pc.selected = 0
		}
	} else {
		pc.selected++
	}

	// Everybody do the wrap-around...
	pc.selected = mod(pc.selected, len(pc.clients))
	pc.highlight()
}

func (pc *promptCycle) prev(keyStr string, activeWrk, visible, iconified bool) {
	if !pc.showing && !pc.show(keyStr, activeWrk, visible, iconified) {
		return
	}

	// First time?
	if pc.selected == -1 {
		pc.selected = len(pc.clients) - 1
	} else {
		pc.selected--
	}

	// Everybody do the wrap-around...
	pc.selected = mod(pc.selected, len(pc.clients))
	pc.highlight()
}

func (pc *promptCycle) highlight() {
	for i, c := range pc.clients {
		iconPar, iok := c.promptStore["cycle_border"]
		winTitle, tok := c.promptStore["cycle_title"]
		if !iok || !tok {
			continue
		}

		if i == pc.selected {
			iconPar.change(xproto.CwBackPixel, uint32(THEME.prompt.borderColor))
			winTitle.map_()
		} else {
			iconPar.change(xproto.CwBackPixel, uint32(THEME.prompt.bgColor))
			winTitle.unmap()
		}
		iconPar.clear()
	}
}

func (pc *promptCycle) choose() {
	if pc.selected >= 0 && pc.selected < len(pc.clients) {
		c := pc.clients[pc.selected]
		if c.iconified {
			c.IconifyToggle()
		}

		c.Focus()
		c.Raise()
	}
	pc.hide()
}

func (pc *promptCycle) show(keyStr string, activeWrk, visible,
	iconified bool) bool {

	// Note that DummyGrab is smart and avoids races. Check it out
	// in xgbutil/keybind.go if you're interested.
	// This makes it impossible to press and release alt-tab too quickly
	// to have it not register.
	if err := keybind.DummyGrab(X); err != nil {
		logger.Warning.Println("Could not grab keyboard for prompt cycle: %v",
			err)
		return false
	}

	// save the modifiers used to initially start this prompt
	pc.grabbedMods, _, _ = keybind.ParseString(X, keyStr)
	bs := THEME.prompt.borderSize
	cbs := THEME.prompt.cycleIconBorderSize
	is := THEME.prompt.cycleIconSize
	padding := THEME.prompt.padding

	// To the top!
	if len(WM.stack) > 0 {
		pc.top.configure(DoStack, 0, 0, 0, 0, 0, xproto.StackModeAbove)
	}

	// get our screen geometry so we can position ourselves
	headGeom := WM.headActive()
	maxWidth := int(float64(headGeom.Width()) * 0.8)

	// x,y correspond to the position of the next window icon.
	// They are updated after mapping each icon to refer to the position
	// of the next icon.
	x, y := bs+padding, bs+padding+cbs+pc.fontHeight

	// width and height correspond to the final width and height of the
	// prompt window. They are updated as each icon is added.
	// width is initialized to account for border size and padding.
	width := 2 * (bs + padding)
	height := (2 * (bs + padding + cbs)) + is + cbs + pc.fontHeight

	// maxFontWidth represents the largest font window.
	// We can use it to sometimes make the prompt bigger to fit window titles.
	maxFontWidth := 0

	// Maintain a list of clients that we show in the dialog.
	pc.clients = []*client{}

	bail := true         // if there's nothing to show, we bail...
	widthStatic := false // when true, we stop increasing width

	var c *client
	for i := len(WM.focus) - 1; i >= 0; i-- {
		c = WM.focus[i]
		if activeWrk && !c.workspace.active {
			continue
		}
		if visible && !c.workspace.visible() {
			continue
		}
		if !iconified && c.iconified {
			continue
		}

		winPar, parok := c.promptStore["cycle_border"]
		winAct, actok := c.promptStore["cycle_act"]
		winInact, inactok := c.promptStore["cycle_inact"]
		winTit, titok := c.promptStore["cycle_title"]
		if !parok || !actok || !inactok || !titok {
			continue
		}

		// We no longer have to bail, since we've found a valid client to show
		bail = false

		// track largest font window
		maxFontWidth = max(maxFontWidth, winTit.geom.Width())

		// Move on to the next row?
		if x+(is+(2*cbs))+padding+bs > maxWidth {
			x = bs + padding
			y += is + (2 * cbs)
			height += is + (2 * cbs)
			widthStatic = true
		}

		// Position the icon window and map its active version or its
		// inactive version if it's iconified.
		winPar.moveresize(DoX|DoY, x, y, 0, 0)
		winPar.map_()
		if c.iconified {
			winInact.map_()
			winAct.unmap()
		} else {
			winAct.map_()
			winInact.unmap()
		}

		// Only increase the width if we're still adding icons to the first row.
		if !widthStatic {
			width += is + (2 * cbs)
		}
		x += is + (2 * cbs)
		pc.clients = append(pc.clients, c)
	}

	if bail {
		pc.hide()
		return false
	}

	// If the computed width is less than the max font width, then increase
	// the width of the prompt to fit the longest window title.
	// Forcefully cap it as the maxWidth, though.
	if maxFontWidth+2*(padding+bs) > width {
		width = min(maxWidth, maxFontWidth+2*(padding+bs))
	}

	// position the damn window based on its width/height (i.e., center it)
	posx := headGeom.X() + headGeom.Width()/2 - width/2
	posy := headGeom.Y() + headGeom.Height()/2 - height/2

	// Issue the configure requests. We also need to adjust the borders.
	pc.top.moveresize(DoX|DoY|DoW|DoH, posx, posy, width, height)
	pc.bTop.moveresize(DoW, 0, 0, width, 0)
	pc.bBot.moveresize(DoY|DoW, 0, height-bs, width, 0)
	pc.bLft.moveresize(DoH, 0, 0, 0, height)
	pc.bRht.moveresize(DoX|DoH, width-bs, 0, 0, height)

	pc.showing = true
	pc.selected = -1
	pc.top.map_()

	return true
}

// hide stops the grab and hides the prompt.
func (pc *promptCycle) hide() {
	pc.top.unmap()
	keybind.DummyUngrab(X)
	pc.showing = false

	for i := len(WM.focus) - 1; i >= 0; i-- {
		c := WM.focus[i]

		winPar, parok := c.promptStore["cycle_border"]
		winAct, actok := c.promptStore["cycle_act"]
		winInact, inactok := c.promptStore["cycle_inact"]
		winTit, titok := c.promptStore["cycle_title"]
		if !parok || !actok || !inactok || !titok {
			continue
		}

		winPar.unmap()
		winAct.unmap()
		winInact.unmap()
		winTit.unmap()
	}
}

// promptCycleAdd adds the client to the prompt cycle dialog.
// Currently, it loads the window icon and the window name.
func (c *client) promptCycleAdd() {
	if PROMPTS.cycle.showing {
		PROMPTS.cycle.hide()
	}

	c.promptStore["cycle_border"] = createWindow(
		PROMPTS.cycle.Id(), xproto.CwBackPixel, uint32(THEME.prompt.bgColor))
	c.promptStore["cycle_border"].moveresize(
		DoW|DoH, 0, 0,
		THEME.prompt.cycleIconSize+2*THEME.prompt.cycleIconBorderSize,
		THEME.prompt.cycleIconSize+2*THEME.prompt.cycleIconBorderSize)

	c.promptCycleUpdateIcon()
	c.promptCycleUpdateName()
}

// promptCycleRemove removes the client from the cycle prompt.
// Basically, we have to destroy all the resources we've allocated.
// (Note that we don't need to free any Pixmaps, since we free those right
// after we draw them.)
func (c *client) promptCycleRemove() {
	if PROMPTS.cycle.showing {
		PROMPTS.cycle.hide()
	}

	if w, ok := c.promptStore["cycle_title"]; ok {
		w.unmap()
		w.destroy()
	}
	if w, ok := c.promptStore["cycle_act"]; ok {
		w.unmap()
		w.destroy()
	}
	if w, ok := c.promptStore["cycle_inact"]; ok {
		w.unmap()
		w.destroy()
	}
	if w, ok := c.promptStore["cycle_border"]; ok {
		w.unmap()
		w.destroy()
	}
}

func (c *client) promptCycleUpdateIcon() {
	par, ok := c.promptStore["cycle_border"]
	if !ok {
		logger.Warning.Printf(
			"BUG: The 'cycle_border' parent window hasn't been " +
				"created yet, and we think it should have been. " +
				"This client: '%s' will probably not show up in " +
				"the cycle prompt.")
		return
	}

	bgc := colorFromInt(THEME.prompt.bgColor)
	iconSize := THEME.prompt.cycleIconSize
	cbs := THEME.prompt.cycleIconBorderSize
	alpha := THEME.prompt.cycleIconTransparency // value checked at startup

	img := c.iconImage(iconSize, iconSize)

	imgAct := xgraphics.BlendBg(img, nil, 100, bgc)
	imgInact := xgraphics.BlendBg(img, nil, alpha, bgc)

	if w, ok := c.promptStore["cycle_act"]; ok {
		xgraphics.PaintImg(X, w.id, imgAct)
	} else {
		c.promptStore["cycle_act"] = createImageWindow(par.id, imgAct, 0)
		c.promptStore["cycle_act"].moveresize(DoX|DoY, cbs, cbs, 0, 0)
	}

	if w, ok := c.promptStore["cycle_inact"]; ok {
		xgraphics.PaintImg(X, w.id, imgInact)
	} else {
		c.promptStore["cycle_inact"] = createImageWindow(par.id, imgInact, 0)
		c.promptStore["cycle_inact"].moveresize(DoX|DoY, cbs, cbs, 0, 0)
	}
}

func (c *client) promptCycleUpdateName() {
	text := c.Name()

	textImg, ew, eh, err := renderTextSolid(
		THEME.prompt.bgColor, THEME.prompt.font, THEME.prompt.fontSize,
		THEME.prompt.fontColor, text)
	if err != nil {
		return
	}

	if w, ok := c.promptStore["cycle_title"]; ok {
		xgraphics.PaintImg(X, w.id, textImg)
	} else {
		c.promptStore["cycle_title"] = createImageWindow(PROMPTS.cycle.Id(),
			textImg, 0)
	}

	bs := THEME.prompt.borderSize
	padding := THEME.prompt.padding
	c.promptStore["cycle_title"].moveresize(DoX|DoY|DoW|DoH,
		bs+padding, bs+padding, ew, eh)
	c.promptStore["cycle_title"].configure(
		DoSibling|DoStack, 0, 0, 0, 0,
		PROMPTS.cycle.bRht.id, xproto.StackModeBelow)

	// Set the largest font size we've seen.
	PROMPTS.cycle.fontHeight = max(PROMPTS.cycle.fontHeight, eh)
}
