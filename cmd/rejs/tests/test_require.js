// Require system comprehensive test
console.log("=== Require System Test ===");

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

// Test 1: Require local lib module (relative path)
addTest("require local lib/math.js", function() {
    var math = require('./lib/math.js');
    return math && typeof math.add === 'function';
});

// Test 2: Test math module functions
addTest("math.add function", function() {
    var math = require('./lib/math.js');
    return math.add(2, 3) === 5;
});

addTest("math.multiply function", function() {
    var math = require('./lib/math.js');
    return math.multiply(4, 5) === 20;
});

addTest("math.factorial function", function() {
    var math = require('./lib/math.js');
    return math.factorial(5) === 120;
});

addTest("math.sum array function", function() {
    var math = require('./lib/math.js');
    return math.sum([1, 2, 3, 4, 5]) === 15;
});

addTest("math.average function", function() {
    var math = require('./lib/math.js');
    return math.average([2, 4, 6, 8]) === 5;
});

// Test 3: Require node_modules (lodash)
addTest("require node_modules lodash", function() {
    var _ = require('lodash');
    return _ && typeof _.map === 'function';
});

// Test 4: Test lodash functions
addTest("lodash.map function", function() {
    var _ = require('lodash');
    var result = _.map([1, 2, 3], function(n) { return n * 2; });
    return Array.isArray(result) && result.length === 3 && result[0] === 2 && result[1] === 4 && result[2] === 6;
});

addTest("lodash.filter function", function() {
    var _ = require('lodash');
    var result = _.filter([1, 2, 3, 4, 5], function(n) { return n % 2 === 0; });
    return Array.isArray(result) && result.length === 2 && result[0] === 2 && result[1] === 4;
});

addTest("lodash.find function", function() {
    var _ = require('lodash');
    var users = [
        { name: 'John', age: 25 },
        { name: 'Jane', age: 30 },
        { name: 'Bob', age: 35 }
    ];
    var result = _.find(users, function(user) { return user.age === 30; });
    return result && result.name === 'Jane';
});

addTest("lodash.reduce function", function() {
    var _ = require('lodash');
    var result = _.reduce([1, 2, 3, 4], function(sum, n) { return sum + n; }, 0);
    return result === 10;
});

addTest("lodash.keys function", function() {
    var _ = require('lodash');
    var obj = { a: 1, b: 2, c: 3 };
    var keys = _.keys(obj);
    return Array.isArray(keys) && keys.length === 3;
});

addTest("lodash.capitalize function", function() {
    var _ = require('lodash');
    return _.capitalize('hello world') === 'Hello world';
});

addTest("lodash.camelCase function", function() {
    var _ = require('lodash');
    return _.camelCase('hello-world_test') === 'helloWorldTest';
});

addTest("lodash.isArray function", function() {
    var _ = require('lodash');
    return _.isArray([1, 2, 3]) === true && _.isArray('hello') === false;
});

// Test 5: Multiple requires of same module (should return same instance)
addTest("multiple requires return same instance", function() {
    var math1 = require('./lib/math.js');
    var math2 = require('./lib/math.js');
    return math1 === math2;
});

addTest("multiple lodash requires return same instance", function() {
    var _1 = require('lodash');
    var _2 = require('lodash');
    return _1 === _2;
});

// Test 6: Built-in modules still work
addTest("require built-in console module", function() {
    var console_module = require('console');
    return console_module && typeof console_module.log === 'function';
});

addTest("require built-in process module", function() {
    var process_module = require('process');
    return process_module && typeof process_module.platform === 'string';
});

addTest("require built-in fs module", function() {
    var fs = require('fs');
    return fs && typeof fs.readFileSync === 'function';
});

// Test 7: Error handling
addTest("require non-existent local file throws error", function() {
    try {
        require('./non-existent-file.js');
        return false; // Should not reach here
    } catch (e) {
        return true; // Expected error
    }
});

addTest("require non-existent node_module throws error", function() {
    try {
        require('non-existent-module');
        return false; // Should not reach here
    } catch (e) {
        return true; // Expected error
    }
});

// Run all tests
console.log("\nRunning require system tests:");
for (var i = 0; i < tests.length; i++) {
    runTest(tests[i].name, tests[i].fn);
}

// Summary
console.log("\n=== Require Test Summary ===");
console.log("Tests: " + passed + "/" + tests.length + " passed");

if (failed === 0) {
    console.log("✅ All require tests passed!");
    process.exit(0);
} else {
    console.log("❌ Some require tests failed!");
    process.exit(1);
}