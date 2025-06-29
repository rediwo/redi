// In-memory posts database
var posts = [
    {
        id: 1,
        title: "Welcome to our blog",
        content: "This is the first post on our test blog. It demonstrates how dynamic routing works with the [id] parameter.",
        excerpt: "This is the first post on our test blog. It demonstrates how dynamic routing works...",
        date: "2024-01-15",
        author: "Admin",
        authorId: 1,
        status: "published",
        tags: ["welcome", "introduction"],
        views: 125,
        comments: 3
    },
    {
        id: 2,
        title: "Getting started with Redi",
        content: "Learn how to use the Redi frontend server. This post covers the basic features and how to set up your own blog.",
        excerpt: "Learn how to use the Redi frontend server. This post covers the basic features...",
        date: "2024-01-20",
        author: "Developer",
        authorId: 2,
        status: "published",
        tags: ["tutorial", "redi", "setup"],
        views: 89,
        comments: 5
    },
    {
        id: 3,
        title: "Dynamic Route Example",
        content: "This demonstrates the dynamic routing feature where [id] in the filename becomes a parameter.",
        excerpt: "This demonstrates the dynamic routing feature where [id] in the filename...",
        date: "2024-01-25",
        author: "System",
        authorId: 1,
        status: "published",
        tags: ["example", "routing"],
        views: 45,
        comments: 1
    },
    {
        id: 4,
        title: "Draft Post Example",
        content: "This is a draft post that should not be visible in the public blog listing.",
        excerpt: "This is a draft post that should not be visible...",
        date: "2024-01-28",
        author: "Editor",
        authorId: 2,
        status: "draft",
        tags: ["draft"],
        views: 0,
        comments: 0
    }
];

var nextId = 5;

// CRUD operations
exports.getAll = function(includeUnpublished) {
    if (includeUnpublished) {
        return posts;
    }
    return posts.filter(function(post) {
        return post.status === 'published';
    });
};

exports.getById = function(id) {
    return posts.find(function(post) {
        return post.id === parseInt(id);
    });
};

exports.getByAuthorId = function(authorId) {
    return posts.filter(function(post) {
        return post.authorId === parseInt(authorId);
    });
};

exports.getByStatus = function(status) {
    return posts.filter(function(post) {
        return post.status === status;
    });
};

exports.create = function(postData) {
    var newPost = {
        id: nextId++,
        title: postData.title || '',
        content: postData.content || '',
        excerpt: postData.excerpt || postData.content.substring(0, 100) + '...',
        date: postData.date || new Date().toISOString().split('T')[0],
        author: postData.author || 'Anonymous',
        authorId: postData.authorId || 1,
        status: postData.status || 'draft',
        tags: postData.tags || [],
        views: 0,
        comments: 0
    };
    posts.push(newPost);
    return newPost;
};

exports.update = function(id, postData) {
    var index = posts.findIndex(function(post) {
        return post.id === parseInt(id);
    });
    
    if (index === -1) {
        return null;
    }
    
    posts[index] = Object.assign({}, posts[index], postData);
    return posts[index];
};

exports.delete = function(id) {
    var index = posts.findIndex(function(post) {
        return post.id === parseInt(id);
    });
    
    if (index === -1) {
        return false;
    }
    
    posts.splice(index, 1);
    return true;
};

exports.incrementViews = function(id) {
    var post = exports.getById(id);
    if (post) {
        post.views++;
        return post;
    }
    return null;
};

exports.count = function(status) {
    if (status) {
        return posts.filter(function(post) {
            return post.status === status;
        }).length;
    }
    return posts.length;
};

exports.getStats = function() {
    return {
        total: posts.length,
        published: exports.count('published'),
        draft: exports.count('draft'),
        totalViews: posts.reduce(function(sum, post) {
            return sum + post.views;
        }, 0),
        totalComments: posts.reduce(function(sum, post) {
            return sum + post.comments;
        }, 0)
    };
};