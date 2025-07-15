# Redi - Modern Web Server & JavaScript Runtime

Redi is a Go-based web development toolkit that provides a dynamic web server (`redi`), a Node.js-compatible JavaScript runtime (`rejs`), and build tools (`redi-build`) for creating embedded applications and desktop apps.

## 🚀 Features

### Redi Web Server
- **Dynamic Routing**: Automatic route discovery from filesystem structure with `[param]` syntax
- **JavaScript API Endpoints**: Execute `.js` files server-side for API routes
- **HTML Template Rendering**: Process `.html` files with Go templates and server-side JavaScript
- **Markdown Support**: Automatic `.md` to HTML conversion with Goldmark parser
- **Svelte Support**: Server-side Svelte compilation with automatic runtime injection and enhanced import system
- **Compilation Cache**: Persistent disk cache for Svelte components with significant performance improvement
- **Pre-build Support**: Vite-like pre-compilation of all components for production deployments
- **Template Layouts**: Nested layouts with `{{layout 'name'}}` syntax
- **Background Mode**: Run server as daemon with `--log` parameter (nohup-like behavior)
- **Session-based VM Management**: Consistent JavaScript state across requests per client
- **Static File Serving**: Efficient serving from `public/` directory
- **Cross-Platform**: Works on Linux, macOS, and Windows
- **JavaScript Engine Pooling**: High-performance concurrent request handling
- **Vimesh Style Integration**: Lightweight CSS generation for Svelte components and HTML templates (enabled by default)
- **Gzip Compression**: Automatic response compression for better performance
- **Custom Error Pages**: Beautiful error pages with template support (404, 500, etc.)

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

## 📊 Framework Support

### Component Frameworks
- **Svelte**: Full support with server-side compilation, enhanced imports, and Vimesh Style integration

Redi focuses on Svelte as the primary component framework, providing deep integration and optimal performance.

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
    <link rel="stylesheet" href="/css/style.css">
</head>
<body>
    <div class="container">
        <h1>{{.Title}}</h1>
        <p>Current time: {{.time}}</p>
        <p>Request method: {{.method}}</p>
        
        {{if .data}}
        <div class="data-display">
            <h2>Submitted Data:</h2>
            <pre>{{.data}}</pre>
        </div>
        {{end}}
    </div>
</body>
</html>
```

**routes/hello.svelte:** (Svelte component example)
```svelte
<script>
    // Import other components
    import Button from './Button.svelte';
    
    // Import assets (transforms to URL strings)
    import styles from '/css/style.css';
    import logo from '/images/logo.svg';
    
    export let name = 'World';
    let count = 0;
    
    function increment() {
        count += 1;
    }
</script>

<style>
    /* Styles are scoped by default in Svelte */
    .greeting {
        font-size: 2rem;
        color: #3b82f6;
    }
</style>

<svelte:head>
    <link rel="stylesheet" href={styles}>
</svelte:head>

<div class="container mx-auto p-8">
    <img src={logo} alt="Logo" class="w-16 h-16 mb-4">
    <h1 class="greeting">Hello {name}!</h1>
    <p class="text-gray-600">You've clicked {count} times</p>
    <Button on:click={increment}>
        Click me
    </Button>
</div>
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

# Enable compilation cache (enabled by default)
redi --root=mysite --port=8080 --cache

# Pre-build all Svelte components for production
redi --root=mysite --prebuild

# Pre-build with custom parallel workers (default: 4)
redi --root=mysite --prebuild --prebuild-parallel=8

# Pre-build then start server (production deployment)
redi --root=mysite --prebuild --port=8080

# Clear cache and exit
redi --root=mysite --clear-cache

# Run in background with logging
redi --root=mysite --port=8080 --log=server.log
```

### Custom Error Pages

Create beautiful error pages in your routes directory:

**routes/404.html:**
```html
{{layout 'base'}}

<div class="min-h-screen flex items-center justify-center bg-gray-100">
    <div class="text-center">
        <h1 class="text-6xl font-bold text-gray-800">{{.Status}}</h1>
        <h2 class="text-2xl font-semibold text-gray-600 mt-4">{{.StatusText}}</h2>
        <p class="text-gray-500 mt-2">{{.Message}}</p>
        <div class="mt-6">
            <p class="text-sm text-gray-400">Requested: {{.Path}}</p>
            <a href="/" class="mt-4 inline-block bg-blue-500 text-white px-6 py-2 rounded hover:bg-blue-600">
                Go Home
            </a>
        </div>
    </div>
