package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	
	js "github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/eventloop"
	"github.com/rediwo/redi/filesystem"
	"github.com/rediwo/redi/modules"
)

const ModuleName = "fs"

// FSModule represents the file system module with event loop support
type FSModule struct {
	basePath string
	loop     *eventloop.EventLoop
	fs       filesystem.FileSystem
}

// NewFSModule creates a new fs module instance with event loop support
func NewFSModule(basePath string, loop *eventloop.EventLoop) *FSModule {
	// Create an OS filesystem for the base path
	osFS := filesystem.NewOSFileSystem("")
	return &FSModule{
		basePath: basePath,
		loop:     loop,
		fs:       osFS,
	}
}

// NewFSModuleWithFS creates a new fs module instance with custom filesystem support
func NewFSModuleWithFS(fs filesystem.FileSystem, basePath string, loop *eventloop.EventLoop) *FSModule {
	return &FSModule{
		basePath: basePath,
		loop:     loop,
		fs:       fs,
	}
}

// init registers the fs module automatically
func init() {
	modules.RegisterModule("fs", initFSModule)
}

// initFSModule initializes the fs module
func initFSModule(config modules.ModuleConfig) error {
	var fsModule *FSModule
	if config.FileSystem != nil {
		fsModule = NewFSModuleWithFS(config.FileSystem, config.BasePath, config.EventLoop)
	} else {
		fsModule = NewFSModule(config.BasePath, config.EventLoop)
	}
	
	config.Registry.RegisterNativeModule(ModuleName, func(runtime *js.Runtime, module *js.Object) {
		exports := module.Get("exports").(*js.Object)
		fsModule.registerFunctions(runtime, exports)
	})
	return nil
}


