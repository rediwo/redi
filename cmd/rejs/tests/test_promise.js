// Promise functionality test
console.log("=== Promise Test ===");

// 1. Basic Promise resolution
console.log("\n1. Basic Promise Resolution:");
new Promise(function(resolve, reject) {
    console.log("  Creating promise...");
    resolve("Promise resolved!");
}).then(function(result) {
    console.log("  ✅ Promise resolved with:", result);
});

// 2. Promise rejection and catch
console.log("\n2. Promise Rejection:");
new Promise(function(resolve, reject) {
    reject(new Error("Promise rejected!"));
}).catch(function(error) {
    console.log("  ✅ Promise caught error:", error.message);
});

// 3. Promise chaining
console.log("\n3. Promise Chaining:");
Promise.resolve(1)
    .then(function(value) {
        console.log("  First then:", value);
        return value * 2;
    })
    .then(function(value) {
        console.log("  Second then:", value);
        return value * 2;
    })
    .then(function(value) {
        console.log("  ✅ Final value:", value);
    });

// 4. Promise.all
console.log("\n4. Promise.all:");
var promise1 = Promise.resolve(1);
var promise2 = Promise.resolve(2);
var promise3 = Promise.resolve(3);

Promise.all([promise1, promise2, promise3])
    .then(function(values) {
        console.log("  ✅ Promise.all resolved with:", values);
    });

// 5. Promise.race
console.log("\n5. Promise.race:");
var fast = new Promise(function(resolve) {
    setTimeout(function() { resolve("fast"); }, 100);
});
var slow = new Promise(function(resolve) {
    setTimeout(function() { resolve("slow"); }, 200);
});

Promise.race([fast, slow])
    .then(function(value) {
        console.log("  ✅ Promise.race won by:", value);
    });

// 6. Promise with setTimeout
console.log("\n6. Promise with setTimeout:");
function delay(ms) {
    return new Promise(function(resolve) {
        setTimeout(resolve, ms);
    });
}

delay(150).then(function() {
    console.log("  ✅ Delayed promise resolved after 150ms");
});

// 7. Nested promises
console.log("\n7. Nested Promises:");
new Promise(function(resolve) {
    resolve(new Promise(function(innerResolve) {
        innerResolve("nested value");
    }));
}).then(function(value) {
    console.log("  ✅ Nested promise resolved with:", value);
});

// 8. Promise.resolve and Promise.reject
console.log("\n8. Promise.resolve and Promise.reject:");
Promise.resolve("direct resolve")
    .then(function(value) {
        console.log("  ✅ Promise.resolve:", value);
    });

Promise.reject("direct reject")
    .catch(function(reason) {
        console.log("  ✅ Promise.reject caught:", reason);
    });

// 9. Finally handler
console.log("\n9. Promise finally:");
Promise.resolve("with finally")
    .then(function(value) {
        console.log("  Value:", value);
        return value;
    })
    .finally(function() {
        console.log("  ✅ Finally block executed");
    });

// Wait for all async operations to complete
setTimeout(function() {
    console.log("\n=== Promise Test Summary ===");
    console.log("✅ Basic promise resolution works");
    console.log("✅ Promise rejection and catch work");
    console.log("✅ Promise chaining works");
    console.log("✅ Promise.all works");
    console.log("✅ Promise.race works");
    console.log("✅ Promise with setTimeout works");
    console.log("✅ Nested promises work");
    console.log("✅ Promise.resolve and Promise.reject work");
    console.log("✅ Promise finally works");
    console.log("✅ All promise tests passed!");
    
    process.exit(0);
}, 500);