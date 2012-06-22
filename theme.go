package main

import (
	"code.google.com/p/freetype-go/freetype"
	"code.google.com/p/freetype-go/freetype/truetype"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xgraphics"

	"github.com/BurntSushi/wingo/bindata"
	"github.com/BurntSushi/wingo/frame"
	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/prompt"
	"github.com/BurntSushi/wingo/render"
	"github.com/BurntSushi/wingo/wini"
)

type theme struct {
	defaultIcon *xgraphics.Image
	full        themeFull
	borders     themeBorders
	slim        themeSlim
	prompt      themePrompt
}

type themeFull struct {
	font                   *truetype.Font
	fontSize               float64
	aFontColor, iFontColor render.Color

	titleSize                int
	aTitleColor, iTitleColor render.Color

	borderSize                 int
	aBorderColor, iBorderColor render.Color

	aCloseButton, iCloseButton *xgraphics.Image
	aCloseColor, iCloseColor   render.Color

	aMaximizeButton, iMaximizeButton *xgraphics.Image
	aMaximizeColor, iMaximizeColor   render.Color

	aMinimizeButton, iMinimizeButton *xgraphics.Image
	aMinimizeColor, iMinimizeColor   render.Color
}

func (tf themeFull) FrameTheme() *frame.FullTheme {
	return &frame.FullTheme{
		Font:            tf.font,
		FontSize:        tf.fontSize,
		AFontColor:      tf.aFontColor,
		IFontColor:      tf.iFontColor,
		TitleSize:       tf.titleSize,
		ATitleColor:     tf.aTitleColor,
		ITitleColor:     tf.iTitleColor,
		BorderSize:      tf.borderSize,
		ABorderColor:    tf.aBorderColor,
		IBorderColor:    tf.iBorderColor,
		ACloseButton:    tf.aCloseButton,
		ICloseButton:    tf.iCloseButton,
		AMaximizeButton: tf.aMaximizeButton,
		IMaximizeButton: tf.iMaximizeButton,
		AMinimizeButton: tf.aMinimizeButton,
		IMinimizeButton: tf.iMinimizeButton,
	}
}

type themeBorders struct {
	borderSize                 int
	aThinColor, iThinColor     render.Color
	aBorderColor, iBorderColor render.Color
}

func (tb themeBorders) FrameTheme() *frame.BordersTheme {
	return &frame.BordersTheme{
		BorderSize:   tb.borderSize,
		AThinColor:   tb.aThinColor,
		IThinColor:   tb.iThinColor,
		ABorderColor: tb.aBorderColor,
		IBorderColor: tb.iBorderColor,
	}
}

type themeSlim struct {
	borderSize                 int
	aBorderColor, iBorderColor render.Color
}

func (ts themeSlim) FrameTheme() *frame.SlimTheme {
	return &frame.SlimTheme{
		BorderSize:   ts.borderSize,
		ABorderColor: ts.aBorderColor,
		IBorderColor: ts.iBorderColor,
	}
}

type themePrompt struct {
	bgColor     render.Color
	borderColor render.Color
	borderSize  int
	padding     int

	font      *truetype.Font
	fontSize  float64
	fontColor render.Color

	cycleIconSize         int
	cycleIconBorderSize   int
	cycleIconTransparency int

	selectActiveBgColor   render.Color
	selectActiveFontColor render.Color

	selectGroupBgColor   render.Color
	selectGroupFont      *truetype.Font
	selectGroupFontSize  float64
	selectGroupFontColor render.Color
}

func (tp themePrompt) CycleTheme() *prompt.CycleTheme {
	return &prompt.CycleTheme{
		BorderSize:       tp.borderSize,
		BgColor:          tp.bgColor,
		BorderColor:      tp.borderColor,
		Padding:          tp.padding,
		Font:             tp.font,
		FontSize:         tp.fontSize,
		FontColor:        tp.fontColor,
		IconSize:         tp.cycleIconSize,
		IconBorderSize:   tp.cycleIconBorderSize,
		IconTransparency: tp.cycleIconTransparency,
	}
}

func (tp themePrompt) SelectTheme() *prompt.SelectTheme {
	return &prompt.SelectTheme{
		BorderSize:      tp.borderSize,
		BgColor:         tp.bgColor,
		BorderColor:     tp.borderColor,
		Padding:         tp.padding,
		Font:            tp.font,
		FontSize:        tp.fontSize,
		FontColor:       tp.fontColor,
		ActiveBgColor:   tp.selectActiveBgColor,
		ActiveFontColor: tp.selectActiveFontColor,
		GroupBgColor:    tp.selectGroupBgColor,
		GroupFont:       tp.selectGroupFont,
		GroupFontSize:   tp.selectGroupFontSize,
		GroupFontColor:  tp.selectGroupFontColor,
		GroupSpacing:    15,
	}
}

