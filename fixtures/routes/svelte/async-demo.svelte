<script>
    // Static imports for immediately needed components
    import Button from './_lib/Button.svelte';
    
    // State for dynamic loading
    let HeavyChart;
    let DataTable;
    let chartLoaded = false;
    let tableLoaded = false;
    let chartLoading = false;
    let tableLoading = false;
    
    // Sample data for components
    const chartData = [
        { label: 'Jan', value: 45 },
        { label: 'Feb', value: 67 },
        { label: 'Mar', value: 34 },
        { label: 'Apr', value: 78 },
        { label: 'May', value: 56 },
        { label: 'Jun', value: 89 }
    ];
    
    // Check if SvelteAsync is available
    let SvelteAsync = window.SvelteAsync;
    
    // Async loading functions
    async function loadChart() {
        if (chartLoaded || chartLoading) return;
        chartLoading = true;
        
        try {
            if (!SvelteAsync) {
                console.error('SvelteAsync library not loaded');
                return;
            }
            // Use the SvelteAsync library which properly handles runtime context
            HeavyChart = await SvelteAsync.import('./_lib/HeavyChart');
            chartLoaded = true;
        } catch (error) {
            console.error('Failed to load chart:', error);
        } finally {
            chartLoading = false;
        }
    }
    
    async function loadTable() {
        if (tableLoaded || tableLoading) return;
        tableLoading = true;
        
        try {
            if (!SvelteAsync) {
                console.error('SvelteAsync library not loaded');
                return;
            }
            // Use the SvelteAsync library which properly handles runtime context
            DataTable = await SvelteAsync.import('./_lib/DataTable');
            tableLoaded = true;
        } catch (error) {
            console.error('Failed to load table:', error);
        } finally {
            tableLoading = false;
        }
    }
    
    // Using SvelteAsync library (alternative approach)
    async function loadChartWithLibrary() {
        if (chartLoaded || chartLoading) return;
        chartLoading = true;
        
        try {
            HeavyChart = await SvelteAsync.import('./_lib/HeavyChart');
            chartLoaded = true;
        } catch (error) {
            console.error('Failed to load chart with library:', error);
        } finally {
            chartLoading = false;
        }
    }
    
    async function loadTableWithLibrary() {
        if (tableLoaded || tableLoading) return;
        tableLoading = true;
        
        try {
            DataTable = await SvelteAsync.import('./_lib/DataTable');
            tableLoaded = true;
        } catch (error) {
            console.error('Failed to load table with library:', error);
        } finally {
            tableLoading = false;
        }
    }
</script>

