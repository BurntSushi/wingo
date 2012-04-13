package main

type layout interface {
	place()
	unplace()
	add(c *client)
	remove(c *client)
}
