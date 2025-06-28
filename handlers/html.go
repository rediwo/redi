package handlers

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	js "github.com/dop251/goja"
	"github.com/gorilla/mux"
	"github.com/rediwo/redi/filesystem"
)

type HTMLHandler struct {
	fs filesystem.FileSystem
}

func NewHTMLHandler(fs filesystem.FileSystem) *HTMLHandler {
	return &HTMLHandler{fs: fs}
}

func (hh *HTMLHandler) Handle(route Route) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		content, err := hh.fs.ReadFile(route.FilePath)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		templateContent, serverScript := hh.extractServerScript(string(content))

		data := make(map[string]interface{})

		if serverScript != "" {
			data, err = hh.executeServerScript(serverScript, r, route, templateContent)
			if err != nil {
				http.Error(w, fmt.Sprintf("Server script error: %v", err), http.StatusInternalServerError)
				return
			}
		}

		finalTemplate, err := hh.processLayouts(templateContent)
		if err != nil {
			http.Error(w, fmt.Sprintf("Layout processing error: %v", err), http.StatusInternalServerError)
			return
		}

		tmpl, err := template.New("page").Parse(finalTemplate)
		if err != nil {
			http.Error(w, fmt.Sprintf("Template parsing error: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// Check if a status code was set in the server script
		if statusCode, ok := data["_statusCode"].(int); ok && statusCode != 200 {
			w.WriteHeader(statusCode)
			delete(data, "_statusCode") // Remove internal field before template execution
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, fmt.Sprintf("Template execution error: %v", err), http.StatusInternalServerError)
		}
	}
}

func (hh *HTMLHandler) extractServerScript(content string) (string, string) {
	re := regexp.MustCompile(`(?s)<script\s+@server>(.*?)</script>`)
	matches := re.FindStringSubmatch(content)

	if len(matches) > 1 {
		cleanContent := re.ReplaceAllString(content, "")
		return cleanContent, strings.TrimSpace(matches[1])
	}

	return content, ""
}

func (hh *HTMLHandler) executeServerScript(script string, r *http.Request, route Route, templateContent string) (map[string]interface{}, error) {
	vm := js.New()
	data := make(map[string]interface{})
	statusCode := 200

	vars := mux.Vars(r)

	reqObj := map[string]interface{}{
		"method":  r.Method,
		"url":     r.URL.String(),
		"path":    r.URL.Path,
		"query":   r.URL.RawQuery,
		"headers": r.Header,
		"params":  vars,
	}

	if r.Method == "POST" {
		body, _ := io.ReadAll(r.Body)
		reqObj["body"] = string(body)
	}

	resObj := map[string]interface{}{
		"render": func(renderData interface{}) {
			if dataMap, ok := renderData.(map[string]interface{}); ok {
				for k, v := range dataMap {
					data[k] = v
				}
			}
		},
		"status": func(code int) {
			statusCode = code
		},
	}

	vm.Set("req", reqObj)
	vm.Set("res", resObj)

	_, err := vm.RunString(script)

	// Store status code in data for the handler to use
	data["_statusCode"] = statusCode

	return data, err
}

func (hh *HTMLHandler) processLayouts(content string) (string, error) {
	layoutRegex := regexp.MustCompile(`{{\s*layout\s+['"]([^'"]+)['"]\s*}}`)

	for {
		matches := layoutRegex.FindStringSubmatch(content)
		if len(matches) == 0 {
			break
		}

		layoutName := matches[1]
		layoutPath := filepath.Join("routes", "_layout", layoutName+".html")

		layoutContent, err := hh.fs.ReadFile(layoutPath)
		if err != nil {
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