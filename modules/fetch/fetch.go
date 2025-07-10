package fetch

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	js "github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/eventloop"
	"github.com/rediwo/redi/registry"
)

// FetchResponse represents the response from a fetch request
type FetchResponse struct {
	Status     int
	StatusText string
	Headers    map[string]string
	Body       string
	URL        string
}

// FetchModule provides Promise-based fetch functionality for JavaScript environments
type FetchModule struct {
	loop *eventloop.EventLoop
}

// init registers the fetch module automatically
func init() {
	registry.RegisterModule("fetch", initFetchModule)
}

// initFetchModule initializes the fetch module
func initFetchModule(config registry.ModuleConfig) error {
	if config.EventLoop != nil && config.VM != nil {
		fetchModule := NewFetchModule(config.EventLoop)
		// Register as global object
		fetchModule.RegisterGlobal(config.VM)
		// Also register as require module
		config.Registry.RegisterNativeModule("fetch", func(vm *js.Runtime, module *js.Object) {
			fetchModule.RegisterGlobal(vm)
			// Return the fetch object for require
			fetchObj := vm.Get("fetch")
			module.Set("exports", fetchObj)
		})
	}
	return nil
}

// NewFetchModule creates a new fetch module instance with event loop
func NewFetchModule(loop *eventloop.EventLoop) *FetchModule {
	return &FetchModule{
		loop: loop,
	}
}

// RegisterGlobal registers the fetch function in the JavaScript VM
func (fm *FetchModule) RegisterGlobal(vm *js.Runtime) {
	vm.Set("fetch", fm.createFetchFunction(vm))
}

// createFetchFunction creates the JavaScript fetch function that returns a Promise
func (fm *FetchModule) createFetchFunction(vm *js.Runtime) func(call js.FunctionCall) js.Value {
	return func(call js.FunctionCall) js.Value {
		if len(call.Arguments) == 0 {
			panic(vm.NewTypeError("fetch requires at least 1 argument"))
		}
		
		url := call.Arguments[0].String()
		var options map[string]interface{}
		
		if len(call.Arguments) > 1 && !js.IsUndefined(call.Arguments[1]) && !js.IsNull(call.Arguments[1]) {
			optionsValue := call.Arguments[1].Export()
			if optMap, ok := optionsValue.(map[string]interface{}); ok {
				options = optMap
			}
		}
		
		if options == nil {
			options = make(map[string]interface{})
		}
		
		promise, resolve, reject := vm.NewPromise()
		
		// Use a single goroutine for the entire operation to reduce complexity
		go func() {
			// Perform HTTP request and immediately resolve/reject within this goroutine
			fetchResp, err := fm.performHTTPRequest(url, options)
			
			// Schedule resolution in the event loop to ensure proper VM context
			fm.loop.RunOnLoop(func(vm *js.Runtime) {
				if err != nil {
					reject(fm.createErrorResponse(err.Error()))
				} else {
					resolve(vm.ToValue(fm.createFetchResponseObject(fetchResp)(vm)))
				}
			})
		}()
		
		return vm.ToValue(promise)
	}
}

// performHTTPRequest executes the HTTP request synchronously and returns the response or error
func (fm *FetchModule) performHTTPRequest(url string, options map[string]interface{}) (*FetchResponse, error) {
	// Use a unified HTTP client with optimized settings
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   10,
			IdleConnTimeout:       30 * time.Second,
			DisableKeepAlives:     false,
			ResponseHeaderTimeout: 5 * time.Second,
		},
	}

	// Parse options
	method := "GET"
	var body []byte
	headers := make(map[string]string)

	if options != nil {
		if m, ok := options["method"].(string); ok {
			method = strings.ToUpper(m)
		}

		if b, ok := options["body"].(string); ok {
			body = []byte(b)
		}

		if h, ok := options["headers"].(map[string]interface{}); ok {
			for key, value := range h {
				if strValue, ok := value.(string); ok {
					headers[key] = strValue
				}
			}
		}
	}

	// Create request
	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(method, url, bytes.NewReader(body))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, err
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Set default Content-Type for POST requests with body
	if method == "POST" && body != nil && req.Header.Get("Content-Type") == "" {
		if fm.isJSON(string(body)) {
			req.Header.Set("Content-Type", "application/json")
		} else {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	}

	// Perform request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Create response headers map
	respHeaders := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			respHeaders[key] = values[0]
		}
	}

	// Create and return fetch response object
	fetchResp := &FetchResponse{
		Status:     resp.StatusCode,
		StatusText: resp.Status,
		Headers:    respHeaders,
		Body:       string(respBody),
		URL:        url,
	}

	return fetchResp, nil
}

// createFetchResponseObject creates the JavaScript response object
func (fm *FetchModule) createFetchResponseObject(fetchResp *FetchResponse) func(*js.Runtime) map[string]interface{} {
	return func(vm *js.Runtime) map[string]interface{} {
		return map[string]interface{}{
			"ok":         fetchResp.Status >= 200 && fetchResp.Status < 300,
			"status":     fetchResp.Status,
			"statusText": fetchResp.StatusText,
			"url":        fetchResp.URL,
			"headers": map[string]interface{}{
				"get": func(name string) string {
					return fetchResp.Headers[name]
				},
				"has": func(name string) bool {
					_, exists := fetchResp.Headers[name]
					return exists
				},
			},
			"text": func() *js.Promise {
				promise, resolve, _ := vm.NewPromise()
				resolve(vm.ToValue(fetchResp.Body))
				return promise
			},
			"json": func() *js.Promise {
				promise, resolve, reject := vm.NewPromise()
				var result interface{}
				err := json.Unmarshal([]byte(fetchResp.Body), &result)
				if err != nil {
					reject(vm.NewTypeError("Failed to parse JSON: " + err.Error()))
				} else {
					resolve(vm.ToValue(result))
				}
				return promise
			},
			"_body": fetchResp.Body, // Internal field for debugging
		}
	}
}

// createErrorResponse creates an error response object
func (fm *FetchModule) createErrorResponse(errorMsg string) map[string]interface{} {
	return map[string]interface{}{
		"ok":         false,
		"status":     0,
		"statusText": "Network Error",
		"url":        "",
		"error":      errorMsg,
		"headers": map[string]interface{}{
			"get": func(name string) string { return "" },
			"has": func(name string) bool { return false },
		},
		"text": func() string {
			return ""
		},
		"json": func() interface{} {
			return nil
		},
	}
}

// isJSON checks if a string is valid JSON
func (fm *FetchModule) isJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}