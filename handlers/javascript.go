package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	js "github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/eventloop"
	"github.com/dop251/goja_nodejs/require"
	"github.com/gorilla/mux"

	"github.com/rediwo/redi/filesystem"
	"github.com/rediwo/redi/modules/console"
	"github.com/rediwo/redi/modules/fetch"
	"github.com/rediwo/redi/modules/fs"
	"github.com/rediwo/redi/modules/path"
	"github.com/rediwo/redi/modules/process"
)

type JavaScriptHandler struct {
	fs      filesystem.FileSystem
	version string
}

func NewJavaScriptHandler(fs filesystem.FileSystem) *JavaScriptHandler {
	return NewJavaScriptHandlerWithVersion(fs, "v20.11.0")
}

func NewJavaScriptHandlerWithVersion(fs filesystem.FileSystem, version string) *JavaScriptHandler {
	return &JavaScriptHandler{
		fs:      fs,
		version: version,
	}
}

type ResponseObject struct {
	StatusCode int
	Headers    map[string]string
	Body       interface{}
}

func (jh *JavaScriptHandler) Handle(route Route) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		content, err := jh.fs.ReadFile(route.FilePath)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}

		// Create a dedicated event loop for this request to ensure complete isolation
		loop := eventloop.NewEventLoop()

		// Create dedicated fetch module for this request
		fetchModule := fetch.NewFetchModule(loop)

		// Start the event loop in a separate goroutine
		go loop.Start()
		defer func() {
			loop.Stop()
		}()

		// Track whether response has been sent
		responseSent := false
		var responseMutex sync.Mutex

		// Channel to wait for script completion or response sent
		done := make(chan error, 1)
		timeout := time.After(10 * time.Second)

		// Create response callback that can be called from any goroutine
		responseCallback := make(chan bool, 1)

		reqObj := jh.createRequestObject(r, route)
		resObj := jh.createResponseObjectWithCallback(w, r.URL.Path, func() {
			responseMutex.Lock()
			responseSent = true
			responseMutex.Unlock()
			select {
			case responseCallback <- true:
			default:
				// Channel is full, response already sent
			}
		})

		// Run in the dedicated event loop for complete isolation
		loop.RunOnLoop(func(vm *js.Runtime) {
			defer func() {
				if recovered := recover(); recovered != nil {
					done <- fmt.Errorf("JavaScript execution error: %v", recovered)
				}
			}()

			// Set up require registry with Node.js modules and file loader
			currentDir := filepath.Dir(route.FilePath)
			registry := require.NewRegistry(
				require.WithGlobalFolders(currentDir),
				require.WithLoader(func(name string) ([]byte, error) {
					// Resolve relative paths from the current route's directory
					var filePath string
					if strings.HasPrefix(name, "./") || strings.HasPrefix(name, "../") {
						// Relative path - join with current directory
						filePath = filepath.Join(currentDir, name)
					} else {
						// Module name without path (try in current directory)
						filePath = filepath.Join(currentDir, name)
					}

					// Add .js extension if not present
					if !strings.HasSuffix(filePath, ".js") && !strings.HasSuffix(filePath, ".json") {
						if _, err := jh.fs.Stat(filePath + ".js"); err == nil {
							filePath += ".js"
						} else if _, err := jh.fs.Stat(filePath + ".json"); err == nil {
							filePath += ".json"
						}
					}

					// Security check: ensure file is within the route directory
					if !strings.HasPrefix(filePath, currentDir) {
						return nil, require.ModuleFileDoesNotExistError
					}

					// Read the file using unified filesystem interface
					return jh.fs.ReadFile(filePath)
				}),
			)

			// Register console module with custom printer
			console.Enable(registry)

			// Register fs module with event loop support (with restricted access to route directory)
			fs.EnableWithEventLoopAndFS(registry, jh.fs, filepath.Dir(route.FilePath), loop)

			// Register path module
			path.Enable(registry)

			// Register process module with event loop support
			process.EnableWithEventLoop(registry, loop, jh.version)

			// Enable the registry on the VM
			registry.Enable(vm)

			// Set up console object
			consoleObj := require.Require(vm, "console")
			vm.Set("console", consoleObj)

			vm.Set("req", reqObj)
			vm.Set("res", resObj)

			// Register fetch functionality with dedicated module
			fetchModule.RegisterGlobal(vm)

			// Note: setTimeout, setInterval, etc. are already available in the EventLoop VM
			_, err := vm.RunString(string(content))

			// If there's an error, send it immediately
			if err != nil {
				done <- err
				return
			}

			// Don't immediately complete - wait for async operations
			// The script has finished synchronous execution but may have pending async work
			go func() {
				// Give more time for immediate synchronous responses
				time.Sleep(5 * time.Millisecond)
				responseMutex.Lock()
				sent := responseSent
				responseMutex.Unlock()

				if sent {
					done <- nil
					return
				}

				// Wait for async operations with reasonable timeout
				maxWait := 15 * time.Second
				start := time.Now()
				ticker := time.NewTicker(50 * time.Millisecond)
				defer ticker.Stop()

				for {
					select {
					case <-ticker.C:
						responseMutex.Lock()
						sent := responseSent
						responseMutex.Unlock()

						if sent {
							done <- nil
							return
						}

						if time.Since(start) >= maxWait {
							done <- fmt.Errorf("script completed without sending response")
							return
						}
					default:
						// Non-blocking check to prevent goroutine from hanging
						time.Sleep(10 * time.Millisecond)
					}
				}
			}()
		})

		// Wait for either script completion or response sent
		scriptCompleted := false

		select {
		case err := <-done:
			scriptCompleted = true
			if err != nil {
				responseMutex.Lock()
				if !responseSent {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				responseMutex.Unlock()
				return
			}
		case <-responseCallback:
			// Response was sent from async operation
			return
		case <-timeout:
			responseMutex.Lock()
			if !responseSent {
				http.Error(w, "Request timeout", http.StatusRequestTimeout)
			}
			responseMutex.Unlock()
			return
		}

		// If script completed successfully, wait for async response
		if scriptCompleted {
			select {
			case <-responseCallback:
				// Async response received
				return
			case <-time.After(5 * time.Second):
				responseMutex.Lock()
				if !responseSent {
					http.Error(w, "Script completed without sending response", http.StatusInternalServerError)
				}
				responseMutex.Unlock()
			}
		}
	}
}

