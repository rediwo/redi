package handlers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/rediwo/redi/filesystem"
)

func TestSvelteHandler_AssetImports(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	handler := NewSvelteHandler(fs)

	tests := []struct {
		name     string
		source   string
		expected []string
	}{
		{
			name: "import CSS file",
			source: `
<script>
import styles from './styles.css';
import Button from './Button.svelte';
</script>
`,
			expected: []string{"./styles.css", "./Button.svelte"},
		},
		{
			name: "import JavaScript library",
			source: `
<script>
import utils from './utils.js';
import { helpers } from '../lib/helpers.js';
import Component from './Component.svelte';
</script>
`,
			expected: []string{"./utils.js", "../lib/helpers.js", "./Component.svelte"},
		},
		{
			name: "import image assets",
			source: `
<script>
import logo from './assets/logo.png';
import icon from '/images/icon.svg';
</script>
`,
			expected: []string{"./assets/logo.png", "/images/icon.svg"},
		},
		{
			name: "import JSON data",
			source: `
<script>
import config from './config.json';
import data from '../data/users.json';
</script>
`,
			expected: []string{"./config.json", "../data/users.json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imports := handler.parseImports(tt.source)
			if len(imports) != len(tt.expected) {
				t.Errorf("Expected %d imports, got %d", len(tt.expected), len(imports))
				return
			}
			for i, imp := range imports {
				if imp != tt.expected[i] {
					t.Errorf("Expected import %s, got %s", tt.expected[i], imp)
				}
			}
		})
	}
}

func TestSvelteHandler_AssetTypeDetection(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	handler := NewSvelteHandler(fs)

	tests := []struct {
		path     string
		expected string
	}{
		{"component.svelte", "svelte"},
		{"script.js", "javascript"},
		{"module.mjs", "javascript"},
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
			assetType := handler.getAssetType(tt.path)
			if assetType != tt.expected {
				t.Errorf("Expected asset type %s for %s, got %s", tt.expected, tt.path, assetType)
			}
		})
	}
}

func TestSvelteHandler_ResolveAssetPath(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	handler := NewSvelteHandler(fs)

	// Create test files
	_ = fs.WriteFile("routes/component.svelte", []byte(""))
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
			currentPath: "routes/component.svelte",
			wantPath:    "routes/styles.css",
			wantType:    "stylesheet",
		},
		{
			name:        "absolute image in public",
			importPath:  "/logo.png",
			currentPath: "routes/component.svelte",
			wantPath:    "public/logo.png",
			wantType:    "image",
		},
		{
			name:        "nested image in public",
			importPath:  "/images/icon.svg",
			currentPath: "routes/component.svelte",
			wantPath:    "public/images/icon.svg",
			wantType:    "image",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, gotType := handler.resolveAssetPath(tt.importPath, tt.currentPath)
			if gotPath != tt.wantPath {
				t.Errorf("Expected path %s, got %s", tt.wantPath, gotPath)
			}
			if gotType != tt.wantType {
				t.Errorf("Expected type %s, got %s", tt.wantType, gotType)
			}
		})
	}
}

func TestSvelteHandler_TransformAssetImports(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	handler := NewSvelteHandler(fs)

	// Create test files
	_ = fs.WriteFile("routes/component.svelte", []byte(""))
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

	imports := make(map[string]string)
	transformed := handler.transformToIIFE(jsCode, "Component", imports, "routes/component.svelte")

	// Check that CSS import was transformed to URL
	if !strings.Contains(transformed, `const styles = '/styles.css'`) {
		t.Error("Expected CSS import to be transformed to URL constant")
	}

	// Check that image import was transformed to URL
	if !strings.Contains(transformed, `const logo = '/logo.png'`) {
		t.Error("Expected image import to be transformed to URL constant")
	}

	// Check that Svelte import was preserved in imports map
	if imports["Button"] != "./Button.svelte" {
		t.Error("Expected Svelte import to be preserved in imports map")
	}

	// Check that import statements were removed
	if strings.Contains(transformed, "import ") {
		t.Error("Expected import statements to be removed")
	}
}

