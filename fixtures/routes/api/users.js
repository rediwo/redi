var usersDb = require('../_data/users');

// Helper function to parse form data
function parseFormData(body) {
    var result = {};
    var pairs = body.split('&');
    
    for (var i = 0; i < pairs.length; i++) {
        var pair = pairs[i].split('=');
        if (pair.length === 2) {
            var key = decodeURIComponent(pair[0]);
            var value = decodeURIComponent(pair[1]);
            result[key] = value;
        }
    }
    
    return result;
}

// GET /api/users
exports.get = function(req, res, next) {
    var userId = req.query ? req.query.id : null;
    
    if (userId) {
        var user = usersDb.getById(userId);
        if (user) {
            res.json({ success: true, data: user });
        } else {
            res.status(404);
            res.json({ success: false, message: "User not found" });
        }
    } else {
        var users = usersDb.getAll();
        res.json({ 
            success: true, 
            data: users,
            count: users.length,
            timestamp: new Date().toISOString()
        });
    }
};

// POST /api/users
exports.post = function(req, res, next) {
    if (req.body) {
        var userData;
        
        // Check if the body is JSON or form data
        try {
            // Try parsing as JSON first
            userData = JSON.parse(req.body);
        } catch (error) {
            // If JSON parsing fails, try parsing as form data
            userData = parseFormData(req.body);
        }
        
        if (!userData.name || !userData.email) {
            res.status(400);
            res.json({ success: false, message: "Name and email are required" });
        } else {
            var existingUser = usersDb.getByEmail(userData.email);
            
            if (existingUser) {
                res.status(409);
                res.json({ success: false, message: "Email already exists" });
            } else {
                var newUser = usersDb.create(userData);
                
                res.status(201);
                res.json({ 
                    success: true, 
                    message: "User created successfully",
                    data: newUser 
                });
            }
        }
    } else {
        res.status(400);
        res.json({ success: false, message: "Request body required" });
    }
};

// PUT /api/users/{id}
exports.put = function(req, res, next) {
    var userId = req.params ? req.params.id : null;
    
    if (!userId) {
        res.status(400);
        res.json({ success: false, message: "User ID required" });
        return;
    }
    
    if (req.body) {
        try {
            var userData = JSON.parse(req.body);
            var updatedUser = usersDb.update(userId, userData);
            
            if (updatedUser) {
                res.json({ 
                    success: true, 
                    message: "User updated successfully",
                    data: updatedUser 
                });
            } else {
                res.status(404);
                res.json({ success: false, message: "User not found" });
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

// DELETE /api/users/{id}
exports.delete = function(req, res, next) {
    var userId = req.params ? req.params.id : null;
    
    if (!userId) {
        res.status(400);
        res.json({ success: false, message: "User ID required" });
        return;
    }
    
    if (usersDb.delete(userId)) {
        res.json({ 
            success: true, 
            message: "User deleted successfully" 
        });
    } else {
        res.status(404);
        res.json({ success: false, message: "User not found" });
    }
};