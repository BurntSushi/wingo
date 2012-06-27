package gribble

import (
	"go/token"
	"reflect"
	"strconv"
)

type Command interface {
	Run() Value
}

func name(cmd Command) string {
	mustBeStruct(reflect.ValueOf(cmd))
	typ := reflect.TypeOf(cmd)
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

func paramFields(cmd Command) []gparam {
	typ := reflect.TypeOf(cmd)
	maxParams := typ.NumField()
	pFields := make(map[int]gparam, maxParams)
	lastParam := -1
	for i := 0; i < maxParams; i++ {
		field := typ.Field(i)
		if pStr := field.Tag.Get("param"); len(pStr) > 0 {
			// newParam panics if this parameter field violates any of the
			// type invariants defined by Gribble.
			p := newParam(cmd, field)

			if pNum, err := strconv.ParseInt(pStr, 10, 32); err == nil {
				ind := int(pNum - 1)
				if _, ok := pFields[ind]; ok {
					panicf("Command '%s' has more than one parameter '%d'.",
						name(cmd), pNum)
				}
				if ind >= maxParams {
					panicf("Command '%s' can have a maximum of %d parameters, "+
						"but found a parameter numbered '%d'.",
						name(cmd), maxParams, pNum)
				}

				pFields[ind] = p
				if ind > lastParam {
					lastParam = ind
				}
			} else {
				panicf("In command '%s', '%s' is not a valid parameter number.",
					name(cmd), pStr)
			}
		}
	}

	// Check to make sure we have contiguous parameters, and load parameters
	// in order into a slice.
	slice := make([]gparam, lastParam+1)
	for i := 0; i <= lastParam; i++ {
		if _, ok := pFields[i]; !ok {
			panicf("Command '%s' is missing parameter '%d'.", name(cmd), i+1)
		}
		slice[i] = pFields[i]
	}
	return slice
}
