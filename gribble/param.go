package gribble

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type gparam struct {
	cmdStruct *commandStruct
	reflect.StructField
	validTypes []string
	number     int
}

func newParam(cmdStruct *commandStruct, field reflect.StructField) *gparam {
	p := &gparam{
		cmdStruct:   cmdStruct,
		StructField: field,
	}
	p.loadParamNumber()
	p.loadValidTypes()
	return p
}

func (p *gparam) isValidType(t string) bool {
	for _, t2 := range p.validTypes {
		if t == t2 {
			return true
		}
	}
	return false
}

func (p *gparam) types() string {
	return strings.Join(p.validTypes, " | ")
}

func (p *gparam) typesOr() string {
	quoted := make([]string, len(p.validTypes))
	for i, t := range p.validTypes {
		quoted[i] = fmt.Sprintf("'%s'", t)
	}
	return strings.Join(quoted, " or ")
}

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

func (p *gparam) loadValidTypes() {
	switch p.Type.Kind() {
	case reflect.Int:
		p.validTypes = []string{"int"}
	case reflect.Float64:
		p.validTypes = []string{"float"}
	case reflect.String:
		p.validTypes = []string{"string"}
	case reflect.Interface:
		switch p.Type.String() {
		case "interface {}":
			p.loadEmptyInterfaceTypes()
		default:
			panicf("In command '%s', parameter '%s' must have type "+
				"'interface{}', but instead it has '%s'.",
				p.cmdStruct.name, p.Name, p.Type.String())
		}
	default:
		panicf("In command '%s', parameter '%s' has type '%s', but "+
			"Gribble only allows int, float64, and string types.",
			p.cmdStruct.name, p.Name, p.Type.String())
	}
}

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
			"interface{} type, but has not specified a list of "+
			"allowable types in a struct tag. i.e., "+
			"'%s interface{} `param:\"%d\" types:\"int,float\"`'.",
			p.cmdStruct.name, p.Name, p.Name, p.number)
	case 1:
		loneType := p.validTypes[0]
		if loneType == "float" {
			loneType = "float64"
		}
		panicf("In command '%s', parameter '%s' is listed as an "+
			"interface{} type, which allows for a parameter to "+
			"represent one of several types. However, only one type "+
			"is listed. Therefore, please use a concrete type instead "+
			"of interface{}. i.e., '%s %s `param:\"%d\"`.",
			p.cmdStruct.name, p.Name, p.Name, loneType, p.number)
	}
	sort.Sort(sort.StringSlice(p.validTypes))
}