func newTheme(X *xgbutil.XUtil) *theme {
	return &theme{
		defaultIcon: builtInIcon(X),
		full: themeFull{
			font:       builtInFont(),
			fontSize:   15,
			aFontColor: render.NewColor(0xffffff),
			iFontColor: render.NewColor(0x000000),

			titleSize:   25,
			aTitleColor: render.NewColor(0x3366ff),
			iTitleColor: render.NewColor(0xdfdcdf),

			borderSize:   10,
			aBorderColor: render.NewColor(0x3366ff),
			iBorderColor: render.NewColor(0xdfdcdf),

			aCloseButton: builtInButton(X, bindata.ClosePng),
			iCloseButton: builtInButton(X, bindata.ClosePng),
			aCloseColor:  render.NewColor(0xffffff),
			iCloseColor:  render.NewColor(0x000000),

			aMaximizeButton: builtInButton(X, bindata.MaximizePng),
			iMaximizeButton: builtInButton(X, bindata.MaximizePng),
			aMaximizeColor:  render.NewColor(0xffffff),
			iMaximizeColor:  render.NewColor(0x000000),

			aMinimizeButton: builtInButton(X, bindata.MinimizePng),
			iMinimizeButton: builtInButton(X, bindata.MinimizePng),
			aMinimizeColor:  render.NewColor(0xffffff),
			iMinimizeColor:  render.NewColor(0x000000),
		},
		borders: themeBorders{
			borderSize:   10,
			aThinColor:   render.NewColor(0x0),
			iThinColor:   render.NewColor(0x0),
			aBorderColor: render.NewColor(0x3366ff),
			iBorderColor: render.NewColor(0xdfdcdf),
		},
		slim: themeSlim{
			borderSize:   10,
			aBorderColor: render.NewColor(0x3366ff),
			iBorderColor: render.NewColor(0xdfdcdf),
		},
		prompt: themePrompt{
			bgColor:               render.NewColor(0xffffff),
			borderColor:           render.NewColor(0x585a5d),
			borderSize:            10,
			padding:               10,
			font:                  builtInFont(),
			fontSize:              15.0,
			fontColor:             render.NewColor(0x000000),
			cycleIconSize:         32,
			cycleIconBorderSize:   3,
			cycleIconTransparency: 50,
			selectActiveBgColor:   render.NewColor(0xffffff),
			selectActiveFontColor: render.NewColor(0x000000),
			selectGroupBgColor:    render.NewColor(0xffffff),
			selectGroupFont:       builtInFont(),
			selectGroupFontSize:   25.0,
			selectGroupFontColor:  render.NewColor(0x0),
		},
	}
}

func loadTheme(X *xgbutil.XUtil) (*theme, error) {
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
	colorize := func(im *xgraphics.Image, clr render.Color) {
		var i int
		r, g, b := clr.RGB8()
		im.ForExp(func(x, y int) (uint8, uint8, uint8, uint8) {
			i = im.PixOffset(x, y)
			return r, g, b, im.Pix[i+3]
		})
	}
	colorize(theme.full.aCloseButton, theme.full.aCloseColor)
	colorize(theme.full.iCloseButton, theme.full.iCloseColor)
	colorize(theme.full.aMaximizeButton, theme.full.aMaximizeColor)
	colorize(theme.full.iMaximizeButton, theme.full.iMaximizeColor)
	colorize(theme.full.aMinimizeButton, theme.full.aMinimizeColor)
	colorize(theme.full.iMinimizeButton, theme.full.iMinimizeColor)

	// Scale some images...
	theme.full.aCloseButton = theme.full.aCloseButton.Scale(
		theme.full.titleSize, theme.full.titleSize)
	theme.full.iCloseButton = theme.full.iCloseButton.Scale(
		theme.full.titleSize, theme.full.titleSize)
	theme.full.aMaximizeButton = theme.full.aMaximizeButton.Scale(
		theme.full.titleSize, theme.full.titleSize)
	theme.full.iMaximizeButton = theme.full.iMaximizeButton.Scale(
		theme.full.titleSize, theme.full.titleSize)
	theme.full.aMinimizeButton = theme.full.aMinimizeButton.Scale(
		theme.full.titleSize, theme.full.titleSize)
	theme.full.iMinimizeButton = theme.full.iMinimizeButton.Scale(
		theme.full.titleSize, theme.full.titleSize)

	return theme, nil
}

func loadThemeFile() (*wini.Data, error) {
	return wini.Parse("config/theme.wini")
}

func loadMiscOption(X *xgbutil.XUtil, theme *theme, k wini.Key) {
	switch k.Name() {
	case "default_icon":
		setImage(X, k, &theme.defaultIcon)
	}
}

