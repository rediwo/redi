package redi

import (
	"net/http"
	"github.com/rediwo/redi/filesystem"
	"github.com/rediwo/redi/handlers"
)

type HandlerManager struct {
	fs              filesystem.FileSystem
	jsHandler       *handlers.JavaScriptHandler
	templateHandler *handlers.TemplateHandler
}

func NewHandlerManager(fs filesystem.FileSystem) *HandlerManager {
	return NewHandlerManagerWithVersion(fs, "")
}

func NewHandlerManagerWithVersion(fs filesystem.FileSystem, version string) *HandlerManager {
	return &HandlerManager{
		fs:              fs,
		jsHandler:       handlers.NewJavaScriptHandlerWithVersion(fs, version),
		templateHandler: handlers.NewTemplateHandler(fs),
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
	default:
		// All non-.js files are handled as templates (HTML, Markdown, JSON, etc.)
		return hm.templateHandler.Handle(handlerRoute)
	}
}
