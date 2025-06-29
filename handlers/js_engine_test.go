package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rediwo/redi/filesystem"
)

func TestJSEnginePool_GetAndReturnEngine(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	pool := GetJSEnginePool(fs, "test-version")

	// Get an engine from the pool
	engine1, err := pool.GetEngine()
	if err != nil {
		t.Errorf("Expected no error getting engine, got: %v", err)
	}
	if engine1 == nil {
		t.Error("Expected engine, got nil")
	}

	// Get another engine
	engine2, err := pool.GetEngine()
	if err != nil {
		t.Errorf("Expected no error getting second engine, got: %v", err)
	}
	if engine2 == nil {
		t.Error("Expected second engine, got nil")
	}

	// Return engines to pool
	pool.ReturnEngine(engine1)
	pool.ReturnEngine(engine2)

	// Verify we can get engines again
	engine3, err := pool.GetEngine()
	if err != nil {
		t.Errorf("Expected no error getting engine after return, got: %v", err)
	}
	if engine3 == nil {
		t.Error("Expected engine after return, got nil")
	}

	pool.ReturnEngine(engine3)
}

func TestJSEnginePool_PoolExhaustion(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	pool := GetJSEnginePool(fs, "test-version")

	// Get all engines from pool
	var engines []*SharedJSEngine
	for i := 0; i < 5; i++ { // Try to get more than pool size
		engine, err := pool.GetEngine()
		if err != nil {
			t.Errorf("Expected no error getting engine %d, got: %v", i, err)
		}
		if engine == nil {
			t.Errorf("Expected engine %d, got nil", i)
		}
		engines = append(engines, engine)
	}

	// Return all engines
	for _, engine := range engines {
		pool.ReturnEngine(engine)
	}
}

func TestSharedJSEngine_ModuleCaching(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	fs.WriteFile("test.js", []byte(`exports.get = function(req, res, next) { res.json({cached: true}); };`))

	engine := &SharedJSEngine{
		fs:          fs,
		version:     "test",
		moduleCache: make(map[string]*CachedModule),
	}

	err := engine.Start()
	if err != nil {
		t.Errorf("Expected no error starting engine, got: %v", err)
	}
	defer engine.Stop()

	// Load module first time
	exports1, err := engine.loadOrGetModule("test.js")
	if err != nil {
		t.Errorf("Expected no error loading module, got: %v", err)
	}
	if exports1 == nil {
		t.Error("Expected exports, got nil")
	}

	// Check that module was cached
	engine.cacheMutex.RLock()
	cached, exists := engine.moduleCache["test.js"]
	engine.cacheMutex.RUnlock()
	if !exists {
		t.Error("Expected module to be cached after first load")
	}
	if cached == nil || cached.Exports != exports1 {
		t.Error("Expected cached module to match first exports")
	}

	// Load module second time (should be cached)
	exports2, err := engine.loadOrGetModule("test.js")
	if err != nil {
		t.Errorf("Expected no error loading cached module, got: %v", err)
	}
	if exports2 == nil {
		t.Error("Expected cached exports, got nil")
	}

	// Verify cache was used (same object reference)
	if exports1 != exports2 {
		t.Errorf("Expected same cached exports object. exports1=%p, exports2=%p", exports1, exports2)
	}
}

func TestSharedJSEngine_FindTemplatePath(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	fs.WriteFile("routes/test.html", []byte(`<h1>Test</h1>`))
	fs.WriteFile("routes/blog.md", []byte(`# Blog`))

	engine := &SharedJSEngine{fs: fs}

	tests := []struct {
		jsPath       string
		expectedPath string
	}{
		{"routes/test.js", "routes/test.html"},
		{"routes/blog.js", "routes/blog.md"},
		{"routes/nonexistent.js", ""},
	}

	for _, test := range tests {
		result := engine.findTemplatePath(test.jsPath)
		if result != test.expectedPath {
			t.Errorf("For JS path %s, expected template %s, got %s", test.jsPath, test.expectedPath, result)
		}
	}
}

