package handlers

import (
	"crypto/md5"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/gorilla/mux"
	"github.com/rediwo/redi/filesystem"
	"github.com/rediwo/redi/utils"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/js"
)

//go:embed svelte-compiler.js
var svelteCompilerJS string

//go:embed svelte-runtime.js
var svelteRuntimeJS string

//go:embed svelte-async.js
var svelteAsyncJS string

// SvelteConfig holds all Svelte-related settings
type SvelteConfig struct {
	// Minification settings
	MinifyRuntime    bool // Enable runtime minification
	MinifyComponents bool // Enable component code minification
	MinifyCSS        bool // Enable CSS minification
	DevMode          bool // Disable minification in development

	// Runtime resource settings
	UseExternalRuntime   bool          // Use external runtime file instead of inline
	RuntimePath          string        // Path for runtime resource (default: "/svelte/runtime.js")
	RuntimeCacheDuration time.Duration // Cache duration for runtime resource

	// Async component loading settings
	EnableAsyncLoading     bool          // Enable async component loading
	ComponentCacheDuration time.Duration // Cache duration for component resources
	AsyncLibraryPath       string        // Path for async library (default: "/svelte/async.js")

	// Vimesh Style settings
	VimeshStyle     *utils.VimeshStyleConfig // Vimesh Style configuration
	VimeshStylePath string                   // Path for Vimesh Style resource (default: "/svelte/vimesh-style.js")
}

// DefaultSvelteConfig returns default Svelte settings
func DefaultSvelteConfig() *SvelteConfig {
	return &SvelteConfig{
		MinifyRuntime:          true,
		MinifyComponents:       true,
		MinifyCSS:              true,
		DevMode:                false,
		UseExternalRuntime:     true,
		RuntimePath:            "/svelte/runtime.js",
		RuntimeCacheDuration:   365 * 24 * time.Hour, // 1 year
		EnableAsyncLoading:     true,
		ComponentCacheDuration: 24 * time.Hour, // 1 day
		AsyncLibraryPath:       "/svelte/async.js",
		VimeshStyle:            utils.DefaultVimeshStyleConfig(),
		VimeshStylePath:        "/svelte/vimesh-style.js",
	}
}

type SvelteHandler struct {
	fs                  filesystem.FileSystem
	vm                  *goja.Runtime
	compileFunc         goja.Callable
	initialized         bool
	templateHandler     *TemplateHandler
	mu                  sync.Mutex
	cache               map[string]*CachedResult
	cacheMu             sync.RWMutex
	config              *SvelteConfig
	minifiedRuntime     string
	runtimeMinified     bool
	runtimeMu           sync.Mutex
	minifiedVimeshStyle string
	vimeshMinified      bool
	vimeshMu            sync.Mutex
	minifiedAsyncLib    string
	asyncLibMinified    bool
	asyncLibMu          sync.Mutex
	minifier            *minify.M
	componentRegistry   map[string]*ComponentInfo // Registry for compiled components
	registryMu          sync.RWMutex              // Mutex for component registry
	importTransformer   *ImportTransformer        // Common import handling
	routesDir           string                    // Routes directory path
}

type CachedResult struct {
	HTML         string
	ContentHash  string   // MD5 hash of the source content
	ConfigHash   string   // Hash of config for cache invalidation
	Dependencies []string // List of component dependencies
}

type SvelteCompileResult struct {
	JS  string `json:"js"`
	CSS string `json:"css"`
}

type AsyncComponentResponse struct {
	Success      bool                     `json:"success"`
	Component    string                   `json:"component"`
	ClassName    string                   `json:"className"`
	JS           string                   `json:"js"`
	CSS          string                   `json:"css"`
	Dependencies []AsyncComponentResponse `json:"dependencies,omitempty"`
	Error        string                   `json:"error,omitempty"`
}

type ComponentInfo struct {
	FilePath     string    // Path to the component file
	Name         string    // Component name (e.g., "Button")
	ClassName    string    // Class name (e.g., "Button")
	CompiledJS   string    // Compiled JavaScript code
	CompiledCSS  string    // Compiled CSS code
	Dependencies []string  // Import dependencies
	LastModified time.Time // Last modification time
	ContentHash  string    // Hash of source content
}

func NewSvelteHandler(fs filesystem.FileSystem) *SvelteHandler {
	return NewSvelteHandlerWithConfig(fs, DefaultSvelteConfig())
}

func NewSvelteHandlerWithConfig(fs filesystem.FileSystem, config *SvelteConfig) *SvelteHandler {
	// Create and configure minifier with safer settings
	m := minify.New()

	// Configure JavaScript minifier with much safer options for Svelte runtime
	jsMinifier := &js.Minifier{
		Precision:    0,    // Keep full precision for numbers
		KeepVarNames: true, // Keep variable names to avoid breaking references
	}
	m.Add("application/javascript", jsMinifier)
	m.AddFunc("text/css", css.Minify)

	return &SvelteHandler{
		fs:                fs,
		templateHandler:   NewTemplateHandler(fs),
		cache:             make(map[string]*CachedResult),
		config:            config,
		minifier:          m,
		componentRegistry: make(map[string]*ComponentInfo),
		importTransformer: NewImportTransformer(fs),
		routesDir:         "routes", // Default value
	}
}

