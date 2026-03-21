// Package logging provides structured logging with file and console output
package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Level represents log severity levels
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel parses a string to a Level
func ParseLevel(s string) Level {
	switch strings.ToLower(s) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

// Logger provides structured logging with file output
type Logger struct {
	mu        sync.Mutex
	level     Level
	console   *log.Logger
	file      *log.Logger
	filePtr   *os.File
	component string
}

// Config holds logger configuration
type Config struct {
	Level     string
	FilePath  string // Path to log file (empty = no file logging)
	Component string // Component name for log prefix
}

// New creates a new logger
func New(cfg Config) (*Logger, error) {
	l := &Logger{
		level:     ParseLevel(cfg.Level),
		console:   log.New(os.Stderr, "", 0),
		component: cfg.Component,
	}

	// Setup file logging if path provided
	if cfg.FilePath != "" {
		// Create log directory if needed
		dir := filepath.Dir(cfg.FilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		// Open log file (append mode)
		f, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}

		l.filePtr = f
		l.file = log.New(f, "", 0)
	}

	return l, nil
}

// Close closes the log file if open
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.filePtr != nil {
		return l.filePtr.Close()
	}
	return nil
}

// WithComponent creates a child logger with a component prefix
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		level:     l.level,
		console:   l.console,
		file:      l.file,
		filePtr:   l.filePtr,
		component: component,
	}
}

// log writes a log entry at the given level
func (l *Logger) log(level Level, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Format timestamp
	now := time.Now()
	timestamp := now.Format("2006-01-02 15:04:05.000")

	// Get caller info for debug level
	var caller string
	if level == LevelDebug {
		_, file, line, ok := runtime.Caller(2)
		if ok {
			caller = fmt.Sprintf(" [%s:%d]", filepath.Base(file), line)
		}
	}

	// Build message
	msg := fmt.Sprintf(format, args...)

	// Build component prefix
	comp := ""
	if l.component != "" {
		comp = fmt.Sprintf("[%s] ", l.component)
	}

	// Format: timestamp LEVEL [component] message [file:line]
	entry := fmt.Sprintf("%s %-5s %s%s%s", timestamp, level.String(), comp, msg, caller)

	// Write to console
	l.console.Println(entry)

	// Write to file if configured
	if l.file != nil {
		l.file.Println(entry)
	}
}

// Debug logs at debug level
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

// Info logs at info level
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

// Warn logs at warn level
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

// Error logs at error level
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

// Printf implements the standard logger interface for compatibility
func (l *Logger) Printf(format string, args ...interface{}) {
	// Parse level from format if present
	msg := fmt.Sprintf(format, args...)
	switch {
	case strings.HasPrefix(msg, "DEBUG:"):
		l.log(LevelDebug, "%s", strings.TrimPrefix(msg, "DEBUG: "))
	case strings.HasPrefix(msg, "INFO:"):
		l.log(LevelInfo, "%s", strings.TrimPrefix(msg, "INFO: "))
	case strings.HasPrefix(msg, "WARN:"):
		l.log(LevelWarn, "%s", strings.TrimPrefix(msg, "WARN: "))
	case strings.HasPrefix(msg, "ERROR:"):
		l.log(LevelError, "%s", strings.TrimPrefix(msg, "ERROR: "))
	default:
		l.log(LevelInfo, "%s", msg)
	}
}

// Println implements the standard logger interface
func (l *Logger) Println(args ...interface{}) {
	l.log(LevelInfo, "%s", fmt.Sprint(args...))
}

// Writer returns an io.Writer that writes to the logger at the given level
func (l *Logger) Writer(level Level) io.Writer {
	return &logWriter{logger: l, level: level}
}

type logWriter struct {
	logger *Logger
	level  Level
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	msg := strings.TrimSpace(string(p))
	if msg != "" {
		w.logger.log(w.level, "%s", msg)
	}
	return len(p), nil
}

// StdLogger returns a standard library *log.Logger that writes to this logger
func (l *Logger) StdLogger() *log.Logger {
	return log.New(l.Writer(LevelInfo), "", 0)
}
