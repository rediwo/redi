package stream

import (
	"bytes"
	"io"
	"sync"

	js "github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/eventloop"
	"github.com/rediwo/redi/registry"
)

const ModuleName = "stream"

// StreamModule provides Node.js-compatible stream functionality
type StreamModule struct {
	runtime   *js.Runtime
	loop      *eventloop.EventLoop
}

// NewStreamModule creates a new stream module instance
func NewStreamModule(vm *js.Runtime, loop *eventloop.EventLoop) *StreamModule {
	return &StreamModule{
		runtime: vm,
		loop:    loop,
	}
}

// init registers the stream module automatically
func init() {
	registry.RegisterModule("stream", initStreamModule)
}

// initStreamModule initializes the stream module
func initStreamModule(config registry.ModuleConfig) error {
	config.Registry.RegisterNativeModule(ModuleName, func(vm *js.Runtime, module *js.Object) {
		exports := module.Get("exports").(*js.Object)
		sm := NewStreamModule(vm, config.EventLoop)
		sm.registerClasses(vm, exports)
	})
	return nil
}

// registerClasses registers all stream classes
func (sm *StreamModule) registerClasses(vm *js.Runtime, exports *js.Object) {
	// Create EventEmitter class first
	eventEmitterClass := sm.createEventEmitterClass(vm)
	
	// Create stream classes
	readableClass := sm.createReadableStreamClass(vm, eventEmitterClass)
	writableClass := sm.createWritableStreamClass(vm, eventEmitterClass)
	duplexClass := sm.createDuplexStreamClass(vm, readableClass, writableClass)
	transformClass := sm.createTransformStreamClass(vm, duplexClass)
	
	// Export stream classes
	exports.Set("Readable", readableClass)
	exports.Set("Writable", writableClass)
	exports.Set("Duplex", duplexClass)
	exports.Set("Transform", transformClass)
	exports.Set("Stream", readableClass) // Stream is an alias for Readable
	
	// Export PassThrough as a convenience
	exports.Set("PassThrough", sm.createPassThroughClass(vm, transformClass))
}

// createEventEmitterClass creates the EventEmitter base class
func (sm *StreamModule) createEventEmitterClass(vm *js.Runtime) js.Value {
	script := `
(function() {
	function EventEmitter() {
		this._events = {};
		this._maxListeners = 10;
	}
	
	EventEmitter.prototype.on = function(event, listener) {
		if (!this._events[event]) {
			this._events[event] = [];
		}
		this._events[event].push(listener);
		return this;
	};
	
	EventEmitter.prototype.addListener = EventEmitter.prototype.on;
	
	EventEmitter.prototype.once = function(event, listener) {
		var self = this;
		var fired = false;
		function wrapper() {
			if (!fired) {
				fired = true;
				self.removeListener(event, wrapper);
				listener.apply(self, arguments);
			}
		}
		wrapper.listener = listener;
		return this.on(event, wrapper);
	};
	
	EventEmitter.prototype.removeListener = function(event, listener) {
		if (!this._events[event]) return this;
		
		var list = this._events[event];
		for (var i = list.length - 1; i >= 0; i--) {
			if (list[i] === listener || (list[i].listener && list[i].listener === listener)) {
				list.splice(i, 1);
				break;
			}
		}
		
		if (list.length === 0) {
			delete this._events[event];
		}
		
		return this;
	};
	
	EventEmitter.prototype.off = EventEmitter.prototype.removeListener;
	
	EventEmitter.prototype.removeAllListeners = function(event) {
		if (event) {
			delete this._events[event];
		} else {
			this._events = {};
		}
		return this;
	};
	
	EventEmitter.prototype.emit = function(event) {
		if (!this._events[event]) return false;
		
		var args = Array.prototype.slice.call(arguments, 1);
		var listeners = this._events[event].slice();
		var hasListeners = false;
		
		for (var i = 0; i < listeners.length; i++) {
			listeners[i].apply(this, args);
			hasListeners = true;
		}
		
		return hasListeners;
	};
	
	EventEmitter.prototype.listenerCount = function(event) {
		return this._events[event] ? this._events[event].length : 0;
	};
	
	EventEmitter.prototype.setMaxListeners = function(n) {
		this._maxListeners = n;
		return this;
	};
	
	return EventEmitter;
})()
`
	
	val, err := vm.RunString(script)
	if err != nil {
		panic(err)
	}
	return val
}

