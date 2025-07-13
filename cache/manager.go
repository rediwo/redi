package cache

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rediwo/redi/filesystem"
)

// CacheManager manages the cache system
type CacheManager struct {
	rootDir   string // Project root directory
	cacheDir  string // Cache directory (default: rootDir/.redi)
	fs        filesystem.FileSystem
	mu        sync.RWMutex
	index     *CacheIndex
	config    *CacheConfig
}

// CacheConfig represents cache configuration
type CacheConfig struct {
	RootDir       string        // Project root directory
	CacheDir      string        // Custom cache directory (optional)
	Enabled       bool          // Whether cache is enabled
	MaxSize       int64         // Maximum cache size in bytes
	TTL           time.Duration // Cache expiration time
	CompressCache bool          // Whether to compress cache
}

// CacheIndex represents the cache index
type CacheIndex struct {
	Version      string                    `json:"version"`
	RediVersion  string                    `json:"rediVersion"`
	Created      time.Time                 `json:"created"`
	LastUpdated  time.Time                 `json:"lastUpdated"`
	Entries      map[string]*CacheEntry    `json:"entries"`
	TotalSize    int64                     `json:"totalSize"`
	EntryCount   int                       `json:"entryCount"`
}

// CacheEntry represents a single cache entry
type CacheEntry struct {
	Path         string    `json:"path"`
	Hash         string    `json:"hash"`         // Content MD5
	ConfigHash   string    `json:"configHash"`   // Config hash
	Size         int64     `json:"size"`
	ModTime      time.Time `json:"modTime"`
	AccessTime   time.Time `json:"accessTime"`
	AccessCount  int       `json:"accessCount"`
	Priority     float64   `json:"priority"`     // Compilation priority
}

// CacheStats provides cache statistics
type CacheStats struct {
	TotalEntries   int           `json:"totalEntries"`
	TotalSize      int64         `json:"totalSize"`
	HitRate        float64       `json:"hitRate"`
	AvgCompileTime time.Duration `json:"avgCompileTime"`
	TopAccessed    []string      `json:"topAccessed"`
}

// NewCacheManager creates a new cache manager
func NewCacheManager(config *CacheConfig) *CacheManager {
	if config.CacheDir == "" {
		config.CacheDir = filepath.Join(config.RootDir, ".redi")
	}
	
	return &CacheManager{
		rootDir:  config.RootDir,
		cacheDir: config.CacheDir,
		config:   config,
		index: &CacheIndex{
			Version:  "1.0",
			Created:  time.Now(),
			Entries:  make(map[string]*CacheEntry),
		},
	}
}

// SetFileSystem sets the filesystem implementation
func (cm *CacheManager) SetFileSystem(fs filesystem.FileSystem) {
	cm.fs = fs
}

// Initialize initializes the cache system
func (cm *CacheManager) Initialize() error {
	if !cm.config.Enabled {
		return nil
	}

	// Create cache directory
	if err := os.MkdirAll(cm.cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Create subdirectories
	dirs := []string{
		filepath.Join(cm.cacheDir, "cache", "svelte", "compiled"),
		filepath.Join(cm.cacheDir, "cache", "svelte", "runtime"),
		filepath.Join(cm.cacheDir, "cache", "svelte", "deps"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create cache subdirectory %s: %w", dir, err)
		}
	}

	// Load existing index
	if err := cm.loadIndex(); err != nil {
		// If index doesn't exist or is corrupted, create new one
		cm.index = &CacheIndex{
			Version:     "1.0",
			RediVersion: cm.getRediVersion(),
			Created:     time.Now(),
			Entries:     make(map[string]*CacheEntry),
		}
		return cm.saveIndex()
	}

	return nil
}

// loadIndex loads the cache index from disk
func (cm *CacheManager) loadIndex() error {
	indexPath := filepath.Join(cm.cacheDir, "cache", "metadata.json")
	
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return err
	}

	var index CacheIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return err
	}

	cm.index = &index
	return nil
}

// saveIndex saves the cache index to disk
func (cm *CacheManager) saveIndex() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.index.LastUpdated = time.Now()
	
	data, err := json.MarshalIndent(cm.index, "", "  ")
	if err != nil {
		return err
	}

	indexPath := filepath.Join(cm.cacheDir, "cache", "metadata.json")
	return os.WriteFile(indexPath, data, 0644)
}

// GetEntry retrieves a cache entry
func (cm *CacheManager) GetEntry(key string) (*CacheEntry, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	entry, exists := cm.index.Entries[key]
	if exists {
		entry.AccessTime = time.Now()
		entry.AccessCount++
	}
	return entry, exists
}

// SetEntry sets a cache entry
func (cm *CacheManager) SetEntry(key string, entry *CacheEntry) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.index.Entries[key] = entry
	cm.index.EntryCount = len(cm.index.Entries)
	
	// Update total size
	cm.index.TotalSize = 0
	for _, e := range cm.index.Entries {
		cm.index.TotalSize += e.Size
	}

	// Check if we need to evict entries
	if cm.config.MaxSize > 0 && cm.index.TotalSize > cm.config.MaxSize {
		cm.evictLRU()
	}

	return nil
}

// evictLRU evicts least recently used entries
func (cm *CacheManager) evictLRU() {
	// Simple LRU implementation
	// TODO: Implement proper LRU eviction
}

// Clear clears all cache
func (cm *CacheManager) Clear() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Remove cache directory
	if err := os.RemoveAll(cm.cacheDir); err != nil {
		return err
	}

	// Reinitialize
	return cm.Initialize()
}

// GetStats returns cache statistics
func (cm *CacheManager) GetStats() *CacheStats {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Calculate hit rate and other stats
	// TODO: Implement proper statistics tracking
	
	return &CacheStats{
		TotalEntries: cm.index.EntryCount,
		TotalSize:    cm.index.TotalSize,
		HitRate:      0.0, // TODO: Track hits/misses
		TopAccessed:  []string{}, // TODO: Track top accessed
	}
}

// generateCacheKey generates a cache key from path and content
func (cm *CacheManager) GenerateCacheKey(path string, content []byte, configHash string) string {
	h := md5.New()
	h.Write([]byte(path))
	h.Write(content)
	h.Write([]byte(configHash))
	h.Write([]byte(cm.getRediVersion()))
	return hex.EncodeToString(h.Sum(nil))
}

// getRediVersion gets the current Redi version
func (cm *CacheManager) getRediVersion() string {
	// TODO: Get from build info
	return "1.0.0"
}

// GetCachePath returns the full path for a cache key
func (cm *CacheManager) GetCachePath(key string) string {
	// Use first 2 chars as subdirectory for better file distribution
	if len(key) >= 2 {
		return filepath.Join(cm.cacheDir, "cache", "svelte", "compiled", key[:2], key+".json")
	}
	return filepath.Join(cm.cacheDir, "cache", "svelte", "compiled", key+".json")
}