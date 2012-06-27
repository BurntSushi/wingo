package gribble

import (
	"reflect"
)

type gparam struct {
	command Command
	reflect.StructField
}

func newParam(cmd Command, field reflect.StructField) gparam {
	p := gparam{
		command:     cmd,
		StructField: field,
	}
	p.mustHaveValidType()
	return p
}

func (p gparam) types() string {
	return p.Type.Name()
}

// XXX: Support multiple types.
// i.e., 'clientId interface{} `param:1,types:int,command`'.
func (p gparam) mustHaveValidType() {
	switch p.Type.Kind() {
	case reflect.Int:
	case reflect.Float64:
	case reflect.String:
	case reflect.Interface:
		if p.Type.Name() != "Command" {
			panicf("In command '%s', parameter '%s' must have type "+
				"'gribble.Command', but instead it has '%s'.",
				name(p.command), p.Name, p.Type.String())
		}
	default:
		panicf("In command '%s', parameter '%s' has type '%s', but "+
			"Gribble only allows int, float64, string and "+
			"gribble.Command types.",
			name(p.command), p.Name, p.Type.String())
	}
}
