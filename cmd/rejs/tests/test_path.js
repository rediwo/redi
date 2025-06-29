// Path module comprehensive test
console.log("=== Path Module Test ===");

var path = require('path');

var tests = [];
var passed = 0;
var failed = 0;

function addTest(name, expected, actual) {
    tests.push({ name: name, expected: expected, actual: actual });
}

function runTests() {
    console.log("\nRunning Path tests:");
    
    for (var i = 0; i < tests.length; i++) {
        var test = tests[i];
        try {
            var actualResult = typeof test.actual === 'function' ? test.actual() : test.actual;
            var expectedResult = test.expected;
            
            if (actualResult === expectedResult) {
                console.log("  ✅ " + test.name);
                console.log("      Result: '" + actualResult + "'");
                passed++;
            } else {
                console.log("  ❌ " + test.name);
                console.log("      Expected: '" + expectedResult + "'");
                console.log("      Actual: '" + actualResult + "'");
                failed++;
            }
        } catch (e) {
            console.log("  ❌ " + test.name + " - " + e.message);
            failed++;
        }
    }
}

// Test data
var testPaths = {
    unix: '/home/user/documents/file.txt',
    unixDir: '/home/user/documents/',
    relative: '../docs/readme.md',
    current: './script.js',
    windows: 'C:\\Users\\user\\file.txt',
    noExt: '/path/to/file'
};

// 1. path.basename tests
console.log("1. Testing path.basename:");
addTest("basename - unix file", "file.txt", function() {
    return path.basename(testPaths.unix);
});

addTest("basename - with extension", "file", function() {
    return path.basename(testPaths.unix, '.txt');
});

addTest("basename - no extension", "file", function() {
    return path.basename(testPaths.noExt);
});

// 2. path.dirname tests
console.log("2. Testing path.dirname:");
addTest("dirname - unix path", "/home/user/documents", function() {
    return path.dirname(testPaths.unix);
});

addTest("dirname - relative path", "../docs", function() {
    return path.dirname(testPaths.relative);
});

// 3. path.extname tests
console.log("3. Testing path.extname:");
addTest("extname - .txt file", ".txt", function() {
    return path.extname(testPaths.unix);
});

addTest("extname - .md file", ".md", function() {
    return path.extname(testPaths.relative);
});

addTest("extname - no extension", "", function() {
    return path.extname(testPaths.noExt);
});

// 4. path.join tests
console.log("4. Testing path.join:");
addTest("join - multiple parts", "/home/user/documents/file.txt", function() {
    return path.join('/home', 'user', 'documents', 'file.txt');
});

addTest("join - with relative parts", "/home/user/file.txt", function() {
    return path.join('/home', 'user', '../user', 'file.txt');
});

addTest("join - current directory", "docs/file.txt", function() {
    return path.join('.', 'docs', 'file.txt');
});

// 5. path.resolve tests
console.log("5. Testing path.resolve:");
addTest("resolve - absolute path", testPaths.unix, function() {
    return path.resolve(testPaths.unix);
});

// 6. path.normalize tests
console.log("6. Testing path.normalize:");
addTest("normalize - redundant separators", "/home/user/docs", function() {
    return path.normalize('/home//user/./docs');
});

addTest("normalize - parent directory", "/home/docs", function() {
    return path.normalize('/home/user/../docs');
});

// 7. path.isAbsolute tests
console.log("7. Testing path.isAbsolute:");
addTest("isAbsolute - absolute path", true, function() {
    return path.isAbsolute(testPaths.unix);
});

addTest("isAbsolute - relative path", false, function() {
    return path.isAbsolute(testPaths.relative);
});

// 8. path.parse tests
console.log("8. Testing path.parse:");
addTest("parse - has all properties", true, function() {
    var parsed = path.parse(testPaths.unix);
    return parsed && 
           typeof parsed.root === 'string' &&
           typeof parsed.dir === 'string' &&
           typeof parsed.base === 'string' &&
           typeof parsed.name === 'string' &&
           typeof parsed.ext === 'string';
});

// 9. path.format tests
console.log("9. Testing path.format:");
addTest("format - reconstruct path", testPaths.unix, function() {
    var parsed = path.parse(testPaths.unix);
    return path.format(parsed);
});

// Run all tests
runTests();

// Summary
console.log("\n=== Path Test Summary ===");
console.log("Total tests: " + tests.length);
console.log("Passed: " + passed);
console.log("Failed: " + failed);

if (failed === 0) {
    console.log("✅ All Path tests passed!");
    process.exit(0);
} else {
    console.log("❌ Some Path tests failed!");
    process.exit(1);
}