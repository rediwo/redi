var postsDb = require('../_data/posts');
var usersDb = require('../_data/users');
var rolesDb = require('../_data/roles');

exports.get = function(req, res, next) {
    var category = req.query ? req.query.category : null;
    var postStats = postsDb.getStats();
    var allUsers = usersDb.getAll();
    var activeUsers = 0;
    
    for (var i = 0; i < allUsers.length; i++) {
        if (allUsers[i].status === 'active') {
            activeUsers++;
        }
    }
    
    var stats = {
        server: {
            name: "Redi Frontend Server",
            version: "1.0.0",
            uptime: process.uptime ? Math.floor(process.uptime()) : 0,
            startTime: new Date(Date.now() - (process.uptime ? process.uptime() * 1000 : 0)).toISOString()
        },
        content: {
            totalPosts: postStats.total,
            publishedPosts: postStats.published,
            draftPosts: postStats.draft,
            totalUsers: allUsers.length,
            activeUsers: activeUsers,
            totalRoles: rolesDb.getAll().length,
            totalPages: 8
        },
        traffic: {
            todayViews: Math.floor(postStats.totalViews * 0.1),
            weeklyViews: Math.floor(postStats.totalViews * 0.3),
            monthlyViews: Math.floor(postStats.totalViews * 0.7),
            totalViews: postStats.totalViews
        },
        performance: {
            averageResponseTime: Math.floor(Math.random() * 100) + 20, // ms
            memoryUsage: process.memoryUsage ? Math.floor(process.memoryUsage().heapUsed / 1024 / 1024) : 0, // MB
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
};