// Test URL module in rejs runtime
const url = require('url');

console.log('Testing URL module...');

// Test 1: URL constructor
const myURL = new url.URL('https://example.com:8080/path?query=123#hash');
console.log('URL href:', myURL.href);
console.log('URL protocol:', myURL.protocol);
console.log('URL hostname:', myURL.hostname);
console.log('URL port:', myURL.port);
console.log('URL pathname:', myURL.pathname);
console.log('URL search:', myURL.search);
console.log('URL hash:', myURL.hash);

// Test 2: URL with base
const relativeURL = new url.URL('/foo', 'https://example.com/bar');
console.log('URL with base:', relativeURL.href);

// Test 3: URLSearchParams
const params = new url.URLSearchParams('foo=bar&baz=qux');
console.log('URLSearchParams get foo:', params.get('foo'));
console.log('URLSearchParams get baz:', params.get('baz'));

// Test 4: URLSearchParams methods
params.set('foo', 'updated');
params.append('new', 'value');
console.log('After set/append:', params.toString());

// Test 5: URL searchParams
const urlWithParams = new url.URL('https://example.com?a=1&b=2');
console.log('URL searchParams a:', urlWithParams.searchParams.get('a'));
console.log('URL searchParams b:', urlWithParams.searchParams.get('b'));

// Test 6: URL modification
myURL.pathname = '/newpath';
myURL.search = '?new=query';
console.log('Modified URL:', myURL.href);

// Test 7: Global constructors
try {
    const globalURL = new URL('https://test.com');
    console.log('Global URL works:', globalURL.hostname);
} catch (e) {
    console.log('Global URL error:', e.message);
}

try {
    const globalParams = new URLSearchParams('test=value');
    console.log('Global URLSearchParams works:', globalParams.get('test'));
} catch (e) {
    console.log('Global URLSearchParams error:', e.message);
}

// Test 8: Domain functions
if (url.domainToASCII) {
    console.log('domainToASCII:', url.domainToASCII('example.com'));
}

if (url.domainToUnicode) {
    console.log('domainToUnicode:', url.domainToUnicode('example.com'));
}

console.log('URL module tests completed!');