package commands

import (
	"strings"

	"github.com/BurntSushi/gribble"

	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/xclient"
)

type MatchClientMapped struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help string `
Returns 1 if the window specified by Client is mapped or not.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd MatchClientMapped) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		matched := false
		withClient(cmd.Client, func(c *xclient.Client) {
			matched = c.IsMapped()
		})
		return boolToInt(matched)
	})
}

type MatchClientClass struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Class string `param:"2"`
	Help string `
Returns 1 if the "class" part of the WM_CLASS property on the window
specified by Client contains the substring specified by Class, and otherwise
returns 0. The search is done case insensitively.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd MatchClientClass) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		matched := false
		withClient(cmd.Client, func(c *xclient.Client) {
			needle := strings.ToLower(cmd.Class)
			haystack := strings.ToLower(c.Class().Class)
			if strings.Contains(haystack, needle) {
				matched = true
			}
		})
		return boolToInt(matched)
	})
}

type MatchClientInstance struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Instance string `param:"2"`
	Help string `
Returns 1 if the "instance" part of the WM_CLASS property on the window
specified by Client contains the substring specified by Instance, and otherwise
returns 0. The search is done case insensitively.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd MatchClientInstance) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		matched := false
		withClient(cmd.Client, func(c *xclient.Client) {
			needle := strings.ToLower(cmd.Instance)
			haystack := strings.ToLower(c.Class().Instance)
			if strings.Contains(haystack, needle) {
				matched = true
			}
		})
		return boolToInt(matched)
	})
}

type MatchClientIsTransient struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Help string `
Returns 1 if the window specified by Client is a transient window, and
otherwise returns 0. A transient window usually corresponds to some kind of
dialog window.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd MatchClientIsTransient) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		matched := false
		withClient(cmd.Client, func(c *xclient.Client) {
			matched = c.IsTransient()
		})
		return boolToInt(matched)
	})
}

type MatchClientName struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Name string `param:"2"`
	Help string `
Returns 1 if the name of the window specified by Client contains the substring
specified by Name, and otherwise returns 0. The search is done case
insensitively.

Client may be the window id or a substring that matches a window name.
`
}

func (cmd MatchClientName) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		matched := false
		withClient(cmd.Client, func(c *xclient.Client) {
			needle := strings.ToLower(cmd.Name)
			haystack := strings.ToLower(c.Name())
			if strings.Contains(haystack, needle) {
				matched = true
			}
		})
		return boolToInt(matched)
	})
}

type MatchClientType struct {
	Client gribble.Any `param:"1" types:"int,string"`
	Type string `param:"2"`
	Help string `
Returns 1 if the type of the window specified by Client matches the type
named by Type, and otherwise returns 0.

Valid window types are "Normal", "Dock" or "Desktop".

Client may be the window id or a substring that matches a window name.
`
}

func (cmd MatchClientType) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		matched := false
		withClient(cmd.Client, func(c *xclient.Client) {
			if strings.ToLower(cmd.Type) == c.PrimaryTypeString() {
				matched = true
			}
		})
		return boolToInt(matched)
	})
}

type True struct {
	Help string `
Always returns 1.
`
}

func (cmd True) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		return boolToInt(true)
	})
}

type False struct {
	Help string `
Always returns 0.
`
}

func (cmd False) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		return boolToInt(false)
	})
}

type Not struct {
	Op int `param:"1"`
	Help string `
Returns the negation of Op. When Op is 0, Not returns 1. When Op is 1, Not
returns 0.

If Op is not in {0, 1}, then a warning is logged and nil is returned.
`
}

func (cmd Not) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		if cmd.Op != 0 && cmd.Op != 1 {
			logger.Warning.Printf(
				"'Not' received a value not in {0, 1}: %d", cmd.Op)
			return nil
		}
		return boolToInt(!intToBool(cmd.Op))
	})
}

type And struct {
	Op1 int `param:"1"`
	Op2 int `param:"2"`
	Help string `
Returns the logical AND of Op1 and Op2.

If Op1 or Op2 is not in {0, 1}, then a warning is logged and nil is returned.
`
}

func (cmd And) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		if cmd.Op1 != 0 && cmd.Op1 != 1 {
			logger.Warning.Printf(
				"Op1 in 'Or' received a value not in {0, 1}: %d", cmd.Op1)
			return nil
		}
		if cmd.Op2 != 0 && cmd.Op2 != 1 {
			logger.Warning.Printf(
				"Op2 in 'Or' received a value not in {0, 1}: %d", cmd.Op2)
			return nil
		}
		return boolToInt(intToBool(cmd.Op1) && intToBool(cmd.Op2))
	})
}

type Or struct {
	Op1 int `param:"1"`
	Op2 int `param:"2"`
	Help string `
Returns the logical OR of Op1 and Op2.

If Op1 or Op2 is not in {0, 1}, then a warning is logged and nil is returned.
`
}

func (cmd Or) Run() gribble.Value {
	return syncRun(func() gribble.Value {
		if cmd.Op1 != 0 && cmd.Op1 != 1 {
			logger.Warning.Printf(
				"Op1 in 'Or' received a value not in {0, 1}: %d", cmd.Op1)
			return nil
		}
		if cmd.Op2 != 0 && cmd.Op2 != 1 {
			logger.Warning.Printf(
				"Op2 in 'Or' received a value not in {0, 1}: %d", cmd.Op2)
			return nil
		}
		return boolToInt(intToBool(cmd.Op1) || intToBool(cmd.Op2))
	})
}
