<script>
    import Card from './_lib/Card.svelte';
    import Button from './_lib/Button.svelte';
    
    const cards = [
        {
            id: 1,
            title: 'Svelte Components',
            description: 'Learn how to build reusable components with Svelte',
            imageUrl: '/images/svelte.png'
        },
        {
            id: 2,
            title: 'Component Composition',
            description: 'Discover how to compose complex UIs from simple components',
            imageUrl: '/images/composition.png'
        },
        {
            id: 3,
            title: 'State Management',
            description: 'Master state management in your Svelte applications',
            imageUrl: '/images/state.png'
        }
    ];
    
    let selectedCard = null;
    
    function selectCard(card) {
        selectedCard = card;
    }
    
    function clearSelection() {
        selectedCard = null;
    }
</script>

<main>
    <h1>Card Gallery with Nested Components</h1>
    <p>This demo shows multiple components imported together (Card and Button).</p>
    
    {#if selectedCard}
        <div class="selected-info">
            <h2>Selected: {selectedCard.title}</h2>
            <p>{selectedCard.description}</p>
            <Button variant="outline" size="small" on:click={clearSelection}>
                Clear Selection
            </Button>
        </div>
    {/if}
    
    <div class="card-grid">
        {#each cards as card}
            <div class="card-wrapper">
                <Card 
                    title={card.title}
                    description={card.description}
                    imageUrl={card.imageUrl}
                >
                    <div slot="actions">
                        <Button 
                            variant="primary" 
                            size="small"
                            on:click={() => selectCard(card)}
                        >
                            Select This Card
                        </Button>
                    </div>
                </Card>
            </div>
        {/each}
    </div>
    
    <div class="footer">
        <h3>Component Import Chain</h3>
        <p>card-gallery.svelte → imports → Card.svelte + Button.svelte</p>
        <p>This demonstrates that multiple components can be imported and used together.</p>
    </div>
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
    
    .selected-info {
        background: #f0f9ff;
        border: 2px solid #3b82f6;
        border-radius: 8px;
        padding: 1.5rem;
        margin: 2rem 0;
    }
    
    .selected-info h2 {
        margin: 0 0 0.5rem 0;
        color: #1e40af;
    }
    
    .card-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
        gap: 2rem;
        margin: 2rem 0;
    }
    
    .card-wrapper {
        height: 100%;
    }
    
    .footer {
        margin-top: 3rem;
        padding: 2rem;
        background: #f5f5f5;
        border-radius: 8px;
    }
    
    .footer h3 {
        margin: 0 0 1rem 0;
        color: #333;
    }
    
    .footer p {
        margin: 0.5rem 0;
        color: #666;
        font-family: monospace;
    }
</style>