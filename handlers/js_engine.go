package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	js "github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/eventloop"
	"github.com/dop251/goja_nodejs/require"
	"github.com/gorilla/mux"

	"github.com/rediwo/redi/filesystem"
)

// CachedModule represents a cached JavaScript module
type CachedModule struct {
	Exports      *js.Object
	LastModified time.Time
}

// SharedJSEngine manages a shared JavaScript environment with module caching
type SharedJSEngine struct {
	fs           filesystem.FileSystem
	version      string
	eventLoop    *eventloop.EventLoop
	vm           *js.Runtime
	registry     *require.Registry
	moduleCache  map[string]*CachedModule
	cacheMutex   sync.RWMutex
	started      bool
	startMutex   sync.Mutex
}

// JSEnginePool manages a pool of JavaScript engines
type JSEnginePool struct {
	engines       []*SharedJSEngine
	available     chan *SharedJSEngine
	fs            filesystem.FileSystem
	version       string
	poolSize      int
	mutex         sync.Mutex
	// Session-based engine allocation
	sessionEngines map[string]*SharedJSEngine
	sessionMutex   sync.RWMutex
}

var (
	globalPools = make(map[string]*JSEnginePool)
	poolMutex   sync.RWMutex
)


// GetJSEnginePool returns a JavaScript engine pool for the given filesystem
func GetJSEnginePool(fs filesystem.FileSystem, version string) *JSEnginePool {
	// Create a key based on filesystem instance pointer and version
	key := fmt.Sprintf("%T-%p-%s", fs, fs, version)
	
	poolMutex.RLock()
	if pool, exists := globalPools[key]; exists {
		poolMutex.RUnlock()
		return pool
	}
	poolMutex.RUnlock()
	
	poolMutex.Lock()
	defer poolMutex.Unlock()
	
	// Double-check after acquiring write lock
	if pool, exists := globalPools[key]; exists {
		return pool
	}
	
	// Create new pool
	pool := &JSEnginePool{
		fs:             fs,
		version:        version,
		poolSize:       3, // Start with 3 engines in the pool
		available:      make(chan *SharedJSEngine, 3),
		sessionEngines: make(map[string]*SharedJSEngine),
	}
	pool.initPool()
	globalPools[key] = pool
	return pool
}

// initPool initializes the engine pool
func (pool *JSEnginePool) initPool() {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	for i := 0; i < pool.poolSize; i++ {
		engine := &SharedJSEngine{
			fs:          pool.fs,
			version:     pool.version,
			moduleCache: make(map[string]*CachedModule),
		}
		
		// Start the engine
		if err := engine.Start(); err != nil {
			// Log error but continue with other engines
			continue
		}
		
		pool.engines = append(pool.engines, engine)
		pool.available <- engine
	}
}

// GetEngine gets an available engine from the pool
func (pool *JSEnginePool) GetEngine() (*SharedJSEngine, error) {
	select {
	case engine := <-pool.available:
		return engine, nil
	default:
		// No engines available, create a new temporary one
		engine := &SharedJSEngine{
			fs:          pool.fs,
			version:     pool.version,
			moduleCache: make(map[string]*CachedModule),
		}
		if err := engine.Start(); err != nil {
			return nil, fmt.Errorf("failed to create temporary engine: %v", err)
		}
		return engine, nil
	}
}

// GetEngineForSession gets an engine specifically for a session/client
func (pool *JSEnginePool) GetEngineForSession(sessionID string) (*SharedJSEngine, error) {
	// Check if we already have an engine for this session
	pool.sessionMutex.RLock()
	if engine, exists := pool.sessionEngines[sessionID]; exists {
		pool.sessionMutex.RUnlock()
		return engine, nil
	}
	pool.sessionMutex.RUnlock()

	// Need to assign an engine to this session
	pool.sessionMutex.Lock()
	defer pool.sessionMutex.Unlock()

	// Double-check after acquiring write lock
	if engine, exists := pool.sessionEngines[sessionID]; exists {
		return engine, nil
	}

	// Try to get an engine from the pool
	var engine *SharedJSEngine
	select {
	case engine = <-pool.available:
		// Got an engine from the pool
	default:
		// No engines available, create a new one
		engine = &SharedJSEngine{
			fs:          pool.fs,
			version:     pool.version,
			moduleCache: make(map[string]*CachedModule),
		}
		if err := engine.Start(); err != nil {
			return nil, fmt.Errorf("failed to create session engine: %v", err)
		}
	}

	// Assign this engine to the session
	pool.sessionEngines[sessionID] = engine
	return engine, nil
}

