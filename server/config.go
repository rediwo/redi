package server

import (
	"io/fs"
	"os"
)

// Config represents server configuration
type Config struct {
	Root    string
	Port    int
	Version string
	LogFile string
	Daemon  bool
	
	// Gzip compression settings
	EnableGzip bool // Enable gzip compression (default: true)
	GzipLevel  int  // Compression level (-1 to 9, -1 = default)
}

// NewConfig creates a new server configuration
func NewConfig() *Config {
	return &Config{
		Port:       8080,
		Version:    "dev",
		EnableGzip: true,
		GzipLevel:  -1, // Use gzip.DefaultCompression
	}
}

// Validate validates the server configuration
func (c *Config) Validate() error {
	if c.Root == "" {
		return ConfigError{Message: "root directory is required"}
	}
	
	if _, err := os.Stat(c.Root); os.IsNotExist(err) {
		return ConfigError{Message: "root directory does not exist", Err: err}
	}
	
	if c.Port <= 0 || c.Port > 65535 {
		return ConfigError{Message: "port must be between 1 and 65535"}
	}
	
	if c.GzipLevel < -1 || c.GzipLevel > 9 {
		return ConfigError{Message: "gzip level must be between -1 and 9"}
	}
	
	return nil
}

// EmbedConfig represents embedded server configuration
type EmbedConfig struct {
	EmbedFS fs.FS
	Port    int
	Version string
	
	// Gzip compression settings
	EnableGzip bool // Enable gzip compression (default: true)
	GzipLevel  int  // Compression level (-1 to 9, -1 = default)
}

// NewEmbedConfig creates a new embedded server configuration
func NewEmbedConfig(embedFS fs.FS) *EmbedConfig {
	return &EmbedConfig{
		EmbedFS:    embedFS,
		Port:       8080,
		Version:    "dev",
		EnableGzip: true,
		GzipLevel:  -1, // Use gzip.DefaultCompression
	}
}

// Validate validates the embedded server configuration
func (c *EmbedConfig) Validate() error {
	if c.EmbedFS == nil {
		return ConfigError{Message: "embedded filesystem is required"}
	}
	
	if c.Port <= 0 || c.Port > 65535 {
		return ConfigError{Message: "port must be between 1 and 65535"}
	}
	
	if c.GzipLevel < -1 || c.GzipLevel > 9 {
		return ConfigError{Message: "gzip level must be between -1 and 9"}
	}
	
	return nil
}

// ConfigError represents a configuration error
type ConfigError struct {
	Message string
	Err     error
}

func (e ConfigError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}