func (jh *JavaScriptHandler) createRequestObject(r *http.Request, route Route) map[string]interface{} {
	vars := mux.Vars(r)

	reqObj := map[string]interface{}{
		"method":  r.Method,
		"url":     r.URL.String(),
		"path":    r.URL.Path,
		"query":   r.URL.RawQuery,
		"headers": r.Header,
		"params":  vars,
	}

	if r.Method == "POST" {
		body, _ := io.ReadAll(r.Body)
		reqObj["body"] = string(body)
	}

	return reqObj
}

func (jh *JavaScriptHandler) createResponseObjectWithCallback(w http.ResponseWriter, requestPath string, onResponse func()) map[string]interface{} {
	statusCode := 200
	var responseMutex sync.Mutex
	responseSent := false

	resObj := map[string]interface{}{
		"json": func(data interface{}) {
			responseMutex.Lock()
			defer responseMutex.Unlock()

			if responseSent {
				return // Prevent duplicate responses
			}
			responseSent = true

			w.Header().Set("Content-Type", "application/json")
			if statusCode != 200 {
				w.WriteHeader(statusCode)
			}
			json.NewEncoder(w).Encode(data)
			onResponse()
		},
		"send": func(data interface{}) {
			responseMutex.Lock()
			defer responseMutex.Unlock()

			if responseSent {
				return // Prevent duplicate responses
			}
			responseSent = true

			w.Header().Set("Content-Type", "text/plain")
			if statusCode != 200 {
				w.WriteHeader(statusCode)
			}
			fmt.Fprint(w, data)
			onResponse()
		},
		"status": func(code int) {
			statusCode = code
		},
		"setHeader": func(key, value string) {
			w.Header().Set(key, value)
		},
	}

	return resObj
}
