package registry

import (
	"fmt"

	js "github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/eventloop"
	"github.com/dop251/goja_nodejs/require"
	"github.com/rediwo/redi/filesystem"
)

// ModuleInitializer is a function that initializes a module with the given configuration
type ModuleInitializer func(config ModuleConfig) error

// ModuleConfig contains all the configuration needed to initialize a module
type ModuleConfig struct {
	Registry   *require.Registry
	EventLoop  *eventloop.EventLoop
	FileSystem filesystem.FileSystem
	BasePath   string
	Version    string
	VM         *js.Runtime
}

// registeredModules holds all registered module initializers
var registeredModules = make(map[string]ModuleInitializer)

// RegisterModule registers a module initializer with the given name
func RegisterModule(name string, initializer ModuleInitializer) {
	registeredModules[name] = initializer
}

// InitializeAllModules initializes all registered modules with the given configuration
func InitializeAllModules(config ModuleConfig) error {
	for name, init := range registeredModules {
		if err := init(config); err != nil {
			return fmt.Errorf("failed to initialize module %s: %v", name, err)
		}
	}
	return nil
}

// GetRegisteredModules returns a copy of all registered module names
func GetRegisteredModules() []string {
	modules := make([]string, 0, len(registeredModules))
	for name := range registeredModules {
		modules = append(modules, name)
	}
	return modules
}