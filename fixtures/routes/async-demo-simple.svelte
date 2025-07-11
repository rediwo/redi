<script>
    import { onMount, onDestroy } from 'svelte';
    import Button from './Button.svelte';
    
    // State for dynamic loading
    let chartLoaded = false;
    let tableLoaded = false;
    let chartLoading = false;
    let tableLoading = false;
    
    // Component instances
    let chartInstance = null;
    let tableInstance = null;
    
    // Container refs
    let chartContainer;
    let tableContainer;
    
    // Check if SvelteAsync is available
    let SvelteAsync = window.SvelteAsync;
    
    // Sample data for components
    const chartData = [
        { label: 'Jan', value: 45 },
        { label: 'Feb', value: 67 },
        { label: 'Mar', value: 34 },
        { label: 'Apr', value: 78 },
        { label: 'May', value: 56 },
        { label: 'Jun', value: 89 }
    ];
    
    // Async loading functions
    async function loadChart() {
        if (chartLoaded || chartLoading) return;
        chartLoading = true;
        
        try {
            if (!SvelteAsync) {
                console.error('SvelteAsync library not loaded');
                return;
            }
            
            // Load the component class
            const HeavyChart = await SvelteAsync.import('HeavyChart');
            
            // Create instance manually
            if (chartContainer) {
                chartInstance = new HeavyChart({
                    target: chartContainer,
                    props: {
                        data: chartData,
                        title: 'Monthly Sales'
                    }
                });
            }
            
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
            
            // Load the component class
            const DataTable = await SvelteAsync.import('DataTable');
            
            // Create instance manually
            if (tableContainer) {
                tableInstance = new DataTable({
                    target: tableContainer,
                    props: {
                        title: 'User Management'
                    }
                });
            }
            
            tableLoaded = true;
        } catch (error) {
            console.error('Failed to load table:', error);
        } finally {
            tableLoading = false;
        }
    }
    
    // Cleanup on destroy
    onDestroy(() => {
        if (chartInstance) {
            chartInstance.$destroy();
        }
        if (tableInstance) {
            tableInstance.$destroy();
        }
    });
</script>

<main>
    <h1>Async Component Loading Demo (Simple)</h1>
    <p>This demo shows how to load Svelte components asynchronously without using svelte:component.</p>
    
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
        <h2>Async Component Loading</h2>
        <p>Click the buttons below to load heavy components on-demand:</p>
        
        <div class="async-section">
            <h3>Heavy Chart Component</h3>
            <Button on:click={loadChart} disabled={chartLoading || chartLoaded}>
                {#if chartLoading}
                    Loading Chart...
                {:else if chartLoaded}
                    Chart Loaded ✓
                {:else}
                    Load Chart
                {/if}
            </Button>
            
            <div bind:this={chartContainer} class="component-container">
                {#if !chartLoaded && !chartLoading}
                    <div class="placeholder">
                        <p>Chart component not loaded yet. Click the button above to load it.</p>
                    </div>
                {:else if chartLoading}
                    <div class="loading-placeholder">
                        <div class="spinner"></div>
                        <p>Loading chart component...</p>
                    </div>
                {/if}
            </div>
        </div>
        
        <div class="async-section">
            <h3>Data Table Component</h3>
            <Button on:click={loadTable} disabled={tableLoading || tableLoaded}>
                {#if tableLoading}
                    Loading Table...
                {:else if tableLoaded}
                    Table Loaded ✓
                {:else}
                    Load Table
                {/if}
            </Button>
            
            <div bind:this={tableContainer} class="component-container">
                {#if !tableLoaded && !tableLoading}
                    <div class="placeholder">
                        <p>Table component not loaded yet. Click the button above to load it.</p>
                    </div>
                {:else if tableLoading}
                    <div class="loading-placeholder">
                        <div class="spinner"></div>
                        <p>Loading table component...</p>
                    </div>
                {/if}
            </div>
        </div>
    </section>
    
    <section class="demo-section">
        <h2>How It Works</h2>
        <p>This approach manually instantiates components after loading:</p>
        <pre><code>{`// Load the component class
const Component = await SvelteAsync.import('ComponentName');

// Create instance manually
const instance = new Component({
    target: containerElement,
    props: { /* props */ }
});

// Later, clean up
instance.$destroy();`}</code></pre>
        <p>This method avoids using <code>&lt;svelte:component&gt;</code> which requires additional runtime context.</p>
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
    
    .component-container {
        margin-top: 1rem;
        min-height: 200px;
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
    
    pre {
        background: #f5f5f5;
        padding: 1rem;
        border-radius: 4px;
        overflow-x: auto;
        font-size: 0.9rem;
        line-height: 1.4;
    }
    
    code {
        color: #333;
        font-family: monospace;
    }
    
    p {
        line-height: 1.6;
        color: #555;
    }
</style>