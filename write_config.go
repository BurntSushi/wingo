package main

import (
	"io"
	"os"
	"path"
	"strings"

	"github.com/BurntSushi/wingo-conc/logger"
	"github.com/BurntSushi/wingo-conc/misc"
)

func writeConfigFiles() {
	var configDir string

	xdgHome := os.Getenv("XDG_CONFIG_HOME")
	home := os.Getenv("HOME")
	if len(xdgHome) > 0 && strings.HasPrefix(xdgHome, "/") {
		configDir = path.Join(xdgHome, "wingo")
	} else if len(home) > 0 && strings.HasPrefix(home, "/") {
		configDir = path.Join(home, ".config", "wingo")
	} else {
		logger.Error.Fatalf("Something is screwy. Wingo could not detect "+
			"valid values in your XDG_CONFIG_HOME ('%s') or HOME ('%s') "+
			"environment variables.", xdgHome, home)
	}

	// If the directory already exists, we quit---avoid accidentally
	// overwriting an existing configuration!
	if _, err := os.Stat(configDir); err == nil || os.IsExist(err) {
		logger.Error.Fatalf("Writing config files failed. The directory '%s' "+
			"already exists. Please remove it if you want a fresh set of "+
			"configuration files.", configDir)
	}

	// Okay, we're all set to continue. Create the directory and copy all of
	// the configuration files.
	if err := os.MkdirAll(configDir, 0777); err != nil {
		logger.Error.Fatalf("Could not create directory '%s': %s",
			configDir, err)
	}

	files := []string{
		"hooks.wini", "key.wini", "mouse.wini", "options.wini", "theme.wini",
	}
	for _, f := range files {
		dst := path.Join(configDir, f)
		if err := copyFile(dst, misc.ConfigFile(f)); err != nil {
			logger.Error.Fatalf("Could not copy file '%s' to '%s': %s",
				f, dst, err)
		}
	}
}

func copyFile(dst, src string) error {
	fsrc, err := os.Open(src)
	if err != nil {
		return err
	}
	fdst, err := os.Create(dst)
	if err != nil {
		return err
	}

	if _, err := io.Copy(fdst, fsrc); err != nil {
		return err
	}
	return nil
}
