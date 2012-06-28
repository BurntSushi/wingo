package gribble

import (
	"go/token"
	"reflect"
)

type commandStruct struct {
	reflect.Value
	name   string
	params []*gparam
}

func newCommandStruct(cmd Command) *commandStruct {
	concreteCmd := concrete(reflect.ValueOf(cmd))
	mustBeStruct(concreteCmd)

	cmdStruct := &commandStruct{
		Value: concreteCmd,
		name:  cmdName(concreteCmd),
	}
	cmdStruct.params = paramFields(cmdStruct)
	return cmdStruct
}

func cmdName(val reflect.Value) string {
	mustBeStruct(val)
	typ := val.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tag := string(field.Tag)
		if field.Name == "name" &&
			field.Type.Kind() == reflect.String &&
			token.Lookup(tag) == token.IDENT {

			return tag
		}
	}
	return typ.Name()
}

func paramFields(cmdStruct *commandStruct) []*gparam {
	typ := cmdStruct.Type()
	maxParams := typ.NumField()
	pFields := make(map[int]*gparam, maxParams)
	lastParam := -1
	for i := 0; i < maxParams; i++ {
		field := typ.Field(i)
		if pStr := field.Tag.Get("param"); len(pStr) > 0 {
			// newParam panics if this parameter field violates any of the
			// type invariants defined by Gribble.
			p := newParam(cmdStruct, field)

			ind := p.number - 1
			if _, ok := pFields[ind]; ok {
				panicf("Command '%s' has more than one parameter '%d'.",
					cmdStruct.name, p.number)
			}
			if ind >= maxParams {
				panicf("Command '%s' can have a maximum of %d parameters, "+
					"but found a parameter numbered '%d'.",
					cmdStruct.name, maxParams, p.number)
			}

			pFields[ind] = p
			if ind > lastParam {
				lastParam = ind
			}
		}
	}

	// Check to make sure we have contiguous parameters, and load parameters
	// in order into a slice.
	slice := make([]*gparam, lastParam+1)
	for i := 0; i <= lastParam; i++ {
		if _, ok := pFields[i]; !ok {
			panicf("Command '%s' is missing parameter '%d'.",
				cmdStruct.name, i+1)
		}
		slice[i] = pFields[i]
	}
	return slice
}
