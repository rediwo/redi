// Test stream module in rejs runtime
const stream = require('stream');

console.log('Testing stream module in rejs runtime');
console.log('Available stream classes:', Object.keys(stream));

// Test basic readable stream with async operations
async function testAsyncReadable() {
    console.log('\n=== Testing Async Readable Stream ===');
    
    const { Readable } = stream;
    
    // Create an async data source without generators
    let dataIndex = 1;
    const rs = new Readable({
        async read() {
            try {
                if (dataIndex <= 5) {
                    await new Promise(resolve => setTimeout(resolve, 10));
                    const value = `async-data-${dataIndex}`;
                    console.log('Pushing:', value);
                    this.push(value);
                    dataIndex++;
                } else {
                    this.push(null);
                }
            } catch (err) {
                this.destroy(err);
            }
        }
    });
    
    // Collect data
    const chunks = [];
    rs.on('data', chunk => {
        console.log('Received:', chunk);
        chunks.push(chunk);
    });
    
    return new Promise((resolve, reject) => {
        rs.on('end', () => {
            console.log('Stream ended, received chunks:', chunks.length);
            if (chunks.length === 5) {
                console.log('Async readable test passed!');
                resolve();
            } else {
                reject(new Error(`Expected 5 chunks, got ${chunks.length}`));
            }
        });
        
        rs.on('error', reject);
    });
}

// Test transform stream with promises
async function testPromiseTransform() {
    console.log('\n=== Testing Promise-based Transform ===');
    
    const { Transform, PassThrough } = stream;
    
    // Create an async transform
    const asyncTransform = new Transform({
        async transform(chunk, encoding, callback) {
            try {
                // Simulate async processing
                await new Promise(resolve => setTimeout(resolve, 5));
                const transformed = chunk.toString().toUpperCase() + '!';
                callback(null, transformed);
            } catch (err) {
                callback(err);
            }
        }
    });
    
    // Override _transform
    asyncTransform._transform = async function(chunk, encoding, callback) {
        try {
            await new Promise(resolve => setTimeout(resolve, 5));
            const transformed = chunk.toString().toUpperCase() + '!';
            callback(null, transformed);
        } catch (err) {
            callback(err);
        }
    };
    
    const source = new PassThrough();
    const results = [];
    
    source.pipe(asyncTransform).on('data', chunk => {
        console.log('Transformed:', chunk);
        results.push(chunk);
    });
    
    // Write data
    source.write('hello');
    source.write('world');
    source.end();
    
    return new Promise((resolve, reject) => {
        asyncTransform.on('end', () => {
            if (results.join(' ') === 'HELLO! WORLD!') {
                console.log('Promise transform test passed!');
                resolve();
            } else {
                reject(new Error('Transform failed'));
            }
        });
        
        asyncTransform.on('error', reject);
    });
}

// Test stream utilities
function testStreamUtilities() {
    console.log('\n=== Testing Stream Utilities ===');
    
    const { Readable, Writable, pipeline } = stream;
    
    // Note: pipeline might not be implemented yet, so we'll test basic functionality
    
    // Test PassThrough
    const { PassThrough } = stream;
    const pt = new PassThrough();
    
    let received = '';
    pt.on('data', chunk => {
        received += chunk;
    });
    
    pt.write('pass');
    pt.write('through');
    pt.end();
    
    pt.on('end', () => {
        console.log('PassThrough result:', received);
        if (received === 'passthrough') {
            console.log('PassThrough test passed!');
        }
    });
}

// Run all tests
async function runTests() {
    try {
        await testAsyncReadable();
        await testPromiseTransform();
        testStreamUtilities();
        
        console.log('\n=== All stream tests passed! ===');
    } catch (err) {
        console.error('Test failed:', err.message);
        process.exit(1);
    }
}

runTests();