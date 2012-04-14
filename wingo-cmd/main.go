package main

/*
	wingo-cmd is a small program that allows one to issue any command (except
	for mouse commands) to Wingo.

	It works by using ClientMessage events and the _WINGO_CMD and
	_WINGO_CMD_STATUS properties on the root window. Note that this approach
	implies that it is NOT a good idea to run two separate instances of
	wingo-cmd at the same time.

	The better solution would be some sort of pipe, but that requires digging
	into the main X event loop.

	Basically, using ClientMessage events is dead simple to implement but easy
	to break when run concurrently. Using pipes is harder to implement, but
	also harder to break when run concurrently. I think.
*/

import (
	"flag"
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

	"github.com/BurntSushi/wingo/cmdusage"
	"github.com/BurntSushi/wingo/logger"
)

func main() {
	X, err := xgbutil.Dial("")
	if err != nil {
		logger.Error.Println(err)
		logger.Error.Println("Error connecting to X, quitting...")
		return
	}
	defer X.Conn().Close()

	// Get command from arguments
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: wingo-cmd CommandName [CommandArgs]")
		return
	}

	commandPieces := flag.Args()
	cmdName := commandPieces[0]
	cmdFull := strings.Join(commandPieces, " ")

	// make sure we start with failure
	cmdusage.StatusSet(X, false)
	success := false

	// Set the command before sending request to run command.
	err = cmdusage.CmdSet(X, cmdFull)
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
				success = cmdusage.StatusGet(X)
				if success {
					os.Exit(0)
				} else {
					logger.Warning.Printf("Error running '%s'", cmdFull)
					cmdusage.ShowUsage(cmdName)
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
