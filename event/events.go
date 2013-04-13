package event

import (
	"github.com/BurntSushi/xgb/xproto"
)

type Event interface{}

type Noop struct{}

type (
	ChangedWorkspace        struct{}
	ChangedVisibleWorkspace struct{}
	ChangedWorkspaceNames   struct{}

	AddedWorkspace struct {
		Name string
	}

	RemovedWorkspace struct {
		Name string
	}
)

type (
	FocusedClient struct {
		Id xproto.Window
	}
	UnfocusedClient struct {
		Id xproto.Window
	}
	MappedClient struct {
		Id xproto.Window
	}
	UnmappedClient struct {
		Id xproto.Window
	}
	ManagedClient struct {
		Id xproto.Window
	}
	UnmanagedClient struct {
		Id xproto.Window
	}
	ChangedClientName struct {
		Id xproto.Window
	}
)

type ChangedLayout struct {
	Workspace string
}