// createReadableStreamClass creates the Readable stream class
func (sm *StreamModule) createReadableStreamClass(vm *js.Runtime, eventEmitter js.Value) js.Value {
	script := `
(function(EventEmitter) {
	// setImmediate polyfill
	var setImmediate = typeof setImmediate !== 'undefined' ? setImmediate : function(fn) {
		setTimeout(fn, 0);
	};
	function Readable(options) {
		EventEmitter.call(this);
		
		options = options || {};
		this._readableState = {
			buffer: [],
			flowing: null,
			ended: false,
			endEmitted: false,
			reading: false,
			objectMode: options.objectMode || false,
			highWaterMark: options.highWaterMark || 16384,
			length: 0,
			pipes: null,
			pipesCount: 0,
			encoding: options.encoding || null
		};
		
		this.readable = true;
		
		if (options.read) {
			this._read = options.read;
		}
	}
	
	// Inherit from EventEmitter
	Readable.prototype = Object.create(EventEmitter.prototype);
	Readable.prototype.constructor = Readable;
	
	// Override on method to handle 'data' event specially
	Readable.prototype.on = function(event, listener) {
		EventEmitter.prototype.on.call(this, event, listener);
		
		// When 'data' listener is added, automatically start flowing
		if (event === 'data' && this._readableState.flowing !== true) {
			this.resume();
		}
		
		return this;
	};
	
	Readable.prototype._read = function(size) {
		// To be implemented by subclasses
		throw new Error('_read() is not implemented');
	};
	
	Readable.prototype.read = function(size) {
		var state = this._readableState;
		var ret = null;
		
		if (state.length > 0 || state.ended) {
			if (size === undefined || size >= state.length) {
				// Read all
				ret = state.buffer;
				state.buffer = [];
				state.length = 0;
			} else {
				// Read partial
				ret = state.buffer.splice(0, size);
				state.length -= size;
			}
			
			if (!state.objectMode) {
				// Join buffers for non-object mode
				ret = ret.join('');
			}
		}
		
		// Try to read more if needed
		if (state.length === 0 && !state.ended && !state.reading) {
			state.reading = true;
			var self = this;
			// Use setImmediate to avoid blocking
			setImmediate(function() {
				self._read(state.highWaterMark);
				state.reading = false;
			});
		}
		
		if (state.length === 0 && state.ended && !state.endEmitted) {
			state.endEmitted = true;
			this.emit('end');
		}
		
		return ret;
	};
	
	Readable.prototype.push = function(chunk) {
		var state = this._readableState;
		
		if (chunk === null) {
			state.ended = true;
			if (state.length === 0 && !state.endEmitted) {
				state.endEmitted = true;
				var self = this;
				setImmediate(function() {
					self.emit('end');
				});
			}
			return false;
		}
		
		state.buffer.push(chunk);
		state.length += state.objectMode ? 1 : chunk.length;
		state.reading = false; // Clear reading flag after push
		
		if (state.flowing) {
			var self = this;
			setImmediate(function() {
				self._flow();
			});
		} else {
			this.emit('readable');
		}
		
		return state.length < state.highWaterMark;
	};
	
	Readable.prototype._flow = function() {
		var state = this._readableState;
		while (state.flowing && state.length > 0) {
			var chunk = state.buffer.shift();
			state.length -= state.objectMode ? 1 : chunk.length;
			this.emit('data', chunk);
		}
		
		if (state.ended && state.length === 0 && !state.endEmitted) {
			state.endEmitted = true;
			this.emit('end');
		} else if (state.length === 0 && !state.ended && !state.reading) {
			// Need more data, trigger read
			state.reading = true;
			var self = this;
			setImmediate(function() {
				self._read(state.highWaterMark);
				state.reading = false;
			});
		}
	};
	
	Readable.prototype.pipe = function(dest, options) {
		var source = this;
		var state = this._readableState;
		
		if (!state.pipes) {
			state.pipes = [];
		}
		state.pipes.push(dest);
		state.pipesCount += 1;
		
		var endFn = function() {
			dest.end();
		};
		
		source.on('data', function(chunk) {
			var ret = dest.write(chunk);
			if (ret === false) {
				source.pause();
			}
		});
		
		dest.on('drain', function() {
			source.resume();
		});
		
		if (!options || options.end !== false) {
			source.on('end', endFn);
		}
		
		// Start flowing if not already
		if (state.flowing !== true) {
			source.resume();
		}
		
		return dest;
	};
	
	Readable.prototype.unpipe = function(dest) {
		var state = this._readableState;
		
		if (state.pipesCount === 0) return this;
		
		if (!dest) {
			// Unpipe all
			state.pipes = null;
			state.pipesCount = 0;
		} else {
			// Unpipe specific destination
			var index = state.pipes.indexOf(dest);
			if (index !== -1) {
				state.pipes.splice(index, 1);
				state.pipesCount -= 1;
			}
		}
		
		return this;
	};
	
	Readable.prototype.pause = function() {
		this._readableState.flowing = false;
		return this;
	};
	
	Readable.prototype.resume = function() {
		var state = this._readableState;
		if (!state.flowing) {
			state.flowing = true;
			var self = this;
			setImmediate(function() {
				self._flow();
				// If buffer is empty and not ended, trigger read
				if (state.length === 0 && !state.ended && !state.reading) {
					state.reading = true;
					self._read(state.highWaterMark);
					state.reading = false;
				}
			});
		}
		return this;
	};
	
	Readable.prototype.setEncoding = function(enc) {
		this._readableState.encoding = enc;
		return this;
	};
	
	return Readable;
})`
	
	val, err := vm.RunString(script)
	if err != nil {
		panic(err)
	}
	
	if fn, ok := js.AssertFunction(val); ok {
		result, err := fn(js.Undefined(), eventEmitter)
		if err != nil {
			panic(err)
		}
		return result
	}
	
	return val
}

