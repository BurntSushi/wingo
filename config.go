package main

import (
    "strings"
)

import "github.com/BurntSushi/wingo/wini"

import (
    "github.com/BurntSushi/xgbutil/ewmh"
)

type conf struct {
    mouse map[string][]mouseCommand
}

func defaultConfig() *conf {
    return &conf{
        mouse: map[string][]mouseCommand{},
    }
}

func loadConfig() error {
    CONF = defaultConfig() // globally defined in wingo.go

    cdata, err := loadConfigFile()
    if err != nil {
        return err
    }

    for _, section := range cdata.Sections() {
        // Each section name *means* something.
        // For example, sections starting with "Mouse" should
        // only contain mouse commands.
        switch {
        case len(section) >= 5 && section[0:5] == "mouse":
            loadMouseConfig(cdata, section)
        }
    }

    return nil
}

// loadMouseConfig does two things:
// 1) Inspects the section name to infer the identifier. In general, the
// "mouse" prefix is removed, and whatever remains is the identifier. There
// are two special cases: "MouseBorders*" turns into "borders_*" and
// "MouseFull*" turns into "full_*".
// 2) Constructs a "mouseCommand" for *every* value.
func loadMouseConfig(cdata *wini.Data, section string) {
    ident := ""
    unmouse := section[5:]
    switch {
    case len(unmouse) > 7 && unmouse[:7] == "borders":
        ident = "borders_" + unmouse[7:]
    case len(unmouse) > 4 && unmouse[:4] == "full":
        ident = "full_" + unmouse[5:]
    default:
        ident = unmouse
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
                if mouseStr[spacei + 1:] == "up" {
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
                    direction = strToDirection(cmd[spacei + 1:])
                    cmd = cmd[:spacei]
                }
            }

            mcmd := mouseCommand{
                cmd: cmd,
                down: down,
                buttonStr: buttonStr,
                direction: direction,
            }
            CONF.mouse[ident] = append(CONF.mouse[ident], mcmd)
        }
    }
}

func loadConfigFile() (*wini.Data, error) {
    return wini.Parse("config.wini")
}

func strToDirection(s string) uint32 {
    switch strings.ToLower(s) {
    case "top": return ewmh.SizeTop
    case "topleft": return ewmh.SizeTopLeft
    case "topright": return ewmh.SizeTopRight
    case "bottom": return ewmh.SizeBottom
    case "bottomleft": return ewmh.SizeBottomLeft
    case "bottomright": return ewmh.SizeBottomRight
    case "left": return ewmh.SizeLeft
    case "right": return ewmh.SizeRight
    }

    return ewmh.Infer
}

