package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/rediwo/redi/filesystem"
	"github.com/rediwo/redi/utils"
)

func TestSvelteHandler_CompilerLoad(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	handler := NewSvelteHandler(fs)

	// Test compiler initialization
	err := handler.initializeCompiler()
	if err != nil {
		t.Fatalf("Failed to initialize Svelte compiler: %v", err)
	}

	// Verify compiler is loaded
	if !handler.initialized {
		t.Error("Compiler should be initialized")
	}

	// Verify compile function exists
	if handler.compileFunc == nil {
		t.Error("Compile function should be available")
	}
}

func TestSvelteHandler_CompileSimpleComponent(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	handler := NewSvelteHandler(fs)

	// Simple Svelte component
	svelteCode := `
<script>
    export let name = 'World';
</script>

<h1>Hello {name}!</h1>

<style>
    h1 {
        color: #ff3e00;
        text-transform: uppercase;
        font-size: 4em;
        font-weight: 100;
    }
</style>
`

	result, err := handler.compileSvelte(svelteCode, "Hello.svelte")
	if err != nil {
		t.Fatalf("Failed to compile Svelte component: %v", err)
	}

	// Check JavaScript output
	if result.JS == "" {
		t.Error("Expected JavaScript output, got empty string")
	}

	// Check CSS output
	if result.CSS == "" {
		t.Error("Expected CSS output, got empty string")
	}

	// Verify component class is created (with capital H)
	if !strings.Contains(result.JS, "class Hello") {
		t.Error("Expected compiled JS to contain component class")
	}

	// Verify CSS contains our styles
	if !strings.Contains(result.CSS, "#ff3e00") {
		t.Error("Expected CSS to contain color style")
	}
}

func TestSvelteHandler_CompileReactiveComponent(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	handler := NewSvelteHandler(fs)

	// Svelte component with reactive statements
	svelteCode := `
<script>
    let count = 0;
    $: doubled = count * 2;

    function increment() {
        count += 1;
    }
</script>

<button on:click={increment}>
    Clicked {count} {count === 1 ? 'time' : 'times'}
</button>

<p>{count} doubled is {doubled}</p>
`

	result, err := handler.compileSvelte(svelteCode, "Counter.svelte")
	if err != nil {
		t.Fatalf("Failed to compile reactive component: %v", err)
	}

	// Check for reactive statement handling
	if !strings.Contains(result.JS, "count") {
		t.Error("Expected compiled JS to contain count variable")
	}

	// Check for event handler
	if !strings.Contains(result.JS, "click") {
		t.Error("Expected compiled JS to contain click handler")
	}
}

func TestSvelteHandler_CompileError(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	handler := NewSvelteHandler(fs)

	// Invalid Svelte code
	invalidCode := `
<script>
    let count = // syntax error
</script>

<div>{count}</div>
`

	_, err := handler.compileSvelte(invalidCode, "Invalid.svelte")
	if err == nil {
		t.Error("Expected compilation error for invalid syntax")
	}
}

func TestSvelteHandler_HTTPRequest(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	
	// Create a test Svelte file
	svelteContent := `
<script>
    let message = 'Hello from Svelte!';
</script>

<h1>{message}</h1>

<style>
    h1 { color: blue; }
</style>
`
	fs.WriteFile("routes/test.svelte", []byte(svelteContent))

	handler := NewSvelteHandler(fs)
	route := Route{
		Path:     "/test",
		FilePath: "routes/test.svelte",
		FileType: "svelte",
	}

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Handle request
	handlerFunc := handler.Handle(route)
	handlerFunc(w, req)

	// Check response
	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body := w.Body.String()

	// Verify HTML structure
	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("Expected HTML doctype")
	}

	// Verify component is mounted
	if !strings.Contains(body, "new Test({") {
		t.Errorf("Expected component mounting code 'new Test({', got body: %s", body)
	}

	// Verify CSS is included
	if !strings.Contains(body, "color: blue") && !strings.Contains(body, "color:blue") {
		t.Errorf("Expected CSS to be included. Body: %s", body)
	}
}

