package handlers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/rediwo/redi/filesystem"
)

func TestImportTransformer_AssetTypeDetection(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	transformer := NewImportTransformer(fs)

	tests := []struct {
		path     string
		expected string
	}{
		{"script.js", "javascript"},
		{"module.mjs", "javascript"},
		{"types.ts", "typescript"},
		{"styles.css", "stylesheet"},
		{"data.json", "json"},
		{"logo.png", "image"},
		{"photo.jpg", "image"},
		{"icon.svg", "image"},
		{"font.woff", "font"},
		{"font.woff2", "font"},
		{"unknown.xyz", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assetType := transformer.GetAssetType(tt.path)
			if assetType != tt.expected {
				t.Errorf("Expected asset type %s for %s, got %s", tt.expected, tt.path, assetType)
			}
		})
	}
}

func TestImportTransformer_ResolveAssetPath(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	transformer := NewImportTransformer(fs)

	// Create test files
	_ = fs.WriteFile("routes/component.js", []byte(""))
	_ = fs.WriteFile("routes/styles.css", []byte(""))
	_ = fs.WriteFile("public/logo.png", []byte(""))
	_ = fs.WriteFile("public/images/icon.svg", []byte(""))

	tests := []struct {
		name        string
		importPath  string
		currentPath string
		wantPath    string
		wantType    string
	}{
		{
			name:        "relative CSS in same directory",
			importPath:  "./styles.css",
			currentPath: "routes/component.js",
			wantPath:    "routes/styles.css",
			wantType:    "stylesheet",
		},
		{
			name:        "absolute image in public",
			importPath:  "/logo.png",
			currentPath: "routes/component.js",
			wantPath:    "public/logo.png",
			wantType:    "image",
		},
		{
			name:        "nested image in public",
			importPath:  "/images/icon.svg",
			currentPath: "routes/component.js",
			wantPath:    "public/images/icon.svg",
			wantType:    "image",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, gotType := transformer.ResolveAssetPath(tt.importPath, tt.currentPath)
			if gotPath != tt.wantPath {
				t.Errorf("Expected path %s, got %s", tt.wantPath, gotPath)
			}
			if gotType != tt.wantType {
				t.Errorf("Expected type %s, got %s", tt.wantType, gotType)
			}
		})
	}
}

func TestImportTransformer_TransformImports(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	transformer := NewImportTransformer(fs)

	// Create test files
	_ = fs.WriteFile("routes/styles.css", []byte(""))
	_ = fs.WriteFile("public/logo.png", []byte(""))

	jsCode := `
import styles from './styles.css';
import logo from '/logo.png';
import Button from './Button.svelte';

function init() {
	console.log(styles);
	console.log(logo);
}
`

	transformed, componentImports := transformer.TransformImports(jsCode, "routes/component.js", []string{".svelte"})

	// Check that CSS import was transformed to URL
	if !strings.Contains(transformed, `const styles = '/styles.css'`) {
		t.Error("Expected CSS import to be transformed to URL constant")
	}

	// Check that image import was transformed to URL
	if !strings.Contains(transformed, `const logo = '/logo.png'`) {
		t.Error("Expected image import to be transformed to URL constant")
	}

	// Check that Svelte import was preserved in component imports
	if componentImports["Button"] != "./Button.svelte" {
		t.Error("Expected Svelte import to be preserved in component imports")
	}

	// Check that import statements were removed
	if strings.Contains(transformed, "import ") {
		t.Error("Expected import statements to be removed")
	}
}

func TestImportTransformer_PublicURLGeneration(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	transformer := NewImportTransformer(fs)

	tests := []struct {
		assetPath string
		expected  string
	}{
		{"public/images/logo.png", "/images/logo.png"},
		{"routes/styles.css", "/styles.css"},
		{"assets/font.woff", "/assets/font.woff"},
		{"public/favicon.ico", "/favicon.ico"},
	}

	for _, tt := range tests {
		t.Run(tt.assetPath, func(t *testing.T) {
			url := transformer.GetPublicURL(tt.assetPath)
			if url != tt.expected {
				t.Errorf("Expected URL %s for path %s, got %s", tt.expected, tt.assetPath, url)
			}
		})
	}
}

