<script>
    export let data = [];
    export let title = "Chart";
    export let width = 400;
    export let height = 300;
    
    let chartContainer;
    let isLoading = true;
    
    // Simulate heavy initialization
    setTimeout(() => {
        isLoading = false;
    }, 1000);
    
    // Generate some sample data if none provided
    if (data.length === 0) {
        data = Array.from({ length: 10 }, (_, i) => ({
            label: `Item ${i + 1}`,
            value: Math.floor(Math.random() * 100)
        }));
    }
    
    function getMaxValue() {
        return Math.max(...data.map(d => d.value));
    }
    
    function getBarHeight(value) {
        return (value / getMaxValue()) * (height - 60);
    }
</script>

<div class="heavy-chart" bind:this={chartContainer}>
    <h3>{title}</h3>
    
    {#if isLoading}
        <div class="loading">
            <div class="spinner"></div>
            <p>Loading chart data...</p>
        </div>
    {:else}
        <div class="chart-container">
            <svg {width} {height} class="chart">
                {#each data as item, i}
                    <g transform="translate({i * (width / data.length)}, 0)">
                        <rect 
                            x="5" 
                            y={height - getBarHeight(item.value) - 40}
                            width={width / data.length - 10}
                            height={getBarHeight(item.value)}
                            fill="#ff3e00"
                            opacity="0.8"
                        />
                        <text 
                            x={width / data.length / 2} 
                            y={height - 25}
                            text-anchor="middle"
                            font-size="12"
                            fill="#333"
                        >
                            {item.label}
                        </text>
                        <text 
                            x={width / data.length / 2} 
                            y={height - getBarHeight(item.value) - 45}
                            text-anchor="middle"
                            font-size="10"
                            fill="#666"
                        >
                            {item.value}
                        </text>
                    </g>
                {/each}
            </svg>
        </div>
    {/if}
</div>

<style>
    .heavy-chart {
        border: 2px solid #ddd;
        border-radius: 8px;
        padding: 1rem;
        margin: 1rem 0;
        background: #f9f9f9;
    }
    
    .heavy-chart h3 {
        margin: 0 0 1rem 0;
        color: #333;
        font-size: 1.2rem;
    }
    
    .loading {
        display: flex;
        flex-direction: column;
        align-items: center;
        padding: 2rem;
        color: #666;
    }
    
    .spinner {
        width: 32px;
        height: 32px;
        border: 3px solid #f3f3f3;
        border-top: 3px solid #ff3e00;
        border-radius: 50%;
        animation: spin 1s linear infinite;
        margin-bottom: 1rem;
    }
    
    @keyframes spin {
        0% { transform: rotate(0deg); }
        100% { transform: rotate(360deg); }
    }
    
    .chart-container {
        display: flex;
        justify-content: center;
        margin: 1rem 0;
    }
    
    .chart {
        border: 1px solid #ddd;
        border-radius: 4px;
        background: white;
    }
    
    .chart rect {
        transition: opacity 0.3s ease;
    }
    
    .chart rect:hover {
        opacity: 1;
    }
    
    .chart text {
        pointer-events: none;
    }
</style>