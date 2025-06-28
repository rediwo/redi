package redi

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
	
	"github.com/rediwo/redi/filesystem"
)

type Route struct {
	Path      string
	FilePath  string
	FileType  string
	IsDynamic bool
	ParamName string
}

type RouteScanner struct {
	fs        filesystem.FileSystem
	routesDir string
}

func NewRouteScanner(fs filesystem.FileSystem, routesDir string) *RouteScanner {
	return &RouteScanner{
		fs:        fs,
		routesDir: routesDir,
	}
}

func (rs *RouteScanner) ScanRoutes() ([]Route, error) {
	var routes []Route

	err := rs.fs.WalkDir(rs.routesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			// Skip directories that start with "_"
			if strings.HasPrefix(d.Name(), "_") {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip files whose name starts with "_"
		name := d.Name()
		if strings.HasPrefix(name, "_") {
			return nil
		}
		
		// Skip files in any directory that starts with "_"
		pathParts := strings.Split(path, "/")
		for _, part := range pathParts {
			if strings.HasPrefix(part, "_") {
				return nil
			}
		}

		ext := filepath.Ext(name)
		if ext != ".html" && ext != ".js" && ext != ".md" {
			return nil
		}

		route := rs.createRoute(path, ext)
		routes = append(routes, route)
		return nil
	})
	
	return routes, err
}

func (rs *RouteScanner) createRoute(filePath, ext string) Route {
	// Remove the routes directory prefix to get the relative path
	// Handle both "routes/file.ext" and "routes\file.ext" (Windows path separator)
	relPath := filePath
	
	// Normalize path separators to forward slashes
	normalizedFilePath := strings.ReplaceAll(filePath, "\\", "/")
	normalizedRoutesDir := strings.ReplaceAll(rs.routesDir, "\\", "/")
	
	// Remove the routes directory prefix
	if strings.HasPrefix(normalizedFilePath, normalizedRoutesDir+"/") {
		relPath = strings.TrimPrefix(normalizedFilePath, normalizedRoutesDir+"/")
	} else if strings.HasPrefix(normalizedFilePath, normalizedRoutesDir) {
		// Handle case where filePath exactly equals routesDir (shouldn't happen for files)
		relPath = strings.TrimPrefix(normalizedFilePath, normalizedRoutesDir)
		relPath = strings.TrimPrefix(relPath, "/")
	}
	
	pathWithoutExt := strings.TrimSuffix(relPath, ext)
	
	// Ensure the URL path starts with "/"
	urlPath := "/" + pathWithoutExt
	
	// Handle index files - convert "/index" to "/"
	if strings.HasSuffix(urlPath, "/index") {
		urlPath = strings.TrimSuffix(urlPath, "/index")
		if urlPath == "" {
			urlPath = "/"
		}
	}
	
	// Handle dynamic routes [param] -> {param}
	isDynamic := false
	paramName := ""
	
	dynamicRegex := regexp.MustCompile(`\[(\w+)\]`)
	if dynamicRegex.MatchString(urlPath) {
		matches := dynamicRegex.FindStringSubmatch(urlPath)
		if len(matches) > 1 {
			paramName = matches[1]
			urlPath = dynamicRegex.ReplaceAllString(urlPath, fmt.Sprintf("{%s}", paramName))
			isDynamic = true
		}
	}
	
	return Route{
		Path:      urlPath,
		FilePath:  filePath,
		FileType:  strings.TrimPrefix(ext, "."),
		IsDynamic: isDynamic,
		ParamName: paramName,
	}
}
