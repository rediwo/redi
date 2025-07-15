package logging

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Field represents a key-value pair for structured logging
type Field struct {
	Key   string
	Value interface{}
}

// Formatter defines the interface for log formatters
type Formatter interface {
	Format(level Level, message string, fields []Field, timestamp time.Time) string
}

// TextFormatter formats log messages as plain text
type TextFormatter struct {
	EnableColors bool
}

// NewTextFormatter creates a new text formatter
func NewTextFormatter(enableColors bool) *TextFormatter {
	return &TextFormatter{
		EnableColors: enableColors,
	}
}

// Format formats a log message as plain text
func (f *TextFormatter) Format(level Level, message string, fields []Field, timestamp time.Time) string {
	var builder strings.Builder
	
	// Timestamp
	builder.WriteString(timestamp.Format("2006-01-02 15:04:05"))
	builder.WriteString(" ")
	
	// Level with optional colors
	levelStr := level.String()
	if f.EnableColors {
		levelStr = f.colorizeLevel(level, levelStr)
	}
	builder.WriteString(fmt.Sprintf("[%s]", levelStr))
	builder.WriteString(" ")
	
	// Message
	builder.WriteString(message)
	
	// Fields
	if len(fields) > 0 {
		builder.WriteString(" ")
		for i, field := range fields {
			if i > 0 {
				builder.WriteString(" ")
			}
			builder.WriteString(fmt.Sprintf("%s=%v", field.Key, field.Value))
		}
	}
	
	return builder.String()
}

// colorizeLevel adds ANSI color codes to the log level
func (f *TextFormatter) colorizeLevel(level Level, levelStr string) string {
	switch level {
	case DEBUG:
		return fmt.Sprintf("\033[36m%s\033[0m", levelStr) // Cyan
	case INFO:
		return fmt.Sprintf("\033[32m%s\033[0m", levelStr) // Green
	case WARN:
		return fmt.Sprintf("\033[33m%s\033[0m", levelStr) // Yellow
	case ERROR:
		return fmt.Sprintf("\033[31m%s\033[0m", levelStr) // Red
	case FATAL:
		return fmt.Sprintf("\033[35m%s\033[0m", levelStr) // Magenta
	default:
		return levelStr
	}
}

// JSONFormatter formats log messages as JSON
type JSONFormatter struct{}

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

// Format formats a log message as JSON
func (f *JSONFormatter) Format(level Level, message string, fields []Field, timestamp time.Time) string {
	logEntry := map[string]interface{}{
		"timestamp": timestamp.Format(time.RFC3339),
		"level":     level.String(),
		"message":   message,
	}
	
	// Add fields to the log entry
	for _, field := range fields {
		logEntry[field.Key] = field.Value
	}
	
	data, err := json.Marshal(logEntry)
	if err != nil {
		// Fallback to plain text if JSON marshaling fails
		return fmt.Sprintf("%s [%s] %s error=failed_to_marshal_json", 
			timestamp.Format("2006-01-02 15:04:05"), level.String(), message)
	}
	
	return string(data)
}