// Test Transform stream specifically
const stream = require('stream');

console.log('=== Testing Transform Stream ===');

// Test 1: Simple synchronous transform
console.log('\n1. Synchronous transform:');
const upperCase = new stream.Transform({
    transform(chunk, encoding, callback) {
        console.log('  Transform called with:', chunk.toString());
        callback(null, chunk.toString().toUpperCase());
    }
});

upperCase.on('data', chunk => {
    console.log('  Output:', chunk.toString());
});

upperCase.on('end', () => {
    console.log('  Transform ended');
    testAsync();
});

upperCase.write('hello');
upperCase.write(' world');
upperCase.end();

// Test 2: Async transform
function testAsync() {
    console.log('\n2. Async transform:');
    const asyncTransform = new stream.Transform({
        transform(chunk, encoding, callback) {
            console.log('  Async transform called with:', chunk.toString());
            // Simulate async operation
            setTimeout(() => {
                console.log('  Async transform completing');
                callback(null, chunk.toString().toUpperCase() + '!');
            }, 10);
        }
    });
    
    const results = [];
    asyncTransform.on('data', chunk => {
        console.log('  Async output:', chunk.toString());
        results.push(chunk.toString());
    });
    
    asyncTransform.on('end', () => {
        console.log('  Async transform ended');
        console.log('  Results:', results);
        testPassThrough();
    });
    
    asyncTransform.write('async');
    asyncTransform.write(' test');
    asyncTransform.end();
}

// Test 3: PassThrough
function testPassThrough() {
    console.log('\n3. PassThrough stream:');
    const passThrough = new stream.PassThrough();
    
    passThrough.on('data', chunk => {
        console.log('  PassThrough data:', chunk.toString());
    });
    
    passThrough.on('end', () => {
        console.log('  PassThrough ended');
        console.log('\n=== Transform Test Complete ===');
        process.exit(0);
    });
    
    passThrough.write('pass');
    passThrough.write('through');
    passThrough.end();
}

// Safety timeout
setTimeout(() => {
    console.log('\n=== Test timed out ===');
    process.exit(1);
}, 5000);