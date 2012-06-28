package gribble

import (
	"fmt"
	"sort"
	"strings"
)

type Environment struct {
	commands map[string]*commandStruct

	// Verbose controls the detail of error messages from the Gribble parser.
	// When true (the default value), a single error message may span
	// multiple lines, and will attempt to pinpoint exactly where an error
	// is occurring.
	// When false, a single error message will always be on one line.
	// This value may be changed at any time.
	Verbose bool
}

func New(commands []Command) *Environment {
	cmds := make(map[string]*commandStruct, len(commands))
	for _, cmd := range commands {
		cmdStruct := newCommandStruct(cmd)
		cmds[cmdStruct.name] = cmdStruct
	}

	return &Environment{
		commands: cmds,
		Verbose:  true,
	}
}

func (env *Environment) Run(cmd string) (interface{}, error) {
	command, err := env.Command(cmd)
	if err != nil {
		return nil, err
	}
	return command.Run(), nil
}

func (env *Environment) Command(cmd string) (Command, error) {
	parsedCmd, err := parse(cmd, env.Verbose)
	if err != nil {
		return nil, err
	}

	filledCommand, err := newCommand(env, parsedCmd)
	if err != nil {
		return nil, err
	}
	return filledCommand, nil
}

func (env *Environment) findCommand(parsedCmd *command) *commandStruct {
	if cmd, ok := env.commands[parsedCmd.name]; ok {
		return cmd
	}
	return nil
}

func (env *Environment) Usage(cmdName string) string {
	return env.usage(cmdName,
		func(param *gparam) string {
			return param.Name
		})
}

func (env *Environment) UsageTypes(cmdName string) string {
	return env.usage(cmdName,
		func(param *gparam) string {
			return fmt.Sprintf("(%s :: %s)", param.Name, param.types())
		})
}

func (env *Environment) usage(cmdName string,
	fieldTrans func(*gparam) string) string {

	cmdStruct, ok := env.commands[cmdName]
	if !ok {
		return fmt.Sprintf("No such command '%s'.", cmdName)
	}

	params := make([]string, len(cmdStruct.params))
	for i, field := range cmdStruct.params {
		params[i] = fieldTrans(field)
	}
	return fmt.Sprintf("%s: %s", cmdStruct.name, strings.Join(params, " "))
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
