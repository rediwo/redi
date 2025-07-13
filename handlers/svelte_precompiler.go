package handlers

import (
	"fmt"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/rediwo/redi/cache"
	"github.com/rediwo/redi/filesystem"
)

// PrecompilePriority represents the priority of a file for precompilation
type PrecompilePriority struct {
	Path     string
	Priority float64
	Size     int64
}

// SveltePrecompiler handles background precompilation of Svelte components
type SveltePrecompiler struct {
	fs              filesystem.FileSystem
	cache           *cache.SvelteCache
	handler         *SvelteHandler
	routesDir       string
	
	// Precompilation queue
	queue           []PrecompilePriority
	queueMu         sync.Mutex
	
	// Worker management
	workers         int
	workerWG        sync.WaitGroup
	stopCh          chan struct{}
	
	// Statistics
	stats           PrecompileStats
	statsMu         sync.RWMutex
}

// PrecompileStats tracks precompilation statistics
type PrecompileStats struct {
	TotalFiles      int
	ProcessedFiles  int
	FailedFiles     int
	SkippedFiles    int
	TotalTime       time.Duration
	StartTime       time.Time
}

// NewSveltePrecompiler creates a new precompiler instance
func NewSveltePrecompiler(fs filesystem.FileSystem, cache *cache.SvelteCache, handler *SvelteHandler, routesDir string, workers int) *SveltePrecompiler {
	if workers <= 0 {
		workers = 2 // Default to 2 workers
	}
	
	return &SveltePrecompiler{
		fs:        fs,
		cache:     cache,
		handler:   handler,
		routesDir: routesDir,
		workers:   workers,
		stopCh:    make(chan struct{}),
		queue:     make([]PrecompilePriority, 0),
	}
}

// Start begins the precompilation process
func (sp *SveltePrecompiler) Start() error {
	sp.statsMu.Lock()
	sp.stats.StartTime = time.Now()
	sp.statsMu.Unlock()

	// Scan for Svelte files
	if err := sp.scanFiles(); err != nil {
		return fmt.Errorf("failed to scan files: %w", err)
	}

	// Sort queue by priority
	sp.sortQueue()

	// Start worker goroutines
	for i := 0; i < sp.workers; i++ {
		sp.workerWG.Add(1)
		go sp.worker(i)
	}

	return nil
}

// Stop stops the precompilation process
func (sp *SveltePrecompiler) Stop() {
	close(sp.stopCh)
	sp.workerWG.Wait()
	
	sp.statsMu.Lock()
	sp.stats.TotalTime = time.Since(sp.stats.StartTime)
	sp.statsMu.Unlock()
}

// scanFiles scans the routes directory for Svelte files
func (sp *SveltePrecompiler) scanFiles() error {
	sp.queueMu.Lock()
	defer sp.queueMu.Unlock()

	return sp.scanDirectory(sp.routesDir, 0)
}

// scanDirectory recursively scans a directory for Svelte files
func (sp *SveltePrecompiler) scanDirectory(dir string, depth int) error {
	entries, err := sp.fs.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		// Skip hidden files and directories
		if entry.Name()[0] == '.' || entry.Name()[0] == '_' {
			continue
		}

		path := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			// Recursively scan subdirectory
			if err := sp.scanDirectory(path, depth+1); err != nil {
				return err
			}
		} else if filepath.Ext(entry.Name()) == ".svelte" {
			// Calculate priority
			priority := sp.calculatePriority(path, depth)
			
			// Get file size
			info, err := sp.fs.Stat(path)
			if err != nil {
				continue
			}

			// Add to queue
			sp.queue = append(sp.queue, PrecompilePriority{
				Path:     path,
				Priority: priority,
				Size:     info.Size(),
			})
			
			sp.statsMu.Lock()
			sp.stats.TotalFiles++
			sp.statsMu.Unlock()
		}
	}

	return nil
}

