package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rediwo/redi/filesystem"
)

func TestTemplateHandler_Handle_HTML(t *testing.T) {
	// Create a mock filesystem with HTML template
	fs := filesystem.NewMemoryFileSystem()
	fs.WriteFile("routes/test.html", []byte(`<h1>{{.Title}}</h1><p>{{.Message}}</p>`))

	handler := NewTemplateHandler(fs)
	route := Route{Path: "/test", FilePath: "routes/test.html"}

	req := httptest.NewRequest("GET", "/test", nil)
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
	if !strings.Contains(body, "<h1>") {
		t.Errorf("Expected HTML content, got: %s", body)
	}
}

func TestTemplateHandler_Handle_Markdown(t *testing.T) {
	// Create a mock filesystem with Markdown template
	fs := filesystem.NewMemoryFileSystem()
	fs.WriteFile("routes/test.md", []byte(`# {{.Title}}\n\n{{.Content}}`))

	handler := NewTemplateHandler(fs)
	route := Route{Path: "/test", FilePath: "routes/test.md"}

	req := httptest.NewRequest("GET", "/test", nil)
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
	if !strings.Contains(body, "<h1>") {
		t.Errorf("Expected HTML content with <h1> tag, got: %s", body)
	}
}

func TestTemplateHandler_Handle_JSON(t *testing.T) {
	// Create a mock filesystem with JSON template
	fs := filesystem.NewMemoryFileSystem()
	fs.WriteFile("routes/test.json", []byte(`{"title": "{{.Title}}", "data": {{.Data}}}`))

	handler := NewTemplateHandler(fs)
	route := Route{Path: "/test", FilePath: "routes/test.json"}

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.Handle(route)(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	expectedContentType := "application/json; charset=utf-8"
	if contentType := w.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("Expected Content-Type %s, got %s", expectedContentType, contentType)
	}
}

func TestTemplateHandler_Handle_FileNotFound(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	handler := NewTemplateHandler(fs)
	route := Route{Path: "/nonexistent", FilePath: "routes/nonexistent.html"}

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()

	handler.Handle(route)(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestTemplateHandler_RenderTemplate_WithData(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	handler := NewTemplateHandler(fs)

	templateContent := `<h1>{{.Title}}</h1><p>{{.Message}}</p>`
	data := map[string]interface{}{
		"Title":   "Test Title",
		"Message": "Test Message",
	}

	w := httptest.NewRecorder()
	err := handler.RenderTemplate("test.html", templateContent, data, w)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Test Title") {
		t.Errorf("Expected 'Test Title' in output, got: %s", body)
	}
	if !strings.Contains(body, "Test Message") {
		t.Errorf("Expected 'Test Message' in output, got: %s", body)
	}
}

func TestTemplateHandler_ProcessLayouts(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	// Create base layout
	fs.WriteFile("routes/_layout/base.html", []byte(`<html><body>{{.Content}}</body></html>`))

	handler := NewTemplateHandler(fs)

	content := `{{layout 'base'}}<h1>Page Content</h1>`
	result, err := handler.processLayouts(content)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expected := `<html><body><h1>Page Content</h1></body></html>`
	if result != expected {
		t.Errorf("Expected: %s, got: %s", expected, result)
	}
}

func TestTemplateHandler_ProcessLayouts_LayoutNotFound(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	handler := NewTemplateHandler(fs)

	content := `{{layout 'nonexistent'}}<h1>Page Content</h1>`
	_, err := handler.processLayouts(content)

	if err == nil {
		t.Error("Expected error for missing layout, got none")
	}

	if !strings.Contains(err.Error(), "layout file not found") {
		t.Errorf("Expected 'layout file not found' error, got: %v", err)
	}
}

func TestTemplateHandler_GuessContentType(t *testing.T) {
	handler := NewTemplateHandler(nil)

	tests := []struct {
		content  string
		expected string
	}{
		{`{"key": "value"}`, "application/json; charset=utf-8"},
		{`[{"key": "value"}]`, "application/json; charset=utf-8"},
		{`# Heading`, "text/markdown; charset=utf-8"},
		{`## Another heading`, "text/markdown; charset=utf-8"},
		{`<!DOCTYPE html>`, "text/html; charset=utf-8"},
		{`<html>`, "text/html; charset=utf-8"},
		{`Plain text content`, "text/plain; charset=utf-8"},
	}

	for _, test := range tests {
		result := handler.guessContentType(test.content)
		if result != test.expected {
			t.Errorf("For content '%s', expected %s, got %s", test.content, test.expected, result)
		}
	}
}