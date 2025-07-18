package process

import (
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	js "github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/eventloop"
	"github.com/dop251/goja_nodejs/require"
	"github.com/rediwo/redi/registry"
)

const testVersion = "v1.2.3"

// setupTestVM creates a VM with process module for testing
func setupTestVM(loop *eventloop.EventLoop) *js.Runtime {
	vm := js.New()
	requireRegistry := require.NewRegistry()
	
	// Use the actual initProcessModule function to ensure we test the real code path
	config := registry.ModuleConfig{
		Registry:  requireRegistry,
		EventLoop: loop,
		Version:   testVersion,
		VM:        vm,
	}
	err := initProcessModule(config)
	if err != nil {
		panic("Failed to initialize process module: " + err.Error())
	}
	requireRegistry.Enable(vm)
	
	return vm
}

func TestProcessModule(t *testing.T) {
	vm := setupTestVM(nil)

	// Test process module availability
	_, err := vm.RunString(`
		var process = require('process');
		if (typeof process !== 'object') {
			throw new Error('process module should export an object');
		}
	`)
	if err != nil {
		t.Fatalf("Failed to load process module: %v", err)
	}

	// Test basic properties
	t.Run("BasicProperties", func(t *testing.T) {
		tests := []struct {
			property string
			jsType   string
		}{
			{"process.version", "string"},
			{"process.versions", "object"},
			{"process.platform", "string"},
			{"process.arch", "string"},
			{"process.pid", "number"},
			{"process.ppid", "number"},
			{"process.env", "object"},
			{"process.argv", "object"},
			{"process.argv0", "string"},
			{"process.execPath", "string"},
			{"process.execArgv", "object"},
			{"process.title", "string"},
		}

		for _, test := range tests {
			t.Run(test.property, func(t *testing.T) {
				script := `
					var process = require('process');
					typeof ` + test.property

				result, err := vm.RunString(script)
				if err != nil {
					t.Fatalf("Failed to get %s: %v", test.property, err)
				}

				if result.String() != test.jsType {
					t.Errorf("Expected %s to be %s, got %s", test.property, test.jsType, result.String())
				}
			})
		}
	})

	// Test process.cwd()
	t.Run("Cwd", func(t *testing.T) {
		expectedCwd, _ := os.Getwd()

		result, err := vm.RunString(`
			var process = require('process');
			process.cwd();
		`)
		if err != nil {
			t.Fatalf("Failed to call process.cwd(): %v", err)
		}

		if result.String() != expectedCwd {
			t.Errorf("Expected cwd to be %s, got %s", expectedCwd, result.String())
		}
	})

	// Test process.uptime()
	t.Run("Uptime", func(t *testing.T) {
		time.Sleep(10 * time.Millisecond) // Ensure some uptime

		result, err := vm.RunString(`
			var process = require('process');
			process.uptime();
		`)
		if err != nil {
			t.Fatalf("Failed to call process.uptime(): %v", err)
		}

		uptime := result.ToFloat()
		if uptime <= 0 {
			t.Errorf("Expected uptime to be > 0, got %f", uptime)
		}
	})

	// Test process.hrtime()
	t.Run("Hrtime", func(t *testing.T) {
		result, err := vm.RunString(`
			var process = require('process');
			var start = process.hrtime();
			// Simulate some work
			for (var i = 0; i < 1000; i++) {}
			var diff = process.hrtime(start);
			({
				start: start,
				diff: diff,
				startIsArray: Array.isArray(start),
				diffIsArray: Array.isArray(diff),
				startLength: start.length,
				diffLength: diff.length
			})
		`)
		if err != nil {
			t.Fatalf("Failed to call process.hrtime(): %v", err)
		}

		obj := result.(*js.Object)
		if !obj.Get("startIsArray").ToBoolean() {
			t.Error("process.hrtime() should return an array")
		}
		if !obj.Get("diffIsArray").ToBoolean() {
			t.Error("process.hrtime(time) should return an array")
		}
		if obj.Get("startLength").ToInteger() != 2 {
			t.Error("process.hrtime() should return array of length 2")
		}
		if obj.Get("diffLength").ToInteger() != 2 {
			t.Error("process.hrtime(time) should return array of length 2")
		}
	})

	// Test process.memoryUsage()
	t.Run("MemoryUsage", func(t *testing.T) {
		result, err := vm.RunString(`
			var process = require('process');
			var mem = process.memoryUsage();
			({
				hasRss: typeof mem.rss === 'number',
				hasHeapTotal: typeof mem.heapTotal === 'number',
				hasHeapUsed: typeof mem.heapUsed === 'number',
				hasExternal: typeof mem.external === 'number',
				hasArrayBuffers: typeof mem.arrayBuffers === 'number'
			})
		`)
		if err != nil {
			t.Fatalf("Failed to call process.memoryUsage(): %v", err)
		}

		obj := result.(*js.Object)
		properties := []string{"hasRss", "hasHeapTotal", "hasHeapUsed", "hasExternal", "hasArrayBuffers"}
		for _, prop := range properties {
			if !obj.Get(prop).ToBoolean() {
				t.Errorf("process.memoryUsage() should have %s as number", prop[3:])
			}
		}
	})

	// Test process.cpuUsage()
	t.Run("CpuUsage", func(t *testing.T) {
		result, err := vm.RunString(`
			var process = require('process');
			var cpu = process.cpuUsage();
			({
				hasUser: typeof cpu.user === 'number',
				hasSystem: typeof cpu.system === 'number'
			})
		`)
		if err != nil {
			t.Fatalf("Failed to call process.cpuUsage(): %v", err)
		}

		obj := result.(*js.Object)
		if !obj.Get("hasUser").ToBoolean() {
			t.Error("process.cpuUsage() should have user as number")
		}
		if !obj.Get("hasSystem").ToBoolean() {
			t.Error("process.cpuUsage() should have system as number")
		}
	})

	// Test process.env
	t.Run("Env", func(t *testing.T) {
		// Set a test environment variable before creating the module
		os.Setenv("TEST_PROCESS_ENV", "test_value")
		defer os.Unsetenv("TEST_PROCESS_ENV")

		// Create a new VM to pick up the new environment variable
		testVM := setupTestVM(nil)

		result, err := testVM.RunString(`
			var process = require('process');
			process.env.TEST_PROCESS_ENV;
		`)
		if err != nil {
			t.Fatalf("Failed to access process.env: %v", err)
		}

		if result.String() != "test_value" {
			t.Errorf("Expected TEST_PROCESS_ENV to be 'test_value', got %s", result.String())
		}
	})

	// Test process.argv
	t.Run("Argv", func(t *testing.T) {
		result, err := vm.RunString(`
			var process = require('process');
			({
				isArray: Array.isArray(process.argv),
				length: process.argv.length,
				firstArg: process.argv[0],
				secondArg: process.argv[1],
				argv0: process.argv0
			})
		`)
		if err != nil {
			t.Fatalf("Failed to access process.argv: %v", err)
		}

		obj := result.(*js.Object)
		if !obj.Get("isArray").ToBoolean() {
			t.Error("process.argv should be an array")
		}
		if obj.Get("length").ToInteger() < 2 {
			t.Error("process.argv should have at least 2 elements")
		}

		firstArg := obj.Get("firstArg").String()
		if firstArg == "" {
			t.Error("process.argv[0] should contain executable path")
		}

		secondArg := obj.Get("secondArg").String()
		if secondArg != "" {
			t.Error("process.argv[1] should be empty string")
		}

		argv0 := obj.Get("argv0").String()
		if argv0 == "" {
			t.Error("process.argv0 should contain executable name")
		}
	})

	// Test platform-specific functions (Unix only)
	if runtime.GOOS != "windows" {
		t.Run("UnixFunctions", func(t *testing.T) {
			functions := []string{"getuid", "getgid", "geteuid", "getegid"}

			for _, funcName := range functions {
				t.Run(funcName, func(t *testing.T) {
					script := `
						var process = require('process');
						typeof process.` + funcName + `();
					`

					result, err := vm.RunString(script)
					if err != nil {
						t.Fatalf("Failed to call process.%s(): %v", funcName, err)
					}

					if result.String() != "number" {
						t.Errorf("process.%s() should return a number, got %s", funcName, result.String())
					}
				})
			}
		})
	}

	// Test stdout/stderr
	t.Run("StdStreams", func(t *testing.T) {
		_, err := vm.RunString(`
			var process = require('process');
			if (typeof process.stdout !== 'object') {
				throw new Error('process.stdout should be an object');
			}
			if (typeof process.stderr !== 'object') {
				throw new Error('process.stderr should be an object');
			}
			if (typeof process.stdin !== 'object') {
				throw new Error('process.stdin should be an object');
			}
			if (typeof process.stdout.write !== 'function') {
				throw new Error('process.stdout.write should be a function');
			}
		`)
		if err != nil {
			t.Fatalf("Failed to test std streams: %v", err)
		}
	})
}

