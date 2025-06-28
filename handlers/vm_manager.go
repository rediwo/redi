package handlers

import (
	"path/filepath"
	"strings"

	js "github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/eventloop"
	"github.com/dop251/goja_nodejs/require"
	"github.com/rediwo/redi/filesystem"
	"github.com/rediwo/redi/modules"
)

// VMManager manages JavaScript VM creation and module initialization
type VMManager struct {
	fs      filesystem.FileSystem
	version string
}

// NewVMManager creates a new VM manager
func NewVMManager(fs filesystem.FileSystem, version string) *VMManager {
	return &VMManager{
		fs:      fs,
		version: version,
	}
}

// SetupRegistry creates and configures a require registry with all modules
func (vm *VMManager) SetupRegistry(loop *eventloop.EventLoop, jsVM *js.Runtime, basePath string) (*require.Registry, error) {
	registry := require.NewRegistry(
		require.WithLoader(vm.createModuleLoader(basePath)),
	)

	config := modules.ModuleConfig{
		Registry:   registry,
		EventLoop:  loop,
		FileSystem: vm.fs,
		BasePath:   basePath,
		Version:    vm.version,
		VM:         jsVM,
	}

	// Initialize all auto-registered modules first
	err := modules.InitializeAllModules(config)
	if err != nil {
		return nil, err
	}

	// Enable the registry in the VM to make modules available
	registry.Enable(jsVM)

	return registry, nil
}

// createModuleLoader creates a module loader function for the require system
func (vm *VMManager) createModuleLoader(basePath string) func(string) ([]byte, error) {
	return func(name string) ([]byte, error) {
		// Resolve relative paths from the current route's directory
		var filePath string
		if strings.HasPrefix(name, "./") || strings.HasPrefix(name, "../") {
			// Relative path - join with current directory
			filePath = filepath.Join(basePath, name)
		} else {
			// Module name without path (try in current directory)
			filePath = filepath.Join(basePath, name)
		}

		// Add .js extension if not present
		if !strings.HasSuffix(filePath, ".js") && !strings.HasSuffix(filePath, ".json") {
			if _, err := vm.fs.Stat(filePath + ".js"); err == nil {
				filePath += ".js"
			} else if _, err := vm.fs.Stat(filePath + ".json"); err == nil {
				filePath += ".json"
			}
		}

		// Security check: ensure file is within the route directory
		if !strings.HasPrefix(filePath, basePath) {
			return nil, require.ModuleFileDoesNotExistError
		}

		// Read the file using unified filesystem interface
		return vm.fs.ReadFile(filePath)
	}
}