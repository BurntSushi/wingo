package misc

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/BurntSushi/xdg"

	"github.com/BurntSushi/wingo/logger"
)

var ConfigPaths = xdg.Paths{
	Override:     "",
	XDGSuffix:    "wingo",
	GoImportPath: "github.com/BurntSushi/wingo/config",
}

var DataPaths = xdg.Paths{
	Override:     "",
	XDGSuffix:    "wingo",
	GoImportPath: "github.com/BurntSushi/wingo/data",
}

var ScriptPaths = xdg.Paths{
	Override:     "",
	XDGSuffix:    "wingo",
	GoImportPath: "github.com/BurntSushi/wingo/config",
}

func ConfigFile(name string) string {
	fpath, err := ConfigPaths.ConfigFile(name)
	if err != nil {
		logger.Error.Fatalln(err)
	}
	return fpath
}

func DataFile(name string) []byte {
	fpath, err := DataPaths.DataFile(name)
	if err != nil {
		logger.Error.Fatalln(err)
	}
	bs, err := ioutil.ReadFile(fpath)
	if err != nil {
		logger.Error.Fatalf("Could not read %s: %s", fpath, err)
	}
	return bs
}

func ScriptPath(name string) string {
	fpath, err := ScriptPaths.ConfigFile(path.Join("scripts", name, name))
	if err != nil {
		fpath, err = ScriptPaths.ConfigFile(path.Join("scripts", name))
		if err != nil {
			logger.Warning.Println(err)
			return ""
		}
	}
	return fpath
}

func ScriptConfigPath(name string) string {
	fname := fmt.Sprintf("%s.cfg", name)
	fp, err := ScriptPaths.ConfigFile(path.Join("scripts", name, fname))
	if err != nil {
		logger.Warning.Println(err)
		return ""
	}
	return fp
}
