package redi

import (
	"compress/gzip"
	"context"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rediwo/redi/cache"
	"github.com/rediwo/redi/filesystem"
	"github.com/rediwo/redi/logging"
	rediHandlers "github.com/rediwo/redi/handlers"
)

type Server struct {
	port           int
	router         *mux.Router
	fs             filesystem.FileSystem
	httpServer     *http.Server
	version        string
	handlerManager *HandlerManager
	enableGzip     bool
	gzipLevel      int
	routesDir      string
	cacheManager   *cache.CacheManager
	svelteCache    *cache.SvelteCache
	enableCache    bool
}

func NewServer(root string, port int) *Server {
	return NewServerWithVersion(root, port, "")
}

func NewServerWithFS(embedFS fs.FS, port int) *Server {
	return NewServerWithFSAndVersion(embedFS, port, "")
}

func NewServerWithVersion(root string, port int, version string) *Server {
	return &Server{
		port:        port,
		router:      mux.NewRouter(),
		fs:          filesystem.NewOSFileSystem(root),
		version:     version,
		enableGzip:  true,
		gzipLevel:   gzip.DefaultCompression,
		routesDir:   "routes",
		enableCache: true,
	}
}

func NewServerWithFSAndVersion(embedFS fs.FS, port int, version string) *Server {
	return &Server{
		port:        port,
		router:      mux.NewRouter(),
		fs:          filesystem.NewEmbedFileSystem(embedFS),
		version:     version,
		enableGzip:  true,
		gzipLevel:   gzip.DefaultCompression,
		routesDir:   "routes",
		enableCache: true,
	}
}

// SetGzipEnabled configures whether gzip compression is enabled
func (s *Server) SetGzipEnabled(enabled bool) {
	s.enableGzip = enabled
}

// SetGzipLevel sets the gzip compression level (-1 to 9)
func (s *Server) SetGzipLevel(level int) {
	if level >= -1 && level <= 9 {
		s.gzipLevel = level
	}
}

// SetRoutesDir sets the directory for routes
func (s *Server) SetRoutesDir(dir string) {
	s.routesDir = dir
}

// SetCacheEnabled configures whether compilation caching is enabled
func (s *Server) SetCacheEnabled(enabled bool) {
	s.enableCache = enabled
}

// initializeCache initializes the cache system if enabled
func (s *Server) initializeCache() error {
	if !s.enableCache {
		logging.Info("Cache is disabled")
		return nil
	}

	logging.Info("Initializing cache system")
	
	// Get the root directory from the filesystem
	rootDir := s.fs.GetRoot()
	logging.Debug("Cache root directory", "path", rootDir)

	// Create cache configuration
	cacheConfig := &cache.CacheConfig{
		RootDir:       rootDir,
		Enabled:       true,
		MaxSize:       500 * 1024 * 1024, // 500MB default
		TTL:           0,                  // No expiration by default
		CompressCache: false,
	}

	// Initialize cache manager
	s.cacheManager = cache.NewCacheManager(cacheConfig)
	s.cacheManager.SetFileSystem(s.fs)

	if err := s.cacheManager.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize cache: %w", err)
	}

	// Create Svelte cache
	s.svelteCache = cache.NewSvelteCache(s.cacheManager, s.fs)

	logging.Info("Cache system initialized", "location", rootDir+"/.redi")
	return nil
}

