package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	js "github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/eventloop"
	"github.com/dop251/goja_nodejs/require"

	"github.com/rediwo/redi/filesystem"
	"github.com/rediwo/redi/handlers"

	// Import all modules to trigger their init() functions
	_ "github.com/rediwo/redi/modules/console"
	_ "github.com/rediwo/redi/modules/fetch"
	_ "github.com/rediwo/redi/modules/fs"
	_ "github.com/rediwo/redi/modules/path"
	_ "github.com/rediwo/redi/modules/process"
)

var (
	// Version will be set by build flags or git tag
	Version = "dev"
)

func main() {
	// Define command line flags
	var (
		showVersion = false
		timeoutMs   = 0
		scriptPath  = ""
		scriptArgs  []string
	)

	// Parse command line arguments manually to handle Node.js-style syntax
	args := os.Args[1:]
	i := 0
	
	for i < len(args) {
		arg := args[i]
		
		if arg == "--version" || arg == "-v" {
			showVersion = true
			i++
		} else if arg == "--timeout" {
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "Error: --timeout requires a value\n")
				os.Exit(1)
			}
			timeout, err := strconv.Atoi(args[i+1])
			if err != nil || timeout < 0 {
				fmt.Fprintf(os.Stderr, "Error: --timeout must be a positive integer (milliseconds)\n")
				os.Exit(1)
			}
			timeoutMs = timeout
			i += 2
		} else if strings.HasPrefix(arg, "--timeout=") {
			timeoutStr := strings.TrimPrefix(arg, "--timeout=")
			timeout, err := strconv.Atoi(timeoutStr)
			if err != nil || timeout < 0 {
				fmt.Fprintf(os.Stderr, "Error: --timeout must be a positive integer (milliseconds)\n")
				os.Exit(1)
			}
			timeoutMs = timeout
			i++
		} else if !strings.HasPrefix(arg, "-") && scriptPath == "" {
			// First non-option argument is the script path
			scriptPath = arg
			scriptArgs = args[i+1:]
			break
		} else {
			fmt.Fprintf(os.Stderr, "Error: Unknown option: %s\n", arg)
			os.Exit(1)
		}
	}

	if showVersion {
		fmt.Printf("rejs version %s\n", getVersion())
		os.Exit(0)
	}

	if scriptPath == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <script.js> [arguments]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nA simplified Node.js-like JavaScript runtime\n")
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		fmt.Fprintf(os.Stderr, "  --version, -v           Show version information\n")
		fmt.Fprintf(os.Stderr, "  --timeout=ms            Set execution timeout in milliseconds\n")
		fmt.Fprintf(os.Stderr, "  --timeout ms            (alternative syntax)\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s hello.js\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --timeout=5000 script.js arg1 arg2\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --timeout 30000 async-script.js\n", os.Args[0])
		os.Exit(1)
	}

	// Check if script file exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Script file '%s' not found\n", scriptPath)
		os.Exit(1)
	}

	// Get absolute path for the script
	absPath, err := filepath.Abs(scriptPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to get absolute path: %v\n", err)
		os.Exit(1)
	}

	// Create filesystem with current directory as root
	fs := filesystem.NewOSFileSystem("/")

	// Run the script with timeout
	timeoutDuration := time.Duration(timeoutMs) * time.Millisecond
	exitCode := runScript(fs, absPath, scriptArgs, timeoutDuration)
	os.Exit(exitCode)
}

func runScript(fs filesystem.FileSystem, scriptPath string, args []string, timeout time.Duration) int {
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
				fmt.Fprintf(os.Stderr, "JavaScript execution error: %v\n", recovered)
				done <- fmt.Errorf("panic: %v", recovered)
			}
		}()

		// Setup module registry using VMManager with custom path resolver
		basePath := filepath.Dir(scriptPath)
		vmManager := handlers.NewVMManager(fs, getVersion())
		
		_, err := vmManager.SetupRegistry(loop, vm, basePath)
		if err != nil {
			done <- fmt.Errorf("failed to setup registry: %v", err)
			return
		}

		// Set up console
		consoleObj := require.Require(vm, "console")
		vm.Set("console", consoleObj)

		// Get the process module and enhance it with rejs-specific properties
		processValue := require.Require(vm, "process")
		if processValue != nil {
			processObj := processValue.ToObject(vm)
			// Update argv with correct values for rejs
			processObj.Set("argv", vm.ToValue(append([]string{"rejs", scriptPath}, args...)))
			// Add exit function
			processObj.Set("exit", func(code int) {
				exitCode = code
				done <- nil
			})
		}
		vm.Set("process", processValue)

		// Instead of running the script directly, require it as a module
		// This ensures proper module context and relative path resolution
		mainModule := require.Require(vm, scriptPath)
		
		// The module should execute when required
		// Check if there was an error in module execution
		if mainModule == nil {
			done <- fmt.Errorf("failed to load main script as module")
			return
		}

		// Script completed successfully, but don't exit immediately
		// Give some time for async operations to complete
		if timeout == 0 {
			// No timeout specified, auto-exit after short delay (for sync scripts)
			go func() {
				time.Sleep(100 * time.Millisecond) // Give async operations a chance
				done <- nil
			}()
		}
		// If timeout is specified, don't auto-exit - let async operations run
	})

	// Wait for completion with optional timeout
	if timeout > 0 {
		// Use timeout
		select {
		case err := <-done:
			if err != nil {
				fmt.Fprintf(os.Stderr, "Script error: %v\n", err)
				return 1
			}
			return exitCode
		case <-time.After(timeout):
			fmt.Fprintf(os.Stderr, "Script timeout: execution exceeded %v\n", timeout)
			return 1
		}
	} else {
		// Wait forever
		if err := <-done; err != nil {
			fmt.Fprintf(os.Stderr, "Script error: %v\n", err)
			return 1
		}
		return exitCode
	}
}


// getVersion returns the version of rejs, trying git tag first, then build-time version
func getVersion() string {
	// If version was set at build time, use it
	if Version != "dev" {
		return Version
	}
	
	// Try to get version from git tag
	if gitVersion := getGitVersion(); gitVersion != "" {
		return gitVersion
	}
	
	// Fallback to dev version
	return Version
}

// getGitVersion attempts to get the current version from git tags
func getGitVersion() string {
	cmd := exec.Command("git", "describe", "--tags", "--exact-match", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		// Try to get the latest tag with commit info
		cmd = exec.Command("git", "describe", "--tags", "--always")
		if output, err = cmd.Output(); err != nil {
			return ""
		}
	}
	
	version := strings.TrimSpace(string(output))
	if version == "" {
		return ""
	}
	
	return version
}