package fs

import (
	"embed"
	"testing"

	js "github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/eventloop"
	"github.com/dop251/goja_nodejs/require"
	"github.com/rediwo/redi/filesystem"
)

//go:embed readonly_test.go
var testFS embed.FS

func TestReadOnlyFilesystem(t *testing.T) {
	// Create embed filesystem (read-only)
	embedFS := filesystem.NewEmbedFileSystem(testFS)
	
	// Create event loop
	loop := eventloop.NewEventLoop()
	go loop.Start()
	defer loop.Stop()
	
	// Test with read-only filesystem
	loop.RunOnLoop(func(vm *js.Runtime) {
		registry := require.NewRegistry()
		EnableWithEventLoopAndFS(registry, embedFS, "", loop)
		registry.Enable(vm)
		
		// Get fs module
		fsModule := require.Require(vm, "fs")
		vm.Set("fs", fsModule)
		
		// Test that write operations throw errors
		writeOperations := []string{
			"fs.writeFileSync('test.txt', 'hello')",
			"fs.mkdirSync('testdir')",
			"fs.unlinkSync('test.txt')",
			"fs.copyFileSync('source.txt', 'dest.txt')",
		}
		
		for _, operation := range writeOperations {
			t.Run("sync_"+operation, func(t *testing.T) {
				defer func() {
					if r := recover(); r != nil {
						// Expected to panic with "filesystem is read-only"
						if err, ok := r.(*js.Exception); ok {
							if err.Error() != "GoError: filesystem is read-only" {
								t.Errorf("Expected 'filesystem is read-only' error, got: %v", err.Error())
							}
						} else {
							t.Errorf("Expected GoError exception, got: %v", r)
						}
					} else {
						t.Error("Expected operation to fail on read-only filesystem")
					}
				}()
				
				_, err := vm.RunString(operation)
				if err == nil {
					t.Error("Expected operation to fail on read-only filesystem")
				}
			})
		}
		
		// Test that read operations work
		readOperations := []string{
			"fs.existsSync('readonly_test.go')",
		}
		
		for _, operation := range readOperations {
			t.Run("read_"+operation, func(t *testing.T) {
				val, err := vm.RunString(operation)
				if err != nil {
					t.Errorf("Read operation failed: %v", err)
				}
				
				if operation == "fs.existsSync('readonly_test.go')" {
					if !val.ToBoolean() {
						t.Error("Expected readonly_test.go to exist")
					}
				}
			})
		}
		
		// Test async operations also fail
		asyncWriteOperations := []struct {
			name string
			code string
		}{
			{"writeFile", "fs.writeFile('test.txt', 'hello', function(err) { if (!err) throw new Error('Should have failed'); })"},
			{"mkdir", "fs.mkdir('testdir', function(err) { if (!err) throw new Error('Should have failed'); })"},
			{"unlink", "fs.unlink('test.txt', function(err) { if (!err) throw new Error('Should have failed'); })"},
			{"copyFile", "fs.copyFile('source.txt', 'dest.txt', function(err) { if (!err) throw new Error('Should have failed'); })"},
		}
		
		for _, operation := range asyncWriteOperations {
			t.Run("async_"+operation.name, func(t *testing.T) {
				defer func() {
					if r := recover(); r != nil {
						// Expected to panic with "filesystem is read-only"
						if err, ok := r.(*js.Exception); ok {
							if err.Error() != "GoError: filesystem is read-only" {
								t.Errorf("Expected 'filesystem is read-only' error, got: %v", err.Error())
							}
						} else {
							t.Errorf("Expected GoError exception, got: %v", r)
						}
					} else {
						t.Error("Expected async operation to fail on read-only filesystem")
					}
				}()
				
				_, err := vm.RunString(operation.code)
				if err == nil {
					t.Error("Expected async operation to fail on read-only filesystem")
				}
			})
		}
	})
}

func TestWritableFilesystem(t *testing.T) {
	// Create OS filesystem (writable)
	osFS := filesystem.NewOSFileSystem("./test_temp")
	
	// Create event loop
	loop := eventloop.NewEventLoop()
	go loop.Start()
	defer loop.Stop()
	
	// Test with writable filesystem
	loop.RunOnLoop(func(vm *js.Runtime) {
		registry := require.NewRegistry()
		EnableWithEventLoopAndFS(registry, osFS, "", loop)
		registry.Enable(vm)
		
		// Get fs module
		fsModule := require.Require(vm, "fs")
		vm.Set("fs", fsModule)
		
		// Test that write operations don't immediately fail due to read-only check
		// (they might fail for other reasons like directory not existing, but not read-only)
		writeOperations := []string{
			"try { fs.writeFileSync('test.txt', 'hello'); } catch (e) { if (e.message.includes('read-only')) throw e; }",
		}
		
		for _, operation := range writeOperations {
			t.Run("writable_"+operation, func(t *testing.T) {
				_, err := vm.RunString(operation)
				if err != nil {
					t.Errorf("Write operation failed with read-only check error: %v", err)
				}
			})
		}
	})
}