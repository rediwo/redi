package buffer

import (
	"github.com/dop251/goja_nodejs/buffer"
	"github.com/rediwo/redi/registry"
)

// init registers the buffer module automatically
func init() {
	registry.RegisterModule("buffer", initBufferModule)
}

// initBufferModule initializes the buffer module
func initBufferModule(config registry.ModuleConfig) error {
	// Enable Buffer globally in the runtime
	buffer.Enable(config.VM)
	
	return nil
}