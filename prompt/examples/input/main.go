// Example input shows how to use an Input prompt from the prompt pacakge.
package main

import (
	"log"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/prompt"
)

// response is the callback that gets executed whenever the user hits
// enter (the "confirm" key). The text parameter contains the string in
// the input box.
func response(inp *prompt.Input, text string) {
	// If you type my name, we exit!
	if text == "Andrew" {
		println("You have the same name as me! Bye!")

		canceled(inp)
		return
	}
	println("Hello " + text + "!")
}

// canceled is the callback that gets executed whenever the prompt is canceled.
// This can occur when the user presses escape (the "cancel" key).
func canceled(inp *prompt.Input) {
	xevent.Quit(inp.X)
}

func main() {
	X, err := xgbutil.NewConn()
	if err != nil {
		log.Fatalln(err)
	}

	// The input box uses the keybind module, so we must initialize it.
	keybind.Initialize(X)

	// Creating a new input prompt is as simple as supply an X connection,
	// a theme and a configuration. We use built in defaults here.
	inpPrompt := prompt.NewInput(X,
		prompt.DefaultInputTheme, prompt.DefaultInputConfig)

	// Show maps the input prompt window and sets the focus. It returns
	// immediately, and the main X event loop is started.
	// Also, we use the root window geometry to make sure the prompt is
	// centered in the middle of the screen. 'response' and 'canceled' are
	// callback functions. See their respective commends for more details.
	inpPrompt.Show(xwindow.RootGeometry(X),
		"What is your name?", response, canceled)

	xevent.Main(X)
}
