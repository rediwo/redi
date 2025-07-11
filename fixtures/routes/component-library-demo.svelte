<script>
    import Button from './Button.svelte';
    import Card from './Card.svelte';
    import Icon from './_components/Icon.svelte';
    
    let rating = 0;
    let favorites = [];
    
    const products = [
        { id: 1, name: 'Svelte Tutorial', price: 'Free', rating: 5 },
        { id: 2, name: 'Advanced Components', price: '$29', rating: 4 },
        { id: 3, name: 'State Management Guide', price: '$19', rating: 5 }
    ];
    
    function toggleFavorite(productId) {
        if (favorites.includes(productId)) {
            favorites = favorites.filter(id => id !== productId);
        } else {
            favorites = [...favorites, productId];
        }
    }
    
    function isFavorite(productId) {
        return favorites.includes(productId);
    }
</script>

<main>
    <h1>Component Library Demo</h1>
    <p>This example shows multiple imported components working together, including nested imports.</p>
    
    <section class="icon-showcase">
        <h2>Icon Components</h2>
        <div class="icon-grid">
            <div class="icon-item">
                <Icon name="star" size={32} color="#ff3e00" />
                <span>Star</span>
            </div>
            <div class="icon-item">
                <Icon name="heart" size={32} color="#e91e63" />
                <span>Heart</span>
            </div>
            <div class="icon-item">
                <Icon name="check" size={32} color="#4caf50" />
                <span>Check</span>
            </div>
            <div class="icon-item">
                <Icon name="menu" size={32} color="#2196f3" />
                <span>Menu</span>
            </div>
        </div>
    </section>
    
    <section class="product-showcase">
        <h2>Product Cards with Interactive Elements</h2>
        <div class="product-grid">
            {#each products as product}
                <Card 
                    title={product.name}
                    description="Learn Svelte with our comprehensive guides"
                >
                    <div slot="actions" class="card-actions">
                        <div class="rating">
                            {#each Array(5) as _, i}
                                <Icon 
                                    name="star" 
                                    size={16} 
                                    color={i < product.rating ? '#ff3e00' : '#ccc'}
                                />
                            {/each}
                        </div>
                        <div class="price">{product.price}</div>
                        <div class="buttons">
                            <Button size="small" variant="primary">
                                <Icon name="arrow_right" size={16} color="white" />
                                View Details
                            </Button>
                            <Button 
                                size="small" 
                                variant={isFavorite(product.id) ? 'primary' : 'outline'}
                                on:click={() => toggleFavorite(product.id)}
                            >
                                <Icon 
                                    name="heart" 
                                    size={16} 
                                    color={isFavorite(product.id) ? 'white' : '#ff3e00'}
                                />
                            </Button>
                        </div>
                    </div>
                </Card>
            {/each}
        </div>
    </section>
    
    <section class="favorites">
        <h2>Your Favorites ({favorites.length})</h2>
        {#if favorites.length > 0}
            <p>You have favorited products with IDs: {favorites.join(', ')}</p>
        {:else}
            <p>Click the heart icon on products to add them to favorites!</p>
        {/if}
    </section>
    
    <div class="import-info">
        <h3>Import Structure</h3>
        <pre>
component-library-demo.svelte
├── Button.svelte
├── Card.svelte
└── components/Icon.svelte</pre>
        <p>This demonstrates how components from different directories can be imported and composed together.</p>
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
    
    h2 {
        color: #333;
        margin: 2rem 0 1rem;
    }
    
    .icon-showcase {
        margin: 2rem 0;
    }
    
    .icon-grid {
        display: flex;
        gap: 2rem;
        flex-wrap: wrap;
    }
    
    .icon-item {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 0.5rem;
    }
    
    .icon-item span {
        font-size: 0.875rem;
        color: #666;
    }
    
    .product-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
        gap: 2rem;
        margin: 2rem 0;
    }
    
    .card-actions {
        display: flex;
        flex-direction: column;
        gap: 1rem;
    }
    
    .rating {
        display: flex;
        gap: 2px;
    }
    
    .price {
        font-size: 1.5rem;
        font-weight: bold;
        color: #ff3e00;
    }
    
    .buttons {
        display: flex;
        gap: 0.5rem;
        align-items: center;
    }
    
    .favorites {
        background: #f5f5f5;
        padding: 2rem;
        border-radius: 8px;
        margin: 2rem 0;
    }
    
    .import-info {
        background: #f0f0f0;
        padding: 2rem;
        border-radius: 8px;
        margin-top: 3rem;
    }
    
    .import-info h3 {
        margin: 0 0 1rem 0;
        color: #333;
    }
    
    .import-info pre {
        background: #fff;
        padding: 1rem;
        border-radius: 4px;
        font-family: monospace;
        font-size: 0.875rem;
        line-height: 1.5;
        margin: 1rem 0;
    }
</style>