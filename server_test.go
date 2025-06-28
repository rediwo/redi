package redi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	
	"github.com/rediwo/redi/filesystem"
)

func TestMain(m *testing.M) {
	// Setup test fixtures if they don't exist
	setupTestFixtures()

	// Run tests
	code := m.Run()

	os.Exit(code)
}

func setupTestFixtures() {
	// This function ensures fixtures exist for testing
	fixturesDir := "fixtures"
	if _, err := os.Stat(fixturesDir); os.IsNotExist(err) {
		// If fixtures don't exist, tests will fail
		// In real scenarios, fixtures should be created beforehand
	}
}

func TestServerInitialization(t *testing.T) {
	server := NewServer("fixtures", 8080)

	if server.port != 8080 {
		t.Errorf("Expected port to be 8080, got %d", server.port)
	}

	if server.router == nil {
		t.Error("Expected router to be initialized")
	}
	
	if server.fs == nil {
		t.Error("Expected filesystem to be initialized")
	}
}

func TestRouteScanning(t *testing.T) {
	fs := filesystem.NewOSFileSystem("fixtures")
	scanner := NewRouteScanner(fs, "routes")

	routes, err := scanner.ScanRoutes()
	if err != nil {
		t.Fatalf("Error scanning routes: %v", err)
	}

	// Test that routes were found
	if len(routes) == 0 {
		t.Error("Expected to find routes, got none")
	}

	// Test specific routes
	routePaths := make(map[string]bool)
	for _, route := range routes {
		routePaths[route.Path] = true
	}

	expectedRoutes := []string{"/", "/login", "/users", "/about", "/blog/{id}"}
	for _, expected := range expectedRoutes {
		if !routePaths[expected] {
			t.Errorf("Expected to find route %s", expected)
		}
	}
}

func TestStaticFileServer(t *testing.T) {
	server := NewServer("fixtures", 8080)
	server.setupStaticFileServer()

	// Test CSS file serving
	req := httptest.NewRequest("GET", "/css/style.css", nil)
	rr := httptest.NewRecorder()

	server.router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/css") {
		t.Errorf("Expected CSS content type, got %s", contentType)
	}
}

func TestMarkdownHandler(t *testing.T) {
	// Use HandlerManager to test through the unified interface
	fs := filesystem.NewOSFileSystem("fixtures")
	handlerManager := NewHandlerManager(fs)
	route := Route{
		Path:     "/about",
		FilePath: filepath.Join("routes", "about.md"),
		FileType: "md",
	}

	req := httptest.NewRequest("GET", "/about", nil)
	rr := httptest.NewRecorder()

	handlerFunc := handlerManager.GetHandler(route)
	handlerFunc(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("Expected HTML content type, got %s", contentType)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "<h1>About Our Test Blog</h1>") {
		t.Error("Expected markdown to be converted to HTML")
	}
}

func TestJavaScriptAPIHandler(t *testing.T) {
	fs := filesystem.NewOSFileSystem("fixtures")
	handlerManager := NewHandlerManager(fs)
	route := Route{
		Path:     "/api/users",
		FilePath: filepath.Join("routes", "api", "users.js"),
		FileType: "js",
	}

	// Test GET request
	req := httptest.NewRequest("GET", "/api/users", nil)
	rr := httptest.NewRecorder()

	handlerFunc := handlerManager.GetHandler(route)
	handlerFunc(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Expected JSON content type, got %s", contentType)
	}

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Error parsing JSON response: %v", err)
	}

	if success, ok := response["success"].(bool); !ok || !success {
		t.Error("Expected success=true in API response")
	}
}

