var usersDb = require('../_data/users');

exports.get = function(req, res, next) {
    res.render({
        Title: "User Management",
        users: usersDb.getAll()
    });
};