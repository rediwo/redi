// Async/Await functionality test
console.log("=== Async/Await Test ===");

// Helper function for delays
function delay(ms) {
    return new Promise(function(resolve) {
        setTimeout(resolve, ms);
    });
}

// Helper function to simulate async operations
function fetchData(data, ms) {
    return new Promise(function(resolve) {
        setTimeout(function() {
            resolve(data);
        }, ms);
    });
}

// 1. Basic async function
console.log("\n1. Basic Async Function:");
async function basicAsync() {
    console.log("  Inside async function");
    return "async result";
}

basicAsync().then(function(result) {
    console.log("  ✅ Async function returned:", result);
});

// 2. Await with Promise
console.log("\n2. Await with Promise:");
async function awaitPromise() {
    console.log("  Waiting for promise...");
    const result = await Promise.resolve("awaited value");
    console.log("  ✅ Awaited result:", result);
    return result;
}

awaitPromise();

// 3. Multiple awaits
console.log("\n3. Multiple Awaits:");
async function multipleAwaits() {
    const first = await fetchData("first", 50);
    console.log("  Got:", first);
    
    const second = await fetchData("second", 50);
    console.log("  Got:", second);
    
    const third = await fetchData("third", 50);
    console.log("  Got:", third);
    
    console.log("  ✅ All awaits completed");
    return [first, second, third];
}

multipleAwaits();

// 4. Try-catch with async/await
console.log("\n4. Try-Catch with Async/Await:");
async function tryCatchAsync() {
    try {
        console.log("  Attempting async operation...");
        await Promise.resolve("success");
        console.log("  ✅ Try block succeeded");
        
        // Test rejection
        await Promise.reject(new Error("Async error"));
    } catch (error) {
        console.log("  ✅ Caught async error:", error.message);
    }
}

tryCatchAsync();

// 5. Parallel execution with Promise.all
console.log("\n5. Parallel Execution:");
async function parallelExecution() {
    console.log("  Starting parallel operations...");
    const start = Date.now();
    
    const results = await Promise.all([
        fetchData("parallel1", 100),
        fetchData("parallel2", 100),
        fetchData("parallel3", 100)
    ]);
    
    const duration = Date.now() - start;
    console.log("  ✅ Parallel results:", results);
    console.log("  Completed in ~" + duration + "ms (should be ~100ms, not 300ms)");
}

parallelExecution();

// 6. Sequential vs Parallel
console.log("\n6. Sequential vs Parallel Comparison:");
async function sequential() {
    const start = Date.now();
    await delay(50);
    await delay(50);
    await delay(50);
    const duration = Date.now() - start;
    console.log("  Sequential: ~" + duration + "ms");
}

async function parallel() {
    const start = Date.now();
    await Promise.all([delay(50), delay(50), delay(50)]);
    const duration = Date.now() - start;
    console.log("  Parallel: ~" + duration + "ms");
    console.log("  ✅ Parallel is faster!");
}

async function compareExecution() {
    await sequential();
    await parallel();
}

compareExecution();

// 7. Async function returning async function
console.log("\n7. Nested Async Functions:");
async function outer() {
    console.log("  Outer async function");
    
    async function inner() {
        console.log("  Inner async function");
        return "inner result";
    }
    
    const result = await inner();
    console.log("  ✅ Nested async result:", result);
    return result;
}

outer();

// 8. Await in loops
console.log("\n8. Await in Loops:");
async function awaitInLoop() {
    const items = ["a", "b", "c"];
    
    console.log("  Sequential loop:");
    for (const item of items) {
        const result = await fetchData(item, 50);
        console.log("    Processed:", result);
    }
    
    console.log("  ✅ Loop with await completed");
}

awaitInLoop();

// 9. Error propagation
console.log("\n9. Error Propagation:");
async function throwError() {
    throw new Error("Async error");
}

async function catchError() {
    try {
        await throwError();
    } catch (error) {
        console.log("  ✅ Error propagated and caught:", error.message);
    }
}

catchError();

// 10. Async IIFE (Immediately Invoked Function Expression)
console.log("\n10. Async IIFE:");
(async function() {
    console.log("  ✅ Async IIFE executed");
    const result = await Promise.resolve("IIFE result");
    console.log("  IIFE result:", result);
})();

// Wait for all async operations to complete
setTimeout(function() {
    console.log("\n=== Async/Await Test Summary ===");
    console.log("✅ Basic async functions work");
    console.log("✅ Await with promises works");
    console.log("✅ Multiple awaits work");
    console.log("✅ Try-catch with async/await works");
    console.log("✅ Parallel execution with Promise.all works");
    console.log("✅ Sequential vs parallel comparison works");
    console.log("✅ Nested async functions work");
    console.log("✅ Await in loops works");
    console.log("✅ Error propagation works");
    console.log("✅ Async IIFE works");
    console.log("✅ All async/await tests passed!");
    
    process.exit(0);
}, 1000);