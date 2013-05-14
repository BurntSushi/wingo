package wm

import (
	"code.google.com/p/jamslam-freetype-go/freetype"
	"code.google.com/p/jamslam-freetype-go/freetype/truetype"

	"github.com/BurntSushi/xgbutil/xgraphics"

	"github.com/BurntSushi/wingo-conc/frame"
	"github.com/BurntSushi/wingo-conc/logger"
	"github.com/BurntSushi/wingo-conc/misc"
	"github.com/BurntSushi/wingo-conc/prompt"
	"github.com/BurntSushi/wingo-conc/render"
	"github.com/BurntSushi/wingo-conc/wini"
)

type ThemeConfig struct {
	DefaultIcon *xgraphics.Image
	Full        ThemeFull
	Borders     ThemeBorders
	Slim        ThemeSlim
	Prompt      ThemePrompt
}

type ThemeFull struct {
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

func (tf ThemeFull) FrameTheme() *frame.FullTheme {
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

type ThemeBorders struct {
	borderSize                 int
	aThinColor, iThinColor     render.Color
	aBorderColor, iBorderColor render.Color
}

func (tb ThemeBorders) FrameTheme() *frame.BordersTheme {
	return &frame.BordersTheme{
		BorderSize:   tb.borderSize,
		AThinColor:   tb.aThinColor,
		IThinColor:   tb.iThinColor,
		ABorderColor: tb.aBorderColor,
		IBorderColor: tb.iBorderColor,
	}
}

type ThemeSlim struct {
	borderSize                 int
	aBorderColor, iBorderColor render.Color
}

func (ts ThemeSlim) FrameTheme() *frame.SlimTheme {
	return &frame.SlimTheme{
		BorderSize:   ts.borderSize,
		ABorderColor: ts.aBorderColor,
		IBorderColor: ts.iBorderColor,
	}
}

type ThemePrompt struct {
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

func (tp ThemePrompt) CycleTheme() *prompt.CycleTheme {
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

func (tp ThemePrompt) SelectTheme() *prompt.SelectTheme {
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

func (tp ThemePrompt) InputTheme() *prompt.InputTheme {
	return &prompt.InputTheme{
		BorderSize:  tp.borderSize,
		BgColor:     tp.bgColor,
		BorderColor: tp.borderColor,
		Padding:     tp.padding,
		Font:        tp.font,
		FontSize:    tp.fontSize,
		FontColor:   tp.fontColor,
		InputWidth:  500,
	}
}

func (tp ThemePrompt) MessageTheme() *prompt.MessageTheme {
	return &prompt.MessageTheme{
		BorderSize:  tp.borderSize,
		BgColor:     tp.bgColor,
		BorderColor: tp.borderColor,
		Padding:     tp.padding,
		Font:        tp.font,
		FontSize:    tp.fontSize,
		FontColor:   tp.fontColor,
	}
}

func newTheme() *ThemeConfig {
	return &ThemeConfig{
		DefaultIcon: builtInIcon(),
		Full: ThemeFull{
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

			aCloseButton: builtInButton(misc.ClosePng),
			iCloseButton: builtInButton(misc.ClosePng),
			aCloseColor:  render.NewColor(0xffffff),
			iCloseColor:  render.NewColor(0x000000),

			aMaximizeButton: builtInButton(misc.MaximizePng),
			iMaximizeButton: builtInButton(misc.MaximizePng),
			aMaximizeColor:  render.NewColor(0xffffff),
			iMaximizeColor:  render.NewColor(0x000000),

			aMinimizeButton: builtInButton(misc.MinimizePng),
			iMinimizeButton: builtInButton(misc.MinimizePng),
			aMinimizeColor:  render.NewColor(0xffffff),
			iMinimizeColor:  render.NewColor(0x000000),
		},
		Borders: ThemeBorders{
			borderSize:   10,
			aThinColor:   render.NewColor(0x0),
			iThinColor:   render.NewColor(0x0),
			aBorderColor: render.NewColor(0x3366ff),
			iBorderColor: render.NewColor(0xdfdcdf),
		},
		Slim: ThemeSlim{
			borderSize:   10,
			aBorderColor: render.NewColor(0x3366ff),
			iBorderColor: render.NewColor(0xdfdcdf),
		},
		Prompt: ThemePrompt{
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

func loadTheme() (*ThemeConfig, error) {
	theme := newTheme()

	tdata, err := wini.Parse(misc.ConfigFile("theme.wini"))
	if err != nil {
		return nil, err
	}

	for _, section := range tdata.Sections() {
		switch section {
		case "misc":
			for _, key := range tdata.Keys(section) {
				loadMiscOption(theme, key)
			}
		case "full":
			for _, key := range tdata.Keys(section) {
				loadFullOption(theme, key)
			}
		case "borders":
			for _, key := range tdata.Keys(section) {
				loadBorderOption(theme, key)
			}
		case "slim":
			for _, key := range tdata.Keys(section) {
				loadSlimOption(theme, key)
			}
		case "prompt":
			for _, key := range tdata.Keys(section) {
				loadPromptOption(theme, key)
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
	colorize(theme.Full.aCloseButton, theme.Full.aCloseColor)
	colorize(theme.Full.iCloseButton, theme.Full.iCloseColor)
	colorize(theme.Full.aMaximizeButton, theme.Full.aMaximizeColor)
	colorize(theme.Full.iMaximizeButton, theme.Full.iMaximizeColor)
	colorize(theme.Full.aMinimizeButton, theme.Full.aMinimizeColor)
	colorize(theme.Full.iMinimizeButton, theme.Full.iMinimizeColor)

	// Scale some images...
	theme.Full.aCloseButton = theme.Full.aCloseButton.Scale(
		theme.Full.titleSize, theme.Full.titleSize)
	theme.Full.iCloseButton = theme.Full.iCloseButton.Scale(
		theme.Full.titleSize, theme.Full.titleSize)
	theme.Full.aMaximizeButton = theme.Full.aMaximizeButton.Scale(
		theme.Full.titleSize, theme.Full.titleSize)
	theme.Full.iMaximizeButton = theme.Full.iMaximizeButton.Scale(
		theme.Full.titleSize, theme.Full.titleSize)
	theme.Full.aMinimizeButton = theme.Full.aMinimizeButton.Scale(
		theme.Full.titleSize, theme.Full.titleSize)
	theme.Full.iMinimizeButton = theme.Full.iMinimizeButton.Scale(
		theme.Full.titleSize, theme.Full.titleSize)

	return theme, nil
}

func loadMiscOption(theme *ThemeConfig, k wini.Key) {
	switch k.Name() {
	case "default_icon":
		setImage(k, &theme.DefaultIcon)
	}
}

func loadFullOption(theme *ThemeConfig, k wini.Key) {
	switch k.Name() {
	case "font":
		setFont(k, &theme.Full.font)
	case "font_size":
		setFloat(k, &theme.Full.fontSize)
	case "a_font_color":
		setNoGradient(k, &theme.Full.aFontColor)
	case "i_font_color":
		setNoGradient(k, &theme.Full.iFontColor)
	case "title_size":
		setInt(k, &theme.Full.titleSize)
	case "title_top_margin":
		logger.Warning.Printf("title_top_margin option has been removed.")
	case "a_title_color":
		setGradient(k, &theme.Full.aTitleColor)
	case "i_title_color":
		setGradient(k, &theme.Full.iTitleColor)
	case "close":
		setImage(k, &theme.Full.aCloseButton)
		setImage(k, &theme.Full.iCloseButton)
	case "a_close_color":
		setNoGradient(k, &theme.Full.aCloseColor)
	case "i_close_color":
		setNoGradient(k, &theme.Full.iCloseColor)
	case "maximize":
		setImage(k, &theme.Full.aMaximizeButton)
		setImage(k, &theme.Full.iMaximizeButton)
	case "a_maximize_color":
		setNoGradient(k, &theme.Full.aMaximizeColor)
	case "i_maximize_color":
		setNoGradient(k, &theme.Full.iMaximizeColor)
	case "minimize":
		setImage(k, &theme.Full.aMinimizeButton)
		setImage(k, &theme.Full.iMinimizeButton)
	case "a_minimize_color":
		setNoGradient(k, &theme.Full.aMinimizeColor)
	case "i_minimize_color":
		setNoGradient(k, &theme.Full.iMinimizeColor)
	case "border_size":
		setInt(k, &theme.Full.borderSize)
	case "a_border_color":
		setNoGradient(k, &theme.Full.aBorderColor)
	case "i_border_color":
		setNoGradient(k, &theme.Full.iBorderColor)
	}
}

func loadBorderOption(theme *ThemeConfig, k wini.Key) {
	switch k.Name() {
	case "border_size":
		setInt(k, &theme.Borders.borderSize)
	case "a_thin_color":
		setNoGradient(k, &theme.Borders.aThinColor)
	case "i_thin_color":
		setNoGradient(k, &theme.Borders.iThinColor)
	case "a_border_color":
		setGradient(k, &theme.Borders.aBorderColor)
	case "i_border_color":
		setGradient(k, &theme.Borders.iBorderColor)
	}
}

func loadSlimOption(theme *ThemeConfig, k wini.Key) {
	switch k.Name() {
	case "border_size":
		setInt(k, &theme.Slim.borderSize)
	case "a_border_color":
		setNoGradient(k, &theme.Slim.aBorderColor)
	case "i_border_color":
		setNoGradient(k, &theme.Slim.iBorderColor)
	}
}

func loadPromptOption(theme *ThemeConfig, k wini.Key) {
	switch k.Name() {
	case "bg_color":
		setNoGradient(k, &theme.Prompt.bgColor)
	case "border_color":
		setNoGradient(k, &theme.Prompt.borderColor)
	case "border_size":
		setInt(k, &theme.Prompt.borderSize)
	case "padding":
		setInt(k, &theme.Prompt.padding)
	case "font":
		setFont(k, &theme.Prompt.font)
	case "font_size":
		setFloat(k, &theme.Prompt.fontSize)
	case "font_color":
		setNoGradient(k, &theme.Prompt.fontColor)
	case "cycle_icon_size":
		setInt(k, &theme.Prompt.cycleIconSize)
	case "cycle_icon_border_size":
		setInt(k, &theme.Prompt.cycleIconBorderSize)
	case "cycle_icon_transparency":
		setInt(k, &theme.Prompt.cycleIconTransparency)

		// naughty!
		if theme.Prompt.cycleIconTransparency < 0 ||
			theme.Prompt.cycleIconTransparency > 100 {
			logger.Warning.Printf("Illegal value '%s' provided for " +
				"'cycle_icon_transparency'. Transparency " +
				"values must be in the range [0, 100], " +
				"inclusive. Using 100 by default.")
			theme.Prompt.cycleIconTransparency = 100
		}
	case "select_active_font_color":
		setNoGradient(k, &theme.Prompt.selectActiveFontColor)
	case "select_active_bg_color":
		setNoGradient(k, &theme.Prompt.selectActiveBgColor)
	case "select_group_bg_color":
		setNoGradient(k, &theme.Prompt.selectGroupBgColor)
	case "select_group_font":
		setFont(k, &theme.Prompt.selectGroupFont)
	case "select_group_font_size":
		setFloat(k, &theme.Prompt.selectGroupFontSize)
	case "select_group_font_color":
		setNoGradient(k, &theme.Prompt.selectGroupFontColor)
	}
}

func builtInIcon() *xgraphics.Image {
	img, err := xgraphics.NewBytes(X, misc.WingoPng)
	if err != nil {
		logger.Warning.Printf("Could not get built in icon image because: %v",
			err)
		return nil
	}
	return img
}

func builtInButton(builtInData []byte) *xgraphics.Image {

	img, err := xgraphics.NewBytes(X, builtInData)
	if err != nil {
		logger.Warning.Printf("Could not get built in button image because: %v",
			err)
		return nil
	}
	return img
}

func builtInFont() *truetype.Font {
	font, err := freetype.ParseFont(misc.DejavusansTtf)
	if err != nil {
		logger.Warning.Printf("Could not parse default font because: %v", err)
		return nil
	}
	return font
}