func TestProcessVersion(t *testing.T) {
	loop := eventloop.NewEventLoop()
	vm := setupTestVM(loop)

	loop.Start()
	defer loop.Stop()

	loop.RunOnLoop(func(loopVm *js.Runtime) {
		// Test process.version specifically
		result, err := vm.RunString(`
			var process = require('process');
			({
				version: process.version,
				versionType: typeof process.version,
				versionLength: process.version.length,
				startsWithV: process.version.charAt(0) === 'v',
				containsDots: process.version.indexOf('.') > 0
			})
		`)
		if err != nil {
			t.Fatalf("Failed to test process.version: %v", err)
		}

		obj := result.(*js.Object)

		// Verify process.version is a string
		if obj.Get("versionType").String() != "string" {
			t.Error("process.version should be a string")
		}

		version := obj.Get("version").String()
		if version == "" {
			t.Error("process.version should not be empty")
		}

		// Check version format (should be like v1.2.3)
		if !obj.Get("startsWithV").ToBoolean() {
			t.Error("process.version should start with 'v'")
		}

		if !obj.Get("containsDots").ToBoolean() {
			t.Error("process.version should contain dots for semantic versioning")
		}

		if obj.Get("versionLength").ToInteger() < 5 {
			t.Error("process.version should be at least 5 characters long (e.g., v1.0.0)")
		}

		t.Logf("process.version: %s", version)
	})
}

