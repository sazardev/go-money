package logger

import (
	"log"
	"os"
	"sync"
)

var (
	defaultLogger Logger
	once          sync.Once
)

// Logger interface defines logging methods
type Logger interface {
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Debug(msg string)
}

// simpleLogger implements Logger interface
type simpleLogger struct {
	infoLog  *log.Logger
	warnLog  *log.Logger
	errorLog *log.Logger
	debugLog *log.Logger
}

// GetLogger returns the singleton logger instance
func GetLogger() Logger {
	once.Do(func() {
		defaultLogger = &simpleLogger{
			infoLog:  log.New(os.Stdout, "[INFO] ", log.LstdFlags),
			warnLog:  log.New(os.Stdout, "[WARN] ", log.LstdFlags),
			errorLog: log.New(os.Stderr, "[ERROR] ", log.LstdFlags),
			debugLog: log.New(os.Stdout, "[DEBUG] ", log.LstdFlags),
		}
	})
	return defaultLogger
}

// Info logs an info message
func (sl *simpleLogger) Info(msg string) {
	sl.infoLog.Println(msg)
}

// Warn logs a warning message
func (sl *simpleLogger) Warn(msg string) {
	sl.warnLog.Println(msg)
}

// Error logs an error message
func (sl *simpleLogger) Error(msg string) {
	sl.errorLog.Println(msg)
}

// Debug logs a debug message
func (sl *simpleLogger) Debug(msg string) {
	sl.debugLog.Println(msg)
}
