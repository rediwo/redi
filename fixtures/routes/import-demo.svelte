<script>
    // Import a Svelte component
    import Button from './Button.svelte';
    
    // Import CSS (will be transformed to URL)
    import styles from '/css/style.css';
    
    // Import an image (will be transformed to URL)
    import logo from '/images/logo.svg';
    
    // Import JSON data (will be transformed to actual data object)
    import packageInfo from '/data.json';
    
    let showInfo = false;
    
    function toggleInfo() {
        showInfo = !showInfo;
    }
    
    // Log imported assets to console
    console.log('CSS URL:', styles);
    console.log('Logo URL:', logo);
    console.log('Package JSON data:', packageInfo);
    console.log('Package name:', packageInfo.name);
    console.log('Package version:', packageInfo.version);
</script>

<main>
    <h1>Enhanced Import System Demo</h1>
    <p>This demo shows how to import various asset types in Svelte components.</p>
    
    <section class="demo-section">
        <h2>Imported Assets</h2>
        
        <div class="asset-demo">
            <h3>1. CSS Import</h3>
            <p>CSS files are imported as URL strings:</p>
            <pre><code>import styles from '/css/style.css';
// Becomes: const styles = '/css/style.css';</code></pre>
            <p>CSS URL: <code>{styles}</code></p>
        </div>
        
        <div class="asset-demo">
            <h3>2. Image Import</h3>
            <p>Images are imported as URL strings:</p>
            <pre><code>import logo from '/images/logo.svg';
// Becomes: const logo = '/images/logo.svg';</code></pre>
            <img src={logo} alt="Svelte Logo" style="width: 100px; height: 100px;">
            <p>Image URL: <code>{logo}</code></p>
        </div>
        
        <div class="asset-demo">
            <h3>3. JSON Import</h3>
            <p>JSON files are imported as JavaScript objects (data is inlined):</p>
            <pre><code>import packageInfo from '/data.json';
// Becomes: const packageInfo = &#123;"name": "...", ...&#125;;</code></pre>
            <div>
                <p>Package name: <code>{packageInfo.name}</code></p>
                <p>Package version: <code>{packageInfo.version}</code></p>
                <p>Full data:</p>
                <pre><code>{JSON.stringify(packageInfo, null, 2)}</code></pre>
            </div>
        </div>
        
        <div class="asset-demo">
            <h3>4. Component Import</h3>
            <p>Svelte components work as before:</p>
            <pre><code>import Button from './Button.svelte';</code></pre>
            <Button on:click={toggleInfo}>
                {showInfo ? 'Hide' : 'Show'} Implementation Details
            </Button>
        </div>
    </section>
    
    {#if showInfo}
    <section class="info-section">
        <h2>Implementation Details</h2>
        <ul>
            <li><strong>Flexible Imports:</strong> Support for CSS, JS, JSON, images, fonts, and more</li>
            <li><strong>URL Transformation:</strong> Non-Svelte imports become URL constants (except JSON)</li>
            <li><strong>JSON Inlining:</strong> JSON imports are inlined as JavaScript objects (up to 100KB)</li>
            <li><strong>Path Resolution:</strong> Supports relative and absolute paths</li>
            <li><strong>Asset Discovery:</strong> Looks in current dir, public/, and project root</li>
            <li><strong>Security:</strong> Prevents directory traversal attacks</li>
            <li><strong>No Hardcoded Paths:</strong> Works with any project structure</li>
        </ul>
        
        <h3>Supported Asset Types</h3>
        <table>
            <tr>
                <th>Extension</th>
                <th>Type</th>
                <th>Import Result</th>
            </tr>
            <tr>
                <td>.svelte</td>
                <td>Component</td>
                <td>Component class</td>
            </tr>
            <tr>
                <td>.js, .mjs</td>
                <td>JavaScript</td>
                <td>URL string</td>
            </tr>
            <tr>
                <td>.css</td>
                <td>Stylesheet</td>
                <td>URL string</td>
            </tr>
            <tr>
                <td>.json</td>
                <td>JSON</td>
                <td>JavaScript object (inlined)</td>
            </tr>
            <tr>
                <td>.png, .jpg, .svg, etc.</td>
                <td>Image</td>
                <td>URL string</td>
            </tr>
            <tr>
                <td>.woff, .woff2, etc.</td>
                <td>Font</td>
                <td>URL string</td>
            </tr>
        </table>
    </section>
    {/if}
</main>

<style>
    main {
        max-width: 1000px;
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
        margin: 1.5rem 0 0.5rem 0;
    }
    
    .demo-section {
        margin: 2rem 0;
    }
    
    .asset-demo {
        margin: 2rem 0;
        padding: 1.5rem;
        background: #f5f5f5;
        border-radius: 8px;
        border: 1px solid #ddd;
    }
    
    pre {
        background: #2d2d2d;
        color: #f8f8f2;
        padding: 1rem;
        border-radius: 4px;
        overflow-x: auto;
        margin: 0.5rem 0;
    }
    
    code {
        font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
        font-size: 0.9rem;
    }
    
    .info-section {
        background: #e8f4fd;
        padding: 2rem;
        border-radius: 8px;
        margin: 2rem 0;
        border: 1px solid #b3d9f2;
    }
    
    .info-section h2 {
        border-bottom-color: #2196f3;
    }
    
    ul {
        margin: 1rem 0;
        padding-left: 2rem;
    }
    
    li {
        margin: 0.5rem 0;
        line-height: 1.6;
    }
    
    table {
        width: 100%;
        border-collapse: collapse;
        margin: 1rem 0;
    }
    
    th, td {
        text-align: left;
        padding: 0.5rem;
        border: 1px solid #ddd;
    }
    
    th {
        background: #f0f0f0;
        font-weight: 600;
    }
    
    img {
        display: block;
        margin: 1rem 0;
    }
</style>