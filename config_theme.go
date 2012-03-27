package main

import (
    "image/color"
    "image/draw"
    "io/ioutil"
    "strconv"
    "strings"
)

import (
    "code.google.com/p/freetype-go/freetype"
    "code.google.com/p/freetype-go/freetype/truetype"
)

// import "code.google.com/p/jamslam-x-go-binding/xgb" 

import "github.com/BurntSushi/xgbutil/xgraphics"

import (
    "github.com/BurntSushi/wingo/bindata"
    "github.com/BurntSushi/wingo/wini"
)

type theme struct {
    defaultIcon string
    full themeFull
    borders themeBorders
    slim themeSlim
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

func defaultTheme() *theme {
    return &theme{
        full: themeFull{
            font: getBuiltInFont(),
            fontSize: 15,
            aFontColor: 0xffffff,
            iFontColor: 0x000000,

            titleSize: 25,
            aTitleColor: newThemeColor(0x3366ff),
            iTitleColor: newThemeColor(0xdfdcdf),

            aCloseButton: getBuiltInButton(bindata.ClosePng),
            iCloseButton: getBuiltInButton(bindata.ClosePng),
            aCloseColor: 0xffffff,
            iCloseColor: 0x000000,

            aMaximizeButton: getBuiltInButton(bindata.MaximizePng),
            iMaximizeButton: getBuiltInButton(bindata.MaximizePng),
            aMaximizeColor: 0xffffff,
            iMaximizeColor: 0x000000,

            aMinimizeButton: getBuiltInButton(bindata.MinimizePng),
            iMinimizeButton: getBuiltInButton(bindata.MinimizePng),
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
        }
    }

    // re-color some images
    xgraphics.ColorImage(THEME.full.aCloseButton,
                         ColorFromInt(THEME.full.aCloseColor))
    xgraphics.ColorImage(THEME.full.iCloseButton,
                         ColorFromInt(THEME.full.iCloseColor))
    xgraphics.ColorImage(THEME.full.aMaximizeButton,
                         ColorFromInt(THEME.full.aMaximizeColor))
    xgraphics.ColorImage(THEME.full.iMaximizeButton,
                         ColorFromInt(THEME.full.iMaximizeColor))
    xgraphics.ColorImage(THEME.full.aMinimizeButton,
                         ColorFromInt(THEME.full.aMinimizeColor))
    xgraphics.ColorImage(THEME.full.iMinimizeButton,
                         ColorFromInt(THEME.full.iMinimizeColor))

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
    return wini.Parse("theme.wini")
}

func loadMiscOption(k wini.Key) {
    strs := k.Strings()

    switch k.Name() {
    case "default_icon":
        THEME.defaultIcon = strs[len(strs) - 1]
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

    rinc := float64(er - sr) / float64(size)
    ginc := float64(eg - sg) / float64(size)
    binc := float64(eb - sb) / float64(size)

    doInc := func(inc float64, start, index int) int {
        return int(float64(start) + inc * float64(index))
    }

    var nr, ng, nb int
    for i := 1; i < size - 1; i++ {
        nr = doInc(rinc, sr, i)
        ng = doInc(ginc, sg, i)
        nb = doInc(binc, sb, i)

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

func getBuiltInButton(loadBuiltIn func() []byte) draw.Image {
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

func getBuiltInFont() *truetype.Font {
    bs := bindata.RobotoregularTtf()
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

func setString(k wini.Key, place *string) {
    if v, ok := getLastString(k); ok {
        *place = v
    }
}

func getLastString(k wini.Key) (string, bool) {
    vals := k.Strings()
    if len(vals) == 0 {
        logWarning.Println(k.Err("No values found."))
        return "", false
    }

    return vals[len(vals) - 1], true
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

func setFloat(k wini.Key, place *float64) {
    if v, ok := getLastFloat(k); ok {
        *place = float64(v)
    }
}

func getLastFloat(k wini.Key) (float64, bool) {
    vals, err := k.Floats()
    if err != nil {
        logWarning.Println(err)
        return 0.0, false
    } else if len(vals) == 0 {
        logWarning.Println(k.Err("No values found."))
        return 0.0, false
    }

    return vals[len(vals) - 1], true
}

