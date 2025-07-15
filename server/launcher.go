package server

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/rediwo/redi"
	"github.com/rediwo/redi/logging"
)

// Launcher handles server startup modes
type Launcher struct {
	factory *Factory
}

// NewLauncher creates a new server launcher
func NewLauncher() *Launcher {
	return &Launcher{
		factory: NewFactory(),
	}
}

// Start starts the server based on configuration
func (l *Launcher) Start(config *Config) error {
	// Initialize logging system
	if err := l.initializeLogging(config); err != nil {
		return fmt.Errorf("failed to initialize logging: %v", err)
	}
	
	if config.Daemon {
		return l.startDaemon(config)
	}
	
	if config.LogFile != "" {
		return l.startBackground(config)
	}
	
	return l.startForeground(config)
}

// initializeLogging initializes the global logging system
func (l *Launcher) initializeLogging(config *Config) error {
	logConfig := config.CreateLoggingConfig()
	logger, err := logging.New(logConfig)
	if err != nil {
		return err
	}
	
	logging.SetGlobalLogger(logger)
	return nil
}

// startForeground starts the server in foreground mode
func (l *Launcher) startForeground(config *Config) error {
	server, err := l.factory.CreateServer(config)
	if err != nil {
		return fmt.Errorf("failed to create server: %v", err)
	}
	
	// Run prebuild if requested
	if config.Prebuild {
		logging.Info("Starting pre-build process")
		if err := server.PreBuild(config.PrebuildParallel); err != nil {
			return fmt.Errorf("pre-build failed: %v", err)
		}
		// If only prebuild was requested, exit
		if config.OnlyPrebuild {
			return nil
		}
	}
	
	logging.Info("Starting redi server", "version", config.Version, "port", config.Port, "root", config.Root)
	
	if err := server.Start(); err != nil {
		return fmt.Errorf("server failed to start: %v", err)
	}
	
	return nil
}

// startBackground starts the server in background mode
func (l *Launcher) startBackground(config *Config) error {
	// Create a new process group and detach from terminal
	cmd := exec.Command(os.Args[0])
	
	// Copy all arguments except --log to the new process, add --daemon
	newArgs := []string{}
	skipNext := false
	for _, arg := range os.Args[1:] {
		if skipNext {
			skipNext = false
			continue
		}
		if arg == "--log" {
			skipNext = true
			continue
		}
		if strings.HasPrefix(arg, "--log=") {
			continue
		}
		newArgs = append(newArgs, arg)
	}
	newArgs = append(newArgs, "--daemon")
	cmd.Args = append([]string{os.Args[0]}, newArgs...)

	// Set up the log file for the background process
	logFile, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %v", config.LogFile, err)
	}
	defer logFile.Close()

	// Write startup header to log file
	fmt.Fprintf(logFile, "\n=== Redi Server Starting in Background Mode ===\n")
	fmt.Fprintf(logFile, "Time: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(logFile, "Version: %s\n", config.Version)
	fmt.Fprintf(logFile, "Port: %d\n", config.Port)
	fmt.Fprintf(logFile, "Root: %s\n", config.Root)
	fmt.Fprintf(logFile, "================================================\n")

	cmd.Stdout = logFile
	cmd.Stderr = logFile
	
	// Platform-specific process attributes
	setPlatformSpecificAttributes(cmd)
	
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start background process: %v", err)
	}

	// Print success message to terminal before exiting
	fmt.Printf("Redi server started in background mode\n")
	fmt.Printf("Version: %s\n", config.Version)
	fmt.Printf("Port: %d\n", config.Port)
	fmt.Printf("Root: %s\n", config.Root)
	fmt.Printf("Log file: %s\n", config.LogFile)
	fmt.Printf("PID: %d\n", cmd.Process.Pid)
	fmt.Printf("Server URL: http://localhost:%d\n", config.Port)
	
	// Write PID file
	pidFile := config.LogFile + ".pid"
	pidContent := fmt.Sprintf("%d", cmd.Process.Pid)
	err = os.WriteFile(pidFile, []byte(pidContent), 0644)
	if err != nil {
		fmt.Printf("Warning: Failed to write PID file %s: %v\n", pidFile, err)
	} else {
		fmt.Printf("PID file: %s\n", pidFile)
	}

	return nil
}

// startDaemon starts the actual server daemon
func (l *Launcher) startDaemon(config *Config) error {
	server, err := l.factory.CreateServer(config)
	if err != nil {
		return fmt.Errorf("failed to create server: %v", err)
	}
	
	// Run prebuild if requested
	if config.Prebuild {
		log.Printf("Starting pre-build process...")
		if err := server.PreBuild(config.PrebuildParallel); err != nil {
			return fmt.Errorf("pre-build failed: %v", err)
		}
		// If only prebuild was requested, exit
		if config.OnlyPrebuild {
			return nil
		}
	}
	
	log.Printf("Starting redi server %s on port %d, serving from %s", config.Version, config.Port, config.Root)

	if err := server.Start(); err != nil {
		return fmt.Errorf("server failed to start: %v", err)
	}
	
	return nil
}

// StartSimple starts a server with simple configuration
func (l *Launcher) StartSimple(root string, port int, version string) error {
	config := &Config{
		Root:    root,
		Port:    port,
		Version: version,
	}
	
	return l.startForeground(config)
}

// StartEmbedded starts an embedded server
func (l *Launcher) StartEmbedded(server *redi.Server) error {
	log.Printf("Starting embedded redi server")
	
	if err := server.Start(); err != nil {
		return fmt.Errorf("server failed to start: %v", err)
	}
	
	return nil
}

// Default launcher instance
var DefaultLauncher = NewLauncher()