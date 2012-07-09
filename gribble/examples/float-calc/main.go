package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/BurntSushi/wingo/gribble"
)

var env *gribble.Environment = gribble.New([]gribble.Command{
	&Add{},
	&Subtract{},
	&Multiply{},
	&Divide{},
})

type Add struct {
	name string      `add`
	Op1  gribble.Any `param:"1" types:"int,float"`
	Op2  gribble.Any `param:"2" types:"int,float"`
}

func (c *Add) Run() gribble.Value {
	return float(c.Op1) + float(c.Op2)
}

type Subtract struct {
	name string      `sub`
	Op1  gribble.Any `param:"1" types:"int,float"`
	Op2  gribble.Any `param:"2" types:"int,float"`
}

func (c *Subtract) Run() gribble.Value {
	return float(c.Op1) - float(c.Op2)
}

type Multiply struct {
	name string      `mul`
	Op1  gribble.Any `param:"1" types:"int,float"`
	Op2  gribble.Any `param:"2" types:"int,float"`
}

func (c *Multiply) Run() gribble.Value {
	return float(c.Op1) * float(c.Op2)
}

type Divide struct {
	name string      `div`
	Op1  gribble.Any `param:"1" types:"int,float"`
	Op2  gribble.Any `param:"2" types:"int,float"`
}

func (c *Divide) Run() gribble.Value {
	return float(c.Op1) / float(c.Op2)
}

func float(val interface{}) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	}
	// This is unreachable because the values allowed in each of 'Op1' and 'Op2'
	// are allowed to be 'int' or 'float'. This invariant is enforced by
	// Gribble. That is, if this panic is hit, Gribble has a bug.
	panic("unreachable")
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s command\n", path.Base(os.Args[0]))
	flag.PrintDefaults()

	fmt.Fprintln(os.Stderr, "\nAvailable commands:")
	fmt.Fprintln(os.Stderr, env.StringTypes())
	os.Exit(1)
}

func main() {
	flag.Usage = usage
	flag.Parse()
	cmd := strings.Join(flag.Args(), " ")

	val, err := env.Run(cmd)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)

		if name := env.CommandName(cmd); len(name) > 0 {
			fmt.Printf("Usage: %s\n", env.UsageTypes(name))
		}
		return
	}
	fmt.Println(val)
}
