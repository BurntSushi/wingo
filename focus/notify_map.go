package focus

import "github.com/BurntSushi/xgb/xproto"

var (
	// A map from mode constants to human readable strings.
	// For the deets:
	// http://tronche.com/gui/x/xlib/events/input-focus/normal-and-grabbed.html
	Modes = map[byte]string{
		xproto.NotifyModeNormal:       "NotifyNormal",
		xproto.NotifyModeGrab:         "NotifyGrab",
		xproto.NotifyModeUngrab:       "NotifyUngrab",
		xproto.NotifyModeWhileGrabbed: "NotifyWhileGrabbed",
	}

	// A map from detail constants to human readable strings.
	// For the deets:
	// http://tronche.com/gui/x/xlib/events/input-focus/normal-and-grabbed.html
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
