package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	
	"github.com/rediwo/redi"
)

var (
	// Version will be set by build flags or git tag
	Version = "dev"
)

func main() {
	// Check for version flag first
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("redi version %s\n", getVersion())
		return
	}
	
	// Check if this is a build command
	if len(os.Args) > 1 && os.Args[1] == "build" {
		buildCommand()
		return
	}
	
	// Check if this is a build-app command
	if len(os.Args) > 1 && os.Args[1] == "build-app" {
		buildAppCommand()
		return
	}

	var root string
	var port int
	var version bool

	flag.StringVar(&root, "root", "", "Root directory containing public and routes folders")
	flag.IntVar(&port, "port", 8080, "Port to serve on")
	flag.BoolVar(&version, "version", false, "Show version information")
	flag.Parse()

	if version {
		fmt.Printf("redi version %s\n", getVersion())
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

	currentVersion := getVersion()
	server := redi.NewServerWithVersion(root, port, currentVersion)
	log.Printf("Starting redi server %s on port %d, serving from %s", currentVersion, port, root)

	if err := server.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func buildCommand() {
	buildFlags := flag.NewFlagSet("build", flag.ExitOnError)
	
	var root string
	var output string
	
	buildFlags.StringVar(&root, "root", "", "Root directory to embed")
	buildFlags.StringVar(&output, "output", "redi-embedded", "Output executable name")
	
	buildFlags.Parse(os.Args[2:])
	
	if root == "" {
		fmt.Fprintf(os.Stderr, "Error: --root flag is required for build command\n")
		buildFlags.Usage()
		os.Exit(1)
	}
	
	if _, err := os.Stat(root); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Root directory %s does not exist\n", root)
		os.Exit(1)
	}
	
	log.Printf("Building embedded version with root directory: %s", root)
	log.Printf("Output executable: %s", output)
	
	// Create temporary directory for build files
	tempDir, err := os.MkdirTemp("", "redi-build-*")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Copy root directory to temp directory
	targetDir := filepath.Join(tempDir, filepath.Base(root))
	if err := copyDir(root, targetDir); err != nil {
		log.Fatalf("Failed to copy root directory: %v", err)
	}
	
	// Generate the embedded main.go
	if err := generateEmbeddedMain(tempDir, filepath.Base(root)); err != nil {
		log.Fatalf("Failed to generate embedded main: %v", err)
	}
	
	// Create a new go.mod for the embedded app
	if err := createEmbeddedGoMod(tempDir); err != nil {
		log.Fatalf("Failed to create embedded go.mod: %v", err)
	}
	
	if err := copyFile("go.sum", filepath.Join(tempDir, "go.sum")); err != nil {
		log.Fatalf("Failed to copy go.sum: %v", err)
	}
	
	// Run go mod tidy
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = tempDir
	if err := tidyCmd.Run(); err != nil {
		log.Fatalf("Failed to run go mod tidy: %v", err)
	}
	
	// Build the embedded executable
	cmd := exec.Command("go", "build", "-o", output, ".")
	cmd.Dir = tempDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to build embedded executable: %v", err)
	}
	
	// Move the executable to current directory
	builtPath := filepath.Join(tempDir, output)
	finalPath := filepath.Join(".", output)
	
	if err := os.Rename(builtPath, finalPath); err != nil {
		log.Fatalf("Failed to move executable: %v", err)
	}
	
	log.Printf("Successfully built embedded executable: %s", finalPath)
}

func generateEmbeddedMain(tempDir, rootDir string) error {
	// Template for embedded main.go
	const mainTemplate = `package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	
	"github.com/rediwo/redi"
)

//go:embed all:{{.RootDir}}
var embeddedFS embed.FS

func main() {
	var port int
	flag.IntVar(&port, "port", 8080, "Port to serve on")
	flag.Parse()

	// Create a sub-filesystem starting from the root directory
	rootFS, err := fs.Sub(embeddedFS, "{{.RootDir}}")
	if err != nil {
		log.Fatalf("Failed to create sub-filesystem: %v", err)
	}

	server := redi.NewServerWithFS(rootFS, port)
	log.Printf("Starting embedded redi server on port %d", port)

	if err := server.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
`

	tmpl, err := template.New("main").Parse(mainTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}
	
	// Clean the root directory path for embed directive
	cleanRoot := strings.ReplaceAll(rootDir, "\\", "/")
	cleanRoot = strings.TrimPrefix(cleanRoot, "./")
	
	data := struct {
		RootDir string
	}{
		RootDir: cleanRoot,
	}
	
	mainFile := filepath.Join(tempDir, "main.go")
	f, err := os.Create(mainFile)
	if err != nil {
		return fmt.Errorf("failed to create main.go: %v", err)
	}
	defer f.Close()
	
	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}
	
	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	
	_, err = dstFile.ReadFrom(srcFile)
	return err
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Create relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		return copyFile(path, dstPath)
	})
}

