// In-memory user database
var users = [
    { id: 1, name: "John Doe", email: "john@example.com", role: "admin", status: "active" },
    { id: 2, name: "Jane Smith", email: "jane@example.com", role: "editor", status: "active" },
    { id: 3, name: "Bob Johnson", email: "bob@example.com", role: "user", status: "inactive" }
];

var nextId = 4;

// CRUD operations
exports.getAll = function() {
    return users;
};

exports.getById = function(id) {
    return users.find(function(user) {
        return user.id === parseInt(id);
    });
};

exports.getByEmail = function(email) {
    return users.find(function(user) {
        return user.email === email;
    });
};

exports.create = function(userData) {
    var newUser = {
        id: nextId++,
        name: userData.name || '',
        email: userData.email || '',
        role: userData.role || 'user',
        status: userData.status || 'active'
    };
    users.push(newUser);
    return newUser;
};

exports.update = function(id, userData) {
    var index = users.findIndex(function(user) {
        return user.id === parseInt(id);
    });
    
    if (index === -1) {
        return null;
    }
    
    users[index] = Object.assign({}, users[index], userData);
    return users[index];
};

exports.delete = function(id) {
    var index = users.findIndex(function(user) {
        return user.id === parseInt(id);
    });
    
    if (index === -1) {
        return false;
    }
    
    users.splice(index, 1);
    return true;
};

exports.count = function() {
    return users.length;
};