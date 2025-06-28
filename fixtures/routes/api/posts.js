var posts = [
    { 
        id: 1, 
        title: "Welcome to our blog", 
        content: "This is the first post on our test blog.",
        author: "Admin",
        status: "published",
        date: "2024-01-15",
        tags: ["welcome", "introduction"]
    },
    { 
        id: 2, 
        title: "Getting started with Redi", 
        content: "Learn how to use the Redi frontend server.",
        author: "Developer",
        status: "published", 
        date: "2024-01-20",
        tags: ["tutorial", "redi"]
    },
    { 
        id: 3, 
        title: "Draft Post", 
        content: "This is a draft post that hasn't been published yet.",
        author: "Editor",
        status: "draft", 
        date: "2024-01-25",
        tags: ["draft"]
    }
];

if (req.method === 'GET') {
    var postId = req.query ? req.query.id : null;
    var status = req.query ? req.query.status : null;
    
    if (postId) {
        var post = null;
        for (var i = 0; i < posts.length; i++) {
            if (posts[i].id === parseInt(postId)) {
                post = posts[i];
                break;
            }
        }
        if (post) {
            res.json({ success: true, data: post });
        } else {
            res.status(404);
            res.json({ success: false, message: "Post not found" });
        }
    } else {
        var filteredPosts = posts;
        
        if (status) {
            filteredPosts = [];
            for (var i = 0; i < posts.length; i++) {
                if (posts[i].status === status) {
                    filteredPosts.push(posts[i]);
                }
            }
        }
        
        res.json({ 
            success: true, 
            data: filteredPosts,
            count: filteredPosts.length,
            filters: { status: status || 'all' },
            timestamp: new Date().toISOString()
        });
    }
    
} else if (req.method === 'POST') {
    if (req.body) {
        try {
            var postData = JSON.parse(req.body);
            
            if (!postData.title || !postData.content) {
                res.status(400);
                res.json({ success: false, message: "Title and content are required" });
            } else {
                var tags = [];
                if (postData.tags) {
                    var tagList = postData.tags.split(',');
                    for (var i = 0; i < tagList.length; i++) {
                        var tag = tagList[i].trim();
                        if (tag) {
                            tags.push(tag);
                        }
                    }
                }
                
                var maxId = 0;
                for (var i = 0; i < posts.length; i++) {
                    if (posts[i].id > maxId) maxId = posts[i].id;
                }
                
                var newPost = {
                    id: maxId + 1,
                    title: postData.title,
                    content: postData.content,
                    author: postData.author || 'Anonymous',
                    status: postData.status || 'draft',
                    date: new Date().toISOString().split('T')[0],
                    tags: tags
                };
                
                posts.push(newPost);
                
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
} else {
    res.status(405);
    res.json({ success: false, message: "Method not allowed" });
}