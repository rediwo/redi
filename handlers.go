package redi

import (
	"net/http"
	"github.com/rediwo/redi/filesystem"
	"github.com/rediwo/redi/handlers"
	"github.com/gorilla/mux"
)

type HandlerManager struct {
	fs              filesystem.FileSystem
	jsHandler       *handlers.JavaScriptHandler
	templateHandler *handlers.TemplateHandler
	svelteHandler   *handlers.SvelteHandler
	errorHandler    *handlers.ErrorHandler
	routesDir       string
}

func NewHandlerManager(fs filesystem.FileSystem) *HandlerManager {
	return NewHandlerManagerWithVersion(fs, "")
}

func NewHandlerManagerWithVersion(fs filesystem.FileSystem, version string) *HandlerManager {
	templateHandler := handlers.NewTemplateHandler(fs)
	return &HandlerManager{
		fs:              fs,
		jsHandler:       handlers.NewJavaScriptHandlerWithVersion(fs, version),
		templateHandler: templateHandler,
		svelteHandler:   handlers.NewSvelteHandler(fs),
		errorHandler:    handlers.NewErrorHandler(fs, templateHandler),
	}
}

// NewHandlerManagerWithServer creates a HandlerManager with server access for route registration
func NewHandlerManagerWithServer(fs filesystem.FileSystem, version string, router *mux.Router, routesDir string) *HandlerManager {
	// Create template config with Vimesh Style enabled
	templateConfig := handlers.DefaultTemplateConfig()
	templateConfig.VimeshStyle.Enable = true
	
	templateHandler := handlers.NewTemplateHandlerWithRouter(fs, templateConfig, router)
	
	jsHandler := handlers.NewJavaScriptHandlerWithVersion(fs, version)
	errorHandler := handlers.NewErrorHandlerWithRoutesDir(fs, templateHandler, routesDir)
	
	// Set error handler on JavaScript handler
	jsHandler.SetErrorHandler(errorHandler)
	
	// Create Svelte config
	svelteConfig := handlers.DefaultSvelteConfig()
	
	return &HandlerManager{
		fs:              fs,
		jsHandler:       jsHandler,
		templateHandler: templateHandler,
		svelteHandler:   handlers.NewSvelteHandlerWithRouterAndRoutesDir(fs, svelteConfig, router, routesDir),
		errorHandler:    errorHandler,
		routesDir:       routesDir,
	}
}

// RegisterAdditionalRoutes registers any additional routes that handlers need
func (hm *HandlerManager) RegisterAdditionalRoutes(router *mux.Router) {
	if hm.svelteHandler != nil {
		hm.svelteHandler.RegisterRoutes(router)
	}
	if hm.templateHandler != nil {
		hm.templateHandler.RegisterRoutes(router)
	}
}

func (hm *HandlerManager) GetHandler(route Route) http.HandlerFunc {
	// Convert main Route to handlers.Route
	handlerRoute := handlers.Route{
		Path:      route.Path,
		FilePath:  route.FilePath,
		FileType:  route.FileType,
		IsDynamic: route.IsDynamic,
		ParamName: route.ParamName,
	}
	
	switch route.FileType {
	case "js":
		return hm.jsHandler.Handle(handlerRoute)
	case "svelte":
		return hm.svelteHandler.Handle(handlerRoute)
	default:
		// All non-component files are handled as templates (HTML, Markdown, JSON, etc.)
		return hm.templateHandler.Handle(handlerRoute)
	}
}

// ComponentRequestHandler handles component requests using dynamic matching
type ComponentRequestHandler struct {
	handlers []handlers.ComponentHandler
}

// NewComponentRequestHandler creates a new component request handler
func NewComponentRequestHandler(componentHandlers []handlers.ComponentHandler) *ComponentRequestHandler {
	return &ComponentRequestHandler{
		handlers: componentHandlers,
	}
}

// Match implements the MatcherFunc interface for gorilla/mux
func (crh *ComponentRequestHandler) Match(r *http.Request, rm *mux.RouteMatch) bool {
	requestPath := r.URL.Path
	
	// Check if any component handler can handle this request
	for _, handler := range crh.handlers {
		if handler.CanHandle(requestPath) {
			return true
		}
	}
	
	return false
}

// ServeHTTP handles the HTTP request by forwarding to the appropriate component handler
func (crh *ComponentRequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestPath := r.URL.Path
	
	// Try each component handler in order
	for _, handler := range crh.handlers {
		if handler.CanHandle(requestPath) {
			handler.ServeComponent(w, r)
			return
		}
	}
	
	// No handler found (should not happen if Match returned true)
	// This should never happen, but just in case
	http.NotFound(w, r)
}

// GetSvelteHandler returns the Svelte handler
func (hm *HandlerManager) GetSvelteHandler() *handlers.SvelteHandler {
	return hm.svelteHandler
}