// calculatePriority calculates the precompilation priority for a file
func (sp *SveltePrecompiler) calculatePriority(path string, depth int) float64 {
	priority := 10.0

	// Shared components get highest priority
	if filepath.Base(filepath.Dir(path)) == "_lib" {
		priority += 50.0
	}

	// Layout components get high priority
	if filepath.Base(filepath.Dir(path)) == "_layout" {
		priority += 30.0
	}

	// Index files get higher priority
	if filepath.Base(path) == "index.svelte" {
		priority += 20.0
	}

	// Files closer to root get higher priority
	priority -= float64(depth) * 2.0

	// Common component names get higher priority
	basename := filepath.Base(path)
	commonNames := []string{"App", "Layout", "Header", "Footer", "Nav", "Menu", "Button", "Card", "Modal", "Form"}
	for _, name := range commonNames {
		if basename == name+".svelte" {
			priority += 10.0
			break
		}
	}

	return priority
}

// sortQueue sorts the precompilation queue by priority
func (sp *SveltePrecompiler) sortQueue() {
	sp.queueMu.Lock()
	defer sp.queueMu.Unlock()

	sort.Slice(sp.queue, func(i, j int) bool {
		// Sort by priority (descending), then by size (ascending)
		if sp.queue[i].Priority != sp.queue[j].Priority {
			return sp.queue[i].Priority > sp.queue[j].Priority
		}
		return sp.queue[i].Size < sp.queue[j].Size
	})
}

// worker is the precompilation worker goroutine
func (sp *SveltePrecompiler) worker(id int) {
	defer sp.workerWG.Done()

	for {
		select {
		case <-sp.stopCh:
			return
		default:
			// Get next file from queue
			file, ok := sp.getNextFile()
			if !ok {
				// No more files, worker can exit
				return
			}

			// Precompile the file
			if err := sp.precompileFile(file.Path); err != nil {
				sp.statsMu.Lock()
				sp.stats.FailedFiles++
				sp.statsMu.Unlock()
				
				// Log error but continue
				fmt.Printf("Precompiler worker %d: failed to compile %s: %v\n", id, file.Path, err)
			} else {
				sp.statsMu.Lock()
				sp.stats.ProcessedFiles++
				sp.statsMu.Unlock()
			}
		}
	}
}

// getNextFile gets the next file from the queue
func (sp *SveltePrecompiler) getNextFile() (PrecompilePriority, bool) {
	sp.queueMu.Lock()
	defer sp.queueMu.Unlock()

	if len(sp.queue) == 0 {
		return PrecompilePriority{}, false
	}

	// Get first item (highest priority)
	file := sp.queue[0]
	sp.queue = sp.queue[1:]

	return file, true
}

// precompileFile precompiles a single Svelte file
func (sp *SveltePrecompiler) precompileFile(path string) error {
	// Read file content
	content, err := sp.fs.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Generate config hash for current configuration
	configHash := sp.handler.generateConfigHash()

	// Check if already cached
	if _, found := sp.cache.Get(path, content, configHash); found {
		sp.statsMu.Lock()
		sp.stats.SkippedFiles++
		sp.stats.ProcessedFiles--  // Adjust count since we're skipping
		sp.statsMu.Unlock()
		return nil
	}

	// Compile using the handler's compile function
	// This will automatically cache the result
	_, err = sp.handler.compileWithCache(path, string(content))
	if err != nil {
		return fmt.Errorf("compilation failed: %w", err)
	}

	return nil
}

// GetStats returns current precompilation statistics
func (sp *SveltePrecompiler) GetStats() PrecompileStats {
	sp.statsMu.RLock()
	defer sp.statsMu.RUnlock()
	
	stats := sp.stats
	if stats.StartTime.IsZero() {
		return stats
	}
	
	// Update total time if still running
	if stats.TotalTime == 0 {
		stats.TotalTime = time.Since(stats.StartTime)
	}
	
	return stats
}

// GetProgress returns the current progress percentage
func (sp *SveltePrecompiler) GetProgress() float64 {
	sp.statsMu.RLock()
	defer sp.statsMu.RUnlock()

	if sp.stats.TotalFiles == 0 {
		return 0
	}

	processed := sp.stats.ProcessedFiles + sp.stats.FailedFiles + sp.stats.SkippedFiles
	return float64(processed) / float64(sp.stats.TotalFiles) * 100
}

// IsComplete returns whether precompilation is complete
func (sp *SveltePrecompiler) IsComplete() bool {
	sp.statsMu.RLock()
	defer sp.statsMu.RUnlock()

	if sp.stats.TotalFiles == 0 {
		return true
	}

	processed := sp.stats.ProcessedFiles + sp.stats.FailedFiles + sp.stats.SkippedFiles
	return processed >= sp.stats.TotalFiles
}