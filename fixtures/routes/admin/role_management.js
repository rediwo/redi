var rolesDb = require('../_data/roles');

exports.get = function(req, res, next) {
    res.render({
        Title: "Role Management",
        roles: rolesDb.getAll()
    });
};