// createWritableStreamClass creates the Writable stream class
func (sm *StreamModule) createWritableStreamClass(vm *js.Runtime, eventEmitter js.Value) js.Value {
	script := `
(function(EventEmitter) {
	// setImmediate polyfill
	var setImmediate = typeof setImmediate !== 'undefined' ? setImmediate : function(fn) {
		setTimeout(fn, 0);
	};
	function Writable(options) {
		EventEmitter.call(this);
		
		options = options || {};
		this._writableState = {
			ended: false,
			ending: false,
			finished: false,
			objectMode: options.objectMode || false,
			highWaterMark: options.highWaterMark || 16384,
			needDrain: false,
			writing: false,
			corked: 0,
			bufferSize: 0,
			buffer: []
		};
		
		this.writable = true;
		
		if (options.write) {
			this._write = options.write;
		}
		if (options.final) {
			this._final = options.final;
		}
	}
	
	// Inherit from EventEmitter
	Writable.prototype = Object.create(EventEmitter.prototype);
	Writable.prototype.constructor = Writable;
	
	Writable.prototype._write = function(chunk, encoding, callback) {
		// To be implemented by subclasses
		throw new Error('_write() is not implemented');
	};
	
	Writable.prototype.write = function(chunk, encoding, callback) {
		var state = this._writableState;
		
		if (typeof encoding === 'function') {
			callback = encoding;
			encoding = null;
		}
		
		if (state.ended) {
			var er = new Error('write after end');
			if (callback) callback(er);
			this.emit('error', er);
			return false;
		}
		
		var ret = state.bufferSize < state.highWaterMark;
		state.needDrain = !ret;
		
		if (state.writing || state.corked) {
			state.buffer.push({chunk: chunk, encoding: encoding, callback: callback});
			state.bufferSize += state.objectMode ? 1 : chunk.length;
		} else {
			state.writing = true;
			var self = this;
			this._write(chunk, encoding, function(err) {
				state.writing = false;
				if (callback) callback(err);
				if (err) self.emit('error', err);
				
				// Process buffered writes
				self._clearBuffer();
			});
		}
		
		return ret;
	};
	
	Writable.prototype._clearBuffer = function() {
		var state = this._writableState;
		
		if (state.buffer.length > 0 && !state.writing) {
			var entry = state.buffer.shift();
			state.bufferSize -= state.objectMode ? 1 : entry.chunk.length;
			
			state.writing = true;
			var self = this;
			this._write(entry.chunk, entry.encoding, function(err) {
				state.writing = false;
				if (entry.callback) entry.callback(err);
				if (err) self.emit('error', err);
				
				// Continue processing buffer
				self._clearBuffer();
			});
		} else if (state.needDrain && state.buffer.length === 0 && !state.writing) {
			state.needDrain = false;
			this.emit('drain');
		}
		
		// Check if we're done
		if (state.ending && state.buffer.length === 0 && !state.writing && !state.finished) {
			state.finished = true;
			this.emit('finish');
		}
	};
	
	Writable.prototype.end = function(chunk, encoding, callback) {
		var state = this._writableState;
		
		if (typeof chunk === 'function') {
			callback = chunk;
			chunk = null;
			encoding = null;
		} else if (typeof encoding === 'function') {
			callback = encoding;
			encoding = null;
		}
		
		if (chunk !== null && chunk !== undefined) {
			this.write(chunk, encoding);
		}
		
		state.ending = true;
		
		// If nothing is buffered, finish immediately
		if (state.buffer.length === 0 && !state.writing) {
			state.finished = true;
			if (callback) this.once('finish', callback);
			this.emit('finish');
		} else if (callback) {
			this.once('finish', callback);
		}
		
		return this;
	};
	
	Writable.prototype.cork = function() {
		this._writableState.corked++;
	};
	
	Writable.prototype.uncork = function() {
		var state = this._writableState;
		if (state.corked) {
			state.corked--;
			if (!state.writing && !state.corked && !state.finished && !state.ending) {
				this._clearBuffer();
			}
		}
	};
	
	return Writable;
})`
	
	val, err := vm.RunString(script)
	if err != nil {
		panic(err)
	}
	
	if fn, ok := js.AssertFunction(val); ok {
		result, err := fn(js.Undefined(), eventEmitter)
		if err != nil {
			panic(err)
		}
		return result
	}
	
	return val
}

