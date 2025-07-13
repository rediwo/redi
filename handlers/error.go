package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"

	"github.com/rediwo/redi/filesystem"
)

// ErrorHandler handles HTTP errors with custom error pages
type ErrorHandler struct {
	fs             filesystem.FileSystem
	templateHandler *TemplateHandler
	cache          map[int]*template.Template
	routesDir      string
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(fs filesystem.FileSystem, templateHandler *TemplateHandler) *ErrorHandler {
	return &ErrorHandler{
		fs:              fs,
		templateHandler: templateHandler,
		cache:           make(map[int]*template.Template),
		routesDir:       "routes", // Default value
	}
}

// NewErrorHandlerWithRoutesDir creates a new error handler with custom routes directory
func NewErrorHandlerWithRoutesDir(fs filesystem.FileSystem, templateHandler *TemplateHandler, routesDir string) *ErrorHandler {
	return &ErrorHandler{
		fs:              fs,
		templateHandler: templateHandler,
		cache:           make(map[int]*template.Template),
		routesDir:       routesDir,
	}
}

// ErrorData represents the data passed to error templates
type ErrorData struct {
	Status      int
	StatusText  string
	Message     string
	Path        string
	Method      string
	RequestID   string
}

// ServeError serves an error response with the given status code and message
func (eh *ErrorHandler) ServeError(w http.ResponseWriter, r *http.Request, status int, message string) {
	// Try to load custom error template
	tmpl := eh.getErrorTemplate(status)
	
	// Prepare error data
	errorData := ErrorData{
		Status:     status,
		StatusText: http.StatusText(status),
		Message:    message,
		Path:       r.URL.Path,
		Method:     r.Method,
	}
	
	// Buffer the response to ensure proper gzip handling
	var buf bytes.Buffer
	
	if tmpl != nil {
		// Render custom error template to buffer
		err := tmpl.Execute(&buf, errorData)
		if err != nil {
			log.Printf("Error rendering error template: %v", err)
			// Fall back to simple error
			buf.Reset()
			eh.writeSimpleError(&buf, errorData)
		}
	} else {
		// Write simple error page to buffer
		eh.writeSimpleError(&buf, errorData)
	}
	
	// Set headers after we know the content
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	
	// Write status and body together
	// This allows the gzip middleware to properly handle the response
	w.WriteHeader(status)
	w.Write(buf.Bytes())
}

// getErrorTemplate loads and caches error templates
func (eh *ErrorHandler) getErrorTemplate(status int) *template.Template {
	// Check cache first
	if tmpl, ok := eh.cache[status]; ok {
		return tmpl
	}
	
	// Try to load template file
	templatePath := filepath.Join(eh.routesDir, fmt.Sprintf("%d.html", status))
	content, err := eh.fs.ReadFile(templatePath)
	if err != nil {
		// Try generic error template
		if status >= 400 && status < 500 {
			templatePath = filepath.Join(eh.routesDir, "4xx.html")
		} else if status >= 500 {
			templatePath = filepath.Join(eh.routesDir, "5xx.html")
		}
		
		content, err = eh.fs.ReadFile(templatePath)
		if err != nil {
			return nil
		}
	}
	
	// Process layouts if template handler is available
	if eh.templateHandler != nil {
		processedContent, err := eh.templateHandler.processLayouts(string(content))
		if err == nil {
			content = []byte(processedContent)
		}
	}
	
	// Parse template
	tmpl, err := template.New(fmt.Sprintf("error-%d", status)).Parse(string(content))
	if err != nil {
		log.Printf("Error parsing error template %s: %v", templatePath, err)
		return nil
	}
	
	// Cache the template
	eh.cache[status] = tmpl
	return tmpl
}

// writeSimpleError writes a simple HTML error page to the given writer
func (eh *ErrorHandler) writeSimpleError(w io.Writer, data ErrorData) {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>%d %s</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            margin: 0;
            padding: 0;
            background-color: #f5f5f5;
            color: #333;
        }
        .container {
            max-width: 600px;
            margin: 100px auto;
            padding: 40px;
            background-color: white;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
            text-align: center;
        }
        h1 {
            font-size: 72px;
            margin: 0 0 10px 0;
            color: #666;
        }
        h2 {
            font-size: 24px;
            margin: 0 0 20px 0;
            font-weight: normal;
            color: #666;
        }
        p {
            font-size: 16px;
            color: #777;
            margin: 20px 0;
        }
        .details {
            margin-top: 30px;
            padding-top: 30px;
            border-top: 1px solid #eee;
            text-align: left;
        }
        .details code {
            background-color: #f5f5f5;
            padding: 2px 4px;
            border-radius: 3px;
            font-family: Consolas, Monaco, monospace;
        }
        a {
            color: #0066cc;
            text-decoration: none;
        }
        a:hover {
            text-decoration: underline;
        }
        .home-link {
            display: inline-block;
            margin-top: 20px;
            padding: 10px 20px;
            background-color: #0066cc;
            color: white;
            border-radius: 4px;
        }
        .home-link:hover {
            background-color: #0052a3;
            text-decoration: none;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>%d</h1>
        <h2>%s</h2>
        <p>%s</p>
        <div class="details">
            <p><strong>Request Path:</strong> <code>%s</code></p>
            <p><strong>Request Method:</strong> <code>%s</code></p>
        </div>
        <a href="/" class="home-link">Go to Homepage</a>
    </div>
</body>
</html>`, data.Status, data.StatusText, data.Status, data.StatusText, data.Message, data.Path, data.Method)
	
	w.Write([]byte(html))
}

// Handle404 is a convenience method for 404 errors
func (eh *ErrorHandler) Handle404(w http.ResponseWriter, r *http.Request) {
	eh.ServeError(w, r, http.StatusNotFound, "The requested page could not be found.")
}

// Handle500 is a convenience method for 500 errors
func (eh *ErrorHandler) Handle500(w http.ResponseWriter, r *http.Request, err error) {
	message := "An internal server error occurred."
	if err != nil {
		// In production, you might want to hide error details
		message = fmt.Sprintf("Internal Server Error: %v", err)
	}
	eh.ServeError(w, r, http.StatusInternalServerError, message)
}