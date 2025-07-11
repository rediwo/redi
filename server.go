package redi

import (
	"compress/gzip"
	"context"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rediwo/redi/filesystem"
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
}

func NewServer(root string, port int) *Server {
	return NewServerWithVersion(root, port, "")
}

func NewServerWithFS(embedFS fs.FS, port int) *Server {
	return NewServerWithFSAndVersion(embedFS, port, "")
}

func NewServerWithVersion(root string, port int, version string) *Server {
	return &Server{
		port:       port,
		router:     mux.NewRouter(),
		fs:         filesystem.NewOSFileSystem(root),
		version:    version,
		enableGzip: true,
		gzipLevel:  gzip.DefaultCompression,
	}
}

func NewServerWithFSAndVersion(embedFS fs.FS, port int, version string) *Server {
	return &Server{
		port:       port,
		router:     mux.NewRouter(),
		fs:         filesystem.NewEmbedFileSystem(embedFS),
		version:    version,
		enableGzip: true,
		gzipLevel:  gzip.DefaultCompression,
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

func (s *Server) Start() error {
	if err := s.setupRoutes(); err != nil {
		return fmt.Errorf("failed to setup routes: %w", err)
	}

	// Register additional routes from handlers before static file server
	if s.handlerManager != nil {
		s.handlerManager.RegisterAdditionalRoutes(s.router)
	}
	
	s.setupStaticFileServer()

	// Apply middleware
	handler := http.Handler(s.router)
	
	// Apply gzip compression if enabled
	if s.enableGzip {
		if s.gzipLevel == gzip.DefaultCompression {
			handler = handlers.CompressHandler(handler)
		} else {
			handler = handlers.CompressHandlerLevel(handler, s.gzipLevel)
		}
		log.Printf("Gzip compression enabled (level: %d)", s.gzipLevel)
	}

	addr := fmt.Sprintf(":%d", s.port)
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	
	log.Printf("Server listening on %s", addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) setupRoutes() error {
	routeScanner := NewRouteScanner(s.fs, "routes")
	s.handlerManager = NewHandlerManagerWithServer(s.fs, s.version, s.router)

	routes, err := routeScanner.ScanRoutes()
	if err != nil {
		return fmt.Errorf("failed to scan routes: %w", err)
	}

	for _, route := range routes {
		handler := s.handlerManager.GetHandler(route)
		s.router.HandleFunc(route.Path, handler).Methods("GET", "POST", "PUT", "DELETE")
		log.Printf("Registered route: %s -> %s", route.Path, route.FilePath)
	}

	// Register additional routes for handlers (e.g., Svelte runtime, Vimesh Style)
	s.handlerManager.RegisterAdditionalRoutes(s.router)

	return nil
}

func (s *Server) setupStaticFileServer() {
	publicFS, err := s.fs.Sub("public")
	if err != nil {
		log.Printf("Warning: No public directory found in filesystem")
		return
	}
	
	// Use unified fs.FS interface
	fileServer := http.FileServer(http.FS(publicFS.GetFS()))
	s.router.PathPrefix("/").Handler(http.StripPrefix("/", fileServer))
	log.Printf("Static file server serving from public directory")
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