// NewSvelteHandlerWithRouter creates a SvelteHandler and registers runtime routes
func NewSvelteHandlerWithRouter(fs filesystem.FileSystem, config *SvelteConfig, router *mux.Router) *SvelteHandler {
	sh := NewSvelteHandlerWithConfig(fs, config)
	sh.RegisterRoutes(router)
	return sh
}

// NewSvelteHandlerWithRouterAndRoutesDir creates a SvelteHandler with custom routes directory
func NewSvelteHandlerWithRouterAndRoutesDir(fs filesystem.FileSystem, config *SvelteConfig, router *mux.Router, routesDir string) *SvelteHandler {
	sh := NewSvelteHandlerWithConfig(fs, config)
	sh.routesDir = routesDir
	sh.RegisterRoutes(router)
	return sh
}

// RegisterRoutes registers Svelte-related routes
func (sh *SvelteHandler) RegisterRoutes(router *mux.Router) {
	if sh.config.UseExternalRuntime {
		router.HandleFunc(sh.config.RuntimePath, sh.ServeSvelteRuntime).Methods("GET", "HEAD")
		log.Printf("Registered Svelte runtime route: %s", sh.config.RuntimePath)
	}

	// Register Vimesh Style route if enabled
	if sh.config.VimeshStyle != nil && sh.config.VimeshStyle.Enable {
		router.HandleFunc(sh.config.VimeshStylePath, sh.ServeVimeshStyle).Methods("GET", "HEAD")
		log.Printf("Registered Svelte Vimesh Style route: %s", sh.config.VimeshStylePath)
	}

	// Register async library if enabled
	if sh.config.EnableAsyncLoading {
		router.HandleFunc(sh.config.AsyncLibraryPath, sh.ServeAsyncLibrary).Methods("GET", "HEAD")
		log.Printf("Registered Svelte async library route: %s", sh.config.AsyncLibraryPath)
	}
}

// ServeSvelteRuntime serves the Svelte runtime as a static resource
func (sh *SvelteHandler) ServeSvelteRuntime(w http.ResponseWriter, r *http.Request) {
	runtime := sh.getMinifiedRuntime()

	// Calculate ETag based on runtime content
	hash := md5.Sum([]byte(runtime))
	etag := `"` + hex.EncodeToString(hash[:]) + `"`

	// Check if client has cached version
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// Set caching headers
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d, immutable", int(sh.config.RuntimeCacheDuration.Seconds())))

	// Add compression hint
	w.Header().Set("Vary", "Accept-Encoding")

	w.Write([]byte(runtime))
}

// ServeVimeshStyle serves the Vimesh Style JavaScript as a static resource
func (sh *SvelteHandler) ServeVimeshStyle(w http.ResponseWriter, r *http.Request) {
	vimeshJS := sh.getMinifiedVimeshStyle()

	// Calculate ETag based on content
	hash := md5.Sum([]byte(vimeshJS))
	etag := `"` + hex.EncodeToString(hash[:]) + `"`

	// Check if client has cached version
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// Set caching headers
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d, immutable", int(sh.config.RuntimeCacheDuration.Seconds())))

	// Add compression hint
	w.Header().Set("Vary", "Accept-Encoding")

	w.Write([]byte(vimeshJS))
}

