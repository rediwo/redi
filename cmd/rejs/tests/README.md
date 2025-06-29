# rejs Test Scripts

This directory contains comprehensive test scripts for the rejs JavaScript runtime.

## Running Tests

### Synchronous Tests (auto-exit)
```bash
# These tests complete quickly and exit automatically
./rejs tests/test_path.js          # ~0.1s - Path module tests
./rejs tests/test_require.js       # ~0.1s - Module system tests
```

### Asynchronous Tests (require --timeout)
```bash
# These tests need timeout for async operations (timers, HTTP requests, processes)
./rejs --timeout=3000 tests/test_basic.js         # ~0.9s - Console, process, timers
./rejs --timeout=10000 tests/test_fs.js           # ~1-2s - File system operations  
./rejs --timeout=15000 tests/test_child_process.js # ~1-3s - Child process operations
./rejs --timeout=20000 tests/test_fetch.js        # ~2-5s - HTTP requests
```

### Alternative Timeout Syntax
```bash
# Both syntax forms are supported
./rejs --timeout=5000 script.js
./rejs --timeout 5000 script.js
```

## Test Files

### test_basic.js ⏱️ (Async - requires --timeout)
Tests fundamental rejs functionality:
- **Console**: log, warn, error output
- **Process**: argv, platform, arch, pid, ppid, version, cwd(), env
- **Globals**: __filename, __dirname
- **Timers**: setTimeout, setInterval, clearInterval
- **Execution**: Uses multiple timers, needs ~3 seconds to complete

### test_fs.js ⏱️ (Mixed - requires --timeout)
Comprehensive file system module tests:
- **Sync file operations**: existsSync, statSync, readFileSync, writeFileSync, unlinkSync
- **Sync directory operations**: mkdirSync, readdirSync, rmdirSync
- **Async file operations**: readFile, writeFile, unlink
- **Async directory operations**: mkdir, rmdir
- **File stats**: isFile(), isDirectory()
- **Comprehensive cleanup**: Both sync and async file/directory removal
- **Execution**: Mixed sync/async, needs timeout for async operations

### test_path.js ⚡ (Sync - auto-exit)
Complete path module functionality:
- **Path parsing**: basename, dirname, extname
- **Path manipulation**: join, resolve, normalize
- **Path utilities**: isAbsolute, parse, format
- **Cross-platform compatibility**
- **Execution**: All synchronous operations, exits automatically

### test_require.js ⚡ (Sync - auto-exit)
Module system and require functionality:
- **Local modules**: Relative path resolution (./lib/math.js)
- **Node modules**: node_modules resolution (lodash)
- **Built-in modules**: console, process, fs modules
- **Module caching**: Multiple requires return same instance
- **Error handling**: Missing module detection
- **Execution**: All synchronous operations, exits automatically

### test_child_process.js ⚙️ (Process - requires --timeout)
Child process module testing with all major functions:
- **execSync**: Synchronous command execution with options and error handling
- **exec**: Asynchronous command execution with callbacks
- **spawn**: Basic process spawning functionality
- **Platform compatibility**: Cross-platform command testing (Windows/Unix)
- **Error handling**: Invalid command detection and proper error reporting
- **Options support**: Working directory, encoding, and timeout settings
- **Execution**: Process operations, needs ~15 seconds for reliability

### test_fetch.js 🌐 (Network - requires --timeout)
HTTP client testing with all methods:
- **GET**: JSON data retrieval and parsing
- **POST**: JSON data submission with request/response validation
- **PUT**: Data updates with timestamp verification
- **DELETE**: Resource deletion with authorization headers
- **Promise handling**: Proper response.json() usage
- **Error handling**: Network timeouts and HTTP errors
- **Execution**: Network operations, needs ~20 seconds for reliability

## Timeout Behavior

### Without --timeout (Synchronous Scripts)
- Scripts exit automatically after ~100ms
- Perfect for synchronous operations (path, require)
- No deadlock or hanging processes

### With --timeout (Asynchronous Scripts)  
- Scripts wait for async operations to complete
- Exit when all operations finish OR timeout is reached
- Required for timers, HTTP requests, file operations

## Expected Output

All tests should run without errors and display:
- ✅ for passing tests
- ❌ for failing tests  
- Test summaries with pass/fail counts
- Detailed error messages for debugging
- Clean exit codes (0 = success, 1 = failure)

## Performance Benchmarks

| Test | Command | Duration | Operations |
|------|---------|----------|------------|
| **Path** | `./rejs tests/test_path.js` | ~0.1s | 18 sync tests |
| **Require** | `./rejs tests/test_require.js` | ~0.1s | 22 sync tests |
| **Basic** | `./rejs --timeout=3000 tests/test_basic.js` | ~0.9s | Timers + console |
| **FS** | `./rejs --timeout=10000 tests/test_fs.js` | ~1-2s | 13 file operations |
| **Child Process** | `./rejs --timeout=15000 tests/test_child_process.js` | ~1-3s | 8 process tests |
| **Fetch** | `./rejs --timeout=20000 tests/test_fetch.js` | ~2-5s | 4 HTTP requests |

## Requirements

- **Internet connection** for fetch tests (uses httpbin.org)
- **File system permissions** for fs tests (creates/deletes test files)
- **Command execution permissions** for child_process tests (runs system commands)
- **No external dependencies** - all modules are built-in

## Troubleshooting

### Script hangs without output
- Add `--timeout=30000` for async scripts
- Check that `process.exit()` is called when operations complete

### "Script timeout" error
- Increase timeout value: `--timeout=60000`
- Check network connectivity for fetch tests
- Ensure file permissions for fs tests

### Tests fail randomly
- Network issues may cause fetch test failures
- Increase timeout for more reliable results
- Run tests individually to isolate issues