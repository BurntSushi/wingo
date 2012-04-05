package main

import "burntsushi.net/go/x-go-binding/xgb"

import (
    "burntsushi.net/go/xgbutil"
    "burntsushi.net/go/xgbutil/keybind"
    "burntsushi.net/go/xgbutil/xevent"
    "burntsushi.net/go/xgbutil/xgraphics"
)

type promptCycle struct {
    showing bool
    selected int
    grabbedMods uint16
    clients []*client
    top *window
    inner *window
    iconBorder *window
}

func (pc *promptCycle) Id() xgb.Id {
    return pc.top.id
}

// promptCycleInitialize sets up the cycle prompt window.
// Note that this does not map it, it just creates all the resources so that
// mapping can be quick.
// XXX: Should also create resources for each already existing client by
// using promptCycleAdd.
func newPromptCycle() *promptCycle {
    top := createWindow(ROOT.id, 0)
    inner := createWindow(top.id, 0)

    // Apply as much theme related crud as we can
    // We don't do width/height until we actually show the prompt.
    bs := THEME.prompt.borderSize
    inner.moveresize(DoX | DoY, bs, bs, 0, 0)

    top.change(xgb.CWBackPixel, uint32(THEME.prompt.borderColor))
    inner.change(xgb.CWBackPixel, uint32(THEME.prompt.bgColor))

    // actual mapping doesn't happen until top is mapped
    inner.map_()

    pc := &promptCycle{
        showing: false,
        selected: 0,
        grabbedMods: 0,
        clients: make([]*client, 0),
        top: top,
        inner: inner,
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

            mods, _ := keybind.DeduceKeyInfo(ev.State, ev.Detail)
            mods &= ^keybind.ModGet(X, ev.Detail)

            if pc.grabbedMods > 0 && mods & pc.grabbedMods == 0 {
                pc.choose()
            }
    }).Connect(X, X.Dummy())

    return pc
}

func (pc *promptCycle) next(keyStr string) {
    if !pc.showing && !pc.show(keyStr) {
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

func (pc *promptCycle) prev(keyStr string) {
    if !pc.showing && !pc.show(keyStr) {
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
        iconPar, ok := c.promptStore["cycle_border"]
        if !ok {
            continue
        }

        if i == pc.selected {
            iconPar.change(xgb.CWBackPixel, uint32(THEME.prompt.borderColor))
        } else {
            iconPar.change(xgb.CWBackPixel, uint32(THEME.prompt.bgColor))
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

func (pc *promptCycle) show(keyStr string) bool {
    // Note that DummyGrab is smart and avoids races. Check it out
    // in xgbutil/keybind.go if you're interested.
    // This makes it impossible to press and release alt-tab too quickly
    // to have it not register.
    if err := keybind.DummyGrab(X); err != nil {
        logWarning.Println("Could not grab keyboard for prompt cycle: %v", err)
        return false
    }

    // save the modifiers used to initially start this prompt
    pc.grabbedMods, _ = keybind.ParseString(X, keyStr)
    bs := THEME.prompt.borderSize
    cbs := THEME.prompt.cycleIconBorderSize
    is := THEME.prompt.cycleIconSize
    padding := 10

    // To the top!
    if len(WM.stack) > 0 {
        pc.top.configure(DoSibling | DoStack, 0, 0, 0, 0,
                         WM.stack[0].Frame().ParentId(), xgb.StackModeAbove)
    }

    // get our screen geometry so we can position ourselves
    headGeom := WM.headActive()
    maxWidth := int(float64(headGeom.Width()) * 0.8)

    // Now let's map and position all of the icons for each window
    x, y := bs + padding + cbs, bs + padding + cbs
    width, height := 2 * x, (2 * y) + is + cbs
    pc.clients = []*client{}
    bail := true // if there's nothing to show, we bail...
    widthStatic := false // when true, we stop increasing width
    var c *client
    for i := len(WM.focus) - 1; i >= 0; i-- {
        c = WM.focus[i]
        winPar, parok := c.promptStore["cycle_border"]
        winAct, actok := c.promptStore["cycle_act"]
        winInact, inactok := c.promptStore["cycle_inact"]
        if !parok || !actok || !inactok {
            continue
        }

        bail = false

        // Move on to the next row?
        if x + (is + (2 * cbs)) + padding + bs > maxWidth {
            x = bs + padding + cbs
            y += is + (2 * cbs)
            height += is + (2 * cbs)
            widthStatic = true
        }

        winPar.moveresize(DoX | DoY, x, y, 0, 0)
        if c.iconified {
            winInact.map_()
            winAct.unmap()
        } else {
            winAct.map_()
            winInact.unmap()
        }

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

    // position the damn window based on its width/height
    posx := headGeom.Width() / 2 - width / 2
    posy := headGeom.Height() / 2 - height / 2

    pc.top.moveresize(DoX | DoY | DoW | DoH, posx, posy, width, height)
    pc.inner.moveresize(DoW | DoH, 0, 0, width - 2 * bs, height - 2 * bs)

    pc.showing = true
    pc.selected = -1
    pc.top.map_()

    return true
}

func (pc *promptCycle) hide() {
    pc.top.unmap()
    keybind.DummyUngrab(X)
    pc.showing = false
}

// promptCycleAdd adds two images to the promptStore:
// the active and inactive images in the dialog.
func (c *client) promptCycleAdd() {
    if PROMPTS.cycle.showing {
        PROMPTS.cycle.hide()
    }

    c.promptStore["cycle_border"] = createWindow(
        PROMPTS.cycle.Id(), xgb.CWBackPixel, uint32(THEME.prompt.bgColor))
    c.promptStore["cycle_border"].moveresize(
        DoW | DoH, 0, 0,
        THEME.prompt.cycleIconSize + 2 * THEME.prompt.cycleIconBorderSize,
        THEME.prompt.cycleIconSize + 2 * THEME.prompt.cycleIconBorderSize)
    c.promptStore["cycle_border"].map_()

    c.promptCycleUpdateIcon()
}

func (c *client) promptCycleDestroy() {
    if PROMPTS.cycle.showing {
        PROMPTS.cycle.hide()
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
        logWarning.Printf("BUG: The 'cycle_border' parent window hasn't been " +
                          "created yet, and we think it should have been. " +
                          "This client: '%s' will probably not show up in " +
                          "the cycle prompt.")
        return
    }

    bgc := ColorFromInt(THEME.prompt.bgColor)
    iconSize := THEME.prompt.cycleIconSize
    cbs := THEME.prompt.cycleIconBorderSize
    alpha := THEME.prompt.cycleIconTransparency // value checked at startup

    img, mask := c.iconImage(iconSize, iconSize)

    imgAct := xgraphics.BlendBg(img, mask, 100, bgc)
    imgInact := xgraphics.BlendBg(img, mask, alpha, bgc)

    if w, ok := c.promptStore["cycle_act"]; ok {
        xgraphics.PaintImg(X, w.id, imgAct)
    } else {
        c.promptStore["cycle_act"] = createImageWindow(par.id, imgAct, 0)
        c.promptStore["cycle_act"].moveresize(DoX | DoY, cbs, cbs, 0, 0)
    }

    if w, ok := c.promptStore["cycle_inact"]; ok {
        xgraphics.PaintImg(X, w.id, imgInact)
    } else {
        c.promptStore["cycle_inact"] = createImageWindow(par.id, imgInact, 0)
        c.promptStore["cycle_inact"].moveresize(DoX | DoY, cbs, cbs, 0, 0)
    }
}

func (c *client) promptCycleUpdateName() {
}

