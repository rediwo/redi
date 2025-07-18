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

// Asset type detection is now tested in import_transformer_test.go

// Asset path resolution is now tested in import_transformer_test.go

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

// Public URL generation is now tested in import_transformer_test.go

func TestSvelteHandler_JavaScriptImportTransformation(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	handler := NewSvelteHandler(fs)

	// Create test JS files with ES6 exports
	utilsJS := `export function hello() { return "Hello"; }
export const version = "1.0.0";
export default { hello, version };`
	_ = fs.WriteFile("routes/utils.js", []byte(utilsJS))
	
	// Create CommonJS module
	commonJS := `function greet(name) { return "Hi " + name; }
module.exports = { greet };`
	_ = fs.WriteFile("routes/common.js", []byte(commonJS))
	
	// Create large JS file
	largeJS := strings.Repeat("export function test() { return 'test'; }\n", 2000)
	_ = fs.WriteFile("routes/large.js", []byte(largeJS))

	// Initialize compiler
	_ = handler.initializeCompiler()

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
			name:   "Large JS file is also transformed",
			jsCode: `import large from './large.js';`,
			checkFunction: func(transformed string) bool {
				// Should also be transformed to IIFE (no size limit)
				return strings.Contains(transformed, "const large = (function()") &&
					strings.Contains(transformed, "__exports")
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
			transformed := handler.transformToIIFE(tt.jsCode, "TestComponent", make(map[string]string), "routes/component.svelte")
			
			// Check if the transformation is correct
			if !tt.checkFunction(transformed) {
				t.Errorf("Transformation did not produce expected result.\nGot:\n%s", transformed)
			}
			
			// Ensure import was removed
			if strings.Contains(transformed, "import ") && !strings.Contains(transformed, "await import") {
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
		wantURL    bool
	}{
		{
			name:       "small JSON file should be inlined",
			jsCode:     `import config from './config.json';`,
			wantResult: fmt.Sprintf("const config = %s;", smallJSON),
			wantURL:    false,
		},
		{
			name:       "large JSON file should also be inlined",
			jsCode:     `import large from './large.json';`,
			wantResult: fmt.Sprintf("const large = %s;", largeJSON),
			wantURL:    false,
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
					if !strings.Contains(transformed, "const config = {") && !strings.Contains(transformed, "const large = {") {
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