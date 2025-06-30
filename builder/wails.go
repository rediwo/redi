package builder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

// WailsBuilder builds Wails desktop applications
type WailsBuilder struct{}

// NewWailsBuilder creates a new Wails builder
func NewWailsBuilder() *WailsBuilder {
	return &WailsBuilder{}
}

// Validate validates the build configuration
func (w *WailsBuilder) Validate(config Config) error {
	if err := ValidateRoot(config.Root); err != nil {
		return err
	}
	
	if config.Output == "" {
		return NewBuildError("output directory name is required", nil)
	}
	
	if config.AppName == "" {
		return NewBuildError("application name is required", nil)
	}
	
	// Check if wails is installed
	if err := w.checkWailsInstalled(); err != nil {
		return err
	}
	
	return nil
}

// Build creates a Wails desktop application
func (w *WailsBuilder) Build(config Config) error {
	if err := w.Validate(config); err != nil {
		return err
	}
	
	// Create app directory
	if err := os.MkdirAll(config.Output, 0755); err != nil {
		return NewBuildError("failed to create output directory", err)
	}
	
	// Create a valid module name (no spaces, lowercase)
	moduleName := strings.ToLower(strings.ReplaceAll(config.AppName, " ", "-"))
	
	// Initialize Wails project in the output directory
	initCmd := exec.Command("wails", "init", "-n", moduleName, "-t", "plain")
	initCmd.Dir = config.Output
	
	if err := initCmd.Run(); err != nil {
		return NewBuildError("failed to initialize Wails project", err)
	}
	
	// Copy root directory to app/embed
	appPath := filepath.Join(config.Output, moduleName)
	embedPath := filepath.Join(appPath, "embed")
	
	if err := os.MkdirAll(embedPath, 0755); err != nil {
		return NewBuildError("failed to create embed directory", err)
	}
	
	targetDir := filepath.Join(embedPath, filepath.Base(config.Root))
	if err := copyDir(config.Root, targetDir); err != nil {
		return NewBuildError("failed to copy root directory", err)
	}
	
	// Generate Wails app.go
	if err := w.generateWailsApp(appPath, filepath.Base(config.Root), config.AppName); err != nil {
		return NewBuildError("failed to generate Wails app", err)
	}
	
	// Update main.go
	if err := w.generateWailsMain(appPath, config.AppName); err != nil {
		return NewBuildError("failed to generate Wails main", err)
	}
	
	// Update go.mod
	if err := w.updateWailsGoMod(appPath); err != nil {
		return NewBuildError("failed to update go.mod", err)
	}
	
	// Copy go.sum
	if err := copyFile("go.sum", filepath.Join(appPath, "go.sum")); err != nil {
		// This is not critical, just log it
	}
	
	// Run go mod tidy
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = appPath
	if err := tidyCmd.Run(); err != nil {
		return NewBuildError("failed to run go mod tidy", err)
	}
	
	// Create a minimal frontend dist directory
	distPath := filepath.Join(appPath, "frontend", "dist")
	if err := os.MkdirAll(distPath, 0755); err != nil {
		return NewBuildError("failed to create frontend/dist", err)
	}
	
	// Copy frontend src to dist
	srcPath := filepath.Join(appPath, "frontend", "src")
	if err := copyDir(srcPath, distPath); err != nil {
		// This is not critical, just continue
	}
	
	return nil
}

func (w *WailsBuilder) checkWailsInstalled() error {
	cmd := exec.Command("wails", "version")
	if err := cmd.Run(); err != nil {
		return NewBuildError("wails command not found. Please install Wails first: go install github.com/wailsapp/wails/v2/cmd/wails@latest", err)
	}
	return nil
}

func (w *WailsBuilder) generateWailsApp(appPath, rootDir, appName string) error {
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

func (w *WailsBuilder) generateWailsMain(appPath, appName string) error {
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

func (w *WailsBuilder) updateWailsGoMod(appPath string) error {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	
	// Check if we're running from source
	isSourceInstall := false
	if goModData, err := os.ReadFile("go.mod"); err == nil {
		goModContent := string(goModData)
		if strings.Contains(goModContent, "module github.com/rediwo/redi") {
			if _, err := os.Stat("server.go"); err == nil {
				if _, err := os.Stat("router.go"); err == nil {
					isSourceInstall = true
				}
			}
		}
	}
	
	var goModContent string
	if isSourceInstall {
		goModContent = fmt.Sprintf(`module changeme

go 1.21

require (
	github.com/wailsapp/wails/v2 v2.9.2
	github.com/rediwo/redi v0.0.0
)

replace github.com/rediwo/redi => %s
`, cwd)
	} else {
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