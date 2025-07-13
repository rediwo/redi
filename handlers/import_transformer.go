package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rediwo/redi/filesystem"
)

// ImportTransformer handles transformation of imports in JavaScript/TypeScript code
type ImportTransformer struct {
	fs filesystem.FileSystem
}

// NewImportTransformer creates a new import transformer
func NewImportTransformer(fs filesystem.FileSystem) *ImportTransformer {
	return &ImportTransformer{
		fs: fs,
	}
}

// TransformImports processes all imports in the given JavaScript code
// It returns the transformed code and a map of component imports (for framework-specific handling)
func (it *ImportTransformer) TransformImports(jsCode string, currentPath string, componentExtensions []string) (string, map[string]string) {
	// First handle complex import patterns (destructuring, etc.)
	// Match various import patterns:
	// - import { a, b } from 'module'
	// - import React, { useState } from 'react'  
	// - import { x }, y from 'module'
	// - import * as name from 'module'
	complexImportRegex := regexp.MustCompile(`import\s+(?:\*\s+as\s+\w+|{[^}]+}|(?:\w+\s*,\s*{[^}]+})|(?:{[^}]+}\s*,\s*\w+)|\w+)\s+from\s+["']([^"']+)["'];?`)
	complexMatches := complexImportRegex.FindAllStringSubmatch(jsCode, -1)
	
	// Track which imports are library imports that should be preserved
	libraryImports := make(map[string]bool)
	for _, match := range complexMatches {
		if len(match) >= 2 {
			importPath := match[1]
			if isLibraryImport(importPath) {
				libraryImports[strings.TrimSpace(match[0])] = true
			}
		}
	}

	// Extract simple imports: import Name from 'path'
	importRegex := regexp.MustCompile(`import\s+(\w+)\s+from\s+["']([^"']+)["'];?\s*`)
	matches := importRegex.FindAllStringSubmatch(jsCode, -1)

	// Build import mappings and handle non-component imports
	componentImports := make(map[string]string)
	
	for _, match := range matches {
		if len(match) >= 3 {
			importName := match[1]
			importPath := match[2]
			
			// Skip library imports
			if isLibraryImport(importPath) {
				libraryImports[strings.TrimSpace(match[0])] = true
				continue
			}
			
			// Check if this is a component import
			isComponent := false
			for _, ext := range componentExtensions {
				if strings.HasSuffix(importPath, ext) {
					isComponent = true
					break
				}
			}
			
			if isComponent {
				// Component imports are handled by the framework
				componentImports[importName] = importPath
			} else {
				// Other assets need URL resolution
				assetPath, assetType := it.ResolveAssetPath(importPath, currentPath)
				if assetPath != "" && assetType != "unknown" {
					// Replace import with URL or inline content based on type
					jsCode = it.ReplaceAssetImport(jsCode, importName, importPath, assetPath, assetType)
				} else {
					log.Printf("Warning: Could not resolve asset import '%s' from '%s'", importPath, currentPath)
				}
			}
		}
	}

	// Remove only non-library import statements
	jsCode = it.removeNonLibraryImports(jsCode, libraryImports)
	
	return jsCode, componentImports
}

// isLibraryImport checks if an import path is a node_modules library
func isLibraryImport(importPath string) bool {
	// Library imports don't start with . or /
	// This covers all node_modules packages
	return !strings.HasPrefix(importPath, ".") && !strings.HasPrefix(importPath, "/")
}

// removeNonLibraryImports removes only non-library import statements from the code
func (it *ImportTransformer) removeNonLibraryImports(jsCode string, libraryImports map[string]bool) string {
	// Define all import patterns
	importPatterns := []string{
		`import\s*\*\s+as\s+\w+\s+from\s*["'][^"']+["'];?`,
		`import\s+\w+\s*,\s*{[^}]+}\s*from\s*["'][^"']+["'];?`,
		`import\s+{[^}]+}\s*,\s*\w+\s*from\s*["'][^"']+["'];?`,
		`import\s*{[^}]+}\s*from\s*["'][^"']+["'];?`,
		`import\s+\w+\s+from\s*["'][^"']+["'];?`,
		`import\s*["'][^"']+["'];?`,
	}
	
	for _, pattern := range importPatterns {
		importRegex := regexp.MustCompile(pattern)
		matches := importRegex.FindAllString(jsCode, -1)
		
		for _, match := range matches {
			// Only remove if it's not a library import
			trimmedMatch := strings.TrimSpace(match)
			if !libraryImports[trimmedMatch] && !libraryImports[match] {
				jsCode = strings.Replace(jsCode, match, "", 1)
			}
		}
	}
	
	return jsCode
}

