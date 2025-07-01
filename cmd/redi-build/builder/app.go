package builder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// AppBuilder builds Wails desktop applications
type AppBuilder struct{}

// NewAppBuilder creates a new app builder
func NewAppBuilder() *AppBuilder {
	return &AppBuilder{}
}

// Validate validates the build configuration
func (a *AppBuilder) Validate(config Config) error {
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
	if err := a.checkWailsInstalled(); err != nil {
		return err
	}
	
	return nil
}

// Build creates a Wails desktop application
func (a *AppBuilder) Build(config Config) error {
	if err := a.Validate(config); err != nil {
		return err
	}
	
	// Create a valid module name (no spaces, lowercase)
	moduleName := strings.ToLower(strings.ReplaceAll(config.AppName, " ", "-"))
	
	// Initialize Wails project in the output directory
	initCmd := exec.Command("wails", "init", "-n", moduleName, "-t", "plain")
	initCmd.Dir = "."
	
	if err := initCmd.Run(); err != nil {
		return NewBuildError("failed to initialize Wails project", err)
	}
	
	// Wails creates the project in the current directory, move it to the desired output
	if moduleName != config.Output {
		if err := os.Rename(moduleName, config.Output); err != nil {
			return NewBuildError("failed to rename Wails project directory", err)
		}
	}
	
	appPath := config.Output
	
	// Copy root directory to app/embed
	embedPath := filepath.Join(appPath, "embed")
	if err := os.MkdirAll(embedPath, 0755); err != nil {
		return NewBuildError("failed to create embed directory", err)
	}
	
	targetDir := filepath.Join(embedPath, filepath.Base(config.Root))
	if err := copyDir(config.Root, targetDir); err != nil {
		return NewBuildError("failed to copy root directory", err)
	}
	
	// Prepare template data
	extensions := expandExtensions(config.Extensions)
	
	data := &TemplateData{
		ModuleName:      moduleName,
		ProjectName:     config.AppName,
		AppName:         config.AppName,
		RootDir:         filepath.Base(config.Root),
		Extensions:      extensions,
		RediVersion:     a.getRediVersion(),
		IsSourceInstall: a.isSourceInstall(),
		ReplaceDir:      a.getReplaceDir(),
	}
	
	// Generate app.go
	if err := a.generateFile("templates/app/app.go.tmpl", filepath.Join(appPath, "app.go"), data); err != nil {
		return NewBuildError("failed to generate app.go", err)
	}
	
	// Generate main.go
	if err := a.generateFile("templates/app/main.go.tmpl", filepath.Join(appPath, "main.go"), data); err != nil {
		return NewBuildError("failed to generate main.go", err)
	}
	
	// Generate go.mod
	if err := a.generateFile("templates/app/go.mod.tmpl", filepath.Join(appPath, "go.mod"), data); err != nil {
		return NewBuildError("failed to generate go.mod", err)
	}
	
	// Copy go.sum if exists
	if err := copyFile("go.sum", filepath.Join(appPath, "go.sum")); err != nil {
		// This is not critical, just continue
	}
	
	// Run go mod tidy
	if err := a.runGoModTidy(appPath); err != nil {
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
	
	// Try to build the app
	if err := a.tryBuild(appPath); err != nil {
		fmt.Printf("App project generated successfully, but wails build failed: %v\n", err)
		fmt.Printf("To build manually, run: cd %s && wails build\n", appPath)
		fmt.Printf("For development mode, run: cd %s && wails dev\n", appPath)
		return nil
	}
	
	fmt.Printf("Wails desktop application successfully created in: %s\n", appPath)
	fmt.Printf("To run in development mode: cd %s && wails dev\n", appPath)
	fmt.Printf("To build the app: cd %s && wails build\n", appPath)
	return nil
}

func (a *AppBuilder) generateFile(templatePath, outputPath string, data *TemplateData) error {
	tmpl, err := GetTemplate(templatePath)
	if err != nil {
		return err
	}
	
	f, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer f.Close()
	
	return tmpl.Execute(f, data)
}

func (a *AppBuilder) checkWailsInstalled() error {
	cmd := exec.Command("wails", "version")
	if err := cmd.Run(); err != nil {
		return NewBuildError("wails command not found. Please install Wails first: go install github.com/wailsapp/wails/v2/cmd/wails@latest", err)
	}
	return nil
}

func (a *AppBuilder) getRediVersion() string {
	if a.isSourceInstall() {
		return "v0.0.0"
	}
	return "v1.0.0"
}

func (a *AppBuilder) isSourceInstall() bool {
	if goModData, err := os.ReadFile("go.mod"); err == nil {
		goModContent := string(goModData)
		if strings.Contains(goModContent, "module github.com/rediwo/redi") {
			if _, err := os.Stat("server.go"); err == nil {
				if _, err := os.Stat("router.go"); err == nil {
					return true
				}
			}
		}
	}
	return false
}

func (a *AppBuilder) getReplaceDir() string {
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}
	return ""
}

func (a *AppBuilder) runGoModTidy(dir string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = dir
	return cmd.Run()
}

func (a *AppBuilder) tryBuild(dir string) error {
	cmd := exec.Command("wails", "build")
	cmd.Dir = dir
	return cmd.Run()
}