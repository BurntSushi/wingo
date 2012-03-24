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
    "strings"
)

type Data map[string]map[string]string // section -> option -> value

func Parse(filename string) (Data, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    data := make(Data)
    reader := bufio.NewReader(file)

    section := "" // options not in a section are not allowed
    lnum := 0 // for nice error messages

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
        if line[len(line) - 1] == '\\' {
            line = line[:len(line) - 1] // remove \\
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

func (d Data) parseLine(section, line string, lnum int) (string, error) {
    // first check for a section name
    if line[0] == '[' && line[len(line) - 1] == ']' {
        s := strings.TrimSpace(line[1:len(line) - 1])
        skey := strings.ToLower(s)

        // Make sure it's not empty
        if len(skey) == 0 {
            return "", winiError(lnum,
                                 "Section names must contain at least " +
                                 "one non-whitespace character.")
        }

        // if we've already seen this section, the user has been naughty
        if _, ok := d[skey]; ok {
            return "", winiError(lnum,
                                 "Section '%s' is defined again. " +
                                 "A section may only be defined once.", s)
        }

        // good to go, make the new section
        d[skey] = make(map[string]string)
        return skey, nil
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

    key, val := strings.TrimSpace(splitted[0]), strings.TrimSpace(splitted[1])

    // If the key already exists, slap the user
    if _, ok := d[section][key]; ok {
        return "", winiError(lnum, "'%s' was already defined.", key)
    }

    // good to go, add the new key!
    d[section][key] = val

    return section, nil
}

func (d Data) SectionsGet() []string {
    sections := make([]string, len(d))

    i := 0
    for s, _ := range d {
        sections[i] = s
        i++
    }

    return sections
}

func winiError(lnum int, formatted string, vals... interface{}) error {
    msg := fmt.Sprintf(formatted, vals...)
    return errors.New(fmt.Sprintf("wini parse error on line %d: %s", lnum, msg))
}

