package handlers

import (
	"testing"
	"github.com/rediwo/redi/filesystem"
)

func TestSvelteHandler_ParseImports(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	handler := NewSvelteHandler(fs)

	tests := []struct {
		name     string
		source   string
		expected []string
	}{
		{
			name: "single import",
			source: `
<script>
import Button from './Button.svelte';
</script>
`,
			expected: []string{"./Button.svelte"},
		},
		{
			name: "multiple imports",
			source: `
<script>
import Button from './Button.svelte';
import Card from '../shared/Card.svelte';
import { Modal } from './Modal.svelte';
</script>
`,
			expected: []string{"./Button.svelte", "../shared/Card.svelte", "./Modal.svelte"},
		},
		{
			name: "no imports",
			source: `
<script>
let count = 0;
</script>
`,
			expected: []string{},
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

func TestSvelteHandler_ResolveComponentPath(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	handler := NewSvelteHandler(fs)

	tests := []struct {
		name        string
		importPath  string
		currentPath string
		expected    string
	}{
		{
			name:        "relative same directory",
			importPath:  "./Button.svelte",
			currentPath: "routes/page.svelte",
			expected:    "routes/Button.svelte",
		},
		{
			name:        "relative parent directory",
			importPath:  "../shared/Card.svelte",
			currentPath: "routes/components/List.svelte",
			expected:    "routes/shared/Card.svelte",
		},
		{
			name:        "simple filename",
			importPath:  "Button.svelte",
			currentPath: "routes/page.svelte",
			expected:    "routes/Button.svelte",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.resolveComponentPath(tt.importPath, tt.currentPath)
			if result != tt.expected {
				t.Errorf("Expected path %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestSvelteHandler_ComponentImports(t *testing.T) {
	fs := filesystem.NewMemoryFileSystem()
	handler := NewSvelteHandler(fs)

	// Create test components
	buttonComponent := `
<script>
    export let text = 'Click me';
</script>
<button>{text}</button>
<style>
    button { padding: 10px; }
</style>
`
	
	pageComponent := `
<script>
    import Button from './Button.svelte';
</script>
<main>
    <h1>Test Page</h1>
    <Button text="Hello" />
</main>
`

	// Write components to filesystem
	err := fs.WriteFile("routes/Button.svelte", []byte(buttonComponent))
	if err != nil {
		t.Fatalf("Failed to write Button component: %v", err)
	}
	
	err = fs.WriteFile("routes/page.svelte", []byte(pageComponent))
	if err != nil {
		t.Fatalf("Failed to write page component: %v", err)
	}

	// Initialize compiler
	err = handler.initializeCompiler()
	if err != nil {
		t.Fatalf("Failed to initialize compiler: %v", err)
	}

	// Collect dependencies for the page component
	deps, err := handler.collectAllDependencies("routes/page.svelte", nil)
	if err != nil {
		t.Fatalf("Failed to collect dependencies: %v", err)
	}

	// Should have 2 components: Button and page
	if len(deps) != 2 {
		t.Errorf("Expected 2 components, got %d", len(deps))
	}

	// First should be Button (dependency), second should be page (main)
	if len(deps) >= 2 {
		if deps[0].Name != "Button" {
			t.Errorf("Expected first component to be Button, got %s", deps[0].Name)
		}
		if deps[1].Name != "page" {
			t.Errorf("Expected second component to be page, got %s", deps[1].Name)
		}
	}
}