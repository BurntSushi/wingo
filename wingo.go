package main

import (
    "log"
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

// state is the master singleton the carries all window manager related state
type state struct {
    clients []*client
}

// global variables!
var X *xgbutil.XUtil
var WM *state

func quit() {
    log.Println("The User has told us to quit.")
    os.Exit(0)
}

func konsole() {
    print("Konsole time\n")
    exec.Command("konsole").Start()
}

func main() {
    var err error
    X, err = xgbutil.Dial(":10")
    if err != nil {
        log.Println(err)
        log.Println("Error connecting to X, quitting...")
        return
    }
    defer X.Conn().Close()

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

    xevent.Main(X)
}

