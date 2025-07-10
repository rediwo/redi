// Debug stream module issue
const stream = require('stream');

console.log('=== Stream Debug Test ===');

// Test 1: Manual flow control
console.log('\n1. Testing manual readable stream:');
const { Readable } = stream;

const readable = new Readable({
    read() {
        console.log('  _read called');
        this.push('data1');
        this.push('data2');
        this.push(null);
    }
});

console.log('  Initial flowing state:', readable._readableState.flowing);

// Try manual read
console.log('  Calling read():');
let chunk;
while ((chunk = readable.read()) !== null) {
    console.log('  Read chunk:', chunk);
}

// Test 2: With data event
console.log('\n2. Testing with data event:');
const readable2 = new Readable({
    read() {
        console.log('  _read called in readable2');
        this.push('event-data1');
        this.push('event-data2');
        this.push(null);
    }
});

console.log('  Before adding listener, flowing:', readable2._readableState.flowing);

readable2.on('data', chunk => {
    console.log('  Data event:', chunk);
});

console.log('  After adding listener, flowing:', readable2._readableState.flowing);

// Manual resume if needed
if (!readable2._readableState.flowing) {
    console.log('  Manually calling resume()');
    readable2.resume();
}

// Test 3: Check EventEmitter
console.log('\n3. Testing EventEmitter inheritance:');
console.log('  readable.on is function?', typeof readable.on === 'function');
console.log('  readable.emit is function?', typeof readable.emit === 'function');

// Test 4: Simple writable
console.log('\n4. Testing simple writable:');
const { Writable } = stream;
const writable = new Writable({
    write(chunk, encoding, callback) {
        console.log('  Write:', chunk.toString());
        callback();
    }
});

writable.write('test');
writable.end();

console.log('\n=== Debug Test Complete ===');