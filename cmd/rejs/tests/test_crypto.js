// Test crypto module in rejs runtime

const crypto = require('crypto');

console.log('=== Testing Crypto Module ===');

// Test 1: Hash functions
console.log('\n1. Testing hash functions:');
const hashes = ['md5', 'sha1', 'sha256', 'sha512'];
hashes.forEach(algo => {
    const hash = crypto.createHash(algo);
    hash.update('test data');
    const result = hash.digest('hex');
    console.log(`  ${algo}: ${result.substring(0, 16)}...`);
});

// Test 2: HMAC
console.log('\n2. Testing HMAC:');
const hmac = crypto.createHmac('sha256', 'my-secret-key');
hmac.update('important message');
console.log('  HMAC-SHA256:', hmac.digest('hex').substring(0, 32) + '...');

// Test 3: Random bytes
console.log('\n3. Testing random bytes:');
const randomSync = crypto.randomBytesSync(8);
console.log('  Sync random bytes (8):', randomSync);

// Test async random bytes
crypto.randomBytes(8, (err, buffer) => {
    if (err) {
        console.log('  Async random bytes error:', err);
    } else {
        console.log('  Async random bytes (8):', buffer);
    }
});

// Test 4: PBKDF2
console.log('\n4. Testing PBKDF2:');
const salt = crypto.randomBytesSync(16);
const keySync = crypto.pbkdf2Sync('password123', salt, 10000, 32, 'sha256');
console.log('  Sync PBKDF2 key length:', keySync.byteLength);

// Test async PBKDF2
crypto.pbkdf2('password123', salt, 10000, 32, 'sha256', (err, key) => {
    if (err) {
        console.log('  Async PBKDF2 error:', err);
    } else {
        console.log('  Async PBKDF2 key length:', key.byteLength);
    }
});

// Test 5: Available hashes
console.log('\n5. Available hash algorithms:');
console.log(' ', crypto.getHashes().join(', '));

// Test 6: Chaining
console.log('\n6. Testing update chaining:');
const chainHash = crypto.createHash('sha256')
    .update('Hello ')
    .update('World')
    .update('!')
    .digest('hex');
console.log('  Chained hash:', chainHash.substring(0, 32) + '...');

// Test 7: Different input types
console.log('\n7. Testing different input types:');
const hash1 = crypto.createHash('sha256');
hash1.update('string input');
console.log('  String input:', hash1.digest('hex').substring(0, 16) + '...');

// Test with ArrayBuffer
const buffer = new ArrayBuffer(8);
const view = new Uint8Array(buffer);
for (let i = 0; i < 8; i++) {
    view[i] = i;
}
const hash2 = crypto.createHash('sha256');
hash2.update(view);
console.log('  TypedArray input:', hash2.digest('hex').substring(0, 16) + '...');

// Test 8: Error handling
console.log('\n8. Testing error handling:');
try {
    crypto.createHash('invalid-algorithm');
} catch (e) {
    console.log('  Invalid algorithm error:', e.message);
}

try {
    crypto.createHmac('sha256'); // Missing key
} catch (e) {
    console.log('  Missing key error:', e.message);
}

console.log('\n=== Crypto Module Tests Complete ===');