func TestSvelteHandler_PublicURLGeneration(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	handler := NewSvelteHandler(fs)

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
			url := handler.getPublicURL(tt.assetPath)
			if url != tt.expected {
				t.Errorf("Expected URL %s for path %s, got %s", tt.expected, tt.assetPath, url)
			}
		})
	}
}

func TestSvelteHandler_JavaScriptImportTransformation(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	handler := NewSvelteHandler(fs)

	// Create test JS files
	_ = fs.WriteFile("routes/utils.js", []byte(`export function hello() { return "Hello"; }`))
	_ = fs.WriteFile("public/js/library.mjs", []byte(`export const VERSION = "1.0.0";`))

	// Initialize compiler
	_ = handler.initializeCompiler()

	tests := []struct {
		name       string
		jsCode     string
		wantResult string
	}{
		{
			name:       "relative JS import",
			jsCode:     `import utils from './utils.js';`,
			wantResult: `const utils = '/utils.js';`,
		},
		{
			name:       "absolute JS import from public",
			jsCode:     `import lib from '/js/library.mjs';`,
			wantResult: `const lib = '/js/library.mjs';`,
		},
		{
			name:       "ES module import",
			jsCode:     `import { helpers } from '../lib/helpers.mjs';`,
			wantResult: ``, // This style of import is removed but not transformed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Transform the code
			transformed := handler.transformToIIFE(tt.jsCode, "TestComponent", make(map[string]string), "routes/component.svelte")
			
			// Check if the result contains what we expect
			if tt.wantResult != "" && !strings.Contains(transformed, tt.wantResult) {
				t.Errorf("Expected transformed code to contain:\n%s\nGot:\n%s", tt.wantResult, transformed)
			}
			
			// Ensure import was removed
			if strings.Contains(transformed, "import ") {
				t.Error("Import statement should be removed")
			}
		})
	}
}

func TestSvelteHandler_JSONImportInlining(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	handler := NewSvelteHandler(fs)

	// Create test JSON files
	smallJSON := `{"name": "test", "version": "1.0.0", "data": [1, 2, 3]}`
	_ = fs.WriteFile("routes/config.json", []byte(smallJSON))
	
	// Create a large JSON file (over 100KB)
	largeData := make([]int, 30000)
	for i := range largeData {
		largeData[i] = i
	}
	largeJSON := fmt.Sprintf(`{"large": %v}`, largeData)
	_ = fs.WriteFile("routes/large.json", []byte(largeJSON))
	
	// Create invalid JSON
	_ = fs.WriteFile("routes/invalid.json", []byte(`{invalid json`))

	tests := []struct {
		name       string
		jsCode     string
		wantResult string
		wantURL    bool
	}{
		{
			name:       "small JSON file should be inlined",
			jsCode:     `import config from './config.json';`,
			wantResult: fmt.Sprintf("const config = %s;", smallJSON),
			wantURL:    false,
		},
		{
			name:       "large JSON file should use URL",
			jsCode:     `import large from './large.json';`,
			wantResult: `const large = '/large.json';`,
			wantURL:    true,
		},
		{
			name:       "invalid JSON should use URL",
			jsCode:     `import invalid from './invalid.json';`,
			wantResult: `const invalid = '/invalid.json';`,
			wantURL:    true,
		},
		{
			name:       "non-existent JSON should not be transformed",
			jsCode:     `import missing from './missing.json';`,
			wantResult: ``,
			wantURL:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Transform the code
			transformed := handler.transformToIIFE(tt.jsCode, "TestComponent", make(map[string]string), "routes/component.svelte")
			
			// Check if the result contains what we expect
			if tt.wantResult != "" {
				if tt.wantURL {
					// Should be a URL string
					if !strings.Contains(transformed, tt.wantResult) {
						t.Errorf("Expected transformed code to contain URL assignment:\n%s\nGot:\n%s", tt.wantResult, transformed)
					}
				} else {
					// Should be inlined JSON data
					if !strings.Contains(transformed, "const config = {") {
						t.Errorf("Expected transformed code to contain inlined JSON data, got:\n%s", transformed)
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