// Math utility library
var math = {};

// Basic operations
math.add = function(a, b) {
    return a + b;
};

math.subtract = function(a, b) {
    return a - b;
};

math.multiply = function(a, b) {
    return a * b;
};

math.divide = function(a, b) {
    if (b === 0) {
        throw new Error('Division by zero');
    }
    return a / b;
};

// Advanced operations
math.power = function(base, exponent) {
    return Math.pow(base, exponent);
};

math.sqrt = function(num) {
    if (num < 0) {
        throw new Error('Square root of negative number');
    }
    return Math.sqrt(num);
};

math.factorial = function(n) {
    if (n < 0) {
        throw new Error('Factorial of negative number');
    }
    if (n === 0 || n === 1) {
        return 1;
    }
    var result = 1;
    for (var i = 2; i <= n; i++) {
        result *= i;
    }
    return result;
};

// Array operations
math.sum = function(numbers) {
    if (!Array.isArray(numbers)) {
        throw new Error('Input must be an array');
    }
    var total = 0;
    for (var i = 0; i < numbers.length; i++) {
        if (typeof numbers[i] !== 'number') {
            throw new Error('All elements must be numbers');
        }
        total += numbers[i];
    }
    return total;
};

math.average = function(numbers) {
    if (!Array.isArray(numbers) || numbers.length === 0) {
        throw new Error('Input must be a non-empty array');
    }
    return math.sum(numbers) / numbers.length;
};

math.max = function(numbers) {
    if (!Array.isArray(numbers) || numbers.length === 0) {
        throw new Error('Input must be a non-empty array');
    }
    var max = numbers[0];
    for (var i = 1; i < numbers.length; i++) {
        if (numbers[i] > max) {
            max = numbers[i];
        }
    }
    return max;
};

math.min = function(numbers) {
    if (!Array.isArray(numbers) || numbers.length === 0) {
        throw new Error('Input must be a non-empty array');
    }
    var min = numbers[0];
    for (var i = 1; i < numbers.length; i++) {
        if (numbers[i] < min) {
            min = numbers[i];
        }
    }
    return min;
};

// Constants
math.PI = Math.PI;
math.E = Math.E;

// Trigonometric functions
math.sin = function(x) {
    return Math.sin(x);
};

math.cos = function(x) {
    return Math.cos(x);
};

math.tan = function(x) {
    return Math.tan(x);
};

// Random utilities
math.random = function(min, max) {
    if (min === undefined && max === undefined) {
        return Math.random();
    }
    if (max === undefined) {
        max = min;
        min = 0;
    }
    return Math.random() * (max - min) + min;
};

math.randomInt = function(min, max) {
    return Math.floor(math.random(min, max));
};

module.exports = math;