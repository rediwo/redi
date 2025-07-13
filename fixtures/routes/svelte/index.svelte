<script>
    const demos = [
        {
            title: 'Counter Demo',
            description: 'Basic state management with reactive updates',
            path: '/svelte/counter'
        },
        {
            title: 'Button Demo',
            description: 'Component imports with various button styles and states',
            path: '/svelte/button-demo'
        },
        {
            title: 'Card Gallery',
            description: 'Multiple component composition with Card and Button',
            path: '/svelte/card-gallery'
        },
        {
            title: 'Component Library',
            description: 'Complex component composition with Icons, Cards, and Buttons',
            path: '/svelte/component-library-demo'
        },
        {
            title: 'Data Table',
            description: 'Advanced data display with sorting and filtering',
            path: '/svelte/data-table-demo'
        },
        {
            title: 'Import Examples',
            description: 'Various import types: CSS, images, JSON, and components',
            path: '/svelte/import-demo'
        },
        {
            title: 'Async Loading',
            description: 'Dynamic component loading for performance optimization',
            path: '/svelte/async-demo'
        }
    ];
</script>

<main>
    <h1>Svelte Framework Demos</h1>
    <p>Explore various Svelte features and component patterns</p>
    
    <div class="demo-grid">
        {#each demos as demo}
            <a href={demo.path} class="demo-card">
                <h3>{demo.title}</h3>
                <p>{demo.description}</p>
            </a>
        {/each}
    </div>
    
    <section class="framework-info">
        <h2>About Svelte</h2>
        <ul>
            <li>Compile-time optimizations for smaller bundles</li>
            <li>Reactive declarations with $: syntax</li>
            <li>No virtual DOM - direct DOM manipulation</li>
            <li>Built-in state management</li>
            <li>Scoped CSS by default</li>
        </ul>
    </section>
</main>

<style>
    main {
        max-width: 1200px;
        margin: 0 auto;
        padding: 2rem;
        font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    }
    
    h1 {
        color: #ff3e00;
        margin-bottom: 0.5rem;
    }
    
    h2 {
        color: #333;
        margin: 2rem 0 1rem;
    }
    
    .demo-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
        gap: 1.5rem;
        margin: 2rem 0;
    }
    
    .demo-card {
        display: block;
        padding: 1.5rem;
        background: white;
        border: 2px solid #eee;
        border-radius: 8px;
        text-decoration: none;
        color: inherit;
        transition: all 0.2s;
        box-shadow: 0 2px 4px rgba(0,0,0,0.05);
    }
    
    .demo-card:hover {
        border-color: #ff3e00;
        transform: translateY(-2px);
        box-shadow: 0 4px 8px rgba(0,0,0,0.1);
    }
    
    .demo-card h3 {
        margin: 0 0 0.5rem 0;
        color: #ff3e00;
    }
    
    .demo-card p {
        margin: 0;
        color: #666;
        font-size: 0.95rem;
    }
    
    .framework-info {
        margin-top: 3rem;
        padding: 2rem;
        background: #f5f5f5;
        border-radius: 8px;
    }
    
    .framework-info ul {
        list-style: none;
        padding: 0;
    }
    
    .framework-info li {
        padding: 0.5rem 0;
        padding-left: 1.5rem;
        position: relative;
    }
    
    .framework-info li:before {
        content: "✓";
        position: absolute;
        left: 0;
        color: #ff3e00;
        font-weight: bold;
    }
</style>