// ServeAsyncComponent serves individual components as JSON for async loading
func (sh *SvelteHandler) ServeAsyncComponent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	componentPath := vars["component"]

	// Validate component path
	if componentPath == "" {
		http.Error(w, "Component path is required", http.StatusBadRequest)
		return
	}

	// Add .svelte extension if not present
	if !strings.HasSuffix(componentPath, ".svelte") {
		componentPath += ".svelte"
	}

	// Use the existing component resolution logic
	fullPath := sh.resolveAsyncComponentPath(componentPath, r)

	if fullPath == "" {
		response := AsyncComponentResponse{
			Success: false,
			Error:   "Component not found",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get component info
	info, err := sh.getComponentInfo(fullPath)
	if err != nil {
		response := AsyncComponentResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to compile component: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Transform component to IIFE format
	imports := make(map[string]string)
	jsCode := sh.transformToIIFE(info.CompiledJS, info.ClassName, imports, info.FilePath)

	// Wrap the component code to ensure it has access to Svelte runtime functions
	// The runtime functions are expected to be available in the global scope
	jsCode = `/* Component: ` + info.Name + ` */
` + jsCode

	jsCode = sh.minifyComponentCode(jsCode)

	// Create response
	response := AsyncComponentResponse{
		Success:   true,
		Component: info.Name,
		ClassName: info.ClassName,
		JS:        jsCode,
		CSS:       sh.minifyCSS(info.CompiledCSS),
	}

	// Add dependencies if requested
	if r.URL.Query().Get("include_deps") == "true" {
		for _, depPath := range info.Dependencies {
			depInfo, err := sh.getComponentInfo(depPath)
			if err != nil {
				continue // Skip failed dependencies
			}

			depImports := make(map[string]string)
			depJS := sh.transformToIIFE(depInfo.CompiledJS, depInfo.ClassName, depImports, depInfo.FilePath)
			depJS = sh.minifyComponentCode(depJS)

			depResponse := AsyncComponentResponse{
				Success:   true,
				Component: depInfo.Name,
				ClassName: depInfo.ClassName,
				JS:        depJS,
				CSS:       sh.minifyCSS(depInfo.CompiledCSS),
			}
			response.Dependencies = append(response.Dependencies, depResponse)
		}
	}

	// Set caching headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", int(sh.config.ComponentCacheDuration.Seconds())))

	// Calculate ETag based on component content
	hash := md5.Sum([]byte(info.ContentHash))
	etag := `"` + hex.EncodeToString(hash[:]) + `"`
	w.Header().Set("ETag", etag)

	// Check if client has cached version
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	json.NewEncoder(w).Encode(response)
}

// ServeAsyncLibrary serves the async component loading library
func (sh *SvelteHandler) ServeAsyncLibrary(w http.ResponseWriter, r *http.Request) {
	asyncLib := sh.getMinifiedAsyncLibrary()

	// Calculate ETag based on library content
	hash := md5.Sum([]byte(asyncLib))
	etag := `"` + hex.EncodeToString(hash[:]) + `"`

	// Check if client has cached version
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// Set caching headers
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d, immutable", int(sh.config.RuntimeCacheDuration.Seconds())))

	// Add compression hint
	w.Header().Set("Vary", "Accept-Encoding")

	w.Write([]byte(asyncLib))
}

// getMinifiedRuntime returns the minified runtime, minifying it on first use
func (sh *SvelteHandler) getMinifiedRuntime() string {
	// If minification is disabled or we're in dev mode, return original
	if !sh.config.MinifyRuntime || sh.config.DevMode {
		return svelteRuntimeJS
	}

	sh.runtimeMu.Lock()
	defer sh.runtimeMu.Unlock()

	// Return cached minified runtime if available
	if sh.runtimeMinified {
		return sh.minifiedRuntime
	}

	// Minify the runtime
	minified, err := sh.minifier.String("application/javascript", svelteRuntimeJS)
	if err != nil {
		log.Printf("Failed to minify Svelte runtime: %v, using original", err)
		sh.minifiedRuntime = svelteRuntimeJS
	} else {
		sh.minifiedRuntime = minified
		reduction := float64(len(svelteRuntimeJS)-len(minified)) / float64(len(svelteRuntimeJS)) * 100
		log.Printf("Svelte runtime minified: %d bytes -> %d bytes (%.1f%% reduction)",
			len(svelteRuntimeJS), len(minified), reduction)
	}

	sh.runtimeMinified = true
	return sh.minifiedRuntime
}

// getMinifiedVimeshStyle returns the minified Vimesh Style, minifying it on first use
func (sh *SvelteHandler) getMinifiedVimeshStyle() string {
	// If minification is disabled or we're in dev mode, return original
	if !sh.config.MinifyRuntime || sh.config.DevMode {
		return utils.GetVimeshStyleJS()
	}

	sh.vimeshMu.Lock()
	defer sh.vimeshMu.Unlock()

	// Return cached minified version if available
	if sh.vimeshMinified {
		return sh.minifiedVimeshStyle
	}

	// Get the original Vimesh Style JavaScript
	vimeshJS := utils.GetVimeshStyleJS()

	// Minify the Vimesh Style
	minified, err := sh.minifier.String("application/javascript", vimeshJS)
	if err != nil {
		log.Printf("Failed to minify Vimesh Style: %v, using original", err)
		sh.minifiedVimeshStyle = vimeshJS
	} else {
		sh.minifiedVimeshStyle = minified
		reduction := float64(len(vimeshJS)-len(minified)) / float64(len(vimeshJS)) * 100
		log.Printf("Vimesh Style minified: %d bytes -> %d bytes (%.1f%% reduction)",
			len(vimeshJS), len(minified), reduction)
	}

	sh.vimeshMinified = true
	return sh.minifiedVimeshStyle
}

// getMinifiedAsyncLibrary returns the minified async library, minifying it on first use
func (sh *SvelteHandler) getMinifiedAsyncLibrary() string {
	// If minification is disabled or we're in dev mode, return original
	if !sh.config.MinifyRuntime || sh.config.DevMode {
		return svelteAsyncJS
	}

	sh.asyncLibMu.Lock()
	defer sh.asyncLibMu.Unlock()

	// Return cached minified version if available
	if sh.asyncLibMinified {
		return sh.minifiedAsyncLib
	}

	// Minify the async library
	minified, err := sh.minifier.String("application/javascript", svelteAsyncJS)
	if err != nil {
		log.Printf("Failed to minify async library: %v, using original", err)
		sh.minifiedAsyncLib = svelteAsyncJS
	} else {
		sh.minifiedAsyncLib = minified
		reduction := float64(len(svelteAsyncJS)-len(minified)) / float64(len(svelteAsyncJS)) * 100
		log.Printf("Async library minified: %d bytes -> %d bytes (%.1f%% reduction)",
			len(svelteAsyncJS), len(minified), reduction)
	}

	sh.asyncLibMinified = true
	return sh.minifiedAsyncLib
}

// minifyComponentCode minifies the compiled component JavaScript if enabled
func (sh *SvelteHandler) minifyComponentCode(jsCode string) string {
	if !sh.config.MinifyComponents || sh.config.DevMode {
		return jsCode
	}

	minified, err := sh.minifier.String("application/javascript", jsCode)
	if err != nil {
		log.Printf("Failed to minify component code: %v, using original", err)
		return jsCode
	}

	return minified
}

// minifyCSS minifies CSS code if enabled
func (sh *SvelteHandler) minifyCSS(cssCode string) string {
	if !sh.config.MinifyCSS || sh.config.DevMode {
		return cssCode
	}

	minified, err := sh.minifier.String("text/css", cssCode)
	if err != nil {
		log.Printf("Failed to minify CSS: %v, using original", err)
		return cssCode
	}

	return minified
}

// calculateConfigHash creates a hash of the config for cache invalidation
func (sh *SvelteHandler) calculateConfigHash() string {
	if sh.config == nil {
		return "none"
	}

	vimeshEnabled := false
	if sh.config.VimeshStyle != nil {
		vimeshEnabled = sh.config.VimeshStyle.Enable
	}

	configString := fmt.Sprintf("%t_%t_%t_%t_%t_%s_%t_%s",
		sh.config.MinifyRuntime,
		sh.config.MinifyComponents,
		sh.config.MinifyCSS,
		sh.config.DevMode,
		sh.config.UseExternalRuntime,
		sh.config.RuntimePath,
		vimeshEnabled,
		sh.config.VimeshStylePath)

	hash := md5.Sum([]byte(configString))
	return hex.EncodeToString(hash[:])
}

// parseImports extracts import statements from Svelte source code
func (sh *SvelteHandler) parseImports(source string) []string {
	var imports []string

	// Match import statements for various file types
	// Supports: import Component from './Component.svelte'
	// Supports: import { Component } from './Component.svelte'
	// Supports: import styles from './styles.css'
	// Supports: import data from '../data.json'
	// Supports: import utils from './utils.js'
	// Supports: import logo from '/images/logo.png'
	importRegex := regexp.MustCompile(`import\s+(?:{[^}]+}|\w+)\s+from\s+['"]([^'"]+)['"];?`)
	matches := importRegex.FindAllStringSubmatch(source, -1)

	for _, match := range matches {
		if len(match) > 1 {
			imports = append(imports, match[1])
		}
	}

	return imports
}

// resolveComponentPath resolves a component import path relative to the current component
func (sh *SvelteHandler) resolveComponentPath(importPath string, currentPath string) string {
	// Handle absolute imports (starting with /)
	if strings.HasPrefix(importPath, "/") {
		// Remove leading slash and treat as relative to root
		return strings.TrimPrefix(importPath, "/")
	}

	// Get the directory of the current component
	currentDir := filepath.Dir(currentPath)

	// Resolve the import path relative to the current directory
	resolvedPath := filepath.Join(currentDir, importPath)

	// Clean the path to handle .. and .
	resolvedPath = filepath.Clean(resolvedPath)

	// Ensure the resolved path doesn't escape the root directory
	// by checking it doesn't start with ../
	if strings.HasPrefix(resolvedPath, "../") {
		// Security: prevent directory traversal attacks
		return ""
	}

	return resolvedPath
}

// resolveAsyncComponentPath resolves the path for an async component request
func (sh *SvelteHandler) resolveAsyncComponentPath(componentPath string, r *http.Request) string {
	// Simply resolve the component path relative to routesDir
	fullPath := filepath.Join(sh.routesDir, componentPath)

	// Check if the file exists
	if _, err := sh.fs.Stat(fullPath); err == nil {
		return fullPath
	}

	// Not found
	return ""
}

// getComponentInfo retrieves or compiles a component
func (sh *SvelteHandler) getComponentInfo(componentPath string) (*ComponentInfo, error) {
	// Check registry first
	sh.registryMu.RLock()
	info, exists := sh.componentRegistry[componentPath]
	sh.registryMu.RUnlock()

	// Check if component needs recompilation
	if exists {
		// Check if file has been modified
		stat, err := sh.fs.Stat(componentPath)
		if err == nil && !stat.ModTime().After(info.LastModified) {
			return info, nil
		}
	}

	// Read and compile the component
	content, err := sh.fs.ReadFile(componentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read component %s: %w", componentPath, err)
	}

	contentStr := string(content)
	contentHash := sh.calculateMD5(contentStr)

	// Parse imports
	imports := sh.parseImports(contentStr)
	dependencies := make([]string, 0)

	// Only add Svelte components as dependencies
	for _, imp := range imports {
		if strings.HasSuffix(imp, ".svelte") {
			resolvedPath := sh.resolveComponentPath(imp, componentPath)
			if resolvedPath != "" {
				dependencies = append(dependencies, resolvedPath)
			}
		}
		// Other asset types are handled during HTML generation
	}

	// Compile the component
	result, err := sh.compileSvelte(contentStr, componentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to compile component %s: %w", componentPath, err)
	}

	// Extract component name from filename
	basename := filepath.Base(componentPath)
	componentName := strings.TrimSuffix(basename, ".svelte")

	// Extract the actual class name from compiled JS
	actualClassName := sh.extractActualClassName(result.JS)
	if actualClassName == "" {
		// Fallback to calculated name if extraction fails
		actualClassName = sh.toComponentClassName(componentName)
	}

	// Create component info
	info = &ComponentInfo{
		FilePath:     componentPath,
		Name:         componentName,
		ClassName:    actualClassName,
		CompiledJS:   result.JS,
		CompiledCSS:  result.CSS,
		Dependencies: dependencies,
		LastModified: time.Now(),
		ContentHash:  contentHash,
	}

	// Store in registry
	sh.registryMu.Lock()
	sh.componentRegistry[componentPath] = info
	sh.registryMu.Unlock()

	return info, nil
}

func (sh *SvelteHandler) initializeCompiler() error {
	if sh.initialized {
		return nil
	}

	sh.vm = goja.New()
	sh.vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	// Provide required globals for Svelte compiler
	sh.vm.Set("global", sh.vm.GlobalObject())
	sh.vm.Set("window", sh.vm.GlobalObject())

	_, err := sh.vm.RunString(`
		// Polyfill performance.now() for Svelte compiler
		if (typeof performance === 'undefined') {
			performance = {
				now: function() {
					return Date.now();
				}
			};
		}
		
		// Polyfill console if needed
		if (typeof console === 'undefined') {
			console = {
				log: function() {},
				warn: function() {},
				error: function() {}
			};
		}
		
		// Polyfill btoa for base64 encoding
		if (typeof btoa === 'undefined') {
			btoa = function(str) {
				var chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=';
				var encoded = '';
				var c1, c2, c3;
				var e1, e2, e3, e4;
				
				for (var i = 0; i < str.length; ) {
					c1 = str.charCodeAt(i++);
					c2 = str.charCodeAt(i++);
					c3 = str.charCodeAt(i++);
					
					e1 = c1 >> 2;
					e2 = ((c1 & 3) << 4) | (c2 >> 4);
					e3 = ((c2 & 15) << 2) | (c3 >> 6);
					e4 = c3 & 63;
					
					if (isNaN(c2)) {
						e3 = e4 = 64;
					} else if (isNaN(c3)) {
						e4 = 64;
					}
					
					encoded += chars.charAt(e1) + chars.charAt(e2) + chars.charAt(e3) + chars.charAt(e4);
				}
				
				return encoded;
			};
		}
		
		// Polyfill atob for base64 decoding
		if (typeof atob === 'undefined') {
			atob = function(str) {
				var chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=';
				var decoded = '';
				var c1, c2, c3;
				var e1, e2, e3, e4;
				
				str = str.replace(/[^A-Za-z0-9\+\/\=]/g, '');
				
				for (var i = 0; i < str.length; ) {
					e1 = chars.indexOf(str.charAt(i++));
					e2 = chars.indexOf(str.charAt(i++));
					e3 = chars.indexOf(str.charAt(i++));
					e4 = chars.indexOf(str.charAt(i++));
					
					c1 = (e1 << 2) | (e2 >> 4);
					c2 = ((e2 & 15) << 4) | (e3 >> 2);
					c3 = ((e3 & 3) << 6) | e4;
					
					decoded += String.fromCharCode(c1);
					
					if (e3 != 64) {
						decoded += String.fromCharCode(c2);
					}
					if (e4 != 64) {
						decoded += String.fromCharCode(c3);
					}
				}
				
				return decoded;
			};
		}
	`)
	if err != nil {
		return fmt.Errorf("failed to setup polyfills: %w", err)
	}

	// Load the Svelte compiler
	_, err = sh.vm.RunString(svelteCompilerJS)
	if err != nil {
		return fmt.Errorf("failed to load Svelte compiler: %w", err)
	}

	// Check if svelte object exists
	svelteObj := sh.vm.Get("svelte")
	if svelteObj == nil {
		return fmt.Errorf("svelte object not found")
	}

	// Get the compile function
	svelteObjValue := svelteObj.ToObject(sh.vm)
	if svelteObjValue == nil {
		return fmt.Errorf("svelte is not an object")
	}

	compileFunc := svelteObjValue.Get("compile")
	if compileFunc == nil {
		return fmt.Errorf("svelte.compile not found")
	}

	callable, ok := goja.AssertFunction(compileFunc)
	if !ok {
		return fmt.Errorf("svelte.compile is not a function")
	}
	sh.compileFunc = callable

	sh.initialized = true
	return nil
}

func (sh *SvelteHandler) compileSvelte(source string, filename string) (*SvelteCompileResult, error) {
	sh.mu.Lock()
	defer sh.mu.Unlock()

	if err := sh.initializeCompiler(); err != nil {
		return nil, err
	}

	// Compile options - use minimal required options
	options := map[string]interface{}{
		"filename": filename,
		"generate": "dom",
		"dev":      false,
		"css":      true,
	}

	// Call compile function
	result, err := sh.compileFunc(goja.Undefined(), sh.vm.ToValue(source), sh.vm.ToValue(options))
	if err != nil {
		return nil, fmt.Errorf("compilation failed: %w", err)
	}

	// Extract result
	resultObj := result.ToObject(sh.vm)
	if resultObj == nil {
		return nil, fmt.Errorf("compilation returned nil result")
	}

	js := ""
	if jsValue := resultObj.Get("js"); jsValue != nil {
		if jsObj := jsValue.ToObject(sh.vm); jsObj != nil {
			if code := jsObj.Get("code"); code != nil {
				js = code.String()
			}
		}
	}

	css := ""
	if cssValue := resultObj.Get("css"); cssValue != nil {
		if cssObj := cssValue.ToObject(sh.vm); cssObj != nil {
			if code := cssObj.Get("code"); code != nil {
				css = code.String()
			}
		}
	}

	return &SvelteCompileResult{
		JS:  js,
		CSS: css,
	}, nil
}

func (sh *SvelteHandler) Handle(route Route) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read the Svelte file
		content, err := sh.fs.ReadFile(route.FilePath)
		if err != nil {
			http.Error(w, "Svelte file not found", http.StatusNotFound)
			return
		}
		contentStr := string(content)
		contentHash := sh.calculateMD5(contentStr)
		configHash := sh.calculateConfigHash()

		// Collect all dependencies
		allComponents, err := sh.collectAllDependencies(route.FilePath, nil)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to collect dependencies: %v", err), http.StatusInternalServerError)
			return
		}

		// Check if any dependency has changed
		dependenciesChanged := false
		var dependencies []string
		for _, comp := range allComponents {
			dependencies = append(dependencies, comp.FilePath)
			// Check if this component has been modified
			stat, err := sh.fs.Stat(comp.FilePath)
			if err == nil && stat.ModTime().After(comp.LastModified) {
				dependenciesChanged = true
				break
			}
		}

		// Check cache first
		sh.cacheMu.RLock()
		cached, exists := sh.cache[route.FilePath]
		sh.cacheMu.RUnlock()

		// If cached and both content hash and config match, and no dependencies changed
		if exists && cached.ContentHash == contentHash && cached.ConfigHash == configHash && !dependenciesChanged {
			// Verify all dependencies are still the same
			sameDepenencies := len(cached.Dependencies) == len(dependencies)
			if sameDepenencies {
				for i, dep := range cached.Dependencies {
					if i >= len(dependencies) || dep != dependencies[i] {
						sameDepenencies = false
						break
					}
				}
			}

			if sameDepenencies {
				log.Printf("Serving cached Svelte component: %s (hash: %s)", route.FilePath, contentHash)
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Header().Set("X-Svelte-Cached", "true")
				w.Write([]byte(cached.HTML))
				return
			}
		}

		log.Printf("Compiling Svelte component: %s (hash: %s) with %d dependencies", route.FilePath, contentHash, len(allComponents)-1)

		// Compile the Svelte component
		result, err := sh.compileSvelte(contentStr, route.FilePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Svelte compilation error: %v", err), http.StatusInternalServerError)
			return
		}

		// Generate HTML response with embedded runtime and all dependencies
		html := sh.generateHTMLWithRuntime(result, filepath.Base(route.FilePath), contentStr, allComponents, route.FilePath)

		// Cache the result with MD5 hash, config, and dependencies
		sh.cacheMu.Lock()
		sh.cache[route.FilePath] = &CachedResult{
			HTML:         html,
			ContentHash:  contentHash,
			ConfigHash:   configHash,
			Dependencies: dependencies,
		}
		sh.cacheMu.Unlock()

		// Send the response
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Svelte-Cached", "false")
		w.Write([]byte(html))
	}
}

