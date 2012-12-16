package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"runtime"
	"runtime/pprof"
	"strings"
	"syscall"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/mousebind"
	"github.com/BurntSushi/xgbutil/xevent"

	"github.com/BurntSushi/wingo/commands"
	"github.com/BurntSushi/wingo/cursors"
	"github.com/BurntSushi/wingo/focus"
	"github.com/BurntSushi/wingo/hook"
	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/misc"
	"github.com/BurntSushi/wingo/stack"
	"github.com/BurntSushi/wingo/wm"
	"github.com/BurntSushi/wingo/xclient"
)

var (
	flagGoMaxProcs     = runtime.NumCPU()
	flagLogLevel       = 2
	flagLogColors      = false
	flagReplace        = false
	flagConfigDir      = ""
	flagDataDir        = ""
	flagWriteConfig    = false
	flagCpuProfile     = ""
	flagWingoRestarted = false
)

func init() {
	flag.IntVar(&flagGoMaxProcs, "p", flagGoMaxProcs,
		"The maximum number of CPUs that can be executing simultaneously.")
	flag.IntVar(&flagLogLevel, "log-level", flagLogLevel,
		"The logging level of Wingo. Valid values are 0, 1, 2, 3 or 4.\n"+
			"Higher numbers result in Wingo producing more output.")
	flag.BoolVar(&flagLogColors, "log-colors", flagLogColors,
		"Whether to output logging data with terminal colors.")
	flag.BoolVar(&flagReplace, "replace", flagReplace,
		"When set, Wingo will attempt to replace a currently running\n"+
			"window manager. If this is not set, and another window manager\n"+
			"is running, Wingo will exit.")
	flag.StringVar(&flagConfigDir, "config-dir", flagConfigDir,
		"Override the location of the configuration files. When this\n"+
			"is not set, the following paths (roughly) will be checked\n"+
			"in order: $XDG_CONFIG_DIR/wingo, /etc/xdg/wingo,\n"+
			"$GOPATH/src/github.com/BurntSushi/wingo/config")
	flag.StringVar(&flagDataDir, "data-dir", flagDataDir,
		"Override the location of the data files (images/fonts). When this\n"+
			"is not set, the following paths (roughly) will be checked\n"+
			"in order: $XDG_DATA_HOME/wingo, /usr/local/share, /usr/share,\n"+
			"$GOPATH/src/github.com/BurntSushi/wingo/data")
	flag.BoolVar(&flagWriteConfig, "write-config", flagWriteConfig,
		"Writes a fresh set of configuration files to $XDG_CONFIG_HOME/wingo\n"+
			"if XDG_CONFIG_HOME is set. Otherwise, configuration files\n"+
			"are written to $HOME/.config/wingo.\n"+
			"This will fail if the 'wingo' configuration directory already\n"+
			"exists, to prevent accidentally overwriting an existing\n"+
			"configuration.\n\n"+
			"When this flag is set, Wingo will not start.")
	flag.BoolVar(&flagWingoRestarted, "wingo-restarted", flagWingoRestarted,
		"DO NOT USE. INTERNAL WINGO USE ONLY.")

	flag.StringVar(&flagCpuProfile, "cpuprofile", flagCpuProfile,
		"When set, a CPU profile will be written to the file specified.")

	flag.Usage = usage
	flag.Parse()

	runtime.GOMAXPROCS(flagGoMaxProcs)
	logger.Colors(flagLogColors)
	logger.LevelSet(flagLogLevel)

	// If the log level is 0, don't show XGB log output either.
	if flagLogLevel == 0 {
		xgb.PrintLog = false
	}
}