// registerFunctions registers all fs functions on the exports object
func (fsm *FSModule) registerFunctions(runtime *js.Runtime, exports *js.Object) {
	// === SYNCHRONOUS FUNCTIONS ===
	
	// readFileSync
	exports.Set("readFileSync", func(call js.FunctionCall) js.Value {
		if len(call.Arguments) == 0 {
			panic(runtime.NewTypeError("path is required"))
		}
		
		filename := call.Arguments[0].String()
		// Resolve relative paths from basePath
		if !filepath.IsAbs(filename) {
			filename = filepath.Join(fsm.basePath, filename)
		}
		
		encoding := "utf8"
		if len(call.Arguments) > 1 {
			encoding = call.Arguments[1].String()
		}
		
		data, err := fsm.fs.ReadFile(filename)
		if err != nil {
			panic(runtime.NewGoError(err))
		}
		
		if encoding == "utf8" || encoding == "utf-8" {
			return runtime.ToValue(string(data))
		}
		
		// Return buffer for binary data
		return runtime.ToValue(data)
	})
	
	// writeFileSync
	exports.Set("writeFileSync", func(call js.FunctionCall) js.Value {
		if len(call.Arguments) < 2 {
			panic(runtime.NewTypeError("path and data are required"))
		}
		
		if fsm.fs.IsReadOnly() {
			panic(runtime.NewGoError(fmt.Errorf("filesystem is read-only")))
		}
		
		filename := call.Arguments[0].String()
		// Resolve relative paths from basePath
		if !filepath.IsAbs(filename) {
			filename = filepath.Join(fsm.basePath, filename)
		}
		
		data := call.Arguments[1].String()
		
		err := os.WriteFile(filename, []byte(data), 0644)
		if err != nil {
			panic(runtime.NewGoError(err))
		}
		
		return js.Undefined()
	})
	
	// existsSync
	exports.Set("existsSync", func(call js.FunctionCall) js.Value {
		if len(call.Arguments) == 0 {
			panic(runtime.NewTypeError("path is required"))
		}
		
		filename := call.Arguments[0].String()
		// Resolve relative paths from basePath
		if !filepath.IsAbs(filename) {
			filename = filepath.Join(fsm.basePath, filename)
		}
		
		_, err := fsm.fs.Stat(filename)
		return runtime.ToValue(err == nil)
	})
	
	// mkdirSync
	exports.Set("mkdirSync", func(call js.FunctionCall) js.Value {
		if len(call.Arguments) == 0 {
			panic(runtime.NewTypeError("path is required"))
		}
		
		if fsm.fs.IsReadOnly() {
			panic(runtime.NewGoError(fmt.Errorf("filesystem is read-only")))
		}
		
		dirname := call.Arguments[0].String()
		// Resolve relative paths from basePath
		if !filepath.IsAbs(dirname) {
			dirname = filepath.Join(fsm.basePath, dirname)
		}
		
		recursive := false
		if len(call.Arguments) > 1 {
			if opts := call.Arguments[1].Export(); opts != nil {
				if optsMap, ok := opts.(map[string]interface{}); ok {
					if r, ok := optsMap["recursive"].(bool); ok {
						recursive = r
					}
				}
			}
		}
		
		var err error
		if recursive {
			err = os.MkdirAll(dirname, 0755)
		} else {
			err = os.Mkdir(dirname, 0755)
		}
		
		if err != nil {
			panic(runtime.NewGoError(err))
		}
		
		return js.Undefined()
	})
	
	// readdirSync
	exports.Set("readdirSync", func(call js.FunctionCall) js.Value {
		if len(call.Arguments) == 0 {
			panic(runtime.NewTypeError("path is required"))
		}
		
		dirname := call.Arguments[0].String()
		// Resolve relative paths from basePath
		if !filepath.IsAbs(dirname) {
			dirname = filepath.Join(fsm.basePath, dirname)
		}
		
		entries, err := os.ReadDir(dirname)
		if err != nil {
			panic(runtime.NewGoError(err))
		}
		
		names := make([]string, len(entries))
		for i, entry := range entries {
			names[i] = entry.Name()
		}
		
		return runtime.ToValue(names)
	})
	
	// statSync
	exports.Set("statSync", func(call js.FunctionCall) js.Value {
		if len(call.Arguments) == 0 {
			panic(runtime.NewTypeError("path is required"))
		}
		
		filename := call.Arguments[0].String()
		// Resolve relative paths from basePath
		if !filepath.IsAbs(filename) {
			filename = filepath.Join(fsm.basePath, filename)
		}
		
		stat, err := os.Stat(filename)
		if err != nil {
			panic(runtime.NewGoError(err))
		}
		
		obj := runtime.NewObject()
		obj.Set("size", stat.Size())
		obj.Set("mode", int(stat.Mode()))
		obj.Set("isDirectory", func() bool { return stat.IsDir() })
		obj.Set("isFile", func() bool { return stat.Mode().IsRegular() })
		
		return obj
	})
	
	// unlinkSync (delete file)
	exports.Set("unlinkSync", func(call js.FunctionCall) js.Value {
		if len(call.Arguments) == 0 {
			panic(runtime.NewTypeError("path is required"))
		}
		
		if fsm.fs.IsReadOnly() {
			panic(runtime.NewGoError(fmt.Errorf("filesystem is read-only")))
		}
		
		filename := call.Arguments[0].String()
		// Resolve relative paths from basePath
		if !filepath.IsAbs(filename) {
			filename = filepath.Join(fsm.basePath, filename)
		}
		
		err := os.Remove(filename)
		if err != nil {
			panic(runtime.NewGoError(err))
		}
		
		return js.Undefined()
	})
	
	// rmdirSync (remove directory)
	exports.Set("rmdirSync", func(call js.FunctionCall) js.Value {
		if len(call.Arguments) < 1 {
			panic(runtime.NewTypeError("directory path is required"))
		}
		
		if fsm.fs.IsReadOnly() {
			panic(runtime.NewGoError(fmt.Errorf("filesystem is read-only")))
		}
		
		dirname := call.Arguments[0].String()
		// Resolve relative paths from basePath
		if !filepath.IsAbs(dirname) {
			dirname = filepath.Join(fsm.basePath, dirname)
		}
		
		err := os.Remove(dirname)
		if err != nil {
			panic(runtime.NewGoError(err))
		}
		
		return js.Undefined()
	})
	
	// copyFileSync
	exports.Set("copyFileSync", func(call js.FunctionCall) js.Value {
		if len(call.Arguments) < 2 {
			panic(runtime.NewTypeError("source and destination are required"))
		}
		
		if fsm.fs.IsReadOnly() {
			panic(runtime.NewGoError(fmt.Errorf("filesystem is read-only")))
		}
		
		src := call.Arguments[0].String()
		dst := call.Arguments[1].String()
		
		// Resolve relative paths from basePath
		if !filepath.IsAbs(src) {
			src = filepath.Join(fsm.basePath, src)
		}
		if !filepath.IsAbs(dst) {
			dst = filepath.Join(fsm.basePath, dst)
		}
		
		sourceFile, err := os.Open(src)
		if err != nil {
			panic(runtime.NewGoError(err))
		}
		defer sourceFile.Close()
		
		destFile, err := os.Create(dst)
		if err != nil {
			panic(runtime.NewGoError(err))
		}
		defer destFile.Close()
		
		_, err = io.Copy(destFile, sourceFile)
		if err != nil {
			panic(runtime.NewGoError(err))
		}
		
		return js.Undefined()
	})

	// === ASYNCHRONOUS FUNCTIONS ===
	// These functions use the event loop for truly asynchronous execution
	
	// readFile (async) - callback style
	exports.Set("readFile", func(call js.FunctionCall) js.Value {
		if len(call.Arguments) < 2 {
			panic(runtime.NewTypeError("path and callback are required"))
		}
		
		filename := call.Arguments[0].String()
		callback := call.Arguments[len(call.Arguments)-1] // callback is always last
		
		// Handle optional encoding parameter
		encoding := "utf8"
		if len(call.Arguments) > 2 {
			if !js.IsNull(call.Arguments[1]) && !js.IsUndefined(call.Arguments[1]) {
				if str, ok := call.Arguments[1].Export().(string); ok {
					encoding = str
				} else if opts, ok := call.Arguments[1].Export().(map[string]interface{}); ok {
					if enc, ok := opts["encoding"].(string); ok {
						encoding = enc
					}
				}
			}
		}
		
		// Resolve relative paths from basePath
		if !filepath.IsAbs(filename) {
			filename = filepath.Join(fsm.basePath, filename)
		}
		
		// Execute asynchronously using the event loop
		if fsm.loop == nil {
			panic(runtime.NewTypeError("fs async functions require event loop"))
		}
		
		// Schedule async operation
		go func() {
			data, err := os.ReadFile(filename)
			
			// Schedule callback on the event loop
			fsm.loop.RunOnLoop(func(vm *js.Runtime) {
				if callFunc, ok := js.AssertFunction(callback); ok {
					if err != nil {
						callFunc(js.Undefined(), vm.ToValue(err.Error()), js.Null())
					} else {
						var result js.Value
						if encoding == "utf8" || encoding == "utf-8" {
							result = vm.ToValue(string(data))
						} else {
							result = vm.ToValue(data)
						}
						callFunc(js.Undefined(), js.Null(), result)
					}
				}
			})
		}()
		
		return js.Undefined()
	})
	
	// writeFile (async) - callback style
	exports.Set("writeFile", func(call js.FunctionCall) js.Value {
		if len(call.Arguments) < 3 {
			panic(runtime.NewTypeError("path, data, and callback are required"))
		}
		
		if fsm.fs.IsReadOnly() {
			panic(runtime.NewGoError(fmt.Errorf("filesystem is read-only")))
		}
		
		filename := call.Arguments[0].String()
		data := call.Arguments[1].String()
		callback := call.Arguments[2]
		
		// Resolve relative paths from basePath
		if !filepath.IsAbs(filename) {
			filename = filepath.Join(fsm.basePath, filename)
		}
		
		// Execute asynchronously using the event loop
		if fsm.loop == nil {
			panic(runtime.NewTypeError("fs async functions require event loop"))
		}
		
		go func() {
			err := os.WriteFile(filename, []byte(data), 0644)
			
			// Schedule callback on the event loop
			fsm.loop.RunOnLoop(func(vm *js.Runtime) {
				if callFunc, ok := js.AssertFunction(callback); ok {
					if err != nil {
						callFunc(js.Undefined(), vm.ToValue(err.Error()))
					} else {
						callFunc(js.Undefined(), js.Null())
					}
				}
			})
		}()
		
		return js.Undefined()
	})
	
	// mkdir (async) - callback style
	exports.Set("mkdir", func(call js.FunctionCall) js.Value {
		if len(call.Arguments) < 2 {
			panic(runtime.NewTypeError("path and callback are required"))
		}
		
		if fsm.fs.IsReadOnly() {
			panic(runtime.NewGoError(fmt.Errorf("filesystem is read-only")))
		}
		
		dirname := call.Arguments[0].String()
		callback := call.Arguments[len(call.Arguments)-1] // callback is always last
		
		// Handle optional options parameter
		recursive := false
		if len(call.Arguments) > 2 {
			if opts := call.Arguments[1].Export(); opts != nil {
				if optsMap, ok := opts.(map[string]interface{}); ok {
					if r, ok := optsMap["recursive"].(bool); ok {
						recursive = r
					}
				}
			}
		}
		
		// Resolve relative paths from basePath
		if !filepath.IsAbs(dirname) {
			dirname = filepath.Join(fsm.basePath, dirname)
		}
		
		// Execute asynchronously using the event loop
		if fsm.loop == nil {
			panic(runtime.NewTypeError("fs async functions require event loop"))
		}
		
		go func() {
			var err error
			if recursive {
				err = os.MkdirAll(dirname, 0755)
			} else {
				err = os.Mkdir(dirname, 0755)
			}
			
			// Schedule callback on the event loop
			fsm.loop.RunOnLoop(func(vm *js.Runtime) {
				if callFunc, ok := js.AssertFunction(callback); ok {
					if err != nil {
						callFunc(js.Undefined(), vm.ToValue(err.Error()))
					} else {
						callFunc(js.Undefined(), js.Null())
					}
				}
			})
		}()
		
		return js.Undefined()
	})
	
	// readdir (async) - callback style
	exports.Set("readdir", func(call js.FunctionCall) js.Value {
		if len(call.Arguments) < 2 {
			panic(runtime.NewTypeError("path and callback are required"))
		}
		
		dirname := call.Arguments[0].String()
		callback := call.Arguments[1]
		
		// Resolve relative paths from basePath
		if !filepath.IsAbs(dirname) {
			dirname = filepath.Join(fsm.basePath, dirname)
		}
		
		// Execute asynchronously using the event loop
		if fsm.loop == nil {
			panic(runtime.NewTypeError("fs async functions require event loop"))
		}
		
		go func() {
			entries, err := os.ReadDir(dirname)
			
			// Schedule callback on the event loop
			fsm.loop.RunOnLoop(func(vm *js.Runtime) {
				if callFunc, ok := js.AssertFunction(callback); ok {
					if err != nil {
						callFunc(js.Undefined(), vm.ToValue(err.Error()), js.Null())
					} else {
						names := make([]string, len(entries))
						for i, entry := range entries {
							names[i] = entry.Name()
						}
						result := vm.ToValue(names)
						callFunc(js.Undefined(), js.Null(), result)
					}
				}
			})
		}()
		
		return js.Undefined()
	})
	
	// stat (async) - callback style
	exports.Set("stat", func(call js.FunctionCall) js.Value {
		if len(call.Arguments) < 2 {
			panic(runtime.NewTypeError("path and callback are required"))
		}
		
		filename := call.Arguments[0].String()
		callback := call.Arguments[1]
		
		// Resolve relative paths from basePath
		if !filepath.IsAbs(filename) {
			filename = filepath.Join(fsm.basePath, filename)
		}
		
		// Execute asynchronously using the event loop
		if fsm.loop == nil {
			panic(runtime.NewTypeError("fs async functions require event loop"))
		}
		
		go func() {
			stat, err := os.Stat(filename)
			
			// Schedule callback on the event loop
			fsm.loop.RunOnLoop(func(vm *js.Runtime) {
				if callFunc, ok := js.AssertFunction(callback); ok {
					if err != nil {
						callFunc(js.Undefined(), vm.ToValue(err.Error()), js.Null())
					} else {
						obj := vm.NewObject()
						obj.Set("size", stat.Size())
						obj.Set("mode", int(stat.Mode()))
						obj.Set("isDirectory", func() bool { return stat.IsDir() })
						obj.Set("isFile", func() bool { return stat.Mode().IsRegular() })
						callFunc(js.Undefined(), js.Null(), obj)
					}
				}
			})
		}()
		
		return js.Undefined()
	})
	
	// unlink (async) - callback style - delete file
	exports.Set("unlink", func(call js.FunctionCall) js.Value {
		if len(call.Arguments) < 2 {
			panic(runtime.NewTypeError("path and callback are required"))
		}
		
		if fsm.fs.IsReadOnly() {
			panic(runtime.NewGoError(fmt.Errorf("filesystem is read-only")))
		}
		
		filename := call.Arguments[0].String()
		callback := call.Arguments[1]
		
		// Resolve relative paths from basePath
		if !filepath.IsAbs(filename) {
			filename = filepath.Join(fsm.basePath, filename)
		}
		
		// Execute asynchronously using the event loop
		if fsm.loop == nil {
			panic(runtime.NewTypeError("fs async functions require event loop"))
		}
		
		go func() {
			err := os.Remove(filename)
			
			// Schedule callback on the event loop
			fsm.loop.RunOnLoop(func(vm *js.Runtime) {
				if callFunc, ok := js.AssertFunction(callback); ok {
					if err != nil {
						callFunc(js.Undefined(), vm.ToValue(err.Error()))
					} else {
						callFunc(js.Undefined(), js.Null())
					}
				}
			})
		}()
		
		return js.Undefined()
	})
	
	// rmdir (async) - callback style
	exports.Set("rmdir", func(call js.FunctionCall) js.Value {
		if len(call.Arguments) < 2 {
			panic(runtime.NewTypeError("directory path and callback are required"))
		}
		
		if fsm.fs.IsReadOnly() {
			panic(runtime.NewGoError(fmt.Errorf("filesystem is read-only")))
		}
		
		dirname := call.Arguments[0].String()
		callback := call.Arguments[1]
		
		// Resolve relative paths from basePath
		if !filepath.IsAbs(dirname) {
			dirname = filepath.Join(fsm.basePath, dirname)
		}
		
		// Execute asynchronously using the event loop
		if fsm.loop == nil {
			panic(runtime.NewTypeError("fs async functions require event loop"))
		}
		
		go func() {
			err := os.Remove(dirname)
			
			// Schedule callback on the event loop
			fsm.loop.RunOnLoop(func(vm *js.Runtime) {
				if callFunc, ok := js.AssertFunction(callback); ok {
					if err != nil {
						callFunc(js.Undefined(), vm.ToValue(err.Error()))
					} else {
						callFunc(js.Undefined(), js.Null())
					}
				}
			})
		}()
		
		return js.Undefined()
	})
	
	// copyFile (async) - callback style
	exports.Set("copyFile", func(call js.FunctionCall) js.Value {
		if len(call.Arguments) < 3 {
			panic(runtime.NewTypeError("source, destination, and callback are required"))
		}
		
		if fsm.fs.IsReadOnly() {
			panic(runtime.NewGoError(fmt.Errorf("filesystem is read-only")))
		}
		
		src := call.Arguments[0].String()
		dst := call.Arguments[1].String()
		callback := call.Arguments[2]
		
		// Resolve relative paths from basePath
		if !filepath.IsAbs(src) {
			src = filepath.Join(fsm.basePath, src)
		}
		if !filepath.IsAbs(dst) {
			dst = filepath.Join(fsm.basePath, dst)
		}
		
		// Execute asynchronously using the event loop
		if fsm.loop == nil {
			panic(runtime.NewTypeError("fs async functions require event loop"))
		}
		
		go func() {
			sourceFile, err := os.Open(src)
			if err != nil {
				// Schedule callback on the event loop
				fsm.loop.RunOnLoop(func(vm *js.Runtime) {
					if callFunc, ok := js.AssertFunction(callback); ok {
						callFunc(js.Undefined(), vm.ToValue(err.Error()))
					}
				})
				return
			}
			defer sourceFile.Close()
			
			destFile, err := os.Create(dst)
			if err != nil {
				// Schedule callback on the event loop
				fsm.loop.RunOnLoop(func(vm *js.Runtime) {
					if callFunc, ok := js.AssertFunction(callback); ok {
						callFunc(js.Undefined(), vm.ToValue(err.Error()))
					}
				})
				return
			}
			defer destFile.Close()
			
			_, err = io.Copy(destFile, sourceFile)
			
			// Schedule callback on the event loop
			fsm.loop.RunOnLoop(func(vm *js.Runtime) {
				if callFunc, ok := js.AssertFunction(callback); ok {
					if err != nil {
						callFunc(js.Undefined(), vm.ToValue(err.Error()))
					} else {
						callFunc(js.Undefined(), js.Null())
					}
				}
			})
		}()
		
		return js.Undefined()
	})
}