<main>
    <h1>Async Component Loading Demo</h1>
    <p>This demo shows how to load Svelte components asynchronously for better performance.</p>
    
    <section class="demo-section">
        <h2>Always Available Components</h2>
        <p>These components are loaded synchronously and are immediately available:</p>
        <div class="button-group">
            <Button>Regular Button</Button>
            <Button variant="secondary">Secondary Button</Button>
            <Button variant="outline">Outline Button</Button>
        </div>
    </section>
    
    <section class="demo-section">
        <h2>Async Loading Demo (Manual)</h2>
        <p>These heavy components are loaded on-demand to improve initial page load time:</p>
        
        <div class="async-section">
            <h3>Heavy Chart Component</h3>
            <div class="load-buttons">
                <Button on:click={loadChart} disabled={chartLoading}>
                    {chartLoading ? 'Loading Chart...' : 'Load Chart (Manual)'}
                </Button>
                <Button on:click={loadChartWithLibrary} disabled={chartLoading}>
                    {chartLoading ? 'Loading Chart...' : 'Load Chart (Library)'}
                </Button>
            </div>
            
            {#if chartLoaded && HeavyChart}
                <svelte:component this={HeavyChart} data={chartData} title="Monthly Sales" />
            {:else if chartLoading}
                <div class="loading-placeholder">
                    <div class="spinner"></div>
                    <p>Loading chart component...</p>
                </div>
            {:else}
                <div class="placeholder">
                    <p>Chart component not loaded yet. Click the button above to load it.</p>
                </div>
            {/if}
        </div>
        
        <div class="async-section">
            <h3>Data Table Component</h3>
            <div class="load-buttons">
                <Button on:click={loadTable} disabled={tableLoading}>
                    {tableLoading ? 'Loading Table...' : 'Load Table (Manual)'}
                </Button>
                <Button on:click={loadTableWithLibrary} disabled={tableLoading}>
                    {tableLoading ? 'Loading Table...' : 'Load Table (Library)'}
                </Button>
            </div>
            
            {#if tableLoaded && DataTable}
                <svelte:component this={DataTable} title="User Management" />
            {:else if tableLoading}
                <div class="loading-placeholder">
                    <div class="spinner"></div>
                    <p>Loading table component...</p>
                </div>
            {:else}
                <div class="placeholder">
                    <p>Table component not loaded yet. Click the button above to load it.</p>
                </div>
            {/if}
        </div>
    </section>
    
    <section class="demo-section">
        <h2>Benefits of Async Loading</h2>
        <ul>
            <li><strong>Faster Initial Load:</strong> Only essential components are loaded immediately</li>
            <li><strong>Better User Experience:</strong> Users see content faster, heavy components load as needed</li>
            <li><strong>Reduced Bundle Size:</strong> Large components don't bloat the initial page</li>
            <li><strong>Progressive Enhancement:</strong> Features are added as they become available</li>
            <li><strong>Better Caching:</strong> Components are cached independently</li>
        </ul>
    </section>
    
    <section class="demo-section">
        <h2>Implementation Approaches</h2>
        <div class="approach-comparison">
            <div class="approach">
                <h3>Manual Fetch</h3>
                <pre><code>// Fetch component manually
const response = await fetch('/svelte/_lib/HeavyChart');
const data = await response.json();
const Component = new Function(data.js + '\nreturn ' + data.className + ';')();
</code></pre>
            </div>
            
            <div class="approach">
                <h3>SvelteAsync Library</h3>
                <pre><code>// Use the convenience library
const Component = await SvelteAsync.import('./_lib/HeavyChart');
// Or with lazy loading
const LazyComponent = SvelteAsync.lazy(() => SvelteAsync.import('./_lib/HeavyChart'));
</code></pre>
            </div>
        </div>
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
        margin-bottom: 1rem;
    }
    
    h2 {
        color: #333;
        margin: 2rem 0 1rem 0;
        border-bottom: 2px solid #ff3e00;
        padding-bottom: 0.5rem;
    }
    
    h3 {
        color: #555;
        margin: 1.5rem 0 1rem 0;
    }
    
    .demo-section {
        margin: 2rem 0;
        padding: 1.5rem;
        border: 1px solid #eee;
        border-radius: 8px;
        background: #fafafa;
    }
    
    .button-group {
        display: flex;
        gap: 1rem;
        flex-wrap: wrap;
        margin: 1rem 0;
    }
    
    .async-section {
        margin: 2rem 0;
        padding: 1.5rem;
        border: 1px solid #ddd;
        border-radius: 8px;
        background: white;
    }
    
    .load-buttons {
        display: flex;
        gap: 1rem;
        margin: 1rem 0;
    }
    
    .placeholder, .loading-placeholder {
        padding: 2rem;
        text-align: center;
        border: 2px dashed #ddd;
        border-radius: 8px;
        margin: 1rem 0;
    }
    
    .placeholder {
        color: #666;
        background: #f9f9f9;
    }
    
    .loading-placeholder {
        color: #666;
        background: #f0f8ff;
        border-color: #87ceeb;
    }
    
    .spinner {
        width: 32px;
        height: 32px;
        border: 3px solid #f3f3f3;
        border-top: 3px solid #ff3e00;
        border-radius: 50%;
        animation: spin 1s linear infinite;
        margin: 0 auto 1rem;
    }
    
    @keyframes spin {
        0% { transform: rotate(0deg); }
        100% { transform: rotate(360deg); }
    }
    
    .demo-section ul {
        list-style: none;
        padding: 0;
    }
    
    .demo-section li {
        padding: 0.5rem 0;
        border-bottom: 1px solid #eee;
    }
    
    .demo-section li:last-child {
        border-bottom: none;
    }
    
    .approach-comparison {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
        gap: 2rem;
        margin: 2rem 0;
    }
    
    .approach {
        padding: 1.5rem;
        border: 1px solid #ddd;
        border-radius: 8px;
        background: white;
    }
    
    .approach h3 {
        margin-top: 0;
        color: #ff3e00;
    }
    
    .approach pre {
        background: #f5f5f5;
        padding: 1rem;
        border-radius: 4px;
        overflow-x: auto;
        font-size: 0.9rem;
        line-height: 1.4;
    }
    
    .approach code {
        color: #333;
    }
    
    p {
        line-height: 1.6;
        color: #555;
    }
</style>