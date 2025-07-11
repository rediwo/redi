<script>
    export let variant = 'primary';
    export let size = 'medium';
    export let disabled = false;
</script>

<button 
    class="btn btn-{variant} btn-{size}" 
    class:disabled={disabled}
    on:click
    {disabled}
>
    <slot>Click me</slot>
</button>

<style>
    .btn {
        padding: 0.5rem 1rem;
        border: none;
        border-radius: 4px;
        font-family: inherit;
        font-size: 1rem;
        cursor: pointer;
        transition: all 0.2s ease;
        outline: none;
    }
    
    .btn:hover:not(.disabled) {
        transform: translateY(-1px);
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    }
    
    .btn:active:not(.disabled) {
        transform: translateY(0);
    }
    
    /* Variants */
    .btn-primary {
        background-color: #ff3e00;
        color: white;
    }
    
    .btn-primary:hover:not(.disabled) {
        background-color: #e63600;
    }
    
    .btn-secondary {
        background-color: #6c757d;
        color: white;
    }
    
    .btn-secondary:hover:not(.disabled) {
        background-color: #5a6268;
    }
    
    .btn-outline {
        background-color: transparent;
        color: #ff3e00;
        border: 2px solid #ff3e00;
    }
    
    .btn-outline:hover:not(.disabled) {
        background-color: #ff3e00;
        color: white;
    }
    
    /* Sizes */
    .btn-small {
        padding: 0.25rem 0.5rem;
        font-size: 0.875rem;
    }
    
    .btn-medium {
        padding: 0.5rem 1rem;
        font-size: 1rem;
    }
    
    .btn-large {
        padding: 0.75rem 1.5rem;
        font-size: 1.25rem;
    }
    
    /* Disabled state */
    .btn.disabled {
        opacity: 0.6;
        cursor: not-allowed;
    }
</style>