// ReleaseSessionEngine releases an engine from a session (for cleanup)
func (pool *JSEnginePool) ReleaseSessionEngine(sessionID string) {
	pool.sessionMutex.Lock()
	defer pool.sessionMutex.Unlock()

	if engine, exists := pool.sessionEngines[sessionID]; exists {
		delete(pool.sessionEngines, sessionID)
		// Try to return to pool
		select {
		case pool.available <- engine:
			// Successfully returned to pool
		default:
			// Pool is full, stop this engine
			engine.Stop()
		}
	}
}

// ReturnEngine returns an engine to the pool
func (pool *JSEnginePool) ReturnEngine(engine *SharedJSEngine) {
	if engine == nil {
		return
	}
	
	select {
	case pool.available <- engine:
		// Successfully returned to pool
	default:
		// Pool is full, stop this engine
		engine.Stop()
	}
}

// Stop stops all engines in the pool
func (pool *JSEnginePool) Stop() {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	// Stop all engines
	for _, engine := range pool.engines {
		engine.Stop()
	}
	
	// Clear the pool
	for len(pool.available) > 0 {
		<-pool.available
	}
	
	pool.engines = nil
}

// Start initializes the shared JavaScript engine
func (engine *SharedJSEngine) Start() error {
	engine.startMutex.Lock()
	defer engine.startMutex.Unlock()

	if engine.started {
		return nil
	}

	// Create event loop and VM
	engine.eventLoop = eventloop.NewEventLoop()
	go engine.eventLoop.Start()

	// Initialize VM in the event loop
	done := make(chan error, 1)
	engine.eventLoop.RunOnLoop(func(vm *js.Runtime) {
		engine.vm = vm

		// Set up VM manager and registry
		vmManager := NewVMManager(engine.fs, engine.version)
		registry, err := vmManager.SetupRegistry(engine.eventLoop, vm, "routes")
		if err != nil {
			done <- fmt.Errorf("failed to setup registry: %v", err)
			return
		}
		engine.registry = registry

		// Set up console
		consoleObj := require.Require(vm, "console")
		vm.Set("console", consoleObj)

		// Set up process as global
		processObj := require.Require(vm, "process")
		vm.Set("process", processObj)

		done <- nil
	})

	if err := <-done; err != nil {
		return err
	}

	engine.started = true
	return nil
}

// Stop shuts down the shared JavaScript engine
func (engine *SharedJSEngine) Stop() {
	engine.startMutex.Lock()
	defer engine.startMutex.Unlock()

	if !engine.started {
		return
	}

	if engine.eventLoop != nil {
		engine.eventLoop.Stop()
	}
	engine.started = false
}

