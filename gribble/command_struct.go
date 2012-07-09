package gribble

import (
	"go/token"
	"reflect"
)

// commandStruct embeds a reflect.Value of a Command interface's concrete
// value, as well as the name of the command and its parameter list.
//
// A list of commandStruct constitutes a Gribble environment.
type commandStruct struct {
	reflect.Value
	name   string
	params []*gparam
}

// newCommandStruct constructs new commandStruct values from user supplied
// values that both implement the Command interface *and* are structs.
//
// The struct itself is scanned for a valid listing of parameters.
// newCommandStruct will panic if a supplied value does not have a valid
// parameter list.
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

// cmdName uses a reflect.Value of a Command to find a command's name.
// A command name can either be specified in a struct tag of a 'name' field,
// or if that is absent, the name of the struct type itself is used.
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

// paramFields returns a list of all parameters for the given command struct
// in the form of a slice of gparam values.
//
// paramFields panics if a new parameter cannot be constructed, or if there
// are more parameters than expected or if there are multiple parameters 
// labeled with the same parameter number or if there are non-contiguous
// parameter numbers.
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

			// We keep track of the last parameter number so that we know when
			// the contiguous block of parameters *should* stop.
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
