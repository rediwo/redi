var postsDb = require('../_data/posts');

exports.get = function(req, res, next) {
    // Get all posts including drafts
    var posts = postsDb.getAll(true);
    
    res.render({
        Title: "Post Management",
        posts: posts
    });
};