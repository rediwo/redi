// File system module comprehensive test
console.log("=== FS Module Test ===");

var fs = require('fs');
var path = require('path');

// Test files
var testDir = path.join(__dirname, 'test_temp');
var testFile = path.join(testDir, 'test.txt');
var testContent = 'Hello, rejs filesystem!\nLine 2\nLine 3';

var tests = [];
var passed = 0;
var failed = 0;

function addTest(name, testFn) {
    tests.push({ name: name, fn: testFn });
}

function runTest(name, testFn) {
    try {
        var result = testFn();
        if (result !== false) {
            console.log("  ✅ " + name);
            passed++;
        } else {
            console.log("  ❌ " + name + " - test returned false");
            failed++;
        }
    } catch (e) {
        console.log("  ❌ " + name + " - " + e.message);
        failed++;
    }
}

// 1. existsSync tests
addTest("existsSync - file exists", function() {
    return fs.existsSync(__filename) === true;
});

addTest("existsSync - file doesn't exist", function() {
    return fs.existsSync('/nonexistent/file.txt') === false;
});

// 2. statSync tests  
addTest("statSync - get file stats", function() {
    var stats = fs.statSync(__filename);
    return stats && typeof stats.isFile === 'function' && stats.isFile() === true;
});

addTest("statSync - isDirectory", function() {
    var stats = fs.statSync(__dirname);
    return stats && stats.isDirectory() === true;
});

// 3. readFileSync tests
addTest("readFileSync - read this file", function() {
    var content = fs.readFileSync(__filename, 'utf8');
    return content && content.indexOf('=== FS Module Test ===') !== -1;
});

// 4. writeFileSync and mkdirSync tests
addTest("mkdirSync - create directory", function() {
    if (fs.existsSync(testDir)) {
        // Clean up first - just remove file if it exists
        if (fs.existsSync(testFile)) {
            fs.unlinkSync(testFile);
        }
        // Skip directory removal since rmdirSync is not available
    }
    fs.mkdirSync(testDir);
    return fs.existsSync(testDir);
});

addTest("writeFileSync - write file", function() {
    fs.writeFileSync(testFile, testContent, 'utf8');
    return fs.existsSync(testFile);
});

addTest("readFileSync - read written file", function() {
    var content = fs.readFileSync(testFile, 'utf8');
    return content === testContent;
});

// 5. readdirSync tests
addTest("readdirSync - read directory", function() {
    var files = fs.readdirSync(testDir);
    return Array.isArray(files) && files.indexOf('test.txt') !== -1;
});

// 6. unlinkSync and rmdirSync tests
addTest("unlinkSync - delete file", function() {
    fs.unlinkSync(testFile);
    return !fs.existsSync(testFile);
});

addTest("rmdirSync - remove directory", function() {
    fs.rmdirSync(testDir);
    return !fs.existsSync(testDir);
});

// Run all tests
console.log("\nRunning FS tests:");
for (var i = 0; i < tests.length; i++) {
    runTest(tests[i].name, tests[i].fn);
}

// Async tests
var asyncTests = [];
var asyncPassed = 0;
var asyncFailed = 0;
var asyncCompleted = 0;

function addAsyncTest(name, testFn) {
    asyncTests.push({ name: name, fn: testFn });
}

function completeAsyncTest(name, success, message) {
    asyncCompleted++;
    if (success) {
        console.log("  ✅ " + name + (message ? " - " + message : ""));
        asyncPassed++;
    } else {
        console.log("  ❌ " + name + " - " + (message || "failed"));
        asyncFailed++;
    }
    
    // Check if all async tests are complete
    if (asyncCompleted === asyncTests.length) {
        showFinalSummary();
    }
}