// createDuplexStreamClass creates the Duplex stream class
func (sm *StreamModule) createDuplexStreamClass(vm *js.Runtime, readable, writable js.Value) js.Value {
	script := `
(function(Readable, Writable) {
	// setImmediate polyfill
	var setImmediate = typeof setImmediate !== 'undefined' ? setImmediate : function(fn) {
		setTimeout(fn, 0);
	};
	function Duplex(options) {
		Readable.call(this, options);
		Writable.call(this, options);
		
		this.allowHalfOpen = options && options.allowHalfOpen !== false;
		
		var self = this;
		this.once('end', function() {
			if (!self.allowHalfOpen) {
				self.end();
			}
		});
	}
	
	// Inherit from Readable
	Duplex.prototype = Object.create(Readable.prototype);
	
	// Mix in Writable methods
	var keys = Object.keys(Writable.prototype);
	for (var i = 0; i < keys.length; i++) {
		var key = keys[i];
		if (!Duplex.prototype[key]) {
			Duplex.prototype[key] = Writable.prototype[key];
		}
	}
	
	Duplex.prototype.constructor = Duplex;
	
	return Duplex;
})`
	
	val, err := vm.RunString(script)
	if err != nil {
		panic(err)
	}
	
	if fn, ok := js.AssertFunction(val); ok {
		result, err := fn(js.Undefined(), readable, writable)
		if err != nil {
			panic(err)
		}
		return result
	}
	
	return val
}

