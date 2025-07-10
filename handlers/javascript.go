package handlers

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"

	"github.com/rediwo/redi/filesystem"

	_ "github.com/rediwo/redi/modules"
)

// generateSessionID creates a session identifier based on client characteristics
func generateSessionID(r *http.Request) string {
	// Use IP address and User-Agent to identify client session
	// In production, you might want to use proper session cookies
	clientIP := r.RemoteAddr
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		clientIP = realIP
	} else if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		clientIP = strings.Split(forwarded, ",")[0]
	}

	// Remove port number from IP address for consistent session identification
	if lastColon := strings.LastIndex(clientIP, ":"); lastColon != -1 {
		clientIP = clientIP[:lastColon]
	}
	// Handle IPv6 addresses wrapped in brackets
	clientIP = strings.Trim(clientIP, "[]")

	userAgent := r.Header.Get("User-Agent")
	sessionData := clientIP + "|" + userAgent

	// Create MD5 hash for shorter session ID
	hash := md5.Sum([]byte(sessionData))
	return hex.EncodeToString(hash[:])
}

type JavaScriptHandler struct {
	fs      filesystem.FileSystem
	version string
}

func NewJavaScriptHandler(fs filesystem.FileSystem) *JavaScriptHandler {
	return NewJavaScriptHandlerWithVersion(fs, "")
}

func NewJavaScriptHandlerWithVersion(fs filesystem.FileSystem, version string) *JavaScriptHandler {
	return &JavaScriptHandler{
		fs:      fs,
		version: version,
	}
}

func (jh *JavaScriptHandler) Handle(route Route) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the engine pool
		pool := GetJSEnginePool(jh.fs, jh.version)

		// Generate session ID for this client
		sessionID := generateSessionID(r)

		// Get an engine for this session
		engine, err := pool.GetEngineForSession(sessionID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get JavaScript engine: %v", err), http.StatusInternalServerError)
			return
		}

		// Note: We don't return the engine to pool immediately since it's session-bound
		// Session engines are managed by the pool and cleaned up separately

		// Execute the HTTP method handler
		if err := engine.ExecuteHTTPMethod(r, w, route); err != nil {
			// Handle different types of errors appropriately
			errMsg := err.Error()
			if strings.Contains(errMsg, "failed to read file") || strings.Contains(errMsg, "failed to stat file") || strings.Contains(errMsg, "no such file") || strings.Contains(errMsg, "file does not exist") {
				http.Error(w, fmt.Sprintf("JavaScript file not found: %s (Path: %s, Method: %s)", route.FilePath, r.URL.Path, r.Method), http.StatusNotFound)
			} else if strings.Contains(errMsg, "failed to compile") || strings.Contains(errMsg, "parsing error") {
				http.Error(w, fmt.Sprintf("JavaScript syntax error in %s: %v", route.FilePath, err), http.StatusInternalServerError)
			} else if methodName, found := strings.CutPrefix(errMsg, "method_not_allowed:"); found {
				http.Error(w, fmt.Sprintf("Method %s not allowed for %s (available methods can be checked in %s)", strings.ToUpper(methodName), r.URL.Path, route.FilePath), http.StatusMethodNotAllowed)
			} else {
				http.Error(w, fmt.Sprintf("JavaScript execution error in %s: %v", route.FilePath, err), http.StatusInternalServerError)
			}
		}
	}
}
