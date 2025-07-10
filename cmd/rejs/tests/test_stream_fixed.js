// Test fixed stream module
const stream = require('stream');

console.log('=== Stream Module Fixed Test ===');

// Test 1: Readable with auto-flow on 'data' event
console.log('\n1. Readable auto-flow test:');
const readable = new stream.Readable({
    read() {
        this.push('auto-flow-data');
        this.push(null);
    }
});

readable.on('data', chunk => {
    console.log('  Auto-flow data:', chunk);
});

readable.on('end', () => {
    console.log('  Auto-flow ended');
    test2();
});

// Test 2: Async readable
function test2() {
    console.log('\n2. Async readable test:');
    let count = 0;
    const asyncReadable = new stream.Readable({
        read() {
            if (count < 3) {
                setTimeout(() => {
                    this.push(`async-${++count}`);
                }, 10);
            } else {
                this.push(null);
            }
        }
    });
    
    asyncReadable.on('data', chunk => {
        console.log('  Async data:', chunk);
    });
    
    asyncReadable.on('end', () => {
        console.log('  Async ended');
        test3();
    });
}

// Test 3: Transform
function test3() {
    console.log('\n3. Transform test:');
    const transform = new stream.Transform({
        transform(chunk, encoding, callback) {
            callback(null, chunk.toString().toUpperCase());
        }
    });
    
    transform.on('data', chunk => {
        console.log('  Transformed:', chunk);
    });
    
    transform.on('end', () => {
        console.log('  Transform ended');
        test4();
    });
    
    transform.write('transform');
    transform.end();
}

// Test 4: Pipe
function test4() {
    console.log('\n4. Pipe test:');
    const source = new stream.Readable({
        read() {
            this.push('piped');
            this.push(null);
        }
    });
    
    const dest = new stream.Writable({
        write(chunk, encoding, callback) {
            console.log('  Piped data:', chunk.toString());
            callback();
        }
    });
    
    dest.on('finish', () => {
        console.log('  Pipe finished');
        console.log('\n=== All Fixed Tests Passed ===');
        process.exit(0);
    });
    
    source.pipe(dest);
}

// Timeout
setTimeout(() => {
    console.log('Test timeout');
    process.exit(1);
}, 3000);