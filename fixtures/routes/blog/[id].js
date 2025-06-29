var postsDb = require('../_data/posts');

exports.get = function(req, res, next) {
    var postId = req.params.id;
    var post = postsDb.getById(postId);

    if (!post) {
        res.status(404);
        res.render({
            Title: "Post Not Found",
            error: "Post not found"
        });
    } else {
        // Increment views
        postsDb.incrementViews(postId);
        
        res.render({
            Title: post.title,
            post: post
        });
    }
};