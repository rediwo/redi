package logging

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Logger defines the interface for logging
type Logger interface {
	Debug(message string, fields ...interface{})
	Info(message string, fields ...interface{})
	Warn(message string, fields ...interface{})
	Error(message string, fields ...interface{})
	Fatal(message string, fields ...interface{})
	
	IsDebugEnabled() bool
	IsInfoEnabled() bool
	IsWarnEnabled() bool
	IsErrorEnabled() bool
	
	WithFields(fields ...interface{}) Logger
	SetLevel(level Level)
	SetOutput(output io.Writer)
	Close() error
}

// logger implements the Logger interface
type logger struct {
	config    *Config
	formatter Formatter
	output    io.Writer
	file      *os.File
	mu        sync.Mutex
	fields    []Field
}

// New creates a new logger with the given configuration
func New(config *Config) (Logger, error) {
	if config == nil {
		config = DefaultConfig()
	}
	
	l := &logger{
		config: config,
		output: config.Output,
		fields: make([]Field, 0),
	}
	
	// Set formatter based on format
	switch config.Format {
	case JSONFormat:
		l.formatter = NewJSONFormatter()
	default:
		l.formatter = NewTextFormatter(config.EnableColors)
	}
	
	// Open log file if specified
	if config.File != "" {
		file, err := os.OpenFile(config.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file %s: %w", config.File, err)
		}
		l.file = file
		l.output = file
	}
	
	return l, nil
}

// log writes a log message with the given level
func (l *logger) log(level Level, message string, fields ...interface{}) {
	if !l.config.ShouldLog(level) {
		return
	}
	
	l.mu.Lock()
	defer l.mu.Unlock()
	
	// Combine instance fields with provided fields
	allFields := make([]Field, len(l.fields))
	copy(allFields, l.fields)
	
	// Parse additional fields
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key := fmt.Sprintf("%v", fields[i])
			value := fields[i+1]
			allFields = append(allFields, Field{Key: key, Value: value})
		}
	}
	
	// Format and write the log message
	timestamp := time.Now()
	formatted := l.formatter.Format(level, message, allFields, timestamp)
	
	// Write to output
	if l.output != nil {
		fmt.Fprintln(l.output, formatted)
	}
	
	// For fatal errors, exit the program
	if level == FATAL {
		os.Exit(1)
	}
}

// Debug logs a debug message
func (l *logger) Debug(message string, fields ...interface{}) {
	l.log(DEBUG, message, fields...)
}

// Info logs an info message
func (l *logger) Info(message string, fields ...interface{}) {
	l.log(INFO, message, fields...)
}

// Warn logs a warning message
func (l *logger) Warn(message string, fields ...interface{}) {
	l.log(WARN, message, fields...)
}

// Error logs an error message
func (l *logger) Error(message string, fields ...interface{}) {
	l.log(ERROR, message, fields...)
}

// Fatal logs a fatal message and exits the program
func (l *logger) Fatal(message string, fields ...interface{}) {
	l.log(FATAL, message, fields...)
}

// IsDebugEnabled checks if debug logging is enabled
func (l *logger) IsDebugEnabled() bool {
	return l.config.ShouldLog(DEBUG)
}

// IsInfoEnabled checks if info logging is enabled
func (l *logger) IsInfoEnabled() bool {
	return l.config.ShouldLog(INFO)
}

// IsWarnEnabled checks if warn logging is enabled
func (l *logger) IsWarnEnabled() bool {
	return l.config.ShouldLog(WARN)
}

// IsErrorEnabled checks if error logging is enabled
func (l *logger) IsErrorEnabled() bool {
	return l.config.ShouldLog(ERROR)
}

// WithFields creates a new logger with additional fields
func (l *logger) WithFields(fields ...interface{}) Logger {
	newLogger := &logger{
		config:    l.config,
		formatter: l.formatter,
		output:    l.output,
		file:      l.file,
		fields:    make([]Field, len(l.fields)),
	}
	copy(newLogger.fields, l.fields)
	
	// Add new fields
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key := fmt.Sprintf("%v", fields[i])
			value := fields[i+1]
			newLogger.fields = append(newLogger.fields, Field{Key: key, Value: value})
		}
	}
	
	return newLogger
}

// SetLevel sets the minimum log level
func (l *logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.config.Level = level
}

// SetOutput sets the output writer
func (l *logger) SetOutput(output io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = output
}

// Close closes the logger and any associated resources
func (l *logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// Global logger instance
var globalLogger Logger

// init initializes the global logger
func init() {
	var err error
	globalLogger, err = New(DefaultConfig())
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize global logger: %v", err))
	}
}

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger Logger) {
	globalLogger = logger
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() Logger {
	return globalLogger
}

// Global logging functions that use the global logger
func Debug(message string, fields ...interface{}) {
	globalLogger.Debug(message, fields...)
}

func Info(message string, fields ...interface{}) {
	globalLogger.Info(message, fields...)
}

func Warn(message string, fields ...interface{}) {
	globalLogger.Warn(message, fields...)
}

func Error(message string, fields ...interface{}) {
	globalLogger.Error(message, fields...)
}

func Fatal(message string, fields ...interface{}) {
	globalLogger.Fatal(message, fields...)
}

func IsDebugEnabled() bool {
	return globalLogger.IsDebugEnabled()
}

func WithFields(fields ...interface{}) Logger {
	return globalLogger.WithFields(fields...)
}