package path

import (
	"path/filepath"
	"strings"
	
	js "github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
)

const ModuleName = "path"

// Enable registers the path module in the given registry
func Enable(registry *require.Registry) {
	registry.RegisterNativeModule(ModuleName, func(runtime *js.Runtime, module *js.Object) {
		exports := module.Get("exports").(*js.Object)
		
		// sep - path separator
		exports.Set("sep", string(filepath.Separator))
		
		// delimiter - path delimiter
		exports.Set("delimiter", string(filepath.ListSeparator))
		
		// join - join path segments
		exports.Set("join", func(call js.FunctionCall) js.Value {
			segments := make([]string, len(call.Arguments))
			for i, arg := range call.Arguments {
				segments[i] = arg.String()
			}
			return runtime.ToValue(filepath.Join(segments...))
		})
		
		// resolve - resolve to absolute path
		exports.Set("resolve", func(call js.FunctionCall) js.Value {
			if len(call.Arguments) == 0 {
				pwd, _ := filepath.Abs(".")
				return runtime.ToValue(pwd)
			}
			
			segments := make([]string, len(call.Arguments))
			for i, arg := range call.Arguments {
				segments[i] = arg.String()
			}
			
			path := filepath.Join(segments...)
			absPath, err := filepath.Abs(path)
			if err != nil {
				panic(runtime.NewGoError(err))
			}
			
			return runtime.ToValue(absPath)
		})
		
		// dirname - get directory name
		exports.Set("dirname", func(call js.FunctionCall) js.Value {
			if len(call.Arguments) == 0 {
				panic(runtime.NewTypeError("path is required"))
			}
			return runtime.ToValue(filepath.Dir(call.Arguments[0].String()))
		})
		
		// basename - get base name
		exports.Set("basename", func(call js.FunctionCall) js.Value {
			if len(call.Arguments) == 0 {
				panic(runtime.NewTypeError("path is required"))
			}
			
			path := call.Arguments[0].String()
			base := filepath.Base(path)
			
			// If ext is provided, remove it from the result
			if len(call.Arguments) > 1 {
				ext := call.Arguments[1].String()
				if strings.HasSuffix(base, ext) {
					base = base[:len(base)-len(ext)]
				}
			}
			
			return runtime.ToValue(base)
		})
		
		// extname - get extension
		exports.Set("extname", func(call js.FunctionCall) js.Value {
			if len(call.Arguments) == 0 {
				panic(runtime.NewTypeError("path is required"))
			}
			path := call.Arguments[0].String()
			
			// Special case: files starting with . but no extension should return empty string
			base := filepath.Base(path)
			if strings.HasPrefix(base, ".") && !strings.Contains(base[1:], ".") {
				return runtime.ToValue("")
			}
			
			return runtime.ToValue(filepath.Ext(path))
		})
		
		// isAbsolute - check if path is absolute
		exports.Set("isAbsolute", func(call js.FunctionCall) js.Value {
			if len(call.Arguments) == 0 {
				panic(runtime.NewTypeError("path is required"))
			}
			return runtime.ToValue(filepath.IsAbs(call.Arguments[0].String()))
		})
		
		// relative - get relative path
		exports.Set("relative", func(call js.FunctionCall) js.Value {
			if len(call.Arguments) < 2 {
				panic(runtime.NewTypeError("from and to paths are required"))
			}
			
			from := call.Arguments[0].String()
			to := call.Arguments[1].String()
			
			rel, err := filepath.Rel(from, to)
			if err != nil {
				panic(runtime.NewGoError(err))
			}
			
			return runtime.ToValue(rel)
		})
		
		// normalize - normalize path
		exports.Set("normalize", func(call js.FunctionCall) js.Value {
			if len(call.Arguments) == 0 {
				panic(runtime.NewTypeError("path is required"))
			}
			return runtime.ToValue(filepath.Clean(call.Arguments[0].String()))
		})
		
		// parse - parse path into components
		exports.Set("parse", func(call js.FunctionCall) js.Value {
			if len(call.Arguments) == 0 {
				panic(runtime.NewTypeError("path is required"))
			}
			
			path := call.Arguments[0].String()
			dir := filepath.Dir(path)
			base := filepath.Base(path)
			ext := filepath.Ext(path)
			name := base
			if ext != "" {
				name = base[:len(base)-len(ext)]
			}
			
			obj := runtime.NewObject()
			obj.Set("root", "") // TODO: properly detect root on different platforms
			obj.Set("dir", dir)
			obj.Set("base", base)
			obj.Set("ext", ext)
			obj.Set("name", name)
			
			return obj
		})
		
		// format - format path from components
		exports.Set("format", func(call js.FunctionCall) js.Value {
			if len(call.Arguments) == 0 {
				panic(runtime.NewTypeError("pathObject is required"))
			}
			
			pathObj := call.Arguments[0].Export()
			if pathMap, ok := pathObj.(map[string]interface{}); ok {
				dir := ""
				base := ""
				
				if d, ok := pathMap["dir"].(string); ok {
					dir = d
				}
				
				if b, ok := pathMap["base"].(string); ok {
					base = b
				} else {
					// If base is not provided, construct from name and ext
					name := ""
					ext := ""
					if n, ok := pathMap["name"].(string); ok {
						name = n
					}
					if e, ok := pathMap["ext"].(string); ok {
						ext = e
					}
					base = name + ext
				}
				
				if dir == "" {
					return runtime.ToValue(base)
				}
				if base == "" {
					return runtime.ToValue(dir)
				}
				
				return runtime.ToValue(filepath.Join(dir, base))
			}
			
			panic(runtime.NewTypeError("pathObject must be an object"))
		})
	})
}