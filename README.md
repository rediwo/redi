# Redi - Modern Web Server & JavaScript Runtime

Redi is a Go-based web development toolkit that provides both a dynamic web server (`redi`) and a Node.js-compatible JavaScript runtime (`rejs`).

## 🚀 Features

### Redi Web Server
- **Dynamic Routing**: Automatic route discovery from filesystem structure with `[param]` syntax
- **JavaScript API Endpoints**: Execute `.js` files server-side for API routes
- **HTML Template Rendering**: Process `.html` files with Go templates and server-side JavaScript
- **Markdown Support**: Automatic `.md` to HTML conversion with Goldmark parser
- **Template Layouts**: Nested layouts with `{{layout 'name'}}` syntax
- **Background Mode**: Run server as daemon with `--log` parameter (nohup-like behavior)
- **Session-based VM Management**: Consistent JavaScript state across requests per client
- **Static File Serving**: Efficient serving from `public/` directory
- **Cross-Platform**: Works on Linux, macOS, and Windows
- **JavaScript Engine Pooling**: High-performance concurrent request handling

### Rejs JavaScript Runtime
- **Node.js Compatible**: Supports CommonJS modules and npm packages
- **Built-in Modules**: `fs`, `path`, `process`, `child_process`, `fetch`, `console`, and more
- **Event Loop**: Full async/await and Promise support
- **ES5+ Syntax**: Modern JavaScript features via Goja engine
- **Module Caching**: Efficient require system with path resolution
- **Command Line**: Run scripts with timeout control
- **Cross-Platform**: Consistent behavior across all platforms

## 📦 Installation

### Download Pre-built Binaries

