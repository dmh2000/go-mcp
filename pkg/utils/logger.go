package utils

import (
	"fmt"
	"io"
	"log"
	"os"
)

// LogLevel defines the severity level for logging.
type LogLevel int

const (
	// LevelInfo logs informational messages.
	LevelInfo LogLevel = iota
	// LevelDebug logs detailed debugging information.
	LevelDebug
)

// Logger wraps the standard Go logger to provide level-based logging.
type Logger struct {
	stdLogger *log.Logger
	level     LogLevel
}

// New creates a new Logger instance.
// It takes an output writer, prefix string, standard log flags, and the minimum LogLevel to output.
func New(out io.Writer, prefix string, flag int, level LogLevel) *Logger {
	return &Logger{
		stdLogger: log.New(out, prefix, flag),
		level:     level,
	}
}

// SetLevel changes the minimum logging level for the logger.
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// shouldLog checks if a message with the given level should be logged based on the logger's configured level.
func (l *Logger) shouldLog(level LogLevel) bool {
	return level <= l.level
}

// Printf logs a formatted string if the message level is appropriate.
// The first argument is the level of the message (LevelInfo or LevelDebug).
func (l *Logger) Printf(level LogLevel, format string, v ...interface{}) {
	if l.shouldLog(level) {
		l.stdLogger.Printf(format, v...)
	}
}

// Println logs a line if the message level is appropriate.
// The first argument is the level of the message (LevelInfo or LevelDebug).
func (l *Logger) Println(level LogLevel, v ...interface{}) {
	if l.shouldLog(level) {
		l.stdLogger.Println(v...)
	}
}

// Fatalf logs a formatted string and then calls os.Exit(1), regardless of the configured log level.
// The first argument is the level of the message (LevelInfo or LevelDebug), but it's mainly for consistency.
// Fatal messages are always output.
func (l *Logger) Fatalf(level LogLevel, format string, v ...interface{}) {
	// Fatal messages are always logged, regardless of level setting.
	l.stdLogger.Output(2, fmt.Sprintf(format, v...)) // Use Output to control call depth for file/line info
	os.Exit(1)
}

// Fatalln logs a line and then calls os.Exit(1), regardless of the configured log level.
// The first argument is the level of the message (LevelInfo or LevelDebug), but it's mainly for consistency.
// Fatal messages are always output.
func (l *Logger) Fatalln(level LogLevel, v ...interface{}) {
	// Fatal messages are always logged, regardless of level setting.
	l.stdLogger.Output(2, fmt.Sprintln(v...)) // Use Output to control call depth
	os.Exit(1)
}

// StandardLogger returns the underlying standard log.Logger instance.
// This can be useful if direct access to the standard logger is needed.
func (l *Logger) StandardLogger() *log.Logger {
	return l.stdLogger
}
