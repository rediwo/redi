package utils

import (
	_ "embed"
	"fmt"
	"sync"

	"github.com/dop251/goja"
)

//go:embed vimesh_style.js
var vimeshStyleJS string

// vimeshStyleInstance holds a Goja runtime instance for Vimesh Style
type vimeshStyleInstance struct {
	vm            *goja.Runtime
	initialized   bool
	mu            sync.Mutex
}

var (
	globalVimeshInstance *vimeshStyleInstance
	vimeshOnce          sync.Once
)

// getVimeshStyleInstance returns a singleton instance of Vimesh Style
func getVimeshStyleInstance() (*vimeshStyleInstance, error) {
	var err error
	vimeshOnce.Do(func() {
		globalVimeshInstance = &vimeshStyleInstance{}
		err = globalVimeshInstance.initialize()
	})
	return globalVimeshInstance, err
}

func (v *vimeshStyleInstance) initialize() error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.initialized {
		return nil
	}

	v.vm = goja.New()
	v.vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	// Create window object as global
	window := v.vm.NewObject()
	v.vm.Set("window", window)
	v.vm.Set("global", window)
	v.vm.Set("self", window)

	// Add console for debugging
	console := v.vm.NewObject()
	console.Set("log", func(args ...interface{}) {
		// Silent logging
	})
	console.Set("error", func(args ...interface{}) {
		// Silent logging
	})
	console.Set("warn", func(args ...interface{}) {
		// Silent logging
	})
	v.vm.Set("console", console)

	// Load Vimesh Style (it will call setupVimeshStyle(window) automatically)
	_, err := v.vm.RunString(vimeshStyleJS)
	if err != nil {
		return fmt.Errorf("failed to load Vimesh Style: %w", err)
	}

	// Verify that window.$vs exists
	windowObj := v.vm.Get("window")
	if windowObj == nil {
		return fmt.Errorf("window object not found")
	}
	
	windowValue := windowObj.ToObject(v.vm)
	if windowValue == nil {
		return fmt.Errorf("window is not an object")
	}
	
	vsValue := windowValue.Get("$vs")
	if vsValue == nil {
		return fmt.Errorf("window.$vs not found - Vimesh Style did not initialize properly")
	}

	// Override the extract function to work around named groups issue in Goja
	_, err = v.vm.RunString(`
		window.$vs.extract = function(html) {
			var match;
			var classes = [];
			var regex = /class\s*=\s*['""]([^'""]*)['"]/g;
			while ((match = regex.exec(html)) !== null) {
				if (match[1]) {
					var classList = match[1].split(' ');
					for (var i = 0; i < classList.length; i++) {
						var cls = classList[i].trim();
						if (cls && classes.indexOf(cls) === -1) {
							classes.push(cls);
						}
					}
				}
			}
			return classes;
		};
	`)
	if err != nil {
		return fmt.Errorf("failed to override extract function: %w", err)
	}

	v.initialized = true
	return nil
}

// GetCSSFromHTML extracts CSS from HTML using Vimesh Style
func GetCSSFromHTML(html string) (string, error) {
	instance, err := getVimeshStyleInstance()
	if err != nil {
		return "", err
	}

	instance.mu.Lock()
	defer instance.mu.Unlock()

	// Get window.$vs object
	windowObj := instance.vm.Get("window")
	if windowObj == nil {
		return "", fmt.Errorf("window object not found")
	}
	
	windowValue := windowObj.ToObject(instance.vm)
	if windowValue == nil {
		return "", fmt.Errorf("window is not an object")
	}
	
	vsObj := windowValue.Get("$vs")
	if vsObj == nil {
		return "", fmt.Errorf("window.$vs not found")
	}
	
	vs := vsObj.ToObject(instance.vm)
	if vs == nil {
		return "", fmt.Errorf("window.$vs is not an object")
	}

	// Get extract function
	extractFunc := vs.Get("extract")
	if extractFunc == nil {
		return "", fmt.Errorf("extract function not found")
	}

	extractCallable, ok := goja.AssertFunction(extractFunc)
	if !ok {
		return "", fmt.Errorf("extract is not a function")
	}

	// Extract classes from HTML
	extractedValue, err := extractCallable(goja.Undefined(), instance.vm.ToValue(html))
	if err != nil {
		return "", fmt.Errorf("failed to extract classes: %w", err)
	}

	// Get add function
	addFunc := vs.Get("add")
	if addFunc == nil {
		return "", fmt.Errorf("add function not found")
	}

	addCallable, ok := goja.AssertFunction(addFunc)
	if !ok {
		return "", fmt.Errorf("add is not a function")
	}

	// Add extracted classes
	_, err = addCallable(goja.Undefined(), extractedValue)
	if err != nil {
		return "", fmt.Errorf("failed to add classes: %w", err)
	}

	// Get styles
	stylesValue := vs.Get("styles")
	if stylesValue == nil || goja.IsUndefined(stylesValue) || goja.IsNull(stylesValue) {
		return "", nil
	}

	styles := stylesValue.String()
	return styles, nil
}

// GetVimeshStyleJS returns the embedded Vimesh Style JavaScript code
func GetVimeshStyleJS() string {
	return vimeshStyleJS
}

// VimeshStyleConfig holds configuration for Vimesh Style integration
type VimeshStyleConfig struct {
	Enable bool `json:"enable"` // Enable Vimesh Style support
}

// DefaultVimeshStyleConfig returns default configuration
func DefaultVimeshStyleConfig() *VimeshStyleConfig {
	return &VimeshStyleConfig{
		Enable: true,
	}
}