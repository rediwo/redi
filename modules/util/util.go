package util

import (
	"github.com/dop251/goja_nodejs/require"
	"github.com/dop251/goja_nodejs/util"
	"github.com/rediwo/redi/registry"
)

// init registers the util module automatically
func init() {
	registry.RegisterModule("util", initUtilModule)
}

// initUtilModule initializes the util module
func initUtilModule(config registry.ModuleConfig) error {
	// Register the util module with goja_nodejs require system
	require.RegisterCoreModule(util.ModuleName, util.Require)
	
	return nil
}