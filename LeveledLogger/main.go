package LeveledLogger

import (
    "fmt"
    "io"
    "log"
    "time"
)

const (
    LL_DEBUG  = 10 - iota
    LL_INFO   = 10 - iota
    LL_CALL   = 10 - iota
    LL_ACTION = 10 - iota
    LL_WARN   = 10 - iota
    LL_ERROR  = 10 - iota
)

type Logger struct {
    level  int
    logger *log.Logger
}

func (ll *Logger) print(level, fname, msg string, v ...interface{}) {
    format := fmt.Sprintf("%d#%s#%s#%s", time.Now().Unix(), level, fname, msg)
    for i := 0; i < len(v); i++ {
        format += "#%v"
    }
    ll.logger.Printf(format, v...)
}

func (ll *Logger) Debug(fname, msg string, fargs ...interface{}) {
    if ll.level >= LL_DEBUG {
        ll.print("DBG", fname, msg, fargs...)
    }
}

func (ll *Logger) Info(fname, msg string, fargs ...interface{}) {
    if ll.level >= LL_INFO {
        ll.print("INF", fname, msg, fargs...)
    }
}

func (ll *Logger) Call(fname string, fargs ...interface{}) {
    if ll.level >= LL_CALL {
        ll.print("CAL", fname, "", fargs...)
    }
}

func (ll *Logger) Action(fname string, fargs ...interface{}) {
    if ll.level >= LL_CALL {
        ll.print("ACT", fname, "", fargs...)
    }
}

func (ll *Logger) Warn(fname, msg string, fargs ...interface{}) {
    if ll.level >= LL_DEBUG {
        ll.print("WRN", fname, msg, fargs...)
    }
}

func (ll *Logger) Error(fname, msg string, fargs ...interface{}) {
    if ll.level >= LL_ERROR {
        ll.print("ERR", fname, msg, fargs...)
        panic(msg)
    }
}

func New(out io.Writer, level int) *Logger {
    return &Logger{
        level:  level,
        logger: log.New(out, "", 0),
    }
}
