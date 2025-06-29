package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rediwo/redi/filesystem"
)

func TestJavaScriptHandler_Handle_API(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	fs.WriteFile("test.js", []byte(`
		exports.get = function(req, res, next) {
			res.json({method: req.method, success: true});
		};
	`))

	handler := NewJavaScriptHandler(fs)
	route := Route{FilePath: "test.js"}

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.Handle(route)(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	expectedContentType := "application/json"
	if contentType := w.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("Expected Content-Type %s, got %s", expectedContentType, contentType)
	}

	body := w.Body.String()
	if !strings.Contains(body, `"success":true`) {
		t.Errorf("Expected success in JSON response, got: %s", body)
	}
}

func TestJavaScriptHandler_Handle_RenderTemplate(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	// Create JS file that calls res.render()
	fs.WriteFile("page.js", []byte(`
		exports.get = function(req, res, next) {
			res.render({
				Title: "Test Page",
				Message: "Hello World"
			});
		};
	`))
	// Create corresponding HTML template
	fs.WriteFile("page.html", []byte(`<h1>{{.Title}}</h1><p>{{.Message}}</p>`))

	handler := NewJavaScriptHandler(fs)
	route := Route{FilePath: "page.js"}

	req := httptest.NewRequest("GET", "/page", nil)
	w := httptest.NewRecorder()

	handler.Handle(route)(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	expectedContentType := "text/html; charset=utf-8"
	if contentType := w.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("Expected Content-Type %s, got %s", expectedContentType, contentType)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Test Page") {
		t.Errorf("Expected 'Test Page' in rendered template, got: %s", body)
	}
	if !strings.Contains(body, "Hello World") {
		t.Errorf("Expected 'Hello World' in rendered template, got: %s", body)
	}
}

func TestJavaScriptHandler_Handle_RenderTemplate_NoTemplate(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	// Create JS file that calls res.render() but no template file
	fs.WriteFile("page.js", []byte(`
		exports.get = function(req, res, next) {
			res.render({Title: "Test"});
		};
	`))

	handler := NewJavaScriptHandler(fs)
	route := Route{FilePath: "page.js"}

	req := httptest.NewRequest("GET", "/page", nil)
	w := httptest.NewRecorder()

	handler.Handle(route)(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Template rendering error") {
		t.Errorf("Expected template error, got: %s", body)
	}
}

func TestJavaScriptHandler_Handle_POST(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	fs.WriteFile("api.js", []byte(`
		exports.post = function(req, res, next) {
			var data = JSON.parse(req.body);
			res.json({received: data, method: req.method});
		};
	`))

	handler := NewJavaScriptHandler(fs)
	route := Route{FilePath: "api.js"}

	body := strings.NewReader(`{"test": "data"}`)
	req := httptest.NewRequest("POST", "/api", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Handle(route)(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	responseBody := w.Body.String()
	if !strings.Contains(responseBody, `"method":"POST"`) {
		t.Errorf("Expected POST method in response, got: %s", responseBody)
	}
	if !strings.Contains(responseBody, `"test":"data"`) {
		t.Errorf("Expected posted data in response, got: %s", responseBody)
	}
}

func TestJavaScriptHandler_Handle_StatusCode(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	fs.WriteFile("error.js", []byte(`
		exports.get = function(req, res, next) {
			res.status(404);
			res.json({error: "Not found"});
		};
	`))

	handler := NewJavaScriptHandler(fs)
	route := Route{FilePath: "error.js"}

	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()

	handler.Handle(route)(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Not found") {
		t.Errorf("Expected error message in response, got: %s", body)
	}
}

func TestJavaScriptHandler_Handle_Headers(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	fs.WriteFile("headers.js", []byte(`
		exports.get = function(req, res, next) {
			res.setHeader("X-Custom-Header", "custom-value");
			res.json({headers: "set"});
		};
	`))

	handler := NewJavaScriptHandler(fs)
	route := Route{FilePath: "headers.js"}

	req := httptest.NewRequest("GET", "/headers", nil)
	w := httptest.NewRecorder()

	handler.Handle(route)(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	customHeader := w.Header().Get("X-Custom-Header")
	if customHeader != "custom-value" {
		t.Errorf("Expected custom header 'custom-value', got: %s", customHeader)
	}
}

func TestJavaScriptHandler_Handle_FileNotFound(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	handler := NewJavaScriptHandler(fs)
	route := Route{FilePath: "nonexistent.js"}

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()

	handler.Handle(route)(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestJavaScriptHandler_Handle_SyntaxError(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	fs.WriteFile("syntax-error.js", []byte(`
		exports.get = function(req, res, next) {
			// Invalid JavaScript syntax
			var x = ;
			res.json({});
		};
	`))

	handler := NewJavaScriptHandler(fs)
	route := Route{FilePath: "syntax-error.js"}

	req := httptest.NewRequest("GET", "/syntax-error", nil)
	w := httptest.NewRecorder()

	handler.Handle(route)(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestJavaScriptHandler_Handle_MethodNotAllowed(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	fs.WriteFile("get-only.js", []byte(`
		exports.get = function(req, res, next) {
			res.json({method: "get"});
		};
	`))

	handler := NewJavaScriptHandler(fs)
	route := Route{FilePath: "get-only.js"}

	req := httptest.NewRequest("POST", "/get-only", nil)
	w := httptest.NewRecorder()

	handler.Handle(route)(w, req)

	if w.Code != 405 {
		t.Errorf("Expected status %d, got %d", 405, w.Code)
	}
}

func TestJavaScriptHandler_Handle_RequireModules(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	fs.WriteFile("require-test.js", []byte(`
		var fs = require('fs');
		var path = require('path');
		
		exports.get = function(req, res, next) {
			res.json({
				hasFS: typeof fs !== 'undefined',
				hasPath: typeof path !== 'undefined'
			});
		};
	`))

	handler := NewJavaScriptHandler(fs)
	route := Route{FilePath: "require-test.js"}

	req := httptest.NewRequest("GET", "/require-test", nil)
	w := httptest.NewRecorder()

	handler.Handle(route)(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, `"hasFS":true`) {
		t.Errorf("Expected fs module to be available, got: %s", body)
	}
	if !strings.Contains(body, `"hasPath":true`) {
		t.Errorf("Expected path module to be available, got: %s", body)
	}
}

func TestJavaScriptHandler_Handle_RenderWithLayout(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	// Create JS file
	fs.WriteFile("layout-test.js", []byte(`
		exports.get = function(req, res, next) {
			res.render({
				Title: "Layout Test",
				Content: "Page content here"
			});
		};
	`))
	// Create template with layout
	fs.WriteFile("layout-test.html", []byte(`{{layout 'base'}}<h1>{{.Title}}</h1><p>{{.Content}}</p>`))
	// Create layout file
	fs.WriteFile("routes/_layout/base.html", []byte(`<html><head><title>{{.Title}}</title></head><body>{{.Content}}</body></html>`))

	handler := NewJavaScriptHandler(fs)
	route := Route{FilePath: "layout-test.js"}

	req := httptest.NewRequest("GET", "/layout-test", nil)
	w := httptest.NewRecorder()

	handler.Handle(route)(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "<html>") {
		t.Errorf("Expected HTML layout structure, got: %s", body)
	}
	if !strings.Contains(body, "Layout Test") {
		t.Errorf("Expected title in layout, got: %s", body)
	}
	if !strings.Contains(body, "Page content here") {
		t.Errorf("Expected page content, got: %s", body)
	}
}