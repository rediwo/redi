package path

import (
	"testing"
	
	js "github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
)

func TestPathModule(t *testing.T) {
	// Set up VM with path module
	vm := js.New()
	registry := require.NewRegistry()
	Enable(registry)
	registry.Enable(vm)
	
	// Get path module
	path := require.Require(vm, "path")
	vm.Set("path", path)
	
	t.Run("join", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{`path.join("a", "b", "c")`, "a/b/c"},
			{`path.join("/a", "b", "c")`, "/a/b/c"},
			{`path.join("a", "", "c")`, "a/c"},
			{`path.join("a", ".", "c")`, "a/c"},
			{`path.join("a", "..", "c")`, "c"},
		}
		
		for _, test := range tests {
			result, err := vm.RunString(test.input)
			if err != nil {
				t.Fatalf("Failed to run %s: %v", test.input, err)
			}
			
			if result.String() != test.expected {
				t.Errorf("For %s: expected %q, got %q", test.input, test.expected, result.String())
			}
		}
	})
	
	t.Run("dirname", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{`path.dirname("/a/b/c.txt")`, "/a/b"},
			{`path.dirname("/a/b/c")`, "/a/b"},
			{`path.dirname("a/b/c.txt")`, "a/b"},
			{`path.dirname("file.txt")`, "."},
			{`path.dirname("/")`, "/"},
		}
		
		for _, test := range tests {
			result, err := vm.RunString(test.input)
			if err != nil {
				t.Fatalf("Failed to run %s: %v", test.input, err)
			}
			
			if result.String() != test.expected {
				t.Errorf("For %s: expected %q, got %q", test.input, test.expected, result.String())
			}
		}
	})
	
	t.Run("basename", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{`path.basename("/a/b/c.txt")`, "c.txt"},
			{`path.basename("/a/b/c")`, "c"},
			{`path.basename("file.txt")`, "file.txt"},
			{`path.basename("/")`, "/"},
			{`path.basename(".")`, "."},
		}
		
		for _, test := range tests {
			result, err := vm.RunString(test.input)
			if err != nil {
				t.Fatalf("Failed to run %s: %v", test.input, err)
			}
			
			if result.String() != test.expected {
				t.Errorf("For %s: expected %q, got %q", test.input, test.expected, result.String())
			}
		}
	})
	
	t.Run("basename with extension", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{`path.basename("/a/b/c.txt", ".txt")`, "c"},
			{`path.basename("file.js", ".js")`, "file"},
			{`path.basename("file.txt", ".js")`, "file.txt"}, // ext doesn't match
		}
		
		for _, test := range tests {
			result, err := vm.RunString(test.input)
			if err != nil {
				t.Fatalf("Failed to run %s: %v", test.input, err)
			}
			
			if result.String() != test.expected {
				t.Errorf("For %s: expected %q, got %q", test.input, test.expected, result.String())
			}
		}
	})
	
	t.Run("extname", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{`path.extname("file.txt")`, ".txt"},
			{`path.extname("file.tar.gz")`, ".gz"},
			{`path.extname("file")`, ""},
			{`path.extname(".file")`, ""},
			{`path.extname("file.")`, "."},
		}
		
		for _, test := range tests {
			result, err := vm.RunString(test.input)
			if err != nil {
				t.Fatalf("Failed to run %s: %v", test.input, err)
			}
			
			if result.String() != test.expected {
				t.Errorf("For %s: expected %q, got %q", test.input, test.expected, result.String())
			}
		}
	})
	
	t.Run("isAbsolute", func(t *testing.T) {
		tests := []struct {
			input    string
			expected bool
		}{
			{`path.isAbsolute("/absolute/path")`, true},
			{`path.isAbsolute("relative/path")`, false},
			{`path.isAbsolute("./relative")`, false},
			{`path.isAbsolute("../relative")`, false},
			{`path.isAbsolute("")`, false},
		}
		
		for _, test := range tests {
			result, err := vm.RunString(test.input)
			if err != nil {
				t.Fatalf("Failed to run %s: %v", test.input, err)
			}
			
			if result.ToBoolean() != test.expected {
				t.Errorf("For %s: expected %t, got %t", test.input, test.expected, result.ToBoolean())
			}
		}
	})
	
	t.Run("normalize", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{`path.normalize("/a/b/../c")`, "/a/c"},
			{`path.normalize("a/b/../c")`, "a/c"},
			{`path.normalize("a/b/./c")`, "a/b/c"},
			{`path.normalize("a//b/c")`, "a/b/c"},
			{`path.normalize(".")`, "."},
		}
		
		for _, test := range tests {
			result, err := vm.RunString(test.input)
			if err != nil {
				t.Fatalf("Failed to run %s: %v", test.input, err)
			}
			
			if result.String() != test.expected {
				t.Errorf("For %s: expected %q, got %q", test.input, test.expected, result.String())
			}
		}
	})
	
	t.Run("relative", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{`path.relative("/a/b", "/a/b/c")`, "c"},
			{`path.relative("/a/b/c", "/a/b")`, ".."},
			{`path.relative("/a/b", "/x/y")`, "../../x/y"},
		}
		
		for _, test := range tests {
			result, err := vm.RunString(test.input)
			if err != nil {
				t.Fatalf("Failed to run %s: %v", test.input, err)
			}
			
			if result.String() != test.expected {
				t.Errorf("For %s: expected %q, got %q", test.input, test.expected, result.String())
			}
		}
	})
	
	t.Run("parse", func(t *testing.T) {
		result, err := vm.RunString(`
			var parsed = path.parse("/home/user/documents/file.txt");
			({
				dir: parsed.dir,
				base: parsed.base,
				ext: parsed.ext,
				name: parsed.name
			})
		`)
		if err != nil {
			t.Fatalf("Failed to parse path: %v", err)
		}
		
		obj := result.(*js.Object)
		dir := obj.Get("dir").String()
		base := obj.Get("base").String()
		ext := obj.Get("ext").String()
		name := obj.Get("name").String()
		
		if dir != "/home/user/documents" {
			t.Errorf("Expected dir '/home/user/documents', got %q", dir)
		}
		if base != "file.txt" {
			t.Errorf("Expected base 'file.txt', got %q", base)
		}
		if ext != ".txt" {
			t.Errorf("Expected ext '.txt', got %q", ext)
		}
		if name != "file" {
			t.Errorf("Expected name 'file', got %q", name)
		}
	})
	
	t.Run("format", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{`path.format({dir: "/home/user", base: "file.txt"})`, "/home/user/file.txt"},
			{`path.format({dir: "/home/user", name: "file", ext: ".txt"})`, "/home/user/file.txt"},
			{`path.format({base: "file.txt"})`, "file.txt"},
			{`path.format({dir: "/home/user"})`, "/home/user"},
		}
		
		for _, test := range tests {
			result, err := vm.RunString(test.input)
			if err != nil {
				t.Fatalf("Failed to run %s: %v", test.input, err)
			}
			
			if result.String() != test.expected {
				t.Errorf("For %s: expected %q, got %q", test.input, test.expected, result.String())
			}
		}
	})
	
	t.Run("resolve", func(t *testing.T) {
		// Test resolve with relative paths
		result, err := vm.RunString(`path.resolve(".")`)
		if err != nil {
			t.Fatalf("Failed to resolve current directory: %v", err)
		}
		
		// Should return an absolute path
		resultStr := result.String()
		if len(resultStr) == 0 {
			t.Error("resolve should return a non-empty path")
		}
		
		// Test with multiple segments
		result, err = vm.RunString(`path.resolve("a", "b", "c")`)
		if err != nil {
			t.Fatalf("Failed to resolve path segments: %v", err)
		}
		
		resultStr = result.String()
		if len(resultStr) == 0 {
			t.Error("resolve should return a non-empty path")
		}
	})
	
	t.Run("constants", func(t *testing.T) {
		// Test sep constant
		result, err := vm.RunString(`path.sep`)
		if err != nil {
			t.Fatalf("Failed to get path.sep: %v", err)
		}
		
		sep := result.String()
		if sep != "/" {
			t.Errorf("Expected sep to be '/', got %q", sep)
		}
		
		// Test delimiter constant
		result, err = vm.RunString(`path.delimiter`)
		if err != nil {
			t.Fatalf("Failed to get path.delimiter: %v", err)
		}
		
		delimiter := result.String()
		if delimiter != ":" {
			t.Errorf("Expected delimiter to be ':', got %q", delimiter)
		}
	})
	
	t.Run("error handling", func(t *testing.T) {
		// Test functions with missing arguments
		errorTests := []string{
			`path.dirname()`,
			`path.basename()`,
			`path.extname()`,
			`path.isAbsolute()`,
			`path.normalize()`,
			`path.parse()`,
		}
		
		for _, test := range errorTests {
			_, err := vm.RunString(test)
			if err == nil {
				t.Errorf("Expected error for %s", test)
			}
		}
		
		// Test relative with insufficient arguments
		_, err := vm.RunString(`path.relative("/a")`)
		if err == nil {
			t.Error("Expected error for path.relative with one argument")
		}
		
		// Test format with invalid argument
		_, err = vm.RunString(`path.format("not an object")`)
		if err == nil {
			t.Error("Expected error for path.format with string argument")
		}
	})
}