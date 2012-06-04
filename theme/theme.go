package theme

import (
	"io/ioutil"
	"strconv"
	"strings"

	"code.google.com/p/freetype-go/freetype"
	"code.google.com/p/freetype-go/freetype/truetype"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xgraphics"

	"github.com/BurntSushi/wingo/bindata"
	"github.com/BurntSushi/wingo/config"
	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/misc"
	"github.com/BurntSushi/wingo/wini"
)

type Theme struct {
	DefaultIcon *xgraphics.Image
	Full        Full
	Borders     Borders
	Slim        Slim
	Prompt      Prompt
}

type Full struct {
	Font                   *truetype.Font
	FontSize               float64
	AFontColor, IFontColor int

	TitleSize                int
	ATitleColor, ITitleColor Color

	ACloseButton, ICloseButton *xgraphics.Image
	ACloseColor, ICloseColor   int

	AMaximizeButton, IMaximizeButton *xgraphics.Image
	AMaximizeColor, IMaximizeColor   int

	AMinimizeButton, IMinimizeButton *xgraphics.Image
	AMinimizeColor, IMinimizeColor   int

	BorderSize                 int
	ABorderColor, IBorderColor int
}

type Borders struct {
	BorderSize                 int
	AThinColor, IThinColor     int
	ABorderColor, IBorderColor Color
}

type Slim struct {
	BorderSize                 int
	ABorderColor, IBorderColor int
}

type Prompt struct {
	BgColor     int
	BorderColor int
	BorderSize  int
	Padding     int

	Font      *truetype.Font
	FontSize  float64
	FontColor int

	CycleIconSize         int
	CycleIconBorderSize   int
	CycleIconTransparency int

	SelectActiveColor   int
	SelectActiveBgColor int
	SelectLabelColor    int
	SelectLabelFontSize float64
}

func newTheme(X *xgbutil.XUtil) *Theme {
	return &Theme{
		DefaultIcon: builtInIcon(X),
		Full: Full{
			Font:       builtInFont(),
			FontSize:   15,
			AFontColor: 0xffffff,
			IFontColor: 0x000000,

			TitleSize:   25,
			ATitleColor: NewColor(0x3366ff),
			ITitleColor: NewColor(0xdfdcdf),

			ACloseButton: builtInButton(X, bindata.ClosePng),
			ICloseButton: builtInButton(X, bindata.ClosePng),
			ACloseColor:  0xffffff,
			ICloseColor:  0x000000,

			AMaximizeButton: builtInButton(X, bindata.MaximizePng),
			IMaximizeButton: builtInButton(X, bindata.MaximizePng),
			AMaximizeColor:  0xffffff,
			IMaximizeColor:  0x000000,

			AMinimizeButton: builtInButton(X, bindata.MinimizePng),
			IMinimizeButton: builtInButton(X, bindata.MinimizePng),
			AMinimizeColor:  0xffffff,
			IMinimizeColor:  0x000000,

			BorderSize:   10,
			ABorderColor: 0x3366ff,
			IBorderColor: 0xdfdcdf,
		},
		Borders: Borders{
			BorderSize:   10,
			AThinColor:   0x0,
			IThinColor:   0x0,
			ABorderColor: NewColor(0x3366ff),
			IBorderColor: NewColor(0xdfdcdf),
		},
		Slim: Slim{
			BorderSize:   10,
			ABorderColor: 0x3366ff,
			IBorderColor: 0xdfdcdf,
		},
		Prompt: Prompt{
			BgColor:               0xffffff,
			BorderColor:           0x585a5d,
			BorderSize:            10,
			Padding:               10,
			Font:                  builtInFont(),
			FontSize:              15,
			FontColor:             0x000000,
			CycleIconSize:         32,
			CycleIconBorderSize:   3,
			CycleIconTransparency: 50,
			SelectActiveColor:     0x000000,
			SelectActiveBgColor:   0xffffff,
			SelectLabelColor:      0xffffff,
			SelectLabelFontSize:   25,
		},
	}
}

