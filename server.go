package redi

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rediwo/redi/filesystem"
)

type Server struct {
	port       int
	router     *mux.Router
	fs         filesystem.FileSystem
	httpServer *http.Server
	version    string
}

func NewServer(root string, port int) *Server {
	return NewServerWithVersion(root, port, "")
}

func NewServerWithFS(embedFS fs.FS, port int) *Server {
	return NewServerWithFSAndVersion(embedFS, port, "")
}

func NewServerWithVersion(root string, port int, version string) *Server {
	return &Server{
		port:    port,
		router:  mux.NewRouter(),
		fs:      filesystem.NewOSFileSystem(root),
		version: version,
	}
}

func NewServerWithFSAndVersion(embedFS fs.FS, port int, version string) *Server {
	return &Server{
		port:    port,
		router:  mux.NewRouter(),
		fs:      filesystem.NewEmbedFileSystem(embedFS),
		version: version,
	}
}

func (s *Server) Start() error {
	if err := s.setupRoutes(); err != nil {
		return fmt.Errorf("failed to setup routes: %w", err)
	}

	s.setupStaticFileServer()

	addr := fmt.Sprintf(":%d", s.port)
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}
	
	log.Printf("Server listening on %s", addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) setupRoutes() error {
	routeScanner := NewRouteScanner(s.fs, "routes")
	handlerManager := NewHandlerManagerWithVersion(s.fs, s.version)

	routes, err := routeScanner.ScanRoutes()
	if err != nil {
		return fmt.Errorf("failed to scan routes: %w", err)
	}

	for _, route := range routes {
		handler := handlerManager.GetHandler(route)
		s.router.HandleFunc(route.Path, handler).Methods("GET", "POST", "PUT", "DELETE")
		log.Printf("Registered route: %s -> %s", route.Path, route.FilePath)
	}

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