func (sh *SvelteHandler) calculateMD5(content string) string {
	hash := md5.Sum([]byte(content))
	return hex.EncodeToString(hash[:])
}

// toComponentClassName converts a component filename to the class name format used by Svelte
// Examples:
// - "hello-world" -> "Hello_world"
// - "my-component" -> "My_component"
// - "simple" -> "Simple"
func (sh *SvelteHandler) toComponentClassName(componentName string) string {
	// Replace hyphens with underscores
	className := strings.ReplaceAll(componentName, "-", "_")

	// Capitalize the first letter
	if len(className) > 0 {
		className = strings.ToUpper(className[:1]) + className[1:]
	}

	return className
}

// extractActualClassName extracts the actual component class name from compiled Svelte JS
func (sh *SvelteHandler) extractActualClassName(compiledJS string) string {
	// Look for "class ClassName extends SvelteComponent"
	classRegex := regexp.MustCompile(`class\s+([A-Za-z_][A-Za-z0-9_]*)\s+extends\s+SvelteComponent`)
	matches := classRegex.FindStringSubmatch(compiledJS)
	if len(matches) > 1 {
		return matches[1]
	}

	// Fallback: return empty string if no class found
	return ""
}

// collectAllDependencies recursively collects all component dependencies
func (sh *SvelteHandler) collectAllDependencies(componentPath string, visited map[string]bool) ([]*ComponentInfo, error) {
	if visited == nil {
		visited = make(map[string]bool)
	}

	// Avoid circular dependencies
	if visited[componentPath] {
		return nil, nil
	}
	visited[componentPath] = true

	// Get component info
	info, err := sh.getComponentInfo(componentPath)
	if err != nil {
		return nil, err
	}

	var allComponents []*ComponentInfo

	// Process dependencies first (for proper load order)
	for _, dep := range info.Dependencies {
		depComponents, err := sh.collectAllDependencies(dep, visited)
		if err != nil {
			return nil, err
		}
		allComponents = append(allComponents, depComponents...)
	}

	// Add this component last
	allComponents = append(allComponents, info)

	return allComponents, nil
}

