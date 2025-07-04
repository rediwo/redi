# Redi - Modern Web Server & JavaScript Runtime

Redi is a Go-based web development toolkit that provides a dynamic web server (`redi`), a Node.js-compatible JavaScript runtime (`rejs`), and build tools (`redi-build`) for creating embedded applications and desktop apps.

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

### Redi Build Tools
- **CLI Applications**: Generate CLI server projects based on redi template
- **Embedded Applications**: Create embedded executable projects with website bundled
- **Desktop Applications**: Generate cross-platform Wails desktop apps with webview
- **Extension System**: Support for extension modules with auto-expansion
- **Template-based Generation**: Source code projects with build automation
- **Third-party Integration**: Clean APIs for building custom applications

## 📦 Installation

### Download Pre-built Binaries

Download the latest release for your platform from the [Releases](https://github.com/rediwo/redi/releases) page. Each archive contains `redi`, `rejs`, and `redi-build`.

```bash
# Extract the archive
tar -xzf redi-v1.0.0-linux-amd64.tar.gz

# Move to PATH (extract individual binaries from archive first)
sudo mv redi-v1.0.0-linux-amd64 /usr/local/bin/redi
sudo mv rejs-v1.0.0-linux-amd64 /usr/local/bin/rejs
sudo mv redi-build-v1.0.0-linux-amd64 /usr/local/bin/redi-build

# Verify installation
redi --version
rejs --version
redi-build --version
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/rediwo/redi.git
cd redi

# Build all tools
make build

# Or build individually
go build -o redi ./cmd/redi
go build -o rejs ./cmd/rejs
go build -o redi-build ./cmd/redi-build
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

### Redi Build Tools

Create CLI applications, embedded projects, and desktop apps:

```bash
# Create CLI application project (based on redi server)
redi-build cli --root=mysite --output=mycli

# Create embedded executable project  
redi-build embed --root=mysite --output=myapp

# Create Wails desktop application project
redi-build app --root=mysite --output=myapp-desktop --name="My Application"

# Use extensions (auto-expands single words to github.com/rediwo/redi-xxx)
redi-build embed --root=mysite --output=myapp --ext=orm,auth,cache

# Use third-party extensions with full URLs
redi-build cli --root=mysite --output=mycli --ext=github.com/example/custom-module

# Use YAML configuration file
redi-build embed --config=build.yaml

# Show help for specific command
redi-build cli --help
redi-build embed --help  
redi-build app --help
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

### Redi Build Tools

#### CLI Commands

**redi-build cli** - Create CLI application projects:
- `--root` - Root directory to include (required)
- `--output` - Output directory name (default: "redi-cli")

**redi-build embed** - Create embedded executable projects:
- `--root` - Root directory to embed (required)
- `--output` - Output directory name (default: "redi-embedded")

**redi-build app** - Create Wails desktop application projects:
- `--root` - Root directory to embed (required)
- `--output` - Output directory name (default: "redi-app")
- `--name` - Application display name (default: "Redi App")

**Global options for all commands:**
- `--ext` - Extension modules (comma-separated, single words auto-expand to github.com/rediwo/redi-xxx)
- `--config` - Configuration file (YAML format)
- `--version` - Show version information
- `--help` - Show help information

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

## 🔧 Build Tools Reference

### Project Types

#### CLI Applications (`redi-build cli`)

Generate standalone CLI server applications based on the redi template:

```bash
# Basic CLI project
redi-build cli --root=mysite --output=my-server

# CLI with extensions
redi-build cli --root=mysite --output=my-server --ext=orm,auth
```

Generated project structure:
```
my-server/
├── main.go           # CLI application entry point
├── go.mod           # Go module with dependencies
├── Makefile         # Build automation
├── bin/             # Compiled binaries
└── mysite/          # Your website files
```

#### Embedded Applications (`redi-build embed`)

Create embedded executable projects with website bundled:

```bash
# Basic embedded project
redi-build embed --root=mysite --output=my-app

# Embedded with extensions
redi-build embed --root=mysite --output=my-app --ext=cache,store
```

Generated project structure:
```
my-app/
├── main.go           # Embedded application with go:embed
├── go.mod           # Go module with dependencies  
├── Makefile         # Build automation
├── bin/             # Compiled binaries
└── mysite/          # Your website files (embedded)
```

#### Desktop Applications (`redi-build app`)

Generate Wails desktop applications with native UI:

```bash
# Basic desktop app
redi-build app --root=mysite --output=my-desktop-app --name="My App"

# Desktop app with extensions
redi-build app --root=mysite --output=my-desktop-app --name="My App" --ext=ui,theme
```

Generated project structure:
```
my-desktop-app/
├── main.go           # Wails application entry point
├── app.go           # Application logic and server management
├── go.mod           # Go module with Wails dependencies
├── frontend/        # Frontend assets
├── embed/           # Embedded website files
└── build/           # Built desktop application
```

### Extension System

#### Official Extensions (Auto-expansion)

Single words automatically expand to official redi extensions:

```bash
# These are equivalent:
--ext=orm,auth,cache
--ext=github.com/rediwo/redi-orm,github.com/rediwo/redi-auth,github.com/rediwo/redi-cache
```

Common official extensions:
- `orm` - Database ORM layer
- `auth` - Authentication and authorization
- `cache` - Caching middleware
- `store` - Key-value storage
- `session` - Session management
- `websocket` - WebSocket support
- `queue` - Background job processing

#### Third-party Extensions

Use full URLs for third-party extensions:

```bash
redi-build cli --root=mysite --ext=github.com/example/custom-auth,github.com/company/logging
```

#### Extension Usage

Extensions use the `import _` pattern for self-registration:

```go
package main

import (
    "github.com/rediwo/redi/runtime"
    "github.com/rediwo/redi/server"
    
    // Extensions auto-register via init() functions
    _ "github.com/rediwo/redi-orm"
    _ "github.com/rediwo/redi-auth"
)
```

### Configuration Files

Use YAML configuration files to avoid repetitive command-line arguments:

**build.yaml:**
```yaml
# Common settings
root: mysite
extensions:
  - orm
  - auth
  - github.com/example/custom

# Command-specific settings
cli:
  output: my-cli-server
  
embed:
  output: my-embedded-app
  
app:
  output: my-desktop-app
  name: "My Application"
```

Usage:
```bash
redi-build cli --config=build.yaml
redi-build embed --config=build.yaml  
redi-build app --config=build.yaml
```

### Build Process

All generated projects include:

1. **Source Code Generation**: Templates filled with your configuration
2. **Dependency Management**: Proper go.mod with required dependencies  
3. **Build Automation**: Makefile with common tasks
4. **Auto-compilation**: Attempts to build the project automatically

#### Build Commands

Each generated project supports standard build commands:

```bash
# Build the project
make build

# Run the project (builds first if needed)
make run

# Clean build artifacts
make clean

# Install to system PATH
make install
```

#### Manual Build Instructions

If auto-compilation fails, projects include helpful instructions:

```bash
# For CLI and embedded projects
cd my-project
go mod tidy
make build

# For desktop applications  
cd my-desktop-app
go mod tidy
wails build
```

## 🛠️ Advanced Features

### Building Embedded Applications

Create embedded executable projects with your website bundled:

```bash
# Build embedded executable project
redi-build embed --root=mysite --output=myapp

# Build and run the embedded app
cd myapp && make build
./bin/myapp --port=8080
```

The embedded project includes:
- Complete website assets and routes
- Redi server runtime  
- Source code and Makefile for customization
- Single executable after building

### Creating Desktop Applications

Generate cross-platform desktop apps using Wails:

```bash
# Create Wails desktop application project
redi-build app --root=mysite --output=myapp-desktop --name="My Application"

# Build the desktop app
cd myapp-desktop
wails build

# Or run in development mode
wails dev
```

Desktop applications include:
- Embedded web server
- Native window with webview
- Menu system with navigation controls
- Start/stop server controls
- Open in external browser option

### Third-Party Integration

Use Redi as a library in your Go applications:

```go
package main

import (
    "github.com/rediwo/redi/server"
    "github.com/rediwo/redi/runtime"
    "github.com/rediwo/redi/cmd/redi-build/builder"
)

func main() {
    // Create and start server
    config := &server.Config{
        Root:    "mysite",
        Port:    8080,
        Version: "1.0.0",
    }
    
    launcher := server.NewLauncher()
    launcher.Start(config)
    
    // Execute JavaScript
    jsConfig, _ := runtime.NewConfig("script.js")
    executor := runtime.NewExecutor()
    executor.Execute(jsConfig)
    
    // Build projects
    cliBuilder := builder.NewCliBuilder()
    cliBuilder.Build(builder.Config{
        Root:       "mysite",
        Output:     "mycli",
        Extensions: []string{"orm", "auth"},
    })
    
    embedBuilder := builder.NewEmbedBuilder()
    embedBuilder.Build(builder.Config{
        Root:       "mysite", 
        Output:     "myapp",
        Extensions: []string{"cache"},
    })
    
    appBuilder := builder.NewAppBuilder()
    appBuilder.Build(builder.Config{
        Root:       "mysite",
        Output:     "mydesktop",
        AppName:    "My App",
        Extensions: []string{"ui"},
    })
}
```

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

# Test build tools
redi-build cli --root=fixtures --output=test-cli
redi-build embed --root=fixtures --output=test-embedded
redi-build app --root=fixtures --output=test-app --name="Test App"
```

### Test Coverage

Redi includes extensive testing for:
- ✅ **Server Integration**: HTTP routing, static files, template rendering
- ✅ **API Endpoints**: JavaScript execution, JSON responses, error handling  
- ✅ **Dynamic Routing**: Parameter extraction, URL patterns
- ✅ **JavaScript Engine**: Module system, require paths, session management
- ✅ **Template System**: Layout processing, data binding, error handling
- ✅ **Build Tools**: CLI, embedded, and desktop app generation
- ✅ **Extension System**: Module auto-expansion and integration
- ✅ **Cross-Platform**: Windows, Linux, macOS compatibility
- ✅ **Built-in Modules**: fs, path, process, fetch, child_process

## 🏗️ Architecture

Redi is built with a modular architecture for easy extension and third-party integration:

### Package Structure

```
redi/
├── cmd/                    # Command-line tools
│   ├── redi/              # Web server CLI
│   ├── rejs/              # JavaScript runtime CLI  
│   └── redi-build/        # Build tools CLI
│       ├── main.go        # Build tool entry point
│       └── builder/       # Builder implementations
│           ├── types.go   # Builder interfaces
│           ├── cli.go     # CLI project builder
│           ├── embed.go   # Embedded app builder
│           ├── app.go     # Desktop app builder
│           ├── templates/ # Embedded templates
│           └── utils.go   # Utility functions
├── server/                # Server management
│   ├── config.go          # Configuration types
│   ├── factory.go         # Server factory
│   ├── launcher.go        # Startup logic
│   └── platform_*.go     # Platform-specific code
├── runtime/               # JavaScript execution
│   ├── config.go          # Runtime configuration
│   ├── executor.go        # JavaScript executor
│   └── version.go         # Version management
├── handlers/              # Request handlers
├── modules/               # JavaScript modules
│   ├── console/           # Console output module
│   ├── fs/                # File system module
│   ├── process/           # Process module
│   └── fetch/             # HTTP client module
├── filesystem/            # File system abstractions
└── fixtures/              # Test website
```

### Design Principles

- **Modularity**: Each package has a single responsibility
- **Extensibility**: Clean interfaces for custom builders and handlers
- **Third-party Friendly**: Simple APIs for integration
- **Platform Support**: Consistent behavior across operating systems
- **No Dependencies**: Embedded applications require no external dependencies

### Core APIs

**Server Management:**
```go
// Create and configure server
config := &server.Config{
    Root:    "site", 
    Port:    8080,
    Version: "1.0.0",
}
launcher := server.NewLauncher()
launcher.Start(config)
```

**JavaScript Execution:**
```go
// Execute JavaScript with runtime
config, _ := runtime.NewConfig("script.js")
executor := runtime.NewExecutor()
exitCode, err := executor.Execute(config)
```

**Building Applications:**
```go
// Import the builder package
import "github.com/rediwo/redi/cmd/redi-build/builder"

// Build CLI application
cliBuilder := builder.NewCliBuilder()
err := cliBuilder.Build(builder.Config{
    Root: "site", Output: "mycli", Extensions: []string{"orm"}
})

// Build embedded executable project
embedBuilder := builder.NewEmbedBuilder()
err := embedBuilder.Build(builder.Config{
    Root: "site", Output: "myapp", Extensions: []string{"cache"}
})

// Build desktop application  
appBuilder := builder.NewAppBuilder()
err := appBuilder.Build(builder.Config{
    Root: "site", Output: "mydesktop", AppName: "My App"
})
```

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