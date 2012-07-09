package gribble

import (
	"fmt"
	"sort"
	"strings"
)

// Environment represents the set of all commands available to the Gribble
// run-time. A new Environment should be created with gribble.New.
type Environment struct {
	// Verbose controls the detail of error messages from the Gribble parser.
	// When true (the default value), a single error message may span
	// multiple lines, and will attempt to pinpoint exactly where an error
	// is occurring.
	// When false, a single error message will always be on one line.
	// This value may be changed at any time.
	Verbose bool

	// commands is map from command name to commandStruct. A commandStruct
	// contains a reflect.Value of a user supplied value that implements
	// the Command interface.
	commands map[string]*commandStruct
}

// New creates a new Gribble environment with the given list of values that
// implement the Command interface. An environment is returned with which
// commands can be executed. An environment can also be queried for usage
// information for a particular command in the environment.
//
// There is currently no way to add or remove commands from an environment
// once it is created.
func New(commands []Command) *Environment {
	cmds := make(map[string]*commandStruct, len(commands))
	for _, cmd := range commands {
		cmdStruct := newCommandStruct(cmd)
		if _, ok := cmds[cmdStruct.name]; ok {
			panicf("Two commands with the same name, '%s', are not allowed "+
				"in the same Gribble environment.", cmdStruct.name)
		}
		cmds[cmdStruct.name] = cmdStruct
	}

	return &Environment{
		commands: cmds,
		Verbose:  true,
	}
}

// Run will execute a given command in the provided environment. An error can
// occur when either parsing or running the command.
//
// The return value of all Gribble commands must implement the Value interface,
// which is empty. The conrete types of values returned are int, float64 or
// string. This is enforced by the Gribble run-time.
func (env *Environment) Run(cmd string) (Value, error) {
	command, err := env.Command(cmd)
	if err != nil {
		return nil, err
	}
	return command.Run(), nil
}

// Command returns a value implementing the Command interface that is ready to
// be executed. In particular, the concrete struct underlying the Command
// interface has had its values filled with literals specified in the "cmd"
// string or by values returned from sub-commands.
//
// This method exists in the event that a particular command requires
// additional information at run-time that cannot be captured by Gribble, or
// if a command needs to be stored and executed later.
//
// As a contrived example, consider the case when SomeOp differs depending
// upon who is running it:
//	package main
//	
//	import (
//		"fmt"
//		"github.com/BurntSushi/gribble"
//	)
//	
//	type SomeOp struct {
//		Op1 int `param:"1"`
//		Op2 int `param:"2"`
//		who string
//	}
//	
//	func (cmd *SomeOp) Run() gribble.Value {
//		switch cmd.who {
//		case "Andrew":
//			return cmd.Op1 * cmd.Op2
//		default:
//			return cmd.Op1 + cmd.Op2
//		}
//		panic("unreachable")
//	}
//	
//	func main() {
//		env := gribble.New([]gribble.Command{&SomeOp{}})
//		cmd, _ := env.Command("SomeOp 2 4")
//		someOp := cmd.(*SomeOp)
//	
//		someOp.who = "Andrew"
//		fmt.Println(someOp.Run()) // outputs "8"
//		someOp.who = "Someone Else"
//		fmt.Println(someOp.Run()) // outputs "6"
//	}
// 
// Note that due to the fact that commands themselves don't know if values have
// come from sub-commands or literals, this sort of state injection is only
// possible with top-level commands. (More precisely, both (*Environment).Run
// and (*Environment).Command execute all sub-commands for you.)
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

// CommandName is a convenience method for returning the name of the command
// used in the command string 'cmd'. The important aspect of this method is that
// it is impervious to most kinds of errors, including parse errors (so long
// as the parse error doesn't occur in the command name itself, in which case,
// the returned string will be empty).
func (env *Environment) CommandName(cmd string) string {
	parsedCmd, _ := parse(cmd, false)
	return parsedCmd.name
}

// findCommand returns the commandStruct corresponding to the command parsed.
// The lookup is by exact case sensitive name matching only.
func (env *Environment) findCommand(parsedCmd *command) *commandStruct {
	if cmd, ok := env.commands[parsedCmd.name]; ok {
		return cmd
	}
	return nil
}

// Usage returns a usage string derived from the command struct, including
// the command and parameter names.
func (env *Environment) Usage(cmdName string) string {
	return env.usage(cmdName,
		func(param *gparam) string {
			return param.Name
		})
}

// Usage returns a usage string derived from the command struct, including
// the command and parameter names, along with the allowable types for
// each parameter.
func (env *Environment) UsageTypes(cmdName string) string {
	return env.usage(cmdName,
		func(param *gparam) string {
			return fmt.Sprintf("(%s :: %s)", param.Name, param.types())
		})
}

// usage generates a usage message with a field transformation function
// 'fieldTrans' applied to each parameter field.
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
	return fmt.Sprintf("%s %s", cmdStruct.name, strings.Join(params, " "))
}

// String outputs a listing of all commands in the environment.
func (env *Environment) String() string {
	cmds := make([]string, 0, len(env.commands))
	for cmdName := range env.commands {
		cmds = append(cmds, env.Usage(cmdName))
	}
	sort.Sort(sort.StringSlice(cmds))
	return strings.Join(cmds, "\n")
}

// StringTypes outputs a listing of all commands with parameter types
// in the environment.
func (env *Environment) StringTypes() string {
	cmds := make([]string, 0, len(env.commands))
	for cmdName := range env.commands {
		cmds = append(cmds, env.UsageTypes(cmdName))
	}
	sort.Sort(sort.StringSlice(cmds))
	return strings.Join(cmds, "\n")
}