func TestProcessVersions(t *testing.T) {
	loop := eventloop.NewEventLoop()
	vm := setupTestVM(loop)

	loop.Start()
	defer loop.Stop()

	loop.RunOnLoop(func(loopVm *js.Runtime) {
		result, err := vm.RunString(`
			var process = require('process');
			({
				hasGo: typeof process.versions.go === 'string',
				processVersion: process.version,
				goVersion: process.versions.go,
				versionKeys: Object.keys(process.versions),
				versionsType: typeof process.versions
			})
		`)
		if err != nil {
			t.Fatalf("Failed to test process.versions: %v", err)
		}

		obj := result.(*js.Object)

		// Test process.versions type
		if obj.Get("versionsType").String() != "object" {
			t.Error("process.versions should be an object")
		}

		if !obj.Get("hasGo").ToBoolean() {
			t.Error("process.versions.go should be a string")
		}

		processVersion := obj.Get("processVersion").String()
		if processVersion != testVersion {
			t.Errorf("Expected process version to be %s, got %s", testVersion, processVersion)
		}

		goVersion := obj.Get("goVersion").String()
		if !strings.HasPrefix(goVersion, "go") {
			t.Errorf("Expected Go version to start with 'go', got %s", goVersion)
		}

		// Check that we have at least some modules listed
		versionKeys := obj.Get("versionKeys").Export()
		if keys, ok := versionKeys.([]any); ok {
			if len(keys) < 1 { // At least go
				t.Errorf("Expected at least 1 version key, got %d", len(keys))
			}
		}
	})
}

