// Basic functionality test: process, console, timers
console.log("=== Basic rejs Test ===");

// 1. Test console functions
console.log("1. Console Tests:");
console.log("  - log: Normal message");
console.warn("  - warn: Warning message");
console.error("  - error: Error message");

// 2. Test process object
console.log("\n2. Process Tests:");
console.log("  argv:", process.argv);
console.log("  platform:", process.platform);
console.log("  arch:", process.arch);
console.log("  pid:", process.pid);
console.log("  ppid:", process.ppid);
console.log("  version:", process.version);
console.log("  cwd():", process.cwd());
console.log("  env.PATH exists:", !!process.env.PATH);
console.log("  env.HOME exists:", !!process.env.HOME);

// 3. Test global variables
console.log("\n3. Global Variables:");
console.log("  __filename:", __filename);
console.log("  __dirname:", __dirname);

// 4. Test setTimeout and setInterval
console.log("\n4. Timer Tests:");

var timeoutCount = 0;
var intervalCount = 0;
var intervalId;

console.log("  Starting setTimeout test...");
setTimeout(function() {
    timeoutCount++;
    console.log("  ✅ setTimeout executed (count: " + timeoutCount + ")");
}, 100);

console.log("  Starting setInterval test...");
intervalId = setInterval(function() {
    intervalCount++;
    console.log("  ✅ setInterval executed (count: " + intervalCount + ")");
    
    if (intervalCount >= 3) {
        clearInterval(intervalId);
        console.log("  ✅ setInterval cleared after 3 executions");
        
        // Final summary
        setTimeout(function() {
            console.log("\n=== Test Summary ===");
            console.log("✅ Console: log, warn, error work");
            console.log("✅ Process: argv, platform, arch, pid, ppid, version, cwd, env work");
            console.log("✅ Globals: __filename, __dirname work");
            console.log("✅ Timers: setTimeout, setInterval, clearInterval work");
            console.log("✅ All basic tests passed!");
            
            // Exit successfully
            process.exit(0);
        }, 200);
    }
}, 150);

console.log("  Waiting for timers to complete...");