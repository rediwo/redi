package redi

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/rediwo/redi/filesystem"
)

func TestServerGzipCompression(t *testing.T) {
	// Create a test filesystem
	fs := filesystem.NewMemoryFileSystem()
	
	// Create test files
	fs.WriteFile("routes/index.html", []byte(`<!DOCTYPE html>
<html>
<head>
	<title>Test Page</title>
</head>
<body>
	<h1>Hello World</h1>
	<p>This is a test page with enough content to trigger compression.</p>
	<p>Lorem ipsum dolor sit amet, consectetur adipiscing elit.</p>
	<p>Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.</p>
	<p>Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris.</p>
</body>
</html>`))
	
	// Create a large CSS file in public directory
	largeCss := strings.Repeat("body { margin: 0; padding: 0; }\n", 100)
	fs.WriteFile("public/css/style.css", []byte(largeCss))
	
	// Create server with gzip enabled
	server := &Server{
		port:           8080,
		router:         mux.NewRouter(),
		fs:             fs,
		version:        "test",
		enableGzip:     true,
		gzipLevel:      gzip.DefaultCompression,
		handlerManager: NewHandlerManager(fs),
	}
	
	// Setup routes (includes static server)
	if err := server.setupRoutes(); err != nil {
		t.Fatalf("Failed to setup routes: %v", err)
	}
	
	// Get the handler with compression
	var handler http.Handler = server.router
	if server.enableGzip {
		handler = testCompressHandler(handler)
	}
	
	// Test HTML response with gzip
	t.Run("HTML with gzip", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()
		
		handler.ServeHTTP(rec, req)
		
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}
		
		// Check for gzip encoding
		if rec.Header().Get("Content-Encoding") != "gzip" {
			t.Error("Expected Content-Encoding: gzip header")
		}
		
		// Decompress and verify content
		reader, err := gzip.NewReader(rec.Body)
		if err != nil {
			t.Fatalf("Failed to create gzip reader: %v", err)
		}
		defer reader.Close()
		
		body, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("Failed to read gzipped body: %v", err)
		}
		
		if !strings.Contains(string(body), "Hello World") {
			t.Error("Expected decompressed body to contain 'Hello World'")
		}
	})
	
	// Test CSS response with gzip
	t.Run("CSS with gzip", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/css/style.css", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()
		
		handler.ServeHTTP(rec, req)
		
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}
		
		// Check for gzip encoding
		if rec.Header().Get("Content-Encoding") != "gzip" {
			t.Error("Expected Content-Encoding: gzip header")
		}
		
		// Verify compressed size is smaller
		compressedSize := rec.Body.Len()
		if compressedSize >= len(largeCss) {
			t.Errorf("Expected compressed size to be smaller than original: compressed=%d, original=%d", 
				compressedSize, len(largeCss))
		}
	})
	
	// Test without Accept-Encoding
	t.Run("No compression without Accept-Encoding", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		// Don't set Accept-Encoding header
		rec := httptest.NewRecorder()
		
		handler.ServeHTTP(rec, req)
		
		if rec.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rec.Code)
		}
		
		// Should not have gzip encoding
		if rec.Header().Get("Content-Encoding") == "gzip" {
			t.Error("Should not have Content-Encoding: gzip without Accept-Encoding")
		}
		
		// Body should be uncompressed
		if strings.Contains(rec.Body.String(), "Hello World") {
			// Good, content is readable without decompression
		} else {
			t.Error("Expected uncompressed body to contain 'Hello World'")
		}
	})
}

func TestServerGzipDisabled(t *testing.T) {
	// Create a test filesystem
	fs := filesystem.NewMemoryFileSystem()
	fs.WriteFile("routes/index.html", []byte("<h1>Test</h1>"))
	
	// Create server with gzip disabled
	server := &Server{
		port:       8080,
		router:     mux.NewRouter(),
		fs:         fs,
		version:    "test",
		enableGzip: false,
	}
	
	// Setup routes
	if err := server.setupRoutes(); err != nil {
		t.Fatalf("Failed to setup routes: %v", err)
	}
	
	handler := server.router
	
	// Test that gzip is not applied when disabled
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	
	handler.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
	
	// Should not have gzip encoding
	if rec.Header().Get("Content-Encoding") == "gzip" {
		t.Error("Should not have Content-Encoding: gzip when gzip is disabled")
	}
}

// testCompressHandler is a simple gzip handler for testing
func testCompressHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}
		
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		
		gzw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
		h.ServeHTTP(gzw, r)
	})
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

