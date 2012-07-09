package gribble

import (
	"fmt"
	"reflect"
)

// e is an alias for fmt.Errorf.
func e(format string, v ...interface{}) error {
	return fmt.Errorf(format, v...)
}

// concrete recursively finds the concrete value pointed to by 'v'.
func concrete(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		return concrete(reflect.Indirect(v))
	}
	return v
}

// panicf is an alias for panic(fmt.Sprintf(...)).
func panicf(format string, v ...interface{}) {
	panic(fmt.Sprintf(format, v...))
}

// mustBe panics if 'val's kind is not 'kind'.
func mustBe(val reflect.Value, kind reflect.Kind) {
	if val.Kind() != kind {
		panicf("Type '%s' has kind '%s', but Gribble requires structs.",
			val.Type(), val.Kind())
	}
}

// mustBeStruct panics if 'val' does not have kind 'reflect.Struct'.
func mustBeStruct(val reflect.Value) {
	mustBe(val, reflect.Struct)
}
