package main

import (
    "os"
    "os/exec"
)

import "code.google.com/p/jamslam-x-go-binding/xgb"

import (
    "github.com/BurntSushi/xgbutil"
    "github.com/BurntSushi/xgbutil/keybind"
    "github.com/BurntSushi/xgbutil/xevent"
    "github.com/BurntSushi/xgbutil/xwindow"
)

// global variables!
var X *xgbutil.XUtil
var WM *state

func quit() {
    logMessage.Println("The User has told us to quit.")
    os.Exit(0)
}

func konsole() {
    exec.Command("konsole").Start()
}

func main() {
    var err error
    X, err = xgbutil.Dial("")
    if err != nil {
        logError.Println(err)
        logError.Println("Error connecting to X, quitting...")
        return
    }
    defer X.Conn().Close()

    // Create WM state
    WM = newState()

    // Allow key bindings to do their thang
    keybind.Initialize(X)

    // Listen to Root. It is all-important.
    xwindow.Listen(X, X.RootWin(), xgb.EventMaskPropertyChange |
                                   xgb.EventMaskSubstructureNotify |
                                   xgb.EventMaskSubstructureRedirect)

    // Oblige map request events
    xevent.MapRequestFun(clientMapRequest).Connect(X, X.RootWin())

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

