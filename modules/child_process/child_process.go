package child_process

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	js "github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/eventloop"
	"github.com/rediwo/redi/modules"
)

const ModuleName = "child_process"

// ChildProcessModule represents the child_process module
type ChildProcessModule struct {
	runtime   *js.Runtime
	loop      *eventloop.EventLoop
}

// NewChildProcessModule creates a new child_process module instance
func NewChildProcessModule(vm *js.Runtime, loop *eventloop.EventLoop) *ChildProcessModule {
	return &ChildProcessModule{
		runtime: vm,
		loop:    loop,
	}
}

// init registers the child_process module automatically
func init() {
	modules.RegisterModule("child_process", initChildProcessModule)
}

// initChildProcessModule initializes the child_process module
func initChildProcessModule(config modules.ModuleConfig) error {
	config.Registry.RegisterNativeModule(ModuleName, func(vm *js.Runtime, module *js.Object) {
		exports := module.Get("exports").(*js.Object)
		cpm := NewChildProcessModule(vm, config.EventLoop)
		cpm.registerGlobals(vm, exports)
	})
	return nil
}

// registerGlobals registers all child_process functions
func (cpm *ChildProcessModule) registerGlobals(vm *js.Runtime, exports *js.Object) {
	// execSync - synchronous command execution
	exports.Set("execSync", func(call js.FunctionCall) js.Value {
		if len(call.Arguments) == 0 {
			panic(vm.NewTypeError("command is required"))
		}

		command := call.Arguments[0].String()
		options := make(map[string]interface{})

		// Parse options if provided
		if len(call.Arguments) > 1 && !js.IsUndefined(call.Arguments[1]) {
			if opt, ok := call.Arguments[1].Export().(map[string]interface{}); ok {
				options = opt
			}
		}

		return cpm.execSync(vm, command, options)
	})

	// exec - asynchronous command execution
	exports.Set("exec", func(call js.FunctionCall) js.Value {
		if len(call.Arguments) == 0 {
			panic(vm.NewTypeError("command is required"))
		}

		command := call.Arguments[0].String()
		var callback js.Value
		options := make(map[string]interface{})

		// Parse arguments - can be (command, callback) or (command, options, callback)
		if len(call.Arguments) >= 2 {
			if len(call.Arguments) == 2 {
				// (command, callback)
				callback = call.Arguments[1]
			} else {
				// (command, options, callback)
				if opt, ok := call.Arguments[1].Export().(map[string]interface{}); ok {
					options = opt
				}
				if len(call.Arguments) >= 3 {
					callback = call.Arguments[2]
				}
			}
		}

		return cpm.exec(vm, command, options, callback)
	})

	// spawn - spawn a new process (simplified version)
	exports.Set("spawn", func(call js.FunctionCall) js.Value {
		if len(call.Arguments) == 0 {
			panic(vm.NewTypeError("command is required"))
		}

		command := call.Arguments[0].String()
		var args []string
		options := make(map[string]interface{})

		// Parse arguments
		if len(call.Arguments) > 1 && !js.IsUndefined(call.Arguments[1]) {
			if argsArray := call.Arguments[1].Export(); argsArray != nil {
				if argsSlice, ok := argsArray.([]interface{}); ok {
					args = make([]string, len(argsSlice))
					for i, arg := range argsSlice {
						args[i] = fmt.Sprintf("%v", arg)
					}
				}
			}
		}

		if len(call.Arguments) > 2 && !js.IsUndefined(call.Arguments[2]) {
			if opt, ok := call.Arguments[2].Export().(map[string]interface{}); ok {
				options = opt
			}
		}

		return cpm.spawn(vm, command, args, options)
	})
}

// execSync executes a command synchronously and returns the output
func (cpm *ChildProcessModule) execSync(vm *js.Runtime, command string, options map[string]interface{}) js.Value {
	// Parse options
	encoding := "utf8"
	cwd := ""
	timeout := 0

	if enc, ok := options["encoding"].(string); ok {
		encoding = enc
	}
	if dir, ok := options["cwd"].(string); ok {
		cwd = dir
	}
	if t, ok := options["timeout"].(float64); ok {
		timeout = int(t)
	}

	// Create command
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	// Set working directory if specified
	if cwd != "" {
		cmd.Dir = cwd
	}

	// Set up context with timeout if specified
	var ctx context.Context
	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
		defer cancel()
		cmd = exec.CommandContext(ctx, cmd.Args[0], cmd.Args[1:]...)
	}

	// Execute command
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Create error object similar to Node.js
		errorObj := vm.NewObject()
		errorObj.Set("message", err.Error())
		errorObj.Set("code", getExitCode(err))
		errorObj.Set("killed", false)
		if ctx != nil && ctx.Err() == context.DeadlineExceeded {
			errorObj.Set("killed", true)
			errorObj.Set("signal", "SIGTERM")
		}
		if encoding != "buffer" {
			errorObj.Set("stdout", string(output))
			errorObj.Set("stderr", string(output))
		} else {
			errorObj.Set("stdout", output)
			errorObj.Set("stderr", output)
		}
		panic(vm.NewGoError(fmt.Errorf("%v", errorObj)))
	}

	// Return output based on encoding
	if encoding == "buffer" {
		// Return as buffer (array of bytes)
		buffer := vm.NewArray()
		for i, b := range output {
			buffer.Set(fmt.Sprintf("%d", i), vm.ToValue(int(b)))
		}
		return buffer
	}

	return vm.ToValue(string(output))
}

