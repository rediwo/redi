package handlers

import (
	"bytes"
	"fmt"
	htmltemplate "html/template"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	texttemplate "text/template"

	"github.com/rediwo/redi/filesystem"
	"github.com/yuin/goldmark"
)

type TemplateHandler struct {
	fs filesystem.FileSystem
}

func NewTemplateHandler(fs filesystem.FileSystem) *TemplateHandler {
	return &TemplateHandler{fs: fs}
}

func (th *TemplateHandler) Handle(route Route) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read the template file
		content, err := th.fs.ReadFile(route.FilePath)
		if err != nil {
			http.Error(w, "Template not found", http.StatusNotFound)
			return
		}

		// Render the template with no data (direct asset access)
		err = th.RenderTemplate(route.FilePath, string(content), nil, w)
		if err != nil {
			http.Error(w, fmt.Sprintf("Template rendering error: %v", err), http.StatusInternalServerError)
		}
	}
}

// RenderTemplate renders a template file with the given data
func (th *TemplateHandler) RenderTemplate(templatePath, templateContent string, data interface{}, w http.ResponseWriter) error {
	ext := strings.ToLower(filepath.Ext(templatePath))
	
	// Process layouts for HTML templates
	if ext == ".html" {
		processedContent, err := th.processLayouts(templateContent)
		if err != nil {
			return fmt.Errorf("layout processing error: %v", err)
		}
		templateContent = processedContent
	}

	// Choose template engine based on file extension
	switch ext {
	case ".html":
		return th.renderHTMLTemplate(templateContent, data, w)
	case ".md":
		return th.renderMarkdownTemplate(templateContent, data, w)
	case ".json", ".txt", ".css", ".js":
		return th.renderTextTemplate(templateContent, data, w)
	default:
		// Default to text template for unknown extensions
		return th.renderTextTemplate(templateContent, data, w)
	}
}

// renderHTMLTemplate renders using html/template
func (th *TemplateHandler) renderHTMLTemplate(content string, data interface{}, w http.ResponseWriter) error {
	tmpl, err := htmltemplate.New("template").Parse(content)
	if err != nil {
		return fmt.Errorf("HTML template parsing error: %v", err)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.Execute(w, data)
}

// renderTextTemplate renders using text/template  
func (th *TemplateHandler) renderTextTemplate(content string, data interface{}, w http.ResponseWriter) error {
	tmpl, err := texttemplate.New("template").Parse(content)
	if err != nil {
		return fmt.Errorf("text template parsing error: %v", err)
	}

	// Set appropriate content type based on file extension in the content
	ext := th.guessContentType(content)
	w.Header().Set("Content-Type", ext)
	
	return tmpl.Execute(w, data)
}

// renderMarkdownTemplate converts markdown to HTML and renders it
func (th *TemplateHandler) renderMarkdownTemplate(content string, data interface{}, w http.ResponseWriter) error {
	var contentToConvert []byte
	
	// Only process as template if data is provided
	if data != nil {
		// Process as text template to handle Go template variables
		tmpl, err := texttemplate.New("template").Parse(content)
		if err != nil {
			return fmt.Errorf("markdown template parsing error: %v", err)
		}
		
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return fmt.Errorf("markdown template execution error: %v", err)
		}
		contentToConvert = buf.Bytes()
	} else {
		// No data provided, just convert markdown directly
		contentToConvert = []byte(content)
	}
	
	// Convert markdown to HTML
	var htmlBuf bytes.Buffer
	if err := goldmark.Convert(contentToConvert, &htmlBuf); err != nil {
		return fmt.Errorf("markdown conversion error: %v", err)
	}
	
	// Set HTML content type and write response
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err := w.Write(htmlBuf.Bytes())
	return err
}

// guessContentType guesses content type from content or defaults to text/plain
func (th *TemplateHandler) guessContentType(content string) string {
	content = strings.TrimSpace(content)
	
	if strings.HasPrefix(content, "{") || strings.HasPrefix(content, "[") {
		return "application/json; charset=utf-8"
	}
	if strings.HasPrefix(content, "#") || strings.Contains(content, "##") {
		return "text/markdown; charset=utf-8"
	}
	if strings.HasPrefix(content, "<!DOCTYPE") || strings.HasPrefix(content, "<html") {
		return "text/html; charset=utf-8"
	}
	
	return "text/plain; charset=utf-8"
}

// processLayouts processes layout directives in HTML templates
func (th *TemplateHandler) processLayouts(content string) (string, error) {
	// This is the same layout processing logic from html.go
	layoutRegex := regexp.MustCompile(`{{\s*layout\s+['"]([^'"]+)['"]\s*}}`)

	for {
		matches := layoutRegex.FindStringSubmatch(content)
		if len(matches) == 0 {
			break
		}

		layoutName := matches[1]
		// Try to find layout in multiple locations
		// First try _layout directory relative to the template
		layoutPaths := []string{
			filepath.Join("_layout", layoutName+".html"),
			filepath.Join("routes", "_layout", layoutName+".html"), // Keep for compatibility
		}
		
		var layoutContent []byte
		var err error
		var found bool
		
		for _, layoutPath := range layoutPaths {
			layoutContent, err = th.fs.ReadFile(layoutPath)
			if err == nil {
				found = true
				break
			}
		}
		
		if !found {
			return "", fmt.Errorf("layout file not found: %s", layoutName)
		}

		contentWithoutLayout := layoutRegex.ReplaceAllString(content, "")

		layoutStr := string(layoutContent)
		if strings.Contains(layoutStr, "{{.Content}}") {
			content = strings.ReplaceAll(layoutStr, "{{.Content}}", contentWithoutLayout)
		} else {
			content = layoutStr + contentWithoutLayout
		}
	}

	return content, nil
}