package focus

import "github.com/BurntSushi/xgb/xproto"

var (
	Modes = map[byte]string{
		xproto.NotifyModeNormal:       "NotifyNormal",
		xproto.NotifyModeGrab:         "NotifyGrab",
		xproto.NotifyModeUngrab:       "NotifyUngrab",
		xproto.NotifyModeWhileGrabbed: "NotifyWhileGrabbed",
	}

	Details = map[byte]string{
		xproto.NotifyDetailAncestor:         "NotifyAncestor",
		xproto.NotifyDetailVirtual:          "NotifyVirtual",
		xproto.NotifyDetailInferior:         "NotifyInferior",
		xproto.NotifyDetailNonlinear:        "NotifyNonlinear",
		xproto.NotifyDetailNonlinearVirtual: "NotifyNonlinearVirtual",
		xproto.NotifyDetailPointer:          "NotifyPointer",
		xproto.NotifyDetailPointerRoot:      "NotifyPointerRoot",
		xproto.NotifyDetailNone:             "NotifyNone",
	}
)
