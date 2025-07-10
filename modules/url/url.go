package url

import (
	"github.com/dop251/goja_nodejs/require"
	"github.com/dop251/goja_nodejs/url"
	"github.com/rediwo/redi/registry"
)

// init registers the url module automatically
func init() {
	registry.RegisterModule("url", initURLModule)
}

// initURLModule initializes the url module
func initURLModule(config registry.ModuleConfig) error {
	// The url module already self-registers in its init() function
	// using require.RegisterCoreModule, so we don't need to register it again
	
	// Make URL and URLSearchParams globally available in the runtime
	urlModule := require.Require(config.VM, url.ModuleName)
	if urlModule != nil {
		urlObj := urlModule.ToObject(config.VM)
		
		// Make URL constructor globally available
		if urlConstructor := urlObj.Get("URL"); urlConstructor != nil {
			config.VM.Set("URL", urlConstructor)
		}
		
		// Make URLSearchParams constructor globally available
		if urlSearchParamsConstructor := urlObj.Get("URLSearchParams"); urlSearchParamsConstructor != nil {
			config.VM.Set("URLSearchParams", urlSearchParamsConstructor)
		}
	}
	
	return nil
}