package runtime

import (
	"os"
	"path/filepath"
	"time"
)

// Config represents JavaScript runtime configuration
type Config struct {
	ScriptPath string
	Args       []string
	Timeout    time.Duration
	Version    string
	BasePath   string
}

// NewConfig creates a new runtime configuration
func NewConfig(scriptPath string) (*Config, error) {
	absPath, err := filepath.Abs(scriptPath)
	if err != nil {
		return nil, RuntimeError{Message: "failed to get absolute path", Err: err}
	}
	
	return &Config{
		ScriptPath: absPath,
		Args:       []string{},
		Timeout:    0, // No timeout by default
		Version:    "dev",
		BasePath:   filepath.Dir(absPath),
	}, nil
}

// WithArgs sets the arguments for the script
func (c *Config) WithArgs(args []string) *Config {
	c.Args = args
	return c
}

// WithTimeout sets the execution timeout
func (c *Config) WithTimeout(timeout time.Duration) *Config {
	c.Timeout = timeout
	return c
}

// WithVersion sets the runtime version
func (c *Config) WithVersion(version string) *Config {
	c.Version = version
	return c
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.ScriptPath == "" {
		return RuntimeError{Message: "script path is required"}
	}
	
	if _, err := os.Stat(c.ScriptPath); os.IsNotExist(err) {
		return RuntimeError{Message: "script file does not exist", Err: err}
	}
	
	return nil
}

// RuntimeError represents a runtime error
type RuntimeError struct {
	Message string
	Err     error
}

func (e RuntimeError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}