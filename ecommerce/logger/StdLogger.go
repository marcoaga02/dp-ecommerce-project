package logger

import (
	"fmt"
	"log"
	"os"
)

// StdLogger is a standard logger that writes log messages to stdout,
// including level, component, timestamp, and caller info.
//
// It filters messages by minimum level and formats output with component context.
type StdLogger struct {
	level     Level // the severity of the log
	logger    *log.Logger
	component string
}

// NewStdLogger creates and returns a new StdLogger instance.
//
// Parameters:
//   - level: the minimum Level of messages to log (Debug, Info, Warn, Error)
//   - component: a string to identify the component or module generating logs
//
// Returns:
//   - *StdLogger: a configured logger instance
func NewStdLogger(level Level, component string) *StdLogger {
	return &StdLogger{
		level:     level,
		logger:    log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds),
		component: component,
	}
}

// log writes a formatted log message at the specified level if it meets
// the logger's minimum level.
//
// Parameters:
//   - lvl: the Level of this log message
//   - prefix: string prefix indicating log severity (e.g., "[INFO]")
//   - msg: the message format string (like fmt.Sprintf)
//   - args: optional arguments for the format string
//
// Behavior:
//   - If lvl is below logger's threshold, message is ignored.
//   - Otherwise, message is formatted with component and printed.
func (l *StdLogger) log(lvl Level, prefix, msg string, args ...interface{}) {
	if lvl < l.level {
		return
	}
	fullMsg := fmt.Sprintf("[%s] %s", l.component, fmt.Sprintf(msg, args...))
	l.logger.Printf("%s %s", prefix, fullMsg)
}

// Debug logs a message at Debug level.
//
// Parameters:
//   - msg: format string for the debug message
//   - args: optional arguments for formatting
func (l *StdLogger) Debug(msg string, args ...interface{}) {
	l.log(Debug, "[DEBUG]", msg, args...)
}

// Info logs a message at Info level.
//
// Parameters:
//   - msg: format string for the info message
//   - args: optional arguments for formatting
func (l *StdLogger) Info(msg string, args ...interface{}) {
	l.log(Info, "[INFO]", msg, args...)
}

// Warn logs a message at Warn level.
//
// Parameters:
//   - msg: format string for the warning message
//   - args: optional arguments for formatting
func (l *StdLogger) Warn(msg string, args ...interface{}) {
	l.log(Warn, "[WARN]", msg, args...)
}

// Error logs a message at Error level.
//
// Parameters:
//   - msg: format string for the error message
//   - args: optional arguments for formatting
func (l *StdLogger) Error(msg string, args ...interface{}) {
	l.log(Error, "[ERROR]", msg, args...)
}

// Fatal logs a message at Error level with a "[FATAL]" prefix,
// then immediately terminates the program with exit code 1.
//
// Parameters:
//   - msg: format string for the fatal error message
//   - args: optional arguments for formatting
//
// Behavior:
//   - This function calls os.Exit(1) after logging and does not return.
func (l *StdLogger) Fatal(msg string, args ...interface{}) {
    l.log(Error, "[FATAL]", msg, args...)
    os.Exit(1)
}