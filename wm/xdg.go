package wm

import (
	"go/build"
	"os"
	"path"
	"strings"

	"github.com/BurntSushi/wingo/logger"
)

// ConfigFile returns a file path containing the configuration file
// specified. If one cannot be found, Wingo will quit and log an error.
func ConfigFile(name string) string {
	readable := func(p string) bool {
		if _, err := os.Open(p); err != nil {
			return false
		}
		return true
	}

	home := os.Getenv("HOME")
	xdgHome := os.Getenv("XDG_CONFIG_HOME")
	xdgDirs := os.Getenv("XDG_CONFIG_DIRS")

	// We're going to accumulate a list of directories for places to inspect
	// for configuration files. Basically, this includes following the
	// xdg basedir spec for the XDG_CONFIG_HOME and XDG_CONFIG_DIRS environment
	// variables.
	try := make([]string, 0)

	// from the command line
	if len(ConfigDir) > 0 {
		try = append(try, ConfigDir)
	}

	// XDG_CONFIG_HOME
	if len(xdgHome) > 0 && strings.HasPrefix(xdgHome, "/") {
		try = append(try, path.Join(xdgHome, "wingo"))
	} else if len(home) > 0 {
		try = append(try, path.Join(home, ".config", "wingo"))
	}

	// XDG_CONFIG_DIRS
	if len(xdgDirs) > 0 {
		for _, p := range strings.Split(xdgDirs, ":") {
			// XDG basedir spec does not allow relative paths
			if !strings.HasPrefix(p, "/") {
				continue
			}
			try = append(try, path.Join(p, "wingo"))
		}
	} else {
		try = append(try, path.Join("/", "etc", "xdg", "wingo"))
	}

	// Add directories from GOPATH. Last resort.
	for _, dir := range build.Default.SrcDirs() {
		d := path.Join(dir, "github.com", "BurntSushi", "wingo", "config")
		try = append(try, d)
	}

	// Now use the first one and keep track of the ones we've tried.
	tried := make([]string, 0, len(try))
	for _, dir := range try {
		if len(dir) == 0 {
			continue
		}

		fpath := path.Join(dir, name)
		if readable(fpath) {
			return fpath
		} else {
			tried = append(tried, fpath)
		}
	}

	// Show the user where we've looked for config files...
	triedStr := strings.Join(tried, ", ")
	logger.Error.Fatalf("Could not find a readable '%s' config file. Wingo "+
		"has tried the following paths: %s", name, triedStr)
	panic("unreachable")
}
