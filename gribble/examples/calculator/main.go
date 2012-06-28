package main

import (
	"fmt"

	"github.com/BurntSushi/wingo/gribble"
)

func wrapFloat(val interface{}) float64 {
	switch v := val.(type) {
	case int:
		return float64(v)
	case float64:
		return v
	}
	panic("unreachable")
}

type Const struct {}

func (c *Const) Run() gribble.Value {
	return 5
}

type Add struct {
	name string `add`
	Op1 interface{} `param:"1" types:"int,float"`
	Op2 interface{} `param:"2" types:"int,float"`
}

func (c *Add) Run() gribble.Value {
	return wrapFloat(c.Op1) + wrapFloat(c.Op2)
}

type Subtract struct {
	name string `sub`
	Op1 interface{} `param:"1" types:"int,float"`
	Op2 interface{} `param:"2" types:"int,float"`
}

func (c *Subtract) Run() gribble.Value {
	return wrapFloat(c.Op1) - wrapFloat(c.Op2)
}

type Multiply struct {
	name string `mul`
	Op1 interface{} `param:"1" types:"int,float"`
	Op2 interface{} `param:"2" types:"int,float"`
}

func (c *Multiply) Run() gribble.Value {
	return wrapFloat(c.Op1) * wrapFloat(c.Op2)
}

type Divide struct {
	name string `div`
	Op1 interface{} `param:"1" types:"int,float"`
	Op2 interface{} `param:"2" types:"int,float"`
}

func (c *Divide) Run() gribble.Value {
	return wrapFloat(c.Op1) / wrapFloat(c.Op2)
}

func test(env *gribble.Environment, cmd string) {
	command, err := env.Command(cmd)
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}
	val := command.Run()
	fmt.Printf("SUCCESS: (%T) %v\n", val, val)
}

func testRun(env *gribble.Environment, cmd string) {
	val, err := env.Run(cmd)
	if err != nil {
		fmt.Println("ERROR:", err)
		fmt.Printf("USAGE: %s\n", env.UsageTypes(env.CommandName(cmd)))
		return
	}
	fmt.Printf("SUCCESS: (%T) %v\n", val, val)
}

func main() {
	cmds := []gribble.Command{
		&Const{},
		&Add{},
		&Subtract{},
		&Multiply{},
		&Divide{},
	}
	env := gribble.New(cmds)
	fmt.Println(env.StringTypes())

	testRun(env, "sub 1 (div 26 (add 5 (mul 2 4)))")
	testRun(env, "div 7 10")
	testRun(env, "mul 500 (Const)")
}

