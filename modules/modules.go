// Package modules provides automatic import of all JavaScript modules.
// Import this package with a blank identifier to automatically register all modules:
//
//	import _ "github.com/rediwo/redi/modules"
//
// This will register all available modules including:
// buffer, child_process, console, crypto, fetch, fs, path, process, stream, url, util
package modules

import (
	// Import all modules to trigger their init() functions
	_ "github.com/rediwo/redi/modules/buffer"
	_ "github.com/rediwo/redi/modules/child_process"
	_ "github.com/rediwo/redi/modules/console"
	_ "github.com/rediwo/redi/modules/crypto"
	_ "github.com/rediwo/redi/modules/fetch"
	_ "github.com/rediwo/redi/modules/fs"
	_ "github.com/rediwo/redi/modules/path"
	_ "github.com/rediwo/redi/modules/process"
	_ "github.com/rediwo/redi/modules/stream"
	_ "github.com/rediwo/redi/modules/url"
	_ "github.com/rediwo/redi/modules/util"
)

// This file serves as a convenience import to automatically register all modules.
// Each module's init() function will call registry.RegisterModule to make itself available.