package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"code.google.com/p/jamslam-x-go-binding/xgb"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/logger"
)

func Get(X *xgbutil.XUtil) (string, error) {
	return xprop.PropValStr(xprop.GetProperty(X, X.RootWin(), "_WINGO_CMD"))
}

func Set(X *xgbutil.XUtil, cmd string) error {
	return xprop.ChangeProp(X, X.RootWin(), 8,
		"_WINGO_CMD", "UTF8_STRING", []byte(cmd))
}

func StatusGet(X *xgbutil.XUtil) bool {
	status, err := xprop.PropValStr(xprop.GetProperty(X, X.RootWin(),
		"_WINGO_CMD_STATUS"))

	return err == nil && strings.ToLower(status) == "success"
}

func StatusSet(X *xgbutil.XUtil, status bool) {
	var statusStr string
	if status {
		statusStr = "Success"
	} else {
		statusStr = "Error"
	}

	// throw away the error
	xprop.ChangeProp(X, X.RootWin(), 8, "_WINGO_CMD_STATUS", "UTF8_STRING",
		[]byte(statusStr))
}

func main() {
	X, err := xgbutil.Dial("")
	if err != nil {
		logger.Error.Println(err)
		logger.Error.Println("Error connecting to X, quitting...")
		return
	}
	defer X.Conn().Close()

	// Get command from arguments
	cmdName := "PromptSelect"
	cmdFull := fmt.Sprintf("%s ClientsAll Prefix", cmdName)

	// make sure we start with failure
	StatusSet(X, false)
	success := false

	// Set the command before sending request to run command.
	err = Set(X, cmdFull)
	if err != nil {
		logger.Error.Printf("Could not set command: %s", err)
		return
	}

	// Issue the command!
	ewmh.ClientEvent(X, X.RootWin(), "_WINGO_CMD")

	// Now let's set up a handler to detect when the status changes
	xevent.PropertyNotifyFun(
		func(X *xgbutil.XUtil, ev xevent.PropertyNotifyEvent) {
			name, err := xprop.AtomName(X, ev.Atom)
			if err != nil {
				logger.Warning.Println(
					"Could not get property atom name for", ev.Atom)
				return
			}

			if name == "_WINGO_CMD_STATUS" {
				success = StatusGet(X)
				if success {
					os.Exit(0)
				} else {
					logger.Warning.Printf("Error running '%s'", cmdFull)
					ShowUsage(cmdName)
					os.Exit(1)
				}
			}
		}).Connect(X, X.RootWin())

	// Listen to Root property change events
	xwindow.Listen(X, X.RootWin(), xgb.EventMaskPropertyChange)

	go xevent.Main(X)

	time.Sleep(time.Second * 5)

	logger.Error.Println(
		"Timed out while trying to issue command to Wingo. " +
		"Are you sure Wingo is running?")
	os.Exit(1)
}

