package main

import (
    "log"
    "os"
    "runtime/pprof"
)

import "burntsushi.net/go/x-go-binding/xgb"

import (
    "burntsushi.net/go/xgbutil"
    "burntsushi.net/go/xgbutil/ewmh"
    "burntsushi.net/go/xgbutil/keybind"
    "burntsushi.net/go/xgbutil/mousebind"
    "burntsushi.net/go/xgbutil/xevent"
)

// global variables!
var X *xgbutil.XUtil
var WM *state
var ROOT *window
var CONF *conf
var THEME *theme
var PROMPTS prompts

func main() {
    var err error

    f, err := os.Create("zzz.prof")
    if err != nil {
        log.Fatal(err)
    }
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()

    X, err = xgbutil.Dial("")
    if err != nil {
        logError.Println(err)
        logError.Println("Error connecting to X, quitting...")
        return
    }
    defer X.Conn().Close()

    // Create a root window abstraction and load its geometry
    ROOT = newWindow(X.RootWin())
    _, err = ROOT.geometry()
    if err != nil {
        logError.Println("Could not get ROOT window geometry because: %v", err)
        logError.Println("Cannot continue. Quitting...")
        return
    }

    // Load configuration
    err = loadConfig()
    if err != nil {
        logError.Println(err)
        logError.Println("No configuration found. Quitting...")
        return
    }

    // Load theme
    err = loadTheme()
    if err != nil {
        logError.Println(err)
        logError.Println("No theme configuration found. Quitting...")
        return
    }

    // Create WM state
    WM = newState()
    WM.headsLoad()

    // Set supported atoms
    ewmh.SupportedSet(X, []string{"_NET_WM_ICON"})

    // Allow key and mouse bindings to do their thang
    keybind.Initialize(X)
    mousebind.Initialize(X)

    // Attach all global key bindings
    attachAllKeys()

    // Attach all root mouse bindings
    rootMouseConfig()

    // Setup some cursors we use
    setupCursors()

    // Initialize prompts
    promptsInitialize()

    // Listen to Root. It is all-important.
    ROOT.listen(xgb.EventMaskPropertyChange |
                xgb.EventMaskStructureNotify |
                xgb.EventMaskSubstructureNotify |
                xgb.EventMaskSubstructureRedirect)

    // Update state when the root window changes size
    xevent.ConfigureNotifyFun(rootGeometryChange).Connect(X, ROOT.id)

    // Oblige map request events
    xevent.MapRequestFun(clientMapRequest).Connect(X, ROOT.id)

    // Oblige configure requests from windows we don't manage.
    xevent.ConfigureRequestFun(configureRequest).Connect(X, ROOT.id)

    xevent.Main(X)

    println("Writing memory profile...")
    f, err = os.Create("zzz.mprof")
    if err != nil {
        log.Fatal(err)
    }
    pprof.WriteHeapProfile(f)
    f.Close()
}