func TestProcessNextTick(t *testing.T) {
	loop := eventloop.NewEventLoop()
	vm := setupTestVM(loop)

	// Set up test variables in the VM
	vm.Set("testCalled", false)
	vm.Set("testResult", js.Null())
	vm.Set("callbackArgs", js.Null())
	vm.Set("callbackThis", js.Null())

	loop.Start()
	defer loop.Stop()

	// Test 1: Basic nextTick execution
	_, err := vm.RunString(`
		var process = require('process');
		
		process.nextTick(function() {
			testCalled = true;
			testResult = 'nextTick executed';
		});
		
		// Initial state should be false (callback should not execute immediately)
		if (testCalled) {
			throw new Error('nextTick should not execute immediately');
		}
	`)
	if err != nil {
		t.Fatalf("Failed to setup process.nextTick(): %v", err)
	}

	// Give event loop time to process
	time.Sleep(100 * time.Millisecond)

	// Check the results
	called := vm.Get("testCalled").ToBoolean()
	result := vm.Get("testResult").String()

	if !called {
		t.Error("process.nextTick() callback should have been called")
	}
	if result != "nextTick executed" {
		t.Errorf("process.nextTick() callback should have executed with correct result, got: %s", result)
	}

	// Test 2: nextTick with arguments
	vm.Set("argTestCalled", false)
	vm.Set("receivedArgs", js.Null())

	_, err = vm.RunString(`
		process.nextTick(function(arg1, arg2, arg3) {
			argTestCalled = true;
			receivedArgs = [arg1, arg2, arg3];
		}, 'hello', 42, true);
	`)
	if err != nil {
		t.Fatalf("Failed to setup process.nextTick() with arguments: %v", err)
	}

	// Give event loop time to process
	time.Sleep(100 * time.Millisecond)

	argCalled := vm.Get("argTestCalled").ToBoolean()
	if !argCalled {
		t.Error("process.nextTick() callback with arguments should have been called")
	}

	// Check received arguments
	argsResult, err := vm.RunString(`
		({
			arg1: receivedArgs[0],
			arg2: receivedArgs[1],
			arg3: receivedArgs[2],
			length: receivedArgs.length
		})
	`)
	if err != nil {
		t.Fatalf("Failed to check nextTick arguments: %v", err)
	}

	argsObj := argsResult.(*js.Object)
	if argsObj.Get("arg1").String() != "hello" {
		t.Errorf("Expected first argument to be 'hello', got: %s", argsObj.Get("arg1").String())
	}
	if argsObj.Get("arg2").ToInteger() != 42 {
		t.Errorf("Expected second argument to be 42, got: %d", argsObj.Get("arg2").ToInteger())
	}
	if !argsObj.Get("arg3").ToBoolean() {
		t.Error("Expected third argument to be true")
	}
	if argsObj.Get("length").ToInteger() != 3 {
		t.Errorf("Expected 3 arguments, got: %d", argsObj.Get("length").ToInteger())
	}

	// Test 3: Multiple nextTick calls (execution order)
	vm.Set("executionOrder", vm.NewArray())

	_, err = vm.RunString(`
		var order = [];
		
		process.nextTick(function() {
			order.push('first');
		});
		
		process.nextTick(function() {
			order.push('second');
		});
		
		process.nextTick(function() {
			order.push('third');
		});
		
		executionOrder = order;
	`)
	if err != nil {
		t.Fatalf("Failed to setup multiple process.nextTick() calls: %v", err)
	}

	// Give event loop time to process all callbacks
	time.Sleep(200 * time.Millisecond)

	orderResult, err := vm.RunString(`
		({
			length: executionOrder.length,
			first: executionOrder[0],
			second: executionOrder[1], 
			third: executionOrder[2]
		})
	`)
	if err != nil {
		t.Fatalf("Failed to check execution order: %v", err)
	}

	orderObj := orderResult.(*js.Object)
	if orderObj.Get("length").ToInteger() != 3 {
		t.Errorf("Expected 3 callbacks to be executed, got: %d", orderObj.Get("length").ToInteger())
	}
	if orderObj.Get("first").String() != "first" {
		t.Errorf("Expected first callback to execute first, got: %s", orderObj.Get("first").String())
	}
	if orderObj.Get("second").String() != "second" {
		t.Errorf("Expected second callback to execute second, got: %s", orderObj.Get("second").String())
	}
	if orderObj.Get("third").String() != "third" {
		t.Errorf("Expected third callback to execute third, got: %s", orderObj.Get("third").String())
	}
}

func TestProcessChdir(t *testing.T) {
	vm := setupTestVM(nil)

	// Get original directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir) // Restore after test

	// Test changing to parent directory
	_, err := vm.RunString(`
		var process = require('process');
		var originalCwd = process.cwd();
		process.chdir('..');
		var newCwd = process.cwd();
		
		if (newCwd === originalCwd) {
			throw new Error('Directory should have changed');
		}
		
		// Change back
		process.chdir(originalCwd);
		var restoredCwd = process.cwd();
		
		if (restoredCwd !== originalCwd) {
			throw new Error('Directory should be restored');
		}
	`)
	if err != nil {
		t.Fatalf("Failed to test process.chdir(): %v", err)
	}
}
