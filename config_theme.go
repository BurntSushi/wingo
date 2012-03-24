package main

import "github.com/BurntSushi/wingo/wini"

type theme struct {
    borders themeBorders
}

type themeBorders struct {
    borderWidth uint16
    cornerSize uint16
    aThinColor, iThinColor int
    aBorderColor, iBorderColor int
}

func defaultTheme() *theme {
    return &theme{
        borders: themeBorders{
            borderWidth: 10,
            cornerSize: 24,
            aThinColor: 0x0,
            iThinColor: 0x0,
            aBorderColor: 0x3366ff,
            iBorderColor: 0xdfdcdf,
        },
    }
}

func loadTheme() error {
    THEME = defaultTheme() // globally defined in wingo.go

    tdata, err := loadThemeFile()
    if err != nil {
        return err
    }

    for _, section := range tdata.Sections() {
        switch {
        case section == "borders":
            loadBorderTheme(tdata, section)
        }
    }

    return nil
}

func loadThemeFile() (*wini.Data, error) {
    return wini.Parse("theme.wini")
}

func loadBorderTheme(tdata *wini.Data, section string) {
    for _, key := range tdata.Keys(section) {
        loadBorderOption(key)
    }
}

func loadBorderOption(k wini.Key) {
    switch k.Name() {
    case "border_width": setFirstUint16(k, &THEME.borders.borderWidth)
    case "corner_size": setFirstUint16(k, &THEME.borders.cornerSize)
    case "a_thin_color": setFirstInt(k, &THEME.borders.aThinColor)
    case "i_thin_color": setFirstInt(k, &THEME.borders.iThinColor)
    case "a_border_color": setFirstInt(k, &THEME.borders.aBorderColor)
    case "i_border_color": setFirstInt(k, &THEME.borders.iBorderColor)
    }
}

func setFirstInt(k wini.Key, place *int) {
    if v, ok := getFirstInt(k); ok {
        *place = int(v)
    }
}

func setFirstUint16(k wini.Key, place *uint16) {
    if v, ok := getFirstInt(k); ok {
        *place = uint16(v)
    }
}

func getFirstInt(k wini.Key) (int, bool) {
    vals, err := k.Ints()
    if err != nil {
        logWarning.Println(err)
        return 0, false
    } else if len(vals) == 0 {
        logWarning.Println(k.Err("No values found."))
        return 0, false
    }

    return vals[0], true
}