// transformToIIFE transforms ES6 module code to IIFE, preserving imports for resolution
func (sh *SvelteHandler) transformToIIFE(jsCode string, componentClassName string, imports map[string]string, componentPath string) string {
	// Use the common import transformer to handle all imports
	var componentImports map[string]string
	jsCode, componentImports = sh.importTransformer.TransformImports(jsCode, componentPath, []string{".svelte"})

	// Merge Svelte component imports into the imports map
	for name, path := range componentImports {
		imports[name] = path
	}

	// Remove ALL Svelte framework imports (they'll be provided by runtime)
	// This includes svelte/internal, svelte/store, etc.
	svelteImportRegex := regexp.MustCompile(`import\s*{[^}]*}\s*from\s*["']svelte[^"']*["'];?\s*`)
	jsCode = svelteImportRegex.ReplaceAllString(jsCode, "")

	// Also remove any bare svelte imports
	bareSvelteImportRegex := regexp.MustCompile(`import\s+\w+\s+from\s*["']svelte[^"']*["'];?\s*`)
	jsCode = bareSvelteImportRegex.ReplaceAllString(jsCode, "")

	// Remove export default statement - handle both forms:
	// 1. export default ComponentName;
	// 2. export default ComponentName at end of file
	exportRegex1 := regexp.MustCompile(`export\s+default\s+` + componentClassName + `\s*;?\s*$`)
	jsCode = exportRegex1.ReplaceAllString(jsCode, "")

	// Also handle any remaining export default statements
	exportRegex2 := regexp.MustCompile(`export\s+default\s+\w+\s*;?\s*$`)
	jsCode = exportRegex2.ReplaceAllString(jsCode, "")

	return jsCode
}