func createEmbeddedGoMod(tempDir string) error {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	
	// Check if we're running from source (has go.mod with github.com/rediwo/redi module)
	// and the current directory contains the actual redi source code
	isSourceInstall := false
	if goModData, err := os.ReadFile("go.mod"); err == nil {
		goModContent := string(goModData)
		if strings.Contains(goModContent, "module github.com/rediwo/redi") {
			// Double check by looking for main redi files
			if _, err := os.Stat("server.go"); err == nil {
				if _, err := os.Stat("router.go"); err == nil {
					isSourceInstall = true
				}
			}
		}
	}
	
	var goModContent string
	if isSourceInstall {
		// Running from source - use replace directive
		goModContent = fmt.Sprintf(`module redi-embedded

go 1.23

require github.com/rediwo/redi v0.0.0

replace github.com/rediwo/redi => %s
`, cwd)
	} else {
		// Running from installed binary - use published version
		goModContent = `module redi-embedded

go 1.23

require github.com/rediwo/redi v1.0.0
`
	}
	
	goModPath := filepath.Join(tempDir, "go.mod")
	return os.WriteFile(goModPath, []byte(goModContent), 0644)
}

func buildAppCommand() {
	buildFlags := flag.NewFlagSet("build-app", flag.ExitOnError)
	
	var root string
	var output string
	var appName string
	
	buildFlags.StringVar(&root, "root", "", "Root directory to embed")
	buildFlags.StringVar(&output, "output", "redi-app", "Output directory name")
	buildFlags.StringVar(&appName, "name", "Redi App", "Application name")
	
	buildFlags.Parse(os.Args[2:])
	
	if root == "" {
		fmt.Fprintf(os.Stderr, "Error: --root flag is required for build-app command\n")
		buildFlags.Usage()
		os.Exit(1)
	}
	
	if _, err := os.Stat(root); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Root directory %s does not exist\n", root)
		os.Exit(1)
	}
	
	log.Printf("Building Wails desktop application with root directory: %s", root)
	log.Printf("Application name: %s", appName)
	log.Printf("Output directory: %s", output)
	
	// Check if wails is installed
	if err := checkWailsInstalled(); err != nil {
		log.Fatalf("Wails is not installed: %v", err)
	}
	
	// Create app directory
	if err := os.MkdirAll(output, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}
	
	// Create a valid module name (no spaces, lowercase)
	moduleName := strings.ToLower(strings.ReplaceAll(appName, " ", "-"))
	
	// Initialize Wails project in the output directory
	log.Printf("Initializing Wails project...")
	initCmd := exec.Command("wails", "init", "-n", moduleName, "-t", "plain")
	initCmd.Dir = output
	initCmd.Stdout = os.Stdout
	initCmd.Stderr = os.Stderr
	
	if err := initCmd.Run(); err != nil {
		log.Fatalf("Failed to initialize Wails project: %v", err)
	}
	
	// Copy root directory to app/embed
	appPath := filepath.Join(output, moduleName)
	embedPath := filepath.Join(appPath, "embed")
	
	if err := os.MkdirAll(embedPath, 0755); err != nil {
		log.Fatalf("Failed to create embed directory: %v", err)
	}
	
	targetDir := filepath.Join(embedPath, filepath.Base(root))
	if err := copyDir(root, targetDir); err != nil {
		log.Fatalf("Failed to copy root directory: %v", err)
	}
	
	// Generate Wails app.go
	if err := generateWailsApp(appPath, filepath.Base(root), appName); err != nil {
		log.Fatalf("Failed to generate Wails app: %v", err)
	}
	
	// Update main.go
	if err := generateWailsMain(appPath, appName); err != nil {
		log.Fatalf("Failed to generate Wails main: %v", err)
	}
	
	// Update go.mod
	if err := updateWailsGoMod(appPath); err != nil {
		log.Fatalf("Failed to update go.mod: %v", err)
	}
	
	// Copy go.sum
	if err := copyFile("go.sum", filepath.Join(appPath, "go.sum")); err != nil {
		log.Printf("Warning: Failed to copy go.sum: %v", err)
	}
	
	// Run go mod tidy
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = appPath
	if err := tidyCmd.Run(); err != nil {
		log.Fatalf("Failed to run go mod tidy: %v", err)
	}
	
	// Create a minimal frontend dist directory
	distPath := filepath.Join(appPath, "frontend", "dist")
	if err := os.MkdirAll(distPath, 0755); err != nil {
		log.Fatalf("Failed to create frontend/dist: %v", err)
	}
	
	// Copy frontend src to dist
	srcPath := filepath.Join(appPath, "frontend", "src")
	if err := copyDir(srcPath, distPath); err != nil {
		log.Printf("Warning: Failed to copy frontend src to dist: %v", err)
	}
	
	log.Printf("Successfully created Wails application in: %s", appPath)
	log.Printf("To run in development mode: cd %s && wails dev", appPath)
	log.Printf("To build the app: cd %s && wails build", appPath)
}

