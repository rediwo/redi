package logging

import (
	"io"
	"os"
)

// Format represents the log output format
type Format int

const (
	// TextFormat outputs logs in plain text format
	TextFormat Format = iota
	// JSONFormat outputs logs in JSON format
	JSONFormat
)

// String returns the string representation of the log format
func (f Format) String() string {
	switch f {
	case TextFormat:
		return "text"
	case JSONFormat:
		return "json"
	default:
		return "text"
	}
}

// ParseFormat parses a string into a log format
func ParseFormat(format string) Format {
	switch format {
	case "json":
		return JSONFormat
	case "text":
		return TextFormat
	default:
		return TextFormat
	}
}

// Config represents the logging configuration
type Config struct {
	// Level is the minimum log level to output
	Level Level
	// Format is the output format (text or json)
	Format Format
	// Output is the writer to output logs to
	Output io.Writer
	// File is the log file path (optional)
	File string
	// EnableColors enables colored output for text format
	EnableColors bool
	// Quiet mode only outputs ERROR and FATAL levels
	Quiet bool
}

// DefaultConfig returns a default logging configuration
func DefaultConfig() *Config {
	return &Config{
		Level:        INFO,
		Format:       TextFormat,
		Output:       os.Stdout,
		EnableColors: true,
		Quiet:        false,
	}
}

// ShouldLog checks if a message with the given level should be logged
func (c *Config) ShouldLog(level Level) bool {
	if c.Quiet {
		return level >= ERROR
	}
	return level >= c.Level
}