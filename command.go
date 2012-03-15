// A set of functions that are key-bindable
package main

func cmd_close_active() {
    focused := WM.focused()
    if focused != nil {
        logMessage.Printf("### Should be umanaging %s", focused)
        focused.close_()
        logMessage.Printf("### Did we?")
    }
}