func checkWailsInstalled() error {
	cmd := exec.Command("wails", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("wails command not found. Please install Wails first: go install github.com/wailsapp/wails/v2/cmd/wails@latest")
	}
	return nil
}

func generateWailsApp(appPath, rootDir, appName string) error {
	const appTemplate = `package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"sync"
	"time"
	
	"github.com/rediwo/redi"
)

//go:embed all:embed/{{.RootDir}}
var embeddedFS embed.FS

// App struct
type App struct {
	ctx    context.Context
	server *redi.Server
	mu     sync.Mutex
	port   int
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		port: 8080,
	}
}

// OnStartup is called when the app starts up
func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
	a.startServer()
}

// OnShutdown is called when the app is shutting down
func (a *App) OnShutdown(ctx context.Context) {
	a.stopServer()
}

// StartServer starts the embedded server
func (a *App) StartServer() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if a.server != nil {
		return fmt.Sprintf("Server already running on port %d", a.port)
	}
	
	a.startServer()
	return fmt.Sprintf("Server started on port %d", a.port)
}

// StopServer stops the embedded server
func (a *App) StopServer() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if a.server == nil {
		return "Server not running"
	}
	
	a.stopServer()
	return "Server stopped"
}

// GetServerStatus returns the current server status
func (a *App) GetServerStatus() map[string]interface{} {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	status := map[string]interface{}{
		"running": a.server != nil,
		"port":    a.port,
		"url":     fmt.Sprintf("http://localhost:%d", a.port),
	}
	
	return status
}

// OpenServer opens the server URL in the default browser
func (a *App) OpenServer() string {
	url := fmt.Sprintf("http://localhost:%d", a.port)
	return url
}

func (a *App) startServer() {
	// Create a sub-filesystem starting from the root directory
	rootFS, err := fs.Sub(embeddedFS, "embed/{{.RootDir}}")
	if err != nil {
		log.Printf("Failed to create sub-filesystem: %v", err)
		return
	}

	// Find available port
	for port := a.port; port < a.port+100; port++ {
		if isPortAvailable(port) {
			a.port = port
			break
		}
	}

	a.server = redi.NewServerWithFS(rootFS, a.port)
	
	go func() {
		log.Printf("Starting embedded redi server on port %d", a.port)
		if err := a.server.Start(); err != nil {
			log.Printf("Server failed to start: %v", err)
			a.mu.Lock()
			a.server = nil
			a.mu.Unlock()
		}
	}()
	
	// Wait a moment for server to start
	time.Sleep(100 * time.Millisecond)
}

func (a *App) stopServer() {
	if a.server != nil {
		a.server.Stop()
		a.server = nil
	}
}

func isPortAvailable(port int) bool {
	conn, err := http.Get(fmt.Sprintf("http://localhost:%d", port))
	if err != nil {
		return true // Port is available
	}
	conn.Body.Close()
	return false // Port is busy
}
`

	tmpl, err := template.New("app").Parse(appTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse app template: %v", err)
	}
	
	cleanRoot := strings.ReplaceAll(rootDir, "\\", "/")
	cleanRoot = strings.TrimPrefix(cleanRoot, "./")
	
	data := struct {
		RootDir string
		AppName string
	}{
		RootDir: cleanRoot,
		AppName: appName,
	}
	
	appFile := filepath.Join(appPath, "app.go")
	f, err := os.Create(appFile)
	if err != nil {
		return fmt.Errorf("failed to create app.go: %v", err)
	}
	defer f.Close()
	
	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("failed to execute app template: %v", err)
	}
	
	return nil
}