func loadFullOption(X *xgbutil.XUtil, theme *theme, k wini.Key) {
	switch k.Name() {
	case "font":
		setFont(k, &theme.full.font)
	case "font_size":
		setFloat(k, &theme.full.fontSize)
	case "a_font_color":
		setNoGradient(k, &theme.full.aFontColor)
	case "i_font_color":
		setNoGradient(k, &theme.full.iFontColor)
	case "title_size":
		setInt(k, &theme.full.titleSize)
	case "a_title_color":
		setGradient(k, &theme.full.aTitleColor)
	case "i_title_color":
		setGradient(k, &theme.full.iTitleColor)
	case "close":
		setImage(X, k, &theme.full.aCloseButton)
		setImage(X, k, &theme.full.iCloseButton)
	case "a_close_color":
		setNoGradient(k, &theme.full.aCloseColor)
	case "i_close_color":
		setNoGradient(k, &theme.full.iCloseColor)
	case "maximize":
		setImage(X, k, &theme.full.aMaximizeButton)
		setImage(X, k, &theme.full.iMaximizeButton)
	case "a_maximize_color":
		setNoGradient(k, &theme.full.aMaximizeColor)
	case "i_maximize_color":
		setNoGradient(k, &theme.full.iMaximizeColor)
	case "minimize":
		setImage(X, k, &theme.full.aMinimizeButton)
		setImage(X, k, &theme.full.iMinimizeButton)
	case "a_minimize_color":
		setNoGradient(k, &theme.full.aMinimizeColor)
	case "i_minimize_color":
		setNoGradient(k, &theme.full.iMinimizeColor)
	case "border_size":
		setInt(k, &theme.full.borderSize)
	case "a_border_color":
		setNoGradient(k, &theme.full.aBorderColor)
	case "i_border_color":
		setNoGradient(k, &theme.full.iBorderColor)
	}
}

func loadBorderOption(X *xgbutil.XUtil, theme *theme, k wini.Key) {
	switch k.Name() {
	case "border_size":
		setInt(k, &theme.borders.borderSize)
	case "a_thin_color":
		setNoGradient(k, &theme.borders.aThinColor)
	case "i_thin_color":
		setNoGradient(k, &theme.borders.iThinColor)
	case "a_border_color":
		setGradient(k, &theme.borders.aBorderColor)
	case "i_border_color":
		setGradient(k, &theme.borders.iBorderColor)
	}
}

func loadSlimOption(X *xgbutil.XUtil, theme *theme, k wini.Key) {
	switch k.Name() {
	case "border_size":
		setInt(k, &theme.slim.borderSize)
	case "a_border_color":
		setNoGradient(k, &theme.slim.aBorderColor)
	case "i_border_color":
		setNoGradient(k, &theme.slim.iBorderColor)
	}
}

func loadPromptOption(X *xgbutil.XUtil, theme *theme, k wini.Key) {
	switch k.Name() {
	case "bg_color":
		setNoGradient(k, &theme.prompt.bgColor)
	case "border_color":
		setNoGradient(k, &theme.prompt.borderColor)
	case "border_size":
		setInt(k, &theme.prompt.borderSize)
	case "padding":
		setInt(k, &theme.prompt.padding)
	case "font":
		setFont(k, &theme.prompt.font)
	case "font_size":
		setFloat(k, &theme.prompt.fontSize)
	case "font_color":
		setNoGradient(k, &theme.prompt.fontColor)
	case "cycle_icon_size":
		setInt(k, &theme.prompt.cycleIconSize)
	case "cycle_icon_border_size":
		setInt(k, &theme.prompt.cycleIconBorderSize)
	case "cycle_icon_transparency":
		setInt(k, &theme.prompt.cycleIconTransparency)

		// naughty!
		if theme.prompt.cycleIconTransparency < 0 ||
			theme.prompt.cycleIconTransparency > 100 {
			logger.Warning.Printf("Illegal value '%s' provided for " +
				"'cycle_icon_transparency'. Transparency " +
				"values must be in the range [0, 100], " +
				"inclusive. Using 100 by default.")
			theme.prompt.cycleIconTransparency = 100
		}
	case "select_active_font_color":
		setNoGradient(k, &theme.prompt.selectActiveFontColor)
	case "select_active_bg_color":
		setNoGradient(k, &theme.prompt.selectActiveBgColor)
	case "select_group_bg_color":
		setNoGradient(k, &theme.prompt.selectGroupBgColor)
	case "select_group_font":
		setFont(k, &theme.prompt.selectGroupFont)
	case "select_group_font_size":
		setFloat(k, &theme.prompt.selectGroupFontSize)
	case "select_group_font_color":
		setNoGradient(k, &theme.prompt.selectGroupFontColor)
	}
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

func builtInFont() *truetype.Font {
	bs := bindata.DejavusansTtf()
	font, err := freetype.ParseFont(bs)
	if err != nil {
		logger.Warning.Printf("Could not parse default font because: %v", err)
		return nil
	}
	return font
}
