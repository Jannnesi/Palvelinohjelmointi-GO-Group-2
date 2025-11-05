package logger

import (
	"log"
	"os"
)

// Logger wraps standard logger
type Logger struct {
	*log.Logger
}

// New creates a new logger instance
func New(level string) *Logger {
	return &Logger{
		Logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.Printf("[INFO] %s", msg)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.Printf("[ERROR] %s", msg)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string) {
	l.Printf("[FATAL] %s", msg)
	os.Exit(1)
}
