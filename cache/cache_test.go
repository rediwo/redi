package cache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rediwo/redi/filesystem"
)

func TestCacheManager(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "redi-cache-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create cache configuration
	config := &CacheConfig{
		RootDir:       tmpDir,
		Enabled:       true,
		MaxSize:       100 * 1024 * 1024, // 100MB
		CompressCache: false,
	}

	// Create cache manager
	manager := NewCacheManager(config)
	manager.SetFileSystem(filesystem.NewOSFileSystem(tmpDir))

	// Initialize cache
	if err := manager.Initialize(); err != nil {
		t.Fatalf("Failed to initialize cache: %v", err)
	}

	// Check that cache directories were created
	cacheDirs := []string{
		filepath.Join(tmpDir, ".redi", "cache", "svelte", "compiled"),
		filepath.Join(tmpDir, ".redi", "cache", "svelte", "runtime"),
		filepath.Join(tmpDir, ".redi", "cache", "svelte", "deps"),
	}

	for _, dir := range cacheDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Cache directory not created: %s", dir)
		}
	}

	// Test cache key generation
	content := []byte("test content")
	configHash := "test-config"
	key := manager.GenerateCacheKey("test.svelte", content, configHash)
	if key == "" {
		t.Error("Generated cache key is empty")
	}

	// Test cache entry management
	entry := &CacheEntry{
		Path:       "test.svelte",
		Hash:       key,
		ConfigHash: configHash,
		Size:       100,
	}

	if err := manager.SetEntry(key, entry); err != nil {
		t.Fatalf("Failed to set cache entry: %v", err)
	}

	// Retrieve the entry
	retrieved, exists := manager.GetEntry(key)
	if !exists {
		t.Error("Cache entry not found")
	}
	if retrieved.Path != entry.Path {
		t.Errorf("Retrieved entry path mismatch: got %s, want %s", retrieved.Path, entry.Path)
	}

	// Test cache stats
	stats := manager.GetStats()
	if stats.TotalEntries != 1 {
		t.Errorf("Expected 1 total entry, got %d", stats.TotalEntries)
	}
}

func TestSvelteCache(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "redi-svelte-cache-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test component file
	componentPath := filepath.Join(tmpDir, "test.svelte")
	componentContent := []byte(`<script>let count = 0;</script><button>{count}</button>`)
	if err := os.WriteFile(componentPath, componentContent, 0644); err != nil {
		t.Fatal(err)
	}

	// Create cache configuration
	config := &CacheConfig{
		RootDir: tmpDir,
		Enabled: true,
	}

	// Create cache manager and Svelte cache
	manager := NewCacheManager(config)
	fs := filesystem.NewOSFileSystem(tmpDir)
	manager.SetFileSystem(fs)
	
	if err := manager.Initialize(); err != nil {
		t.Fatal(err)
	}

	svelteCache := NewSvelteCache(manager, fs)

	// Test cache miss
	configHash := "test-config"
	cached, found := svelteCache.Get("test.svelte", componentContent, configHash)
	if found {
		t.Error("Expected cache miss, but found entry")
	}
	if cached != nil {
		t.Error("Expected nil cached entry on miss")
	}

	// Test setting cache
	entry := &SvelteCacheEntry{
		Path:       "test.svelte",
		JavaScript: "compiled js",
		CSS:        "compiled css",
		ClassName:  "Test",
	}

	if err := svelteCache.Set("test.svelte", componentContent, configHash, entry); err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// Test cache hit
	cached, found = svelteCache.Get("test.svelte", componentContent, configHash)
	if !found {
		t.Error("Expected cache hit, but not found")
	}
	if cached == nil {
		t.Error("Expected cached entry, got nil")
	}
	if cached.JavaScript != entry.JavaScript {
		t.Errorf("Cached JS mismatch: got %s, want %s", cached.JavaScript, entry.JavaScript)
	}

	// Test stats
	stats := svelteCache.GetStats()
	if stats["hits"].(int64) != 1 {
		t.Errorf("Expected 1 hit, got %v", stats["hits"])
	}
	if stats["misses"].(int64) != 1 {
		t.Errorf("Expected 1 miss, got %v", stats["misses"])
	}
}