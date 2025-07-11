package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/rediwo/redi/filesystem"
)

func TestSvelteHandler_ServeAsyncComponent(t *testing.T) {
	// Create filesystem with test components
	fs := filesystem.NewMemoryFileSystem()

	// Create a simple test component
	testComponent := `
<script>
    export let message = 'Hello World';
</script>
<div class="test-component">
    <h1>{message}</h1>
</div>
<style>
    .test-component { padding: 20px; }
</style>
`

	err := fs.WriteFile("routes/TestComponent.svelte", []byte(testComponent))
	if err != nil {
		t.Fatalf("Failed to create test component: %v", err)
	}

	// Create handler with async loading enabled
	config := DefaultSvelteConfig()
	config.EnableAsyncLoading = true
	handler := NewSvelteHandlerWithConfig(fs, config)

	// Initialize compiler
	err = handler.initializeCompiler()
	if err != nil {
		t.Fatalf("Failed to initialize compiler: %v", err)
	}

	// Create router and register routes
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Test successful component loading
	req := httptest.NewRequest("GET", "/svelte/components/TestComponent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check response content type
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}

	// Parse response
	var response AsyncComponentResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Verify response
	if !response.Success {
		t.Errorf("Expected success=true, got %v", response.Success)
	}

	if response.Component != "TestComponent" {
		t.Errorf("Expected component=TestComponent, got %s", response.Component)
	}

	if response.ClassName != "TestComponent" {
		t.Errorf("Expected className=TestComponent, got %s", response.ClassName)
	}

	if response.JS == "" {
		t.Error("Expected JS code to be non-empty")
	}

	if response.CSS == "" {
		t.Error("Expected CSS code to be non-empty")
	}
}

func TestSvelteHandler_ServeAsyncComponent_NotFound(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()

	config := DefaultSvelteConfig()
	config.EnableAsyncLoading = true
	handler := NewSvelteHandlerWithConfig(fs, config)

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Test non-existent component
	req := httptest.NewRequest("GET", "/svelte/components/NonExistentComponent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	var response AsyncComponentResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Success {
		t.Error("Expected success=false for non-existent component")
	}

	if response.Error == "" {
		t.Error("Expected error message for non-existent component")
	}
}

func TestSvelteHandler_ServeAsyncComponent_WithDependencies(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()

	// Create base component
	baseComponent := `
<script>
    export let text = 'Base Component';
</script>
<div class="base">{text}</div>
<style>
    .base { color: blue; }
</style>
`

	// Create component with dependency
	mainComponent := `
<script>
    import Base from './BaseComponent.svelte';
    export let title = 'Main Component';
</script>
<div class="main">
    <h1>{title}</h1>
    <Base text="From Main" />
</div>
<style>
    .main { padding: 10px; }
</style>
`

	err := fs.WriteFile("routes/BaseComponent.svelte", []byte(baseComponent))
	if err != nil {
		t.Fatalf("Failed to create base component: %v", err)
	}

	err = fs.WriteFile("routes/MainComponent.svelte", []byte(mainComponent))
	if err != nil {
		t.Fatalf("Failed to create main component: %v", err)
	}

	config := DefaultSvelteConfig()
	config.EnableAsyncLoading = true
	handler := NewSvelteHandlerWithConfig(fs, config)

	// Initialize compiler
	err = handler.initializeCompiler()
	if err != nil {
		t.Fatalf("Failed to initialize compiler: %v", err)
	}

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Test component with dependencies
	req := httptest.NewRequest("GET", "/svelte/components/MainComponent?include_deps=true", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response AsyncComponentResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success=true, got %v", response.Success)
	}

	if len(response.Dependencies) == 0 {
		t.Error("Expected dependencies to be included")
	}

	// Check dependency details
	if len(response.Dependencies) > 0 {
		dep := response.Dependencies[0]
		if dep.Component != "BaseComponent" {
			t.Errorf("Expected dependency component=BaseComponent, got %s", dep.Component)
		}

		if dep.JS == "" {
			t.Error("Expected dependency JS to be non-empty")
		}
	}
}

