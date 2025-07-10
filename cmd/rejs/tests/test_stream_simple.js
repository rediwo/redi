// Simple stream test
const stream = require('stream');

console.log('=== Simple Stream Test ===');

// Test 1: Most basic readable
console.log('\n1. Basic readable with immediate data:');
const readable = new stream.Readable({
    read() {
        console.log('  _read called');
        this.push('data');
        this.push(null); // End immediately
    }
});

readable.on('data', chunk => {
    console.log('  Got data:', chunk);
});

readable.on('end', () => {
    console.log('  Stream ended');
});

// Test 2: Basic writable
console.log('\n2. Basic writable:');
const writable = new stream.Writable({
    write(chunk, encoding, callback) {
        console.log('  Write:', chunk.toString());
        callback();
    }
});

writable.write('test');
writable.end(() => {
    console.log('  Write complete');
});

// Give time for async operations
setTimeout(() => {
    console.log('\n=== Test Complete ===');
    process.exit(0);
}, 1000);