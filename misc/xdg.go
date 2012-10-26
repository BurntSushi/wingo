package misc

import (
	"io/ioutil"

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
