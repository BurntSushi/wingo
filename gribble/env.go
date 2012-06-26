package gribble

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/BurntSushi/wingo/logger"
)

type Command interface {
	Run() interface{}
}

type Environment struct {
	commands map[string]Command
}

func New(commands []Command) *Environment {
	cmds := make(map[string]Command, len(commands))
	for _, cmd := range commands {
		val := reflect.ValueOf(cmd)
		if val.Kind() != reflect.Struct {
			logger.Warning.Printf(
				"Type '%s' has kind '%s', but Gribble requires structs.",
				val.Type(), val.Kind())
			continue
		}
		cmds[val.Type().Name()] = cmd
	}

	return &Environment{commands: cmds}
}

func (env *Environment) Usage(cmdName string) string {
	cmd, ok := env.commands[cmdName]
	if !ok {
		return fmt.Sprintf("Usage: No such command '%s'.", cmdName)
	}
	val := reflect.ValueOf(cmd)
	if val.Kind() != reflect.Struct {
		panic("unreachable")
	}
	return fmt.Sprintf("%s: USAGE", val.Type().Name())
}

func (env *Environment) String() string {
	cmds := make([]string, 0, len(env.commands))
	for cmdName := range env.commands {
		cmds = append(cmds, env.Usage(cmdName))
	}
	sort.Sort(sort.StringSlice(cmds))
	return strings.Join(cmds, "\n")
}
