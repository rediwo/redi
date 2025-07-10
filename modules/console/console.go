package console

import (
	"fmt"
	"os"
	
	"github.com/dop251/goja_nodejs/console"
	"github.com/rediwo/redi/registry"
)

// CustomPrinter implements the console.Printer interface
// without timestamps, just plain output
type CustomPrinter struct{}

// Log prints to stdout without timestamp
func (p *CustomPrinter) Log(s string) {
	fmt.Fprintln(os.Stdout, s)
}

// Warn prints to stderr without timestamp
func (p *CustomPrinter) Warn(s string) {
	fmt.Fprintln(os.Stderr, s)
}

// Error prints to stderr without timestamp
func (p *CustomPrinter) Error(s string) {
	fmt.Fprintln(os.Stderr, s)
}

// NewCustomPrinter creates a new custom printer instance
func NewCustomPrinter() *CustomPrinter {
	return &CustomPrinter{}
}

// init registers the console module automatically
func init() {
	registry.RegisterModule("console", initConsoleModule)
}

// initConsoleModule initializes the console module
func initConsoleModule(config registry.ModuleConfig) error {
	config.Registry.RegisterNativeModule(console.ModuleName, console.RequireWithPrinter(NewCustomPrinter()))
	return nil
}