func TestSvelteHandler_ServeAsyncLibrary(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()

	config := DefaultSvelteConfig()
	config.EnableAsyncLoading = true
	handler := NewSvelteHandlerWithConfig(fs, config)

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Test async library serving
	req := httptest.NewRequest("GET", config.AsyncLibraryPath, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check content type
	if w.Header().Get("Content-Type") != "application/javascript; charset=utf-8" {
		t.Errorf("Expected Content-Type application/javascript, got %s", w.Header().Get("Content-Type"))
	}

	// Check that response contains async library code
	body := w.Body.String()
	if body == "" {
		t.Error("Expected async library code to be non-empty")
	}

	// Check for key async library components
	if !containsString(body, "SvelteAsync") {
		t.Error("Expected async library to contain SvelteAsync object")
	}

	if !containsString(body, "loadComponent") {
		t.Error("Expected async library to contain loadComponent function")
	}

	if !containsString(body, "lazy") {
		t.Error("Expected async library to contain lazy function")
	}
}

func TestSvelteHandler_AsyncComponentCaching(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()

	testComponent := `
<script>
    export let message = 'Cached Component';
</script>
<div>{message}</div>
`

	err := fs.WriteFile("routes/CachedComponent.svelte", []byte(testComponent))
	if err != nil {
		t.Fatalf("Failed to create test component: %v", err)
	}

	config := DefaultSvelteConfig()
	config.EnableAsyncLoading = true
	config.ComponentCacheDuration = 1 * time.Hour
	handler := NewSvelteHandlerWithConfig(fs, config)

	// Initialize compiler
	err = handler.initializeCompiler()
	if err != nil {
		t.Fatalf("Failed to initialize compiler: %v", err)
	}

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// First request
	req1 := httptest.NewRequest("GET", "/svelte/components/CachedComponent", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w1.Code)
	}

	etag1 := w1.Header().Get("ETag")
	if etag1 == "" {
		t.Error("Expected ETag header to be set")
	}

	// Second request with same ETag should return 304
	req2 := httptest.NewRequest("GET", "/svelte/components/CachedComponent", nil)
	req2.Header.Set("If-None-Match", etag1)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusNotModified {
		t.Errorf("Expected status 304, got %d", w2.Code)
	}
}

func TestSvelteHandler_AsyncRouteRegistration(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()

	// Test with async loading disabled
	config1 := DefaultSvelteConfig()
	config1.EnableAsyncLoading = false
	handler1 := NewSvelteHandlerWithConfig(fs, config1)

	router1 := mux.NewRouter()
	handler1.RegisterRoutes(router1)

	// Should not register async routes
	req1 := httptest.NewRequest("GET", "/svelte/components/test", nil)
	w1 := httptest.NewRecorder()
	router1.ServeHTTP(w1, req1)

	if w1.Code != http.StatusNotFound {
		t.Errorf("Expected 404 when async loading disabled, got %d", w1.Code)
	}

	// Test with async loading enabled
	config2 := DefaultSvelteConfig()
	config2.EnableAsyncLoading = true
	handler2 := NewSvelteHandlerWithConfig(fs, config2)

	router2 := mux.NewRouter()
	handler2.RegisterRoutes(router2)

	// Should register async routes (404 for non-existent component is expected)
	req2 := httptest.NewRequest("GET", "/svelte/components/test", nil)
	w2 := httptest.NewRecorder()
	router2.ServeHTTP(w2, req2)

	// Should get 404 for non-existent component (not route not found)
	if w2.Code != http.StatusNotFound {
		t.Errorf("Expected 404 for non-existent component, got %d", w2.Code)
	}

	// But should be JSON response, not HTML
	if w2.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected JSON response for async component, got %s", w2.Header().Get("Content-Type"))
	}
}

// Helper function to check if string contains substring
func containsString(str, substr string) bool {
	return len(str) >= len(substr) && findSubstring(str, substr) != -1
}

func findSubstring(str, substr string) int {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