// loadOrGetModule loads a JavaScript module file and caches it, or returns cached version
func (engine *SharedJSEngine) loadOrGetModule(filePath string) (*js.Object, error) {
	// Get file modification time first
	info, err := engine.fs.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file %s (filesystem type: %T): %v", filePath, engine.fs, err)
	}
	
	// Check if module is already cached
	engine.cacheMutex.RLock()
	if cached, exists := engine.moduleCache[filePath]; exists {
		// Check if file has been modified using the mod time we just got
		if !info.ModTime().After(cached.LastModified) {
			// File hasn't changed, return cached module
			engine.cacheMutex.RUnlock()
			return cached.Exports, nil
		}
	}
	engine.cacheMutex.RUnlock()

	// Read the file
	content, err := engine.fs.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s (filesystem type: %T): %v", filePath, engine.fs, err)
	}

	// Load module in the shared event loop
	result := make(chan *js.Object, 1)
	errChan := make(chan error, 1)

	engine.eventLoop.RunOnLoop(func(vm *js.Runtime) {
		// Create module wrapper
		moduleCode := fmt.Sprintf(`
			(function(exports, require, module, __filename, __dirname) {
				%s
				return exports;
			})
		`, string(content))

		// Create module objects
		exports := vm.NewObject()
		module := vm.NewObject()
		module.Set("exports", exports)

		// Execute the module wrapper
		fn, err := vm.RunString(moduleCode)
		if err != nil {
			errChan <- fmt.Errorf("failed to compile module %s: %v", filePath, err)
			return
		}

		// Get require function
		requireFunc := vm.Get("require")
		basePath := filepath.Dir(filePath)

		// Call the module wrapper function
		if callable, ok := js.AssertFunction(fn); ok {
			moduleResult, err := callable(js.Undefined(),
				exports,
				requireFunc,
				module,
				vm.ToValue(filePath),
				vm.ToValue(basePath))
			if err != nil {
				errChan <- fmt.Errorf("failed to execute module %s: %v", filePath, err)
				return
			}

			// Get the exports from the result
			if resultObj, ok := moduleResult.(*js.Object); ok {
				exports = resultObj
			}
		}

		result <- exports
	})

	// Wait for module loading to complete
	select {
	case exports := <-result:
		// Cache the module using the modification time we got earlier
		engine.cacheMutex.Lock()
		engine.moduleCache[filePath] = &CachedModule{
			Exports:      exports,
			LastModified: info.ModTime(),
		}
		engine.cacheMutex.Unlock()
		return exports, nil
	case err := <-errChan:
		return nil, err
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("timeout loading module %s", filePath)
	}
}

// ExecuteHTTPMethod executes an HTTP method handler from a JavaScript module
func (engine *SharedJSEngine) ExecuteHTTPMethod(r *http.Request, w http.ResponseWriter, route Route) error {
	if !engine.started {
		return fmt.Errorf("JavaScript engine not started")
	}

	// Load the module
	exports, err := engine.loadOrGetModule(route.FilePath)
	if err != nil {
		return err
	}

	// Check if method handler exists before executing
	httpMethod := strings.ToLower(r.Method)
	if httpMethod == "delete" {
		// Check both 'delete' and 'del' for DELETE method
		if js.IsUndefined(exports.Get("delete")) {
			httpMethod = "del"
		}
	}

	// Check if handler exists
	handler := exports.Get(httpMethod)
	if handler == nil || js.IsUndefined(handler) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return fmt.Errorf("method_not_allowed:%s", httpMethod)
	}

	// Create request object
	reqObj := engine.createRequestObject(r, route)

	// Track response
	responseSent := false
	var responseMutex sync.Mutex
	done := make(chan error, 1)

	// Create response object
	resObj := engine.createResponseObject(w, &responseSent, &responseMutex, func() {
		done <- nil
	}, route)

	// Execute the method handler in the event loop
	engine.eventLoop.RunOnLoop(func(vm *js.Runtime) {
		defer func() {
			if recovered := recover(); recovered != nil {
				done <- fmt.Errorf("JavaScript execution error: %v", recovered)
			}
		}()

		// Get the method handler (we already checked it exists)
		handler := exports.Get(httpMethod)

		// Execute the specific method handler
		// Create next function
		nextFunc := vm.ToValue(func(call js.FunctionCall) js.Value {
			return js.Undefined()
		})

		// Execute the method handler
		if callable, ok := js.AssertFunction(handler); ok {
			_, err := callable(js.Undefined(), vm.ToValue(reqObj), vm.ToValue(resObj), nextFunc)
			if err != nil {
				done <- fmt.Errorf("failed to execute %s handler: %v", httpMethod, err)
				return
			}
		}
	})

	// Wait for completion with timeout
	select {
	case err := <-done:
		return err
	case <-time.After(10 * time.Second):
		responseMutex.Lock()
		if !responseSent {
			http.Error(w, "Request timeout", http.StatusRequestTimeout)
		}
		responseMutex.Unlock()
		return fmt.Errorf("request timeout")
	}
}