Download the latest release for your platform from the [Releases](https://github.com/rediwo/redi/releases) page. Each archive contains both `redi` and `rejs`.

```bash
# Extract the archive
tar -xzf redi-v1.0.0-linux-amd64.tar.gz

# Move to PATH
sudo mv redi-v1.0.0-linux-amd64 /usr/local/bin/redi
sudo mv rejs-v1.0.0-linux-amd64 /usr/local/bin/rejs

# Verify installation
redi --version
rejs --version
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/rediwo/redi.git
cd redi

# Build both tools
make build

# Or build individually
go build -o redi ./cmd/redi
go build -o rejs ./cmd/rejs
```

## 🎯 Quick Start

### Redi Web Server

Create a simple website structure:

```
mysite/
├── public/
│   ├── css/
│   │   └── style.css
│   └── js/
│       └── main.js
└── routes/
    ├── index.html
    └── api/
        └── hello.js
```

**routes/index.html:**
```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - My Site</title>
    
    <!-- Tailwind CSS + Vimesh Style -->
    <script src="https://unpkg.com/@vimesh/style"></script>
    <!-- Alpine.js -->
    <script defer src="https://unpkg.com/alpinejs"></script>
    
    <script>
        $vs.reset({
            aliasColors: {
                primary: "#3b82f6"
            }
        });
    </script>
    <style>[x-cloak] { display: none !important; }</style>
</head>
<body x-data class="bg-gray-50 min-h-screen">
    <div class="container mx-auto px-4 py-8">
        <div class="bg-white rounded-xl shadow-lg p-8">
            <h1 class="text-3xl font-bold text-gray-900 mb-4">{{.Title}}</h1>
            <p class="text-gray-600 mb-4">Current time: {{.time}}</p>
            <p class="text-sm text-gray-500">Request method: {{.method}}</p>
            
            {{if .data}}
            <div class="mt-6 p-4 bg-blue-50 rounded-lg">
                <h2 class="text-xl font-semibold text-blue-900 mb-2">Submitted Data:</h2>
                <pre class="text-sm text-blue-800">{{.data}}</pre>
            </div>
            {{end}}
        </div>
    </div>
</body>
</html>
```

**routes/index.js:** (corresponding JavaScript file)
```javascript
// Handle GET requests  
exports.get = function(req, res, next) {
    res.render({
        Title: "Welcome to Redi",
        time: new Date().toLocaleTimeString(),
        method: req.method
    });
};

// Handle POST requests
exports.post = function(req, res, next) {
    var body = req.body ? JSON.parse(req.body) : {};
    res.render({
        Title: "Form Submitted",
        data: JSON.stringify(body, null, 2),
        time: new Date().toLocaleTimeString(),
        method: req.method
    });
};
```

**routes/api/hello.js:**
```javascript
// Handle GET requests
exports.get = function(req, res, next) {
    res.json({
        message: "Hello from Redi API",
        timestamp: Date.now(),
        method: req.method,
        userAgent: req.headers['user-agent']
    });
};

// Handle POST requests
exports.post = function(req, res, next) {
    var body = req.body ? JSON.parse(req.body) : {};
    res.json({
        message: "Data received",
        receivedData: body,
        timestamp: Date.now()
    });
};
```

Run the server:
```bash
# Run in foreground
redi --root=mysite --port=8080

# Run in background with logging
redi --root=mysite --port=8080 --log=server.log
```

### Rejs JavaScript Runtime

**script.js:**
```javascript
var fs = require('fs');
var path = require('path');

console.log('Running on:', process.platform);
console.log('Script path:', __filename);

// Read a file
var content = fs.readFileSync('data.txt', 'utf8');
console.log('File content:', content);

// Make an HTTP request
var fetch = require('fetch');
fetch('https://api.github.com/users/github')
    .then(function(response) {
        return response.json();
    })
    .then(function(data) {
        console.log('GitHub user:', data.name);
        process.exit(0);
    });
```

Run the script:
```bash
# Run with auto-exit for sync scripts
rejs script.js

# Run with timeout for async operations
rejs --timeout=5000 async-script.js
```

## 📚 Documentation

### Redi Web Server

#### CLI Options
- `--root` - Root directory containing public and routes folders (required)
- `--port` - Port to serve on (default: 8080)
- `--log` - Log file path (enables background/daemon mode like nohup)
- `--version` - Show version information

#### Directory Structure
- `public/` - Static assets (CSS, JS, images)
- `routes/` - Dynamic routes and API endpoints

#### Route Types
- `.html` - HTML templates processed with Go templates
- `.js` - JavaScript files for API endpoints and server-side logic
- `.md` - Markdown files auto-converted to HTML with Goldmark

#### Dynamic Routes
Use `[param]` syntax for dynamic segments:
- `routes/blog/[id].html` → `/blog/123`
- `routes/users/[name]/profile.html` → `/users/john/profile`

### Rejs JavaScript Runtime

#### CLI Options
- `--timeout=ms` - Set execution timeout in milliseconds
- `--version` - Show version information

#### Built-in Modules

**Core Modules:**
- `console` - Console output with colors
- `process` - Process information and control
- `fs` - File system operations (sync and async)
- `path` - Path manipulation utilities
- `child_process` - Execute system commands

**Network Modules:**
- `fetch` - HTTP client with Promise support

**Global Objects:**
- `process` - Process information and control (Node.js compatible)
- `__filename` - Current script absolute path
- `__dirname` - Current script directory
- `setTimeout`, `setInterval`, `clearTimeout`, `clearInterval`

#### Module System
```javascript
// Import built-in module
var fs = require('fs');

// Import local module
var math = require('./lib/math.js');

// Import from node_modules
var lodash = require('lodash');
```

## 🛠️ Advanced Features

### Background Mode & Process Management

Run Redi server in background mode (daemon) with automatic logging:

```bash
# Start server in background with logging
redi --root=mysite --port=8080 --log=server.log

# Server runs in background, logs to server.log
# PID saved to server.log.pid for process management

# Stop the background server
kill $(cat server.log.pid)
```

### Session-Based JavaScript State

Redi maintains JavaScript engine state per client session, enabling:
- **Persistent variables** across requests from the same client
- **Session-specific data** storage
- **Consistent user experience** within a session

```javascript
// routes/api/counter.js
// Each client gets their own counter state
if (!global.counter) {
    global.counter = 0;
}

exports.get = function(req, res, next) {
    global.counter++;
    res.json({ 
        counter: global.counter,
        sessionId: req.sessionId // Automatically generated
    });
};
```

### Server-Side Rendering with Layouts

**routes/_layout/base.html:**
```html
<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}} - My Site</title>
</head>
<body>
    <nav><!-- Navigation --></nav>
    {{.Content}}
    <footer><!-- Footer --></footer>
</body>
</html>
```

**routes/page.html:**
```html
{{layout "base"}}
<script @server>
res.render({
    Title: "Page Title"
});
</script>

<h1>Page Content</h1>
```

### HTTP Method Testing

Redi provides comprehensive HTTP method support with built-in testing capabilities:

**routes/method-example.js:**
```javascript
exports.get = function(req, res, next) {
    res.render({
        Title: "HTTP Method Testing",
        method: req.method,
        message: "GET request received successfully"
    });
};

exports.post = function(req, res, next) {
    var body = req.body ? JSON.parse(req.body) : {};
    res.render({
        Title: "HTTP Method Testing", 
        method: req.method,
        message: "POST request received",
        receivedData: JSON.stringify(body, null, 2)
    });
};

exports.put = function(req, res, next) {
    var body = req.body ? JSON.parse(req.body) : {};
    res.render({
        Title: "HTTP Method Testing",
        method: req.method, 
        message: "PUT request received",
        receivedData: JSON.stringify(body, null, 2)
    });
};

exports.delete = function(req, res, next) {
    res.render({
        Title: "HTTP Method Testing",
        method: req.method,
        message: "DELETE request received"
    });
};
```

### Child Process Execution

```javascript
var child_process = require('child_process');

// Synchronous execution
var output = child_process.execSync('ls -la');
console.log(output.toString());

// Asynchronous execution
child_process.exec('git status', function(error, stdout, stderr) {
    if (error) {
        console.error('Error:', error);
        return;
    }
    console.log('Output:', stdout);
});
```

## 🧪 Testing

Run the comprehensive test suite:

```bash
# Run all tests (includes integration, API, and module tests)
make test

# Run specific test categories
make test-unit           # Unit tests only
make test-integration    # Integration tests only
make test-api           # API endpoint tests only

# Manual testing with rejs runtime
./rejs fixtures/routes/tests/process_test.js
./rejs --timeout=10000 fixtures/routes/tests/fetch_test.js
```

### Test Coverage

Redi includes extensive testing for:
- ✅ **Server Integration**: HTTP routing, static files, template rendering
- ✅ **API Endpoints**: JavaScript execution, JSON responses, error handling  
- ✅ **Dynamic Routing**: Parameter extraction, URL patterns
- ✅ **JavaScript Engine**: Module system, require paths, session management
- ✅ **Template System**: Layout processing, data binding, error handling
- ✅ **Cross-Platform**: Windows, Linux, macOS compatibility
- ✅ **Built-in Modules**: fs, path, process, fetch, child_process

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Links

- [GitHub Repository](https://github.com/rediwo/redi)
- [Releases](https://github.com/rediwo/redi/releases)
- [Issues](https://github.com/rediwo/redi/issues)

---

Made with ❤️ using Go and JavaScript