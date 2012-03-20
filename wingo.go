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
    "github.com/BurntSushi/xgbutil/xwindow"
)

// global variables!
var X *xgbutil.XUtil
var WM *state

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

    // Allow key and mouse bindings to do their thang
    keybind.Initialize(X)
    mousebind.Initialize(X)

    // Setup some cursors we use
    setupCursors()

    // Listen to Root. It is all-important.
    xwindow.Listen(X, X.RootWin(), xgb.EventMaskPropertyChange |
                                   xgb.EventMaskSubstructureNotify |
                                   xgb.EventMaskSubstructureRedirect)

    // Oblige map request events
    xevent.MapRequestFun(clientMapRequest).Connect(X, X.RootWin())

    // Oblige configure requests from windows we don't manage.
    xevent.ConfigureRequestFun(configureRequest).Connect(X, X.RootWin())

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
            cmd_close_active()
    }).Connect(X, X.RootWin(), "Mod4-c")

    xevent.Main(X)
}
