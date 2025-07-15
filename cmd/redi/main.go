package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/rediwo/redi/runtime"
	"github.com/rediwo/redi/server"
)

var (
	// Version will be set by build flags or git tag
	Version = "dev"
)

func main() {
	// Check for version flag first
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		versionProvider := runtime.NewVersionProvider(Version)
		fmt.Printf("redi version %s\n", versionProvider.GetVersion())
		return
	}

	var root string
	var port int
	var version bool
	var logFile string
	var daemon bool
	var disableGzip bool
	var enableCache bool
	var clearCache bool
	var prebuild bool
	var prebuildParallel int
	var logLevel string
	var logFormat string
	var logQuiet bool

	flag.StringVar(&root, "root", "", "Root directory containing public and routes folders")
	flag.IntVar(&port, "port", 8080, "Port to serve on")
	flag.BoolVar(&version, "version", false, "Show version information")
	flag.StringVar(&logFile, "log", "", "Log file path (enables background mode like nohup)")
	flag.BoolVar(&daemon, "daemon", false, "Internal flag for daemon mode")
	flag.BoolVar(&disableGzip, "disable-gzip", false, "Disable gzip compression")
	flag.BoolVar(&enableCache, "cache", true, "Enable compilation cache")
	flag.BoolVar(&clearCache, "clear-cache", false, "Clear existing cache and exit")
	flag.BoolVar(&prebuild, "prebuild", false, "Pre-compile all Svelte components before starting server")
	flag.IntVar(&prebuildParallel, "prebuild-parallel", 4, "Number of parallel workers for pre-building (default: 4)")
	flag.StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	flag.StringVar(&logFormat, "log-format", "text", "Log format (text, json)")
	flag.BoolVar(&logQuiet, "quiet", false, "Quiet mode (only ERROR and FATAL messages)")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Redi Frontend Server - Dynamic web serving with JavaScript, Markdown, and templates\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s --root=mysite --port=8080          # Run in foreground\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --root=mysite --log=server.log     # Run in background (like nohup)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --root=mysite --cache              # Enable compilation cache\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --root=mysite --clear-cache        # Clear cache and exit\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --root=mysite --prebuild           # Pre-compile all Svelte components\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --root=mysite --prebuild --port=8080  # Pre-build then start server\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --root=mysite --log-level=debug    # Enable debug logging\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --root=mysite --log-format=json    # Use JSON log format\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --root=mysite --quiet              # Quiet mode (errors only)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --version                          # Show version\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nLogging:\n")
		fmt.Fprintf(os.Stderr, "  Levels: debug, info, warn, error\n")
		fmt.Fprintf(os.Stderr, "  Formats: text (colored), json\n")
		fmt.Fprintf(os.Stderr, "\nWhen using --log, the server runs in background mode and all output\n")
		fmt.Fprintf(os.Stderr, "is redirected to the log file. A PID file (.pid) is also created.\n")
		fmt.Fprintf(os.Stderr, "Use 'kill $(cat logfile.pid)' to stop the background server.\n")
	}

	flag.Parse()

	if version {
		versionProvider := runtime.NewVersionProvider(Version)
		fmt.Printf("redi version %s\n", versionProvider.GetVersion())
		return
	}

	if root == "" {
		fmt.Fprintf(os.Stderr, "Error: --root flag is required\n")
		flag.Usage()
		os.Exit(1)
	}

	if _, err := os.Stat(root); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Root directory %s does not exist\n", root)
		os.Exit(1)
	}

	// Handle cache clearing
	if clearCache {
		cachePath := root + "/.redi"
		if err := os.RemoveAll(cachePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error clearing cache: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Cache cleared successfully: %s\n", cachePath)
		os.Exit(0)
	}

	versionProvider := runtime.NewVersionProvider(Version)
	currentVersion := versionProvider.GetVersion()

	// Check if user explicitly wants only prebuild (by checking if they provided other server flags)
	onlyPrebuild := prebuild && logFile == "" && !daemon

	config := &server.Config{
		Root:        root,
		Port:        port,
		Version:     currentVersion,
		LogFile:     logFile,
		Daemon:      daemon,
		EnableGzip:  !disableGzip,
		GzipLevel:   -1, // Use gzip.DefaultCompression
		EnableCache: enableCache,
		Prebuild:    prebuild,
		PrebuildParallel: prebuildParallel,
		OnlyPrebuild: onlyPrebuild,
		LogLevel:    logLevel,
		LogFormat:   logFormat,
		LogQuiet:    logQuiet,
	}

	launcher := server.NewLauncher()
	if err := launcher.Start(config); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

