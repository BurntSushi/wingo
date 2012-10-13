// Example message shows how to use a Message prompt from the prompt pacakge.
package main

import (
	"log"
	"os"
	"time"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/prompt"
)

// hidden is a callback that is executed when the message disappears.
func hidden(msg *prompt.Message) {
	os.Exit(0)
}

func main() {
	X, err := xgbutil.NewConn()
	if err != nil {
		log.Fatalln(err)
	}

	// The message box uses the keybind module, so we must initialize it.
	keybind.Initialize(X)

	// Creating a new message prompt is as simple as supply an X connection,
	// a theme and a configuration. We use built in defaults here.
	msgPrompt := prompt.NewMessage(X,
		prompt.DefaultMessageTheme, prompt.DefaultMessageConfig)

	// Show maps the message prompt window.
	// If a duration is specified, the window does NOT acquire focus and
	// automatically disappears after the specified time.
	// If a duration is not specified (i.e., '0'), then the window is mapped,
	// and acquires focus. It does not disappear until it loses focus or when
	// the user hits the "confirm" or "cancel" keys (usually "enter" and
	// "escape").
	timeout := 2 * time.Second // or "0" for no timeout.
	msgPrompt.Show(xwindow.RootGeometry(X), "Hello, world!", timeout, hidden)

	xevent.Main(X)
}
