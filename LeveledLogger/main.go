package LeveledLogger

import (
    "fmt"
    "io"
    "log"
)

const (
    LL_DEBUG = 10
    LL_INFO  = 5
    LL_ERROR = 0
)

type Logger struct {
    level  int
    logger *log.Logger
}

func (ll *Logger) printf(level string, format string, v ...interface{}) {
    ll.logger.Printf("["+level+"] "+format, v...)
}

func (ll *Logger) Debugf(format string, v ...interface{}) {
    if ll.level >= LL_DEBUG {
        ll.printf("DBG", format, v...)
    }
}

func (ll *Logger) Infof(format string, v ...interface{}) {
    if ll.level >= LL_INFO {
        ll.printf("INF", format, v...)
    }
}

func (ll *Logger) Errorf(format string, v ...interface{}) {
    if ll.level >= LL_ERROR {
        ll.printf("ERR", format, v...)
        panic(fmt.Sprintf(format, v...))
    }
}

func New(out io.Writer, level int) *Logger {
    return &Logger{
        level:  level,
        logger: log.New(out, "", log.Ldate|log.Ltime),
    }
}