func (s *Server) Start() error {
	if err := s.setupRoutes(); err != nil {
		return fmt.Errorf("failed to setup routes: %w", err)
	}

	// Apply middleware
	handler := http.Handler(s.router)
	
	// Apply gzip compression if enabled
	if s.enableGzip {
		if s.gzipLevel == gzip.DefaultCompression {
			handler = handlers.CompressHandler(handler)
		} else {
			handler = handlers.CompressHandlerLevel(handler, s.gzipLevel)
		}
		logging.Info("Gzip compression enabled", "level", s.gzipLevel)
	}

	addr := fmt.Sprintf(":%d", s.port)
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	
	logging.Info("Server listening", "address", addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) setupRoutes() error {
	// Initialize cache if enabled
	if err := s.initializeCache(); err != nil {
		logging.Warn("Failed to initialize cache", "error", err)
		// Continue without cache
	}

	routeScanner := NewRouteScanner(s.fs, s.routesDir)
	s.handlerManager = NewHandlerManagerWithServer(s.fs, s.version, s.router, s.routesDir)

	// Set persistent cache on Svelte handler if available
	if s.svelteCache != nil && s.handlerManager.svelteHandler != nil {
		s.handlerManager.svelteHandler.SetPersistentCache(s.svelteCache)
	}

	// Set custom 404 handler using the error handler
	s.router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.handlerManager.errorHandler.Handle404(w, r)
	})

	routes, err := routeScanner.ScanRoutes()
	if err != nil {
		return fmt.Errorf("failed to scan routes: %w", err)
	}

	for _, route := range routes {
		handler := s.handlerManager.GetHandler(route)
		s.router.HandleFunc(route.Path, handler).Methods("GET", "POST", "PUT", "DELETE", "HEAD")
		logging.Debug("Registered route", "path", route.Path, "file", route.FilePath, "type", route.FileType)
		
		// Only register with trailing slash for index files (but not the root "/")
		if route.IsIndex && route.Path != "/" && route.Path != "" {
			s.router.HandleFunc(route.Path+"/", handler).Methods("GET", "POST", "PUT", "DELETE", "HEAD")
			logging.Debug("Registered route with trailing slash", "path", route.Path+"/", "file", route.FilePath)
		}
	}

	// Register additional routes for handlers (e.g., Svelte runtime, Vimesh Style)
	s.handlerManager.RegisterAdditionalRoutes(s.router)

	// Register dynamic component handler using MatcherFunc
	componentHandler := NewComponentRequestHandler([]rediHandlers.ComponentHandler{
		s.handlerManager.svelteHandler,
	})
	s.router.MatcherFunc(componentHandler.Match).HandlerFunc(componentHandler.ServeHTTP).Methods("GET", "HEAD")

	// Setup static file server last - catches remaining requests
	s.setupStaticFileServer()

	return nil
}


func (s *Server) setupStaticFileServer() {
	publicFS, err := s.fs.Sub("public")
	if err != nil {
		log.Printf("Warning: No public directory found in filesystem")
		return
	}
	
	// Custom handler that checks file existence before serving
	staticHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't handle component paths (containing /_)
		if strings.Contains(r.URL.Path, "/_") {
			s.handlerManager.errorHandler.Handle404(w, r)
			return
		}
		
		// Clean the path and check if file exists
		cleanPath := strings.TrimPrefix(r.URL.Path, "/")
		if cleanPath == "" {
			cleanPath = "index.html"
		}
		
		// Check if file exists in public directory
		file, err := publicFS.GetFS().Open(cleanPath)
		if err != nil {
			// File not found, use our custom 404 handler
			s.handlerManager.errorHandler.Handle404(w, r)
			return
		}
		file.Close()
		
		// File exists, serve it using the file server
		fileServer := http.FileServer(http.FS(publicFS.GetFS()))
		http.StripPrefix("/", fileServer).ServeHTTP(w, r)
	})
	
	s.router.PathPrefix("/").Handler(staticHandler)
	logging.Debug("Static file server enabled", "directory", "public")
}

// Stop gracefully shuts down the server
func (s *Server) Stop() error {
	if s.httpServer == nil {
		return nil
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	return s.httpServer.Shutdown(ctx)
}

// PreBuild pre-compiles all Svelte components
func (s *Server) PreBuild(parallelWorkers int) error {
	// Ensure cache is enabled
	if !s.enableCache {
		return fmt.Errorf("cache must be enabled for pre-building (use --cache flag)")
	}
	
	// Initialize cache if not already done
	if s.cacheManager == nil {
		if err := s.initializeCache(); err != nil {
			return fmt.Errorf("failed to initialize cache: %w", err)
		}
	}
	
	// Setup handlers to get the SvelteHandler
	if err := s.setupRoutes(); err != nil {
		return fmt.Errorf("failed to setup routes: %w", err)
	}
	
	// Get the SvelteHandler from the handler manager
	svelteHandler := s.handlerManager.GetSvelteHandler()
	if svelteHandler == nil {
		return fmt.Errorf("Svelte handler not initialized")
	}
	
	// Create and run the pre-builder
	preBuilder := NewServerPreBuilder(s.fs, svelteHandler, s.routesDir, parallelWorkers)
	return preBuilder.Build()
}

