// Svelte runtime implementation
(function(global) {
    'use strict';
    
    // Core functions
    const noop = () => {};
    const run = (fn) => fn();
    const blank_object = () => Object.create(null);
    const run_all = (fns) => fns.forEach(run);
    const is_function = (thing) => typeof thing === 'function';
    const safe_not_equal = (a, b) => a != a ? b == b : a !== b || ((a && typeof a === 'object') || typeof a === 'function');
    const is_empty = obj => Object.keys(obj).length === 0;
    const subscribe = (store, ...callbacks) => {
        if (store == null) {
            return noop;
        }
        const unsub = store.subscribe(...callbacks);
        return unsub.unsubscribe ? () => unsub.unsubscribe() : unsub;
    };
    const get_store_value = (store) => {
        let value;
        subscribe(store, _ => value = _)();
        return value;
    };
    const component_subscribe = (component, store, callback) => {
        component.$$.on_destroy.push(subscribe(store, callback));
    };
    const create_slot = (definition, ctx, $$scope, fn) => {
        if (definition) {
            const slot_ctx = get_slot_context(definition, ctx, $$scope, fn);
            return definition[0](slot_ctx);
        }
    };
    const get_slot_context = (definition, ctx, $$scope, fn) => {
        return definition[1] && fn
            ? assign($$scope.ctx.slice(), definition[1](fn(ctx)))
            : $$scope.ctx;
    };
    const get_slot_changes = (definition, $$scope, dirty, fn) => {
        if (definition[2] && fn) {
            const lets = definition[2](fn(dirty));
            if ($$scope.dirty === undefined) {
                return lets;
            }
            if (typeof lets === 'object') {
                const merged = [];
                const len = Math.max($$scope.dirty.length, lets.length);
                for (let i = 0; i < len; i += 1) {
                    merged[i] = $$scope.dirty[i] | lets[i];
                }
                return merged;
            }
            return $$scope.dirty | lets;
        }
        return $$scope.dirty;
    };
    const update_slot_base = (slot, slot_definition, ctx, $$scope, slot_changes, get_slot_context_fn) => {
        if (slot_changes) {
            const slot_context = get_slot_context(slot_definition, ctx, $$scope, get_slot_context_fn);
            slot.p(slot_context, slot_changes);
        }
    };
    const update_slot = (slot, slot_definition, ctx, $$scope, dirty, get_slot_changes_fn, get_slot_context_fn) => {
        const slot_changes = get_slot_changes(slot_definition, $$scope, dirty, get_slot_changes_fn);
        update_slot_base(slot, slot_definition, ctx, $$scope, slot_changes, get_slot_context_fn);
    };
    const exclude_internal_props = (props) => {
        const result = {};
        for (const k in props) if (k[0] !== '$') result[k] = props[k];
        return result;
    };
    const compute_rest_props = (props, keys) => {
        const rest = {};
        keys = new Set(keys);
        for (const k in props) if (!keys.has(k) && k[0] !== '$') rest[k] = props[k];
        return rest;
    };
    const compute_slots = (slots) => {
        const result = {};
        for (const key in slots) {
            result[key] = true;
        }
        return result;
    };
    const assign = (tar, src) => {
        for (const k in src) tar[k] = src[k];
        return tar;
    };
    const set_store_value = (store, ret, value) => {
        store.set(value);
        return ret;
    };
    
    // DOM helpers
    function append(target, node) {
        target.appendChild(node);
    }
    
    function insert(target, node, anchor) {
        target.insertBefore(node, anchor || null);
    }
    
    function detach(node) {
        if (node.parentNode) {
            node.parentNode.removeChild(node);
        }
    }
    
    function destroy_each(iterations, detaching) {
        for (let i = 0; i < iterations.length; i += 1) {
            if (iterations[i]) iterations[i].d(detaching);
        }
    }
    
    function element(name) {
        return document.createElement(name);
    }
    
    function element_is(name, is) {
        return document.createElement(name, { is });
    }
    
    function object_without_properties(obj, exclude) {
        const target = {};
        for (const k in obj) {
            if (Object.prototype.hasOwnProperty.call(obj, k) && exclude.indexOf(k) === -1) {
                target[k] = obj[k];
            }
        }
        return target;
    }
    
    function svg_element(name) {
        return document.createElementNS('http://www.w3.org/2000/svg', name);
    }
    
    function text(data) {
        return document.createTextNode(data);
    }
    
    function space() {
        return text(' ');
    }
    
    function empty() {
        return text('');
    }
    
    function listen(node, event, handler, options) {
        node.addEventListener(event, handler, options);
        return () => node.removeEventListener(event, handler, options);
    }
    
    function prevent_default(fn) {
        return function(event) {
            event.preventDefault();
            return fn.call(this, event);
        };
    }
    
    function stop_propagation(fn) {
        return function(event) {
            event.stopPropagation();
            return fn.call(this, event);
        };
    }
    
    function self(fn) {
        return function(event) {
            if (event.target === this) fn.call(this, event);
        };
    }
    
    function trusted(fn) {
        return function(event) {
            if (event.isTrusted) fn.call(this, event);
        };
    }
    
    function attr(node, attribute, value) {
        if (value == null)
            node.removeAttribute(attribute);
        else if (node.getAttribute(attribute) !== value)
            node.setAttribute(attribute, value);
    }
    
    function set_attributes(node, attributes) {
        const descriptors = Object.getOwnPropertyDescriptors(node.__proto__);
        for (const key in attributes) {
            if (attributes[key] == null) {
                node.removeAttribute(key);
            } else if (key === 'style') {
                node.style.cssText = attributes[key];
            } else if (key === '__value') {
                node.value = node[key] = attributes[key];
            } else if (descriptors[key] && descriptors[key].set) {
                node[key] = attributes[key];
            } else {
                attr(node, key, attributes[key]);
            }
        }
    }
    
    function set_svg_attributes(node, attributes) {
        for (const key in attributes) {
            attr(node, key, attributes[key]);
        }
    }
    
    function set_custom_element_data(node, prop, value) {
        if (prop in node) {
            node[prop] = typeof node[prop] === 'boolean' && value === '' ? true : value;
        } else {
            attr(node, prop, value);
        }
    }
    
    function xlink_attr(node, attribute, value) {
        node.setAttributeNS('http://www.w3.org/1999/xlink', attribute, value);
    }
    
    function get_binding_group_value(group, __value, checked) {
        const value = new Set();
        for (let i = 0; i < group.length; i += 1) {
            if (group[i].checked) value.add(group[i].__value);
        }
        if (!checked) {
            value.delete(__value);
        }
        return Array.from(value);
    }
    
    function to_number(value) {
        return value === '' ? null : +value;
    }
    
    function time_ranges_to_array(ranges) {
        const array = [];
        for (let i = 0; i < ranges.length; i += 1) {
            array.push({ start: ranges.start(i), end: ranges.end(i) });
        }
        return array;
    }
    
    function children(element) {
        return Array.from(element.childNodes);
    }
    
    function claim_element(nodes, name, attributes, svg) {
        for (let i = 0; i < nodes.length; i += 1) {
            const node = nodes[i];
            if (node.nodeName === name) {
                let j = 0;
                const remove = [];
                while (j < node.attributes.length) {
                    const attribute = node.attributes[j++];
                    if (!attributes[attribute.name]) {
                        remove.push(attribute.name);
                    }
                }
                for (let k = 0; k < remove.length; k++) {
                    node.removeAttribute(remove[k]);
                }
                return nodes.splice(i, 1)[0];
            }
        }
        return svg ? svg_element(name) : element(name);
    }
    
    function claim_text(nodes, data) {
        for (let i = 0; i < nodes.length; i += 1) {
            const node = nodes[i];
            if (node.nodeType === 3) {
                node.data = '' + data;
                return nodes.splice(i, 1)[0];
            }
        }
        return text(data);
    }
    
    function claim_space(nodes) {
        return claim_text(nodes, ' ');
    }
    
    function set_data(text, data) {
        data = '' + data;
        if (text.wholeText !== data)
            text.data = data;
    }
    
    function set_input_value(input, value) {
        input.value = value == null ? '' : value;
    }
    
    function set_input_type(input, type) {
        try {
            input.type = type;
        } catch (e) {}
    }
    
    function set_style(node, key, value, important) {
        if (value === null) {
            node.style.removeProperty(key);
        } else {
            node.style.setProperty(key, value, important ? 'important' : '');
        }
    }
    
    function select_option(select, value) {
        for (let i = 0; i < select.options.length; i += 1) {
            const option = select.options[i];
            if (option.__value === value) {
                option.selected = true;
                return;
            }
        }
        select.selectedIndex = -1;
    }
    
    function select_options(select, value) {
        for (let i = 0; i < select.options.length; i += 1) {
            const option = select.options[i];
            option.selected = ~value.indexOf(option.__value);
        }
    }
    
    function select_value(select) {
        const selected_option = select.querySelector(':checked') || select.options[0];
        return selected_option && selected_option.__value;
    }
    
    function select_multiple_value(select) {
        return [].map.call(select.querySelectorAll(':checked'), option => option.__value);
    }
    
    function custom_event(type, detail, { bubbles = false, cancelable = false } = {}) {
        const e = document.createEvent('CustomEvent');
        e.initCustomEvent(type, bubbles, cancelable, detail);
        return e;
    }
    
    function query_selector_all(selector, parent = document.body) {
        return Array.from(parent.querySelectorAll(selector));
    }
    
    // Component system
    let current_component;
    
    function set_current_component(component) {
        current_component = component;
    }
    
    function get_current_component() {
        if (!current_component) throw new Error('Function called outside component initialization');
        return current_component;
    }
    
    function onMount(fn) {
        get_current_component().$$.on_mount.push(fn);
    }
    
    function onDestroy(fn) {
        get_current_component().$$.on_destroy.push(fn);
    }
    
    function createEventDispatcher() {
        const component = get_current_component();
        return (type, detail, { cancelable = false } = {}) => {
            const callbacks = component.$$.callbacks[type];
            if (callbacks) {
                const event = custom_event(type, detail, { cancelable });
                callbacks.slice().forEach(fn => {
                    fn.call(component, event);
                });
                return !event.defaultPrevented;
            }
            return true;
        };
    }
    
    function setContext(key, context) {
        get_current_component().$$.context.set(key, context);
        return context;
    }
    
    function getContext(key) {
        return get_current_component().$$.context.get(key);
    }
    
    function getAllContexts() {
        return get_current_component().$$.context;
    }
    
    function hasContext(key) {
        return get_current_component().$$.context.has(key);
    }
    
    function bubble(component, event) {
        const callbacks = component.$$.callbacks[event.type];
        if (callbacks) {
            callbacks.slice().forEach(fn => fn.call(this, event));
        }
    }
    
    const dirty_components = [];
    const binding_callbacks = [];
    const render_callbacks = [];
    const flush_callbacks = [];
    const resolved_promise = Promise.resolve();
    let update_scheduled = false;
    
    function schedule_update() {
        if (!update_scheduled) {
            update_scheduled = true;
            resolved_promise.then(flush);
        }
    }
    
    function tick() {
        schedule_update();
        return resolved_promise;
    }
    
    function add_render_callback(fn) {
        render_callbacks.push(fn);
    }
    
    function add_flush_callback(fn) {
        flush_callbacks.push(fn);
    }
    
    const seen_callbacks = new Set();
    let flushidx = 0;
    
    function flush() {
        if (flushidx !== 0) {
            return;
        }
        const saved_component = current_component;
        do {
            try {
                while (flushidx < dirty_components.length) {
                    const component = dirty_components[flushidx];
                    flushidx++;
                    set_current_component(component);
                    update(component.$$);
                }
            } catch (e) {
                dirty_components.length = 0;
                flushidx = 0;
                throw e;
            }
            set_current_component(null);
            dirty_components.length = 0;
            flushidx = 0;
            while (binding_callbacks.length) binding_callbacks.pop()();
            for (let i = 0; i < render_callbacks.length; i += 1) {
                const callback = render_callbacks[i];
                if (!seen_callbacks.has(callback)) {
                    seen_callbacks.add(callback);
                    callback();
                }
            }
            render_callbacks.length = 0;
        } while (dirty_components.length);
        while (flush_callbacks.length) {
            flush_callbacks.pop()();
        }
        update_scheduled = false;
        seen_callbacks.clear();
        set_current_component(saved_component);
    }
    
    function update($$) {
        if ($$.fragment !== null) {
            $$.update();
            run_all($$.before_update);
            const dirty = $$.dirty;
            $$.dirty = [-1];
            $$.fragment && $$.fragment.p($$.ctx, dirty);
            $$.after_update.forEach(add_render_callback);
        }
    }
    
    const outroing = new Set();
    let outros;
    
    function group_outros() {
        outros = {
            r: 0,
            c: [],
            p: outros
        };
    }
    
    function check_outros() {
        if (!outros.r) {
            run_all(outros.c);
        }
        outros = outros.p;
    }
    
    function transition_in(block, local) {
        if (block && block.i) {
            outroing.delete(block);
            block.i(local);
        }
    }
    
    function transition_out(block, local, detach, callback) {
        if (block && block.o) {
            if (outroing.has(block)) return;
            outroing.add(block);
            outros.c.push(() => {
                outroing.delete(block);
                if (callback) {
                    if (detach) block.d(1);
                    callback();
                }
            });
            block.o(local);
        } else if (callback) {
            callback();
        }
    }
    
    // Lifecycle
    function create_in_transition(node, fn, params) {
        const options = { direction: 'in' };
        let config = fn(node, params, options);
        let running = false;
        let animation_name;
        let task;
        let uid = 0;
        
        function cleanup() {
            if (animation_name) delete_rule(node, animation_name);
        }
        
        function go() {
            const { delay = 0, duration = 300, easing = identity, tick = noop, css } = config || null_transition;
            if (css) animation_name = create_rule(node, 0, 1, duration, delay, easing, css, uid++);
            tick(0, 1);
            const start_time = now() + delay;
            const end_time = start_time + duration;
            if (task) task.abort();
            running = true;
            add_render_callback(() => dispatch(node, true, 'start'));
            task = loop(now => {
                if (running) {
                    if (now >= end_time) {
                        tick(1, 0);
                        dispatch(node, true, 'end');
                        cleanup();
                        return running = false;
                    }
                    if (now >= start_time) {
                        const t = easing((now - start_time) / duration);
                        tick(t, 1 - t);
                    }
                }
                return running;
            });
        }
        
        let started = false;
        
        return {
            start() {
                if (started) return;
                started = true;
                delete_rule(node);
                if (is_function(config)) {
                    config = config(options);
                    wait().then(go);
                } else {
                    go();
                }
            },
            invalidate() {
                started = false;
            },
            end() {
                if (running) {
                    cleanup();
                    running = false;
                }
            }
        };
    }
    
    function create_out_transition(node, fn, params) {
        const options = { direction: 'out' };
        let config = fn(node, params, options);
        let running = true;
        let animation_name;
        const group = outros;
        group.r += 1;
        
        function go() {
            const { delay = 0, duration = 300, easing = identity, tick = noop, css } = config || null_transition;
            if (css) animation_name = create_rule(node, 1, 0, duration, delay, easing, css);
            const start_time = now() + delay;
            const end_time = start_time + duration;
            add_render_callback(() => dispatch(node, false, 'start'));
            loop(now => {
                if (running) {
                    if (now >= end_time) {
                        tick(0, 1);
                        dispatch(node, false, 'end');
                        if (!--group.r) {
                            run_all(group.c);
                        }
                        return false;
                    }
                    if (now >= start_time) {
                        const t = easing((now - start_time) / duration);
                        tick(1 - t, t);
                    }
                }
                return running;
            });
        }
        
        if (is_function(config)) {
            wait().then(() => {
                config = config(options);
                go();
            });
        } else {
            go();
        }
        
        return {
            end(reset) {
                if (reset && config.tick) {
                    config.tick(1, 0);
                }
                if (running) {
                    if (animation_name) delete_rule(node, animation_name);
                    running = false;
                }
            }
        };
    }
    
    // Bindings
    function bind(component, name, callback) {
        const index = component.$$.props[name];
        if (index !== undefined) {
            component.$$.bound[index] = callback;
            callback(component.$$.ctx[index]);
        }
    }
    
    function create_component(block) {
        block && block.c();
    }
    
    function mount_component(component, target, anchor, customElement) {
        const { fragment, on_mount, on_destroy, after_update } = component.$$;
        
        fragment && fragment.m(target, anchor);
        
        if (!customElement) {
            add_render_callback(() => {
                const new_on_destroy = on_mount.map(run).filter(is_function);
                if (on_destroy) {
                    on_destroy.push(...new_on_destroy);
                } else {
                    run_all(new_on_destroy);
                }
                component.$$.on_mount = [];
            });
        }
        
        after_update.forEach(add_render_callback);
    }
    
    function destroy_component(component, detaching) {
        const $$ = component.$$;
        if ($$.fragment !== null) {
            run_all($$.on_destroy);
            $$.fragment && $$.fragment.d(detaching);
            $$.on_destroy = $$.fragment = null;
            $$.ctx = [];
        }
    }
    
    function make_dirty(component, i) {
        if (component.$$.dirty[0] === -1) {
            dirty_components.push(component);
            schedule_update();
            component.$$.dirty.fill(0);
        }
        component.$$.dirty[(i / 31) | 0] |= (1 << (i % 31));
    }
    
    function init(component, options, instance, create_fragment, not_equal, props, append_styles, dirty = [-1]) {
        const parent_component = current_component;
        set_current_component(component);
        
        const $$ = component.$$ = {
            fragment: null,
            ctx: null,
            props,
            update: noop,
            not_equal,
            bound: blank_object(),
            on_mount: [],
            on_destroy: [],
            on_disconnect: [],
            before_update: [],
            after_update: [],
            context: new Map(parent_component ? parent_component.$$.context : []),
            callbacks: blank_object(),
            dirty,
            skip_bound: false,
            root: options.target || parent_component.$$.root
        };
        
        append_styles && append_styles($$.root);
        
        let ready = false;
        
        $$.ctx = instance
            ? instance(component, options.props || {}, (i, ret, ...rest) => {
                const value = rest.length ? rest[0] : ret;
                if ($$.ctx && not_equal($$.ctx[i], $$.ctx[i] = value)) {
                    if (!$$.skip_bound && $$.bound[i])
                        $$.bound[i](value);
                    if (ready)
                        make_dirty(component, i);
                }
                return ret;
            })
            : [];
        
        $$.update();
        ready = true;
        run_all($$.before_update);
        
        $$.fragment = create_fragment ? create_fragment($$.ctx) : false;
        
        if (options.target) {
            if (options.hydrate) {
                const nodes = children(options.target);
                $$.fragment && $$.fragment.l(nodes);
                nodes.forEach(detach);
            } else {
                $$.fragment && $$.fragment.c();
            }
            
            if (options.intro) transition_in(component.$$.fragment);
            mount_component(component, options.target, options.anchor, options.customElement);
            flush();
        }
        
        set_current_component(parent_component);
    }
    
    class SvelteComponent {
        $destroy() {
            destroy_component(this, 1);
            this.$destroy = noop;
        }
        
        $on(type, callback) {
            const callbacks = (this.$$.callbacks[type] || (this.$$.callbacks[type] = []));
            callbacks.push(callback);
            return () => {
                const index = callbacks.indexOf(callback);
                if (index !== -1)
                    callbacks.splice(index, 1);
            };
        }
        
        $set($$props) {
            if (this.$$set && !is_empty($$props)) {
                this.$$.skip_bound = true;
                this.$$set($$props);
                this.$$.skip_bound = false;
            }
        }
    }
    
    // Array handling
    function ensure_array_like(array_like_or_iterator) {
        return array_like_or_iterator?.length !== undefined
            ? array_like_or_iterator
            : Array.from(array_like_or_iterator);
    }
    
    // Each block handling
    function destroy_block(block, lookup) {
        block.d(1);
        lookup.delete(block.key);
    }
    
    function outro_and_destroy_block(block, lookup) {
        transition_out(block, 1, 1, () => {
            lookup.delete(block.key);
        });
    }
    
    function fix_and_destroy_block(block, lookup) {
        block.f();
        destroy_block(block, lookup);
    }
    
    function fix_and_outro_and_destroy_block(block, lookup) {
        block.f();
        outro_and_destroy_block(block, lookup);
    }
    
    function update_keyed_each(old_blocks, dirty, get_key, dynamic, ctx, list, lookup, node, destroy, create_each_block, next, get_context) {
        let o = old_blocks.length;
        let n = list.length;
        let i = o;
        const old_indexes = {};
        while (i--) old_indexes[old_blocks[i].key] = i;
        const new_blocks = [];
        const new_lookup = new Map();
        const deltas = new Map();
        i = n;
        while (i--) {
            const child_ctx = get_context(ctx, list, i);
            const key = get_key(child_ctx);
            let block = lookup.get(key);
            if (!block) {
                block = create_each_block(key, child_ctx);
                block.c();
            } else if (dynamic) {
                block.p(child_ctx, dirty);
            }
            new_lookup.set(key, new_blocks[i] = block);
            if (key in old_indexes) deltas.set(key, Math.abs(i - old_indexes[key]));
        }
        const will_move = new Set();
        const did_move = new Set();
        function insert(block) {
            transition_in(block, 1);
            block.m(node, next);
            lookup.set(block.key, block);
            next = block.first;
            n--;
        }
        while (o && n) {
            const new_block = new_blocks[n - 1];
            const old_block = old_blocks[o - 1];
            const new_key = new_block.key;
            const old_key = old_block.key;
            if (new_block === old_block) {
                next = new_block.first;
                o--;
                n--;
            }
            else if (!new_lookup.has(old_key)) {
                destroy(old_block, lookup);
                o--;
            }
            else if (!lookup.has(new_key) || will_move.has(new_key)) {
                insert(new_block);
            }
            else if (did_move.has(old_key)) {
                o--;
            } else if (deltas.get(new_key) > deltas.get(old_key)) {
                did_move.add(new_key);
                insert(new_block);
            } else {
                will_move.add(old_key);
                o--;
            }
        }
        while (o--) {
            const old_block = old_blocks[o];
            if (!new_lookup.has(old_block.key)) destroy(old_block, lookup);
        }
        while (n) insert(new_blocks[n - 1]);
        return new_blocks;
    }
    
    function validate_each_keys(ctx, list, get_context, get_key) {
        const keys = new Set();
        for (let i = 0; i < list.length; i++) {
            const key = get_key(get_context(ctx, list, i));
            if (keys.has(key)) {
                throw new Error('Cannot have duplicate keys in a keyed each');
            }
            keys.add(key);
        }
    }
    
    // Spread props
    function spread(args, classes_to_add) {
        const attributes = assign({}, ...args);
        if (classes_to_add) {
            if (attributes.class == null) {
                attributes.class = classes_to_add;
            } else {
                attributes.class += ' ' + classes_to_add;
            }
        }
        let str = '';
        Object.keys(attributes).forEach(name => {
            if (exclude_internal_props(attributes[name])) {
                str += ' ' + name + '="' + attributes[name] + '"';
            }
        });
        return str;
    }
    
    // Helper functions
    const has_prop = (obj, prop) => Object.prototype.hasOwnProperty.call(obj, prop);
    const identity = x => x;
    const not_equal = (a, b) => a != a ? b == b : a !== b;
    const is_promise = (value) => !!value && (typeof value === 'object' || typeof value === 'function') && typeof value.then === 'function';
    
    // Hydration helpers (aliases)
    const append_hydration = append;
    const insert_hydration = insert;
    const detach_dev = detach;
    const listen_dev = listen;
    const attr_dev = attr;
    const set_data_dev = set_data;
    const children_dev = children;
    const insert_hydration_dev = insert;
    const append_hydration_dev = append;
    const insert_dev = insert;
    
    // Lifecycle hooks export
    const globals = typeof window !== 'undefined'
        ? window
        : typeof globalThis !== 'undefined'
        ? globalThis
        : global;
    
    // construct_svelte_component - used by <svelte:component> directive
    function construct_svelte_component(component, props) {
        return new component({
            props: props,
            $$inline: true
        });
    }
    
    // construct_svelte_component_dev - development version
    function construct_svelte_component_dev(component, props) {
        const instance = construct_svelte_component(component, props);
        return instance;
    }
    
    // Export everything globally
    const exports = {
        SvelteComponent,
        init,
        noop,
        safe_not_equal,
        append,
        insert,
        detach,
        detach_dev,
        element,
        text,
        space,
        empty,
        listen,
        attr,
        attr_dev,
        set_data,
        set_data_dev,
        prop_dev: (node, property, value) => {
            node[property] = value;
        },
        add_location: noop,
        append_hydration,
        append_hydration_dev,
        insert_hydration,
        insert_hydration_dev,
        set_input_value,
        set_input_type,
        set_style,
        select_option,
        select_options,
        select_value,
        select_multiple_value,
        add_resize_listener: noop,
        toggle_class: (element, name, toggle) => {
            element.classList[toggle ? 'add' : 'remove'](name);
        },
        custom_event,
        destroy_each,
        run_all,
        is_function,
        blank_object,
        subscribe,
        get_store_value,
        component_subscribe,
        create_slot,
        get_slot_context,
        get_slot_changes,
        update_slot,
        update_slot_base,
        assign,
        exclude_internal_props,
        compute_rest_props,
        compute_slots,
        set_store_value,
        mount_component,
        destroy_component,
        transition_in,
        transition_out,
        check_outros,
        group_outros,
        create_component,
        construct_svelte_component,
        construct_svelte_component_dev,
        ensure_array_like,
        update_keyed_each,
        destroy_block,
        outro_and_destroy_block,
        fix_and_destroy_block,
        fix_and_outro_and_destroy_block,
        validate_each_keys,
        spread,
        prevent_default,
        stop_propagation,
        self,
        trusted,
        set_attributes,
        set_svg_attributes,
        set_custom_element_data,
        xlink_attr,
        get_binding_group_value,
        to_number,
        time_ranges_to_array,
        children,
        claim_element,
        claim_text,
        claim_space,
        claim_component: noop,
        query_selector_all,
        set_current_component,
        get_current_component,
        onMount,
        onDestroy,
        createEventDispatcher,
        setContext,
        getContext,
        getAllContexts,
        hasContext,
        bubble,
        clear_loops: noop,
        loop: noop,
        dirty_components,
        schedule_update,
        tick,
        binding_callbacks,
        render_callbacks,
        flush_callbacks,
        resolved_promise,
        update_scheduled,
        add_render_callback,
        add_flush_callback,
        flush,
        make_dirty,
        bind,
        create_in_transition,
        create_out_transition,
        create_bidirectional_transition: noop,
        null_to_empty: (value) => value == null ? '' : value,
        is_empty,
        validate_component: noop,
        validate_slots: noop,
        validate_each_argument: ensure_array_like,
        validate_dynamic_element: noop,
        validate_void_dynamic_element: noop,
        once: (fn) => {
            let ran = false;
            return function(...args) {
                if (ran) return;
                ran = true;
                fn.call(this, ...args);
            };
        },
        add_transform: noop,
        add_classes: noop,
        add_flush_callback,
        add_iframe_resize_listener: noop,
        afterUpdate: noop,
        append_dev: append,
        append_empty_stylesheet: noop,
        append_styles: noop,
        beforeUpdate: noop,
        bind_this: noop,
        bubble_event: bubble,
        check_dispose: noop,
        comment: noop,
        component_subscribe,
        contenteditable_truthy_values: ['', true, 1, 'true', 'contenteditable'],
        create_animation: noop,
        create_bidirectional_transition: noop,
        create_component,
        create_in_transition,
        create_out_transition,
        create_slot,
        create_ssr_component: noop,
        current_component,
        dataset_dev: noop,
        debug: noop,
        destroy_block,
        destroy_component,
        destroy_each,
        detach,
        detach_after_dev: noop,
        detach_before_dev: noop,
        detach_between_dev: noop,
        detach_dev,
        dispatch_dev: noop,
        element,
        element_is,
        empty,
        end_hydrating: noop,
        ensure_array_like,
        escape: noop,
        escape_attribute_value: noop,
        escape_object: noop,
        escaped: x => x,
        exclude_internal_props,
        fix_and_destroy_block,
        fix_and_outro_and_destroy_block,
        fix_position: noop,
        flush,
        get_all_dirty_from_scope: noop,
        get_binding_group_value,
        get_current_component,
        get_custom_elements_slots: noop,
        get_slot_changes,
        get_slot_context,
        get_spread_object: () => ({}),
        get_spread_update: () => ({}),
        get_store_value,
        globals,
        group_outros,
        handle_promise: noop,
        has_prop,
        HtmlTag: class {},
        HtmlTagHydration: class {},
        identity,
        init,
        insert,
        insert_dev,
        insert_hydration,
        insert_hydration_dev,
        intros: { enabled: false },
        invalid_attribute_name_character: /[\s'">/=\u{FDD0}-\u{FDEF}\u{FFFE}\u{FFFF}\u{1FFFE}\u{1FFFF}\u{2FFFE}\u{2FFFF}\u{3FFFE}\u{3FFFF}\u{4FFFE}\u{4FFFF}\u{5FFFE}\u{5FFFF}\u{6FFFE}\u{6FFFF}\u{7FFFE}\u{7FFFF}\u{8FFFE}\u{8FFFF}\u{9FFFE}\u{9FFFF}\u{AFFFE}\u{AFFFF}\u{BFFFE}\u{BFFFF}\u{CFFFE}\u{CFFFF}\u{DFFFE}\u{DFFFF}\u{EFFFE}\u{EFFFF}\u{FFFFE}\u{FFFFF}\u{10FFFE}\u{10FFFF}]/u,
        is_client: true,
        is_crossorigin: noop,
        is_empty,
        is_function,
        is_promise,
        listen,
        listen_dev,
        loop: noop,
        loop_guard: noop,
        merge_ssr_styles: noop,
        missing_component: { $$render: () => '' },
        mount_component,
        noop,
        not_equal,
        null_to_empty: value => value == null ? '' : value,
        object_without_properties,
        onDestroy,
        onMount,
        once: fn => {
            let ran = false;
            return function(...args) {
                if (ran) return;
                ran = true;
                fn.call(this, ...args);
            };
        },
        outro_and_destroy_block,
        prevent_default,
        prop_dev: (node, property, value) => {
            node[property] = value;
        },
        query_selector_all,
        run,
        run_all,
        safe_not_equal,
        schedule_update,
        select_multiple_value,
        select_option,
        select_options,
        select_value,
        self,
        set_attributes,
        set_current_component,
        set_custom_element_data,
        set_data,
        set_data_dev: set_data,
        set_data_maybe_contenteditable: set_data,
        set_data_maybe_contenteditable_dev: set_data,
        set_input_type,
        set_input_value,
        set_now: f => () => f(),
        set_raf: f => () => f(),
        set_store_value,
        set_style,
        set_svg_attributes,
        space,
        spread,
        src_url_equal: (a, b) => a == b,
        start_hydrating: noop,
        stop_propagation,
        subscribe,
        svg_element,
        text,
        tick,
        time_ranges_to_array,
        to_number,
        toggle_class: (element, name, toggle) => {
            element.classList[toggle ? 'add' : 'remove'](name);
        },
        transition_in,
        transition_out,
        trusted,
        update_await_block_branch: noop,
        update_keyed_each,
        update_slot,
        update_slot_base,
        validate_component: noop,
        validate_dynamic_element: noop,
        validate_each_argument: ensure_array_like,
        validate_each_keys,
        validate_slots: noop,
        validate_store: noop,
        validate_void_dynamic_element: noop,
        xlink_attr,
        
        // Debug/dev exports
        $$: {
            on_destroy: [],
            context: new Map(),
            on_mount: [],
            before_update: [],
            after_update: [],
            callbacks: {}
        }
    };
    
    // Export to global
    if (typeof global.Svelte === 'undefined') {
        global.Svelte = exports;
    }
    
    // Also attach individual exports to global
    Object.keys(exports).forEach(key => {
        if (typeof global[key] === 'undefined') {
            global[key] = exports[key];
        }
    });
    
})(typeof window !== 'undefined' ? window : global);