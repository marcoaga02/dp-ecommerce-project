package logger

import (
	"os"
)

// ParseLogLevel reads the environment variable identified by key and converts its value
// to the corresponding logger.Level.
//
// If the environment variable is not set or contains an unrecognized value, it returns logger.Info as the default.
//
// Parameters:
//   - key: the name of the environment variable to read
//
// Returns:
//   - logger.Level: the corresponding log level constant
func ParseLogLevel(key string) Level {
    level := os.Getenv(key)
    switch level {
    case "debug", "DEBUG":
        return Debug
    case "info", "INFO":
        return Info
    case "warn", "WARN", "warning", "WARNING":
        return Warn
    case "error", "ERROR":
        return Error
    default:
        return Info // default level
    }
}