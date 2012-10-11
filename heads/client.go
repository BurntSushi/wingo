package heads

import (
	"github.com/BurntSushi/xgb/xproto"
)

type Clients interface {
	Get(i int) Client
	Len() int
}

type Client interface {
	Id() xproto.Window
	IsMaximized() bool
	Remaximize()
}
