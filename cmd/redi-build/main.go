package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/rediwo/redi/builder"
	"github.com/rediwo/redi/runtime"
)

var (
	// Version will be set by build flags or git tag
	Version = "dev"
)

func main() {
	// Check for version flag first
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		versionProvider := runtime.NewVersionProvider(Version)
		fmt.Printf("redi-build version %s\n", versionProvider.GetVersion())
		return
	}

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "embed":
		buildEmbedded()
	case "wails":
		buildWails()
	case "--help", "-h", "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Redi Build Tool - Create embedded executables and Wails desktop apps\n\n")
	fmt.Fprintf(os.Stderr, "Usage: %s <command> [options]\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  embed     Build embedded executable\n")
	fmt.Fprintf(os.Stderr, "  wails     Build Wails desktop application\n")
	fmt.Fprintf(os.Stderr, "  help      Show this help message\n")
	fmt.Fprintf(os.Stderr, "  --version Show version information\n")
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  %s embed --root=mysite --output=myapp\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s wails --root=mysite --output=myapp --name=\"My App\"\n", os.Args[0])
}

func buildEmbedded() {
	embedFlags := flag.NewFlagSet("embed", flag.ExitOnError)
	
	var root string
	var output string
	
	embedFlags.StringVar(&root, "root", "", "Root directory to embed")
	embedFlags.StringVar(&output, "output", "redi-embedded", "Output executable name")
	
	embedFlags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Build embedded executable\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s embed [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		embedFlags.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s embed --root=fixtures --output=my-app\n", os.Args[0])
	}
	
	embedFlags.Parse(os.Args[2:])
	
	config := builder.Config{
		Root:   root,
		Output: output,
	}
	
	embedBuilder := builder.NewEmbedBuilder()
	
	log.Printf("Building embedded version with root directory: %s", root)
	log.Printf("Output executable: %s", output)
	
	if err := embedBuilder.Build(config); err != nil {
		log.Fatalf("Build failed: %v", err)
	}
	
	log.Printf("Successfully built embedded executable: %s", output)
}

func buildWails() {
	wailsFlags := flag.NewFlagSet("wails", flag.ExitOnError)
	
	var root string
	var output string
	var appName string
	
	wailsFlags.StringVar(&root, "root", "", "Root directory to embed")
	wailsFlags.StringVar(&output, "output", "redi-app", "Output directory name")
	wailsFlags.StringVar(&appName, "name", "Redi App", "Application name")
	
	wailsFlags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Build Wails desktop application\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s wails [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		wailsFlags.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s wails --root=fixtures --output=my-app --name=\"My Application\"\n", os.Args[0])
	}
	
	wailsFlags.Parse(os.Args[2:])
	
	config := builder.Config{
		Root:    root,
		Output:  output,
		AppName: appName,
	}
	
	wailsBuilder := builder.NewWailsBuilder()
	
	log.Printf("Building Wails desktop application with root directory: %s", root)
	log.Printf("Application name: %s", appName)
	log.Printf("Output directory: %s", output)
	
	if err := wailsBuilder.Build(config); err != nil {
		log.Fatalf("Build failed: %v", err)
	}
	
	appPath := fmt.Sprintf("%s/%s", output, strings.ToLower(strings.ReplaceAll(appName, " ", "-")))
	log.Printf("Successfully created Wails application in: %s", appPath)
	log.Printf("To run in development mode: cd %s && wails dev", appPath)
	log.Printf("To build the app: cd %s && wails build", appPath)
}

