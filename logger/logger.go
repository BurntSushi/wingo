// Package logger is a very light wrapper around 'log' so we can control IO.
// It can also print pretty colors.
// XXX: Only 'Print*' functions have been wrapped. The rest of the methods
// directly use those provided by 'log'.
package logger

import (
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
	flags  = FlagDebug | FlagMessage | FlagWarning | FlagError
	colors = true

	Debug, Lots, Message, Warning, Error *logger
)

func init() {
	Debug = newLogger(
		FlagDebug,
		log.New(os.Stderr, "WINGO DEBUG: *** ", log.Ldate|log.Ltime),
		log.New(os.Stderr,
			BgGreen(Blue("WINGO DEBUG:")).String()+" ",
			log.Ldate|log.Ltime))
	Lots = newLogger(
		FlagLots,
		log.New(os.Stderr, "WINGO LOTS: ", log.Ldate|log.Ltime),
		log.New(os.Stderr, "WINGO LOTS: ", log.Ldate|log.Ltime))
	Message = newLogger(
		FlagMessage,
		log.New(os.Stderr, "WINGO MESSAGE: ", log.Ldate|log.Ltime),
		log.New(os.Stderr, "WINGO MESSAGE: ", log.Ldate|log.Ltime))
	Warning = newLogger(
		FlagWarning,
		log.New(os.Stderr, "WINGO WARNING: ", log.Ldate|log.Ltime),
		log.New(os.Stderr,
			Bold(Red("WINGO WARNING:")).String()+" ",
			log.Ldate|log.Ltime))
	Error = newLogger(
		FlagError,
		log.New(os.Stderr, "WINGO ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		log.New(os.Stderr,
			BgMagenta("WINGO ERROR:").String()+" ",
			log.Ldate|log.Ltime|log.Lshortfile))
}

func newLogger(logType int, plain *log.Logger, colored *log.Logger) *logger {
	return &logger{logType, plain, colored}
}

func FlagsSet(newFlags int) {
	flags = newFlags
}

func Colors(enable bool) {
	colors = enable
}

func (lg *logger) Print(v ...interface{}) {
	if lg.logType&flags == 0 {
		return
	}

	if colors {
		lg.colored.Print(v...)
	} else {
		lg.plain.Print(v...)
	}
}

func (lg *logger) Printf(format string, v ...interface{}) {
	if lg.logType&flags == 0 {
		return
	}

	if colors {
		lg.colored.Printf(format, v...)
	} else {
		lg.plain.Printf(format, v...)
	}
}

func (lg *logger) Println(v ...interface{}) {
	if lg.logType&flags == 0 {
		return
	}

	if colors {
		lg.colored.Println(v...)
	} else {
		lg.plain.Println(v...)
	}
}
