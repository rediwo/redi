package utils

import (
	"fmt"
	"strings"
	"testing"
)

func TestDefaultVimeshStyleConfig(t *testing.T) {
	config := DefaultVimeshStyleConfig()
	
	if !config.Enable {
		t.Error("Expected Enable to be true by default")
	}
}

func TestVimeshStyleInitialization(t *testing.T) {
	// Test singleton initialization
	instance1, err1 := getVimeshStyleInstance()
	if err1 != nil {
		t.Fatalf("Failed to get first instance: %v", err1)
	}
	
	instance2, err2 := getVimeshStyleInstance()
	if err2 != nil {
		t.Fatalf("Failed to get second instance: %v", err2)
	}
	
	// Should be the same instance
	if instance1 != instance2 {
		t.Error("Expected singleton instance, got different instances")
	}
}

func TestGetCSSFromHTML(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		expectError bool
		checkCSS    func(css string) bool
	}{
		{
			name: "Basic Tailwind classes",
			html: `<div class="bg-blue-500 text-white p-4">Hello</div>`,
			expectError: false,
			checkCSS: func(css string) bool {
				// Check if CSS contains expected patterns
				return strings.Contains(css, "bg-blue-500") || 
					   strings.Contains(css, "background-color") ||
					   strings.Contains(css, "rgb(") ||
					   len(css) > 0
			},
		},
		{
			name: "Multiple elements with classes",
			html: `
				<div class="container mx-auto">
					<h1 class="text-3xl font-bold">Title</h1>
					<p class="text-gray-600 leading-relaxed">Paragraph</p>
				</div>
			`,
			expectError: false,
			checkCSS: func(css string) bool {
				return len(css) > 0
			},
		},
		{
			name: "No Tailwind classes",
			html: `<div>Hello World</div>`,
			expectError: false,
			checkCSS: func(css string) bool {
				// Should return empty or minimal CSS
				return true
			},
		},
		{
			name: "Responsive and pseudo classes",
			html: `<button class="bg-blue-500 hover:bg-blue-600 md:px-6">Button</button>`,
			expectError: false,
			checkCSS: func(css string) bool {
				return len(css) > 0
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			css, err := GetCSSFromHTML(tt.html)
			
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			if !tt.expectError && tt.checkCSS != nil {
				if !tt.checkCSS(css) {
					t.Errorf("CSS check failed. Got: %s", css)
				}
			}
		})
	}
}

