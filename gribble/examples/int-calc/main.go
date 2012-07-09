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
	name string `add`
	Op1  int    `param:"1"`
	Op2  int    `param:"2"`
}

func (c *Add) Run() gribble.Value {
	return c.Op1 + c.Op2
}

type Subtract struct {
	name string `sub`
	Op1  int    `param:"1"`
	Op2  int    `param:"2"`
}

func (c *Subtract) Run() gribble.Value {
	return c.Op1 - c.Op2
}

type Multiply struct {
	name string `mul`
	Op1  int    `param:"1"`
	Op2  int    `param:"2"`
}

func (c *Multiply) Run() gribble.Value {
	return c.Op1 * c.Op2
}

type Divide struct {
	name string `div`
	Op1  int    `param:"1"`
	Op2  int    `param:"2"`
}

func (c *Divide) Run() gribble.Value {
	return c.Op1 / c.Op2
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
