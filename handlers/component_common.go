package handlers

import (
	"net/http"
)

// ComponentHandler defines the interface for handling component requests
type ComponentHandler interface {
	// CanHandle determines if this handler can process the given request path
	CanHandle(requestPath string) bool
	// ServeComponent handles the component request
	ServeComponent(w http.ResponseWriter, r *http.Request)
}