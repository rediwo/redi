package console

import (
	"fmt"
	"os"
	
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
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

// Enable registers the console module with custom printer in the given registry
func Enable(registry *require.Registry) {
	registry.RegisterNativeModule(console.ModuleName, console.RequireWithPrinter(NewCustomPrinter()))
}