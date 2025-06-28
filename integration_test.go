package redi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Integration tests that test the complete server functionality
func TestServerIntegration(t *testing.T) {
	// Skip if fixtures don't exist
	if _, err := os.Stat("fixtures"); os.IsNotExist(err) {
		t.Skip("Fixtures directory not found, skipping integration tests")
	}

	// Start server in background
	server := NewServer("fixtures", 8899) // Use different port for testing

	go func() {
		if err := server.Start(); err != nil {
			t.Errorf("Server failed to start: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Test cases
	testCases := []struct {
		name           string
		path           string
		method         string
		body           string
		expectedStatus int
		expectedType   string
		expectedText   string
	}{
		{
			name:           "Homepage",
			path:           "/",
			method:         "GET",
			expectedStatus: 200,
			expectedType:   "text/html",
			expectedText:   "Welcome to Test Blog",
		},
		{
			name:           "Static CSS",
			path:           "/css/style.css",
			method:         "GET",
			expectedStatus: 200,
			expectedType:   "text/css",
			expectedText:   "font-family",
		},
		{
			name:           "Static JavaScript",
			path:           "/js/main.js",
			method:         "GET",
			expectedStatus: 200,
			expectedType:   "javascript",
			expectedText:   "DOMContentLoaded",
		},
		{
			name:           "Markdown Page",
			path:           "/about",
			method:         "GET",
			expectedStatus: 200,
			expectedType:   "text/html",
			expectedText:   "About Our Test Blog",
		},
		{
			name:           "API Users",
			path:           "/api/users",
			method:         "GET",
			expectedStatus: 200,
			expectedType:   "application/json",
			expectedText:   "John Doe",
		},
		{
			name:           "API Stats",
			path:           "/api/stats",
			method:         "GET",
			expectedStatus: 200,
			expectedType:   "application/json",
			expectedText:   "Redi Frontend Server",
		},
		{
			name:           "Dynamic Route",
			path:           "/blog/123",
			method:         "GET",
			expectedStatus: 200,
			expectedType:   "text/html",
			expectedText:   "Dynamic Route Example",
		},
		{
			name:           "Admin Dashboard",
			path:           "/admin",
			method:         "GET",
			expectedStatus: 200,
			expectedType:   "text/html",
			expectedText:   "Dashboard",
		},
		{
			name:           "404 Not Found",
			path:           "/nonexistent",
			method:         "GET",
			expectedStatus: 404,
		},
	}

	baseURL := "http://localhost:8899"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url := baseURL + tc.path

			var resp *http.Response
			var err error

			if tc.method == "GET" {
				resp, err = http.Get(url)
			} else {
				req, _ := http.NewRequest(tc.method, url, strings.NewReader(tc.body))
				if tc.body != "" {
					req.Header.Set("Content-Type", "application/json")
				}
				client := &http.Client{}
				resp, err = client.Do(req)
			}

			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			// Check status code
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}

			// Check content type if specified
			if tc.expectedType != "" {
				contentType := resp.Header.Get("Content-Type")
				if !strings.Contains(contentType, tc.expectedType) {
					t.Errorf("Expected content type to contain %s, got %s", tc.expectedType, contentType)
				}
			}

			// Check response body if specified
			if tc.expectedText != "" {
				body := make([]byte, 4096)
				n, _ := resp.Body.Read(body)
				bodyStr := string(body[:n])

				if !strings.Contains(bodyStr, tc.expectedText) {
					t.Errorf("Expected response to contain %s, got first 500 chars: %s", tc.expectedText, bodyStr[:min(500, len(bodyStr))])
				}
			}
		})
	}
}

