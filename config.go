package main

import (
	"strings"

	"github.com/BurntSushi/xgbutil/ewmh"

	"github.com/BurntSushi/wingo/config"
	"github.com/BurntSushi/wingo/wini"
)

type conf struct {
	mouse                 map[string][]mouseCommand
	key                   map[string][]keyCommand
	ffm                   bool
	workspaces            []string
	alwaysFloating        []string
	confirmKey, cancelKey string
	backspaceKey          string
	tabKey, revTabKey     string
}

func defaultConfig() *conf {
	return &conf{
		mouse:          map[string][]mouseCommand{},
		key:            map[string][]keyCommand{},
		ffm:            false,
		workspaces:     []string{"1", "2", "3", "4"},
		alwaysFloating: []string{},
		confirmKey:     "Return",
		cancelKey:      "Escape",
		backspaceKey:   "BackSpace",
		tabKey:         "Tab",
		revTabKey:      "ISO_Left_Tab",
	}
}

func loadConfig() error {
	CONF = defaultConfig() // globally defined in wingo.go

	if err := loadMouseConfig(); err != nil {
		return err
	}
	if err := loadKeyConfig(); err != nil {
		return err
	}
	if err := loadOptionsConfig(); err != nil {
		return err
	}

	return nil
}

func loadMouseConfig() error {
	cdata, err := loadMouseConfigFile()
	if err != nil {
		return err
	}

	for _, section := range cdata.Sections() {
		loadMouseConfigSection(cdata, section)
	}

	return nil
}

func loadKeyConfig() error {
	cdata, err := loadKeyConfigFile()
	if err != nil {
		return err
	}

	for _, section := range cdata.Sections() {
		loadKeyConfigSection(cdata, section)
	}

	return nil
}

func loadOptionsConfig() error {
	cdata, err := loadOptionsConfigFile()
	if err != nil {
		return err
	}

	for _, section := range cdata.Sections() {
		loadOptionsConfigSection(cdata, section)
	}

	return nil
}

// loadMouseConfigSection does two things:
// 1) Inspects the section name to infer the identifier. In general, the
// "mouse" prefix is removed, and whatever remains is the identifier. There
// are two special cases: "MouseBorders*" turns into "borders_*" and
// "MouseFull*" turns into "full_*".
// 2) Constructs a "mouseCommand" for *every* value.
func loadMouseConfigSection(cdata *wini.Data, section string) {
	ident := ""
	switch {
	case len(section) > 7 && section[:7] == "borders":
		ident = "borders_" + section[7:]
	case len(section) > 4 && section[:4] == "full":
		ident = "full_" + section[4:]
	default:
		ident = section
	}

	for _, key := range cdata.Keys(section) {
		mouseStr := key.Name()
		for _, cmd := range key.Strings() {
			if _, ok := CONF.mouse[ident]; !ok {
				CONF.mouse[ident] = make([]mouseCommand, 0)
			}

			// "mouseStr" is actually made up of a button string
			// and a toggle of "down" or "up" corresponding to a button press
			// or a button release, respectively. Look for that.
			// If it isn't there, assume 'down'
			spacei := strings.Index(mouseStr, " ")
			down := true
			buttonStr := mouseStr
			if spacei > -1 {
				buttonStr = mouseStr[:spacei]
				if mouseStr[spacei+1:] == "up" {
					down = false
				}
			}

			// 'cmd' can sometimes take parameters. For mouse commands,
			// only one such command does so: Resize. Check for that.
			// (The parameter is which way to resize. If it's absent, default
			// to "Infer".)
			var direction uint32 = ewmh.Infer
			if len(cmd) > 6 && strings.ToLower(cmd[:6]) == "resize" {
				spacei := strings.Index(cmd, " ")
				if spacei > -1 {
					direction = strToDirection(cmd[spacei+1:])
					cmd = cmd[:spacei]
				}
			}

			mcmd := mouseCommand{
				cmd:       cmd,
				down:      down,
				buttonStr: buttonStr,
				direction: direction,
			}
			CONF.mouse[ident] = append(CONF.mouse[ident], mcmd)
		}
	}
}

func loadKeyConfigSection(cdata *wini.Data, section string) {
	for _, key := range cdata.Keys(section) {
		keyStr := key.Name()
		for _, cmd := range key.Strings() {
			if _, ok := CONF.key[section]; !ok {
				CONF.key[section] = make([]keyCommand, 0)
			}

			// "keyStr" is actually made up of a key string
			// and a toggle of "down" or "up" corresponding to a key press
			// or a key release, respectively. Look for that.
			// If it isn't there, assume 'down'
			spacei := strings.Index(keyStr, " ")
			down := true
			if spacei > -1 {
				if keyStr[spacei+1:] == "up" {
					down = false
				}
				keyStr = keyStr[:spacei]
			}

			// 'cmd' might have space separated parameters
			cmdName, args := commandParse(cmd)

			kcmd := keyCommand{
				cmd:    cmdName,
				args:   args,
				down:   down,
				keyStr: keyStr,
			}
			CONF.key[section] = append(CONF.key[section], kcmd)
		}
	}
}

func loadOptionsConfigSection(cdata *wini.Data, section string) {
	for _, key := range cdata.Keys(section) {
		option := key.Name()
		switch option {
		case "workspaces":
			if workspaces, ok := config.GetLastString(key); ok {
				CONF.workspaces = strings.Split(workspaces, " ")
			}
		case "always_floating":
			if alwaysFloating, ok := config.GetLastString(key); ok {
				CONF.alwaysFloating = strings.Split(alwaysFloating, " ")
			}
		case "focus_follows_mouse":
			config.SetBool(key, &CONF.ffm)
		case "cancel":
			config.SetString(key, &CONF.cancelKey)
		case "confirm":
			config.SetString(key, &CONF.confirmKey)
		}
	}
}

func loadMouseConfigFile() (*wini.Data, error) {
	return wini.Parse("configdata/mouse.wini")
}

func loadKeyConfigFile() (*wini.Data, error) {
	return wini.Parse("configdata/key.wini")
}

func loadOptionsConfigFile() (*wini.Data, error) {
	return wini.Parse("configdata/options.wini")
}

func strToDirection(s string) uint32 {
	switch strings.ToLower(s) {
	case "top":
		return ewmh.SizeTop
	case "bottom":
		return ewmh.SizeBottom
	case "left":
		return ewmh.SizeLeft
	case "right":
		return ewmh.SizeRight
	case "topleft":
		return ewmh.SizeTopLeft
	case "topright":
		return ewmh.SizeTopRight
	case "bottomleft":
		return ewmh.SizeBottomLeft
	case "bottomright":
		return ewmh.SizeBottomRight
	}

	return ewmh.Infer
}
