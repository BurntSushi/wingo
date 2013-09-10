package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path"

	"github.com/BurntSushi/xgbutil"

	"github.com/BurntSushi/wingo/commands"
	"github.com/BurntSushi/wingo/logger"
)

// ipc starts the command server via a unix domain socket. It accepts
// connections, reads Wingo commands, and attempts to execute them. If the
// command results in an error, it is sent back to the client. Otherwise, the
// return value of the command is sent to the user.
//
// Note that every message between the server and client MUST be null
// terminated.
func ipc(X *xgbutil.XUtil) {
	fpath := socketFilePath(X)

	// Remove the domain socket if it already exists.
	os.Remove(fpath) // don't care if there's an error

	listener, err := net.Listen("unix", fpath)
	if err != nil {
		logger.Error.Fatalln("Could not start IPC listener: %s", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Warning.Printf("Error accepting IPC connection: %s", err)
			continue
		}

		// Read the command from the connection. All messages are
		// null-terminated.
		go handleClient(conn)
	}
}

func socketFilePath(X *xgbutil.XUtil) string {
	xc := X.Conn()
	name := fmt.Sprintf(":%d.%d", xc.DisplayNumber, xc.DefaultScreen)

	var runtimeDir string
	xdgRuntime := os.Getenv("XDG_RUNTIME_DIR")
	if len(xdgRuntime) > 0 {
		runtimeDir = path.Join(xdgRuntime, "wingo")
	} else {
		runtimeDir = path.Join(os.TempDir(), "wingo")
	}

	if err := os.MkdirAll(runtimeDir, 0777); err != nil {
		logger.Error.Fatalf("Could not create directory '%s': %s",
			runtimeDir, err)
	}

	return path.Join(runtimeDir, name)
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	for {
		reader := bufio.NewReader(conn)
		msg, err := reader.ReadString(0)
		if err == io.EOF {
			return
		}
		if err != nil {
			logger.Warning.Printf("Error reading command '%s': %s", msg, err)
			return
		}
		msg = msg[:len(msg)-1] // get rid of null terminator

		logger.Lots.Printf("Running command from IPC: '%s'.", msg)

		// Run the command. We set the error reporting to verbose. Be kind!
		// If the command resulted in an error, we stop and send the error back
		// to the user. (This would be a Gribble parse/type error, not a
		// Wingo error.)
		commands.Env.Verbose = true
		val, err := commands.Env.RunMany(msg)
		commands.Env.Verbose = false
		if err != nil {
			logger.Lots.Printf("ERROR running command: '%s'.", err)
			fmt.Fprintf(conn, "ERROR: %s%c", err, 0)

			// One command failing doesn't mean we should close the conn.
			continue
		}

		// Fetch the return value of the command that was executed, and send
		// it back to the client. If the return value is nil, send an empty
		// response back. Otherwise, we need to type switch on all possible
		// return values.
		if val != nil {
			var retVal string
			switch v := val.(type) {
			case string:
				retVal = v
			case int:
				retVal = fmt.Sprintf("%d", v)
			case float64:
				retVal = fmt.Sprintf("%f", v)
			default:
				logger.Error.Fatalf("BUG: Unknown Gribble return type: %T", v)
			}
			fmt.Fprintf(conn, "%s%c", retVal, 0)
		} else {
			fmt.Fprintf(conn, "%c", 0)
		}
	}
}
