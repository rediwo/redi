package redi

import (
	"net/http"
	"github.com/rediwo/redi/filesystem"
	"github.com/rediwo/redi/handlers"
)

type HandlerManager struct {
	fs              filesystem.FileSystem
	markdownHandler *handlers.MarkdownHandler
	jsHandler       *handlers.JavaScriptHandler
	htmlHandler     *handlers.HTMLHandler
}

func NewHandlerManager(fs filesystem.FileSystem) *HandlerManager {
	return NewHandlerManagerWithVersion(fs, "")
}

func NewHandlerManagerWithVersion(fs filesystem.FileSystem, version string) *HandlerManager {
	return &HandlerManager{
		fs:              fs,
		markdownHandler: handlers.NewMarkdownHandler(fs),
		jsHandler:       handlers.NewJavaScriptHandlerWithVersion(fs, version),
		htmlHandler:     handlers.NewHTMLHandler(fs),
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
	case "md":
		return hm.markdownHandler.Handle(handlerRoute)
	case "js":
		return hm.jsHandler.Handle(handlerRoute)
	case "html":
		return hm.htmlHandler.Handle(handlerRoute)
	default:
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Unsupported file type", http.StatusInternalServerError)
		}
	}
}
