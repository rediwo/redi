package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rediwo/redi/runtime"
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
		versionProvider := runtime.NewVersionProvider(Version)
		fmt.Printf("rejs version %s\n", versionProvider.GetVersion())
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

	// Run the script with timeout
	timeoutDuration := time.Duration(timeoutMs) * time.Millisecond
	versionProvider := runtime.NewVersionProvider(Version)
	
	exitCode, err := runtime.ExecuteScript(scriptPath, scriptArgs, timeoutDuration, versionProvider.GetVersion())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Script error: %v\n", err)
	}
	
	os.Exit(exitCode)
}

