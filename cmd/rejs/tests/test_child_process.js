// Child process module comprehensive test
console.log("=== Child Process Module Test ===");

var child_process = require('child_process');

var tests = [];
var passed = 0;
var failed = 0;
var completed = 0;

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

function completeTest(name, success, message) {
    completed++;
    if (success) {
        console.log("  ✅ " + name + (message ? " - " + message : ""));
        passed++;
    } else {
        console.log("  ❌ " + name + " - " + (message || "failed"));
        failed++;
    }
    
    // Check if all tests are complete
    if (completed === tests.length) {
        showSummary();
    }
}

function showSummary() {
    console.log("\n=== Child Process Test Summary ===");
    console.log("Total tests: " + tests.length);
    console.log("Passed: " + passed);
    console.log("Failed: " + failed);

    if (failed === 0) {
        console.log("✅ All Child Process tests passed!");
        process.exit(0);
    } else {
        console.log("❌ Some Child Process tests failed!");
        process.exit(1);
    }
}

// 1. execSync tests
addTest("execSync - simple command", function() {
    var output = child_process.execSync('echo "Hello from execSync"');
    return output && output.toString().indexOf("Hello from execSync") !== -1;
});

addTest("execSync - with options", function() {
    try {
        var output = child_process.execSync('echo "Working directory test"', {
            encoding: 'utf8'
        });
        return output && output.indexOf("Working directory test") !== -1;
    } catch (e) {
        return false;
    }
});

addTest("execSync - error handling", function() {
    try {
        child_process.execSync('invalidcommandthatdoesnotexist');
        return false; // Should have thrown an error
    } catch (e) {
        return true; // Expected error
    }
});

// 2. Platform-specific tests
var isWindows = process.platform === 'win32';
var listCommand = isWindows ? 'dir' : 'ls';
var echoCommand = 'echo "test output"';

addTest("execSync - platform command (" + listCommand + ")", function() {
    try {
        var output = child_process.execSync(listCommand);
        return output && output.length > 0;
    } catch (e) {
        console.log("      Platform command error: " + e.message);
        return false;
    }
});

// Run all synchronous tests first
console.log("\nRunning Child Process synchronous tests:");
for (var i = 0; i < tests.length; i++) {
    runTest(tests[i].name, tests[i].fn);
}

// 3. Async tests (exec)
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
        showAsyncSummary();
    }
}

function showAsyncSummary() {
    console.log("\n=== Child Process Test Summary ===");
    console.log("Sync tests: " + passed + "/" + tests.length + " passed");
    console.log("Async tests: " + asyncPassed + "/" + asyncTests.length + " passed");
    console.log("Total: " + (passed + asyncPassed) + "/" + (tests.length + asyncTests.length) + " passed");
    
    if (failed === 0 && asyncFailed === 0) {
        console.log("✅ All Child Process tests passed!");
        process.exit(0);
    } else {
        console.log("❌ Some Child Process tests failed!");
        process.exit(1);
    }
}

// Async exec tests
addAsyncTest("exec - simple command", function() {
    child_process.exec('echo "Hello from exec"', function(error, stdout, stderr) {
        if (error) {
            completeAsyncTest("exec - simple command", false, error.message);
        } else {
            var success = stdout && stdout.indexOf("Hello from exec") !== -1;
            completeAsyncTest("exec - simple command", success, success ? "output correct" : "unexpected output");
        }
    });
});

addAsyncTest("exec - with options", function() {
    child_process.exec(echoCommand, { encoding: 'utf8' }, function(error, stdout, stderr) {
        if (error) {
            completeAsyncTest("exec - with options", false, error.message);
        } else {
            var success = stdout && stdout.indexOf("test output") !== -1;
            completeAsyncTest("exec - with options", success, success ? "options handled" : "options failed");
        }
    });
});

addAsyncTest("exec - error command", function() {
    child_process.exec('invalidcommandthatdoesnotexist', function(error, stdout, stderr) {
        if (error) {
            completeAsyncTest("exec - error command", true, "error properly caught");
        } else {
            completeAsyncTest("exec - error command", false, "should have failed");
        }
    });
});

// Simple spawn test (spawn is more complex, so we'll test basic functionality)
addAsyncTest("spawn - basic test", function() {
    try {
        var child = child_process.spawn('echo', ['Hello from spawn']);
        if (child && typeof child.pid !== 'undefined') {
            completeAsyncTest("spawn - basic test", true, "child process created");
        } else {
            completeAsyncTest("spawn - basic test", false, "child process not created properly");
        }
    } catch (e) {
        completeAsyncTest("spawn - basic test", false, e.message);
    }
});

// Wait for sync tests to complete, then run async tests
setTimeout(function() {
    if (failed === 0) {
        console.log("\n=== Running Child Process Async Tests ===");
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
        }, 10000);
    } else {
        console.log("\n=== Child Process Test Summary ===");
        console.log("Sync tests: " + passed + "/" + tests.length + " passed");
        console.log("❌ Skipping async tests due to sync test failures!");
        process.exit(1);
    }
}, 100);