package builder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CliBuilder builds CLI applications based on cmd/redi
type CliBuilder struct{}

// NewCliBuilder creates a new CLI builder
func NewCliBuilder() *CliBuilder {
	return &CliBuilder{}
}

// Validate validates the build configuration
func (c *CliBuilder) Validate(config Config) error {
	if err := ValidateRoot(config.Root); err != nil {
		return err
	}
	
	if config.Output == "" {
		return NewBuildError("output directory name is required", nil)
	}
	
	return nil
}

// Build creates a CLI application project
func (c *CliBuilder) Build(config Config) error {
	if err := c.Validate(config); err != nil {
		return err
	}
	
	// Create output directory
	if err := os.MkdirAll(config.Output, 0755); err != nil {
		return NewBuildError("failed to create output directory", err)
	}
	
	// Create bin directory for compiled binaries
	binDir := filepath.Join(config.Output, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return NewBuildError("failed to create bin directory", err)
	}
	
	// Copy root directory to output
	targetDir := filepath.Join(config.Output, filepath.Base(config.Root))
	if err := copyDir(config.Root, targetDir); err != nil {
		return NewBuildError("failed to copy root directory", err)
	}
	
	// Prepare template data
	moduleName := config.Output
	if config.AppName != "" {
		moduleName = strings.ToLower(strings.ReplaceAll(config.AppName, " ", "-"))
	}
	
	binaryName := moduleName
	projectName := config.AppName
	if projectName == "" {
		projectName = strings.Title(moduleName)
	}
	
	extensions := expandExtensions(config.Extensions)
	
	data := &TemplateData{
		ModuleName:      moduleName,
		ProjectName:     projectName,
		BinaryName:      binaryName,
		AppName:         projectName,
		RootDir:         filepath.Base(config.Root),
		Extensions:      extensions,
		RediVersion:     c.getRediVersion(),
		IsSourceInstall: c.isSourceInstall(),
		ReplaceDir:      c.getReplaceDir(),
	}
	
	// Generate main.go
	if err := c.generateFile("templates/cli/main.go.tmpl", filepath.Join(config.Output, "main.go"), data); err != nil {
		return NewBuildError("failed to generate main.go", err)
	}
	
	// Generate go.mod
	if err := c.generateFile("templates/cli/go.mod.tmpl", filepath.Join(config.Output, "go.mod"), data); err != nil {
		return NewBuildError("failed to generate go.mod", err)
	}
	
	// Generate Makefile
	if err := c.generateFile("templates/cli/Makefile.tmpl", filepath.Join(config.Output, "Makefile"), data); err != nil {
		return NewBuildError("failed to generate Makefile", err)
	}
	
	// Copy go.sum if exists
	if err := copyFile("go.sum", filepath.Join(config.Output, "go.sum")); err != nil {
		// This is not critical, just continue
	}
	
	// Run go mod tidy
	if err := c.runGoModTidy(config.Output); err != nil {
		return NewBuildError("failed to run go mod tidy", err)
	}
	
	// Try to build the project
	if err := c.tryBuild(config.Output, binaryName); err != nil {
		fmt.Printf("Project generated successfully, but compilation failed: %v\n", err)
		fmt.Printf("To build manually, run: cd %s && make build\n", config.Output)
		return nil
	}
	
	fmt.Printf("CLI project successfully created and built in: %s\n", config.Output)
	fmt.Printf("Binary available at: %s\n", filepath.Join(config.Output, "bin", binaryName))
	return nil
}

func (c *CliBuilder) generateFile(templatePath, outputPath string, data *TemplateData) error {
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

func (c *CliBuilder) getRediVersion() string {
	if c.isSourceInstall() {
		return "v0.0.0"
	}
	return "v1.0.0"
}

func (c *CliBuilder) isSourceInstall() bool {
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

func (c *CliBuilder) getReplaceDir() string {
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}
	return ""
}

func (c *CliBuilder) runGoModTidy(dir string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = dir
	return cmd.Run()
}

func (c *CliBuilder) tryBuild(dir, binaryName string) error {
	cmd := exec.Command("make", "build")
	cmd.Dir = dir
	return cmd.Run()
}