// In-memory roles database
var roles = [
    { 
        id: 1,
        name: "admin", 
        description: "Full system access", 
        permissions: ["create", "read", "update", "delete", "manage_users", "manage_roles", "system_admin"],
        userCount: 1
    },
    { 
        id: 2,
        name: "editor", 
        description: "Content management", 
        permissions: ["create", "read", "update", "publish"],
        userCount: 1
    },
    { 
        id: 3,
        name: "user", 
        description: "Basic access", 
        permissions: ["read", "comment"],
        userCount: 1
    }
];

var nextId = 4;

// CRUD operations
exports.getAll = function() {
    return roles;
};

exports.getById = function(id) {
    return roles.find(function(role) {
        return role.id === parseInt(id);
    });
};

exports.getByName = function(name) {
    return roles.find(function(role) {
        return role.name === name;
    });
};

exports.create = function(roleData) {
    var newRole = {
        id: nextId++,
        name: roleData.name || '',
        description: roleData.description || '',
        permissions: roleData.permissions || [],
        userCount: 0
    };
    roles.push(newRole);
    return newRole;
};

exports.update = function(id, roleData) {
    var index = roles.findIndex(function(role) {
        return role.id === parseInt(id);
    });
    
    if (index === -1) {
        return null;
    }
    
    roles[index] = Object.assign({}, roles[index], roleData);
    return roles[index];
};

exports.delete = function(id) {
    var role = exports.getById(id);
    if (!role || role.userCount > 0) {
        return false; // Can't delete role with users
    }
    
    var index = roles.findIndex(function(r) {
        return r.id === parseInt(id);
    });
    
    if (index === -1) {
        return false;
    }
    
    roles.splice(index, 1);
    return true;
};

exports.hasPermission = function(roleName, permission) {
    var role = exports.getByName(roleName);
    if (!role) {
        return false;
    }
    return role.permissions.indexOf(permission) !== -1;
};