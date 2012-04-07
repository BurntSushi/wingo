package main

import (
    "image/draw"
    "io/ioutil"
    "strconv"
    "strings"
)

import (
    "code.google.com/p/freetype-go/freetype"
    "code.google.com/p/freetype-go/freetype/truetype"
)

// import "burntsushi.net/go/x-go-binding/xgb" 

import "burntsushi.net/go/xgbutil/xgraphics"

import (
    "burntsushi.net/go/wingo/bindata"
    "burntsushi.net/go/wingo/wini"
)

type theme struct {
    defaultIcon draw.Image
    full themeFull
    borders themeBorders
    slim themeSlim
    prompt themePrompt
}

type themeFull struct {
    font *truetype.Font
    fontSize float64
    aFontColor, iFontColor int

    titleSize int
    aTitleColor, iTitleColor themeColor

    aCloseButton, iCloseButton draw.Image
    aCloseColor, iCloseColor int

    aMaximizeButton, iMaximizeButton draw.Image
    aMaximizeColor, iMaximizeColor int

    aMinimizeButton, iMinimizeButton draw.Image
    aMinimizeColor, iMinimizeColor int

    borderSize int
    aBorderColor, iBorderColor int
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

type themePrompt struct {
    bgColor int
    borderColor int
    borderSize int
    padding int

    font *truetype.Font
    fontSize float64
    fontColor int

    cycleIconSize int
    cycleIconBorderSize int
    cycleIconTransparency int
}

func defaultTheme() *theme {
    return &theme{
        defaultIcon: builtInIcon(),
        full: themeFull{
            font: builtInFont(),
            fontSize: 15,
            aFontColor: 0xffffff,
            iFontColor: 0x000000,

            titleSize: 25,
            aTitleColor: newThemeColor(0x3366ff),
            iTitleColor: newThemeColor(0xdfdcdf),

            aCloseButton: builtInButton(bindata.ClosePng),
            iCloseButton: builtInButton(bindata.ClosePng),
            aCloseColor: 0xffffff,
            iCloseColor: 0x000000,

            aMaximizeButton: builtInButton(bindata.MaximizePng),
            iMaximizeButton: builtInButton(bindata.MaximizePng),
            aMaximizeColor: 0xffffff,
            iMaximizeColor: 0x000000,

            aMinimizeButton: builtInButton(bindata.MinimizePng),
            iMinimizeButton: builtInButton(bindata.MinimizePng),
            aMinimizeColor: 0xffffff,
            iMinimizeColor: 0x000000,

            borderSize: 10,
            aBorderColor: 0x3366ff,
            iBorderColor: 0xdfdcdf,
        },
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
        prompt: themePrompt{
            bgColor: 0xffffff,
            borderColor: 0x585a5d,
            borderSize: 10,
            padding: 10,
            font: builtInFont(),
            fontSize: 15,
            fontColor: 0x000000,
            cycleIconSize: 32,
            cycleIconBorderSize: 3,
            cycleIconTransparency: 50,
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
        switch section {
        case "misc":
            for _, key := range tdata.Keys(section) {
                loadMiscOption(key)
            }
        case "full":
            for _, key := range tdata.Keys(section) {
                loadFullOption(key)
            }
        case "borders":
            for _, key := range tdata.Keys(section) {
                loadBorderOption(key)
            }
        case "slim":
            for _, key := range tdata.Keys(section) {
                loadSlimOption(key)
            }
        case "prompt":
            for _, key := range tdata.Keys(section) {
                loadPromptOption(key)
            }
        }
    }

    // re-color some images
    xgraphics.ColorImage(THEME.full.aCloseButton,
                         colorFromInt(THEME.full.aCloseColor))
    xgraphics.ColorImage(THEME.full.iCloseButton,
                         colorFromInt(THEME.full.iCloseColor))
    xgraphics.ColorImage(THEME.full.aMaximizeButton,
                         colorFromInt(THEME.full.aMaximizeColor))
    xgraphics.ColorImage(THEME.full.iMaximizeButton,
                         colorFromInt(THEME.full.iMaximizeColor))
    xgraphics.ColorImage(THEME.full.aMinimizeButton,
                         colorFromInt(THEME.full.aMinimizeColor))
    xgraphics.ColorImage(THEME.full.iMinimizeButton,
                         colorFromInt(THEME.full.iMinimizeColor))

    // Scale some images...
    THEME.full.aCloseButton = xgraphics.Scale(THEME.full.aCloseButton,
                                              THEME.full.titleSize,
                                              THEME.full.titleSize)
    THEME.full.iCloseButton = xgraphics.Scale(THEME.full.iCloseButton,
                                              THEME.full.titleSize,
                                              THEME.full.titleSize)
    THEME.full.aMaximizeButton = xgraphics.Scale(THEME.full.aMaximizeButton,
                                              THEME.full.titleSize,
                                              THEME.full.titleSize)
    THEME.full.iMaximizeButton = xgraphics.Scale(THEME.full.iMaximizeButton,
                                              THEME.full.titleSize,
                                              THEME.full.titleSize)
    THEME.full.aMinimizeButton = xgraphics.Scale(THEME.full.aMinimizeButton,
                                              THEME.full.titleSize,
                                              THEME.full.titleSize)
    THEME.full.iMinimizeButton = xgraphics.Scale(THEME.full.iMinimizeButton,
                                              THEME.full.titleSize,
                                              THEME.full.titleSize)

    return nil
}

func loadThemeFile() (*wini.Data, error) {
    return wini.Parse("config/theme.wini")
}

func loadMiscOption(k wini.Key) {
    switch k.Name() {
    case "default_icon": setImage(k, &THEME.defaultIcon)
    }
}

func loadFullOption(k wini.Key) {
    switch k.Name() {
    case "font": setFont(k, &THEME.full.font)
    case "font_size": setFloat(k, &THEME.full.fontSize)
    case "a_font_color": setInt(k, &THEME.full.aFontColor)
    case "i_font_color": setInt(k, &THEME.full.iFontColor)
    case "title_size": setInt(k, &THEME.full.titleSize)
    case "a_title_color": setGradient(k, &THEME.full.aTitleColor)
    case "i_title_color": setGradient(k, &THEME.full.iTitleColor)
    case "close":
        setImage(k, &THEME.full.aCloseButton)
        setImage(k, &THEME.full.iCloseButton)
    case "a_close_color": setInt(k, &THEME.full.aCloseColor)
    case "i_close_color": setInt(k, &THEME.full.iCloseColor)
    case "maximize":
        setImage(k, &THEME.full.aMaximizeButton)
        setImage(k, &THEME.full.iMaximizeButton)
    case "a_maximize_color": setInt(k, &THEME.full.aMaximizeColor)
    case "i_maximize_color": setInt(k, &THEME.full.iMaximizeColor)
    case "minimize":
        setImage(k, &THEME.full.aMinimizeButton)
        setImage(k, &THEME.full.iMinimizeButton)
    case "a_minimize_color": setInt(k, &THEME.full.aMinimizeColor)
    case "i_minimize_color": setInt(k, &THEME.full.iMinimizeColor)
    case "border_size": setInt(k, &THEME.full.borderSize)
    case "a_border_color": setInt(k, &THEME.full.aBorderColor)
    case "i_border_color": setInt(k, &THEME.full.iBorderColor)
    }
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

func loadPromptOption(k wini.Key) {
    switch k.Name() {
    case "bg_color": setInt(k, &THEME.prompt.bgColor)
    case "border_color": setInt(k, &THEME.prompt.borderColor)
    case "border_size": setInt(k, &THEME.prompt.borderSize)
    case "padding": setInt(k, &THEME.prompt.padding)
    case "font": setFont(k, &THEME.prompt.font)
    case "font_size": setFloat(k, &THEME.prompt.fontSize)
    case "font_color": setInt(k, &THEME.prompt.fontColor)
    case "cycle_icon_size": setInt(k, &THEME.prompt.cycleIconSize)
    case "cycle_icon_border_size": setInt(k, &THEME.prompt.cycleIconBorderSize)
    case "cycle_icon_transparency":
        setInt(k, &THEME.prompt.cycleIconTransparency)

        // naughty!
        if THEME.prompt.cycleIconTransparency < 0 ||
           THEME.prompt.cycleIconTransparency > 100 {
            logWarning.Printf("Illegal value '%s' provided for " +
                              "'cycle_icon_transparency'. Transparency " +
                              "values must be in the range [0, 100], " +
                              "inclusive. Using 100 by default.")
            THEME.prompt.cycleIconTransparency = 100
        }
    }
}

type themeColor struct {
    start, end int
}

func newThemeColor(clr int) themeColor {
    return themeColor{start: clr, end: -1}
}

func newThemeGradient(start, end int) themeColor {
    return themeColor{start: start, end: end}
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

func builtInIcon() draw.Image {
    img, err := xgraphics.LoadPngFromBytes(bindata.WingoPng())
    if err != nil {
        logWarning.Printf("Could not get built in icon image because: %v",
                          err)
        return nil
    }
    return img
}

func builtInButton(loadBuiltIn func() []byte) draw.Image {
    img, err := xgraphics.LoadPngFromBytes(loadBuiltIn())
    if err != nil {
        logWarning.Printf("Could not get built in button image because: %v",
                          err)
        return nil
    }
    return img
}

func setImage(k wini.Key, place *draw.Image) {
    if v, ok := getLastString(k); ok {
        img, err := xgraphics.LoadPngFromFile(v)
        if err != nil {
            logWarning.Printf("Could not load '%s' as a png image because: %v",
                              v, err)
            return
        }
        *place = img
    }
}

func builtInFont() *truetype.Font {
    bs := bindata.DejavusansTtf()
    font, err := freetype.ParseFont(bs)
    if err != nil {
        logWarning.Printf("Could not parse default font because: %v", err)
        return nil
    }
    return font
}

func setFont(k wini.Key, place **truetype.Font) {
    if v, ok := getLastString(k); ok {
        bs, err := ioutil.ReadFile(v)
        if err != nil {
            logWarning.Printf("Could not get font data from '%s' because: %v",
                              v, err)
            return
        }

        font, err := freetype.ParseFont(bs)
        if err != nil {
            logWarning.Printf("Could not parse font data from '%s' because: %v",
                              v, err)
            return
        }

        *place = font
    }
}

