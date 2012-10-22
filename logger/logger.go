// package logger is a very light wrapper around 'log' so we can control IO.
// It can also print pretty colors.
package logger

import (
	"fmt"
	"log"
	"os"
)

import (
	. "github.com/str1ngs/ansi/color"
)

type logger struct {
	logType int
	plain   *log.Logger
	colored *log.Logger
}

const (
	FlagError   = 1 << iota // errors are very bad and unrecoverable
	FlagWarning             // warnings are bad things that don't stop us
	FlagMessage             // casual; typically describing state changes
	FlagLots                // more output than you could possibly need
	FlagDebug               // random debug output formatted differently
)

var (
	flags    = FlagDebug | FlagMessage | FlagWarning | FlagError
	logFlags = log.Lshortfile
	colors   = true
	levels   = []int{
		FlagDebug,
		FlagDebug | FlagError,
		FlagDebug | FlagError | FlagWarning,
		FlagDebug | FlagError | FlagWarning | FlagMessage,
		FlagDebug | FlagError | FlagWarning | FlagMessage | FlagLots,
	}

	Debug, Lots, Message, Warning, Error *logger
)

func init() {
	Debug = newLogger(
		FlagDebug,
		log.New(os.Stderr, "WINGO DEBUG: *** ", logFlags),
		log.New(os.Stderr,
			BgGreen(Blue("WINGO DEBUG:")).String()+" ", logFlags))
	Lots = newLogger(
		FlagLots,
		log.New(os.Stderr, "WINGO LOTS: ", logFlags),
		log.New(os.Stderr, "WINGO LOTS: ", logFlags))
	Message = newLogger(
		FlagMessage,
		log.New(os.Stderr, "WINGO MESSAGE: ", logFlags),
		log.New(os.Stderr, "WINGO MESSAGE: ", logFlags))
	Warning = newLogger(
		FlagWarning,
		log.New(os.Stderr, "WINGO WARNING: ", logFlags),
		log.New(os.Stderr,
			Bold(Red("WINGO WARNING:")).String()+" ", logFlags))
	Error = newLogger(
		FlagError,
		log.New(os.Stderr, "WINGO ERROR: ", logFlags),
		log.New(os.Stderr, BgMagenta("WINGO ERROR:").String()+" ", logFlags))
}

func newLogger(logType int, plain *log.Logger, colored *log.Logger) *logger {
	return &logger{logType, plain, colored}
}

func FlagsSet(newFlags int) {
	flags = newFlags
}

// LevelSet is a shortcut for setting log output flags. Valid log levels
// are integers in the range 0 to 4, inclusive. The higher the level, the
// more output.
func LevelSet(lvl int) {
	if lvl < 0 || lvl > 4 {
		panic("log level must be in [0, 4]")
	}
	FlagsSet(levels[lvl])
}

func Colors(enable bool) {
	colors = enable
}

func (lg *logger) Print(v ...interface{}) {
	if lg.logType&flags == 0 {
		return
	}

	if colors {
		lg.colored.Output(2, fmt.Sprint(v...))
	} else {
		lg.plain.Output(2, fmt.Sprint(v...))
	}
}

func (lg *logger) Printf(format string, v ...interface{}) {
	if lg.logType&flags == 0 {
		return
	}

	if colors {
		lg.colored.Output(2, fmt.Sprintf(format, v...))
	} else {
		lg.plain.Output(2, fmt.Sprintf(format, v...))
	}
}

func (lg *logger) Println(v ...interface{}) {
	if lg.logType&flags == 0 {
		return
	}

	if colors {
		lg.colored.Output(2, fmt.Sprintln(v...))
	} else {
		lg.plain.Output(2, fmt.Sprintln(v...))
	}
}

func (lg *logger) Fatal(v ...interface{}) {
	if lg.logType&flags == 0 {
		return
	}

	if colors {
		lg.colored.Output(2, fmt.Sprint(v...))
	} else {
		lg.plain.Output(2, fmt.Sprint(v...))
	}
	os.Exit(1)
}

func (lg *logger) Fatalf(format string, v ...interface{}) {
	if lg.logType&flags == 0 {
		return
	}

	if colors {
		lg.colored.Output(2, fmt.Sprintf(format, v...))
	} else {
		lg.plain.Output(2, fmt.Sprintf(format, v...))
	}
	os.Exit(1)
}

func (lg *logger) Fatalln(v ...interface{}) {
	if lg.logType&flags == 0 {
		return
	}

	if colors {
		lg.colored.Output(2, fmt.Sprintln(v...))
	} else {
		lg.plain.Output(2, fmt.Sprintln(v...))
	}
	os.Exit(1)
}

func (lg *logger) Panic(v ...interface{}) {
	if lg.logType&flags == 0 {
		return
	}

	if colors {
		lg.colored.Output(2, fmt.Sprint(v...))
	} else {
		lg.plain.Output(2, fmt.Sprint(v...))
	}
	panic("")
}

func (lg *logger) Panicf(format string, v ...interface{}) {
	if lg.logType&flags == 0 {
		return
	}

	if colors {
		lg.colored.Output(2, fmt.Sprintf(format, v...))
	} else {
		lg.plain.Output(2, fmt.Sprintf(format, v...))
	}
	panic("")
}

func (lg *logger) Panicln(v ...interface{}) {
	if lg.logType&flags == 0 {
		return
	}

	if colors {
		lg.colored.Output(2, fmt.Sprintln(v...))
	} else {
		lg.plain.Output(2, fmt.Sprintln(v...))
	}
	panic("")
}