func TestSvelteHandler_FileSize(t *testing.T) {
	// Test embedded compiler size
	compilerSize := len(svelteCompilerJS)
	compilerSizeMB := float64(compilerSize) / (1024 * 1024)
	
	t.Logf("Embedded Svelte compiler size: %.2f MB", compilerSizeMB)
	
	// Verify it's within expected range (1-2 MB)
	if compilerSizeMB < 1.0 || compilerSizeMB > 2.0 {
		t.Errorf("Unexpected compiler size: %.2f MB, expected 1-2 MB", compilerSizeMB)
	}
}

func BenchmarkSvelteCompilation(b *testing.B) {
	fs := filesystem.NewMemoryFileSystem()
	handler := NewSvelteHandler(fs)

	svelteCode := `
<script>
    export let name = 'World';
    let count = 0;
    
    function increment() {
        count += 1;
    }
</script>

<h1>Hello {name}!</h1>
<button on:click={increment}>Count: {count}</button>

<style>
    h1 { color: #ff3e00; }
    button { padding: 10px; }
</style>
`

	// Initialize compiler once
	handler.initializeCompiler()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := handler.compileSvelte(svelteCode, "Benchmark.svelte")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Tests from svelte_vimesh_test.go

func TestSvelteHandler_VimeshStyleIntegration(t *testing.T) {
	// Create test filesystem with a Svelte component using Tailwind classes
	fs := filesystem.NewMemoryFileSystem()
	
	svelteComponent := `
<script>
	let count = 0;
	
	function increment() {
		count += 1;
	}
</script>

<div class="bg-blue-500 text-white p-4 rounded-lg shadow-md max-w-md mx-auto mt-8">
	<h1 class="text-2xl font-bold mb-2">Vimesh Style Test</h1>
	<p class="text-lg mb-4">Count: {count}</p>
	<button 
		class="bg-white text-blue-500 px-4 py-2 rounded hover:bg-gray-100 transition-colors"
		on:click={increment}>
		Increment
	</button>
</div>

<style>
	/* Component-specific styles */
	div {
		user-select: none;
	}
</style>
`
	
	// Write the component to the filesystem
	err := fs.WriteFile("routes/test.svelte", []byte(svelteComponent))
	if err != nil {
		t.Fatalf("Failed to write test component: %v", err)
	}
	
	// Create config with Vimesh Style enabled
	config := DefaultSvelteConfig()
	config.VimeshStyle = &utils.VimeshStyleConfig{Enable: true}
	
	// Create router
	router := mux.NewRouter()
	
	// Create handler with router
	handler := NewSvelteHandlerWithRouter(fs, config, router)
	
	// Create route for the component
	route := Route{
		Path:     "/test",
		FilePath: "routes/test.svelte",
		FileType: "svelte",
	}
	
	// Register the component handler
	router.HandleFunc(route.Path, handler.Handle(route)).Methods("GET")
	
	// Test 1: Vimesh Style JavaScript route
	t.Run("VimeshStyleJSRoute", func(t *testing.T) {
		req := httptest.NewRequest("GET", config.VimeshStylePath, nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		
		contentType := w.Header().Get("Content-Type")
		if !strings.Contains(contentType, "application/javascript") {
			t.Errorf("Expected JavaScript content type, got %s", contentType)
		}
		
		body := w.Body.String()
		if len(body) == 0 {
			t.Error("Vimesh Style JS should not be empty")
		}
		
		// Check for Vimesh Style markers
		if !strings.Contains(body, "$vs") && !strings.Contains(body, "vimesh") {
			t.Error("Response doesn't contain Vimesh Style code")
		}
	})
	
	// Test 2: Svelte Runtime route
	t.Run("SvelteRuntimeRoute", func(t *testing.T) {
		req := httptest.NewRequest("GET", config.RuntimePath, nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		
		contentType := w.Header().Get("Content-Type")
		if !strings.Contains(contentType, "application/javascript") {
			t.Errorf("Expected JavaScript content type, got %s", contentType)
		}
		
		body := w.Body.String()
		if len(body) == 0 {
			t.Error("Svelte runtime should not be empty")
		}
	})
	
	// Test 3: Component with Vimesh Style
	t.Run("ComponentWithVimeshStyle", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		
		body := w.Body.String()
		
		// Check for Vimesh Style CSS
		if !strings.Contains(body, `<style id="vimesh-styles">`) {
			t.Error("Component should include Vimesh Style CSS")
		}
		
		// Check for Vimesh Style script tag
		if !strings.Contains(body, config.VimeshStylePath) {
			t.Error("Component should include Vimesh Style script tag")
		}
		
		// Check that some CSS was extracted
		if !strings.Contains(body, "bg-blue-500") || !strings.Contains(body, "text-white") {
			// The CSS might be transformed, so just check that the style tag has content
			start := strings.Index(body, `<style id="vimesh-styles">`)
			end := strings.Index(body, `</style>`)
			if start >= 0 && end > start {
				cssContent := body[start+26 : end]
				if len(strings.TrimSpace(cssContent)) == 0 {
					t.Error("Vimesh Style CSS should not be empty")
				}
			}
		}
	})
	
	// Test 4: Caching headers
	t.Run("CachingHeaders", func(t *testing.T) {
		req := httptest.NewRequest("GET", config.VimeshStylePath, nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		// Check ETag
		etag := w.Header().Get("ETag")
		if etag == "" {
			t.Error("Expected ETag header")
		}
		
		// Check Cache-Control
		cacheControl := w.Header().Get("Cache-Control")
		if !strings.Contains(cacheControl, "max-age=") {
			t.Error("Expected Cache-Control header with max-age")
		}
		
		// Test conditional request
		req2 := httptest.NewRequest("GET", config.VimeshStylePath, nil)
		req2.Header.Set("If-None-Match", etag)
		w2 := httptest.NewRecorder()
		
		router.ServeHTTP(w2, req2)
		
		if w2.Code != http.StatusNotModified {
			t.Errorf("Expected 304 Not Modified for matching ETag, got %d", w2.Code)
		}
	})
}

func TestSvelteHandler_VimeshStyleDisabled(t *testing.T) {
	// Create test filesystem
	fs := filesystem.NewMemoryFileSystem()
	
	// Write a simple component
	err := fs.WriteFile("routes/test.svelte", []byte(`<div class="bg-red-500">Test</div>`))
	if err != nil {
		t.Fatalf("Failed to write test component: %v", err)
	}
	
	// Create config with Vimesh Style disabled
	config := DefaultSvelteConfig()
	config.VimeshStyle = &utils.VimeshStyleConfig{Enable: false}
	
	// Create router
	router := mux.NewRouter()
	
	// Create handler with router
	handler := NewSvelteHandlerWithRouter(fs, config, router)
	
	// Create route
	route := Route{
		Path:     "/test",
		FilePath: "routes/test.svelte",
		FileType: "svelte",
	}
	
	// Register the component handler
	router.HandleFunc(route.Path, handler.Handle(route)).Methods("GET")
	
	// Test that Vimesh Style route is not registered
	t.Run("VimeshStyleRouteNotRegistered", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/svelte-vimesh-style.js", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		// Should get 404 since route is not registered
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected 404 for disabled Vimesh Style, got %d", w.Code)
		}
	})
	
	// Test that component doesn't include Vimesh Style
	t.Run("ComponentWithoutVimeshStyle", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		
		body := w.Body.String()
		
		// Should not include Vimesh Style CSS
		if strings.Contains(body, `<style id="vimesh-styles">`) {
			t.Error("Component should not include Vimesh Style CSS when disabled")
		}
		
		// Should not include Vimesh Style script
		if strings.Contains(body, "/svelte-vimesh-style.js") {
			t.Error("Component should not include Vimesh Style script when disabled")
		}
	})
}