// ResolveAssetPath attempts to find an asset in various locations
func (it *ImportTransformer) ResolveAssetPath(importPath string, currentPath string) (string, string) {
	// For absolute imports starting with /, try public directory first
	if strings.HasPrefix(importPath, "/") {
		// Remove leading slash
		trimmedPath := strings.TrimPrefix(importPath, "/")
		
		// Try in public directory first
		publicPath := filepath.Join("public", trimmedPath)
		if _, err := it.fs.Stat(publicPath); err == nil {
			return publicPath, it.GetAssetType(publicPath)
		}
		
		// Try at root
		if _, err := it.fs.Stat(trimmedPath); err == nil {
			return trimmedPath, it.GetAssetType(trimmedPath)
		}
		
		// Try with routes prefix
		routesPath := filepath.Join("routes", trimmedPath)
		if _, err := it.fs.Stat(routesPath); err == nil {
			return routesPath, it.GetAssetType(routesPath)
		}
		
		return "", ""
	}
	
	// For relative imports, resolve relative to current file
	resolvedPath := it.resolveRelativePath(importPath, currentPath)
	if resolvedPath == "" {
		return "", ""
	}

	// Try the path as-is first
	if _, err := it.fs.Stat(resolvedPath); err == nil {
		return resolvedPath, it.GetAssetType(resolvedPath)
	}

	// If not found, try common asset directories
	// Try in public directory for static assets
	publicPath := filepath.Join("public", resolvedPath)
	if _, err := it.fs.Stat(publicPath); err == nil {
		return publicPath, it.GetAssetType(publicPath)
	}

	// Try without public prefix if the import already contains it
	if strings.HasPrefix(resolvedPath, "public/") {
		directPath := strings.TrimPrefix(resolvedPath, "public/")
		if _, err := it.fs.Stat(directPath); err == nil {
			return directPath, it.GetAssetType(directPath)
		}
	}

	return "", ""
}

// resolveRelativePath resolves a relative import path
func (it *ImportTransformer) resolveRelativePath(importPath string, currentPath string) string {
	// Handle absolute imports (starting with /)
	if strings.HasPrefix(importPath, "/") {
		// Remove leading slash and treat as relative to root
		return strings.TrimPrefix(importPath, "/")
	}

	// Get the directory of the current file
	currentDir := filepath.Dir(currentPath)

	// Resolve the import path relative to the current directory
	resolvedPath := filepath.Join(currentDir, importPath)

	// Clean the path to handle .. and .
	resolvedPath = filepath.Clean(resolvedPath)

	// Ensure the resolved path doesn't escape the root directory
	// by checking it doesn't start with ../
	if strings.HasPrefix(resolvedPath, "../") {
		// Security: prevent directory traversal attacks
		return ""
	}

	return resolvedPath
}

// GetAssetType determines the type of asset based on file extension
func (it *ImportTransformer) GetAssetType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".js", ".mjs":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".css":
		return "stylesheet"
	case ".json":
		return "json"
	case ".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp":
		return "image"
	case ".woff", ".woff2", ".ttf", ".eot":
		return "font"
	default:
		return "unknown"
	}
}

// ReplaceAssetImport handles non-component asset imports
func (it *ImportTransformer) ReplaceAssetImport(jsCode, importName, importPath, assetPath, assetType string) string {
	var replacement string
	
	// Generate the public URL for the asset
	publicURL := it.GetPublicURL(assetPath)
	
	switch assetType {
	case "image", "font":
		// For images and fonts, replace with URL string
		replacement = fmt.Sprintf("const %s = '%s';", importName, publicURL)
	case "stylesheet":
		// For CSS, we could inject a link tag or return URL
		replacement = fmt.Sprintf("const %s = '%s';", importName, publicURL)
	case "javascript", "typescript":
		// For JS/TS modules, try to inline and transform to make exports available
		jsData, err := it.fs.ReadFile(assetPath)
		if err != nil {
			// If we can't read it, fall back to URL
			log.Printf("Warning: Could not read JS file '%s': %v", assetPath, err)
			replacement = fmt.Sprintf("const %s = '%s';", importName, publicURL)
		} else {
			// Transform the JS module to extract exports
			jsContent := string(jsData)
			transformed := it.TransformJSModule(jsContent, importName)
			replacement = transformed
		}
	case "json":
		// For JSON, inline the content as JavaScript object
		jsonData, err := it.fs.ReadFile(assetPath)
		if err != nil {
			// If we can't read it, fall back to URL
			log.Printf("Warning: Could not read JSON file '%s': %v", assetPath, err)
			replacement = fmt.Sprintf("const %s = '%s';", importName, publicURL)
		} else {
			// Validate JSON
			var jsonObj any
			if err := json.Unmarshal(jsonData, &jsonObj); err != nil {
				log.Printf("Warning: Invalid JSON in file '%s': %v", assetPath, err)
				replacement = fmt.Sprintf("const %s = '%s';", importName, publicURL)
			} else {
				// Inline the JSON data
				replacement = fmt.Sprintf("const %s = %s;", importName, string(jsonData))
			}
		}
	default:
		replacement = fmt.Sprintf("const %s = '%s';", importName, publicURL)
	}
	
	// Replace the specific import statement
	importPattern := fmt.Sprintf(`import\s+%s\s+from\s+["']%s["'];?\s*`, importName, regexp.QuoteMeta(importPath))
	re := regexp.MustCompile(importPattern)
	return re.ReplaceAllString(jsCode, replacement+"\n")
}

