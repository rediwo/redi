package builder

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

// EmbedBuilder builds embedded executables
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

// Build creates an embedded executable
func (e *EmbedBuilder) Build(config Config) error {
	if err := e.Validate(config); err != nil {
		return err
	}
	
	// Create temporary directory for build files
	tempDir, err := os.MkdirTemp("", "redi-build-*")
	if err != nil {
		return NewBuildError("failed to create temp directory", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Copy root directory to temp directory
	targetDir := filepath.Join(tempDir, filepath.Base(config.Root))
	if err := copyDir(config.Root, targetDir); err != nil {
		return NewBuildError("failed to copy root directory", err)
	}
	
	// Generate the embedded main.go
	if err := e.generateEmbeddedMain(tempDir, filepath.Base(config.Root)); err != nil {
		return NewBuildError("failed to generate embedded main", err)
	}
	
	// Create a new go.mod for the embedded app
	if err := e.createEmbeddedGoMod(tempDir); err != nil {
		return NewBuildError("failed to create embedded go.mod", err)
	}
	
	if err := copyFile("go.sum", filepath.Join(tempDir, "go.sum")); err != nil {
		return NewBuildError("failed to copy go.sum", err)
	}
	
	// Run go mod tidy
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = tempDir
	if err := tidyCmd.Run(); err != nil {
		return NewBuildError("failed to run go mod tidy", err)
	}
	
	// Build the embedded executable
	cmd := exec.Command("go", "build", "-o", config.Output, ".")
	cmd.Dir = tempDir
	
	if err := cmd.Run(); err != nil {
		return NewBuildError("failed to build embedded executable", err)
	}
	
	// Move the executable to current directory
	builtPath := filepath.Join(tempDir, config.Output)
	finalPath := filepath.Join(".", config.Output)
	
	if err := os.Rename(builtPath, finalPath); err != nil {
		return NewBuildError("failed to move executable", err)
	}
	
	return nil
}

func (e *EmbedBuilder) generateEmbeddedMain(tempDir, rootDir string) error {
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

func (e *EmbedBuilder) createEmbeddedGoMod(tempDir string) error {
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
		goModContent = fmt.Sprintf(`module redi-embedded

go 1.23

require github.com/rediwo/redi v0.0.0

replace github.com/rediwo/redi => %s
`, cwd)
	} else {
		goModContent = `module redi-embedded

go 1.23

require github.com/rediwo/redi v1.0.0
`
	}
	
	goModPath := filepath.Join(tempDir, "go.mod")
	return os.WriteFile(goModPath, []byte(goModContent), 0644)
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
	
	_, err = io.Copy(dstFile, srcFile)
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