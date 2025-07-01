package builder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// EmbedBuilder builds embedded project source code
type EmbedBuilder struct{}

// NewEmbedBuilder creates a new embed builder
func NewEmbedBuilder() *EmbedBuilder {
	return &EmbedBuilder{}
}

// Validate validates the build configuration
func (e *EmbedBuilder) Validate(config Config) error {
	if err := ValidateRoot(config.Root); err != nil {
		return err
	}
	
	if config.Output == "" {
		return NewBuildError("output name is required", nil)
	}
	
	return nil
}

// Build creates an embedded executable project
func (e *EmbedBuilder) Build(config Config) error {
	if err := e.Validate(config); err != nil {
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
	moduleName := config.Output + "-embedded"
	binaryName := config.Output
	projectName := strings.Title(config.Output) + " Embedded"
	
	extensions := expandExtensions(config.Extensions)
	
	data := &TemplateData{
		ModuleName:      moduleName,
		ProjectName:     projectName,
		BinaryName:      binaryName,
		RootDir:         filepath.Base(config.Root),
		Extensions:      extensions,
		RediVersion:     e.getRediVersion(),
		IsSourceInstall: e.isSourceInstall(),
		ReplaceDir:      e.getReplaceDir(),
	}
	
	// Generate main.go
	if err := e.generateFile("templates/embed/main.go.tmpl", filepath.Join(config.Output, "main.go"), data); err != nil {
		return NewBuildError("failed to generate main.go", err)
	}
	
	// Generate go.mod
	if err := e.generateFile("templates/embed/go.mod.tmpl", filepath.Join(config.Output, "go.mod"), data); err != nil {
		return NewBuildError("failed to generate go.mod", err)
	}
	
	// Generate Makefile
	if err := e.generateFile("templates/embed/Makefile.tmpl", filepath.Join(config.Output, "Makefile"), data); err != nil {
		return NewBuildError("failed to generate Makefile", err)
	}
	
	// Copy go.sum if exists
	if err := copyFile("go.sum", filepath.Join(config.Output, "go.sum")); err != nil {
		// This is not critical, just continue
	}
	
	// Run go mod tidy
	if err := e.runGoModTidy(config.Output); err != nil {
		return NewBuildError("failed to run go mod tidy", err)
	}
	
	// Try to build the project
	if err := e.tryBuild(config.Output, binaryName); err != nil {
		fmt.Printf("Project generated successfully, but compilation failed: %v\n", err)
		fmt.Printf("To build manually, run: cd %s && make build\n", config.Output)
		return nil
	}
	
	fmt.Printf("Embedded project successfully created and built in: %s\n", config.Output)
	fmt.Printf("Binary available at: %s\n", filepath.Join(config.Output, "bin", binaryName))
	return nil
}

func (e *EmbedBuilder) generateFile(templatePath, outputPath string, data *TemplateData) error {
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

func (e *EmbedBuilder) getRediVersion() string {
	if e.isSourceInstall() {
		return "v0.0.0"
	}
	return "v1.0.0"
}

func (e *EmbedBuilder) isSourceInstall() bool {
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

func (e *EmbedBuilder) getReplaceDir() string {
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}
	return ""
}

func (e *EmbedBuilder) runGoModTidy(dir string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = dir
	return cmd.Run()
}

func (e *EmbedBuilder) tryBuild(dir, binaryName string) error {
	cmd := exec.Command("make", "build")
	cmd.Dir = dir
	return cmd.Run()
}