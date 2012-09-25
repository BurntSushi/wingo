package main

import (
	"strings"

	"github.com/BurntSushi/gribble"

	"github.com/BurntSushi/xgbutil/ewmh"

	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/wini"
)

type conf struct {
	cmdEnv                *gribble.Environment
	mouse                 map[string][]mouseCommand
	key                   map[string][]keyCommand
	ffm                   bool
	workspaces            []string
	alwaysFloating        []string
	confirmKey, cancelKey string
	backspaceKey          string
	tabKey, revTabKey     string
}

func newConf() *conf {
	cmdEnv := gribble.New([]gribble.Command{
		&CmdClose{},
		&CmdFocus{},
		&CmdFocusRaise{},
		&CmdIconifyToggle{},
		&CmdMouseMove{},
		&CmdMouseResize{},
		&CmdMove{},
		&CmdPromptCycleNext{},
		&CmdRaise{},
		&CmdResize{},
		&CmdQuit{},
		&CmdShell{},
	})
	return &conf{
		cmdEnv:         cmdEnv,
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

func loadConfig() (*conf, error) {
	conf := newConf() // globally defined in wingo.go
	if err := conf.loadMouseConfig(); err != nil {
		return nil, err
	}
	if err := conf.loadKeyConfig(); err != nil {
		return nil, err
	}
	if err := conf.loadOptionsConfig(); err != nil {
		return nil, err
	}
	return conf, nil
}

func (conf *conf) loadMouseConfig() error {
	cdata, err := loadMouseConfigFile()
	if err != nil {
		return err
	}
	for _, section := range cdata.Sections() {
		conf.loadMouseConfigSection(cdata, section)
	}
	return nil
}

func (conf *conf) loadKeyConfig() error {
	cdata, err := loadKeyConfigFile()
	if err != nil {
		return err
	}
	for _, section := range cdata.Sections() {
		conf.loadKeyConfigSection(cdata, section)
	}
	return nil
}

func (conf *conf) loadOptionsConfig() error {
	cdata, err := loadOptionsConfigFile()
	if err != nil {
		return err
	}
	for _, section := range cdata.Sections() {
		conf.loadOptionsConfigSection(cdata, section)
	}
	return nil
}

// loadMouseConfigSection does two things:
// 1) Inspects the section name to infer the identifier. In general, the
// "mouse" prefix is removed, and whatever remains is the identifier. There
// are two special cases: "MouseBorders*" turns into "borders_*" and
// "MouseFull*" turns into "full_*".
// 2) Constructs a "mouseCommand" for *every* value.
func (conf *conf) loadMouseConfigSection(cdata *wini.Data, section string) {
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
			if _, ok := conf.mouse[ident]; !ok {
				conf.mouse[ident] = make([]mouseCommand, 0)
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

			gribbleCmd, err := conf.cmdEnv.Command(cmd)
			if err == nil {
				mcmd := mouseCommand{
					cmd:       gribbleCmd,
					cmdName:   conf.cmdEnv.CommandName(cmd),
					down:      down,
					buttonStr: buttonStr,
				}
				conf.mouse[ident] = append(conf.mouse[ident], mcmd)
			} else {
				logger.Warning.Printf(
					"Could not parse command '%s' because: %s", cmd, err)
			}
		}
	}
}

func (conf *conf) loadKeyConfigSection(cdata *wini.Data, section string) {
	for _, key := range cdata.Keys(section) {
		keyStr := key.Name()
		for _, cmd := range key.Strings() {
			if _, ok := conf.key[section]; !ok {
				conf.key[section] = make([]keyCommand, 0)
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

			gribbleCmd, err := conf.cmdEnv.Command(cmd)
			if err == nil {
				kcmd := keyCommand{
					cmd:     gribbleCmd,
					cmdName: conf.cmdEnv.CommandName(cmd),
					down:    down,
					keyStr:  keyStr,
				}
				conf.key[section] = append(conf.key[section], kcmd)
			} else {
				logger.Warning.Printf(
					"Could not parse command '%s' because: %s", cmd, err)
			}
		}
	}
}

func (conf *conf) loadOptionsConfigSection(cdata *wini.Data, section string) {
	for _, key := range cdata.Keys(section) {
		option := key.Name()
		switch option {
		case "workspaces":
			if workspaces, ok := getLastString(key); ok {
				conf.workspaces = strings.Split(workspaces, " ")
			}
		case "always_floating":
			if alwaysFloating, ok := getLastString(key); ok {
				conf.alwaysFloating = strings.Split(alwaysFloating, " ")
			}
		case "focus_follows_mouse":
			setBool(key, &conf.ffm)
		case "cancel":
			setString(key, &conf.cancelKey)
		case "confirm":
			setString(key, &conf.confirmKey)
		}
	}
}

func loadMouseConfigFile() (*wini.Data, error) {
	return wini.Parse("config/mouse.wini")
}

func loadKeyConfigFile() (*wini.Data, error) {
	return wini.Parse("config/key.wini")
}

func loadOptionsConfigFile() (*wini.Data, error) {
	return wini.Parse("config/options.wini")
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
