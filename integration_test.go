package redi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/rediwo/redi/filesystem"
)

func setupIntegrationMemoryFS() *filesystem.MemoryFileSystem {
	memFS := filesystem.NewMemoryFileSystem()
	
	// Layout files
	memFS.WriteFile("routes/_layout/base.html", []byte(`<!DOCTYPE html>
<html><head><title>{{.Title}} - Test Blog</title></head>
<body><main>{{.Content}}</main></body></html>`))
	
	memFS.WriteFile("routes/_layout/admin.html", []byte(`<!DOCTYPE html>
<html><head><title>Admin - {{.Title}}</title></head>
<body><header>Admin Panel</header><main>{{.Content}}</main></body></html>`))
	
	// Index page
	memFS.WriteFile("routes/index.js", []byte(`exports.get = function(req, res, next) {
    res.render({
        Title: "Welcome to Test Blog"
    });
};`))
	
	memFS.WriteFile("routes/index.html", []byte(`{{layout 'base'}}

<h1>Welcome to Test Blog</h1>
<p>This is a test blog built with Redi frontend server.</p>`))
	
	// About markdown page
	memFS.WriteFile("routes/about.md", []byte(`# About Our Test Blog

This is a **test blog** built using the **Redi frontend server**.`))
	
	// API endpoints
	memFS.WriteFile("routes/api/users.js", []byte(`var users = [
    { id: 1, name: "John Doe", email: "john@example.com", role: "admin" },
    { id: 2, name: "Jane Smith", email: "jane@example.com", role: "editor" }
];

exports.get = function(req, res, next) {
    res.json({ success: true, data: users, count: users.length });
};

exports.post = function(req, res, next) {
    var userData = JSON.parse(req.body);
    var newUser = {
        id: users.length + 1,
        name: userData.name,
        email: userData.email,
        role: userData.role || 'user'
    };
    users.push(newUser);
    res.status(201);
    res.json({ success: true, data: newUser });
};`))
	
	memFS.WriteFile("routes/api/login.js", []byte(`exports.post = function(req, res, next) {
    var credentials = JSON.parse(req.body);
    if (credentials.username === "admin" && credentials.password === "admin123") {
        res.json({
            success: true,
            message: "Login successful",
            data: { username: "admin", role: "admin", token: "mock-token" }
        });
    } else {
        res.status(401);
        res.json({ success: false, message: "Invalid credentials" });
    }
};`))
	
	memFS.WriteFile("routes/api/stats.js", []byte(`var stats = {
    server: { name: "Redi Frontend Server", version: "1.0.0" },
    content: { totalPosts: 15, publishedPosts: 12 },
    traffic: { todayViews: 100, weeklyViews: 500 },
    performance: { averageResponseTime: 50, memoryUsage: 75 },
    features: { markdownSupport: true, jsEngineSupport: true }
};

exports.get = function(req, res, next) {
    var category = req.query ? req.query.category : null;
    if (category && stats[category]) {
        res.json({ success: true, category: category, data: stats[category] });
    } else {
        res.json({ success: true, data: stats });
    }
};`))
	
	memFS.WriteFile("routes/api/posts.js", []byte(`var posts = [
    { id: 1, title: "First Post", content: "Content 1", status: "published" },
    { id: 2, title: "Second Post", content: "Content 2", status: "draft" }
];

exports.get = function(req, res, next) {
    var status = req.query ? req.query.status : null;
    var filteredPosts = status ? posts.filter(function(p) { return p.status === status; }) : posts;
    res.json({ success: true, data: filteredPosts });
};

exports.post = function(req, res, next) {
    var postData = JSON.parse(req.body);
    var newPost = {
        id: posts.length + 1,
        title: postData.title,
        content: postData.content,
        author: postData.author,
        status: "published"
    };
    posts.push(newPost);
    res.status(201);
    res.json({ success: true, data: newPost });
};`))
	
	// Dynamic blog route
	memFS.WriteFile("routes/blog/[id].js", []byte(`var posts = {
    "1": { id: 1, title: "Welcome to our blog", content: "First blog post content" },
    "2": { id: 2, title: "Getting started with Redi", content: "Second blog post content" },
    "123": { id: 123, title: "Dynamic Route Example", content: "Example dynamic content" }
};

exports.get = function(req, res, next) {
    var postId = req.params.id;
    var post = posts[postId];

    if (!post) {
        res.status(404);
        res.render({ Title: "Post Not Found", error: "Post not found" });
    } else {
        res.render({ Title: post.title, post: post });
    }
};`))
	
	memFS.WriteFile("routes/blog/[id].html", []byte(`{{layout 'base'}}

{{if .error}}
<h1>404 - Post Not Found</h1>
<p>The requested blog post could not be found.</p>
{{else}}
<h1>{{.post.title}}</h1>
<p>{{.post.content}}</p>
{{end}}`))
	
	// Admin page
	memFS.WriteFile("routes/admin/index.js", []byte(`exports.get = function(req, res, next) {
    res.render({
        Title: "Dashboard"
    });
};`))
	
	memFS.WriteFile("routes/admin/index.html", []byte(`{{layout 'admin'}}

<h1>Admin Dashboard</h1>
<p>Welcome to the admin panel.</p>`))
	
	// Static files
	memFS.WriteFile("public/css/style.css", []byte(`body { 
    font-family: Arial, sans-serif; 
    margin: 0; 
    padding: 20px; 
}`))
	
	memFS.WriteFile("public/js/main.js", []byte(`document.addEventListener('DOMContentLoaded', function() {
    console.log('Page loaded');
});`))
	
	return memFS
}

