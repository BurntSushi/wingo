package main

import (
	"strings"

	"github.com/BurntSushi/gribble"

	"github.com/BurntSushi/xgbutil/ewmh"

	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/wini"
)

type config struct {
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

// newConfig
func newConfig() *config {
	return &config{
		cmdEnv:         gribbleCommandEnv, // see command.go
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

// loadConfig reads all configuration files and loads them into the
// a single config value.
//
// Most of this code is incredibly boring.
func loadConfig() (*config, error) {
	conf := newConfig() // globally defined in wingo.go

	type confFile struct {
		fpath       string
		loadSection func(*config, *wini.Data, string)
	}
	cfiles := []confFile{
		{"config/mouse.wini", (*config).loadMouseConfigSection},
		{"config/key.wini", (*config).loadKeyConfigSection},
		{"config/options.wini", (*config).loadOptionsConfigSection},
	}
	for _, cfile := range cfiles {
		cdata, err := wini.Parse(cfile.fpath)
		if err != nil {
			return nil, err
		}
		for _, section := range cdata.Sections() {
			cfile.loadSection(conf, cdata, section)
		}
	}
	return conf, nil
}

// loadMouseConfigSection does two things:
// 1) Inspects the section name to infer the identifier. In general, the
// "mouse" prefix is removed, and whatever remains is the identifier. There
// are two special cases: "MouseBorders*" turns into "borders_*" and
// "MouseFull*" turns into "full_*".
// 2) Constructs a "mouseCommand" for *every* value.
//
// The idents are used for attaching mouse commands to the corresponding
// frames. (See the mouseCommand methods.)
func (conf *config) loadMouseConfigSection(cdata *wini.Data, section string) {
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

			gribbleCmd, err := conf.cmdEnv.Command(cmd)
			if err != nil {
				logger.Warning.Printf(
					"Could not parse command '%s' because: %s", cmd, err)
			} else {
				down, justMouseStr := isDown(mouseStr)
				mcmd := mouseCommand{
					cmd:       gribbleCmd,
					cmdName:   conf.cmdEnv.CommandName(cmd),
					down:      down,
					buttonStr: justMouseStr,
				}
				conf.mouse[ident] = append(conf.mouse[ident], mcmd)
			}
		}
	}
}

func (conf *config) loadKeyConfigSection(cdata *wini.Data, section string) {
	for _, key := range cdata.Keys(section) {
		keyStr := key.Name()
		for _, cmd := range key.Strings() {
			if _, ok := conf.key[section]; !ok {
				conf.key[section] = make([]keyCommand, 0)
			}

			gribbleCmd, err := conf.cmdEnv.Command(cmd)
			if err != nil {
				logger.Warning.Printf(
					"Could not parse command '%s' because: %s", cmd, err)
			} else {
				down, justKeyStr := isDown(keyStr)
				kcmd := keyCommand{
					cmd:     gribbleCmd,
					cmdName: conf.cmdEnv.CommandName(cmd),
					down:    down,
					keyStr:  justKeyStr,
				}
				conf.key[section] = append(conf.key[section], kcmd)
			}
		}
	}
}

func (conf *config) loadOptionsConfigSection(cdata *wini.Data, section string) {
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

// strToDirection converts a string representation of a mouse direction
// to an xgbutil.ewmh constant value. It is case insensitive.
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

// isDown takes a key/mouse combination, and looks for the keyword "up".
// If "up" exists, isDown returns false. Otherwise, true.
// It also returns the key/mouse string without "up" or "down".
func isDown(keyStr string) (bool, string) {
	spacei := strings.Index(keyStr, " ")
	down := true
	if spacei > -1 {
		if strings.ToLower(keyStr[spacei+1:]) == "up" {
			down = false
		}
		keyStr = keyStr[:spacei]
	}
	return down, keyStr
}