// createRequestObject creates a request object for JavaScript
func (engine *SharedJSEngine) createRequestObject(r *http.Request, route Route) map[string]interface{} {
	vars := mux.Vars(r)

	reqObj := map[string]interface{}{
		"method":  r.Method,
		"url":     r.URL.String(),
		"path":    r.URL.Path,
		"query":   r.URL.RawQuery,
		"headers": r.Header,
		"params":  vars,
	}

	// Add body for non-GET requests
	if r.Method != "GET" && r.Method != "HEAD" {
		body, _ := io.ReadAll(r.Body)
		reqObj["body"] = string(body)
	}

	return reqObj
}

// createResponseObject creates a response object
func (engine *SharedJSEngine) createResponseObject(w http.ResponseWriter, responseSent *bool, responseMutex *sync.Mutex, onResponse func(), route Route) map[string]interface{} {
	statusCode := 200

	return map[string]interface{}{
		"json": func(data interface{}) {
			responseMutex.Lock()
			defer responseMutex.Unlock()

			if *responseSent {
				return // Prevent duplicate responses
			}
			*responseSent = true

			w.Header().Set("Content-Type", "application/json")
			if statusCode != 200 {
				w.WriteHeader(statusCode)
			}
			json.NewEncoder(w).Encode(data)
			onResponse()
		},
		"send": func(data interface{}) {
			responseMutex.Lock()
			defer responseMutex.Unlock()

			if *responseSent {
				return // Prevent duplicate responses
			}
			*responseSent = true

			w.Header().Set("Content-Type", "text/plain")
			if statusCode != 200 {
				w.WriteHeader(statusCode)
			}
			fmt.Fprint(w, data)
			onResponse()
		},
		"render": func(data interface{}) {
			responseMutex.Lock()
			defer responseMutex.Unlock()

			if *responseSent {
				return // Prevent duplicate responses
			}
			*responseSent = true

			// Auto-find template file based on JS file path
			err := engine.renderTemplate(route, data, w, statusCode)
			if err != nil {
				// Log the error instead of sending HTTP error (which would cause duplicate WriteHeader)
				fmt.Printf("Template rendering error: %v\n", err)
				// Write error content directly without calling WriteHeader again
				fmt.Fprintf(w, "Template rendering error: %v", err)
			}
			onResponse()
		},
		"status": func(code int) {
			statusCode = code
		},
		"setHeader": func(key, value string) {
			w.Header().Set(key, value)
		},
	}
}


// renderTemplate finds and renders the template file corresponding to the JS file
func (engine *SharedJSEngine) renderTemplate(route Route, data interface{}, w http.ResponseWriter, statusCode int) error {
	// Convert .js file path to template file path
	templatePath := engine.findTemplatePath(route.FilePath)
	if templatePath == "" {
		return fmt.Errorf("no template file found for %s", route.FilePath)
	}

	// Read template file
	templateContent, err := engine.fs.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("template file not found: %s", templatePath)
	}

	// Create template handler to render the template
	templateHandler := NewTemplateHandler(engine.fs)
	
	// Set status code if not 200
	if statusCode != 200 {
		w.WriteHeader(statusCode)
	}

	// Render the template
	return templateHandler.RenderTemplate(templatePath, string(templateContent), data, w)
}


// findTemplatePath finds the template file corresponding to a JS file
func (engine *SharedJSEngine) findTemplatePath(jsFilePath string) string {
	// Remove .js extension
	basePath := strings.TrimSuffix(jsFilePath, ".js")
	
	// Try different template extensions in order of preference
	templateExtensions := []string{".html", ".md", ".txt", ".json"}
	
	for _, ext := range templateExtensions {
		templatePath := basePath + ext
		if _, err := engine.fs.Stat(templatePath); err == nil {
			return templatePath
		}
	}
	
	return ""
}