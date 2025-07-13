<script>
    let count = 0;
    
    function increment() {
        count += 1;
    }
    
    function decrement() {
        count -= 1;
    }
</script>

<div class="max-w-md mx-auto bg-white rounded-xl shadow-lg overflow-hidden">
    <div class="p-8">
        <div class="text-center">
            <h1 class="text-3xl font-bold text-gray-900 mb-4">Vimesh Style Test</h1>
            <p class="text-gray-600 mb-6">Testing Tailwind CSS classes with Vimesh Style runtime</p>
        </div>
        
        <div class="bg-gray-50 rounded-lg p-6 mb-6">
            <div class="text-center">
                <div class="text-6xl font-bold text-blue-600 mb-4">{count}</div>
                <p class="text-gray-500 text-sm">Current counter value</p>
            </div>
        </div>
        
        <div class="flex space-x-4 justify-center">
            <button 
                on:click={decrement}
                class="bg-red-500 hover:bg-red-600 text-white font-medium py-2 px-4 rounded-lg transition-colors duration-200 shadow-sm"
            >
                Decrement
            </button>
            
            <button 
                on:click={increment}
                class="bg-green-500 hover:bg-green-600 text-white font-medium py-2 px-4 rounded-lg transition-colors duration-200 shadow-sm"
            >
                Increment
            </button>
        </div>
        
        <div class="mt-8 p-4 bg-blue-50 rounded-lg">
            <h2 class="text-lg font-semibold text-blue-900 mb-2">Features Demo</h2>
            <div class="grid grid-cols-2 gap-4">
                <div class="bg-white p-3 rounded border border-blue-200">
                    <div class="text-sm font-medium text-blue-700">Responsive</div>
                    <div class="text-xs text-blue-500">Works on all devices</div>
                </div>
                <div class="bg-white p-3 rounded border border-green-200">
                    <div class="text-sm font-medium text-green-700">Fast</div>
                    <div class="text-xs text-green-500">Vimesh Style runtime</div>
                </div>
            </div>
        </div>
        
        <div class="mt-6 text-center">
            <p class="text-xs text-gray-400">
                Powered by Svelte + Vimesh Style + Redi
            </p>
        </div>
    </div>
</div>

<style>
    /* Custom styles can still be used alongside Vimesh Style */
    .custom-animation {
        animation: pulse 2s infinite;
    }
</style>