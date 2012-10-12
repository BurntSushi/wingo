package main

import (
	"log"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/BurntSushi/wingo/prompt"
)

var X *xgbutil.XUtil

func response(text string) {
	if text == "quit" {
		xevent.Quit(X)
		return
	}
	println(text)
}

func main() {
	var err error

	X, err = xgbutil.NewConn()
	if err != nil {
		log.Fatalln(err)
	}

	keybind.Initialize(X)

	inpPrompt := prompt.NewInput(X,
		prompt.DefaultInputTheme, prompt.DefaultInputConfig)

	inpPrompt.Show(xwindow.RootGeometry(X), "Hello: ", response)

	xevent.Main(X)
}
