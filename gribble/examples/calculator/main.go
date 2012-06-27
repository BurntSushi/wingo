package main

import (
	"fmt"

	"github.com/BurntSushi/wingo/gribble"
)

type Add struct {
	Op1 int `param:"1"`
	Op2 int `param:"2"`
	Cmd gribble.Command `param:"3"`
}

func (c Add) Run() gribble.Value {
	return gribble.Int(c.Op1 + c.Op2)
}

type Multiply struct {
	Op1 int `param:"1"`
	Op2 int `param:"2"`
}

func (c Multiply) Run() gribble.Value {
	return gribble.Int(c.Op1 * c.Op2)
}

func main() {
	cmds := []gribble.Command{
		Add{},
		Multiply{},
	}
	env := gribble.New(cmds)
	fmt.Println(env.StringTypes())
}

