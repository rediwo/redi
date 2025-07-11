<script>
    export let title = '';
    export let description = '';
    export let imageUrl = '';
    export let showButton = true;
</script>

<div class="card">
    {#if imageUrl}
        <img src={imageUrl} alt={title} class="card-image" />
    {/if}
    <div class="card-content">
        {#if title}
            <h3 class="card-title">{title}</h3>
        {/if}
        {#if description}
            <p class="card-description">{description}</p>
        {/if}
        {#if showButton}
            <div class="card-actions">
                <slot name="actions">
                    <button class="card-button">Learn More</button>
                </slot>
            </div>
        {/if}
    </div>
</div>

<style>
    .card {
        background: white;
        border-radius: 8px;
        box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
        overflow: hidden;
        transition: transform 0.2s, box-shadow 0.2s;
    }
    
    .card:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    }
    
    .card-image {
        width: 100%;
        height: 200px;
        object-fit: cover;
    }
    
    .card-content {
        padding: 1.5rem;
    }
    
    .card-title {
        margin: 0 0 0.5rem 0;
        font-size: 1.25rem;
        font-weight: 600;
        color: #333;
    }
    
    .card-description {
        margin: 0 0 1rem 0;
        color: #666;
        line-height: 1.5;
    }
    
    .card-actions {
        margin-top: 1rem;
    }
    
    .card-button {
        background: #ff3e00;
        color: white;
        border: none;
        padding: 0.5rem 1rem;
        border-radius: 4px;
        cursor: pointer;
        font-size: 0.875rem;
        transition: background 0.2s;
    }
    
    .card-button:hover {
        background: #e63600;
    }
</style>