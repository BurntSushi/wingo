package gribble

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// Any is an empty interface that can be used as a type for parameters that
// can accept values of more than one type. When using the 'Any' type as a
// parameter type, the struct tag 'types' MUST be set.
type Any interface{}

// gparam represents a single parameter in a user supplied command struct.
type gparam struct {
	// cmdStruct is the commandStruct containing this parameter.
	// It is stored here for outputting nicer error messages (i.e., include
	// the command name).
	cmdStruct *commandStruct

	// Every parameter corresponds to a single struct field inside a user
	// supplied command struct. Embedding it here allows a gparam to act like
	// one.
	reflect.StructField

	// validTypes is a simple list of strings of allowable types for this 
	// parameter. This list is derived either from the concrete type of the
	// parameter from the user supplied command struct, or from the list of
	// types specified in this struct field's struct tag.
	//
	// validTypes can only have more than one type if the struct field has
	// type 'Any'.
	//
	// Valid types are 'int', 'float' or 'string'.
	validTypes []string

	// The parameter number offset (starting at 1).
	number int
}

// newParam uses a struct field to infer all necessary information to create
// a new parameter.
func newParam(cmdStruct *commandStruct, field reflect.StructField) *gparam {
	p := &gparam{
		cmdStruct:   cmdStruct,
		StructField: field,
	}
	p.loadParamNumber()
	p.loadValidTypes()
	return p
}

// isValidType returns true if 't' is in 'validTypes' and false otherwise.
func (p *gparam) isValidType(t string) bool {
	for _, t2 := range p.validTypes {
		if t == t2 {
			return true
		}
	}
	return false
}

// types returns a string of all allowable types for this parameter with each
// type separated by a "|".
func (p *gparam) types() string {
	return strings.Join(p.validTypes, " | ")
}

// typesOr returns a string of all allowable types for this parameter with
// each type quoted and separated by an "or".
func (p *gparam) typesOr() string {
	quoted := make([]string, len(p.validTypes))
	for i, t := range p.validTypes {
		quoted[i] = fmt.Sprintf("'%s'", t)
	}
	return strings.Join(quoted, " or ")
}

// loadParamNumber tries to find this parameter's number by inspecting the
// struct field's struct tag. loadParamNumber panics if the 'param' field is
// missing from the struct tag, or if its value is not an integer.
func (p *gparam) loadParamNumber() {
	pStr := p.Tag.Get("param")
	if len(pStr) == 0 {
		panicf("In command '%s', '%s' is not a valid parameter because it is "+
			"missing 'param' in a struct tag.", p.cmdStruct.name, p.Name)
	}
	if pNum, err := strconv.ParseInt(pStr, 10, 32); err == nil {
		p.number = int(pNum)
	} else {
		panicf("In command '%s', '%s' is not a valid parameter number.",
			p.cmdStruct.name, pStr)
	}
}

// loadValidTypes populates the 'validTypes' member of the 'gparam' struct.
// Population is straight-forward for fields with concrete types corresponding
// to int, float64 or string. For fields with 'Any' type, the valid
// types are inferred from the 'types' struct tag. (If said types cannot be
// inferred, then loadValidTypes panics.) If a field has any type other than
// int, float64, string or Any, loadValidTypes panics.
func (p *gparam) loadValidTypes() {
	switch p.Type.Kind() {
	case reflect.Int:
		p.validTypes = []string{"int"}
	case reflect.Float64:
		p.validTypes = []string{"float"}
	case reflect.String:
		p.validTypes = []string{"string"}
	case reflect.Interface:
		// Found this little trick in Go's rpc/server.go.
		// The idea here is that TypeOf takes an empty interface value and
		// always returns the concrete type inside the interface. Get around
		// it by passing a pointer a benign interface value, and fetching
		// the interface type using Elem.
		if reflect.TypeOf((*Any)(nil)).Elem() == p.Type {
			p.loadEmptyInterfaceTypes()
		} else {
			panicf("In command '%s', parameter '%s' must have type "+
				"'Any', but instead it has '%s'.",
				p.cmdStruct.name, p.Name, p.Type.String())
		}
	default:
		panicf("In command '%s', parameter '%s' has type '%s', but "+
			"Gribble only allows int, float64, string and Any types.",
			p.cmdStruct.name, p.Name, p.Type.String())
	}
}

// loadEmptyInterfaceTypes populates the 'validTypes' member of the 'gparam'
// struct by inspecting the 'types' struct tag. The only types allowed are
// int, float and string.
//
// loadEmptyInterfaceTypes panics if: Any type other than int, float or string
// is found in the 'types' struct tag. If the 'types' struct tag doesn't exist
// or is empty. Or if the 'types' struct tag contains only a single type (in
// which case, a concrete type should be used instead).
func (p *gparam) loadEmptyInterfaceTypes() {
	validTypes := strings.Split(p.Tag.Get("types"), ",")
	p.validTypes = make([]string, 0, len(validTypes))
	for _, t := range validTypes {
		switch t {
		case "":
			continue
		case "int":
		case "float":
		case "string":
		default:
			panicf("In command '%s', parameter '%s' has listed an "+
				"invalid type '%s'. Valid types are int, float and string.",
				p.cmdStruct.name, p.Name, t)
		}
		p.validTypes = append(p.validTypes, t)
	}
	switch len(p.validTypes) {
	case 0:
		panicf("In command '%s', parameter '%s' is listed as an "+
			"Any type, but has not specified a list of "+
			"allowable types in a struct tag. i.e., "+
			"'%s Any `param:\"%d\" types:\"int,float\"`'.",
			p.cmdStruct.name, p.Name, p.Name, p.number)
	case 1:
		loneType := p.validTypes[0]
		if loneType == "float" {
			loneType = "float64"
		}
		panicf("In command '%s', parameter '%s' is listed as an "+
			"Any type, which allows for a parameter to "+
			"represent one of several types. However, only one type "+
			"is listed. Therefore, please use a concrete type instead "+
			"of Any. i.e., '%s %s `param:\"%d\"`.",
			p.cmdStruct.name, p.Name, p.Name, loneType, p.number)
	}
	sort.Sort(sort.StringSlice(p.validTypes))
}
