package main

type layout interface {
	floating() bool
	place()
	unplace()
	add(c *client)
	remove(c *client)
	maximizable() bool
	move(c *client, x, y int)
	resize(c *client, w, h int)
	moveresize(c *client, x, y, w, h int)
}
