var fs = require('fs');

// Handle GET requests
exports.get = function(req, res, next) {
    res.render({
        Title: "HTTP Method Example",
        method: req.method,
        path: req.path,
        message: "This page supports GET, POST, PUT, and DELETE methods."
    });
};

// Handle POST requests
exports.post = function(req, res, next) {
    var body = {};
    if (req.body) {
        var contentType = req.headers['Content-Type'] || req.headers['content-type'] || '';
        if (contentType.indexOf('application/json') !== -1) {
            // Parse JSON data
            try {
                body = JSON.parse(req.body);
            } catch (e) {
                body = { error: 'Invalid JSON: ' + e.message };
            }
        } else if (contentType.indexOf('application/x-www-form-urlencoded') !== -1) {
            // Parse form data
            var pairs = req.body.split('&');
            for (var i = 0; i < pairs.length; i++) {
                var pair = pairs[i].split('=');
                if (pair.length === 2) {
                    // Simple URL decode (replace + with space and basic % escapes)
                    var key = pair[0].replace(/\+/g, ' ').replace(/%20/g, ' ');
                    var value = pair[1].replace(/\+/g, ' ').replace(/%20/g, ' ');
                    body[key] = value;
                }
            }
        } else {
            // Raw body
            body = { rawData: req.body };
        }
    }
    res.render({
        Title: "POST Request Received",
        method: req.method,
        receivedData: body,
        message: "POST data was received successfully."
    });
};

// Handle PUT requests
exports.put = function(req, res, next) {
    var body = {};
    if (req.body) {
        try {
            body = JSON.parse(req.body);
        } catch (e) {
            body = { error: 'Invalid JSON: ' + e.message, rawData: req.body };
        }
    }
    res.render({
        Title: "PUT Request Received", 
        method: req.method,
        receivedData: body,
        message: "PUT data was received successfully."
    });
};

// Handle DELETE requests (using 'del' since 'delete' is a reserved word)
exports.del = function(req, res, next) {
    res.render({
        Title: "DELETE Request Received",
        method: req.method,
        message: "DELETE request was processed successfully."
    });
};