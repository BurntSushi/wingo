package misc

import (
	"strings"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
)

// keyMatch is a utility function for comparing two keysym strings for equality.
// It automatically converts a (mods, byte) pair to a string.
func KeyMatch(X *xgbutil.XUtil, target string,
	mods uint16, keycode xproto.Keycode) bool {

	guess := keybind.LookupString(X, mods, keycode)
	return strings.ToLower(guess) == strings.ToLower(target)
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// This exists because '%' isn't really modulus; it's *remainder*.
// e.g., (-1) % 2 = -1 but (-1) mod 2 = 1.
func Mod(x, m int) int {
	r := x % m
	if r < 0 {
		r += m
	}
	return r
}
