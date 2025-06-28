var validCredentials = [
    { username: "admin", password: "admin123", role: "admin" },
    { username: "editor", password: "editor123", role: "editor" },
    { username: "user", password: "user123", role: "user" }
];

if (req.method === 'POST') {
    if (req.body) {
        try {
            var credentials = JSON.parse(req.body);
            
            if (!credentials.username || !credentials.password) {
                res.status(400);
                res.json({ 
                    success: false, 
                    message: "Username and password are required" 
                });
            } else {
                var user = null;
                for (var i = 0; i < validCredentials.length; i++) {
                    if (validCredentials[i].username === credentials.username && 
                        validCredentials[i].password === credentials.password) {
                        user = validCredentials[i];
                        break;
                    }
                }
                
                if (user) {
                    res.json({
                        success: true,
                        message: "Login successful",
                        data: {
                            username: user.username,
                            role: user.role,
                            token: "mock-jwt-token-" + Date.now(),
                            expiresAt: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString()
                        }
                    });
                } else {
                    res.status(401);
                    res.json({
                        success: false,
                        message: "Invalid username or password"
                    });
                }
            }
        } catch (error) {
            res.status(400);
            res.json({ 
                success: false, 
                message: "Invalid JSON data" 
            });
        }
    } else {
        res.status(400);
        res.json({ 
            success: false, 
            message: "Request body required" 
        });
    }
} else {
    res.status(405);
    res.json({ 
        success: false, 
        message: "Method not allowed. Use POST to login." 
    });
}