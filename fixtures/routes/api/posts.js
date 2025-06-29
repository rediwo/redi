var postsDb = require('../_data/posts');

// GET /api/posts
exports.get = function(req, res, next) {
    var postId = req.query ? req.query.id : null;
    var status = req.query ? req.query.status : null;
    
    if (postId) {
        var post = postsDb.getById(postId);
        if (post) {
            res.json({ success: true, data: post });
        } else {
            res.status(404);
            res.json({ success: false, message: "Post not found" });
        }
    } else {
        var posts;
        
        if (status === 'all') {
            posts = postsDb.getAll(true); // Include unpublished
        } else if (status) {
            posts = postsDb.getByStatus(status);
        } else {
            posts = postsDb.getAll(); // Only published
        }
        
        res.json({ 
            success: true, 
            data: posts,
            count: posts.length,
            filters: { status: status || 'published' },
            timestamp: new Date().toISOString()
        });
    }
};

// POST /api/posts
exports.post = function(req, res, next) {
    if (req.body) {
        try {
            var postData = JSON.parse(req.body);
            
            if (!postData.title || !postData.content) {
                res.status(400);
                res.json({ success: false, message: "Title and content are required" });
            } else {
                // Process tags
                var tags = [];
                if (postData.tags) {
                    if (typeof postData.tags === 'string') {
                        var tagList = postData.tags.split(',');
                        for (var i = 0; i < tagList.length; i++) {
                            var tag = tagList[i].trim();
                            if (tag) {
                                tags.push(tag);
                            }
                        }
                    } else if (Array.isArray(postData.tags)) {
                        tags = postData.tags;
                    }
                }
                postData.tags = tags;
                
                var newPost = postsDb.create(postData);
                
                res.status(201);
                res.json({ 
                    success: true, 
                    message: "Post created successfully",
                    data: newPost 
                });
            }
        } catch (error) {
            res.status(400);
            res.json({ success: false, message: "Invalid JSON data" });
        }
    } else {
        res.status(400);
        res.json({ success: false, message: "Request body required" });
    }
};

// PUT /api/posts/{id}
exports.put = function(req, res, next) {
    var postId = req.params ? req.params.id : null;
    
    if (!postId) {
        res.status(400);
        res.json({ success: false, message: "Post ID required" });
        return;
    }
    
    if (req.body) {
        try {
            var postData = JSON.parse(req.body);
            
            // Process tags if provided
            if (postData.tags && typeof postData.tags === 'string') {
                var tags = [];
                var tagList = postData.tags.split(',');
                for (var i = 0; i < tagList.length; i++) {
                    var tag = tagList[i].trim();
                    if (tag) {
                        tags.push(tag);
                    }
                }
                postData.tags = tags;
            }
            
            var updatedPost = postsDb.update(postId, postData);
            
            if (updatedPost) {
                res.json({ 
                    success: true, 
                    message: "Post updated successfully",
                    data: updatedPost 
                });
            } else {
                res.status(404);
                res.json({ success: false, message: "Post not found" });
            }
        } catch (error) {
            res.status(400);
            res.json({ success: false, message: "Invalid JSON data" });
        }
    } else {
        res.status(400);
        res.json({ success: false, message: "Request body required" });
    }
};

// DELETE /api/posts/{id}
exports.delete = function(req, res, next) {
    var postId = req.params ? req.params.id : null;
    
    if (!postId) {
        res.status(400);
        res.json({ success: false, message: "Post ID required" });
        return;
    }
    
    if (postsDb.delete(postId)) {
        res.json({ 
            success: true, 
            message: "Post deleted successfully" 
        });
    } else {
        res.status(404);
        res.json({ success: false, message: "Post not found" });
    }
};