package server

import (
	"io/fs"
	"os"
	
	"github.com/rediwo/redi/logging"
)

// Config represents server configuration
type Config struct {
	Root      string
	Port      int
	Version   string
	LogFile   string
	Daemon    bool
	RoutesDir string // Directory for routes (default: "routes")
	
	// Gzip compression settings
	EnableGzip bool // Enable gzip compression (default: true)
	GzipLevel  int  // Compression level (-1 to 9, -1 = default)
	
	// Cache settings
	EnableCache bool // Enable compilation cache (default: false)
	
	// Prebuild settings
	Prebuild         bool // Pre-compile all Svelte components before starting
	PrebuildParallel int  // Number of parallel workers for pre-building
	OnlyPrebuild     bool // Only run prebuild without starting server
	
	// Logging settings
	LogLevel    string // Log level (debug, info, warn, error)
	LogFormat   string // Log format (text, json)
	LogQuiet    bool   // Quiet mode (only ERROR and FATAL)
}

// NewConfig creates a new server configuration
func NewConfig() *Config {
	return &Config{
		Port:             8080,
		Version:          "dev",
		RoutesDir:        "routes",
		EnableGzip:       true,
		GzipLevel:        -1, // Use gzip.DefaultCompression
		EnableCache:      false,
		Prebuild:         false,
		PrebuildParallel: 4,
		LogLevel:         "info",
		LogFormat:        "text",
		LogQuiet:         false,
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

// CreateLoggingConfig creates a logging configuration from server config
func (c *Config) CreateLoggingConfig() *logging.Config {
	logConfig := logging.DefaultConfig()
	logConfig.Level = logging.ParseLevel(c.LogLevel)
	logConfig.Format = logging.ParseFormat(c.LogFormat)
	logConfig.Quiet = c.LogQuiet
	
	// Set log file if specified
	if c.LogFile != "" {
		logConfig.File = c.LogFile
		logConfig.EnableColors = false // Disable colors for file output
	}
	
	return logConfig
}

// EmbedConfig represents embedded server configuration
type EmbedConfig struct {
	EmbedFS   fs.FS
	Port      int
	Version   string
	RoutesDir string // Directory for routes (default: "routes")
	
	// Gzip compression settings
	EnableGzip bool // Enable gzip compression (default: true)
	GzipLevel  int  // Compression level (-1 to 9, -1 = default)
	
	// Cache settings
	EnableCache bool // Enable compilation cache (default: false)
}

// NewEmbedConfig creates a new embedded server configuration
func NewEmbedConfig(embedFS fs.FS) *EmbedConfig {
	return &EmbedConfig{
		EmbedFS:     embedFS,
		Port:        8080,
		Version:     "dev",
		RoutesDir:   "routes",
		EnableGzip:  true,
		GzipLevel:   -1, // Use gzip.DefaultCompression
		EnableCache: false,
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