func TestSharedJSEngine_RenderTemplate(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	fs.WriteFile("routes/test.html", []byte(`<h1>{{.Title}}</h1>`))

	engine := &SharedJSEngine{fs: fs}
	err := engine.Start()
	if err != nil {
		t.Errorf("Expected no error starting engine, got: %v", err)
	}
	defer engine.Stop()

	route := Route{FilePath: "routes/test.js"}
	data := map[string]interface{}{"Title": "Test Title"}
	w := httptest.NewRecorder()

	err = engine.renderTemplate(route, data, w, 200)
	if err != nil {
		t.Errorf("Expected no error rendering template, got: %v", err)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Test Title") {
		t.Errorf("Expected 'Test Title' in output, got: %s", body)
	}

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestSharedJSEngine_RenderTemplate_NoTemplate(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()

	engine := &SharedJSEngine{fs: fs}
	err := engine.Start()
	if err != nil {
		t.Errorf("Expected no error starting engine, got: %v", err)
	}
	defer engine.Stop()

	route := Route{FilePath: "routes/nonexistent.js"}
	data := map[string]interface{}{}
	w := httptest.NewRecorder()

	err = engine.renderTemplate(route, data, w, 200)
	if err == nil {
		t.Error("Expected error for missing template, got none")
	}

	if !strings.Contains(err.Error(), "no template file found") {
		t.Errorf("Expected 'no template file found' error, got: %v", err)
	}
}

func TestSharedJSEngine_ExecuteHTTPMethod(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	fs.WriteFile("test.js", []byte(`
		exports.get = function(req, res, next) {
			res.json({method: req.method, success: true});
		};
		exports.post = function(req, res, next) {
			res.json({method: req.method, body: req.body});
		};
	`))

	engine := &SharedJSEngine{
		fs:          fs,
		version:     "test",
		moduleCache: make(map[string]*CachedModule),
	}

	err := engine.Start()
	if err != nil {
		t.Errorf("Expected no error starting engine, got: %v", err)
	}
	defer engine.Stop()

	// Test GET request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	route := Route{FilePath: "test.js"}

	err = engine.ExecuteHTTPMethod(req, w, route)
	if err != nil {
		t.Errorf("Expected no error executing GET, got: %v", err)
	}

	// Test POST request
	req = httptest.NewRequest("POST", "/test", strings.NewReader(`{"data": "test"}`))
	w = httptest.NewRecorder()

	err = engine.ExecuteHTTPMethod(req, w, route)
	if err != nil {
		t.Errorf("Expected no error executing POST, got: %v", err)
	}
}

func TestSharedJSEngine_ExecuteHTTPMethod_MethodNotAllowed(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	fs.WriteFile("test.js", []byte(`
		exports.get = function(req, res, next) {
			res.json({success: true});
		};
	`))

	engine := &SharedJSEngine{
		fs:          fs,
		version:     "test",
		moduleCache: make(map[string]*CachedModule),
	}

	err := engine.Start()
	if err != nil {
		t.Errorf("Expected no error starting engine, got: %v", err)
	}
	defer engine.Stop()

	// Test unsupported method
	req := httptest.NewRequest("PATCH", "/test", nil)
	w := httptest.NewRecorder()
	route := Route{FilePath: "test.js"}

	err = engine.ExecuteHTTPMethod(req, w, route)
	if err == nil {
		t.Error("Expected error for unsupported method, got none")
	}

	if w.Code != 405 {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestSharedJSEngine_StartStop(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	engine := &SharedJSEngine{
		fs:          fs,
		version:     "test",
		moduleCache: make(map[string]*CachedModule),
	}

	// Test start
	err := engine.Start()
	if err != nil {
		t.Errorf("Expected no error starting engine, got: %v", err)
	}

	if !engine.started {
		t.Error("Expected engine to be started")
	}

	// Test double start (should be no-op)
	err = engine.Start()
	if err != nil {
		t.Errorf("Expected no error on double start, got: %v", err)
	}

	// Test stop
	engine.Stop()
	if engine.started {
		t.Error("Expected engine to be stopped")
	}

	// Test double stop (should be no-op)
	engine.Stop()
}

func TestSharedJSEngine_ConcurrentAccess(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	fs.WriteFile("test.js", []byte(`
		exports.get = function(req, res, next) {
			res.json({timestamp: Date.now()});
		};
	`))

	engine := &SharedJSEngine{
		fs:          fs,
		version:     "test",
		moduleCache: make(map[string]*CachedModule),
	}

	err := engine.Start()
	if err != nil {
		t.Errorf("Expected no error starting engine, got: %v", err)
	}
	defer engine.Stop()

	// Run multiple concurrent requests
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			route := Route{FilePath: "test.js"}

			err := engine.ExecuteHTTPMethod(req, w, route)
			if err != nil {
				t.Errorf("Request %d failed: %v", i, err)
			}
			done <- true
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// Request completed
		case <-time.After(5 * time.Second):
			t.Errorf("Request %d timed out", i)
		}
	}
}