// TransformJSModule transforms a JavaScript module to make its exports available
func (it *ImportTransformer) TransformJSModule(jsContent string, importName string) string {
	// Check for ES6 exports
	hasESExports := regexp.MustCompile(`export\s+(const|let|var|function|class|default)`).MatchString(jsContent)
	hasNamedExports := regexp.MustCompile(`export\s*{[^}]+}`).MatchString(jsContent)
	
	// Check for CommonJS exports
	hasCommonJSExports := regexp.MustCompile(`module\.exports\s*=`).MatchString(jsContent) ||
		regexp.MustCompile(`exports\.\w+\s*=`).MatchString(jsContent)
	
	if hasESExports || hasNamedExports {
		// Handle ES6 modules
		return it.transformES6Module(jsContent, importName)
	} else if hasCommonJSExports {
		// Handle CommonJS modules
		return it.transformCommonJSModule(jsContent, importName)
	} else {
		// If no exports detected, wrap in IIFE and return empty object
		return fmt.Sprintf(`const %s = (function() {
%s
    return {};
})();`, importName, jsContent)
	}
}

// transformES6Module handles ES6 module transformation
func (it *ImportTransformer) transformES6Module(jsContent string, importName string) string {
	// Create an IIFE that collects exports
	var result strings.Builder
	result.WriteString(fmt.Sprintf("const %s = (function() {\n", importName))
	result.WriteString("    const __exports = {};\n")
	
	// Collect exported variables/constants
	exportVarRegex := regexp.MustCompile(`export\s+(const|let|var)\s+(\w+)`)
	varMatches := exportVarRegex.FindAllStringSubmatch(jsContent, -1)
	var exportedVars []string
	for _, match := range varMatches {
		if len(match) > 2 {
			exportedVars = append(exportedVars, match[2])
		}
	}
	
	// Collect exported functions
	exportFuncRegex := regexp.MustCompile(`export\s+function\s+(\w+)`)
	funcMatches := exportFuncRegex.FindAllStringSubmatch(jsContent, -1)
	for _, match := range funcMatches {
		if len(match) > 1 {
			exportedVars = append(exportedVars, match[1])
		}
	}
	
	// Now transform the content
	jsContent = exportVarRegex.ReplaceAllString(jsContent, "$1 $2")
	jsContent = exportFuncRegex.ReplaceAllString(jsContent, "function $1")
	
	// Handle: export default ...
	defaultExportRegex := regexp.MustCompile(`export\s+default\s+`)
	hasDefault := defaultExportRegex.MatchString(jsContent)
	if hasDefault {
		jsContent = defaultExportRegex.ReplaceAllString(jsContent, "__exports.default = ")
	}
	
	// Add the transformed content
	result.WriteString(jsContent)
	result.WriteString("\n")
	
	// Export all collected variables
	for _, varName := range exportedVars {
		result.WriteString(fmt.Sprintf("    __exports.%s = %s;\n", varName, varName))
	}
	
	result.WriteString("    return __exports;\n")
	result.WriteString("})();\n")
	
	return result.String()
}

// transformCommonJSModule handles CommonJS module transformation
func (it *ImportTransformer) transformCommonJSModule(jsContent string, importName string) string {
	// Wrap in IIFE with module and exports objects
	return fmt.Sprintf(`const %s = (function() {
    const module = { exports: {} };
    const exports = module.exports;
    
%s
    
    return module.exports;
})();`, importName, jsContent)
}

// GetPublicURL converts a filesystem path to a public URL
func (it *ImportTransformer) GetPublicURL(assetPath string) string {
	// Remove public/ prefix if present
	if prefix, found := strings.CutPrefix(assetPath, "public/"); found {
		return "/" + prefix
	}
	// For paths in routes/, we might need different handling
	if prefix, found := strings.CutPrefix(assetPath, "routes/"); found {
		return "/" + prefix
	}
	// Default: assume it's relative to root
	return "/" + assetPath
}