func LoadTheme(X *xgbutil.XUtil) (*Theme, error) {
	theme := newTheme(X)

	tdata, err := loadThemeFile()
	if err != nil {
		return nil, err
	}

	for _, section := range tdata.Sections() {
		switch section {
		case "misc":
			for _, key := range tdata.Keys(section) {
				loadMiscOption(X, theme, key)
			}
		case "full":
			for _, key := range tdata.Keys(section) {
				loadFullOption(X, theme, key)
			}
		case "borders":
			for _, key := range tdata.Keys(section) {
				loadBorderOption(X, theme, key)
			}
		case "slim":
			for _, key := range tdata.Keys(section) {
				loadSlimOption(X, theme, key)
			}
		case "prompt":
			for _, key := range tdata.Keys(section) {
				loadPromptOption(X, theme, key)
			}
		}
	}

	// re-color some images
	colorize := func(im *xgraphics.Image, clr int) {
		var i int
		r8, g8, b8 := misc.RGBFromInt(clr)
		r, g, b := uint8(r8), uint8(g8), uint8(b8)
		im.ForExp(func(x, y int) (uint8, uint8, uint8, uint8) {
			i = im.PixOffset(x, y)
			return r, g, b, im.Pix[i+3]
		})
	}
	colorize(theme.Full.ACloseButton, theme.Full.ACloseColor)
	colorize(theme.Full.ICloseButton, theme.Full.ICloseColor)
	colorize(theme.Full.AMaximizeButton, theme.Full.AMaximizeColor)
	colorize(theme.Full.IMaximizeButton, theme.Full.IMaximizeColor)
	colorize(theme.Full.AMinimizeButton, theme.Full.AMinimizeColor)
	colorize(theme.Full.IMinimizeButton, theme.Full.IMinimizeColor)

	// Scale some images...
	theme.Full.ACloseButton = theme.Full.ACloseButton.Scale(
		theme.Full.TitleSize, theme.Full.TitleSize)
	theme.Full.ICloseButton = theme.Full.ICloseButton.Scale(
		theme.Full.TitleSize, theme.Full.TitleSize)
	theme.Full.AMaximizeButton = theme.Full.AMaximizeButton.Scale(
		theme.Full.TitleSize, theme.Full.TitleSize)
	theme.Full.IMaximizeButton = theme.Full.IMaximizeButton.Scale(
		theme.Full.TitleSize, theme.Full.TitleSize)
	theme.Full.AMinimizeButton = theme.Full.AMinimizeButton.Scale(
		theme.Full.TitleSize, theme.Full.TitleSize)
	theme.Full.IMinimizeButton = theme.Full.IMinimizeButton.Scale(
		theme.Full.TitleSize, theme.Full.TitleSize)

	return theme, nil
}

func loadThemeFile() (*wini.Data, error) {
	return wini.Parse("configdata/theme.wini")
}

func loadMiscOption(X *xgbutil.XUtil, theme *Theme, k wini.Key) {
	switch k.Name() {
	case "default_icon":
		setImage(X, k, &theme.DefaultIcon)
	}
}

func loadFullOption(X *xgbutil.XUtil, theme *Theme, k wini.Key) {
	switch k.Name() {
	case "font":
		setFont(k, &theme.Full.Font)
	case "font_size":
		config.SetFloat(k, &theme.Full.FontSize)
	case "a_font_color":
		config.SetInt(k, &theme.Full.AFontColor)
	case "i_font_color":
		config.SetInt(k, &theme.Full.IFontColor)
	case "title_size":
		config.SetInt(k, &theme.Full.TitleSize)
	case "a_title_color":
		setGradient(k, &theme.Full.ATitleColor)
	case "i_title_color":
		setGradient(k, &theme.Full.ITitleColor)
	case "close":
		setImage(X, k, &theme.Full.ACloseButton)
		setImage(X, k, &theme.Full.ICloseButton)
	case "a_close_color":
		config.SetInt(k, &theme.Full.ACloseColor)
	case "i_close_color":
		config.SetInt(k, &theme.Full.ICloseColor)
	case "maximize":
		setImage(X, k, &theme.Full.AMaximizeButton)
		setImage(X, k, &theme.Full.IMaximizeButton)
	case "a_maximize_color":
		config.SetInt(k, &theme.Full.AMaximizeColor)
	case "i_maximize_color":
		config.SetInt(k, &theme.Full.IMaximizeColor)
	case "minimize":
		setImage(X, k, &theme.Full.AMinimizeButton)
		setImage(X, k, &theme.Full.IMinimizeButton)
	case "a_minimize_color":
		config.SetInt(k, &theme.Full.AMinimizeColor)
	case "i_minimize_color":
		config.SetInt(k, &theme.Full.IMinimizeColor)
	case "border_size":
		config.SetInt(k, &theme.Full.BorderSize)
	case "a_border_color":
		config.SetInt(k, &theme.Full.ABorderColor)
	case "i_border_color":
		config.SetInt(k, &theme.Full.IBorderColor)
	}
}

func loadBorderOption(X *xgbutil.XUtil, theme *Theme, k wini.Key) {
	switch k.Name() {
	case "border_size":
		config.SetInt(k, &theme.Borders.BorderSize)
	case "a_thin_color":
		config.SetInt(k, &theme.Borders.AThinColor)
	case "i_thin_color":
		config.SetInt(k, &theme.Borders.IThinColor)
	case "a_border_color":
		setGradient(k, &theme.Borders.ABorderColor)
	case "i_border_color":
		setGradient(k, &theme.Borders.IBorderColor)
	}
}

func loadSlimOption(X *xgbutil.XUtil, theme *Theme, k wini.Key) {
	switch k.Name() {
	case "border_size":
		config.SetInt(k, &theme.Slim.BorderSize)
	case "a_border_color":
		config.SetInt(k, &theme.Slim.ABorderColor)
	case "i_border_color":
		config.SetInt(k, &theme.Slim.IBorderColor)
	}
}

