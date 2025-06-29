var postsDb = require('../_data/posts');

exports.get = function(req, res, next) {
    var posts = postsDb.getAll(); // Only get published posts
    res.render({
        Title: "Blog",
        posts: posts
    });
};