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
	case "server":
		buildServer()
	case "standalone":
		buildStandalone()
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
	fmt.Fprintf(os.Stderr, "  cli         Build JavaScript CLI tool (based on cmd/rejs)\n")
	fmt.Fprintf(os.Stderr, "  server      Build server application project (based on cmd/redi)\n")
	fmt.Fprintf(os.Stderr, "  standalone  Build standalone executable project\n")
	fmt.Fprintf(os.Stderr, "  app         Build Wails desktop application project\n")
	fmt.Fprintf(os.Stderr, "  help        Show this help message\n")
	fmt.Fprintf(os.Stderr, "  --version   Show version information\n")
	fmt.Fprintf(os.Stderr, "\nGlobal Options:\n")
	fmt.Fprintf(os.Stderr, "  --ext=ext1,ext2,ext3    Extension modules to include\n")
	fmt.Fprintf(os.Stderr, "                          Single words auto-expand to github.com/rediwo/redi-xxx\n")
	fmt.Fprintf(os.Stderr, "  --config=config.yaml    Configuration file (YAML format)\n")
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  %s cli --script=main.js --output=mycli\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s server --root=mysite --output=myserver\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s standalone --root=mysite --output=myapp\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s app --root=mysite --output=myapp --name=\"My App\"\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s cli --script=app.js --ext=orm,auth --output=mycli\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s server --config=build.yaml\n", os.Args[0])
}

func buildServer() {
	serverFlags := flag.NewFlagSet("server", flag.ExitOnError)
	
	var root string
	var output string
	var extensions string
	var configFile string
	
	serverFlags.StringVar(&root, "root", "", "Root directory to include")
	serverFlags.StringVar(&output, "output", "redi-server", "Output directory name")
	serverFlags.StringVar(&extensions, "ext", "", "Extension modules (comma-separated)")
	serverFlags.StringVar(&configFile, "config", "", "Configuration file (YAML)")
	
	serverFlags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Build server application project\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s server [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		serverFlags.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExtensions:\n")
		fmt.Fprintf(os.Stderr, "  Single words like 'orm' expand to 'github.com/rediwo/redi-orm'\n")
		fmt.Fprintf(os.Stderr, "  Use full URLs for third-party extensions\n")
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s server --root=fixtures --output=my-server --ext=orm,auth\n", os.Args[0])
	}
	
	serverFlags.Parse(os.Args[2:])
	
	config := parseConfig(configFile, builder.Config{
		Root:       root,
		Output:     output,
		Extensions: parseExtensions(extensions),
		ConfigFile: configFile,
	})
	
	serverBuilder := builder.NewServerBuilder()
	
	log.Printf("Building server application with root directory: %s", config.Root)
	log.Printf("Output directory: %s", config.Output)
	if len(config.Extensions) > 0 {
		log.Printf("Extensions: %s", strings.Join(config.Extensions, ", "))
	}
	
	if err := serverBuilder.Build(config); err != nil {
		log.Fatalf("Build failed: %v", err)
	}
}

func buildStandalone() {
	standaloneFlags := flag.NewFlagSet("standalone", flag.ExitOnError)
	
	var root string
	var output string
	var extensions string
	var configFile string
	
	standaloneFlags.StringVar(&root, "root", "", "Root directory to embed")
	standaloneFlags.StringVar(&output, "output", "redi-standalone", "Output directory name")
	standaloneFlags.StringVar(&extensions, "ext", "", "Extension modules (comma-separated)")
	standaloneFlags.StringVar(&configFile, "config", "", "Configuration file (YAML)")
	
	standaloneFlags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Build standalone executable project\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s standalone [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		standaloneFlags.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExtensions:\n")
		fmt.Fprintf(os.Stderr, "  Single words like 'orm' expand to 'github.com/rediwo/redi-orm'\n")
		fmt.Fprintf(os.Stderr, "  Use full URLs for third-party extensions\n")
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s standalone --root=fixtures --output=my-app --ext=orm,cache\n", os.Args[0])
	}
	
	standaloneFlags.Parse(os.Args[2:])
	
	config := parseConfig(configFile, builder.Config{
		Root:       root,
		Output:     output,
		Extensions: parseExtensions(extensions),
		ConfigFile: configFile,
	})
	
	standaloneBuilder := builder.NewStandaloneBuilder()
	
	log.Printf("Building standalone application with root directory: %s", config.Root)
	log.Printf("Output directory: %s", config.Output)
	if len(config.Extensions) > 0 {
		log.Printf("Extensions: %s", strings.Join(config.Extensions, ", "))
	}
	
	if err := standaloneBuilder.Build(config); err != nil {
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

func buildCli() {
	cliFlags := flag.NewFlagSet("cli", flag.ExitOnError)
	
	var scriptPath string
	var output string
	var extensions string
	var configFile string
	
	cliFlags.StringVar(&scriptPath, "script", "", "Main JavaScript file")
	cliFlags.StringVar(&output, "output", "redi-cli", "Output binary name")
	cliFlags.StringVar(&extensions, "ext", "", "Extension modules (comma-separated)")
	cliFlags.StringVar(&configFile, "config", "", "Configuration file (YAML)")
	
	cliFlags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Build JavaScript CLI tool (based on rejs)\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s cli [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		cliFlags.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExtensions:\n")
		fmt.Fprintf(os.Stderr, "  Single words like 'orm' expand to 'github.com/rediwo/redi-orm'\n")
		fmt.Fprintf(os.Stderr, "  Use full URLs for third-party extensions\n")
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s cli --script=main.js --output=mycli --ext=orm,auth\n", os.Args[0])
	}
	
	cliFlags.Parse(os.Args[2:])
	
	config := parseConfig(configFile, builder.Config{
		ScriptPath: scriptPath,
		Output:     output,
		Extensions: parseExtensions(extensions),
		ConfigFile: configFile,
	})
	
	cliBuilder := builder.NewCliBuilder()
	
	log.Printf("Building JavaScript CLI with main script: %s", config.ScriptPath)
	log.Printf("Output binary: %s", config.Output)
	if len(config.Extensions) > 0 {
		log.Printf("Extensions: %s", strings.Join(config.Extensions, ", "))
	}
	
	if err := cliBuilder.Build(config); err != nil {
		log.Fatalf("Build failed: %v", err)
	}
}

func parseConfig(configFile string, defaultConfig builder.Config) builder.Config {
	if configFile == "" {
		return defaultConfig
	}
	
	// TODO: Implement YAML config parsing
	log.Printf("Config file support not yet implemented: %s", configFile)
	return defaultConfig
}