func loadPromptOption(X *xgbutil.XUtil, theme *Theme, k wini.Key) {
	switch k.Name() {
	case "bg_color":
		config.SetInt(k, &theme.Prompt.BgColor)
	case "border_color":
		config.SetInt(k, &theme.Prompt.BorderColor)
	case "border_size":
		config.SetInt(k, &theme.Prompt.BorderSize)
	case "padding":
		config.SetInt(k, &theme.Prompt.Padding)
	case "font":
		setFont(k, &theme.Prompt.Font)
	case "font_size":
		config.SetFloat(k, &theme.Prompt.FontSize)
	case "font_color":
		config.SetInt(k, &theme.Prompt.FontColor)
	case "cycle_icon_size":
		config.SetInt(k, &theme.Prompt.CycleIconSize)
	case "cycle_icon_border_size":
		config.SetInt(k, &theme.Prompt.CycleIconBorderSize)
	case "cycle_icon_transparency":
		config.SetInt(k, &theme.Prompt.CycleIconTransparency)

		// naughty!
		if theme.Prompt.CycleIconTransparency < 0 ||
			theme.Prompt.CycleIconTransparency > 100 {
			logger.Warning.Printf("Illegal value '%s' provided for " +
				"'cycle_icon_transparency'. Transparency " +
				"values must be in the range [0, 100], " +
				"inclusive. Using 100 by default.")
			theme.Prompt.CycleIconTransparency = 100
		}
	case "select_active_color":
		config.SetInt(k, &theme.Prompt.SelectActiveColor)
	case "select_active_bg_color":
		config.SetInt(k, &theme.Prompt.SelectActiveBgColor)
	case "select_label_color":
		config.SetInt(k, &theme.Prompt.SelectLabelColor)
	case "select_label_font_size":
		config.SetFloat(k, &theme.Prompt.SelectLabelFontSize)
	}
}

func setGradient(k wini.Key, clr *Color) {
	// Check to make sure we have a value for this key
	vals := k.Strings()
	if len(vals) == 0 {
		logger.Warning.Println(k.Err("No values found."))
		return
	}

	// Use the last value
	val := vals[len(vals)-1]

	// If there are no spaces, it can't be a gradient.
	if strings.Index(val, " ") == -1 {
		if start, ok := config.GetLastInt(k); ok {
			clr.Start = start
		}
		return
	}

	// Okay, now we have to do things manually.
	// Split up the value into two pieces separated by whitespace and parse
	// each piece as an int.
	splitted := strings.Split(val, " ")
	if len(splitted) != 2 {
		logger.Warning.Println(k.Err("Expected a gradient value (two colors "+
			"separated by a space), but found '%s' "+
			"instead.", val))
		return
	}

	start, err := strconv.ParseInt(strings.TrimSpace(splitted[0]), 0, 0)
	if err != nil {
		logger.Warning.Println(k.Err("'%s' is not an integer. (%s)",
			splitted[0], err))
		return
	}

	end, err := strconv.ParseInt(strings.TrimSpace(splitted[1]), 0, 0)
	if err != nil {
		logger.Warning.Println(k.Err("'%s' is not an integer. (%s)",
			splitted[1], err))
		return
	}

	// finally...
	clr.Start, clr.End = int(start), int(end)
}

func builtInIcon(X *xgbutil.XUtil) *xgraphics.Image {
	img, err := xgraphics.NewBytes(X, bindata.WingoPng())
	if err != nil {
		logger.Warning.Printf("Could not get built in icon image because: %v",
			err)
		return nil
	}
	return img
}

func builtInButton(X *xgbutil.XUtil,
	loadBuiltIn func() []byte) *xgraphics.Image {

	img, err := xgraphics.NewBytes(X, loadBuiltIn())
	if err != nil {
		logger.Warning.Printf("Could not get built in button image because: %v",
			err)
		return nil
	}
	return img
}

func setImage(X *xgbutil.XUtil, k wini.Key, place **xgraphics.Image) {
	if v, ok := config.GetLastString(k); ok {
		img, err := xgraphics.NewFileName(X, v)
		if err != nil {
			logger.Warning.Printf(
				"Could not load '%s' as a png image because: %v", v, err)
			return
		}
		*place = img
	}
}

func builtInFont() *truetype.Font {
	bs := bindata.DejavusansTtf()
	font, err := freetype.ParseFont(bs)
	if err != nil {
		logger.Warning.Printf("Could not parse default font because: %v", err)
		return nil
	}
	return font
}

func setFont(k wini.Key, place **truetype.Font) {
	if v, ok := config.GetLastString(k); ok {
		bs, err := ioutil.ReadFile(v)
		if err != nil {
			logger.Warning.Printf(
				"Could not get font data from '%s' because: %v", v, err)
			return
		}

		font, err := freetype.ParseFont(bs)
		if err != nil {
			logger.Warning.Printf(
				"Could not parse font data from '%s' because: %v", v, err)
			return
		}

		*place = font
	}
}
