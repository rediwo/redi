package fs

import (
	"os"
	"path/filepath"
	"testing"
	"time"
	
	js "github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/eventloop"
	"github.com/dop251/goja_nodejs/require"
	"github.com/rediwo/redi/filesystem"
	"github.com/rediwo/redi/registry"
)

func TestFSModule(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "fs_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Set up VM and registry
	vm := js.New()
	requireRegistry := require.NewRegistry()
	
	// Use the actual initFSModule function to ensure we test the real code path
	osFS := filesystem.NewOSFileSystem("")
	config := registry.ModuleConfig{
		Registry:   requireRegistry,
		FileSystem: osFS,
		BasePath:   tmpDir,
		VM:         vm,
	}
	err = initFSModule(config)
	if err != nil {
		t.Fatalf("Failed to initialize fs module: %v", err)
	}
	requireRegistry.Enable(vm)
	
	// Get fs module
	fs := require.Require(vm, "fs")
	vm.Set("fs", fs)
	
	t.Run("writeFileSync and readFileSync", func(t *testing.T) {
		testContent := "Hello, World!"
		testFile := "test.txt"
		
		// Write file
		_, err := vm.RunString(`fs.writeFileSync("` + testFile + `", "` + testContent + `")`)
		if err != nil {
			t.Fatalf("writeFileSync failed: %v", err)
		}
		
		// Check file was created
		filePath := filepath.Join(tmpDir, testFile)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Fatal("File was not created")
		}
		
		// Read file
		result, err := vm.RunString(`fs.readFileSync("` + testFile + `")`)
		if err != nil {
			t.Fatalf("readFileSync failed: %v", err)
		}
		
		if result.String() != testContent {
			t.Errorf("Expected %q, got %q", testContent, result.String())
		}
	})
	
	t.Run("existsSync", func(t *testing.T) {
		// Test existing file
		result, err := vm.RunString(`fs.existsSync("test.txt")`)
		if err != nil {
			t.Fatalf("existsSync failed: %v", err)
		}
		
		if !result.ToBoolean() {
			t.Error("existsSync should return true for existing file")
		}
		
		// Test non-existing file
		result, err = vm.RunString(`fs.existsSync("nonexistent.txt")`)
		if err != nil {
			t.Fatalf("existsSync failed: %v", err)
		}
		
		if result.ToBoolean() {
			t.Error("existsSync should return false for non-existing file")
		}
	})
	
	t.Run("mkdirSync", func(t *testing.T) {
		// Create directory
		_, err := vm.RunString(`fs.mkdirSync("testdir")`)
		if err != nil {
			t.Fatalf("mkdirSync failed: %v", err)
		}
		
		// Check directory was created
		dirPath := filepath.Join(tmpDir, "testdir")
		if stat, err := os.Stat(dirPath); os.IsNotExist(err) || !stat.IsDir() {
			t.Fatal("Directory was not created")
		}
		
		// Test recursive creation
		_, err = vm.RunString(`fs.mkdirSync("nested/deep/dir", {recursive: true})`)
		if err != nil {
			t.Fatalf("mkdirSync recursive failed: %v", err)
		}
		
		nestedPath := filepath.Join(tmpDir, "nested", "deep", "dir")
		if stat, err := os.Stat(nestedPath); os.IsNotExist(err) || !stat.IsDir() {
			t.Fatal("Nested directory was not created")
		}
	})
	
	t.Run("readdirSync", func(t *testing.T) {
		// Create some test files
		testFiles := []string{"file1.txt", "file2.txt"}
		for _, file := range testFiles {
			filePath := filepath.Join(tmpDir, file)
			if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
		}
		
		// Read directory
		result, err := vm.RunString(`fs.readdirSync(".")`)
		if err != nil {
			t.Fatalf("readdirSync failed: %v", err)
		}
		
		// Convert result to slice
		if result.ExportType().Kind().String() != "slice" {
			t.Fatal("readdirSync should return an array")
		}
		
		arr := result.(*js.Object)
		length := arr.Get("length").ToInteger()
		
		if length < 2 {
			t.Errorf("Expected at least 2 files, got %d", length)
		}
	})
	
	t.Run("statSync", func(t *testing.T) {
		// Test file stat
		result, err := vm.RunString(`
			var stat = fs.statSync("test.txt");
			({
				size: stat.size,
				isFile: stat.isFile(),
				isDirectory: stat.isDirectory()
			})
		`)
		if err != nil {
			t.Fatalf("statSync failed: %v", err)
		}
		
		obj := result.(*js.Object)
		isFile := obj.Get("isFile").ToBoolean()
		isDirectory := obj.Get("isDirectory").ToBoolean()
		
		if !isFile {
			t.Error("Expected isFile to be true")
		}
		if isDirectory {
			t.Error("Expected isDirectory to be false")
		}
		
		// Test directory stat
		result, err = vm.RunString(`
			var stat = fs.statSync("testdir");
			({
				isFile: stat.isFile(),
				isDirectory: stat.isDirectory()
			})
		`)
		if err != nil {
			t.Fatalf("statSync on directory failed: %v", err)
		}
		
		obj = result.(*js.Object)
		isFile = obj.Get("isFile").ToBoolean()
		isDirectory = obj.Get("isDirectory").ToBoolean()
		
		if isFile {
			t.Error("Expected isFile to be false for directory")
		}
		if !isDirectory {
			t.Error("Expected isDirectory to be true for directory")
		}
	})
	
	t.Run("copyFileSync", func(t *testing.T) {
		srcContent := "Copy test content"
		srcFile := "source.txt"
		dstFile := "destination.txt"
		
		// Create source file
		_, err := vm.RunString(`fs.writeFileSync("` + srcFile + `", "` + srcContent + `")`)
		if err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}
		
		// Copy file
		_, err = vm.RunString(`fs.copyFileSync("` + srcFile + `", "` + dstFile + `")`)
		if err != nil {
			t.Fatalf("copyFileSync failed: %v", err)
		}
		
		// Read destination file
		result, err := vm.RunString(`fs.readFileSync("` + dstFile + `")`)
		if err != nil {
			t.Fatalf("Failed to read destination file: %v", err)
		}
		
		if result.String() != srcContent {
			t.Errorf("Expected %q, got %q", srcContent, result.String())
		}
	})
	
	t.Run("unlinkSync", func(t *testing.T) {
		testFile := "delete_me.txt"
		
		// Create file
		_, err := vm.RunString(`fs.writeFileSync("` + testFile + `", "delete me")`)
		if err != nil {
			t.Fatalf("Failed to create file for deletion: %v", err)
		}
		
		// Verify file exists
		result, err := vm.RunString(`fs.existsSync("` + testFile + `")`)
		if err != nil {
			t.Fatalf("existsSync failed: %v", err)
		}
		if !result.ToBoolean() {
			t.Fatal("File should exist before deletion")
		}
		
		// Delete file
		_, err = vm.RunString(`fs.unlinkSync("` + testFile + `")`)
		if err != nil {
			t.Fatalf("unlinkSync failed: %v", err)
		}
		
		// Verify file is deleted
		result, err = vm.RunString(`fs.existsSync("` + testFile + `")`)
		if err != nil {
			t.Fatalf("existsSync failed: %v", err)
		}
		if result.ToBoolean() {
			t.Error("File should not exist after deletion")
		}
	})
	
	t.Run("error handling", func(t *testing.T) {
		// Test reading non-existent file
		_, err := vm.RunString(`fs.readFileSync("nonexistent.txt")`)
		if err == nil {
			t.Error("Expected error when reading non-existent file")
		}
		
		// Test writing to invalid path (assuming no permission)
		_, err = vm.RunString(`fs.writeFileSync("/invalid/path/file.txt", "content")`)
		if err == nil {
			t.Error("Expected error when writing to invalid path")
		}
	})
	
	t.Run("path resolution", func(t *testing.T) {
		// Test that relative paths are resolved from base directory
		subDir := "subdir"
		subDirPath := filepath.Join(tmpDir, subDir)
		if err := os.Mkdir(subDirPath, 0755); err != nil {
			t.Fatalf("Failed to create subdirectory: %v", err)
		}
		
		testFile := filepath.Join(subDir, "subfile.txt")
		testContent := "subdirectory file"
		
		// Write file in subdirectory
		_, err := vm.RunString(`fs.writeFileSync("` + testFile + `", "` + testContent + `")`)
		if err != nil {
			t.Fatalf("Failed to write file in subdirectory: %v", err)
		}
		
		// Read file from subdirectory
		result, err := vm.RunString(`fs.readFileSync("` + testFile + `")`)
		if err != nil {
			t.Fatalf("Failed to read file from subdirectory: %v", err)
		}
		
		if result.String() != testContent {
			t.Errorf("Expected %q, got %q", testContent, result.String())
		}
	})
	
	// === ASYNC FUNCTION TESTS ===
	// Set up new VM with event loop for async functions
	loop := eventloop.NewEventLoop()
	go loop.Start()
	defer loop.Stop()
	
	// Create new VM for async tests
	vmAsync := js.New()
	requireRegistryAsync := require.NewRegistry()
	
	// Use the actual initFSModule function for async testing
	osfsAsync := filesystem.NewOSFileSystem("")
	configAsync := registry.ModuleConfig{
		Registry:   requireRegistryAsync,
		FileSystem: osfsAsync,
		BasePath:   tmpDir,
		EventLoop:  loop,
		VM:         vmAsync,
	}
	err = initFSModule(configAsync)
	if err != nil {
		t.Fatalf("Failed to initialize fs module for async tests: %v", err)
	}
	requireRegistryAsync.Enable(vmAsync)
	
	// Get fs module for async tests
	fsAsync := require.Require(vmAsync, "fs")
	vmAsync.Set("fs", fsAsync)
	
	t.Run("readFile async", func(t *testing.T) {
		testContent := "Async read test"
		testFile := "async_read.txt"
		
		// First create the file synchronously
		_, err := vmAsync.RunString(`fs.writeFileSync("` + testFile + `", "` + testContent + `")`)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		
		// Test async read with callback
		done := make(chan bool, 1)
		var resultContent string
		var callbackError string
		
		vmAsync.Set("testCallback", func(err js.Value, data js.Value) {
			if !js.IsNull(err) && !js.IsUndefined(err) {
				callbackError = err.String()
			} else {
				resultContent = data.String()
			}
			done <- true
		})
		
		_, err = vmAsync.RunString(`fs.readFile("` + testFile + `", testCallback)`)
		if err != nil {
			t.Fatalf("Failed to call readFile: %v", err)
		}
		
		// Wait for callback
		select {
		case <-done:
			if callbackError != "" {
				t.Fatalf("readFile callback error: %s", callbackError)
			}
			if resultContent != testContent {
				t.Errorf("Expected %q, got %q", testContent, resultContent)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("readFile callback not called within timeout")
		}
	})
}