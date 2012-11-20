/*
Package text provides text rendering helper functions and an abstraction to
create input text windows.

The DrawText rendering function included takes a window, font information and
text and creates an image with the text written on it. The image is then
painted to the window provided, and the window is resized to "snugly" fit the
text. (Note that DrawText has a subtle bug that will manifest itself with large
font sizes. Please see the bugs section.)

The other useful part of this package is the Input window type. It is an
abstraction over xwindow.Window that provides an input box like window. The
Input type's methods can then be used to easily add and remove text from the
input box in response to KeyPress events. (You must write the KeyPress event
handler.)

Here's a minimal example for creating an input window and allowing the user to
type into it:

	input := text.NewInput(XUtilValue, RootId, 500, 0, font, 20.0,
		textColor, bgColor)
	input.Listen(xproto.EventMaskKeyPress)
	xevent.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			if keybind.KeyMatch(X, "BackSpace", ev.State, ev.Detail) {
				input.Remove()
			} else {
				input.Add(ev.State, ev.Detail)
			}
		}).Connect(X, input.Id)

Since the Input type embeds an xwindow.Window, it can be thought of as a
regular window with special methods for handling text display.

Note that a slightly more involved and working example can be found in
text/examples/input/main.go.
*/
package text
