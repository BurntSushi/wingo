// A set of functions that are key-bindable
package main

import (
    "bytes"
    "os/exec"
    "strings"
    "time"
    "unicode"
)

// commandShellFun takes a command specified in a configuration file and
// tries to parse it as an executable command. The command must be wrapped
// in "`" and "`" (back-quotes). If it's not, we return nil. Otherwise, we
// return a function that will execute the command.
// This provides rudimentary support for quoted values in the command.
func commandShellFun(cmd string) func() {
    if cmd[0] != '`' || cmd[len(cmd) - 1] != '`' {
        return nil
    }

    return func() {
        var stderr bytes.Buffer

        allCmd := cmd[1:len(cmd) - 1]

        splitCmdName := strings.SplitN(allCmd, " ", 2)
        cmdName := splitCmdName[0]
        args := make([]string, 0)
        addArg := func(start, end int) {
            args = append(args, strings.TrimSpace(splitCmdName[1][start:end]))
        }

        if len(splitCmdName) > 1 {
            startArgPos := 0
            inQuote := false
            for i, char := range splitCmdName[1] {
                // Add arguments enclosed in quotes
                // Yes, this mixes up quotes.
                if char == '"' || char == '\'' {
                    inQuote = !inQuote

                    if !inQuote {
                        addArg(startArgPos, i)
                    }
                    startArgPos = i + 1 // skip the end quote character
                }

                // Add arguments separated by spaces without quotes
                if !inQuote && unicode.IsSpace(char) {
                    addArg(startArgPos, i)
                    startArgPos = i
                }
            }

            // add anything that's left over
            addArg(startArgPos, len(splitCmdName[1]))
        }

        cmd := exec.Command(cmdName, args...)
        cmd.Stderr = &stderr

        err := cmd.Run()
        if err != nil {
            logWarning.Printf("Error running '%s': %s", allCmd, err)
            if stderr.Len() > 0 {
                logWarning.Printf("Error running '%s': %s",
                                  allCmd, stderr.String())
            }
        }
    }
}


// This is a start, but it is not quite ready for prime-time yet.
// 1. If the window is destroyed while the go routine is still running,
// we're in big trouble.
// 2. This has no way to stop from some external event (like focus).
// Basically, both of these things can be solved by using channels to tell
// the goroutine to quit. Not difficult but also not worth my time atm.
func cmd_active_flash() {
    focused := WM.focused()

    if focused != nil {
        go func(c *client) {
            for i := 0; i < 10; i++ {
                if c.Frame().State() == StateActive {
                    c.Frame().Inactive()
                } else {
                    c.Frame().Active()
                }

                time.Sleep(time.Second)
            }
        }(focused)
    }
}

