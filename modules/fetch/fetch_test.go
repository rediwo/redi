package fetch

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	js "github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/eventloop"
)

func TestFetchModule_BasicFetch(t *testing.T) {
	// Create a test HTTP server for fetch requests
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		switch r.Method {
		case "GET":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"method": "GET",
				"success": true,
			})
		case "POST":
			body := make([]byte, r.ContentLength)
			r.Body.Read(body)
			
			var data map[string]interface{}
			json.Unmarshal(body, &data)
			
			json.NewEncoder(w).Encode(map[string]interface{}{
				"method": "POST",
				"received": data,
				"success": true,
			})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Method not allowed",
			})
		}
	}))
	defer testServer.Close()

	// Create a dedicated event loop for testing
	loop := eventloop.NewEventLoop()
	go loop.Start()
	defer loop.Stop()

	// Create fetch module
	fetchModule := NewFetchModule(loop)

	// Channel to wait for result
	done := make(chan map[string]interface{}, 1)
	timeout := time.After(10 * time.Second)

	// Run test in event loop
	loop.RunOnLoop(func(vm *js.Runtime) {
		// Register fetch functionality
		fetchModule.RegisterGlobal(vm)

		// Test JavaScript code that uses fetch
		jsCode := `
		fetch('` + testServer.URL + `', {
			method: 'GET'
		})
		.then(function(response) {
			return response.json();
		})
		.then(function(data) {
			return {
				success: true,
				message: "Fetch test completed",
				data: data
			};
		})
		.catch(function(error) {
			return {
				success: false,
				error: error.toString()
			};
		});
		`

		val, err := vm.RunString(jsCode)
		if err != nil {
			t.Fatalf("Failed to execute JavaScript: %v", err)
			return
		}

		// The result should be a Promise
		if promise, ok := val.Export().(*js.Promise); ok {
			// Wait for promise to resolve/reject
			go func() {
				for {
					state := promise.State()
					if state == js.PromiseStateFulfilled {
						result := make(map[string]interface{})
						if obj, ok := promise.Result().Export().(map[string]interface{}); ok {
							result = obj
						}
						select {
						case done <- result:
						default:
						}
						return
					} else if state == js.PromiseStateRejected {
						result := map[string]interface{}{
							"success": false,
							"error": promise.Result().String(),
						}
						select {
						case done <- result:
						default:
						}
						return
					}
					time.Sleep(10 * time.Millisecond)
				}
			}()
		} else {
			// Direct result
			if obj, ok := val.Export().(map[string]interface{}); ok {
				select {
				case done <- obj:
				default:
				}
			}
		}
	})

	// Wait for result
	select {
	case result := <-done:
		// Verify the fetch operation completed successfully
		if !result["success"].(bool) {
			t.Errorf("Fetch operation should succeed, got error: %v", result["error"])
			return
		}

		// Verify response message
		if result["message"].(string) != "Fetch test completed" {
			t.Errorf("Expected message 'Fetch test completed', got %v", result["message"])
		}

		// Verify data was received
		data := result["data"].(map[string]interface{})
		if data["method"].(string) != "GET" {
			t.Errorf("Expected method GET, got %v", data["method"])
		}
	case <-timeout:
		t.Fatal("Test timed out - fetch operation did not complete")
	}
}