// Tests from svelte_transform_test.go

func TestTransformToIIFE(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	sh := NewSvelteHandler(fs)
	
	tests := []struct {
		name           string
		jsCode         string
		componentName  string
		shouldContain  []string
		shouldNotContain []string
	}{
		{
			name: "Remove export default with semicolon",
			jsCode: `
import { SvelteComponent } from 'svelte';
class MyComponent extends SvelteComponent {
	constructor(options) {
		super();
	}
}
export default MyComponent;`,
			componentName: "MyComponent",
			shouldContain: []string{"class MyComponent extends SvelteComponent"},
			shouldNotContain: []string{"import", "export default"},
		},
		{
			name: "Remove export default without semicolon",
			jsCode: `
import { SvelteComponent } from 'svelte';
class MyComponent extends SvelteComponent {
	constructor(options) {
		super();
	}
}
export default MyComponent`,
			componentName: "MyComponent",
			shouldContain: []string{"class MyComponent extends SvelteComponent"},
			shouldNotContain: []string{"import", "export default"},
		},
		{
			name: "Real Svelte output format",
			jsCode: `/* generated by Svelte */
import {
	SvelteComponent,
	init,
	safe_not_equal
} from "svelte/internal";

function create_fragment(ctx) {
	// component logic
}

class Vimesh_test extends SvelteComponent {
	constructor(options) {
		super(),
		init(this, options, instance, create_fragment, safe_not_equal, {})
	}
}
export default Vimesh_test`,
			componentName: "Vimesh_test",
			shouldContain: []string{"class Vimesh_test extends SvelteComponent"},
			shouldNotContain: []string{"import", "export default"},
		},
		{
			name: "Multiple imports",
			jsCode: `import { SvelteComponent } from "svelte";
import { onMount } from "svelte";
import utils from "./utils";
class Component extends SvelteComponent {}
export default Component;`,
			componentName: "Component",
			shouldContain: []string{"class Component extends SvelteComponent"},
			shouldNotContain: []string{"import", "export default"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imports := make(map[string]string)
			result := sh.transformToIIFE(tt.jsCode, tt.componentName, imports, "routes/test.svelte")
			
			// Check that required content is present
			for _, expected := range tt.shouldContain {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain '%s', but it didn't.\nResult:\n%s", expected, result)
				}
			}
			
			// Check that unwanted content is removed
			for _, unexpected := range tt.shouldNotContain {
				if strings.Contains(result, unexpected) {
					t.Errorf("Expected result NOT to contain '%s', but it did.\nResult:\n%s", unexpected, result)
				}
			}
			
			// Ensure no trailing/leading whitespace issues
			trimmed := strings.TrimSpace(result)
			if trimmed != result {
				t.Logf("Warning: Result has leading/trailing whitespace")
			}
		})
	}
}