func generateWailsMain(appPath, appName string) error {
	const mainTemplate = `package main

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// main function
func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create menu
	appMenu := menu.NewMenu()
	navigationMenu := appMenu.AddSubmenu("Navigation")
	
	navigationMenu.AddText("Back", keys.CmdOrCtrl("left"), func(_ *menu.CallbackData) {
		runtime.WindowExecJS(app.ctx, "history.back()")
	})
	
	navigationMenu.AddText("Forward", keys.CmdOrCtrl("right"), func(_ *menu.CallbackData) {
		runtime.WindowExecJS(app.ctx, "history.forward()")
	})
	
	navigationMenu.AddSeparator()
	
	navigationMenu.AddText("Refresh", keys.CmdOrCtrl("r"), func(_ *menu.CallbackData) {
		runtime.WindowReload(app.ctx)
	})
	
	navigationMenu.AddSeparator()
	
	navigationMenu.AddText("Home", keys.CmdOrCtrl("h"), func(_ *menu.CallbackData) {
		runtime.WindowExecJS(app.ctx, "window.location.href = '/'")
	})

	// Server control menu
	serverMenu := appMenu.AddSubmenu("Server")
	
	serverMenu.AddText("Start Server", nil, func(_ *menu.CallbackData) {
		result := app.StartServer()
		runtime.LogInfo(app.ctx, result)
	})
	
	serverMenu.AddText("Stop Server", nil, func(_ *menu.CallbackData) {
		result := app.StopServer()
		runtime.LogInfo(app.ctx, result)
	})
	
	serverMenu.AddSeparator()
	
	serverMenu.AddText("Open in Browser", keys.CmdOrCtrl("o"), func(_ *menu.CallbackData) {
		status := app.GetServerStatus()
		if status["running"].(bool) {
			url := status["url"].(string)
			runtime.BrowserOpenURL(app.ctx, url)
		}
	})

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "{{.AppName}}",
		Width:  1024,
		Height: 768,
		Menu:   appMenu,
		AssetServer: &assetserver.Options{
			Handler: NewAssetHandler(app),
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.OnStartup,
		OnShutdown:       app.OnShutdown,
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

// NewAssetHandler creates a custom asset handler that proxies to the embedded Redi server
func NewAssetHandler(app *App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Wait for server to be ready
		status := app.GetServerStatus()
		if !status["running"].(bool) {
			http.Error(w, "Server not ready", http.StatusServiceUnavailable)
			return
		}

		// Create proxy request to local server
		port := status["port"].(int)
		proxyURL := fmt.Sprintf("http://localhost:%d%s", port, r.URL.Path)
		
		// Handle query parameters
		if r.URL.RawQuery != "" {
			proxyURL += "?" + r.URL.RawQuery
		}

		// Create new request
		proxyReq, err := http.NewRequestWithContext(context.Background(), r.Method, proxyURL, r.Body)
		if err != nil {
			http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
			return
		}

		// Copy headers
		for key, values := range r.Header {
			for _, value := range values {
				proxyReq.Header.Add(key, value)
			}
		}

		// Make request to embedded server
		client := &http.Client{}
		resp, err := client.Do(proxyReq)
		if err != nil {
			http.Error(w, "Failed to proxy request", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		// Copy status code
		w.WriteHeader(resp.StatusCode)

		// Copy response body
		io.Copy(w, resp.Body)
	}
}
`

	tmpl, err := template.New("main").Parse(mainTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse main template: %v", err)
	}
	
	data := struct {
		AppName string
	}{
		AppName: appName,
	}
	
	mainFile := filepath.Join(appPath, "main.go")
	f, err := os.Create(mainFile)
	if err != nil {
		return fmt.Errorf("failed to create main.go: %v", err)
	}
	defer f.Close()
	
	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("failed to execute main template: %v", err)
	}
	
	return nil
}

func updateWailsGoMod(appPath string) error {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	
	// Check if we're running from source (has go.mod with github.com/rediwo/redi module)
	// and the current directory contains the actual redi source code
	isSourceInstall := false
	if goModData, err := os.ReadFile("go.mod"); err == nil {
		goModContent := string(goModData)
		if strings.Contains(goModContent, "module github.com/rediwo/redi") {
			// Double check by looking for main redi files
			if _, err := os.Stat("server.go"); err == nil {
				if _, err := os.Stat("router.go"); err == nil {
					isSourceInstall = true
				}
			}
		}
	}
	
	var goModContent string
	if isSourceInstall {
		// Running from source - use replace directive
		goModContent = fmt.Sprintf(`module changeme

go 1.21

require (
	github.com/wailsapp/wails/v2 v2.9.2
	github.com/rediwo/redi v0.0.0
)

replace github.com/rediwo/redi => %s
`, cwd)
	} else {
		// Running from installed binary - use published version
		goModContent = `module changeme

go 1.21

require (
	github.com/wailsapp/wails/v2 v2.9.2
	github.com/rediwo/redi v1.0.0
)
`
	}
	
	goModPath := filepath.Join(appPath, "go.mod")
	return os.WriteFile(goModPath, []byte(goModContent), 0644)
}

// getVersion returns the version of redi, trying git tag first, then build-time version
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
	if err == nil {
		return strings.TrimSpace(string(output))
	}
	
	// If no exact tag match, try to get the latest tag with commit info
	cmd = exec.Command("git", "describe", "--tags", "--always")
	output, err = cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output))
	}
	
	// If git is not available or not a git repository, return empty
	return ""
}