</div>
```

Error pages support:
- Template layouts
- Go template syntax
- Vimesh Style CSS utilities
- Custom styling
- Generic pages (4xx.html, 5xx.html)

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
- `--cache` - Enable compilation cache (default: true)
- `--prebuild` - Pre-compile all Svelte components before starting server
- `--prebuild-parallel` - Number of parallel workers for pre-building (default: 4)
- `--clear-cache` - Clear existing cache and exit
- `--log` - Log file path (enables background/daemon mode like nohup)
- `--log-level` - Log level: debug, info, warn, error (default: info)
- `--log-format` - Log format: text, json (default: text)
- `--quiet` - Quiet mode (only ERROR and FATAL messages)
- `--version` - Show version information

#### Directory Structure
- `public/` - Static assets (CSS, JS, images)
- `routes/` - Dynamic routes and API endpoints
- `.redi/` - Cache directory (auto-generated)
  - `cache/` - Compiled component cache
  - `metadata.json` - Cache metadata and statistics

#### Route Types
- `.html` - HTML templates processed with Go templates
- `.js` - JavaScript files for API endpoints and server-side logic
- `.md` - Markdown files auto-converted to HTML with Goldmark
- `.svelte` - Svelte components compiled server-side with automatic runtime injection

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

### Custom Error Pages

Redi supports custom error pages that integrate with your site's design:

**routes/404.html:**
```html
{{layout "base"}}

<div class="min-h-screen flex items-center justify-center">
    <div class="text-center">
        <h1 class="text-9xl font-bold text-gray-300">{{.Status}}</h1>
        <h2 class="text-3xl font-semibold text-gray-800 mb-4">{{.StatusText}}</h2>
        <p class="text-gray-600 mb-8">{{.Message}}</p>
        <div class="space-y-2 text-sm text-gray-500">
            <p>Path: <code class="bg-gray-100 px-2 py-1 rounded">{{.Path}}</code></p>
            <p>Method: <code class="bg-gray-100 px-2 py-1 rounded">{{.Method}}</code></p>
        </div>
        <a href="/" class="mt-8 inline-block bg-blue-500 text-white px-6 py-3 rounded hover:bg-blue-600">
            Go Home
        </a>
    </div>
</div>
```

Error pages can be:
- **Specific**: `404.html`, `500.html` for specific status codes
- **Generic**: `4xx.html`, `5xx.html` for error ranges
- **Styled**: Use layouts and Vimesh Style utility classes
- **Dynamic**: Access error data in templates

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

### Svelte Components with Enhanced Import System

Redi provides built-in support for Svelte components with automatic server-side compilation and enhanced import capabilities:

#### Enhanced Import Features
- **Asset Imports**: Import CSS, images, and fonts as URL strings
- **JSON Imports**: Import JSON files as JavaScript objects (automatically inlined for files < 100KB)
- **Automatic Transformation**: Non-Svelte imports are intelligently transformed
- **Flexible Path Resolution**: Supports relative and absolute imports
- **No Build Step**: All imports are resolved and transformed at runtime

**Example with imports:**
```svelte
<script>
    // Import Svelte components (works as before)
    import Card from './Card.svelte';
    import Button from '../shared/Button.svelte';
    
    // Import CSS (transforms to URL string)
    import styles from './component.css';
    
    // Import images (transforms to URL string)
    import logo from '/images/logo.svg';
    import heroImage from './assets/hero.jpg';
    
    // Import JSON data (transforms to JavaScript object)
    import config from './config.json';
    
    // Import fonts (transforms to URL string)
    import customFont from '/fonts/custom.woff2';
    
    // Use the imported assets
    console.log('Styles URL:', styles);  // '/component.css'
    console.log('Logo URL:', logo);      // '/images/logo.svg'
    
    // JSON data is directly available - no fetch needed!
    console.log('Config:', config);      // {name: "app", version: "1.0.0", ...}
    console.log('App name:', config.name);
</script>

<svelte:head>
    <link rel="stylesheet" href={styles}>
    <style>
        @font-face {
            font-family: 'CustomFont';
            src: url({customFont}) format('woff2');
        }
    </style>
</svelte:head>

<main>
    <img src={logo} alt="Logo">
    <img src={heroImage} alt="Hero">
    <Card />
    <Button>Click me</Button>
</main>
```

### Svelte Components with Vimesh Style

Vimesh Style is enabled by default for both Svelte components and HTML templates, providing Tailwind-compatible utility classes with minimal overhead.

**Configuration (optional - already enabled by default):**
```go
import "github.com/rediwo/redi/handlers"

// Vimesh Style is enabled by default, but can be configured:
svelteConfig := &handlers.SvelteConfig{
    MinifyRuntime:    true,
    MinifyComponents: true,
    MinifyCSS:       true,
    UseExternalRuntime: true,
    VimeshStyle: &utils.VimeshStyleConfig{
        Enable: true,  // Default: true
    },
    VimeshStylePath: "/svelte-vimesh-style.js",
}

