package commands

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/gribble"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil/xrect"

	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/prompt"
	"github.com/BurntSushi/wingo/workspace"
	"github.com/BurntSushi/wingo/wm"
	"github.com/BurntSushi/wingo/xclient"
)

// parsePos takes a string and parses an x or y position from it.
// The magic here is that while a string could just be a simple integer,
// it could also be a float greater than 0 but <= 1 in terms of the current
// head's geometry.
func parsePos(geom xrect.Rect, gribblePos gribble.Any, y bool) (int, bool) {
	switch pos := gribblePos.(type) {
	case int:
		return pos, true
	case float64:
		if pos <= 0 || pos > 1 {
			logger.Warning.Printf("'%s' not in the valid range (0, 1].", pos)
			return 0, false
		}

		if y {
			return geom.Y() + int(float64(geom.Height())*pos), true
		} else {
			return geom.X() + int(float64(geom.Width())*pos), true
		}
	}
	panic("unreachable")
}

// parseDim takes a string and parses a width or height dimension from it.
// The magic here is that while a string could just be a simple integer,
// it could also be a float greater than 0 but <= 1 in terms of the current
// head's geometry.
func parseDim(geom xrect.Rect, gribbleDim gribble.Any, hght bool) (int, bool) {
	switch dim := gribbleDim.(type) {
	case int:
		return dim, true
	case float64:
		if dim <= 0 || dim > 1 {
			logger.Warning.Printf("'%s' not in the valid range (0, 1].", dim)
			return 0, false
		}

		if hght {
			return int(float64(geom.Height()) * dim), true
		} else {
			return int(float64(geom.Width()) * dim), true
		}
	}
	panic("unreachable")
}

// stringBool takes a string and returns true if the string corresponds
// to a "true" value. i.e., "Yes", "Y", "y", "YES", "yEs", etc.
func stringBool(s string) bool {
	sl := strings.ToLower(s)
	return sl == "yes" || sl == "y"
}

// boolToInt converts a boolean value to an integer. (True = 1 and False = 0.)
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// intToBool converts an integer value to a boolean.
// An integer other than 0 or 1 causes a panic.
func intToBool(n int) bool {
	switch n {
	case 0:
		return false
	case 1:
		return true
	}
	panic(fmt.Sprintf("Unexpected boolean integer: %d", n))
}

// stringTabComp takes a string and converts it to a tab completion constant
// defined in the prompt package. Valid values are "Prefix", "Any" and
// "Multiple"
func stringTabComp(s string) int {
	switch s {
	case "Prefix":
		return prompt.TabCompletePrefix
	case "Any":
		return prompt.TabCompleteAny
	case "Multiple":
		return prompt.TabCompleteMultiple
	default:
		logger.Warning.Printf(
			"Tab completion mode '%s' not supported.", s)
	}
	return prompt.TabCompletePrefix
}

// Shortcut for executing Client interface functions that have no parameters
// and no return values on the currently focused window.
func withFocused(f func(c *xclient.Client)) gribble.Any {
	if focused := wm.LastFocused(); focused != nil {
		client := focused.(*xclient.Client)
		f(client)
		return int(client.Id())
	}
	return ":void:"
}

func withClient(cArg gribble.Any, f func(c *xclient.Client)) gribble.Any {
	switch c := cArg.(type) {
	case int:
		if c == 0 {
			return ":void:"
		}
		for _, client_ := range wm.Clients {
			client := client_.(*xclient.Client)
			if int(client.Id()) == c {
				f(client)
				return int(client.Id())
			}
		}
		return ":void:"
	case string:
		switch c {
		case ":void:":
			return ":void:"
		case ":mouse:":
			wid := xproto.Window(wm.MouseClientClicked)
			if client := wm.FindManagedClient(wid); client != nil {
				c := client.(*xclient.Client)
				f(c)
				return int(c.Id())
			} else {
				f(nil)
				return ":void:"
			}
		default:
			for _, client_ := range wm.Clients {
				client := client_.(*xclient.Client)
				name := strings.ToLower(client.Name())
				if strings.Contains(name, strings.ToLower(c)) {
					f(client)
					return int(client.Id())
				}
			}
			return ":void:"
		}
	default:
		panic(fmt.Sprintf("BUG: Unknown Gribble return type: %T", c))
	}
	panic("unreachable")
}

func withWorkspace(wArg gribble.Any, f func(wrk *workspace.Workspace)) {
	switch w := wArg.(type) {
	case int:
		if wrk := wm.Heads.Workspaces.Get(w); wrk != nil {
			f(wrk)
		}
	case string:
		if wrk := wm.Heads.Workspaces.Find(w); wrk != nil {
			f(wrk)
		}
	}
}
