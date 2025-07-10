package builder

import "os"

// Config represents build configuration
type Config struct {
	Root        string
	Output      string
	AppName     string
	Platform    string
	Extensions  []string
	ConfigFile  string
	ScriptPath  string   // For CLI builder - main JavaScript file
}

// Builder interface for different build types
type Builder interface {
	Build(config Config) error
	Validate(config Config) error
}

// BuildError represents a build error
type BuildError struct {
	Message string
	Err     error
}

func (e BuildError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// NewBuildError creates a new build error
func NewBuildError(message string, err error) BuildError {
	return BuildError{Message: message, Err: err}
}

// ValidateRoot checks if the root directory exists
func ValidateRoot(root string) error {
	if root == "" {
		return NewBuildError("root directory is required", nil)
	}
	
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return NewBuildError("root directory does not exist", err)
	}
	
	return nil
}