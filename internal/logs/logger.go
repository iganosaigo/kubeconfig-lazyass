package logs

import (
	"fmt"
	"log"
	"os"
)

type Logger struct {
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		info:  log.New(os.Stdout, "INFO: ", log.LstdFlags),
		warn:  log.New(os.Stdout, "WARN: ", log.LstdFlags),
		error: log.New(os.Stderr, "ERROR: ", log.LstdFlags),
	}
}

func (l *Logger) Info(v ...interface{}) {
	l.info.Println(v...)
}

func (l *Logger) Warn(v ...interface{}) {
	l.warn.Println(v...)
}

func (l *Logger) Error(v ...interface{}) {
	l.error.Println(v...)
}

func (l *Logger) Fatal(exitCode int, v ...interface{}) {
	l.Error(v...)
	os.Exit(exitCode)
}

func (l *Logger) Fatalf(exitCode int, format string, v ...interface{}) {
	l.Error(fmt.Sprintf(format, v...))
	os.Exit(exitCode)
}