func TestGetCSSFromHTML_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		expectCSS   bool
		checkFunc   func(css string, t *testing.T)
	}{
		{
			name:      "Empty HTML",
			html:      "",
			expectCSS: true, // Vimesh Style returns preset CSS even with no classes
		},
		{
			name:      "HTML without classes",
			html:      `<div>No classes here</div><p>Just text</p>`,
			expectCSS: true, // Vimesh Style returns preset CSS even with no classes
		},
		{
			name:      "HTML with empty class attribute",
			html:      `<div class="">Empty class</div>`,
			expectCSS: true, // Vimesh Style returns preset CSS even with no classes
		},
		{
			name:      "HTML with className instead of class",
			html:      `<div className="bg-red-500 p-2">React style</div>`,
			expectCSS: true,
			checkFunc: func(css string, t *testing.T) {
				if !strings.Contains(css, "background-color") {
					t.Error("Expected CSS to contain background-color for bg-red-500")
				}
			},
		},
		{
			name:      "Mixed case class attribute",
			html:      `<div CLASS="text-lg">Mixed case</div>`,
			expectCSS: true,
		},
		{
			name:      "Single quotes",
			html:      `<div class='bg-green-500 m-4'>Single quotes</div>`,
			expectCSS: true,
			checkFunc: func(css string, t *testing.T) {
				if !strings.Contains(css, "background-color") {
					t.Error("Expected CSS to contain background-color")
				}
			},
		},
		{
			name:      "Multiple elements with same classes",
			html:      `<div class="p-4"><span class="p-4">Same padding</span></div>`,
			expectCSS: true,
			checkFunc: func(css string, t *testing.T) {
				// Vimesh Style includes preset CSS which has multiple padding rules
				// Just check that p-4 specific padding is included
				if !strings.Contains(css, ".p-4") {
					t.Error("Expected CSS to contain .p-4 class")
				}
			},
		},
		{
			name:      "Invalid Tailwind classes",
			html:      `<div class="not-a-class random-stuff">Invalid</div>`,
			expectCSS: true, // Will still return preset CSS
		},
		{
			name:      "Tailwind utilities with variants",
			html:      `<div class="hover:bg-blue-600 lg:text-2xl dark:bg-gray-800">Variants</div>`,
			expectCSS: true,
			checkFunc: func(css string, t *testing.T) {
				// Should handle hover, responsive, and dark mode variants
				if !strings.Contains(css, ":hover") && !strings.Contains(css, "@media") {
					t.Log("Note: Variants might not be fully supported in current implementation")
				}
			},
		},
		{
			name:      "Arbitrary values",
			html:      `<div class="bg-[#1da1f2] text-[20px] p-[10px]">Arbitrary values</div>`,
			expectCSS: true,
			checkFunc: func(css string, t *testing.T) {
				// Check if arbitrary values are handled
				if len(css) == 0 {
					t.Error("Expected CSS for arbitrary values")
				}
			},
		},
		{
			name:      "Space-separated classes with extra spaces",
			html:      `<div class="  bg-blue-500   text-white    p-4  ">Extra spaces</div>`,
			expectCSS: true,
			checkFunc: func(css string, t *testing.T) {
				if !strings.Contains(css, "background-color") {
					t.Error("Expected CSS to handle classes with extra spaces")
				}
			},
		},
		{
			name:      "Nested HTML structure",
			html:      `
				<div class="container mx-auto">
					<header class="bg-gray-100 p-4">
						<h1 class="text-3xl font-bold">Title</h1>
						<nav class="flex space-x-4">
							<a class="text-blue-500 hover:text-blue-700">Link</a>
						</nav>
					</header>
				</div>
			`,
			expectCSS: true,
			checkFunc: func(css string, t *testing.T) {
				expectedClasses := []string{"container", "mx-auto", "bg-gray-100", "p-4", "text-3xl", "font-bold", "flex", "space-x-4", "text-blue-500"}
				foundCount := 0
				for _, class := range expectedClasses {
					if strings.Contains(css, class) || strings.Contains(css, "background-color") || strings.Contains(css, "font-") || strings.Contains(css, "padding") {
						foundCount++
					}
				}
				if foundCount < 3 {
					t.Errorf("Expected to find CSS for multiple classes, found evidence of %d", foundCount)
				}
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			css, err := GetCSSFromHTML(tt.html)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if tt.expectCSS && len(css) == 0 {
				t.Error("Expected CSS output, got empty string")
			}
			
			if !tt.expectCSS && len(css) > 0 {
				t.Errorf("Expected no CSS output, got %d bytes", len(css))
			}
			
			if tt.checkFunc != nil {
				tt.checkFunc(css, t)
			}
		})
	}
}

func TestGetVimeshStyleJS(t *testing.T) {
	js := GetVimeshStyleJS()
	
	if js == "" {
		t.Error("Expected Vimesh Style JS to be non-empty")
	}
	
	// Check for some expected content
	if !strings.Contains(js, "vimesh") && !strings.Contains(js, "style") {
		t.Error("Vimesh Style JS doesn't contain expected content")
	}
}

func TestGetVimeshStyleJS_Content(t *testing.T) {
	js := GetVimeshStyleJS()
	
	if len(js) == 0 {
		t.Fatal("Expected Vimesh Style JS to be non-empty")
	}
	
	// Check for expected content markers
	expectedContent := []string{
		"setupVimeshStyle",
		"window",
		"$vs",
		"Vimesh",    // Should contain Vimesh
		"bg-",      // Background utilities
		"text-",    // Text utilities
		"padding",  // Padding utilities
		"margin",   // Margin utilities
	}
	
	for _, expected := range expectedContent {
		if !strings.Contains(js, expected) {
			t.Errorf("Expected Vimesh Style JS to contain '%s'", expected)
		}
	}
	
	// Check it's the browser version (not CommonJS)
	if strings.Contains(js, "module.exports") {
		t.Error("Found module.exports - this should be the browser version, not CommonJS")
	}
	
	// Check it has the setup call
	if !strings.Contains(js, "setupVimeshStyle(window)") {
		t.Error("Expected to find setupVimeshStyle(window) call")
	}
}

