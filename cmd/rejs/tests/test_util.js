// Test util module in rejs runtime
const util = require('util');

console.log('Testing util module...');

// Test 1: util.format with string
console.log('util.format string:', util.format('Hello %s', 'World'));

// Test 2: util.format with number
console.log('util.format number:', util.format('Number: %d', 42));

// Test 3: util.format with JSON
console.log('util.format JSON:', util.format('Object: %j', { name: 'test', value: 123 }));

// Test 4: util.format with percentage
console.log('util.format percentage:', util.format('Progress: 100%%'));

// Test 5: util.format with multiple arguments
console.log('util.format multiple:', util.format('Hello %s, you are %d years old', 'Alice', 25));

// Test 6: util.format with extra arguments
console.log('util.format extra args:', util.format('Hello %s', 'World', 'Extra', 'Args'));

// Test 7: util.format with no placeholders
console.log('util.format no placeholders:', util.format('No placeholders', 'arg1', 'arg2'));

// Test 8: Check for other util methods
if (util.inspect) {
    console.log('util.inspect available:', typeof util.inspect);
    const obj = { a: 1, b: 'test', nested: { c: true } };
    console.log('util.inspect object:', util.inspect(obj));
}

if (util.inherits) {
    console.log('util.inherits available:', typeof util.inherits);
}

if (util.promisify) {
    console.log('util.promisify available:', typeof util.promisify);
}

console.log('util module tests completed!');