package builder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

// CliBuilder builds JavaScript CLI applications based on cmd/rejs
type CliBuilder struct{}

// NewCliBuilder creates a new CLI builder
func NewCliBuilder() *CliBuilder {
	return &CliBuilder{}
}

// Validate validates the build configuration
func (c *CliBuilder) Validate(config Config) error {
	if config.ScriptPath == "" {
		return NewBuildError("script path is required", nil)
	}
	
	// Check if script exists
	if _, err := os.Stat(config.ScriptPath); os.IsNotExist(err) {
		return NewBuildError("script file does not exist", err)
	}
	
	if config.Output == "" {
		return NewBuildError("output binary name is required", nil)
	}
	
	return nil
}

// Build creates a JavaScript CLI application
func (c *CliBuilder) Build(config Config) error {
	if err := c.Validate(config); err != nil {
		return err
	}
	
	// Create output directory
	outputDir := filepath.Dir(config.Output)
	if outputDir != "." && outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return NewBuildError("failed to create output directory", err)
		}
	}
	
	// Get script filename
	scriptName := filepath.Base(config.ScriptPath)
	
	// Prepare template data
	binaryName := filepath.Base(config.Output)
	extensions := expandExtensions(config.Extensions)
	
	data := &TemplateData{
		BinaryName:      binaryName,
		ScriptName:      scriptName,
		Extensions:      extensions,
		RediVersion:     c.getRediVersion(),
		IsSourceInstall: c.isSourceInstall(),
		ReplaceDir:      c.getReplaceDir(),
	}
	
	// Create temporary build directory
	tempDir, err := os.MkdirTemp("", "redi-cli-build-*")
	if err != nil {
		return NewBuildError("failed to create temp directory", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Copy the script file to temp directory
	scriptDest := filepath.Join(tempDir, scriptName)
	if err := copyFile(config.ScriptPath, scriptDest); err != nil {
		return NewBuildError("failed to copy script file", err)
	}
	
	// Generate main.go with embed directive
	mainPath := filepath.Join(tempDir, "main.go")
	if err := c.generateMainGo(mainPath, data); err != nil {
		return NewBuildError("failed to generate main.go", err)
	}
	
	// Generate go.mod
	goModPath := filepath.Join(tempDir, "go.mod")
	if err := c.generateFile("templates/cli/go.mod.tmpl", goModPath, data); err != nil {
		return NewBuildError("failed to generate go.mod", err)
	}
	
	// Copy go.sum if exists
	if err := copyFile("go.sum", filepath.Join(tempDir, "go.sum")); err != nil {
		// This is not critical, just continue
	}
	
	// Run go mod tidy
	if err := c.runGoModTidy(tempDir); err != nil {
		return NewBuildError("failed to run go mod tidy", err)
	}
	
	// Build the binary
	outputPath := config.Output
	if !filepath.IsAbs(outputPath) {
		cwd, _ := os.Getwd()
		outputPath = filepath.Join(cwd, outputPath)
	}
	if err := c.buildBinary(tempDir, outputPath); err != nil {
		return NewBuildError("failed to build binary", err)
	}
	
	fmt.Printf("JavaScript CLI successfully built: %s\n", config.Output)
	fmt.Printf("You can now run: %s <script-args>\n", config.Output)
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
			if _, err := os.Stat("runtime"); err == nil {
				return true
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
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("go mod tidy error: %s\n", string(output))
	}
	return err
}

func (c *CliBuilder) buildBinary(sourceDir, outputPath string) error {
	cmd := exec.Command("go", "build", "-o", outputPath, ".")
	cmd.Dir = sourceDir
	return cmd.Run()
}

func (c *CliBuilder) generateMainGo(outputPath string, data *TemplateData) error {
	mainTemplate := `package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	_ "embed"

	"github.com/rediwo/redi/runtime"
{{- range .Extensions }}
	_ "{{ . }}"
{{- end }}
)

var (
	// Version will be set by build flags or git tag
	Version = "dev"
)

//go:embed {{.ScriptName}}
var scriptContent string

func main() {
	// Define command line flags
	var (
		showVersion = false
		timeoutMs   = 0
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
		} else if !strings.HasPrefix(arg, "-") {
			// All remaining arguments are script arguments
			scriptArgs = args[i:]
			break
		} else {
			fmt.Fprintf(os.Stderr, "Error: Unknown option: %s\n", arg)
			os.Exit(1)
		}
	}

	if showVersion {
		versionProvider := runtime.NewVersionProvider(Version)
		fmt.Printf("{{.BinaryName}} version %s\n", versionProvider.GetVersion())
		os.Exit(0)
	}

	// Create temporary file for script
	tempFile, err := os.CreateTemp("", "{{.BinaryName}}-*.js")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create temp file: %v\n", err)
		os.Exit(1)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString(scriptContent); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to write script: %v\n", err)
		os.Exit(1)
	}
	tempFile.Close()

	// Run the script with timeout
	timeoutDuration := time.Duration(timeoutMs) * time.Millisecond
	versionProvider := runtime.NewVersionProvider(Version)
	
	exitCode, err := runtime.ExecuteScript(tempFile.Name(), scriptArgs, timeoutDuration, versionProvider.GetVersion())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Script error: %v\n", err)
	}
	
	os.Exit(exitCode)
}
`
	tmpl, err := template.New("main").Parse(mainTemplate)
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