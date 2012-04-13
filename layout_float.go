package main

type floating struct {
	workspace *workspace
}

func newFloating(wrk *workspace) *floating {
	return &floating{
		workspace: wrk,
	}
}

func (ly *floating) place()           {}
func (ly *floating) unplace()         {}
func (ly *floating) add(c *client)    {}
func (ly *floating) remove(c *client) {}
