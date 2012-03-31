// A very light wrapper around 'log' so we can control IO.
// XXX: Only 'Print*' functions have been wrapped. The rest of the methods
// directly use those provided by 'log'.
package main

import (
    "log"
    "os"
)

type WingoLogger struct {
    logType int
    *log.Logger
}

const (
    DebugError = 1 << iota // errors are very bad and unrecoverable
    DebugWarning // warnings are bad things that don't stop us from chugging
    DebugMessage // casual; typically describing state changes
    DebugLots // more output than you could possibly need
    DebugDebug // random debug output formatted differently
)

var DebugFlags int = DebugDebug | DebugMessage | DebugWarning | DebugError

var logDebug, logLots, logMessage, logWarning, logError *WingoLogger

func init() {
    logDebug = newWingoLogger(DebugDebug,
                              log.New(os.Stderr, "*** WINGO DEBUG: ",
                                      log.Ldate | log.Ltime))
    logLots = newWingoLogger(DebugLots,
                             log.New(os.Stderr, "WINGO LOTS: ",
                                     log.Ldate | log.Ltime))
    logMessage = newWingoLogger(DebugMessage,
                                log.New(os.Stderr, "WINGO MESSAGE: ",
                                        log.Ldate | log.Ltime))
    logWarning = newWingoLogger(DebugWarning,
                                log.New(os.Stderr, "WINGO WARNING: ",
                                        log.Ldate | log.Ltime))
    logError = newWingoLogger(DebugError,
                              log.New(os.Stderr, "WINGO ERROR: ",
                                      log.Ldate | log.Ltime | log.Lshortfile))
}

func newWingoLogger(logType int, logger *log.Logger) *WingoLogger {
    return &WingoLogger{logType, logger}
}

func (wl *WingoLogger) Print(v ...interface{}) {
    if wl.logType & DebugFlags == 0 {
        return
    }

    wl.Logger.Print(v...)
}

func (wl *WingoLogger) Printf(format string, v ...interface{}) {
    if wl.logType & DebugFlags == 0 {
        return
    }

    wl.Logger.Printf(format, v...)
}

func (wl *WingoLogger) Println(v ...interface{}) {
    if wl.logType & DebugFlags == 0 {
        return
    }

    wl.Logger.Println(v...)
}

