// Package common provides shared logging utilities
package common

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// LogLevel represents logging verbosity
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// String returns the string representation of LogLevel
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// Logger provides structured logging
type loggerImpl struct {
	level  LogLevel
	output io.Writer
	logger *log.Logger
}

// NewLogger creates a new Logger instance
func NewLogger(level LogLevel, output io.Writer) Logger {
	if output == nil {
		output = os.Stdout
	}

	return &loggerImpl{
		level:  level,
		output: output,
		logger: log.New(output, "", 0),
	}
}

// Debug logs a debug message
func (l *loggerImpl) Debug(format string, args ...interface{}) {
	if l.level <= LogLevelDebug {
		l.log(LogLevelDebug, format, args...)
	}
}

// Info logs an info message
func (l *loggerImpl) Info(format string, args ...interface{}) {
	if l.level <= LogLevelInfo {
		l.log(LogLevelInfo, format, args...)
	}
}

// Warn logs a warning message
func (l *loggerImpl) Warn(format string, args ...interface{}) {
	if l.level <= LogLevelWarn {
		l.log(LogLevelWarn, format, args...)
	}
}

// Error logs an error message
func (l *loggerImpl) Error(format string, args ...interface{}) {
	if l.level <= LogLevelError {
		l.log(LogLevelError, format, args...)
	}
}

func (l *loggerImpl) log(level LogLevel, format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	l.logger.Printf("[%s] %s %s", level.String(), timestamp, message)
}

// SetLevel changes the logging level
func (l *loggerImpl) SetLevel(level LogLevel) {
	l.level = level
}
