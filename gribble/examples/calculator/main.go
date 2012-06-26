package main

import (
	"fmt"

	"github.com/BurntSushi/wingo/gribble"
)

type Add struct {
	Op1 int
	Op2 int
}

func (c Add) Run() interface{} {
	return c.Op1 + c.Op2
}

type Multiply struct {
	Op1 int
	Op2 int
}

func (c Multiply) Run() interface{} {
	return c.Op1 * c.Op2
}

func main() {
	cmds := []gribble.Command{
		Add{},
		Multiply{},
	}
	env := gribble.New(cmds)
	fmt.Println(env)
}