func (sh *SvelteHandler) generateHTMLWithRuntime(result *SvelteCompileResult, componentName string, svelteSource string, allComponents []*ComponentInfo, componentPath string) string {
	// Remove .svelte extension
	componentName = strings.TrimSuffix(componentName, ".svelte")

	// Find the main component in allComponents to get the actual class name
	var componentClassName string
	for _, comp := range allComponents {
		if comp.FilePath == componentPath {
			componentClassName = comp.ClassName
			break
		}
	}

	// Fallback to calculated name if not found in components
	if componentClassName == "" {
		componentClassName = sh.toComponentClassName(componentName)
	}

	// Build import mappings for main component
	imports := make(map[string]string)

	// Transform the ES6 module code to IIFE
	jsCode := sh.transformToIIFE(result.JS, componentClassName, imports, componentPath)

	// Minify component code if enabled (only the component code, not mounting logic)
	jsCode = sh.minifyComponentCode(jsCode)

	// Minify CSS if enabled
	css := sh.minifyCSS(result.CSS)

	// Extract Vimesh Style CSS if enabled
	var vimeshCSS string
	var vimeshScript string
	if sh.config.VimeshStyle != nil && sh.config.VimeshStyle.Enable {
		// Extract CSS from the Svelte source
		extractedCSS, err := utils.GetCSSFromHTML(svelteSource)
		if err != nil {
			log.Printf("Failed to extract Vimesh Style CSS: %v", err)
		} else if extractedCSS != "" {
			vimeshCSS = fmt.Sprintf(`<style id="vimesh-styles">%s</style>`, extractedCSS)
			// Add Vimesh Style runtime script
			vimeshScript = fmt.Sprintf(`<script src="%s"></script>`, sh.config.VimeshStylePath)
		}
	}

	// Add async library script if enabled
	var asyncScript string
	if sh.config.EnableAsyncLoading {
		asyncScript = fmt.Sprintf(`<script src="%s"></script>`, sh.config.AsyncLibraryPath)
	}

	var runtimeScript string
	if sh.config.UseExternalRuntime {
		// Use external runtime script
		runtimeScript = fmt.Sprintf(`<script src="%s"></script>`, sh.config.RuntimePath)
	} else {
		// Inline runtime script
		runtime := sh.getMinifiedRuntime()
		runtimeComment := "// Svelte runtime"
		if sh.config.MinifyRuntime && !sh.config.DevMode {
			runtimeComment += " (minified)"
		}
		runtimeScript = fmt.Sprintf(`<script>
        %s
        %s
    </script>`, runtimeComment, runtime)
	}

	// Collect all CSS from dependencies
	var allCSS strings.Builder
	allCSS.WriteString(css) // Main component CSS

	// Collect all JS from dependencies and create component registry
	var allJS strings.Builder
	var componentRegistry strings.Builder
	componentRegistry.WriteString("// Component registry\n")
	componentRegistry.WriteString("const __svelteComponents = {};\n\n")

	// Process dependencies first (they're already in correct order)
	for _, dep := range allComponents {
		if dep.ClassName == componentClassName {
			continue // Skip main component, we'll add it last
		}

		// Add CSS
		if dep.CompiledCSS != "" {
			allCSS.WriteString("\n/* Component: " + dep.Name + " */\n")
			allCSS.WriteString(sh.minifyCSS(dep.CompiledCSS))
		}

		// Transform and add JS
		depImports := make(map[string]string)
		depJS := sh.transformToIIFE(dep.CompiledJS, dep.ClassName, depImports, dep.FilePath)
		depJS = sh.minifyComponentCode(depJS)

		allJS.WriteString("// Component: " + dep.Name + "\n")
		allJS.WriteString("(function() {\n")

		// Add import references at the beginning of the component scope
		for importName, importPath := range depImports {
			resolvedPath := sh.resolveComponentPath(importPath, dep.FilePath)
			depInfo, _ := sh.getComponentInfo(resolvedPath)
			if depInfo != nil {
				allJS.WriteString("  const " + importName + " = __svelteComponents['" + depInfo.ClassName + "'];\n")
			}
		}

		allJS.WriteString(depJS)
		allJS.WriteString("\n  __svelteComponents['" + dep.ClassName + "'] = " + dep.ClassName + ";\n")
		allJS.WriteString("})();\n\n")
	}

	componentComment := "// Main component code"
	if sh.config.MinifyComponents && !sh.config.DevMode {
		componentComment += " (minified)"
	}

	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>` + componentName + `</title>
    <style>` + allCSS.String() + `</style>
    ` + vimeshCSS + `
    ` + vimeshScript + `
    ` + asyncScript + `
</head>
<body>
    <div id="app"></div>
    ` + runtimeScript + `
    <script>
        ` + componentRegistry.String() + `
        ` + allJS.String() + `
        ` + componentComment + `
        (function() {
            // Import references for main component
`
	// Add import references for main component
	for importName, importPath := range imports {
		// Find the main component's file path from allComponents
		var mainComponentPath string
		for _, comp := range allComponents {
			if comp.ClassName == componentClassName {
				mainComponentPath = comp.FilePath
				break
			}
		}

		if mainComponentPath != "" {
			resolvedPath := sh.resolveComponentPath(importPath, mainComponentPath)
			depInfo, _ := sh.getComponentInfo(resolvedPath)
			if depInfo != nil {
				html += "            const " + importName + " = __svelteComponents['" + depInfo.ClassName + "'];\n"
			}
		}
	}

	html += `            ` + jsCode + `
            
            // Mount the component
            const app = new ` + componentClassName + `({
                target: document.getElementById('app'),
                props: {}
            });
            
            // Make it available globally for debugging
            window.svelteApp = app;
        })();
    </script>
</body>
</html>`

	return html
}

// CanHandle implements ComponentHandler interface
func (sh *SvelteHandler) CanHandle(requestPath string) bool {
	// Only handle paths that contain underscore (component paths)
	if !strings.Contains(requestPath, "/_") {
		return false
	}

	// Check if the corresponding .svelte file exists
	filePath := sh.routesDir + requestPath + ".svelte"
	_, err := sh.fs.Stat(filePath)
	return err == nil
}

// ServeComponent implements ComponentHandler interface
func (sh *SvelteHandler) ServeComponent(w http.ResponseWriter, r *http.Request) {
	// Extract component path from request
	requestPath := r.URL.Path
	filePath := sh.routesDir + requestPath + ".svelte"

	// Check if file exists first
	_, err := sh.fs.Stat(filePath)
	if err != nil {
		response := AsyncComponentResponse{
			Success: false,
			Error:   "Component not found",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Reuse existing async component logic
	info, err := sh.getComponentInfo(filePath)
	if err != nil {
		response := AsyncComponentResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to compile component: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Transform component to IIFE format
	imports := make(map[string]string)
	jsCode := sh.transformToIIFE(info.CompiledJS, info.ClassName, imports, info.FilePath)

	// Wrap the component code
	jsCode = `/* Component: ` + info.Name + ` */
` + jsCode

	jsCode = sh.minifyComponentCode(jsCode)

	// Create response
	response := AsyncComponentResponse{
		Success:   true,
		Component: info.Name,
		ClassName: info.ClassName,
		JS:        jsCode,
		CSS:       sh.minifyCSS(info.CompiledCSS),
	}

	// Add dependencies if requested
	if r.URL.Query().Get("include_deps") == "true" {
		for _, depPath := range info.Dependencies {
			depInfo, err := sh.getComponentInfo(depPath)
			if err != nil {
				continue // Skip failed dependencies
			}

			depImports := make(map[string]string)
			depJS := sh.transformToIIFE(depInfo.CompiledJS, depInfo.ClassName, depImports, depInfo.FilePath)
			depJS = sh.minifyComponentCode(depJS)

			depResponse := AsyncComponentResponse{
				Success:   true,
				Component: depInfo.Name,
				ClassName: depInfo.ClassName,
				JS:        depJS,
				CSS:       sh.minifyCSS(depInfo.CompiledCSS),
			}
			response.Dependencies = append(response.Dependencies, depResponse)
		}
	}

	// Set caching headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", int(sh.config.ComponentCacheDuration.Seconds())))

	// Calculate ETag based on component content
	hash := md5.Sum([]byte(info.ContentHash))
	etag := `"` + hex.EncodeToString(hash[:]) + `"`
	w.Header().Set("ETag", etag)

	// Check if client has cached version
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	json.NewEncoder(w).Encode(response)
}
