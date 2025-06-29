// Fetch module comprehensive test (GET, POST, PUT, DELETE)
console.log("=== Fetch Module Test ===");

var fetch = require('fetch');

var tests = [];
var passed = 0;
var failed = 0;
var completed = 0;
var totalTests = 4; // GET, POST, PUT, DELETE

function addTest(name, testFn) {
    tests.push({ name: name, fn: testFn });
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
    if (completed === totalTests) {
        showSummary();
    }
}

function showSummary() {
    console.log("\n=== Fetch Test Summary ===");
    console.log("Total tests: " + totalTests);
    console.log("Passed: " + passed);
    console.log("Failed: " + failed);

    if (failed === 0) {
        console.log("✅ All Fetch tests passed!");
        process.exit(0);
    } else {
        console.log("❌ Some Fetch tests failed!");
        process.exit(1);
    }
}

// Test URLs (using httpbin.org for testing HTTP methods)
var baseUrl = 'https://httpbin.org';

console.log("Running Fetch tests (this may take a moment...):");

// 1. GET request test
console.log("\n1. Testing GET request:");
fetch(baseUrl + '/json')
    .then(function (response) {
        if (response.status === 200) {
            return response.json().then(function (data) {
                if (data && typeof data === 'object') {
                    completeTest("GET request", true, "status " + response.status + ", valid JSON");
                } else {
                    completeTest("GET request", false, "invalid JSON response");
                }
            });
        } else {
            completeTest("GET request", false, "status " + response.status);
        }
    })
    .catch(function (error) {
        completeTest("GET request", false, error.toString());
    });

// 2. POST request test
console.log("2. Testing POST request:");
var postData = {
    name: "rejs",
    version: "1.0.0",
    test: true
};

fetch(baseUrl + '/post', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json'
    },
    body: JSON.stringify(postData)
})
    .then(function (response) {
        if (response.status === 200) {
            return response.json().then(function (data) {
                if (data && data.json && data.json.name === "rejs") {
                    completeTest("POST request", true, "status " + response.status + ", data echoed correctly");
                } else {
                    completeTest("POST request", false, "data not echoed correctly");
                }
            });
        } else {
            completeTest("POST request", false, "status " + response.status);
        }
    })
    .catch(function (error) {
        completeTest("POST request", false, error.toString());
    });

// 3. PUT request test
console.log("3. Testing PUT request:");
var putData = {
    action: "update",
    timestamp: Date.now()
};

fetch(baseUrl + '/put', {
    method: 'PUT',
    headers: {
        'Content-Type': 'application/json'
    },
    body: JSON.stringify(putData)
})
    .then(function (response) {
        if (response.status === 200) {
            return response.json().then(function (data) {
                if (data && data.json && data.json.action === "update") {
                    completeTest("PUT request", true, "status " + response.status + ", data sent correctly");
                } else {
                    completeTest("PUT request", false, "data not processed correctly");
                }
            });
        } else {
            completeTest("PUT request", false, "status " + response.status);
        }
    })
    .catch(function (error) {
        completeTest("PUT request", false, error.toString());
    });

// 4. DELETE request test
console.log("4. Testing DELETE request:");
fetch(baseUrl + '/delete', {
    method: 'DELETE',
    headers: {
        'Authorization': 'Bearer test-token'
    }
})
    .then(function (response) {
        if (response.status === 200) {
            return response.json().then(function (data) {
                if (data && data.headers && data.headers.Authorization === 'Bearer test-token') {
                    completeTest("DELETE request", true, "status " + response.status + ", headers sent correctly");
                } else {
                    completeTest("DELETE request", false, "headers not sent correctly");
                }
            });
        } else {
            completeTest("DELETE request", false, "status " + response.status);
        }
    })
    .catch(function (error) {
        completeTest("DELETE request", false, error.toString());
    });

// Set a timeout in case requests hang
setTimeout(function () {
    if (completed < totalTests) {
        console.log("\n⚠️  Timeout: Not all requests completed");
        console.log("Completed: " + completed + "/" + totalTests);
        process.exit(1);
    }
}, 30000); // 30 second timeout

console.log("\nWaiting for HTTP requests to complete...");