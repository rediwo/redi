// Test Buffer module in rejs runtime
const Buffer = require('buffer').Buffer;

console.log('Testing Buffer module...');

// Test 1: Create Buffer from string
const buf1 = Buffer.from('hello world', 'utf8');
console.log('Buffer from string:', buf1.toString());

// Test 2: Create Buffer from array
const buf2 = Buffer.from([0x68, 0x65, 0x6c, 0x6c, 0x6f]);
console.log('Buffer from array:', buf2.toString());

// Test 3: Buffer allocation
const buf3 = Buffer.alloc(10);
console.log('Buffer alloc length:', buf3.length);

// Test 4: Buffer write
const buf4 = Buffer.alloc(256);
const len = buf4.write('Hello Buffer!');
console.log('Buffer write:', buf4.toString('utf8', 0, len));

// Test 5: Buffer concat
const buf5 = Buffer.concat([buf1, Buffer.from(' ')]);
console.log('Buffer concat:', buf5.toString());

// Test 6: Buffer encoding
const base64 = Buffer.from('hello').toString('base64');
console.log('Base64 encoding:', base64);

const hex = Buffer.from('hello').toString('hex');
console.log('Hex encoding:', hex);

// Test 7: Global Buffer
try {
    const globalBuf = new Buffer(5);
    console.log('Global Buffer constructor works, length:', globalBuf.length);
} catch (e) {
    console.log('Global Buffer constructor error:', e.message);
}

console.log('Buffer module tests completed!');