// createTransformStreamClass creates the Transform stream class
func (sm *StreamModule) createTransformStreamClass(vm *js.Runtime, duplex js.Value) js.Value {
	script := `
(function(Duplex) {
	// setImmediate polyfill
	var setImmediate = typeof setImmediate !== 'undefined' ? setImmediate : function(fn) {
		setTimeout(fn, 0);
	};
	function Transform(options) {
		Duplex.call(this, options);
		
		options = options || {};
		
		this._transformState = {
			writechunk: null,
			writeencoding: null,
			writecb: null,
			transforming: false
		};
		
		// Set transform method from options
		if (options.transform) {
			this._transform = options.transform;
		}
		if (options.flush) {
			this._flush = options.flush;
		}
		
		var self = this;
		
		// Override _write
		this._write = function(chunk, encoding, callback) {
			var ts = self._transformState;
			ts.writechunk = chunk;
			ts.writeencoding = encoding;
			ts.writecb = callback;
			
			if (!ts.transforming) {
				ts.transforming = true;
				self._transform(chunk, encoding, function(err, data) {
					ts.transforming = false;
					
					if (err) {
						if (ts.writecb) ts.writecb(err);
						self.emit('error', err);
						return;
					}
					
					if (data !== null && data !== undefined) {
						self.push(data);
					}
					
					if (ts.writecb) {
						ts.writecb();
					}
				});
			}
		};
		
		// Handle end
		this.once('finish', function() {
			if (self._flush) {
				self._flush(function(err, data) {
					if (err) {
						self.emit('error', err);
						return;
					}
					if (data !== null && data !== undefined) {
						self.push(data);
					}
					self.push(null);
				});
			} else {
				self.push(null);
			}
		});
	}
	
	// Inherit from Duplex
	Transform.prototype = Object.create(Duplex.prototype);
	Transform.prototype.constructor = Transform;
	
	Transform.prototype._transform = function(chunk, encoding, callback) {
		// To be implemented by subclasses
		throw new Error('_transform() is not implemented');
	};
	
	return Transform;
})`
	
	val, err := vm.RunString(script)
	if err != nil {
		panic(err)
	}
	
	if fn, ok := js.AssertFunction(val); ok {
		result, err := fn(js.Undefined(), duplex)
		if err != nil {
			panic(err)
		}
		return result
	}
	
	return val
}

// createPassThroughClass creates the PassThrough stream class
func (sm *StreamModule) createPassThroughClass(vm *js.Runtime, transform js.Value) js.Value {
	script := `
(function(Transform) {
	function PassThrough(options) {
		Transform.call(this, options);
	}
	
	// Inherit from Transform
	PassThrough.prototype = Object.create(Transform.prototype);
	PassThrough.prototype.constructor = PassThrough;
	
	PassThrough.prototype._transform = function(chunk, encoding, callback) {
		callback(null, chunk);
	};
	
	return PassThrough;
})`
	
	val, err := vm.RunString(script)
	if err != nil {
		panic(err)
	}
	
	if fn, ok := js.AssertFunction(val); ok {
		result, err := fn(js.Undefined(), transform)
		if err != nil {
			panic(err)
		}
		return result
	}
	
	return val
}

// StreamBuffer is a helper for managing stream data in Go
type StreamBuffer struct {
	mu     sync.Mutex
	buffer bytes.Buffer
	closed bool
}

// Write adds data to the buffer
func (sb *StreamBuffer) Write(data []byte) (int, error) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	
	if sb.closed {
		return 0, io.ErrClosedPipe
	}
	
	return sb.buffer.Write(data)
}

// Read reads data from the buffer
func (sb *StreamBuffer) Read(p []byte) (int, error) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	
	n, err := sb.buffer.Read(p)
	if err == io.EOF && sb.closed {
		return n, io.EOF
	}
	
	return n, err
}

// Close marks the buffer as closed
func (sb *StreamBuffer) Close() error {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	
	sb.closed = true
	return nil
}

// Len returns the current buffer length
func (sb *StreamBuffer) Len() int {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	
	return sb.buffer.Len()
}