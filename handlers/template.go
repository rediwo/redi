package handlers

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	htmltemplate "html/template"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	texttemplate "text/template"

	"github.com/gorilla/mux"
	"github.com/rediwo/redi/filesystem"
	"github.com/rediwo/redi/utils"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/js"
	"github.com/yuin/goldmark"
)

// TemplateConfig holds all template-related settings
type TemplateConfig struct {
	// Vimesh Style settings
	VimeshStyle     *utils.VimeshStyleConfig // Vimesh Style configuration
	VimeshStylePath string                   // Path for Vimesh Style resource (default: "/vimesh-style.js")
	MinifyVimesh    bool                     // Enable Vimesh Style minification
	DevMode         bool                     // Development mode (disable minification)
}

// DefaultTemplateConfig returns default template settings
func DefaultTemplateConfig() *TemplateConfig {
	return &TemplateConfig{
		VimeshStyle:     utils.DefaultVimeshStyleConfig(),
		VimeshStylePath: "/vimesh-style.js",
		MinifyVimesh:    true,
		DevMode:         false,
	}
}

type TemplateHandler struct {
	fs                  filesystem.FileSystem
	config              *TemplateConfig
	minifiedVimeshStyle string
	vimeshMinified      bool
	vimeshMu            sync.Mutex
	minifier            *minify.M
}

func NewTemplateHandler(fs filesystem.FileSystem) *TemplateHandler {
	return NewTemplateHandlerWithConfig(fs, DefaultTemplateConfig())
}

func NewTemplateHandlerWithConfig(fs filesystem.FileSystem, config *TemplateConfig) *TemplateHandler {
	m := minify.New()
	m.AddFunc("text/javascript", js.Minify)

	return &TemplateHandler{
		fs:       fs,
		config:   config,
		minifier: m,
	}
}

// NewTemplateHandlerWithRouter creates a TemplateHandler with router for registering routes
func NewTemplateHandlerWithRouter(fs filesystem.FileSystem, config *TemplateConfig, router *mux.Router) *TemplateHandler {
	handler := NewTemplateHandlerWithConfig(fs, config)
	handler.RegisterRoutes(router)
	return handler
}

// RegisterRoutes registers additional routes for the template handler
func (th *TemplateHandler) RegisterRoutes(router *mux.Router) {
	// Register Vimesh Style route if enabled
	if th.config.VimeshStyle != nil && th.config.VimeshStyle.Enable && th.config.VimeshStylePath != "" {
		router.HandleFunc(th.config.VimeshStylePath, th.ServeVimeshStyle).Methods("GET", "HEAD")
		log.Printf("Registered HTML Vimesh Style route: %s", th.config.VimeshStylePath)
	}
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
	// Process Vimesh Style if enabled
	if th.config.VimeshStyle != nil && th.config.VimeshStyle.Enable {
		content = th.processVimeshStyle(content)
	}

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

// ServeVimeshStyle serves the Vimesh Style JavaScript as a static resource
func (th *TemplateHandler) ServeVimeshStyle(w http.ResponseWriter, r *http.Request) {
	vimeshJS := th.getMinifiedVimeshStyle()

	// Calculate ETag based on content
	hash := md5.Sum([]byte(vimeshJS))
	etag := `"` + hex.EncodeToString(hash[:]) + `"`

	// Check if client has cached version
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// Set caching headers
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")

	// Add compression hint
	w.Header().Set("Vary", "Accept-Encoding")

	w.Write([]byte(vimeshJS))
}

// getMinifiedVimeshStyle returns the Vimesh Style JS, minified if enabled
func (th *TemplateHandler) getMinifiedVimeshStyle() string {
	th.vimeshMu.Lock()
	defer th.vimeshMu.Unlock()

	// Return cached version if available
	if th.vimeshMinified {
		return th.minifiedVimeshStyle
	}

	// Get the Vimesh Style code
	vimeshCode := utils.GetVimeshStyleJS()

	// Minify if enabled and not in dev mode
	if th.config.MinifyVimesh && !th.config.DevMode {
		minified, err := th.minifier.String("text/javascript", vimeshCode)
		if err != nil {
			log.Printf("Failed to minify Vimesh Style: %v", err)
			th.minifiedVimeshStyle = vimeshCode
		} else {
			th.minifiedVimeshStyle = minified
			sizeBefore := len(vimeshCode)
			sizeAfter := len(minified)
			reduction := float64(sizeBefore-sizeAfter) / float64(sizeBefore) * 100
			log.Printf("Vimesh Style minified: %d bytes -> %d bytes (%.1f%% reduction)",
				sizeBefore, sizeAfter, reduction)
		}
	} else {
		th.minifiedVimeshStyle = vimeshCode
	}

	th.vimeshMinified = true
	return th.minifiedVimeshStyle
}

// processVimeshStyle processes HTML content to inject Vimesh Style if enabled
func (th *TemplateHandler) processVimeshStyle(content string) string {
	if th.config.VimeshStyle == nil || !th.config.VimeshStyle.Enable {
		return content
	}

	// Use GetCSSFromHTML to extract and generate CSS from the HTML
	generatedCSS, err := utils.GetCSSFromHTML(content)
	if err != nil {
		log.Printf("Failed to generate Vimesh CSS: %v", err)
		generatedCSS = ""
	}

	// Create style tag with id="vimesh-styles"
	css := ""
	if generatedCSS != "" {
		css = fmt.Sprintf(`<style id="vimesh-styles">%s</style>`, generatedCSS)
	}

	// Inject Vimesh Style runtime and CSS before closing </head>
	vimeshScript := fmt.Sprintf(`<script src="%s"></script>`, th.config.VimeshStylePath)

	// Try to inject before </head>
	if strings.Contains(content, "</head>") {
		content = strings.Replace(content, "</head>",
			css+"\n"+vimeshScript+"\n</head>", 1)
	} else if strings.Contains(content, "<body>") {
		// If no </head>, inject after <body>
		content = strings.Replace(content, "<body>",
			"<body>\n"+css+"\n"+vimeshScript+"\n", 1)
	} else {
		// As last resort, prepend to content
		content = css + "\n" + vimeshScript + "\n" + content
	}

	return content
}
