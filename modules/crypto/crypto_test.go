package crypto

import (
	"testing"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/require"
	"github.com/rediwo/redi/registry"
)

func TestCryptoModule(t *testing.T) {
	vm := goja.New()
	requireRegistry := require.NewRegistry()

	// Initialize crypto module
	config := registry.ModuleConfig{
		Registry: requireRegistry,
		VM:       vm,
	}
	
	if err := initCryptoModule(config); err != nil {
		t.Fatalf("Failed to initialize crypto module: %v", err)
	}

	requireRegistry.Enable(vm)

	tests := []struct {
		name     string
		script   string
		expected interface{}
	}{
		{
			name: "MD5 hash",
			script: `
				var crypto = require('crypto');
				var hash = crypto.createHash('md5');
				hash.update('hello world');
				hash.digest('hex');
			`,
			expected: "5eb63bbbe01eeed093cb22bb8f5acdc3",
		},
		{
			name: "SHA256 hash",
			script: `
				var crypto = require('crypto');
				var hash = crypto.createHash('sha256');
				hash.update('hello world');
				hash.digest('hex');
			`,
			expected: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
		{
			name: "HMAC-SHA256",
			script: `
				var crypto = require('crypto');
				var hmac = crypto.createHmac('sha256', 'secret-key');
				hmac.update('hello world');
				hmac.digest('hex');
			`,
			expected: "095d5a21fe6d0646db223fdf3de6436bb8dfb2fab0b51677ecf6441fcf5f2a67",
		},
		{
			name: "Hash chaining",
			script: `
				var crypto = require('crypto');
				var hash = crypto.createHash('sha1');
				hash.update('hello');
				hash.update(' ');
				hash.update('world');
				hash.digest('hex');
			`,
			expected: "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed",
		},
		{
			name: "Random bytes sync",
			script: `
				var crypto = require('crypto');
				var bytes = crypto.randomBytesSync(16);
				bytes.byteLength;
			`,
			expected: int64(16),
		},
		{
			name: "PBKDF2 sync",
			script: `
				var crypto = require('crypto');
				var key = crypto.pbkdf2Sync('password', 'salt', 1000, 32, 'sha256');
				key.byteLength;
			`,
			expected: int64(32),
		},
		{
			name: "Get hashes",
			script: `
				var crypto = require('crypto');
				var hashes = crypto.getHashes();
				hashes.length > 0 && hashes.indexOf('sha256') >= 0;
			`,
			expected: true,
		},
		{
			name: "Unknown algorithm error",
			script: `
				var crypto = require('crypto');
				try {
					crypto.createHash('unknown-algo');
					false;
				} catch (e) {
					e.message.indexOf('Unknown hash algorithm') >= 0;
				}
			`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := vm.RunString(tt.script)
			if err != nil {
				t.Fatalf("Script execution failed: %v", err)
			}

			actual := result.Export()
			if actual != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, actual)
			}
		})
	}
}

func TestCryptoAsync(t *testing.T) {
	vm := goja.New()
	requireRegistry := require.NewRegistry()

	// Initialize crypto module without event loop (using fallback)
	config := registry.ModuleConfig{
		Registry: requireRegistry,
		VM:       vm,
		EventLoop: nil, // Test without event loop
	}
	
	if err := initCryptoModule(config); err != nil {
		t.Fatalf("Failed to initialize crypto module: %v", err)
	}

	requireRegistry.Enable(vm)

	// Test async randomBytes
	script := `
		var crypto = require('crypto');
		var result = null;
		var error = null;
		var called = false;
		
		crypto.randomBytes(16, function(err, buf) {
			called = true;
			error = err;
			result = buf;
		});
		
		// For test without event loop, callback is called immediately in goroutine
		// Just check that the function was registered
		true;
	`

	result, err := vm.RunString(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	// Just verify the script runs without error
	if result.Export() != true {
		t.Errorf("Script should return true")
	}
}

func TestPBKDF2Async(t *testing.T) {
	vm := goja.New()
	requireRegistry := require.NewRegistry()

	// Initialize crypto module without event loop (using fallback)
	config := registry.ModuleConfig{
		Registry: requireRegistry,
		VM:       vm,
		EventLoop: nil, // Test without event loop
	}
	
	if err := initCryptoModule(config); err != nil {
		t.Fatalf("Failed to initialize crypto module: %v", err)
	}

	requireRegistry.Enable(vm)

	// Test async pbkdf2 - with 5 arguments
	script := `
		var crypto = require('crypto');
		var called = false;
		
		crypto.pbkdf2('password', 'salt', 1000, 32, function(err, key) {
			called = true;
		});
		
		// Also test with 6 arguments
		crypto.pbkdf2('password', 'salt', 1000, 32, 'sha256', function(err, key) {
			called = true;
		});
		
		true;
	`

	result, err := vm.RunString(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	// Just verify the script runs without error
	if result.Export() != true {
		t.Errorf("Script should return true")
	}
}