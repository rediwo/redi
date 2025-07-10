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
}

func NewHandlerManager(fs filesystem.FileSystem) *HandlerManager {
	return NewHandlerManagerWithVersion(fs, "")
}

func NewHandlerManagerWithVersion(fs filesystem.FileSystem, version string) *HandlerManager {
	return &HandlerManager{
		fs:              fs,
		jsHandler:       handlers.NewJavaScriptHandlerWithVersion(fs, version),
		templateHandler: handlers.NewTemplateHandler(fs),
		svelteHandler:   handlers.NewSvelteHandler(fs),
	}
}

// NewHandlerManagerWithServer creates a HandlerManager with server access for route registration
func NewHandlerManagerWithServer(fs filesystem.FileSystem, version string, router *mux.Router) *HandlerManager {
	return &HandlerManager{
		fs:              fs,
		jsHandler:       handlers.NewJavaScriptHandlerWithVersion(fs, version),
		templateHandler: handlers.NewTemplateHandler(fs),
		svelteHandler:   handlers.NewSvelteHandlerWithRouter(fs, handlers.DefaultSvelteConfig(), router),
	}
}

// RegisterAdditionalRoutes registers any additional routes that handlers need
func (hm *HandlerManager) RegisterAdditionalRoutes(router *mux.Router) {
	if hm.svelteHandler != nil {
		hm.svelteHandler.RegisterRoutes(router)
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
		// All non-.js/.svelte files are handled as templates (HTML, Markdown, JSON, etc.)
		return hm.templateHandler.Handle(handlerRoute)
	}
}
