package gribble

import (
	"fmt"
	"reflect"
)

func panicf(format string, v ...interface{}) {
	panic(fmt.Sprintf(format, v...))
}

func mustBe(val reflect.Value, kind reflect.Kind) {
	if val.Kind() != kind {
		panicf("Type '%s' has kind '%s', but Gribble requires structs.",
			val.Type(), val.Kind())
	}
}

func mustBeStruct(val reflect.Value) {
	mustBe(val, reflect.Struct)
}
