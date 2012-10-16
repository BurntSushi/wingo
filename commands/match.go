package commands

import (
	"strings"

	"github.com/BurntSushi/gribble"

	"github.com/BurntSushi/wingo/logger"
	// "github.com/BurntSushi/wingo/wm" 
	"github.com/BurntSushi/wingo/xclient"
)

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
Always returns 1.
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
		return intToBool(cmd.Op)
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
