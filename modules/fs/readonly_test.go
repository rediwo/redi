package fs

import (
	"embed"
	"os"
	"strings"
	"testing"

	js "github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/eventloop"
	"github.com/dop251/goja_nodejs/require"
	"github.com/rediwo/redi/filesystem"
	"github.com/rediwo/redi/registry"
)

//go:embed readonly_test.go
var testFS embed.FS

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestReadOnlyFilesystem(t *testing.T) {
	// Create embed filesystem (read-only)
	embedFS := filesystem.NewEmbedFileSystem(testFS)
	
	// Create event loop
	loop := eventloop.NewEventLoop()
	go loop.Start()
	defer loop.Stop()
	
	// Set up VM and modules outside of subtests
	vm := js.New()
	requireRegistry := require.NewRegistry()
	
	// Use the actual initFSModule function with read-only filesystem
	config := registry.ModuleConfig{
		Registry:   requireRegistry,
		FileSystem: embedFS,
		BasePath:   "",
		EventLoop:  loop,
		VM:         vm,
	}
	err := initFSModule(config)
	if err != nil {
		t.Fatalf("Failed to initialize fs module: %v", err)
	}
	requireRegistry.Enable(vm)
	
	// Get fs module
	fsObject := require.Require(vm, "fs")
	vm.Set("fs", fsObject)
	
	// Test that write operations throw errors
	writeOperations := []string{
		"fs.writeFileSync('test.txt', 'hello')",
		"fs.mkdirSync('testdir')",
		"fs.unlinkSync('test.txt')",
		"fs.copyFileSync('source.txt', 'dest.txt')",
	}
	
	for _, operation := range writeOperations {
		t.Run("sync_"+operation, func(t *testing.T) {
			_, err := vm.RunString(operation)
			if err == nil {
				t.Error("Expected operation to fail on read-only filesystem")
			} else {
				// Check that it's the expected read-only error
				if !contains(err.Error(), "filesystem is read-only") {
					t.Errorf("Expected 'filesystem is read-only' error, got: %v", err.Error())
				}
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
			_, err := vm.RunString(operation.code)
			if err == nil {
				t.Error("Expected async operation to fail on read-only filesystem")
			} else {
				// Check that it's the expected read-only error
				if !contains(err.Error(), "filesystem is read-only") {
					t.Errorf("Expected 'filesystem is read-only' error, got: %v", err.Error())
				}
			}
		})
	}
}

func TestWritableFilesystem(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "fs_writable_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create OS filesystem (writable)
	osFS := filesystem.NewOSFileSystem("")
	
	// Create event loop
	loop := eventloop.NewEventLoop()
	go loop.Start()
	defer loop.Stop()
	
	// Set up VM and modules outside of subtests
	vm := js.New()
	requireRegistryOS := require.NewRegistry()
	
	// Use the actual initFSModule function with writable filesystem
	config := registry.ModuleConfig{
		Registry:   requireRegistryOS,
		FileSystem: osFS,
		BasePath:   tmpDir,
		EventLoop:  loop,
		VM:         vm,
	}
	err = initFSModule(config)
	if err != nil {
		t.Fatalf("Failed to initialize fs module: %v", err)
	}
	requireRegistryOS.Enable(vm)
	
	// Get fs module
	fsObject := require.Require(vm, "fs")
	vm.Set("fs", fsObject)
	
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
}