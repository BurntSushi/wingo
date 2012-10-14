package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path"

	"github.com/BurntSushi/wingo/commands"
	"github.com/BurntSushi/wingo/logger"
)

func ipc() {
	fpath := path.Join(os.TempDir(), "wingo-ipc")

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

		reader := bufio.NewReader(conn)
		msg, err := reader.ReadString(0)
		if err != nil {
			logger.Warning.Printf("Error reading command '%s': %s", msg, err)
			continue
		}
		msg = msg[:len(msg)-1] // get rid of null terminator

		logger.Message.Printf("Running command from IPC: '%s'.", msg)

		commands.Env.Verbose = true
		val, err := commands.Env.RunMany(msg)
		commands.Env.Verbose = false
		if err != nil {
			fmt.Fprintf(conn, "%s%c", err, 0)
			continue
		}

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
