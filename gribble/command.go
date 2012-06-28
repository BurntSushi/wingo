package gribble

import (
	"fmt"
	"reflect"
)

type Command interface {
	Run() Value
}

func newCommand(env *Environment, parsedCmd *command) (Command, error) {
	cmdStruct := env.findCommand(parsedCmd)
	if cmdStruct == nil {
		return nil, e("Command '%s' does not exist.", parsedCmd.name)
	}

	if len(cmdStruct.params) != len(parsedCmd.params) {
		return nil, e("Command '%s' has %d parameters, but %d were given.",
			cmdStruct.name, len(cmdStruct.params), len(parsedCmd.params))
	}

	filled := reflect.New(cmdStruct.Type())
	for i, pField := range cmdStruct.params {
		pNum := i + 1
		toFill := filled.Elem().FieldByName(pField.Name)

		var val Value
		if parsedSubCmd, ok := parsedCmd.params[i].(*command); ok {
			subCmd, err := newCommand(env, parsedSubCmd)
			if err != nil {
				return nil, err
			}
			val = subCmd.Run()
		} else {
			val = parsedCmd.params[i]
		}

		typeError := func() error {
			hasType := fmt.Sprintf("%T", val)
			if hasType == "float64" {
				hasType = "float"
			}
			return e("When executing command '%s', parameter %d has "+
				"type '%s', but '%s' expected %s.",
				cmdStruct.name, pNum, hasType, cmdStruct.name,
				pField.typesOr())
		}
		if err := fillNonCmdParam(typeError, toFill, pField, val); err != nil {
			return nil, err
		}
	}

	// We can blithely type assert because 'filled' was created with a type
	// returned by reflection from 'command', which is guaranteed to implement
	// the Command interface by Go's type system.
	return filled.Interface().(Command), nil
}

func fillNonCmdParam(typeError func() error,
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
		// the way we "fill" parameter values.
		panic(typeError())
	}

	return nil
}
