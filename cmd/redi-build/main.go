package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/rediwo/redi/cmd/redi-build/builder"
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
	case "cli":
		buildCli()
	case "embed":
		buildEmbedded()
	case "app":
		buildApp()
	case "--help", "-h", "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Redi Build Tool - Create CLI tools, embedded apps, and desktop applications\n\n")
	fmt.Fprintf(os.Stderr, "Usage: %s <command> [options]\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  cli       Build CLI application project (based on cmd/redi)\n")
	fmt.Fprintf(os.Stderr, "  embed     Build embedded executable project\n")
	fmt.Fprintf(os.Stderr, "  app       Build Wails desktop application project\n")
	fmt.Fprintf(os.Stderr, "  help      Show this help message\n")
	fmt.Fprintf(os.Stderr, "  --version Show version information\n")
	fmt.Fprintf(os.Stderr, "\nGlobal Options:\n")
	fmt.Fprintf(os.Stderr, "  --ext=ext1,ext2,ext3    Extension modules to include\n")
	fmt.Fprintf(os.Stderr, "                          Single words auto-expand to github.com/rediwo/redi-xxx\n")
	fmt.Fprintf(os.Stderr, "  --config=config.yaml    Configuration file (YAML format)\n")
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  %s cli --root=mysite --output=mycli\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s embed --root=mysite --output=myapp\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s app --root=mysite --output=myapp --name=\"My App\"\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s embed --root=mysite --ext=orm,auth,cache --output=myapp\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s cli --config=build.yaml\n", os.Args[0])
}

func buildCli() {
	cliFlags := flag.NewFlagSet("cli", flag.ExitOnError)
	
	var root string
	var output string
	var extensions string
	var configFile string
	
	cliFlags.StringVar(&root, "root", "", "Root directory to include")
	cliFlags.StringVar(&output, "output", "redi-cli", "Output directory name")
	cliFlags.StringVar(&extensions, "ext", "", "Extension modules (comma-separated)")
	cliFlags.StringVar(&configFile, "config", "", "Configuration file (YAML)")
	
	cliFlags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Build CLI application project\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s cli [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		cliFlags.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExtensions:\n")
		fmt.Fprintf(os.Stderr, "  Single words like 'orm' expand to 'github.com/rediwo/redi-orm'\n")
		fmt.Fprintf(os.Stderr, "  Use full URLs for third-party extensions\n")
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s cli --root=fixtures --output=my-cli --ext=orm,auth\n", os.Args[0])
	}
	
	cliFlags.Parse(os.Args[2:])
	
	config := parseConfig(configFile, builder.Config{
		Root:       root,
		Output:     output,
		Extensions: parseExtensions(extensions),
		ConfigFile: configFile,
	})
	
	cliBuilder := builder.NewCliBuilder()
	
	log.Printf("Building CLI application with root directory: %s", config.Root)
	log.Printf("Output directory: %s", config.Output)
	if len(config.Extensions) > 0 {
		log.Printf("Extensions: %s", strings.Join(config.Extensions, ", "))
	}
	
	if err := cliBuilder.Build(config); err != nil {
		log.Fatalf("Build failed: %v", err)
	}
}

func buildEmbedded() {
	embedFlags := flag.NewFlagSet("embed", flag.ExitOnError)
	
	var root string
	var output string
	var extensions string
	var configFile string
	
	embedFlags.StringVar(&root, "root", "", "Root directory to embed")
	embedFlags.StringVar(&output, "output", "redi-embedded", "Output directory name")
	embedFlags.StringVar(&extensions, "ext", "", "Extension modules (comma-separated)")
	embedFlags.StringVar(&configFile, "config", "", "Configuration file (YAML)")
	
	embedFlags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Build embedded executable project\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s embed [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		embedFlags.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExtensions:\n")
		fmt.Fprintf(os.Stderr, "  Single words like 'orm' expand to 'github.com/rediwo/redi-orm'\n")
		fmt.Fprintf(os.Stderr, "  Use full URLs for third-party extensions\n")
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s embed --root=fixtures --output=my-app --ext=orm,cache\n", os.Args[0])
	}
	
	embedFlags.Parse(os.Args[2:])
	
	config := parseConfig(configFile, builder.Config{
		Root:       root,
		Output:     output,
		Extensions: parseExtensions(extensions),
		ConfigFile: configFile,
	})
	
	embedBuilder := builder.NewEmbedBuilder()
	
	log.Printf("Building embedded application with root directory: %s", config.Root)
	log.Printf("Output directory: %s", config.Output)
	if len(config.Extensions) > 0 {
		log.Printf("Extensions: %s", strings.Join(config.Extensions, ", "))
	}
	
	if err := embedBuilder.Build(config); err != nil {
		log.Fatalf("Build failed: %v", err)
	}
}

func buildApp() {
	appFlags := flag.NewFlagSet("app", flag.ExitOnError)
	
	var root string
	var output string
	var appName string
	var extensions string
	var configFile string
	
	appFlags.StringVar(&root, "root", "", "Root directory to embed")
	appFlags.StringVar(&output, "output", "redi-app", "Output directory name")
	appFlags.StringVar(&appName, "name", "Redi App", "Application name")
	appFlags.StringVar(&extensions, "ext", "", "Extension modules (comma-separated)")
	appFlags.StringVar(&configFile, "config", "", "Configuration file (YAML)")
	
	appFlags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Build Wails desktop application project\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s app [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		appFlags.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExtensions:\n")
		fmt.Fprintf(os.Stderr, "  Single words like 'orm' expand to 'github.com/rediwo/redi-orm'\n")
		fmt.Fprintf(os.Stderr, "  Use full URLs for third-party extensions\n")
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s app --root=fixtures --output=my-app --name=\"My Application\" --ext=orm,auth\n", os.Args[0])
	}
	
	appFlags.Parse(os.Args[2:])
	
	config := parseConfig(configFile, builder.Config{
		Root:       root,
		Output:     output,
		AppName:    appName,
		Extensions: parseExtensions(extensions),
		ConfigFile: configFile,
	})
	
	appBuilder := builder.NewAppBuilder()
	
	log.Printf("Building Wails desktop application with root directory: %s", config.Root)
	log.Printf("Application name: %s", config.AppName)
	log.Printf("Output directory: %s", config.Output)
	if len(config.Extensions) > 0 {
		log.Printf("Extensions: %s", strings.Join(config.Extensions, ", "))
	}
	
	if err := appBuilder.Build(config); err != nil {
		log.Fatalf("Build failed: %v", err)
	}
}

func parseExtensions(extensions string) []string {
	if extensions == "" {
		return nil
	}
	
	var result []string
	for _, ext := range strings.Split(extensions, ",") {
		ext = strings.TrimSpace(ext)
		if ext != "" {
			result = append(result, ext)
		}
	}
	return result
}

func parseConfig(configFile string, defaultConfig builder.Config) builder.Config {
	if configFile == "" {
		return defaultConfig
	}
	
	// TODO: Implement YAML config parsing
	log.Printf("Config file support not yet implemented: %s", configFile)
	return defaultConfig
}