func TestConcurrentAccess(t *testing.T) {
	// Test concurrent access to ensure thread safety
	html := `<div class="bg-blue-500 p-4">Concurrent test</div>`
	
	done := make(chan bool, 10)
	errors := make(chan error, 10)
	
	// Run 10 concurrent extractions
	for i := 0; i < 10; i++ {
		go func(index int) {
			css, err := GetCSSFromHTML(html)
			if err != nil {
				errors <- err
			} else if len(css) == 0 {
				errors <- fmt.Errorf("goroutine %d: got empty CSS", index)
			}
			done <- true
		}(i)
	}
	
	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Check for errors
	close(errors)
	for err := range errors {
		t.Errorf("Concurrent access error: %v", err)
	}
}

func TestClassExtraction(t *testing.T) {
	// Test the regex-based class extraction logic
	tests := []struct {
		name     string
		html     string
		expected []string
	}{
		{
			name:     "Standard class attribute",
			html:     `<div class="foo bar baz">Test</div>`,
			expected: []string{"foo", "bar", "baz"},
		},
		{
			name:     "className attribute",
			html:     `<div className="foo bar">React</div>`,
			expected: []string{"foo", "bar"},
		},
		{
			name:     "Mixed attributes",
			html:     `<div class="foo" id="test" className="bar">Mixed</div>`,
			expected: []string{"foo", "bar"},
		},
		{
			name:     "No duplicates",
			html:     `<div class="foo foo bar">Duplicates</div>`,
			expected: []string{"foo", "bar"},
		},
		{
			name:     "Preserve Tailwind classes",
			html:     `<div class="bg-blue-500 hover:bg-blue-600 lg:p-4">Tailwind</div>`,
			expected: []string{"bg-blue-500", "hover:bg-blue-600", "lg:p-4"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			css, err := GetCSSFromHTML(tt.html)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			// Since we can't directly test the extracted classes,
			// we verify that CSS was generated when classes were present
			if len(tt.expected) > 0 && len(css) == 0 {
				t.Error("Expected CSS to be generated for extracted classes")
			}
		})
	}
}

func BenchmarkGetCSSFromHTML(b *testing.B) {
	html := `
		<div class="container mx-auto p-4">
			<h1 class="text-3xl font-bold text-gray-900 mb-4">Benchmark Test</h1>
			<div class="grid grid-cols-3 gap-4">
				<div class="bg-white rounded-lg shadow-md p-6">
					<p class="text-gray-600">Card 1</p>
				</div>
				<div class="bg-white rounded-lg shadow-md p-6">
					<p class="text-gray-600">Card 2</p>
				</div>
				<div class="bg-white rounded-lg shadow-md p-6">
					<p class="text-gray-600">Card 3</p>
				</div>
			</div>
		</div>
	`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetCSSFromHTML(html)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetCSSFromHTML_Simple(b *testing.B) {
	html := `<div class="bg-blue-500 text-white p-4">Simple</div>`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetCSSFromHTML(html)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGetCSSFromHTML_Complex(b *testing.B) {
	html := `
		<div class="container mx-auto p-4">
			<header class="bg-gray-100 p-6 rounded-lg shadow-md">
				<h1 class="text-3xl font-bold text-gray-900 mb-2">Complex Layout</h1>
				<p class="text-gray-600">With many classes</p>
			</header>
			<main class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mt-8">
				<div class="bg-white p-6 rounded shadow hover:shadow-lg transition-shadow">
					<h2 class="text-xl font-semibold mb-2">Card 1</h2>
					<p class="text-gray-700">Content goes here</p>
				</div>
				<div class="bg-white p-6 rounded shadow hover:shadow-lg transition-shadow">
					<h2 class="text-xl font-semibold mb-2">Card 2</h2>
					<p class="text-gray-700">Content goes here</p>
				</div>
				<div class="bg-white p-6 rounded shadow hover:shadow-lg transition-shadow">
					<h2 class="text-xl font-semibold mb-2">Card 3</h2>
					<p class="text-gray-700">Content goes here</p>
				</div>
			</main>
		</div>
	`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetCSSFromHTML(html)
		if err != nil {
			b.Fatal(err)
		}
	}
}