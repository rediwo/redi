// Performance test endpoints for concurrent load testing
exports.get = function(req, res, next) {
    var testType = (req.query.split('type=')[1] || 'fast').split('&')[0];
    var delay = parseInt(req.query.split('delay=')[1]) || 0;
    
    if (testType === 'fast') {
        // Fast response test - minimal processing
        res.json({
            success: true,
            message: "Fast response",
            timestamp: new Date().toISOString(),
            requestId: Math.random().toString(36).substr(2, 9)
        });
        
    } else if (testType === 'cpu') {
        // CPU intensive test
        var start = Date.now();
        var result = 0;
        for (var i = 0; i < 100000; i++) {
            result += Math.sqrt(i) * Math.sin(i);
        }
        var duration = Date.now() - start;
        
        res.json({
            success: true,
            message: "CPU intensive test completed",
            duration: duration + "ms",
            result: result,
            timestamp: new Date().toISOString()
        });
        
    } else if (testType === 'async') {
        // Async test with setTimeout
        setTimeout(function() {
            res.json({
                success: true,
                message: "Async test completed",
                delay: delay + "ms",
                timestamp: new Date().toISOString()
            });
        }, delay);
        
    } else if (testType === 'promise') {
        // Promise chain test
        Promise.resolve(42)
            .then(function(value) {
                return value * 2;
            })
            .then(function(value) {
                return new Promise(function(resolve) {
                    setTimeout(function() {
                        resolve(value + 10);
                    }, delay);
                });
            })
            .then(function(result) {
                res.json({
                    success: true,
                    message: "Promise chain test completed",
                    result: result,
                    delay: delay + "ms",
                    timestamp: new Date().toISOString()
                });
            });
            
    } else if (testType === 'fetch_external') {
        // External fetch test (to avoid localhost issues)
        fetch('https://httpbin.org/delay/' + Math.min(delay/1000, 5))
            .then(function(response) {
                return response.json();
            })
            .then(function(data) {
                res.json({
                    success: true,
                    message: "External fetch test completed",
                    delay: delay + "ms",
                    hasData: !!data,
                    timestamp: new Date().toISOString()
                });
            })
            .catch(function(error) {
                res.status(500);
                res.json({
                    success: false,
                    error: error.toString(),
                    timestamp: new Date().toISOString()
                });
            });
            
    } else if (testType === 'memory') {
        // Memory allocation test
        var data = [];
        for (var i = 0; i < 10000; i++) {
            data.push({
                id: i,
                text: "Sample data item " + i,
                timestamp: new Date().toISOString(),
                random: Math.random()
            });
        }
        
        res.json({
            success: true,
            message: "Memory test completed",
            itemCount: data.length,
            timestamp: new Date().toISOString()
        });
    }
};