// exec executes a command asynchronously
func (cpm *ChildProcessModule) exec(vm *js.Runtime, command string, options map[string]interface{}, callback js.Value) js.Value {
	// Parse options
	encoding := "utf8"
	cwd := ""
	timeout := 0

	if enc, ok := options["encoding"].(string); ok {
		encoding = enc
	}
	if dir, ok := options["cwd"].(string); ok {
		cwd = dir
	}
	if t, ok := options["timeout"].(float64); ok {
		timeout = int(t)
	}

	// Execute asynchronously
	go func() {
		// Create command
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd", "/C", command)
		} else {
			cmd = exec.Command("sh", "-c", command)
		}

		// Set working directory if specified
		if cwd != "" {
			cmd.Dir = cwd
		}

		// Set up context with timeout if specified
		var ctx context.Context
		var cancel context.CancelFunc
		if timeout > 0 {
			ctx, cancel = context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
			defer cancel()
			cmd = exec.CommandContext(ctx, cmd.Args[0], cmd.Args[1:]...)
		}

		// Execute command
		output, err := cmd.CombinedOutput()

		// Call back in event loop
		if cpm.loop != nil && !js.IsUndefined(callback) {
			cpm.loop.RunOnLoop(func(vm *js.Runtime) {
				if callFunc, ok := js.AssertFunction(callback); ok {
					var errorArg js.Value = js.Null()
					var stdout, stderr js.Value

					if err != nil {
						// Create error object
						errorObj := vm.NewObject()
						errorObj.Set("message", err.Error())
						errorObj.Set("code", getExitCode(err))
						errorObj.Set("killed", false)
						if ctx != nil && ctx.Err() == context.DeadlineExceeded {
							errorObj.Set("killed", true)
							errorObj.Set("signal", "SIGTERM")
						}
						errorArg = errorObj
					}

					// Set stdout and stderr
					if encoding == "buffer" {
						// Return as buffer (array of bytes)
						buffer := vm.NewArray()
						for i, b := range output {
							buffer.Set(fmt.Sprintf("%d", i), vm.ToValue(int(b)))
						}
						stdout = buffer
						stderr = buffer
					} else {
						stdout = vm.ToValue(string(output))
						stderr = vm.ToValue("")
					}

					callFunc(js.Undefined(), errorArg, stdout, stderr)
				}
			})
		}
	}()

	return js.Undefined()
}

// spawn creates a new process (simplified implementation)
func (cpm *ChildProcessModule) spawn(vm *js.Runtime, command string, args []string, options map[string]interface{}) js.Value {
	// Create a child process object
	childObj := vm.NewObject()

	// Parse options
	cwd := ""
	if dir, ok := options["cwd"].(string); ok {
		cwd = dir
	}

	// Create command
	cmd := exec.Command(command, args...)
	if cwd != "" {
		cmd.Dir = cwd
	}

	// Create event emitter-like object
	childObj.Set("pid", 0) // Will be set after start

	// stdout stream (simplified)
	stdoutObj := vm.NewObject()
	stdoutObj.Set("_data", vm.ToValue(""))
	childObj.Set("stdout", stdoutObj)

	// stderr stream (simplified)
	stderrObj := vm.NewObject()
	stderrObj.Set("_data", vm.ToValue(""))
	childObj.Set("stderr", stderrObj)

	// kill method
	childObj.Set("kill", func(call js.FunctionCall) js.Value {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return vm.ToValue(true)
	})

	// Start the process asynchronously
	go func() {
		if err := cmd.Start(); err != nil {
			// Emit error event
			return
		}

		// Set PID
		if cpm.loop != nil {
			cpm.loop.RunOnLoop(func(vm *js.Runtime) {
				childObj.Set("pid", cmd.Process.Pid)
			})
		}

		// Wait for process to complete
		err := cmd.Wait()
		
		if cpm.loop != nil {
			cpm.loop.RunOnLoop(func(vm *js.Runtime) {
				// Emit close event
				exitCode := 0
				if err != nil {
					exitCode = getExitCode(err)
				}
				// In a real implementation, we'd emit events here
				// For now, just set the exit code
				childObj.Set("exitCode", exitCode)
			})
		}
	}()

	return childObj
}


// getExitCode extracts exit code from error
func getExitCode(err error) int {
	if exitError, ok := err.(*exec.ExitError); ok {
		return exitError.ExitCode()
	}
	return 1
}