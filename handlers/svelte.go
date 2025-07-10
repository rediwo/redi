package handlers

import (
	"crypto/md5"
	_ "embed"
	"encoding/hex"
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

// SvelteConfig holds all Svelte-related settings
type SvelteConfig struct {
	// Minification settings
	MinifyRuntime    bool // Enable runtime minification
	MinifyComponents bool // Enable component code minification
	MinifyCSS        bool // Enable CSS minification
	DevMode          bool // Disable minification in development
	
	// Runtime resource settings
	UseExternalRuntime   bool          // Use external runtime file instead of inline
	RuntimePath          string        // Path for runtime resource (default: "/svelte-runtime.js")
	RuntimeCacheDuration time.Duration // Cache duration for runtime resource
	
	// Vimesh Style settings
	VimeshStyle       *utils.VimeshStyleConfig // Vimesh Style configuration
	VimeshStylePath   string                   // Path for Vimesh Style resource (default: "/svelte-vimesh-style.js")
}

// DefaultSvelteConfig returns default Svelte settings
func DefaultSvelteConfig() *SvelteConfig {
	return &SvelteConfig{
		MinifyRuntime:        true,
		MinifyComponents:     true,
		MinifyCSS:           true,
		DevMode:             false,
		UseExternalRuntime:  true,
		RuntimePath:         "/svelte-runtime.js",
		RuntimeCacheDuration: 365 * 24 * time.Hour, // 1 year
		VimeshStyle:         utils.DefaultVimeshStyleConfig(),
		VimeshStylePath:     "/svelte-vimesh-style.js",
	}
}

type SvelteHandler struct {
	fs               filesystem.FileSystem
	vm               *goja.Runtime
	compileFunc      goja.Callable
	initialized      bool
	templateHandler  *TemplateHandler
	mu               sync.Mutex
	cache            map[string]*CachedResult
	cacheMu          sync.RWMutex
	config           *SvelteConfig
	minifiedRuntime  string
	runtimeMinified  bool
	runtimeMu        sync.Mutex
	minifier         *minify.M
}

type CachedResult struct {
	HTML        string
	ContentHash string // MD5 hash of the source content
	ConfigHash  string // Hash of config for cache invalidation
}

type SvelteCompileResult struct {
	JS  string `json:"js"`
	CSS string `json:"css"`
}

func NewSvelteHandler(fs filesystem.FileSystem) *SvelteHandler {
	return NewSvelteHandlerWithConfig(fs, DefaultSvelteConfig())
}

func NewSvelteHandlerWithConfig(fs filesystem.FileSystem, config *SvelteConfig) *SvelteHandler {
	// Create and configure minifier with safer settings
	m := minify.New()
	
	// Configure JavaScript minifier with much safer options for Svelte runtime
	jsMinifier := &js.Minifier{
		Precision: 0, // Keep full precision for numbers
		KeepVarNames: true, // Keep variable names to avoid breaking references
	}
	m.Add("application/javascript", jsMinifier)
	m.AddFunc("text/css", css.Minify)
	
	return &SvelteHandler{
		fs:              fs,
		templateHandler: NewTemplateHandler(fs),
		cache:           make(map[string]*CachedResult),
		config:          config,
		minifier:        m,
	}
}

// NewSvelteHandlerWithRouter creates a SvelteHandler and registers runtime routes
func NewSvelteHandlerWithRouter(fs filesystem.FileSystem, config *SvelteConfig, router *mux.Router) *SvelteHandler {
	sh := NewSvelteHandlerWithConfig(fs, config)
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
	vimeshJS := utils.GetVimeshStyleJS()
	
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

		// Check cache first
		sh.cacheMu.RLock()
		cached, exists := sh.cache[route.FilePath]
		sh.cacheMu.RUnlock()

		// If cached and both content hash and config match, return cached HTML
		if exists && cached.ContentHash == contentHash && cached.ConfigHash == configHash {
			log.Printf("Serving cached Svelte component: %s (hash: %s)", route.FilePath, contentHash)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("X-Svelte-Cached", "true")
			w.Write([]byte(cached.HTML))
			return
		}

		log.Printf("Compiling Svelte component: %s (hash: %s)", route.FilePath, contentHash)

		// Compile the Svelte component
		result, err := sh.compileSvelte(contentStr, route.FilePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Svelte compilation error: %v", err), http.StatusInternalServerError)
			return
		}

		// Generate HTML response with embedded runtime
		html := sh.generateHTMLWithRuntime(result, filepath.Base(route.FilePath), contentStr)

		// Cache the result with MD5 hash and config
		sh.cacheMu.Lock()
		sh.cache[route.FilePath] = &CachedResult{
			HTML:        html,
			ContentHash: contentHash,
			ConfigHash:  configHash,
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

func (sh *SvelteHandler) transformToIIFE(jsCode string, componentClassName string) string {
	// Remove all types of import statements
	importRegex1 := regexp.MustCompile(`import\s*{[^}]+}\s*from\s*["'][^"']+["'];?\s*`)
	jsCode = importRegex1.ReplaceAllString(jsCode, "")
	
	importRegex2 := regexp.MustCompile(`import\s+\w+\s+from\s*["'][^"']+["'];?\s*`)
	jsCode = importRegex2.ReplaceAllString(jsCode, "")
	
	importRegex3 := regexp.MustCompile(`import\s*["'][^"']+["'];?\s*`)
	jsCode = importRegex3.ReplaceAllString(jsCode, "")
	
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

func (sh *SvelteHandler) generateHTMLWithRuntime(result *SvelteCompileResult, componentName string, svelteSource string) string {
	// Remove .svelte extension
	componentName = strings.TrimSuffix(componentName, ".svelte")
	
	// Convert component name to class name format (handle hyphens)
	// vimesh-test -> Vimesh_test
	componentClassName := sh.toComponentClassName(componentName)
	
	// Transform the ES6 module code to IIFE
	jsCode := sh.transformToIIFE(result.JS, componentClassName)
	
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
	
	componentComment := "// Component code"
	if sh.config.MinifyComponents && !sh.config.DevMode {
		componentComment += " (minified)"
	}
	
	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>` + componentName + `</title>
    <style>` + css + `</style>
    ` + vimeshCSS + `
    ` + vimeshScript + `
</head>
<body>
    <div id="app"></div>
    ` + runtimeScript + `
    <script>
        ` + componentComment + `
        ` + jsCode + `
        
        // Mount the component
        const app = new ` + componentClassName + `({
            target: document.getElementById('app'),
            props: {}
        });
        
        // Make it available globally for debugging
        window.svelteApp = app;
    </script>
</body>
</html>`

	return html
}