function showFinalSummary() {
    // Check for leftover directories
    console.log("\n=== Cleanup Verification ===");
    var leftoverDirs = [testDir, asyncTestDir];
    var cleanupIssues = 0;
    
    leftoverDirs.forEach(function(dir) {
        if (fs.existsSync(dir)) {
            console.log("⚠️  Leftover directory: " + path.basename(dir));
            cleanupIssues++;
            try {
                // Read all files in directory and remove them first
                var files = fs.readdirSync(dir);
                files.forEach(function(file) {
                    var filePath = path.join(dir, file);
                    var stats = fs.statSync(filePath);
                    if (stats.isFile()) {
                        fs.unlinkSync(filePath);
                        console.log("   🗑️ Removed file: " + file);
                    }
                });
                
                // Then remove the empty directory
                fs.rmdirSync(dir);
                console.log("   ✅ Cleaned up: " + path.basename(dir));
            } catch (e) {
                console.log("   ❌ Failed to clean: " + e.message);
            }
        }
    });
    
    if (cleanupIssues === 0) {
        console.log("✅ No leftover directories");
    }
    
    console.log("\n=== FS Test Summary ===");
    console.log("Sync tests: " + passed + "/" + tests.length + " passed");
    console.log("Async tests: " + asyncPassed + "/" + asyncTests.length + " passed");
    console.log("Total: " + (passed + asyncPassed) + "/" + (tests.length + asyncTests.length) + " passed");
    
    if (failed === 0 && asyncFailed === 0) {
        console.log("✅ All FS tests passed!");
        process.exit(0);
    } else {
        console.log("❌ Some FS tests failed!");
        process.exit(1);
    }
}

// Async test setup
var asyncTestDir = path.join(__dirname, 'async_test_temp');
var asyncTestFile = path.join(asyncTestDir, 'async_test.txt');
var asyncTestContent = 'Async test content';

// 1. Async readFile test
addAsyncTest("readFile - read this file", function() {
    fs.readFile(__filename, 'utf8', function(err, data) {
        if (err) {
            completeAsyncTest("readFile - read this file", false, err);
        } else {
            var success = data && data.indexOf('=== FS Module Test ===') !== -1;
            completeAsyncTest("readFile - read this file", success);
        }
    });
});

// 2. Async mkdir + writeFile + readFile + unlink + rmdir test
addAsyncTest("mkdir - create directory async", function() {
    fs.mkdir(asyncTestDir, function(err) {
        if (err) {
            completeAsyncTest("mkdir - create directory async", false, err);
        } else {
            completeAsyncTest("mkdir - create directory async", true);
            
            // Chain next test
            fs.writeFile(asyncTestFile, asyncTestContent, 'utf8', function(err) {
                if (err) {
                    completeAsyncTest("writeFile - write file async", false, err);
                } else {
                    completeAsyncTest("writeFile - write file async", true);
                    
                    // Chain next test
                    fs.readFile(asyncTestFile, 'utf8', function(err, data) {
                        if (err) {
                            completeAsyncTest("readFile - read written file async", false, err);
                        } else {
                            var success = data === asyncTestContent;
                            completeAsyncTest("readFile - read written file async", success);
                            
                            // Chain cleanup
                            fs.unlink(asyncTestFile, function(err) {
                                if (err) {
                                    completeAsyncTest("unlink - delete file async", false, err);
                                } else {
                                    completeAsyncTest("unlink - delete file async", true);
                                    
                                    // Final cleanup
                                    fs.rmdir(asyncTestDir, function(err) {
                                        if (err) {
                                            completeAsyncTest("rmdir - remove directory async", false, err);
                                        } else {
                                            completeAsyncTest("rmdir - remove directory async", true);
                                        }
                                    });
                                }
                            });
                        }
                    });
                }
            });
        }
    });
});

// Wait for sync tests to complete first
setTimeout(function() {
    if (failed === 0) {
        console.log("\n=== Running Async FS Tests ===");
        // Start async tests
        for (var i = 0; i < asyncTests.length; i++) {
            asyncTests[i].fn();
        }
        
        // Set timeout for async tests
        setTimeout(function() {
            if (asyncCompleted < asyncTests.length) {
                console.log("\n⚠️  Timeout: Not all async tests completed");
                console.log("Completed: " + asyncCompleted + "/" + asyncTests.length);
                process.exit(1);
            }
        }, 5000);
    } else {
        console.log("\n=== FS Test Summary ===");
        console.log("Sync tests: " + passed + "/" + tests.length + " passed");
        console.log("❌ Skipping async tests due to sync test failures!");
        process.exit(1);
    }
}, 100);