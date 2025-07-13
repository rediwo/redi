package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/rediwo/redi/filesystem"
)

// SvelteCache manages Svelte component compilation caching
type SvelteCache struct {
	manager         *CacheManager
	fs              filesystem.FileSystem
	mu              sync.RWMutex
	hits            int64
	misses          int64
	totalCompileTime time.Duration
	compileCount    int64
}

// SvelteCacheEntry represents a cached Svelte component
type SvelteCacheEntry struct {
	// Metadata
	Path         string    `json:"path"`
	Hash         string    `json:"hash"`
	ConfigHash   string    `json:"configHash"`
	CompileTime  time.Duration `json:"compileTime"`
	Timestamp    time.Time `json:"timestamp"`
	
	// Compiled output
	JavaScript   string   `json:"javascript"`
	CSS          string   `json:"css"`
	ClassName    string   `json:"className"`
	
	// Dependencies
	Dependencies []string `json:"dependencies"`
	Imports      map[string]string `json:"imports"`
	
	// Runtime flags
	HasVimeshStyle bool `json:"hasVimeshStyle"`
	IsMinified     bool `json:"isMinified"`
	
	// Runtime-only fields (not persisted)
	WasFromCache bool `json:"-"` // Indicates if this entry was loaded from cache
}

// NewSvelteCache creates a new Svelte cache instance
func NewSvelteCache(manager *CacheManager, fs filesystem.FileSystem) *SvelteCache {
	return &SvelteCache{
		manager: manager,
		fs:      fs,
	}
}

// Get retrieves a cached component
func (sc *SvelteCache) Get(path string, content []byte, configHash string) (*SvelteCacheEntry, bool) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	// Generate cache key
	key := sc.manager.GenerateCacheKey(path, content, configHash)
	
	// Check if entry exists in index
	entry, exists := sc.manager.GetEntry(key)
	if !exists {
		sc.misses++
		return nil, false
	}

	// Load cached data
	cachePath := sc.manager.GetCachePath(key)
	data, err := os.ReadFile(cachePath)
	if err != nil {
		sc.misses++
		return nil, false
	}

	var cacheEntry SvelteCacheEntry
	if err := json.Unmarshal(data, &cacheEntry); err != nil {
		sc.misses++
		return nil, false
	}

	// Verify file hasn't changed
	stat, err := sc.fs.Stat(path)
	if err != nil || stat.ModTime().After(entry.ModTime) {
		sc.misses++
		return nil, false
	}

	// Mark this entry as loaded from cache
	cacheEntry.WasFromCache = true
	
	sc.hits++
	return &cacheEntry, true
}

// Set stores a compiled component in cache
func (sc *SvelteCache) Set(path string, content []byte, configHash string, compiled *SvelteCacheEntry) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	// Update compile statistics
	sc.totalCompileTime += compiled.CompileTime
	sc.compileCount++

	// Generate cache key
	key := sc.manager.GenerateCacheKey(path, content, configHash)
	
	// Serialize cache entry
	data, err := json.MarshalIndent(compiled, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize cache entry: %w", err)
	}

	// Create cache directory if needed
	cachePath := sc.manager.GetCachePath(key)
	cacheDir := filepath.Dir(cachePath)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Write cache file
	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	// Get file info
	stat, err := sc.fs.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	// Update cache index
	entry := &CacheEntry{
		Path:        path,
		Hash:        key,
		ConfigHash:  configHash,
		Size:        int64(len(data)),
		ModTime:     stat.ModTime(),
		AccessTime:  time.Now(),
		AccessCount: 1,
		Priority:    sc.calculatePriority(path),
	}

	if err := sc.manager.SetEntry(key, entry); err != nil {
		return fmt.Errorf("failed to update cache index: %w", err)
	}

	// Save index
	return sc.manager.saveIndex()
}

// CompileWithCache compiles a Svelte component with caching
func (sc *SvelteCache) CompileWithCache(
	path string,
	content []byte,
	configHash string,
	compileFunc func() (*SvelteCacheEntry, error),
) (*SvelteCacheEntry, error) {
	// Try to get from cache first
	if cached, found := sc.Get(path, content, configHash); found {
		return cached, nil
	}

	// Compile the component
	start := time.Now()
	compiled, err := compileFunc()
	if err != nil {
		return nil, err
	}
	compiled.CompileTime = time.Since(start)
	compiled.WasFromCache = false // This is a fresh compilation

	// Store in cache
	if err := sc.Set(path, content, configHash, compiled); err != nil {
		// Log error but continue - cache failure shouldn't break compilation
		fmt.Printf("Warning: failed to cache component %s: %v\n", path, err)
	}

	return compiled, nil
}

// GetStats returns cache statistics
func (sc *SvelteCache) GetStats() map[string]interface{} {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	hitRate := float64(0)
	if total := sc.hits + sc.misses; total > 0 {
		hitRate = float64(sc.hits) / float64(total) * 100
	}

	avgCompileTime := time.Duration(0)
	if sc.compileCount > 0 {
		avgCompileTime = sc.totalCompileTime / time.Duration(sc.compileCount)
	}

	return map[string]interface{}{
		"hits":            sc.hits,
		"misses":          sc.misses,
		"hitRate":         hitRate,
		"avgCompileTime":  avgCompileTime.String(),
		"totalCompiles":   sc.compileCount,
	}
}

// calculatePriority calculates cache priority for a component
func (sc *SvelteCache) calculatePriority(path string) float64 {
	// Higher priority for:
	// - Components in _lib directory (shared components)
	// - Components closer to root
	// - Frequently accessed components (will be updated on access)
	
	priority := 1.0
	
	// Shared components get higher priority
	if filepath.Base(filepath.Dir(path)) == "_lib" {
		priority += 2.0
	}
	
	// Root level components get higher priority
	depth := len(strings.Split(path, string(filepath.Separator)))
	priority += 1.0 / float64(depth)
	
	return priority
}

// InvalidatePath invalidates cache for a specific path
func (sc *SvelteCache) InvalidatePath(path string) error {
	// TODO: Implement path-based invalidation
	// This would need to track reverse mapping from paths to cache keys
	return nil
}

// InvalidateDependents invalidates cache for components that depend on a path
func (sc *SvelteCache) InvalidateDependents(path string) error {
	// TODO: Implement dependency-based invalidation
	// This would need to maintain a dependency graph
	return nil
}

// Precompile attempts to precompile a component and cache it
func (sc *SvelteCache) Precompile(path string, content []byte, configHash string, vm *goja.Runtime) error {
	// Check if already cached
	if _, found := sc.Get(path, content, configHash); found {
		return nil
	}

	// TODO: Implement actual precompilation using the VM
	// This would involve calling the Svelte compiler
	
	return nil
}