var postsDb = require('./_data/posts');

exports.get = function(req, res, next) {
    // Get the 3 most recent published posts
    var posts = postsDb.getAll().slice(0, 3);
    
    res.render({
        Title: "Home",
        posts: posts
    });
};