func TestImportTransformer_JavaScriptTransformation(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	transformer := NewImportTransformer(fs)

	// Create test JS files with ES6 exports
	utilsJS := `export function hello() { return "Hello"; }
export const version = "1.0.0";
export default { hello, version };`
	_ = fs.WriteFile("routes/utils.js", []byte(utilsJS))
	
	// Create CommonJS module
	commonJS := `function greet(name) { return "Hi " + name; }
module.exports = { greet };`
	_ = fs.WriteFile("routes/common.js", []byte(commonJS))

	tests := []struct {
		name          string
		jsCode        string
		checkFunction func(transformed string) bool
	}{
		{
			name:   "ES6 module with exports",
			jsCode: `import utils from './utils.js';`,
			checkFunction: func(transformed string) bool {
				// Should contain IIFE that returns exports
				return strings.Contains(transformed, "const utils = (function()") &&
					strings.Contains(transformed, "__exports") &&
					strings.Contains(transformed, "return __exports")
			},
		},
		{
			name:   "CommonJS module",
			jsCode: `import common from './common.js';`,
			checkFunction: func(transformed string) bool {
				// Should contain module.exports wrapper
				return strings.Contains(transformed, "const common = (function()") &&
					strings.Contains(transformed, "module.exports")
			},
		},
		{
			name:   "Non-existent JS file",
			jsCode: `import missing from './missing.js';`,
			checkFunction: func(transformed string) bool {
				// Should fall back to URL
				return !strings.Contains(transformed, "const missing")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Transform the code
			transformed, _ := transformer.TransformImports(tt.jsCode, "routes/component.js", []string{})
			
			// Check if the transformation is correct
			if !tt.checkFunction(transformed) {
				t.Errorf("Transformation did not produce expected result.\nGot:\n%s", transformed)
			}
			
			// Ensure import was removed
			if strings.Contains(transformed, "import ") {
				t.Error("Import statement should be removed")
			}
		})
	}
}

func TestImportTransformer_JSONInlining(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	transformer := NewImportTransformer(fs)

	// Create test JSON files
	smallJSON := `{"name": "test", "version": "1.0.0", "data": [1, 2, 3]}`
	_ = fs.WriteFile("routes/config.json", []byte(smallJSON))
	
	// Create a large JSON file
	largeData := make([]int, 30000)
	for i := range largeData {
		largeData[i] = i
	}
	// Properly format the JSON array
	dataStrs := make([]string, len(largeData))
	for i, v := range largeData {
		dataStrs[i] = fmt.Sprintf("%d", v)
	}
	largeJSON := fmt.Sprintf(`{"large": [%s]}`, strings.Join(dataStrs, ","))
	_ = fs.WriteFile("routes/large.json", []byte(largeJSON))
	
	// Create invalid JSON
	_ = fs.WriteFile("routes/invalid.json", []byte(`{invalid json`))

	tests := []struct {
		name       string
		jsCode     string
		wantResult string
		wantInline bool
	}{
		{
			name:       "small JSON file should be inlined",
			jsCode:     `import config from './config.json';`,
			wantResult: fmt.Sprintf("const config = %s;", smallJSON),
			wantInline: true,
		},
		{
			name:       "large JSON file should also be inlined",
			jsCode:     `import large from './large.json';`,
			wantResult: fmt.Sprintf("const large = %s;", largeJSON),
			wantInline: true,
		},
		{
			name:       "invalid JSON should use URL",
			jsCode:     `import invalid from './invalid.json';`,
			wantResult: `const invalid = '/invalid.json';`,
			wantInline: false,
		},
		{
			name:       "non-existent JSON should not be transformed",
			jsCode:     `import missing from './missing.json';`,
			wantResult: ``,
			wantInline: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Transform the code
			transformed, _ := transformer.TransformImports(tt.jsCode, "routes/component.js", []string{})
			
			// Check if the result contains what we expect
			if tt.wantResult != "" {
				if tt.wantInline {
					// Should be inlined JSON data
					if !strings.Contains(transformed, "const config = {") && !strings.Contains(transformed, "const large = {") {
						t.Errorf("Expected transformed code to contain inlined JSON data, got:\n%s", transformed)
					}
				} else {
					// Should be a URL string
					if !strings.Contains(transformed, tt.wantResult) {
						t.Errorf("Expected transformed code to contain URL assignment:\n%s\nGot:\n%s", tt.wantResult, transformed)
					}
				}
			}
			
			// Ensure import was removed
			if strings.Contains(transformed, "import ") {
				t.Error("Import statement should be removed")
			}
		})
	}
}