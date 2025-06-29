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
		require.WithPathResolver(vm.createPathResolver(basePath)), // Custom path resolver for absolute paths
		require.WithGlobalFolders(), // Enable global folders resolution
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
		// The name parameter is already resolved by the PathResolver
		// If it's an absolute path, use it directly
		var filePath string
		if filepath.IsAbs(name) {
			filePath = name
		} else {
			// For relative paths, resolve from basePath
			if strings.HasPrefix(name, "./") || strings.HasPrefix(name, "../") {
				filePath = filepath.Join(basePath, name)
			} else {
				filePath = filepath.Join(basePath, name)
			}
		}

		// Check if this is a directory and look for index.js
		if info, err := vm.fs.Stat(filePath); err == nil && info.IsDir() {
			// Try index.js in the directory
			indexPath := filepath.Join(filePath, "index.js")
			if _, err := vm.fs.Stat(indexPath); err == nil {
				filePath = indexPath
			} else {
				// Try index.json
				indexPath = filepath.Join(filePath, "index.json")
				if _, err := vm.fs.Stat(indexPath); err == nil {
					filePath = indexPath
				} else {
					return nil, require.ModuleFileDoesNotExistError
				}
			}
		} else {
			// Add .js extension if not present and file doesn't exist as-is
			if !strings.HasSuffix(filePath, ".js") && !strings.HasSuffix(filePath, ".json") {
				if _, err := vm.fs.Stat(filePath); err != nil {
					// File doesn't exist as-is, try with extensions
					if _, err := vm.fs.Stat(filePath + ".js"); err == nil {
						filePath += ".js"
					} else if _, err := vm.fs.Stat(filePath + ".json"); err == nil {
						filePath += ".json"
					} else {
						return nil, require.ModuleFileDoesNotExistError
					}
				}
			}
		}

		// Read the file using unified filesystem interface
		return vm.fs.ReadFile(filePath)
	}
}

// createPathResolver creates a path resolver function that returns absolute paths
// This ensures __dirname and __filename in required modules have absolute paths
func (vm *VMManager) createPathResolver(basePath string) func(string, string) string {
	return func(base, name string) string {
		// If name is already absolute, return it as-is
		if filepath.IsAbs(name) {
			return filepath.Clean(name)
		}

		// Determine the base directory for resolution
		var baseDir string
		if base == "" {
			// First require - use basePath
			baseDir = basePath
		} else {
			// base could be either a file path or directory path
			// If it's already absolute, use it as-is if it's a directory
			// or get its directory if it's a file
			if filepath.IsAbs(base) {
				// Check if base is a directory or file
				if info, err := vm.fs.Stat(base); err == nil && info.IsDir() {
					baseDir = base
				} else {
					// It's a file path, get its directory
					baseDir = filepath.Dir(base)
				}
			} else {
				// If base is relative, resolve it relative to basePath
				absBase, err := filepath.Abs(filepath.Join(basePath, base))
				if err == nil {
					if info, err := vm.fs.Stat(absBase); err == nil && info.IsDir() {
						baseDir = absBase
					} else {
						baseDir = filepath.Dir(absBase)
					}
				} else {
					baseDir = filepath.Dir(filepath.Join(basePath, base))
				}
			}
		}

		// Resolve the module path relative to baseDir
		resolvedPath := filepath.Join(baseDir, name)

		// Convert to absolute path and clean it
		if absPath, err := filepath.Abs(resolvedPath); err == nil {
			return filepath.Clean(absPath)
		}

		// Fallback to cleaned resolved path
		return filepath.Clean(resolvedPath)
	}
}