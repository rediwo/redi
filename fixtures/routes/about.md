# About Our Test Blog

This is a **test blog** built using the **Redi frontend server**. Redi is a lightweight Go-based server that provides:

## Features

- **Static File Serving**: Serves files from the `public` directory
- **Markdown Support**: Automatically converts `.md` files to HTML
- **Server-side JavaScript**: Execute JavaScript on the server with Goja engine
- **Template Layouts**: Support for nested layouts with `{{layout 'base'}}` syntax
- **Dynamic Routing**: Use `[param]` syntax for dynamic route parameters
- **REST API**: Create API endpoints with `.js` files

## How It Works

### Markdown Files
Files with `.md` extension are automatically converted to HTML using the Goldmark parser. This page you're reading is an example of a markdown file being served as HTML.

### HTML Templates
HTML files can include server-side JavaScript in `<script @server>` blocks. The JavaScript has access to `req` and `res` objects, similar to Express.js.

### JavaScript APIs
Files with `.js` extension are executed as API endpoints. They have access to:
- `req` object with method, URL, headers, params, and body
- `res` object with json(), send(), status(), and setHeader() methods

### Layout System
Templates can use `{{layout 'layoutName'}}` to wrap content in layout files located in `routes/_layout/`.

## Example Usage

```bash
# Start the server
./redi --root=/path/to/site --port=8080
```

## Directory Structure

```
site/
├── public/          # Static files (CSS, JS, images)
├── routes/          # Dynamic routes
│   ├── _layout/     # Layout templates
│   ├── index.html   # Homepage
│   ├── about.md     # This markdown file
│   ├── blog/
│   │   └── [id].html # Dynamic blog post route
│   └── api/
│       └── users.js # API endpoint
```

## Technology Stack

- **Go**: Backend server implementation
- **Gorilla Mux**: HTTP routing
- **Goja**: JavaScript engine for server-side execution
- **Goldmark**: Markdown parser
- **html/template**: Go template engine

---

*This site demonstrates all the features of the Redi frontend server in action.*