package main

import "github.com/BurntSushi/wingo/wini"

import (
    "github.com/BurntSushi/xgbutil/ewmh"
)

type conf struct {
    mouse map[string][]mouseCommand
}

func defaultConfig() *conf {
    return &conf{
        mouse: map[string][]mouseCommand{
            "client": []mouseCommand{{"FocusRaise", true, "1", 0}},
            "frame": []mouseCommand{{"Move", true, "Mod4-1", 0},
                                    {"Resize", true, "Mod4-3", ewmh.Infer}},
            "borders_topside": []mouseCommand{{"Resize", true, "1", ewmh.SizeTop},
                                              {"FocusRaise", true, "1", 0}},
        },
    }
}

func loadConfig() error {
    CONF = defaultConfig()

    _, err := wini.Parse("config.wini")
    if err != nil {
        return err
    }

    return nil
}

