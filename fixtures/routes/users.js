var usersDb = require('./_data/users');

exports.get = function(req, res, next) {
    var users = usersDb.getAll();
    res.render({
        Title: "Users",
        users: users,
        totalUsers: users.length
    });
};