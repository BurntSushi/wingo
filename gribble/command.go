package gribble

import (
	"fmt"
	"reflect"
)

// Command is an interface that all commands in the Gribble environment must
// satisfy. Namely, every command must define a 'Run' method that returns some
// 'Value'. The 'Run' method ought to perform the computation required by the
// command.
//
// For example, consider an 'add' command with two parameters:
//
//	type Add struct {
//		Op1 int `param:"1"`
//		Op2 int `param:"2"`
//	}
//	func (cmd *Add) Run() gribble.Value {
//		return c.Op1 + c.Op2
//	}
//
// Where 'Value' is an empty interface whose concrete type is guaranteed to
// be an int, float64 or string by the Gribble runtime.
type Command interface {
	Run() Value
}

// newCommand takes an environment and a valid parsed command and returns a
// value implementing the Command interface.
//
// Note that all sub-commands are executed, and every parameter value in the
// Command value returned is filled in.
func newCommand(env *Environment, parsedCmd *command) (Command, error) {
	cmdStruct := env.findCommand(parsedCmd)
	if cmdStruct == nil {
		return nil, e("Command '%s' does not exist.", parsedCmd.name)
	}
	if len(cmdStruct.params) != len(parsedCmd.params) {
		return nil, e("Command '%s' has %d parameters, but %d were given.",
			cmdStruct.name, len(cmdStruct.params), len(parsedCmd.params))
	}

	// filled is a new reflect.Value instance of the concrete command
	// represented by the embedded 'reflect.Value' type in the 'commandStruct'
	// type. Since the 'newCommandStruct' only allows values of type 'Command'
	// interface, we are guaranteed that 'filled's underlying concrete type
	// can be type asserted to a Command interface type.
	filled := reflect.New(cmdStruct.Type())
	for i, pField := range cmdStruct.params {
		// i is the parameter index starting from 0, and pNum is the
		// parameter index starting from 1 that is displayed in error messages.
		pNum := i + 1

		// toFill is the reflect.Value of this parameter's field inside
		// the new instance of the user's command struct.
		toFill := filled.Elem().FieldByName(pField.Name)

		// val is the parameter value, which can come from a literal or
		// a sub-command. We make sure that its concrete type is an int, float64
		// or a string.
		var val Value

		// If the parameter value can be type asserted to a *command, then we
		// have a sub-command on our hands. Look it up in the environment, run
		// it, and keep its value.
		if parsedSubCmd, ok := parsedCmd.params[i].(*command); ok {
			subCmd, err := newCommand(env, parsedSubCmd)
			if err != nil {
				return nil, err
			}
			val = subCmd.Run()
		} else {
			val = parsedCmd.params[i]
		}

		// typeError closes over several parameters to report a nice error
		// message when the type of a parameter is not what we expect.
		typeError := func() error {
			hasType := fmt.Sprintf("%T", val)
			if hasType == "float64" {
				hasType = "float" // keep type names consistent
			}
			return e("When executing command '%s', parameter %d has "+
				"type '%s', but '%s' expected %s.",
				cmdStruct.name, pNum, hasType, cmdStruct.name,
				pField.typesOr())
		}
		if err := fillParam(typeError, toFill, pField, val); err != nil {
			return nil, err
		}
	}

	// We can blithely type assert because 'filled' was created with a type
	// returned by reflection from a Command interface value, which is 
	// of course guaranteed to implement the Command interface.
	return filled.Interface().(Command), nil
}

// fillParam checks to make sure that the argument given (val) has a
// type that matches the expected parameter (pField) type. If it doesn't,
// the 'typeError' function is called and returned. If the types match, then
// the parameter field in the user's struct (toFill) is filled with the value
// found in the given argument (val).
//
// A more succinct overview: parameter and argument types much match, and 
// all gribble.Value values returned by sub-commands
// must have a concrete type of int, float64 or string.
func fillParam(typeError func() error,
	toFill reflect.Value, pField *gparam, val Value) error {

	switch paramVal := val.(type) {
	case int:
		if !pField.isValidType("int") {
			return typeError()
		}
		// toFill.Set(int64(paramVal)) 
		toFill.Set(reflect.ValueOf(paramVal))
	case float64:
		if !pField.isValidType("float") {
			return typeError()
		}
		// toFill.Set(paramVal) 
		toFill.Set(reflect.ValueOf(paramVal))
	case string:
		if !pField.isValidType("string") {
			return typeError()
		}
		toFill.Set(reflect.ValueOf(paramVal))
	default:
		// This is a panic because if a parameter has a value *other*
		// than int, string, or float, there's a bug in the parser or in
		// the way we "fill" parameter values. Alternatively, this could be
		// reached if the user violates the invariant that a command can only
		// return an int, float64 or a string.
		panic(typeError())
	}

	return nil
}

// Value is a named empty interface that corresponds to the type of any value
// returned by a Gribble command. All Value values returned by sub-commands
// must have a concrete type of int, float64 or string. If this invariant is
// violated, Gribble will panic with a run-time type error.
type Value interface{}
