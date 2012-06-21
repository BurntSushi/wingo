package main

import (
	"io/ioutil"
	"strconv"
	"strings"

	"code.google.com/p/freetype-go/freetype"
	"code.google.com/p/freetype-go/freetype/truetype"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xgraphics"

	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/render"
	"github.com/BurntSushi/wingo/wini"
)

func setString(k wini.Key, place *string) {
	if v, ok := getLastString(k); ok {
		*place = v
	}
}

func getLastString(k wini.Key) (string, bool) {
	vals := k.Strings()
	if len(vals) == 0 {
		logger.Warning.Println(k.Err("No values found."))
		return "", false
	}

	return vals[len(vals)-1], true
}

func setBool(k wini.Key, place *bool) {
	if v, ok := getLastBool(k); ok {
		*place = v
	}
}

func getLastBool(k wini.Key) (bool, bool) {
	vals, err := k.Bools()
	if err != nil {
		logger.Warning.Println(err)
		return false, false
	} else if len(vals) == 0 {
		logger.Warning.Println(k.Err("No values found."))
		return false, false
	}

	return vals[len(vals)-1], true
}

func setInt(k wini.Key, place *int) {
	if v, ok := getLastInt(k); ok {
		*place = int(v)
	}
}

func getLastInt(k wini.Key) (int, bool) {
	vals, err := k.Ints()
	if err != nil {
		logger.Warning.Println(err)
		return 0, false
	} else if len(vals) == 0 {
		logger.Warning.Println(k.Err("No values found."))
		return 0, false
	}

	return vals[len(vals)-1], true
}

func setFloat(k wini.Key, place *float64) {
	if v, ok := getLastFloat(k); ok {
		*place = float64(v)
	}
}

func getLastFloat(k wini.Key) (float64, bool) {
	vals, err := k.Floats()
	if err != nil {
		logger.Warning.Println(err)
		return 0.0, false
	} else if len(vals) == 0 {
		logger.Warning.Println(k.Err("No values found."))
		return 0.0, false
	}

	return vals[len(vals)-1], true
}

func setNoGradient(k wini.Key, clr *render.Color) {
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
		if start, ok := getLastInt(k); ok {
			clr.Start = start
		}
		return
	}

	logger.Warning.Printf(
		k.Err("Gradients are not supported for this theme option."))
}

func setGradient(k wini.Key, clr *render.Color) {
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
		if start, ok := getLastInt(k); ok {
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

func setImage(X *xgbutil.XUtil, k wini.Key, place **xgraphics.Image) {
	if v, ok := getLastString(k); ok {
		img, err := xgraphics.NewFileName(X, v)
		if err != nil {
			logger.Warning.Printf(
				"Could not load '%s' as a png image because: %v", v, err)
			return
		}
		*place = img
	}
}

func setFont(k wini.Key, place **truetype.Font) {
	if v, ok := getLastString(k); ok {
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
