// Basic stream module test
const stream = require('stream');

console.log('Testing basic stream functionality...');
console.log('Available stream classes:', Object.keys(stream).join(', '));

// Test 1: Basic Readable Stream (synchronous)
console.log('\n1. Testing Basic Readable Stream:');
const { Readable } = stream;

class MyReadable extends Readable {
    constructor(options) {
        super(options);
        this.current = 1;
    }
    
    _read() {
        if (this.current <= 3) {
            const chunk = `data-${this.current}`;
            console.log('  Pushing:', chunk);
            this.push(chunk);
            this.current++;
        } else {
            console.log('  Pushing null (end)');
            this.push(null);
        }
    }
}

const readable = new MyReadable();
const chunks = [];

readable.on('data', chunk => {
    console.log('  Received:', chunk.toString());
    chunks.push(chunk.toString());
});

readable.on('end', () => {
    console.log('  Stream ended. Total chunks:', chunks.length);
    console.log('  Data:', chunks.join(', '));
});

// Test 2: Basic Writable Stream
console.log('\n2. Testing Basic Writable Stream:');
const { Writable } = stream;

class MyWritable extends Writable {
    constructor(options) {
        super(options);
        this.data = [];
    }
    
    _write(chunk, encoding, callback) {
        const str = chunk.toString();
        console.log('  Writing:', str);
        this.data.push(str);
        callback();
    }
}

const writable = new MyWritable();
writable.write('hello');
writable.write(' ');
writable.write('world');
writable.end();

writable.on('finish', () => {
    console.log('  Write finished. Data:', writable.data.join(''));
});

// Test 3: Transform Stream
console.log('\n3. Testing Transform Stream:');
const { Transform } = stream;

class UppercaseTransform extends Transform {
    _transform(chunk, encoding, callback) {
        const transformed = chunk.toString().toUpperCase();
        console.log('  Transforming:', chunk.toString(), '->', transformed);
        callback(null, transformed);
    }
}

const transform = new UppercaseTransform();
const transformResults = [];

transform.on('data', chunk => {
    transformResults.push(chunk.toString());
});

transform.on('end', () => {
    console.log('  Transform result:', transformResults.join(''));
});

transform.write('hello');
transform.write(' world');
transform.end();

// Test 4: Pipe functionality
console.log('\n4. Testing Pipe:');
const source = new Readable({
    read() {
        this.push('piped data');
        this.push(null);
    }
});

const dest = new Writable({
    write(chunk, encoding, callback) {
        console.log('  Piped:', chunk.toString());
        callback();
    }
});

source.pipe(dest);

// Test 5: PassThrough
console.log('\n5. Testing PassThrough:');
const { PassThrough } = stream;
const passThrough = new PassThrough();

passThrough.on('data', chunk => {
    console.log('  PassThrough data:', chunk.toString());
});

passThrough.write('pass');
passThrough.write('through');
passThrough.end();

// Test 6: Stream events
console.log('\n6. Testing Stream Events:');
const eventStream = new Readable({
    read() {
        this.push('event test');
        this.push(null);
    }
});

eventStream.on('readable', () => {
    console.log('  Event: readable');
});

eventStream.on('data', chunk => {
    console.log('  Event: data -', chunk.toString());
});

eventStream.on('end', () => {
    console.log('  Event: end');
});

eventStream.on('close', () => {
    console.log('  Event: close');
});

// Allow some time for events to fire
setTimeout(() => {
    console.log('\n=== Basic Stream Tests Complete ===');
}, 100);