func TestFetchModule_Headers(t *testing.T) {
	// Create test server that echoes headers
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Custom-Header", "test-value")
		
		headers := make(map[string]string)
		for key, values := range r.Header {
			if len(values) > 0 {
				headers[key] = values[0]
			}
		}
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"receivedHeaders": headers,
		})
	}))
	defer testServer.Close()

	// Create a dedicated event loop for testing
	loop := eventloop.NewEventLoop()
	go loop.Start()
	defer loop.Stop()

	// Create fetch module
	fetchModule := NewFetchModule(loop)

	// Channel to wait for result
	done := make(chan map[string]interface{}, 1)
	timeout := time.After(10 * time.Second)

	// Run test in event loop
	loop.RunOnLoop(func(vm *js.Runtime) {
		// Register fetch functionality
		fetchModule.RegisterGlobal(vm)

		jsCode := `
		fetch('` + testServer.URL + `', {
			method: 'GET',
			headers: {
				'X-Test-Header': 'test-value',
				'Authorization': 'Bearer token123'
			}
		})
		.then(function(response) {
			return response.json().then(function(data) {
				return {
					status: response.status,
					ok: response.ok,
					contentType: response.headers.get('Content-Type'),
					customHeader: response.headers.get('X-Custom-Header'),
					hasContentType: response.headers.has('Content-Type'),
					hasNonExistent: response.headers.has('X-Non-Existent'),
					data: data
				};
			});
		})
		.catch(function(error) {
			return {
				success: false,
				error: error.toString()
			};
		});
		`

		val, err := vm.RunString(jsCode)
		if err != nil {
			t.Fatalf("Failed to execute JavaScript: %v", err)
			return
		}

		// Handle promise result
		if promise, ok := val.Export().(*js.Promise); ok {
			go func() {
				for {
					state := promise.State()
					if state == js.PromiseStateFulfilled {
						result := make(map[string]interface{})
						if obj, ok := promise.Result().Export().(map[string]interface{}); ok {
							result = obj
						}
						select {
						case done <- result:
						default:
						}
						return
					} else if state == js.PromiseStateRejected {
						result := map[string]interface{}{
							"success": false,
							"error": promise.Result().String(),
						}
						select {
						case done <- result:
						default:
						}
						return
					}
					time.Sleep(10 * time.Millisecond)
				}
			}()
		}
	})

	// Wait for result
	select {
	case result := <-done:
		// Check if we have a success response first
		if result["success"] != nil && !result["success"].(bool) {
			t.Errorf("Fetch operation failed: %v", result["error"])
			return
		}

		// Verify headers functionality
		if result["contentType"] == nil {
			t.Error("contentType field is missing from response")
			return
		}
		if result["contentType"].(string) != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got %v", result["contentType"])
		}
		
		if result["customHeader"].(string) != "test-value" {
			t.Errorf("Expected X-Custom-Header 'test-value', got %v", result["customHeader"])
		}
		
		if !result["hasContentType"].(bool) {
			t.Error("Should have Content-Type header")
		}
		
		if result["hasNonExistent"].(bool) {
			t.Error("Should not have non-existent header")
		}

		// Verify that our sent headers were received
		data := result["data"].(map[string]interface{})
		receivedHeaders := data["receivedHeaders"].(map[string]interface{})
		
		if receivedHeaders["X-Test-Header"] != "test-value" {
			t.Errorf("Expected sent header 'X-Test-Header: test-value', got %v", receivedHeaders["X-Test-Header"])
		}
	case <-timeout:
		t.Fatal("Test timed out - fetch operation did not complete")
	}
}

