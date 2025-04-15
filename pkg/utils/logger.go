package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings" // Added for ToUpper
)

// Define valid log level strings
const (
	LevelInfo  = "INFO"
	LevelDebug = "DEBUG"
)

// Logger wraps the standard Go logger to provide level-based logging.
type Logger struct {
	stdLogger *log.Logger
	level     string // Store level as a string ("INFO" or "DEBUG")
}

// New creates a new Logger instance.
// It takes an output writer, prefix string, standard log flags, and the minimum level string ("INFO" or "DEBUG") to output.
// Defaults to "INFO" if an invalid level string is provided.
func New(out io.Writer, prefix string, flag int, level string) *Logger {
	normalizedLevel := strings.ToUpper(level)
	if normalizedLevel != LevelDebug {
		normalizedLevel = LevelInfo // Default to INFO
	}
	return &Logger{
		stdLogger: log.New(out, prefix, flag),
		level:     normalizedLevel,
	}
}

// SetLevel changes the minimum logging level for the logger using a string ("INFO" or "DEBUG").
// Defaults to "INFO" if an invalid level string is provided.
func (l *Logger) SetLevel(level string) {
	l.level = strings.ToUpper(level)
}

// shouldLog checks if a message with the given level string should be logged.
func (l *Logger) shouldLog(messageLevel string) bool {
	// print only if it matches, otherwise nothing
	return messageLevel == l.level
}

// Printf logs a formatted string if the message level is appropriate.
// The first argument is the level string ("INFO" or "DEBUG").
func (l *Logger) Printf(level string, format string, v ...interface{}) {
	if l.shouldLog(level) {
		// Call Output with depth 3 to capture the caller's file/line
		l.stdLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

// Println logs a line if the message level is appropriate.
// The first argument is the level string ("INFO" or "DEBUG").
func (l *Logger) Println(level string, v ...interface{}) {
	if l.shouldLog(level) {
		// Call Output with depth 3 to capture the caller's file/line
		l.stdLogger.Output(2, fmt.Sprintln(v...))
	}
}

// Fatalf logs a formatted string and then calls os.Exit(1), regardless of the configured log level.
// The first argument is the level string ("INFO" or "DEBUG"), but it's mainly for consistency.
// Fatal messages are always output.
func (l *Logger) Fatalf(level string, format string, v ...interface{}) {
	// Fatal messages are always logged, regardless of level setting.
	l.stdLogger.Output(2, fmt.Sprintf(format, v...)) // Use Output with depth 3 to capture the caller's file/line
	os.Exit(1)
}

// Fatalln logs a line and then calls os.Exit(1), regardless of the configured log level.
// The first argument is the level string ("INFO" or "DEBUG"), but it's mainly for consistency.
// Fatal messages are always output.
func (l *Logger) Fatalln(level string, v ...interface{}) {
	// Fatal messages are always logged, regardless of level setting.
	l.stdLogger.Output(2, fmt.Sprintln(v...)) // Use Output with depth 3 to capture the caller's file/line
	os.Exit(1)
}

// StandardLogger returns the underlying standard log.Logger instance.
// This can be useful if direct access to the standard logger is needed.
func (l *Logger) StandardLogger() *log.Logger {
	return l.stdLogger
}
