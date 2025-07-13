<script>
    import Button from './_lib/Button.svelte';
    
    let clickCount = 0;
    let lastVariant = '';
    
    function handlePrimaryClick() {
        clickCount++;
        lastVariant = 'primary';
    }
    
    function handleSecondaryClick() {
        clickCount++;
        lastVariant = 'secondary';
    }
    
    function handleOutlineClick() {
        clickCount++;
        lastVariant = 'outline';
    }
    
    function reset() {
        clickCount = 0;
        lastVariant = '';
    }
</script>

<main>
    <h1>Svelte Component Import Demo</h1>
    <p>This demonstrates importing and using another Svelte component.</p>
    
    <section class="button-showcase">
        <h2>Button Variants</h2>
        <div class="button-group">
            <Button on:click={handlePrimaryClick}>Primary Button</Button>
            <Button variant="secondary" on:click={handleSecondaryClick}>Secondary Button</Button>
            <Button variant="outline" on:click={handleOutlineClick}>Outline Button</Button>
        </div>
        
        <h2>Button Sizes</h2>
        <div class="button-group">
            <Button size="small">Small</Button>
            <Button size="medium">Medium</Button>
            <Button size="large">Large</Button>
        </div>
        
        <h2>Disabled State</h2>
        <div class="button-group">
            <Button disabled={true}>Disabled Primary</Button>
            <Button variant="secondary" disabled={true}>Disabled Secondary</Button>
            <Button variant="outline" disabled={true}>Disabled Outline</Button>
        </div>
        
        <h2>Click Counter</h2>
        <div class="stats">
            <p>Total clicks: <strong>{clickCount}</strong></p>
            {#if lastVariant}
                <p>Last clicked: <strong>{lastVariant}</strong> button</p>
            {/if}
            {#if clickCount > 0}
                <Button variant="outline" size="small" on:click={reset}>Reset Counter</Button>
            {/if}
        </div>
    </section>
</main>

<style>
    main {
        max-width: 800px;
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
        margin-top: 2rem;
        margin-bottom: 1rem;
        font-size: 1.5rem;
    }
    
    p {
        color: #666;
        line-height: 1.6;
    }
    
    .button-showcase {
        margin-top: 2rem;
    }
    
    .button-group {
        display: flex;
        gap: 1rem;
        flex-wrap: wrap;
        margin-bottom: 1.5rem;
    }
    
    .stats {
        background: #f5f5f5;
        padding: 1.5rem;
        border-radius: 8px;
        margin-top: 1rem;
    }
    
    .stats p {
        margin: 0.5rem 0;
    }
    
    .stats strong {
        color: #ff3e00;
    }
</style>