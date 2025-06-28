var users = [
    { id: 1, name: "John Doe", email: "john@example.com", role: "admin", status: "active" },
    { id: 2, name: "Jane Smith", email: "jane@example.com", role: "editor", status: "active" },
    { id: 3, name: "Bob Johnson", email: "bob@example.com", role: "user", status: "inactive" }
];

if (req.method === 'GET') {
    var userId = req.query ? req.query.id : null;
    
    if (userId) {
        var user = null;
        for (var i = 0; i < users.length; i++) {
            if (users[i].id === parseInt(userId)) {
                user = users[i];
                break;
            }
        }
        if (user) {
            res.json({ success: true, data: user });
        } else {
            res.status(404);
            res.json({ success: false, message: "User not found" });
        }
    } else {
        res.json({ 
            success: true, 
            data: users,
            count: users.length,
            timestamp: new Date().toISOString()
        });
    }
    
} else if (req.method === 'POST') {
    if (req.body) {
        try {
            var userData = JSON.parse(req.body);
            
            if (!userData.name || !userData.email) {
                res.status(400);
                res.json({ success: false, message: "Name and email are required" });
            } else {
                var existingUser = null;
                for (var i = 0; i < users.length; i++) {
                    if (users[i].email === userData.email) {
                        existingUser = users[i];
                        break;
                    }
                }
                
                if (existingUser) {
                    res.status(409);
                    res.json({ success: false, message: "Email already exists" });
                } else {
                    var maxId = 0;
                    for (var i = 0; i < users.length; i++) {
                        if (users[i].id > maxId) maxId = users[i].id;
                    }
                    
                    var newUser = {
                        id: maxId + 1,
                        name: userData.name,
                        email: userData.email,
                        role: userData.role || 'user',
                        status: userData.status || 'active'
                    };
                    
                    users.push(newUser);
                    
                    res.status(201);
                    res.json({ 
                        success: true, 
                        message: "User created successfully",
                        data: newUser 
                    });
                }
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