func main() {
	if flagWriteConfig {
		writeConfigFiles()
		os.Exit(0)
	}

	X, err := xgbutil.NewConn()
	if err != nil {
		logger.Error.Println(err)
		logger.Error.Fatalln("Error connecting to X, quitting...")
	}
	defer X.Conn().Close()

	// Do this first! Attempt to retrieve window manager ownership.
	// This includes waiting for any existing window manager to die.
	// 'own' also sets up handlers for quitting when a window manager tries
	// to replace *us*.
	if err := own(X, flagReplace); err != nil {
		logger.Error.Fatalf(
			"Could not establish window manager ownership: %s", err)
	}

	if len(flagConfigDir) > 0 {
		misc.ConfigPaths.Override = flagConfigDir
	}
	if len(flagDataDir) > 0 {
		misc.DataPaths.Override = flagDataDir
	}
	misc.ReadData()

	keybind.Initialize(X)
	mousebind.Initialize(X)
	focus.Initialize(X)
	stack.Initialize(X)
	cursors.Initialize(X)
	wm.Initialize(X, commands.Env, newHacks())
	hook.Initialize(commands.Env, misc.ConfigFile("hooks.wini"))

	// Initialize event handlers on the root window.
	rootInit(X)

	// Tell everyone what we support.
	setSupported()

	// Start up the IPC command listener.
	go ipc()

	// Just before starting the main event loop, check to see if there are
	// any clients that already exist that we should manage.
	manageExistingClients()

	// Now make sure that clients are in the appropriate visible state.
	for _, wrk := range wm.Heads.Workspaces.Wrks {
		if wrk.IsVisible() {
			wrk.Show()
		} else {
			wrk.Hide()
		}
	}
	wm.Heads.ApplyStruts(wm.Clients)

	wm.FocusFallback()
	wm.Startup = false
	pingBefore, pingAfter, pingQuit := xevent.MainPing(X)

	if len(flagCpuProfile) > 0 {
		f, err := os.Create(flagCpuProfile)
		if err != nil {
			logger.Error.Fatalf("%s\n", err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if flagWingoRestarted {
		hook.Fire(hook.Restarted, hook.Args{})
	} else {
		hook.Fire(hook.Startup, hook.Args{})
	}

EVENTLOOP:
	for {
		select {
		case <-pingBefore:
			// Wait for the event to finish processing.
			<-pingAfter
		case f := <-commands.SafeExec:
			commands.SafeReturn <- f()
		case <-pingQuit:
			break EVENTLOOP
		}
	}
	if wm.Restart {
		// We need to tell the next invocation of Wingo that it is being
		// *restarted*. (So that we don't refire the startup hook.)
		// Thus, search os.Args for "--wingo-restarted". If it doesn't exist,
		// add it.
		found := false
		for _, arg := range os.Args {
			if strings.ToLower(strings.TrimSpace(arg)) == "--wingo-restarted" {
				found = true
			}
		}
		if !found {
			os.Args = append(os.Args, "--wingo-restarted")
		}
		logger.Message.Println("The user has told us to restart...\n\n\n")
		if err := syscall.Exec(os.Args[0], os.Args, os.Environ()); err != nil {
			logger.Error.Fatalf("Could not exec '%s': %s",
				strings.Join(os.Args, " "), err)
		}
	}
}

func setSupported() {
	// See COMPLIANCE for what's supported and what isn't (plus rationale).
	// ewmhSupported comes from ewmh_supported.go and is automatically
	// generated from the COMPLIANCE file.

	// Set supported atoms
	ewmh.SupportedSet(wm.X, ewmhSupported)

	// While we're at it, set the supporting wm hint too.
	ewmh.SupportingWmCheckSet(wm.X, wm.X.RootWin(), wm.X.Dummy())
	ewmh.SupportingWmCheckSet(wm.X, wm.X.Dummy(), wm.X.Dummy())
	ewmh.WmNameSet(wm.X, wm.X.Dummy(), "Wingo")
}

// manageExistingClients traverse the window tree and tries to manage all
// top-level clients. Clients that are not in the Unmapped state will be
// managed.
func manageExistingClients() {
	tree, err := xproto.QueryTree(wm.X.Conn(), wm.Root.Id).Reply()
	if err != nil {
		logger.Warning.Printf("Could not issue QueryTree request: %s", err)
		return
	}
	for _, potential := range tree.Children {
		// Ignore our own dummy window...
		if potential == wm.X.Dummy() {
			continue
		}

		attrs, err := xproto.GetWindowAttributes(wm.X.Conn(), potential).Reply()
		if err != nil {
			continue
		}
		if attrs.MapState == xproto.MapStateUnmapped {
			continue
		}
		logger.Message.Printf("Managing existing client %d", potential)
		xclient.New(potential)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "\nUsage: %s [flags]\n", path.Base(os.Args[0]))
	flag.VisitAll(func(fg *flag.Flag) {
		// Don't let users know about flags they shouldn't use.
		if fg.Name == "wingo-restarted" {
			return
		}
		fmt.Printf("--%s=\"%s\"\n\t%s\n", fg.Name, fg.DefValue,
			strings.Replace(fg.Usage, "\n", "\n\t", -1))
	})
	os.Exit(1)
}
