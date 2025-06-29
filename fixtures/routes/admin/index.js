var postsDb = require('../_data/posts');
var usersDb = require('../_data/users');

exports.get = function(req, res, next) {
    var postStats = postsDb.getStats();
    
    res.render({
        Title: "Dashboard",
        stats: {
            totalPosts: postStats.published,
            totalUsers: usersDb.count(),
            totalViews: postStats.totalViews,
            totalComments: postStats.totalComments
        },
        recentActivity: [
            { action: "New post created", time: "2 hours ago", user: "Admin" },
            { action: "User registered", time: "4 hours ago", user: "newuser@example.com" },
            { action: "Post edited", time: "1 day ago", user: "Editor" }
        ]
    });
};