// Svelte Async Component Loading Library
// This library provides client-side utilities for dynamic component loading

(function(global) {
    'use strict';

    // Global component registry for async loaded components
    global.__svelteAsyncComponents = global.__svelteAsyncComponents || {};
    
    // Component loading cache
    var loadingCache = {};
    var loadedComponents = {};
    
    // CSS injection helper
    function injectCSS(css, componentName) {
        if (!css) return;
        
        var styleId = 'svelte-async-' + componentName;
        var existingStyle = document.getElementById(styleId);
        
        if (!existingStyle) {
            var style = document.createElement('style');
            style.id = styleId;
            style.type = 'text/css';
            style.innerHTML = css;
            document.head.appendChild(style);
        }
    }
    
    // Resolve component path - handle relative paths
    function resolveComponentPath(componentPath) {
        // If it's a relative path starting with './', resolve it based on current page
        if (componentPath.startsWith('./')) {
            var currentPath = window.location.pathname;
            // Remove filename part, keep directory
            var basePath = currentPath.substring(0, currentPath.lastIndexOf('/'));
            return basePath + '/' + componentPath.substring(2);
        }
        
        // If it's an absolute path starting with '/', use as-is
        if (componentPath.startsWith('/')) {
            return componentPath;
        }
        
        // For backwards compatibility, assume it's a component name in current framework
        var currentPath = window.location.pathname;
        var pathParts = currentPath.split('/');
        if (pathParts.length >= 2 && pathParts[1]) {
            // e.g., /svelte/page -> /svelte/_lib/ComponentName
            return '/' + pathParts[1] + '/_lib/' + componentPath;
        }
        
        // Fallback
        return componentPath;
    }

    // Load and register a component
    function loadComponent(componentPath, options) {
        options = options || {};
        
        // Check if already loaded
        if (loadedComponents[componentPath]) {
            return Promise.resolve(loadedComponents[componentPath]);
        }
        
        // Check if already loading
        if (loadingCache[componentPath]) {
            return loadingCache[componentPath];
        }
        
        // Resolve the component path
        var resolvedPath = resolveComponentPath(componentPath);
        
        // Start loading
        var promise = fetch(resolvedPath, {
            method: 'GET',
            headers: {
                'Accept': 'application/json',
                'Cache-Control': 'public, max-age=3600'
            }
        })
        .then(function(response) {
            if (!response.ok) {
                throw new Error('Failed to load component: ' + response.status);
            }
            return response.json();
        })
        .then(function(data) {
            if (!data.success) {
                throw new Error(data.error || 'Component loading failed');
            }
            
            // Inject CSS
            injectCSS(data.css, data.component);
            
            // Load dependencies first
            if (data.dependencies && data.dependencies.length > 0) {
                data.dependencies.forEach(function(dep) {
                    injectCSS(dep.css, dep.component);
                    
                    // Execute dependency component code
                    try {
                        var depCode = '(function() {\n' + dep.js + '\nreturn ' + dep.className + ';\n})()';
                        global.__svelteAsyncComponents[dep.className] = eval(depCode);
                    } catch (e) {
                        console.error('Failed to execute dependency component:', dep.component, e);
                    }
                });
            }
            
            // Execute main component code
            try {
                // Create a function that executes in the global context with all Svelte runtime functions available
                var componentCode = '(function() {\n' + data.js + '\nreturn ' + data.className + ';\n})()';
                var ComponentClass = eval(componentCode);
                
                // Register component
                global.__svelteAsyncComponents[data.className] = ComponentClass;
                loadedComponents[componentPath] = ComponentClass;
                
                return ComponentClass;
            } catch (e) {
                console.error('Failed to execute component:', data.component, e);
                throw e;
            }
        })
        .catch(function(error) {
            // Remove from loading cache on error
            delete loadingCache[componentPath];
            throw error;
        });
        
        // Cache the loading promise
        loadingCache[componentPath] = promise;
        
        return promise;
    }
    
    // Lazy component wrapper
    function lazy(importFn) {
        return function LazyComponent(options) {
            var target = options.target;
            var props = options.props || {};
            
            var loadingComponent = null;
            var actualComponent = null;
            var mounted = false;
            
            // Create loading placeholder
            if (target) {
                target.innerHTML = '<div class="svelte-async-loading">Loading...</div>';
            }
            
            // Load the component
            importFn()
                .then(function(ComponentClass) {
                    if (!mounted) return; // Component was destroyed before loading completed
                    
                    // Remove loading placeholder
                    if (target) {
                        target.innerHTML = '';
                    }
                    
                    // Create actual component
                    actualComponent = new ComponentClass({
                        target: target,
                        props: props
                    });
                })
                .catch(function(error) {
                    if (!mounted) return;
                    
                    console.error('Failed to load lazy component:', error);
                    if (target) {
                        target.innerHTML = '<div class="svelte-async-error">Failed to load component</div>';
                    }
                });
            
            mounted = true;
            
            // Return component-like interface
            return {
                $destroy: function() {
                    mounted = false;
                    if (actualComponent && actualComponent.$destroy) {
                        actualComponent.$destroy();
                    }
                },
                $set: function(newProps) {
                    if (actualComponent && actualComponent.$set) {
                        actualComponent.$set(newProps);
                    }
                }
            };
        };
    }
    
    // Dynamic import function
    function dynamicImport(componentPath) {
        return loadComponent(componentPath);
    }
    
    // Export API
    global.SvelteAsync = {
        loadComponent: loadComponent,
        lazy: lazy,
        import: dynamicImport,
        
        // Utility function for creating lazy imports
        createLazyImport: function(componentPath) {
            return function() {
                return dynamicImport(componentPath);
            };
        }
    };
    
    // Add some default CSS for loading states
    injectCSS(`
        .svelte-async-loading {
            padding: 20px;
            text-align: center;
            color: #666;
            font-style: italic;
        }
        
        .svelte-async-error {
            padding: 20px;
            text-align: center;
            color: #d32f2f;
            background: #ffebee;
            border: 1px solid #ffcdd2;
            border-radius: 4px;
        }
    `, 'async-styles');
    
})(typeof window !== 'undefined' ? window : global);