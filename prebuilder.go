package redi

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rediwo/redi/filesystem"
	"github.com/rediwo/redi/handlers"
	"github.com/rediwo/redi/logging"
)

// PreBuilder handles pre-compilation of Svelte components
type PreBuilder struct {
	fs              filesystem.FileSystem
	svelteHandler   *handlers.SvelteHandler
	routesDir       string
	parallelWorkers int
	components      []string
	totalComponents int32
	compiledCount   int32
	errorCount      int32
	startTime       time.Time
	mu              sync.Mutex
	errors          []error
}

// NewServerPreBuilder creates a new PreBuilder instance
func NewServerPreBuilder(fs filesystem.FileSystem, svelteHandler *handlers.SvelteHandler, routesDir string, parallelWorkers int) *PreBuilder {
	if parallelWorkers <= 0 {
		parallelWorkers = 4
	}
	return &PreBuilder{
		fs:              fs,
		svelteHandler:   svelteHandler,
		routesDir:       routesDir,
		parallelWorkers: parallelWorkers,
		components:      make([]string, 0),
		errors:          make([]error, 0),
	}
}

// Build scans for all Svelte components and pre-compiles them
func (pb *PreBuilder) Build() error {
	pb.startTime = time.Now()
	
	// Get all Svelte files from the handler
	logging.Info("Scanning for Svelte components")
	components, err := pb.svelteHandler.GetAllSvelteFiles()
	if err != nil {
		return fmt.Errorf("failed to scan components: %w", err)
	}
	
	pb.components = components
	
	if len(pb.components) == 0 {
		logging.Info("No Svelte components found to pre-build")
		return nil
	}
	
	pb.totalComponents = int32(len(pb.components))
	logging.Info("Starting pre-build", "components", pb.totalComponents, "workers", pb.parallelWorkers)
	
	// Create worker pool
	jobs := make(chan string, len(pb.components))
	var wg sync.WaitGroup
	
	// Start workers
	for i := 0; i < pb.parallelWorkers; i++ {
		wg.Add(1)
		go pb.worker(i+1, jobs, &wg)
	}
	
	// Send all components to job queue
	for _, component := range pb.components {
		jobs <- component
	}
	close(jobs)
	
	// Wait for all workers to complete
	wg.Wait()
	
	// Print summary
	elapsed := time.Since(pb.startTime)
	compiled := atomic.LoadInt32(&pb.compiledCount)
	errors := atomic.LoadInt32(&pb.errorCount)
	
	if compiled > 0 {
		avgTime := elapsed / time.Duration(compiled)
		logging.Info("Pre-build completed", 
			"duration", elapsed.Round(time.Millisecond).String(),
			"compiled", compiled,
			"errors", errors,
			"avgTime", avgTime.Round(time.Millisecond).String())
	} else {
		logging.Info("Pre-build completed", 
			"duration", elapsed.Round(time.Millisecond).String(),
			"compiled", compiled,
			"errors", errors)
	}
	
	if errors > 0 {
		pb.printErrors()
		return fmt.Errorf("pre-build completed with %d errors", errors)
	}
	
	return nil
}

// worker processes components from the job queue
func (pb *PreBuilder) worker(id int, jobs <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	
	for component := range jobs {
		start := time.Now()
		relativePath := strings.TrimPrefix(component, pb.routesDir+"/")
		
		// Update progress
		current := atomic.AddInt32(&pb.compiledCount, 1)
		progress := float64(current) / float64(pb.totalComponents) * 100
		
		if logging.IsDebugEnabled() {
			logging.Debug("Compiling component", 
				"worker", id,
				"progress", fmt.Sprintf("%.1f%%", progress),
				"current", current,
				"total", pb.totalComponents,
				"file", relativePath)
		} else {
			// For non-debug levels, show a simple progress indicator
			logging.Info("Compiling", "progress", fmt.Sprintf("%d/%d", current, pb.totalComponents), "file", relativePath)
		}
		
		// Compile the component
		if err := pb.compileComponent(component); err != nil {
			atomic.AddInt32(&pb.errorCount, 1)
			atomic.AddInt32(&pb.compiledCount, -1) // Correct the count
			pb.mu.Lock()
			pb.errors = append(pb.errors, fmt.Errorf("%s: %w", relativePath, err))
			pb.mu.Unlock()
			logging.Error("Compilation failed", 
				"file", relativePath, 
				"duration", fmt.Sprintf("%.2fs", time.Since(start).Seconds()), 
				"error", err)
		} else {
			logging.Debug("Compilation successful", 
				"file", relativePath, 
				"duration", fmt.Sprintf("%.2fs", time.Since(start).Seconds()))
		}
	}
}

// compileComponent pre-compiles a single Svelte component
func (pb *PreBuilder) compileComponent(path string) error {
	// Read the component file
	content, err := pb.fs.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	// Use the SvelteHandler's PrecompileComponent method to trigger compilation and caching
	// This will automatically cache the compiled result
	if err := pb.svelteHandler.PrecompileComponent(path, string(content)); err != nil {
		return fmt.Errorf("compilation failed: %w", err)
	}
	
	return nil
}

// printErrors prints all compilation errors
func (pb *PreBuilder) printErrors() {
	logging.Error("Compilation errors summary", "errorCount", len(pb.errors))
	for _, err := range pb.errors {
		logging.Error("Compilation error", "error", err)
	}
}

// GetStats returns pre-build statistics
func (pb *PreBuilder) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"totalComponents": pb.totalComponents,
		"compiledCount":   atomic.LoadInt32(&pb.compiledCount),
		"errorCount":      atomic.LoadInt32(&pb.errorCount),
		"parallelWorkers": pb.parallelWorkers,
		"elapsed":         time.Since(pb.startTime).String(),
	}
}