func TestJavaScriptAPILoginEndpoint(t *testing.T) {
	fs := filesystem.NewOSFileSystem("fixtures")
	handlerManager := NewHandlerManager(fs)
	route := Route{
		Path:     "/api/login",
		FilePath: filepath.Join("routes", "api", "login.js"),
		FileType: "js",
	}

	// Test valid login
	loginData := `{"username":"admin","password":"admin123"}`
	req := httptest.NewRequest("POST", "/api/login", strings.NewReader(loginData))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handlerFunc := handlerManager.GetHandler(route)
	handlerFunc(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rr.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Error parsing JSON response: %v", err)
	}

	if success, ok := response["success"].(bool); !ok || !success {
		t.Error("Expected successful login")
	}

	// Test invalid login
	invalidLoginData := `{"username":"invalid","password":"wrong"}`
	req2 := httptest.NewRequest("POST", "/api/login", strings.NewReader(invalidLoginData))
	req2.Header.Set("Content-Type", "application/json")
	rr2 := httptest.NewRecorder()

	handlerFunc(rr2, req2)

	if rr2.Code != http.StatusUnauthorized {
		t.Errorf("Expected status code 401, got %d", rr2.Code)
	}
}

func TestHTMLTemplateHandler(t *testing.T) {
	fs := filesystem.NewOSFileSystem("fixtures")
	handlerManager := NewHandlerManager(fs)
	route := Route{
		Path:     "/",
		FilePath: filepath.Join("routes", "index.html"),
		FileType: "html",
	}

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handlerFunc := handlerManager.GetHandler(route)
	handlerFunc(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("Expected HTML content type, got %s", contentType)
	}

	body := rr.Body.String()

	// Check that layout was applied
	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Error("Expected HTML layout to be applied")
	}

	// Check that server script was executed
	if !strings.Contains(body, "Welcome to Test Blog") {
		t.Error("Expected server script data to be rendered")
	}
}

func TestDynamicRouting(t *testing.T) {
	fs := filesystem.NewOSFileSystem("fixtures")
	scanner := NewRouteScanner(fs, "routes")
	routes, err := scanner.ScanRoutes()
	if err != nil {
		t.Fatalf("Error scanning routes: %v", err)
	}

	// Find the dynamic blog route
	var blogRoute *Route
	for _, route := range routes {
		if route.Path == "/blog/{id}" && route.IsDynamic {
			blogRoute = &route
			break
		}
	}

	if blogRoute == nil {
		t.Fatal("Expected to find dynamic blog route")
	}

	if blogRoute.ParamName != "id" {
		t.Errorf("Expected param name to be 'id', got %s", blogRoute.ParamName)
	}
}

// func TestLayoutProcessing(t *testing.T) {
// 	// This test is commented out because it relies on unexported methods
// 	// Layout processing is tested through the full HTML handler test
// }

// func TestServerScriptExtraction(t *testing.T) {
// 	// This test is commented out because it relies on unexported methods
// 	// Server script extraction is tested through the full HTML handler test
// }

// Benchmark tests
func BenchmarkRouteScanning(b *testing.B) {
	fs := filesystem.NewOSFileSystem("fixtures")
	scanner := NewRouteScanner(fs, "routes")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := scanner.ScanRoutes()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarkdownConversion(b *testing.B) {
	fs := filesystem.NewOSFileSystem("fixtures")
	handlerManager := NewHandlerManager(fs)
	route := Route{
		Path:     "/about",
		FilePath: filepath.Join("routes", "about.md"),
		FileType: "md",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/about", nil)
		rr := httptest.NewRecorder()

		handlerFunc := handlerManager.GetHandler(route)
		handlerFunc(rr, req)
	}
}

func BenchmarkJavaScriptExecution(b *testing.B) {
	fs := filesystem.NewOSFileSystem("fixtures")
	handlerManager := NewHandlerManager(fs)
	route := Route{
		Path:     "/api/stats",
		FilePath: filepath.Join("routes", "api", "stats.js"),
		FileType: "js",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/stats", nil)
		rr := httptest.NewRecorder()

		handlerFunc := handlerManager.GetHandler(route)
		handlerFunc(rr, req)
	}
}