// Register Svelte handler with router
svelteHandler := handlers.NewSvelteHandlerWithRouter(fs, svelteConfig, router)
```

**routes/app.svelte:**
```svelte
<script>
    let items = ['Apple', 'Banana', 'Orange'];
    let newItem = '';
    
    function addItem() {
        if (newItem) {
            items = [...items, newItem];
            newItem = '';
        }
    }
</script>

<!-- Using Vimesh Style utility classes -->
<div class="max-w-md mx-auto p-6 bg-white rounded-lg shadow-md">
    <h1 class="text-2xl font-bold mb-4">Shopping List</h1>
    
    <div class="flex gap-2 mb-4">
        <input 
            bind:value={newItem}
            class="flex-1 px-3 py-2 border rounded"
            placeholder="Add item..."
            on:keydown={(e) => e.key === 'Enter' && addItem()}
        />
        <button 
            on:click={addItem}
            class="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600">
            Add
        </button>
    </div>
    
    <ul class="space-y-2">
        {#each items as item}
            <li class="p-2 bg-gray-100 rounded">{item}</li>
        {/each}
    </ul>
</div>
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

## 📝 Logging System

Redi includes a comprehensive logging system with multiple levels and formats:

### Log Levels
- **DEBUG**: Detailed diagnostic information (route registration, cache operations)
- **INFO**: General informational messages (server startup, compilation progress)
- **WARN**: Warning messages (cache failures, deprecated features)
- **ERROR**: Error messages (compilation failures, file not found)
- **FATAL**: Fatal errors that cause program termination

### Log Formats
- **Text**: Human-readable colored output for development
- **JSON**: Structured logging for production and log analysis

### Usage Examples
```bash
# Different log levels
redi --root=mysite --log-level=debug    # Show all messages
redi --root=mysite --log-level=info     # Default level
redi --root=mysite --log-level=warn     # Only warnings and errors
redi --root=mysite --log-level=error    # Only errors

# Different formats
redi --root=mysite --log-format=text    # Colored text (default)
redi --root=mysite --log-format=json    # JSON format

# Quiet mode (only errors and fatal)
redi --root=mysite --quiet

# Combine with file logging
redi --root=mysite --log=server.log --log-format=json --log-level=info
```

### Structured Logging
The logging system supports structured fields for better analysis:
```json
{"level":"INFO","message":"Server starting","port":8080,"root":"mysite","timestamp":"2025-07-14T16:40:00+08:00","version":"1.0.0"}
{"level":"DEBUG","message":"Registered route","file":"routes/index.html","path":"/","type":"html","timestamp":"2025-07-14T16:40:01+08:00"}
{"level":"INFO","message":"Pre-build completed","compiled":15,"duration":"2.5s","errors":0,"timestamp":"2025-07-14T16:40:03+08:00"}
```

## ⚡ Performance

### Compilation Cache

Redi includes a sophisticated caching system for optimal Svelte component performance:

#### Cache Features
- **Persistent Cache**: Components cached to disk in `.redi/cache/` directory
- **Performance Boost**: Significant improvement in component loading speed after initial compilation
- **Cache Headers**: `X-Svelte-Cached` header indicates whether component was served from cache
- **Dependency Tracking**: Automatically invalidates cache when components or dependencies change
- **Metadata Management**: Tracks access patterns, compilation times, and cache priorities

#### Pre-build Support (Vite-like)
```bash
# Pre-compile all Svelte components for production
redi --root=mysite --prebuild

# Use parallel workers for faster compilation
redi --root=mysite --prebuild --prebuild-parallel=8

# Production deployment: pre-build then start server
redi --root=mysite --prebuild --port=8080
```

#### Cache Management
```bash
# Clear all cached components
redi --root=mysite --clear-cache

# Check cache status via response headers
curl -I http://localhost:8080/svelte/_lib/DataTable
# X-Svelte-Cached: true  (served from cache)
# X-Svelte-Cached: false (freshly compiled)
```

#### Performance Benefits
- **Cold Start**: First request compiles component and caches result
- **Warm Cache**: Subsequent requests served from memory (fastest)
- **Persistent Cache**: After server restart, components loaded from disk cache
- **Production Ready**: Pre-build all components during deployment for optimal performance

#### Cache Scenarios
1. **Development**: Cache automatically manages compilation for fast iteration
2. **Production**: Pre-build components during deployment for immediate availability
3. **CI/CD**: Clear cache and pre-build as part of deployment pipeline

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
- ✅ **Svelte Support**: Server-side compilation, runtime injection, caching
- ✅ **Vimesh Style**: CSS extraction, runtime generation, Svelte integration
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
│           ├── types.go       # Builder interfaces
│           ├── cli.go         # JavaScript CLI builder
│           ├── server.go      # Server app builder
│           ├── standalone.go  # Standalone app builder
│           ├── app.go         # Desktop app builder
│           ├── templates/     # Embedded templates
│           └── utils.go       # Utility functions
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