package runtime

import (
	"fmt"
	"time"

	js "github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/eventloop"

	"github.com/rediwo/redi/filesystem"
	"github.com/rediwo/redi/handlers"
	_ "github.com/rediwo/redi/modules"
)

// Executor executes JavaScript code
type Executor struct {
	fs filesystem.FileSystem
}

// NewExecutor creates a new JavaScript executor
func NewExecutor() *Executor {
	return &Executor{
		fs: filesystem.NewOSFileSystem("/"),
	}
}

// NewExecutorWithFS creates a new JavaScript executor with custom filesystem
func NewExecutorWithFS(fs filesystem.FileSystem) *Executor {
	return &Executor{
		fs: fs,
	}
}

// Execute runs a JavaScript script with the given configuration
func (e *Executor) Execute(config *Config) (int, error) {
	if err := config.Validate(); err != nil {
		return 1, err
	}

	return e.runScript(config)
}

// runScript executes the JavaScript script
func (e *Executor) runScript(config *Config) (int, error) {
	// Create event loop
	loop := eventloop.NewEventLoop()
	go loop.Start()
	defer loop.Stop()

	// Channel to wait for script completion
	done := make(chan error, 1)
	exitCode := 0

	// Run in event loop
	loop.RunOnLoop(func(vm *js.Runtime) {
		defer func() {
			if recovered := recover(); recovered != nil {
				done <- RuntimeError{Message: fmt.Sprintf("JavaScript execution error: %v", recovered)}
			}
		}()

		// Setup module registry using VMManager with custom path resolver
		vmManager := handlers.NewVMManager(e.fs, config.Version)

		_, requireModule, err := vmManager.SetupRegistry(loop, vm, config.BasePath)
		if err != nil {
			done <- RuntimeError{Message: "failed to setup registry", Err: err}
			return
		}

		// Set up console
		consoleObj, err := requireModule.Require("console")
		if err != nil {
			done <- RuntimeError{Message: "failed to require console", Err: err}
			return
		}
		vm.Set("console", consoleObj)

		// Get the process module and enhance it with rejs-specific properties
		processValue, err := requireModule.Require("process")
		if err != nil {
			done <- RuntimeError{Message: "failed to require process", Err: err}
			return
		}
		if processValue != nil {
			processObj := processValue.ToObject(vm)
			// Update argv with correct values for rejs
			processObj.Set("argv", vm.ToValue(append([]string{"rejs", config.ScriptPath}, config.Args...)))
			// Add exit function
			processObj.Set("exit", func(code int) {
				exitCode = code
				done <- nil
			})
		}
		vm.Set("process", processValue)

		// Instead of running the script directly, require it as a module
		// This ensures proper module context and relative path resolution
		mainModule, err := requireModule.Require(config.ScriptPath)
		if err != nil {
			done <- RuntimeError{Message: "failed to load main script as module", Err: err}
			return
		}

		// The module should execute when required
		// Check if there was an error in module execution
		if mainModule == nil {
			done <- RuntimeError{Message: "failed to load main script as module"}
			return
		}

		// Script completed successfully, but don't exit immediately
		// Give some time for async operations to complete
		if config.Timeout == 0 {
			// No timeout specified, auto-exit after short delay (for sync scripts)
			go func() {
				time.Sleep(100 * time.Millisecond) // Give async operations a chance
				done <- nil
			}()
		}
		// If timeout is specified, don't auto-exit - let async operations run
	})

	// Wait for completion with optional timeout
	if config.Timeout > 0 {
		// Use timeout
		select {
		case err := <-done:
			if err != nil {
				return 1, err
			}
			return exitCode, nil
		case <-time.After(config.Timeout):
			return 1, RuntimeError{Message: fmt.Sprintf("script timeout: execution exceeded %v", config.Timeout)}
		}
	} else {
		// Wait forever
		if err := <-done; err != nil {
			return 1, err
		}
		return exitCode, nil
	}
}

// ExecuteScript is a convenience function for simple script execution
func ExecuteScript(scriptPath string, args []string, timeout time.Duration, version string) (int, error) {
	config, err := NewConfig(scriptPath)
	if err != nil {
		return 1, err
	}

	config.WithArgs(args).WithTimeout(timeout).WithVersion(version)

	executor := NewExecutor()
	return executor.Execute(config)
}

// ExecuteScriptSimple is a convenience function for very simple script execution
func ExecuteScriptSimple(scriptPath string) (int, error) {
	return ExecuteScript(scriptPath, []string{}, 0, "dev")
}
