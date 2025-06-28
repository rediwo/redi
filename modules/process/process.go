package process

import (
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	js "github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/eventloop"
	"github.com/rediwo/redi/modules"
)

const ModuleName = "process"

// ProcessModule represents the process module
type ProcessModule struct {
	runtime   *js.Runtime
	startTime time.Time
	loop      *eventloop.EventLoop
	version   string
}

// NewProcessModule creates a new process module instance
func NewProcessModule(vm *js.Runtime, loop *eventloop.EventLoop, version string) *ProcessModule {
	return &ProcessModule{
		runtime:   vm,
		startTime: time.Now(),
		loop:      loop,
		version:   version,
	}
}

// init registers the process module automatically
func init() {
	modules.RegisterModule("process", initProcessModule)
}

// initProcessModule initializes the process module
func initProcessModule(config modules.ModuleConfig) error {
	config.Registry.RegisterNativeModule(ModuleName, func(vm *js.Runtime, module *js.Object) {
		exports := module.Get("exports").(*js.Object)
		pm := NewProcessModule(vm, config.EventLoop, config.Version)
		pm.registerGlobals(vm, exports)
	})
	return nil
}


// registerGlobals registers all process globals and functions
func (pm *ProcessModule) registerGlobals(vm *js.Runtime, exports *js.Object) {
	// process.version - use external version
	exports.Set("version", pm.version)

	// process.versions - include Go runtime and module versions
	versions := vm.NewObject()
	versions.Set("go", runtime.Version())

	// Get module versions from build info
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range buildInfo.Deps {
			// Clean module path to use as key (remove version info and path)
			moduleName := dep.Path
			if strings.Contains(moduleName, "/") {
				parts := strings.Split(moduleName, "/")
				moduleName = parts[len(parts)-1] // Use last part as name
			}
			versions.Set(moduleName, dep.Version)
		}
	}

	exports.Set("versions", versions)

	// process.platform
	exports.Set("platform", runtime.GOOS)

	// process.arch
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x64"
	}
	exports.Set("arch", arch)

	// process.pid
	exports.Set("pid", os.Getpid())

	// process.ppid
	exports.Set("ppid", os.Getppid())

	// process.cwd()
	exports.Set("cwd", func(call js.FunctionCall) js.Value {
		cwd, err := os.Getwd()
		if err != nil {
			panic(vm.NewGoError(err))
		}
		return vm.ToValue(cwd)
	})

	// process.chdir()
	exports.Set("chdir", func(call js.FunctionCall) js.Value {
		if len(call.Arguments) == 0 {
			panic(vm.NewTypeError("directory is required"))
		}

		dir := call.Arguments[0].String()
		if err := os.Chdir(dir); err != nil {
			panic(vm.NewGoError(err))
		}

		return js.Undefined()
	})

	// process.env
	env := vm.NewObject()
	for _, envVar := range os.Environ() {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) == 2 {
			env.Set(parts[0], parts[1])
		}
	}
	exports.Set("env", env)

	// process.argv
	argv := vm.NewArray()
	// Use actual executable name
	execPath, _ := os.Executable()
	if execPath == "" {
		execPath = os.Args[0] // Fallback to first argument
	}
	argv.Set("0", vm.ToValue(execPath))
	// Add empty script name (not applicable for server-side execution)
	argv.Set("1", vm.ToValue(""))
	// Add actual command line arguments if available
	args := os.Args
	if len(args) > 1 {
		for i, arg := range args[1:] {
			argv.Set(string(rune(i+2)), vm.ToValue(arg))
		}
	}
	exports.Set("argv", argv)

	// process.argv0
	execName := filepath.Base(execPath)
	if execName == "" {
		execName = filepath.Base(os.Args[0]) // Fallback
	}
	exports.Set("argv0", execName)

	// process.execPath (already set above)
	exports.Set("execPath", execPath)

	// process.execArgv
	exports.Set("execArgv", vm.NewArray())

	// process.title
	exports.Set("title", "redi")

	// process.uptime()
	exports.Set("uptime", func(call js.FunctionCall) js.Value {
		uptime := time.Since(pm.startTime).Seconds()
		return vm.ToValue(uptime)
	})

	// process.hrtime()
	exports.Set("hrtime", func(call js.FunctionCall) js.Value {
		now := time.Now()
		var start time.Time

		if len(call.Arguments) > 0 && !js.IsUndefined(call.Arguments[0]) {
			// hrtime([time]) - get difference from previous hrtime
			if prevTime, ok := call.Arguments[0].Export().([]any); ok && len(prevTime) == 2 {
				if sec, ok := prevTime[0].(int64); ok {
					if nsec, ok := prevTime[1].(int64); ok {
						start = time.Unix(sec, nsec)
					}
				}
			}
		}

		var elapsed time.Duration
		if !start.IsZero() {
			elapsed = now.Sub(start)
		} else {
			elapsed = now.Sub(time.Unix(0, 0))
		}

		seconds := elapsed.Nanoseconds() / 1e9
		nanoseconds := elapsed.Nanoseconds() % 1e9

		result := vm.NewArray()
		result.Set("0", vm.ToValue(seconds))
		result.Set("1", vm.ToValue(nanoseconds))
		return result
	})

	// process.hrtime.bigint()
	hrtimeFunc := exports.Get("hrtime").(*js.Object)
	hrtimeFunc.Set("bigint", func(call js.FunctionCall) js.Value {
		now := time.Now().UnixNano()
		return vm.ToValue(now)
	})

	// process.memoryUsage()
	exports.Set("memoryUsage", func(call js.FunctionCall) js.Value {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		result := vm.NewObject()
		result.Set("rss", m.Sys)
		result.Set("heapTotal", m.HeapSys)
		result.Set("heapUsed", m.HeapInuse)
		result.Set("external", 0)     // Not available in Go
		result.Set("arrayBuffers", 0) // Not available in Go
		return result
	})

	// process.cpuUsage()
	exports.Set("cpuUsage", func(call js.FunctionCall) js.Value {
		// Mock CPU usage - Go doesn't provide easy access to process CPU time
		result := vm.NewObject()
		result.Set("user", 1000000)  // Microseconds
		result.Set("system", 500000) // Microseconds
		return result
	})

	// process.nextTick()
	exports.Set("nextTick", func(call js.FunctionCall) js.Value {
		if len(call.Arguments) == 0 {
			panic(vm.NewTypeError("callback is required"))
		}

		callback := call.Arguments[0]
		args := call.Arguments[1:]

		// If event loop is available, use it for proper nextTick behavior
		if pm.loop != nil {
			pm.loop.RunOnLoop(func(vm *js.Runtime) {
				if callFunc, ok := js.AssertFunction(callback); ok {
					// Convert args to js.Value slice
					jsArgs := make([]js.Value, len(args))
					copy(jsArgs, args)
					callFunc(js.Undefined(), jsArgs...)
				}
			})
		} else {
			// Fallback to goroutine if no event loop (for tests)
			go func() {
				if callFunc, ok := js.AssertFunction(callback); ok {
					// Convert args to js.Value slice
					jsArgs := make([]js.Value, len(args))
					copy(jsArgs, args)
					callFunc(js.Undefined(), jsArgs...)
				}
			}()
		}

		return js.Undefined()
	})

	// process.exit()
	exports.Set("exit", func(call js.FunctionCall) js.Value {
		code := 0
		if len(call.Arguments) > 0 && !js.IsUndefined(call.Arguments[0]) {
			code = int(call.Arguments[0].ToInteger())
		}
		os.Exit(code)
		return js.Undefined()
	})

	// process.abort()
	exports.Set("abort", func(call js.FunctionCall) js.Value {
		os.Exit(134) // SIGABRT exit code
		return js.Undefined()
	})

	// process.kill()
	exports.Set("kill", func(call js.FunctionCall) js.Value {
		if len(call.Arguments) < 1 {
			panic(vm.NewTypeError("pid is required"))
		}

		pid := int(call.Arguments[0].ToInteger())
		// signal := "SIGTERM"
		// if len(call.Arguments) > 1 {
		// 	signal = call.Arguments[1].String()
		// }

		// Find process
		process, err := os.FindProcess(pid)
		if err != nil {
			panic(vm.NewGoError(err))
		}

		// Send signal (simplified - only supports killing on Unix-like systems)
		if err := process.Kill(); err != nil {
			panic(vm.NewGoError(err))
		}

		return vm.ToValue(true)
	})

	// process.getuid() (Unix only)
	exports.Set("getuid", func(call js.FunctionCall) js.Value {
		if runtime.GOOS == "windows" {
			panic(vm.NewTypeError("getuid is not supported on Windows"))
		}
		return vm.ToValue(os.Getuid())
	})

	// process.getgid() (Unix only)
	exports.Set("getgid", func(call js.FunctionCall) js.Value {
		if runtime.GOOS == "windows" {
			panic(vm.NewTypeError("getgid is not supported on Windows"))
		}
		return vm.ToValue(os.Getgid())
	})

	// process.geteuid() (Unix only)
	exports.Set("geteuid", func(call js.FunctionCall) js.Value {
		if runtime.GOOS == "windows" {
			panic(vm.NewTypeError("geteuid is not supported on Windows"))
		}
		return vm.ToValue(os.Geteuid())
	})

	// process.getegid() (Unix only)
	exports.Set("getegid", func(call js.FunctionCall) js.Value {
		if runtime.GOOS == "windows" {
			panic(vm.NewTypeError("getegid is not supported on Windows"))
		}
		return vm.ToValue(os.Getegid())
	})

	// process.stdout, process.stderr, process.stdin (mock objects)
	stdout := vm.NewObject()
	stdout.Set("isTTY", true)
	stdout.Set("write", func(call js.FunctionCall) js.Value {
		if len(call.Arguments) > 0 {
			os.Stdout.WriteString(call.Arguments[0].String())
		}
		return js.Undefined()
	})
	exports.Set("stdout", stdout)

	stderr := vm.NewObject()
	stderr.Set("isTTY", true)
	stderr.Set("write", func(call js.FunctionCall) js.Value {
		if len(call.Arguments) > 0 {
			os.Stderr.WriteString(call.Arguments[0].String())
		}
		return js.Undefined()
	})
	exports.Set("stderr", stderr)

	stdin := vm.NewObject()
	stdin.Set("isTTY", true)
	exports.Set("stdin", stdin)
}