func TestAPIEndpointsIntegration(t *testing.T) {
	// Skip if fixtures don't exist
	if _, err := os.Stat("fixtures"); os.IsNotExist(err) {
		t.Skip("Fixtures directory not found, skipping API integration tests")
	}

	// Start server in background
	server := NewServer("fixtures", 8898) // Use different port for testing

	go func() {
		if err := server.Start(); err != nil {
			t.Errorf("Server failed to start: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	baseURL := "http://localhost:8898"

	t.Run("Login API", func(t *testing.T) {
		// Test valid login
		loginData := `{"username":"admin","password":"admin123"}`
		resp, err := http.Post(baseURL+"/api/login", "application/json",
			strings.NewReader(loginData))

		if err != nil {
			t.Fatalf("Login request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if success, ok := result["success"].(bool); !ok || !success {
			t.Error("Expected successful login")
		}

		// Test invalid login
		invalidData := `{"username":"invalid","password":"wrong"}`
		resp2, err := http.Post(baseURL+"/api/login", "application/json",
			strings.NewReader(invalidData))

		if err != nil {
			t.Fatalf("Invalid login request failed: %v", err)
		}
		defer resp2.Body.Close()

		if resp2.StatusCode != 401 {
			t.Errorf("Expected status 401 for invalid login, got %d", resp2.StatusCode)
		}
	})

	t.Run("Users API", func(t *testing.T) {
		// Test GET users
		resp, err := http.Get(baseURL + "/api/users")
		if err != nil {
			t.Fatalf("GET users request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if success, ok := result["success"].(bool); !ok || !success {
			t.Error("Expected successful response")
		}

		if data, ok := result["data"].([]interface{}); !ok || len(data) == 0 {
			t.Error("Expected user data in response")
		}

		// Test POST new user
		userData := `{"name":"Test User","email":"test@example.com","role":"user"}`
		resp2, err := http.Post(baseURL+"/api/users", "application/json",
			strings.NewReader(userData))

		if err != nil {
			t.Fatalf("POST user request failed: %v", err)
		}
		defer resp2.Body.Close()

		if resp2.StatusCode != 201 {
			t.Errorf("Expected status 201 for new user, got %d", resp2.StatusCode)
		}
	})

	t.Run("Posts API", func(t *testing.T) {
		// Test GET posts
		resp, err := http.Get(baseURL + "/api/posts")
		if err != nil {
			t.Fatalf("GET posts request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Test GET posts with filter
		resp2, err := http.Get(baseURL + "/api/posts?status=published")
		if err != nil {
			t.Fatalf("GET filtered posts request failed: %v", err)
		}
		defer resp2.Body.Close()

		if resp2.StatusCode != 200 {
			t.Errorf("Expected status 200 for filtered posts, got %d", resp2.StatusCode)
		}

		// Test POST new post
		postData := `{"title":"Test Post","content":"Test content","author":"Test Author"}`
		resp3, err := http.Post(baseURL+"/api/posts", "application/json",
			strings.NewReader(postData))

		if err != nil {
			t.Fatalf("POST post request failed: %v", err)
		}
		defer resp3.Body.Close()

		if resp3.StatusCode != 201 {
			t.Errorf("Expected status 201 for new post, got %d", resp3.StatusCode)
		}
	})

	t.Run("Stats API", func(t *testing.T) {
		// Test GET all stats
		resp, err := http.Get(baseURL + "/api/stats")
		if err != nil {
			t.Fatalf("GET stats request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if data, ok := result["data"].(map[string]interface{}); !ok {
			t.Error("Expected stats data in response")
		} else {
			// Check that expected stat categories exist
			expectedCategories := []string{"server", "content", "traffic", "performance", "features"}
			for _, category := range expectedCategories {
				if _, exists := data[category]; !exists {
					t.Errorf("Expected category %s in stats", category)
				}
			}
		}

		// Test GET specific category
		resp2, err := http.Get(baseURL + "/api/stats?category=server")
		if err != nil {
			t.Fatalf("GET server stats request failed: %v", err)
		}
		defer resp2.Body.Close()

		if resp2.StatusCode != 200 {
			t.Errorf("Expected status 200 for server stats, got %d", resp2.StatusCode)
		}
	})
}

func TestDynamicRoutingIntegration(t *testing.T) {
	// Skip if fixtures don't exist
	if _, err := os.Stat("fixtures"); os.IsNotExist(err) {
		t.Skip("Fixtures directory not found, skipping dynamic routing tests")
	}

	// Start server in background
	server := NewServer("fixtures", 8897) // Use different port for testing

	go func() {
		if err := server.Start(); err != nil {
			t.Errorf("Server failed to start: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	baseURL := "http://localhost:8897"

	testCases := []struct {
		name         string
		path         string
		expectedText string
		shouldFind   bool
	}{
		{
			name:         "Blog Post 1",
			path:         "/blog/1",
			expectedText: "Welcome to our blog",
			shouldFind:   true,
		},
		{
			name:         "Blog Post 2",
			path:         "/blog/2",
			expectedText: "Getting started with Redi",
			shouldFind:   true,
		},
		{
			name:         "Blog Post 123",
			path:         "/blog/123",
			expectedText: "Dynamic Route Example",
			shouldFind:   true,
		},
		{
			name:         "Non-existent Blog Post",
			path:         "/blog/999",
			expectedText: "404 - Post Not Found",
			shouldFind:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Get(baseURL + tc.path)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if tc.shouldFind {
				if resp.StatusCode != 200 {
					t.Errorf("Expected status 200, got %d", resp.StatusCode)
				}
			} else {
				if resp.StatusCode != 404 {
					t.Errorf("Expected status 404, got %d", resp.StatusCode)
				}
			}

			body := make([]byte, 2048)
			n, _ := resp.Body.Read(body)
			bodyStr := string(body[:n])

			if !strings.Contains(bodyStr, tc.expectedText) {
				t.Errorf("Expected response to contain %s, got %s", tc.expectedText, bodyStr)
			}
		})
	}
}

// Helper function to check if server is running
func waitForServer(url string, timeout time.Duration) error {
	start := time.Now()
	for time.Since(start) < timeout {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return fmt.Errorf("server did not start within %v", timeout)
}