func TestFetchModule_JSONParsing(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		response := map[string]interface{}{
			"users": []map[string]interface{}{
				{"id": 1, "name": "John", "active": true},
				{"id": 2, "name": "Jane", "active": false},
			},
			"total": 2,
			"page": 1,
		}
		
		json.NewEncoder(w).Encode(response)
	}))
	defer testServer.Close()

	// Create a dedicated event loop for testing
	loop := eventloop.NewEventLoop()
	go loop.Start()
	defer loop.Stop()

	// Create fetch module
	fetchModule := NewFetchModule(loop)

	// Channel to wait for result
	done := make(chan map[string]interface{}, 1)
	timeout := time.After(10 * time.Second)

	// Run test in event loop
	loop.RunOnLoop(func(vm *js.Runtime) {
		// Register fetch functionality
		fetchModule.RegisterGlobal(vm)

		jsCode := `
		fetch('` + testServer.URL + `')
			.then(function(resp) {
				return resp.json().then(function(data) {
					return {
						ok: resp.ok,
						status: resp.status,
						dataType: typeof data,
						usersLength: data.users ? data.users.length : 0,
						total: data.total,
						firstUserName: data.users && data.users[0] ? data.users[0].name : null
					};
				});
			})
			.catch(function(error) {
				return {
					success: false,
					error: error.toString()
				};
			});
		`

		val, err := vm.RunString(jsCode)
		if err != nil {
			t.Fatalf("Failed to execute JavaScript: %v", err)
			return
		}

		// Handle promise result
		if promise, ok := val.Export().(*js.Promise); ok {
			go func() {
				for {
					state := promise.State()
					if state == js.PromiseStateFulfilled {
						result := make(map[string]interface{})
						if obj, ok := promise.Result().Export().(map[string]interface{}); ok {
							result = obj
						}
						select {
						case done <- result:
						default:
						}
						return
					} else if state == js.PromiseStateRejected {
						result := map[string]interface{}{
							"success": false,
							"error": promise.Result().String(),
						}
						select {
						case done <- result:
						default:
						}
						return
					}
					time.Sleep(10 * time.Millisecond)
				}
			}()
		}
	})

	// Wait for result
	select {
	case result := <-done:
		if result["dataType"].(string) != "object" {
			t.Errorf("Expected parsed data to be object, got %v", result["dataType"])
		}
		
		// JavaScript numbers are always float64 in Goja
		usersLength := result["usersLength"].(int64)
		if usersLength != 2 {
			t.Errorf("Expected 2 users, got %v", usersLength)
		}
		
		total := result["total"].(int64)
		if total != 2 {
			t.Errorf("Expected total 2, got %v", total)
		}
		
		if result["firstUserName"].(string) != "John" {
			t.Errorf("Expected first user name 'John', got %v", result["firstUserName"])
		}
	case <-timeout:
		t.Fatal("Test timed out - fetch operation did not complete")
	}
}

func TestFetchModule_ErrorHandling(t *testing.T) {
	// Create a dedicated event loop for testing
	loop := eventloop.NewEventLoop()
	go loop.Start()
	defer loop.Stop()

	// Create fetch module
	fetchModule := NewFetchModule(loop)

	// Channel to wait for result
	done := make(chan map[string]interface{}, 1)
	timeout := time.After(10 * time.Second)

	// Run test in event loop
	loop.RunOnLoop(func(vm *js.Runtime) {
		// Register fetch functionality
		fetchModule.RegisterGlobal(vm)

		// Test fetch to a non-existent URL
		jsCode := `
		fetch('http://nonexistent-server-12345.com')
			.then(function(resp) {
				return {
					success: true,
					status: resp.status
				};
			})
			.catch(function(error) {
				return {
					success: false,
					error: error.toString(),
					hasError: true
				};
			});
		`

		val, err := vm.RunString(jsCode)
		if err != nil {
			t.Fatalf("Failed to execute JavaScript: %v", err)
			return
		}

		// Handle promise result
		if promise, ok := val.Export().(*js.Promise); ok {
			go func() {
				for {
					state := promise.State()
					if state == js.PromiseStateFulfilled {
						result := make(map[string]interface{})
						if obj, ok := promise.Result().Export().(map[string]interface{}); ok {
							result = obj
						}
						select {
						case done <- result:
						default:
						}
						return
					} else if state == js.PromiseStateRejected {
						result := map[string]interface{}{
							"success": false,
							"error": promise.Result().String(),
						}
						select {
						case done <- result:
						default:
						}
						return
					}
					time.Sleep(10 * time.Millisecond)
				}
			}()
		}
	})

	// Wait for result
	select {
	case result := <-done:
		// Should have caught an error
		if result["success"].(bool) {
			t.Error("Expected fetch to fail for non-existent server")
		}
		
		if !result["hasError"].(bool) {
			t.Error("Expected hasError to be true")
		}
		
		if result["error"] == nil || result["error"].(string) == "" {
			t.Error("Expected error message to be present")
		}
	case <-timeout:
		t.Fatal("Test timed out - fetch operation did not complete")
	}
}