// Test Node.js core modules integration in rejs runtime

console.log('=== Testing Node.js Core Modules Integration ===\n');

// Test 1: Buffer module
console.log('1. Testing Buffer module:');
const Buffer = require('buffer').Buffer;
const buf = Buffer.from('Hello World', 'utf8');
console.log('  Buffer content:', buf.toString());
console.log('  Buffer length:', buf.length);
console.log('  Buffer hex:', buf.toString('hex').substring(0, 20) + '...');

// Test global Buffer
const globalBuf = new Buffer('Global Buffer Test');
console.log('  Global Buffer:', globalBuf.toString());

// Test 2: Util module
console.log('\n2. Testing Util module:');
const util = require('util');
console.log('  util.format:', util.format('Hello %s, number: %d', 'World', 42));
console.log('  util.format JSON:', util.format('Data: %j', { test: true, value: 123 }));

// Test 3: URL module
console.log('\n3. Testing URL module:');
const url = require('url');
const myURL = new url.URL('https://example.com:8080/path?key=value#section');
console.log('  URL parts:', {
    hostname: myURL.hostname,
    port: myURL.port,
    pathname: myURL.pathname,
    search: myURL.search
});

// Test global URL
const globalURL = new URL('https://test.global.com/test');
console.log('  Global URL:', globalURL.href);

// Test 4: Combined usage
console.log('\n4. Testing combined usage:');
// Create a URL with query parameters
const apiURL = new URL('https://api.example.com/data');
apiURL.searchParams.append('format', 'json');
apiURL.searchParams.append('limit', '10');

// Format the URL info with util
const urlInfo = util.format('API URL: %s with params: %j', 
    apiURL.hostname, 
    Object.fromEntries(apiURL.searchParams)
);
console.log('  ' + urlInfo);

// Create a buffer from URL string
const urlBuffer = Buffer.from(apiURL.href);
console.log('  URL as Buffer length:', urlBuffer.length);
console.log('  URL Buffer (first 30 chars):', urlBuffer.toString().substring(0, 30) + '...');

// Test 5: Error handling
console.log('\n5. Testing error handling:');
try {
    new URL('invalid-url');
} catch (e) {
    console.log('  URL error caught:', e.message);
}

// Test 6: Check other core modules availability
console.log('\n6. Checking other core modules:');
const moduleList = ['fs', 'path', 'crypto', 'stream', 'process'];
moduleList.forEach(moduleName => {
    try {
        const mod = require(moduleName);
        console.log(`  ${moduleName}: ✓ available`);
    } catch (e) {
        console.log(`  ${moduleName}: ✗ not available`);
    }
});

console.log('\n=== All Node.js Core Modules Tests Completed ===');