package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/wingo/commands"
)

var (
	flagFileInput         = ""
	flagListCommands      = false
	flagListTypeCommands  = false
	flagListUsageCommands = false
	flagPoll              = 0
	flagUsageCommand      = ""
)

func init() {
	flag.StringVar(&flagFileInput, "f", flagFileInput,
		"When set, commands will be read from the specified file.\n"+
			"If '-' is used, commands will be read from stdin.")
	flag.BoolVar(&flagListCommands, "list", flagListCommands,
		"Print a list of all commands and their parameters.")
	flag.BoolVar(&flagListTypeCommands, "list-types", flagListTypeCommands,
		"Print a list of all commands and their parameters (with type info).")
	flag.BoolVar(&flagListUsageCommands, "list-usage", flagListUsageCommands,
		"Print a list of all commands, their parameters (with type info),\n"+
			"and usage information for each command.")
	flag.StringVar(&flagUsageCommand, "usage", flagUsageCommand,
		"Print usage information for a particular command.")
	flag.IntVar(&flagPoll, "poll", flagPoll,
		"When greater than 0, the commands specified will be repeated at "+
			"the interval specified in milliseconds.")

	flag.Usage = usage
	flag.Parse()

	log.SetFlags(0)
}

func main() {
	// If a list/usage flag is set, print the stuff and exit.
	handleFlags()

	// Get the commands from file/stdin/argument.
	cmds := getCommands()

	// Connect to the Wingo command server.
	conn, err := net.Dial("unix", path.Join(os.TempDir(), "wingo-ipc"))
	if err != nil {
		log.Fatalf("Could not connect to Wingo IPC: %s", err)
	}

	// If the 'poll' flag is set, then we'll need to send the command and
	// print the result repeatedly.
	for {
		// Send the command.
		if _, err = fmt.Fprintf(conn, "%s%c", cmds, 0); err != nil {
			log.Fatalf("Error writing command: %s", err)
		}

		// Read the response.
		reader := bufio.NewReader(conn)
		msg, err := reader.ReadString(0)
		if err != nil {
			log.Fatalf("Could not read response: %s", err)
		}
		msg = msg[:len(msg)-1] // get rid of null terminator

		fmt.Println(msg)

		if flagPoll == 0 {
			break
		} else {
			time.Sleep(time.Duration(flagPoll) * time.Millisecond)
		}
	}
}

// getCommands inspects the arguments to extra a series of commands. If the
// '-f' flag is set, then the commands are read from a file or stdin ("-").
// Otherwise, the command is read from the first command line argument.
func getCommands() (cmds string) {
	// If '-f' is set, use commands from the file specified.
	// Otherwise, make sure there is one and only one argument (the command).
	if len(flagFileInput) > 0 {
		var contents []byte
		var err error

		if flagFileInput == "-" { // stdin
			contents, err = ioutil.ReadAll(os.Stdin)
			if err != nil {
				log.Fatalf("Could not read stdin: %s", err)
			}
		} else {
			contents, err = ioutil.ReadFile(flagFileInput)
			if err != nil {
				log.Fatalf("Could not read file '%s': %s", flagFileInput, err)
			}
		}

		// Ignore any line starting with '#' after trimming.
		lines := make([]string, 0)
		for _, line := range bytes.Split(contents, []byte{'\n'}) {
			line = bytes.TrimSpace(line)
			if len(line) == 0 || line[0] == '#' {
				continue
			}
			lines = append(lines, string(line))
		}
		cmds = strings.Join(lines, "\n")
	} else {
		if flag.NArg() != 1 {
			log.Printf("Expected 1 argument but got %d arguments.", flag.NArg())
			usage()
		}
		cmds = flag.Arg(0)
	}
	return
}

// handleFlags inspects the list/usage flags, and will print the appropriate
// output and exit with success if one of them is set.
func handleFlags() {
	switch {
	case flagListCommands:
		fmt.Println(commands.Env.String())
		os.Exit(0)
	case flagListTypeCommands:
		fmt.Println(commands.Env.StringTypes())
		os.Exit(0)
	case flagListUsageCommands:
		cmds := make([]string, 0)
		commands.Env.Each(func(name, help string) {
			usage := commands.Env.UsageTypes(name)
			help = strings.Replace(help, "\n", "\n\t", -1)

			if len(help) > 0 {
				cmds = append(cmds, fmt.Sprintf("%s\n\t%s\n", usage, help))
			} else {
				cmds = append(cmds, usage)
			}
		})
		sort.Sort(sort.StringSlice(cmds))
		fmt.Println(strings.Join(cmds, "\n"))
		os.Exit(0)
	case len(flagUsageCommand) > 0:
		fmt.Println(commands.Env.UsageTypes(flagUsageCommand))

		help := commands.Env.Help(flagUsageCommand)
		fmt.Printf("\t%s\n", strings.Replace(help, "\n", "\n\t", -1))
		os.Exit(0)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "\nUsage: %s [flags] [command]\n",
		path.Base(os.Args[0]))
	flag.VisitAll(func(fg *flag.Flag) {
		fmt.Printf("--%s=\"%s\"\n\t%s\n", fg.Name, fg.DefValue,
			strings.Replace(fg.Usage, "\n", "\n\t", -1))
	})
	os.Exit(1)
}
