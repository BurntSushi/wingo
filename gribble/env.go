package gribble

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type Environment struct {
	commands map[string]Command
}

func New(commands []Command) *Environment {
	cmds := make(map[string]Command, len(commands))
	for _, cmd := range commands {
		mustBeStruct(reflect.ValueOf(cmd))
		paramFields(cmd) // for possible panic side effect
		cmds[name(cmd)] = cmd
	}

	return &Environment{commands: cmds}
}

func (env *Environment) Usage(cmdName string) string {
	return env.usage(cmdName,
		func(param gparam) string {
			return param.Name
		})
}

func (env *Environment) UsageTypes(cmdName string) string {
	return env.usage(cmdName,
		func(param gparam) string {
			return fmt.Sprintf("(%s :: %s)", param.Name, param.types())
		})
}

func (env *Environment) usage(cmdName string,
	fieldTrans func(gparam) string) string {

	cmd, ok := env.commands[cmdName]
	if !ok {
		return fmt.Sprintf("Usage: No such command '%s'.", cmdName)
	}

	pFields := paramFields(cmd)
	params := make([]string, len(pFields))
	for i, field := range paramFields(cmd) {
		params[i] = fieldTrans(field)
	}
	return fmt.Sprintf("%s: %s", name(cmd), strings.Join(params, " "))
}

func (env *Environment) String() string {
	cmds := make([]string, 0, len(env.commands))
	for cmdName := range env.commands {
		cmds = append(cmds, env.Usage(cmdName))
	}
	sort.Sort(sort.StringSlice(cmds))
	return strings.Join(cmds, "\n")
}

func (env *Environment) StringTypes() string {
	cmds := make([]string, 0, len(env.commands))
	for cmdName := range env.commands {
		cmds = append(cmds, env.UsageTypes(cmdName))
	}
	sort.Sort(sort.StringSlice(cmds))
	return strings.Join(cmds, "\n")
}
