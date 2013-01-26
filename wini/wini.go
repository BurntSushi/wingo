/*
   Package wini provides an ini-like file parser. Namely, a wini parser.

   A wini file is very similar to a traditional ini file, except it doesn't
   have any quoting mechanism, uses ':=' instead of '=', and allows simple
   variables.

   Variables are extremely useful when theming.

   We also do our best to provide helpful error messages, so we can more
   precisely slap the user when they've gone wrong :-)

   This package is heavily inspired by glacjay's "goini" package:
   https://github.com/glacjay/goini
*/
package wini

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Data struct {
	data      map[string]Section // section -> option -> values
	variables map[string]string
}
type Section map[string]Value
type Value []string

type Key struct {
	data                      *Data
	section, key, niceSection string
}

var findVar *regexp.Regexp = regexp.MustCompile("\\$[a-zA-Z0-9_]+")

func Parse(filename string) (*Data, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data := &Data{
		data:      make(map[string]Section),
		variables: make(map[string]string),
	}
	reader := bufio.NewReader(file)

	section := "" // options not in a section are not allowed
	lnum := 0     // for nice error messages

	for {
		lnum += 1

		ln, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		// trim the fat
		line := strings.TrimSpace(string(ln))

		// If the line is a comment or empty, skip it
		if len(line) == 0 || line[0] == '#' || line[0] == ';' {
			continue
		}

		// If the line has a continuation, gobble up the rest
		for line[len(line)-1] == '\\' {
			line = line[:len(line)-1] // remove \\
			ln, _, err := reader.ReadLine()
			if err != nil {
				return nil, err
			}

			// just do a concatenation. Inefficient, but we don't really care.
			line += strings.TrimSpace(string(ln))
		}

		section, err = data.parseLine(section, line, lnum)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

func (d *Data) parseLine(section, line string, lnum int) (string, error) {
	// first check for a section name
	if line[0] == '[' && line[len(line)-1] == ']' {
		s := strings.TrimSpace(line[1 : len(line)-1])
		skey := strings.ToLower(s)

		// Make sure it's not empty
		if len(skey) == 0 {
			return "", winiError(lnum,
				"Section names must contain at least "+
					"one non-whitespace character.")
		}

		// if we've already seen this section, the user has been naughty
		if _, ok := d.data[skey]; ok {
			return "", winiError(lnum,
				"Section '%s' is defined again. "+
					"A section may only be defined once.", s)
		}

		// good to go, make the new section
		d.data[skey] = make(Section)
		return skey, nil
	}

	// Now check for a variable
	if line[0] == '$' {
		splitted := strings.SplitN(line, ":=", 2)
		if len(splitted) != 2 {
			return "", winiError(lnum,
				"Expected ':=' but could not find one in '%s'.",
				line)
		}

		varName := strings.TrimSpace(splitted[0][1:])
		varVal := strings.TrimSpace(splitted[1])

		// add it to the variable mapping --- it's okay if we overwrite!
		d.variables[varName] = varVal

		return section, nil
	}

	// now we're looking for 'key := val', which means we *must* have
	// a non-empty section
	if len(section) == 0 {
		return "", winiError(lnum, "Every option must belong to a '[section]'.")
	}

	// Okay, now actually parse 'key := val'
	splitted := strings.SplitN(line, ":=", 2)
	if len(splitted) != 2 {
		return "", winiError(lnum,
			"Expected ':=' but could not find one in '%s'.",
			line)
	}

	key := strings.TrimSpace(splitted[0])
	val := strings.TrimSpace(splitted[1])

	// If the key doesn't exist, allocate a slice
	if _, ok := d.data[section][key]; !ok {
		d.data[section][key] = make(Value, 0)
	}

	// good to go, add the new key!
	// don't forget to do variable replacement on val!
	d.data[section][key] = append(d.data[section][key],
		d.varReplace(val))

	return section, nil
}

func (d *Data) varReplace(val string) string {
	replace := func(varName string) string {
		if varVal, ok := d.variables[varName[1:]]; ok {
			return varVal
		}
		return val
	}
	return findVar.ReplaceAllStringFunc(val, replace)
}

func (d *Data) Sections() []string {
	sections := make([]string, len(d.data))

	i := 0
	for s, _ := range d.data {
		sections[i] = s
		i++
	}

	return sections
}

func (d *Data) GetKey(section, keyName string) *Key {
	skey := strings.ToLower(section)
	if keys, ok := d.data[skey]; ok {
		if _, ok := keys[keyName]; ok {
			return &Key{
				data:        d,
				section:     skey,
				key:         keyName,
				niceSection: section,
			}
		}
	}
	return nil
}

func (d *Data) Keys(section string) []Key {
	skey := strings.ToLower(section)
	if s, ok := d.data[skey]; ok {
		keys := make([]Key, len(s))
		i := 0
		for k, _ := range s {
			keys[i] = Key{data: d, section: skey, key: k,
				niceSection: section}
			i++
		}
		return keys
	}
	return nil
}

func (k Key) Strings() []string {
	return k.vals()
}

func (k Key) Bools() ([]bool, error) {
	bvals := make([]bool, len(k.vals()))
	for i, val := range k.vals() {
		v := strings.ToLower(val)[0]
		if v == 'y' || v == '1' || v == 't' {
			bvals[i] = true
			continue
		}
		if v == 'n' || v == '0' || v == 'f' {
			bvals[i] = false
			continue
		}
		return nil, k.Err("Not a valid boolean value: '%s'.", val)
	}
	return bvals, nil
}

func (k Key) Ints() ([]int, error) {
	ivals := make([]int, len(k.vals()))
	for i, val := range k.vals() {
		ival, err := strconv.ParseInt(val, 0, 0)
		if err != nil {
			return nil, k.Err("'%s' is not an integer. (%s)", val, err)
		}
		ivals[i] = int(ival)
	}
	return ivals, nil
}

func (k Key) Floats() ([]float64, error) {
	fvals := make([]float64, len(k.vals()))
	for i, val := range k.vals() {
		fval, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return nil, k.Err("'%s' is not a decimal. (%s)", val, err)
		}
		fvals[i] = fval
	}
	return fvals, nil
}

func (k Key) Name() string {
	return k.key
}

func (k Key) String() string {
	return fmt.Sprintf("(%s, %s)", k.niceSection, k.key)
}

func (k Key) vals() []string {
	return k.data.data[k.section][k.key]
}

func winiError(lnum int, formatted string, vals ...interface{}) error {
	msg := fmt.Sprintf(formatted, vals...)
	return errors.New(fmt.Sprintf("wini parse error on line %d: %s", lnum, msg))
}

func (k Key) Err(formatted string, vals ...interface{}) error {
	msg := fmt.Sprintf(formatted, vals...)
	return errors.New(fmt.Sprintf("There was an error reading the value for "+
		"the option '%s' in section '%s': %s",
		k.key, k.niceSection, msg))
}