func TestServerIntegration(t *testing.T) {
	memFS := setupIntegrationMemoryFS()
	
	// Create server with memory filesystem
	server := &Server{
		router: mux.NewRouter(),
		port:   8080,
		fs:     memFS,
	}
	
	// Setup routes and static server
	if err := server.setupRoutes(); err != nil {
		t.Fatalf("Failed to setup routes: %v", err)
	}
	server.setupStaticFileServer()
	
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
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request
			if tc.body != "" {
				req = httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}
			
			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, req)
			
			// Check status code
			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, w.Code)
			}
			
			// Check content type if specified
			if tc.expectedType != "" {
				contentType := w.Header().Get("Content-Type")
				if !strings.Contains(contentType, tc.expectedType) {
					t.Errorf("Expected content type to contain %s, got %s", tc.expectedType, contentType)
				}
			}
			
			// Check response body if specified
			if tc.expectedText != "" {
				body := w.Body.String()
				if !strings.Contains(body, tc.expectedText) {
					maxLen := 500
					if len(body) < maxLen {
						maxLen = len(body)
					}
					t.Errorf("Expected response to contain %s, got first %d chars: %s", tc.expectedText, maxLen, body[:maxLen])
				}
			}
		})
	}
}

func TestAPIEndpointsIntegration(t *testing.T) {
	memFS := setupIntegrationMemoryFS()
	
	server := &Server{
		router: mux.NewRouter(),
		port:   8080,
		fs:     memFS,
	}
	
	if err := server.setupRoutes(); err != nil {
		t.Fatalf("Failed to setup routes: %v", err)
	}
	
	t.Run("Login API", func(t *testing.T) {
		// Test valid login
		loginData := `{"username":"admin","password":"admin123"}`
		req := httptest.NewRequest("POST", "/api/login", strings.NewReader(loginData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		server.router.ServeHTTP(w, req)
		
		if w.Code != 200 {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		
		var result map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		
		if success, ok := result["success"].(bool); !ok || !success {
			t.Error("Expected successful login")
		}
		
		// Test invalid login
		invalidData := `{"username":"invalid","password":"wrong"}`
		req2 := httptest.NewRequest("POST", "/api/login", strings.NewReader(invalidData))
		req2.Header.Set("Content-Type", "application/json")
		w2 := httptest.NewRecorder()
		
		server.router.ServeHTTP(w2, req2)
		
		if w2.Code != 401 {
			t.Errorf("Expected status 401 for invalid login, got %d", w2.Code)
		}
	})
	
	t.Run("Users API", func(t *testing.T) {
		// Test GET users
		req := httptest.NewRequest("GET", "/api/users", nil)
		w := httptest.NewRecorder()
		
		server.router.ServeHTTP(w, req)
		
		if w.Code != 200 {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		
		var result map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
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
		req2 := httptest.NewRequest("POST", "/api/users", strings.NewReader(userData))
		req2.Header.Set("Content-Type", "application/json")
		w2 := httptest.NewRecorder()
		
		server.router.ServeHTTP(w2, req2)
		
		if w2.Code != 201 {
			t.Errorf("Expected status 201 for new user, got %d", w2.Code)
		}
	})
	
	t.Run("Posts API", func(t *testing.T) {
		// Test GET posts
		req := httptest.NewRequest("GET", "/api/posts", nil)
		w := httptest.NewRecorder()
		
		server.router.ServeHTTP(w, req)
		
		if w.Code != 200 {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		
		// Test GET posts with filter
		req2 := httptest.NewRequest("GET", "/api/posts?status=published", nil)
		w2 := httptest.NewRecorder()
		
		server.router.ServeHTTP(w2, req2)
		
		if w2.Code != 200 {
			t.Errorf("Expected status 200 for filtered posts, got %d", w2.Code)
		}
		
		// Test POST new post
		postData := `{"title":"Test Post","content":"Test content","author":"Test Author"}`
		req3 := httptest.NewRequest("POST", "/api/posts", strings.NewReader(postData))
		req3.Header.Set("Content-Type", "application/json")
		w3 := httptest.NewRecorder()
		
		server.router.ServeHTTP(w3, req3)
		
		if w3.Code != 201 {
			t.Errorf("Expected status 201 for new post, got %d", w3.Code)
		}
	})
	
	t.Run("Stats API", func(t *testing.T) {
		// Test GET all stats
		req := httptest.NewRequest("GET", "/api/stats", nil)
		w := httptest.NewRecorder()
		
		server.router.ServeHTTP(w, req)
		
		if w.Code != 200 {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		
		var result map[string]interface{}
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
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
		req2 := httptest.NewRequest("GET", "/api/stats?category=server", nil)
		w2 := httptest.NewRecorder()
		
		server.router.ServeHTTP(w2, req2)
		
		if w2.Code != 200 {
			t.Errorf("Expected status 200 for server stats, got %d", w2.Code)
		}
	})
}

func TestDynamicRoutingIntegration(t *testing.T) {
	memFS := setupIntegrationMemoryFS()
	
	server := &Server{
		router: mux.NewRouter(),
		port:   8080,
		fs:     memFS,
	}
	
	if err := server.setupRoutes(); err != nil {
		t.Fatalf("Failed to setup routes: %v", err)
	}
	
	testCases := []struct {
		name         string
		path         string
		expectedText string
		expectedStatus int
	}{
		{
			name:         "Blog Post 1",
			path:         "/blog/1",
			expectedText: "Welcome to our blog",
			expectedStatus: 200,
		},
		{
			name:         "Blog Post 2",
			path:         "/blog/2",
			expectedText: "Getting started with Redi",
			expectedStatus: 200,
		},
		{
			name:         "Blog Post 123",
			path:         "/blog/123",
			expectedText: "Dynamic Route Example",
			expectedStatus: 200,
		},
		{
			name:         "Non-existent Blog Post",
			path:         "/blog/999",
			expectedText: "404 - Post Not Found",
			expectedStatus: 404,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.path, nil)
			w := httptest.NewRecorder()
			
			server.router.ServeHTTP(w, req)
			
			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, w.Code)
			}
			
			body := w.Body.String()
			if !strings.Contains(body, tc.expectedText) {
				t.Errorf("Expected response to contain %s, got %s", tc.expectedText, body)
			}
		})
	}
}