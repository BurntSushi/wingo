package main

import (
    "image/color"
    "strconv"
    "strings"
)

import "github.com/BurntSushi/wingo/wini"

type theme struct {
    borders themeBorders
    slim themeSlim
}

type themeBorders struct {
    borderSize int
    aThinColor, iThinColor int
    aBorderColor, iBorderColor themeColor
}

type themeSlim struct {
    borderSize int
    aBorderColor, iBorderColor int
}

func defaultTheme() *theme {
    return &theme{
        borders: themeBorders{
            borderSize: 10,
            aThinColor: 0x0,
            iThinColor: 0x0,
            aBorderColor: newThemeColor(0x3366ff),
            iBorderColor: newThemeColor(0xdfdcdf),
        },
        slim: themeSlim{
            borderSize: 10,
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
            for _, key := range tdata.Keys(section) {
                loadBorderOption(key)
            }
        case section == "slim":
            for _, key := range tdata.Keys(section) {
                loadSlimOption(key)
            }
        }
    }

    return nil
}

func loadThemeFile() (*wini.Data, error) {
    return wini.Parse("theme.wini")
}

func loadBorderOption(k wini.Key) {
    switch k.Name() {
    case "border_size": setInt(k, &THEME.borders.borderSize)
    case "a_thin_color": setInt(k, &THEME.borders.aThinColor)
    case "i_thin_color": setInt(k, &THEME.borders.iThinColor)
    case "a_border_color": setGradient(k, &THEME.borders.aBorderColor)
    case "i_border_color": setGradient(k, &THEME.borders.iBorderColor)
    }
}

func loadSlimOption(k wini.Key) {
    switch k.Name() {
    case "border_size": setInt(k, &THEME.slim.borderSize)
    case "a_border_color": setInt(k, &THEME.slim.aBorderColor)
    case "i_border_color": setInt(k, &THEME.slim.iBorderColor)
    }
}

func setGradient(k wini.Key, clr *themeColor) {
    // Check to make sure we have a value for this key
    vals := k.Strings()
    if len(vals) == 0 {
        logWarning.Println(k.Err("No values found."))
        return
    }

    // Use the last value
    val := vals[len(vals) - 1]

    // If there are no spaces, it can't be a gradient.
    if strings.Index(val, " ") == -1 {
        if start, ok := getLastInt(k); ok {
            clr.start = start
        }
        return
    }

    // Okay, now we have to do things manually.
    // Split up the value into two pieces separated by whitespace and parse
    // each piece as an int.
    splitted := strings.Split(val, " ")
    if len(splitted) != 2 {
        logWarning.Println(k.Err("Expected a gradient value (two colors " +
                                 "separated by a space), but found '%s' " +
                                 "instead.", val))
        return
    }

    start, err := strconv.ParseInt(strings.TrimSpace(splitted[0]), 0, 0)
    if err != nil {
        logWarning.Println(k.Err("'%s' is not an integer. (%s)",
                                 splitted[0], err))
        return
    }

    end, err := strconv.ParseInt(strings.TrimSpace(splitted[1]), 0, 0)
    if err != nil {
        logWarning.Println(k.Err("'%s' is not an integer. (%s)",
                                 splitted[1], err))
        return
    }

    // finally...
    clr.start, clr.end = int(start), int(end)
}

type themeColor struct {
    start, end int
}

func (tc themeColor) isGradient() bool {
    return tc.start >= 0 && tc.start <= 0xffffff &&
           tc.end >= 0 && tc.end <= 0xffffff
}

// steps returns a slice of colors corresponding to the gradient
// of colors from start to end. The size is determined by the 'size' parameter.
// The first and last colors in the slice are guaranteed to be
// tc.start and tc.end. (Unless the size is 1, in which case, the first and
// only color in the slice is tc.start.)
func (tc themeColor) steps(size int) []color.RGBA {
    // naughty
    if !tc.isGradient() {
        stps := make([]color.RGBA, size)
        for i := 0; i < size; i++ {
            stps[i] = ColorFromInt(tc.start)
        }
    }

    // yikes
    if size == 0 || size == 1 {
        return []color.RGBA{ColorFromInt(tc.start)}
    }

    stps := make([]color.RGBA, size)
    stps[0], stps[size - 1] = ColorFromInt(tc.start), ColorFromInt(tc.end)

    // no more?
    if size == 2 {
        return stps
    }

    sr, sg, sb := RGBFromInt(tc.start)
    er, eg, eb := RGBFromInt(tc.end)

    rinc := int(round(float64(er - sr) / float64(size)))
    ginc := int(round(float64(eg - sg) / float64(size)))
    binc := int(round(float64(eb - sb) / float64(size)))

    var nr, ng, nb int
    for i := 1; i < size - 1; i++ {
        nr = sr + rinc * i
        ng = sg + ginc * i
        nb = sb + binc * i
        stps[i] = ColorFromInt(IntFromRGB(nr, ng, nb))
    }

    return stps
}

func newThemeColor(clr int) themeColor {
    return themeColor{start: clr, end: -1}
}

func newThemeGradient(start, end int) themeColor {
    return themeColor{start: start, end: end}
}

func setInt(k wini.Key, place *int) {
    if v, ok := getLastInt(k); ok {
        *place = int(v)
    }
}

func getLastInt(k wini.Key) (int, bool) {
    vals, err := k.Ints()
    if err != nil {
        logWarning.Println(err)
        return 0, false
    } else if len(vals) == 0 {
        logWarning.Println(k.Err("No values found."))
        return 0, false
    }

    return vals[len(vals) - 1], true
}