func TestToComponentClassName(t *testing.T) {
	sh := &SvelteHandler{}
	
	tests := []struct {
		input    string
		expected string
	}{
		{"hello-world", "Hello_world"},
		{"my-component", "My_component"},
		{"simple", "Simple"},
		{"vimesh-test", "Vimesh_test"},
		{"foo-bar-baz", "Foo_bar_baz"},
		{"component", "Component"},
		{"", ""},
		{"a", "A"},
		{"test-component-name", "Test_component_name"},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sh.toComponentClassName(tt.input)
			if result != tt.expected {
				t.Errorf("toComponentClassName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestVimeshStyleMinification(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	
	testCases := []struct {
		name           string
		minifyEnabled  bool
		devMode        bool
		expectMinified bool
	}{
		{"Minification enabled", true, false, true},
		{"Minification disabled", false, false, false},
		{"Dev mode enabled", true, true, false},
		{"Both disabled", false, true, false},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := DefaultSvelteConfig()
			config.MinifyRuntime = tc.minifyEnabled
			config.DevMode = tc.devMode
			config.VimeshStyle = &utils.VimeshStyleConfig{Enable: true}
			
			handler := NewSvelteHandlerWithConfig(fs, config)
			vimeshJS := handler.getMinifiedVimeshStyle()
			
			originalSize := len(utils.GetVimeshStyleJS())
			minifiedSize := len(vimeshJS)
			
			if tc.expectMinified {
				if minifiedSize >= originalSize {
					t.Errorf("Expected minified version to be smaller: original=%d, minified=%d", 
						originalSize, minifiedSize)
				}
				// Should have some reasonable compression
				reduction := float64(originalSize-minifiedSize) / float64(originalSize) * 100
				if reduction < 10 { // Expect at least 10% reduction
					t.Errorf("Expected better compression, got only %.1f%% reduction", reduction)
				}
			} else {
				if minifiedSize != originalSize {
					t.Errorf("Expected original size %d, got %d", originalSize, minifiedSize)
				}
			}
		})
	}
}