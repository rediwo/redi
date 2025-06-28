var stats = {
    server: {
        name: "Redi Frontend Server",
        version: "1.0.0",
        uptime: Math.floor(Math.random() * 86400), // Random uptime in seconds
        startTime: new Date(Date.now() - Math.floor(Math.random() * 86400000)).toISOString()
    },
    content: {
        totalPosts: 15,
        publishedPosts: 12,
        draftPosts: 3,
        totalUsers: 3,
        activeUsers: 2,
        totalPages: 8
    },
    traffic: {
        todayViews: Math.floor(Math.random() * 1000) + 100,
        weeklyViews: Math.floor(Math.random() * 5000) + 500,
        monthlyViews: Math.floor(Math.random() * 20000) + 2000,
        totalViews: Math.floor(Math.random() * 100000) + 10000
    },
    performance: {
        averageResponseTime: Math.floor(Math.random() * 100) + 20, // ms
        memoryUsage: Math.floor(Math.random() * 100) + 50, // MB
        cpuUsage: Math.floor(Math.random() * 50) + 10 // %
    },
    features: {
        markdownSupport: true,
        jsEngineSupport: true,
        templateLayouts: true,
        dynamicRouting: true,
        staticFileServing: true,
        apiEndpoints: true
    }
};

if (req.method === 'GET') {
    var category = req.query ? req.query.category : null;
    
    if (category && stats[category]) {
        res.json({
            success: true,
            category: category,
            data: stats[category],
            timestamp: new Date().toISOString()
        });
    } else if (category) {
        res.status(404);
        res.json({
            success: false,
            message: "Category not found",
            availableCategories: Object.keys(stats)
        });
    } else {
        res.json({
            success: true,
            data: stats,
            timestamp: new Date().toISOString(),
            generatedAt: new Date().toISOString()
        });
    }
    
} else {
    res.status(405);
    res.json({ 
        success: false, 
        message: "Method not allowed. Use GET to retrieve stats." 
    });
}