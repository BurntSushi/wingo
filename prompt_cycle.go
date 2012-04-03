package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

import (
    "github.com/BurntSushi/xgbutil"
    "github.com/BurntSushi/xgbutil/keybind"
    "github.com/BurntSushi/xgbutil/xevent"
    "github.com/BurntSushi/xgbutil/xgraphics"
)

type promptCycle struct {
    showing bool
    selected int
    grabbedMods uint16
    top *window
    inner *window
}

func (pc *promptCycle) Id() xgb.Id {
    return pc.top.id
}

// promptCycleInitialize sets up the cycle prompt window.
// Note that this does not map it, it just creates all the resources so that
// mapping can be quick.
// XXX: Should also create resources for each already existing client by
// using promptCylceAdd.
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
        top: top,
        inner: inner,
    }

    // This bears explanation.
    // The prompt cycle dialog needs to choose the selection when the 
    // modifiers (i.e., "alt" in "alt-tab") are released.
    // The only way to do this is to check the raw KeyRelease event.
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
    if !pc.showing {
        pc.show(keyStr)
    }

    pc.selected++
    println("next ", pc.selected)
}

func (pc *promptCycle) prev(keyStr string) {
    if !pc.showing {
        pc.show(keyStr)
    }

    pc.selected--
    println("prev ", pc.selected)
}

func (pc *promptCycle) choose() {
    println("Choosing ", pc.selected)
    pc.hide()
}

func (pc *promptCycle) show(keyStr string) {
    // Note that DummyGrab is smart and avoids races. Check it out
    // in xgbutil/keybind.go if you're interested.
    if err := keybind.DummyGrab(X); err != nil {
        logWarning.Println("Could not grab keyboard for prompt cycle: %v", err)
        return
    }

    // save the modifiers used to initially start this prompt
    pc.grabbedMods, _ = keybind.ParseString(X, keyStr)
    bs := THEME.prompt.borderSize

    pc.top.moveresize(DoW | DoH, 0, 0, 500, 200)
    pc.inner.moveresize(DoW | DoH, 0, 0, 500 - 2 * bs, 200 - 2 * bs)

    // To the top!
    if len(WM.stack) > 0 {
        pc.top.configure(DoSibling | DoStack, 0, 0, 0, 0,
                         WM.stack[0].Frame().ParentId(), xgb.StackModeAbove)
    }

    pc.showing = true
    pc.selected = -1
    pc.top.map_()
}

func (pc *promptCycle) hide() {
    pc.top.unmap()
    keybind.DummyUngrab(X)
    pc.showing = false
}

// promptCycleAdd adds two images to the promptStore:
// the active and inactive images in the dialog.
func (c *client) promptCycleAdd() {
    bgc := ColorFromInt(THEME.prompt.bgColor)
    iconSize := THEME.prompt.cycleIconSize
    alpha := THEME.prompt.cycleIconTransparency // value checked at startup

    img, mask := c.iconImage(iconSize, iconSize)

    imgAct := xgraphics.BlendBg(img, mask, 100, bgc)
    imgInact := xgraphics.BlendBg(img, mask, alpha, bgc)

    pid := PROMPTS.cycle.Id()
    c.promptStore["cycle_act"] = createImageWindow(pid, imgAct, 0)
    c.promptStore["cycle_inact"] = createImageWindow(pid, imgInact, 0)
}

