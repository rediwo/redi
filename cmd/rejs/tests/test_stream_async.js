// Test async stream operations
const stream = require('stream');

console.log('=== Testing Async Stream Operations ===');

// Test 1: Simple async readable
console.log('\n1. Simple async readable:');
let count = 0;
const readable = new stream.Readable({
    read() {
        console.log('  _read called, count:', count);
        if (count < 3) {
            setTimeout(() => {
                count++;
                console.log('  Pushing data', count);
                this.push(`data-${count}`);
            }, 10);
        } else if (count === 3) {
            count++;
            console.log('  Pushing null');
            this.push(null);
        }
    }
});

const chunks = [];
readable.on('data', chunk => {
    console.log('  Received:', chunk);
    chunks.push(chunk);
});

readable.on('end', () => {
    console.log('  Stream ended, total chunks:', chunks.length);
    console.log('  Chunks:', chunks);
    
    // Test 2: Transform stream
    console.log('\n2. Transform stream:');
    testTransform();
});

function testTransform() {
    const transform = new stream.Transform({
        transform(chunk, encoding, callback) {
            console.log('  Transform input:', chunk.toString());
            // Simulate async transform
            setTimeout(() => {
                callback(null, chunk.toString().toUpperCase());
            }, 10);
        }
    });
    
    transform.on('data', chunk => {
        console.log('  Transform output:', chunk);
    });
    
    transform.on('end', () => {
        console.log('  Transform ended');
        console.log('\n=== Async Test Complete ===');
        process.exit(0);
    });
    
    transform.write('hello');
    transform.write(' world');
    transform.end();
}

// Safety timeout
setTimeout(() => {
    console.log('\n=== Test timed out ===');
    process.exit(1);
}, 5000);