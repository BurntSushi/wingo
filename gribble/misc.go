package gribble

import (
	"fmt"
	"reflect"
)

func e(format string, v ...interface{}) error {
	return fmt.Errorf(format, v...)
}

func concrete(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		return concrete(reflect.Indirect(v))
	}
	return v
}

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
