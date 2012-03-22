package main

import (
    // "log" 
    // "os" 
    "os/exec"
    // "runtime/pprof" 
)

import "code.google.com/p/jamslam-x-go-binding/xgb"

import (
    "github.com/BurntSushi/xgbutil"
    "github.com/BurntSushi/xgbutil/keybind"
    "github.com/BurntSushi/xgbutil/mousebind"
    "github.com/BurntSushi/xgbutil/xevent"
)

// global variables!
var X *xgbutil.XUtil
var WM *state
var ROOT *window

func quit() {
    logMessage.Println("The User has told us to quit.")
    X.Quit()
}

func konsole() {
    exec.Command("konsole").Start()
}

func main() {
    var err error

    // f, err := os.Create("zzz.prof") 
    // if err != nil { 
        // log.Fatal(err) 
    // } 
    // pprof.StartCPUProfile(f) 
    // defer pprof.StopCPUProfile() 

    X, err = xgbutil.Dial("")
    if err != nil {
        logError.Println(err)
        logError.Println("Error connecting to X, quitting...")
        return
    }
    defer X.Conn().Close()

    // Create WM state
    WM = newState()

    // Create a root window abstraction and load its geometry
    ROOT = newWindow(X.RootWin())
    _, err = ROOT.geometry()
    if err != nil {
        logError.Println("Could not get ROOT window geometry because: %v", err)
        logError.Println("Cannot continue. Quitting...")
        return
    }

    // Allow key and mouse bindings to do their thang
    keybind.Initialize(X)
    mousebind.Initialize(X)

    // Setup some cursors we use
    setupCursors()

    // Listen to Root. It is all-important.
    ROOT.listen(xgb.EventMaskPropertyChange |
                xgb.EventMaskSubstructureNotify |
                xgb.EventMaskSubstructureRedirect |
                xgb.EventMaskButtonPress)

    // Oblige map request events
    xevent.MapRequestFun(clientMapRequest).Connect(X, X.RootWin())

    // Oblige configure requests from windows we don't manage.
    xevent.ConfigureRequestFun(configureRequest).Connect(X, X.RootWin())

    mousebind.ButtonPressFun(
        func(X *xgbutil.XUtil, ev xevent.ButtonPressEvent) {
            ROOT.focus()
            WM.unfocusExcept(0)
    }).Connect(X, ROOT.id, "1", false, false)

    keybind.KeyPressFun(
        func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
            quit()
    }).Connect(X, X.RootWin(), "Mod1-Shift-c")

    keybind.KeyPressFun(
        func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
            konsole()
    }).Connect(X, X.RootWin(), "Mod4-j")

    keybind.KeyPressFun(
        func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
            WM.logClientList()
    }).Connect(X, X.RootWin(), "Mod4-l")

    keybind.KeyPressFun(
        func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
            cmd_active_close()
    }).Connect(X, X.RootWin(), "Mod4-c")

    keybind.KeyPressFun(
        func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
            cmd_active_frame_nada()
    }).Connect(X, X.RootWin(), "Mod1-1")
    keybind.KeyPressFun(
        func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
            cmd_active_frame_slim()
    }).Connect(X, X.RootWin(), "Mod1-2")
    keybind.KeyPressFun(
        func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
            cmd_active_frame_borders()
    }).Connect(X, X.RootWin(), "Mod1-3")
    keybind.KeyPressFun(
        func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
            cmd_active_frame_full()
    }).Connect(X, X.RootWin(), "Mod1-4")

    keybind.KeyPressFun(
        func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
            cmd_active_flash()
    }).Connect(X, X.RootWin(), "Mod4-f")

    keybind.KeyPressFun(
        func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
            cmd_active_test1()
    }).Connect(X, X.RootWin(), "Mod4-1")

    xevent.Main(X)
}

