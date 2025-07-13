<script>
	let count = 0;
	$: doubled = count * 2;
	
	function increment() {
		count += 1;
	}
	
	function decrement() {
		count -= 1;
	}
	
	function reset() {
		count = 0;
	}
</script>

<main>
	<h1>Svelte Counter Demo</h1>
	
	<div class="counter">
		<button on:click={decrement} disabled={count <= 0}>
			-
		</button>
		
		<span class="count">{count}</span>
		
		<button on:click={increment}>
			+
		</button>
	</div>
	
	<p>Doubled: {doubled}</p>
	
	<button class="reset" on:click={reset} disabled={count === 0}>
		Reset
	</button>
</main>

<style>
	main {
		text-align: center;
		padding: 2em;
		max-width: 400px;
		margin: 0 auto;
		font-family: system-ui, -apple-system, sans-serif;
	}

	h1 {
		color: #ff3e00;
		text-transform: uppercase;
		font-size: 2.5em;
		font-weight: 100;
		margin-bottom: 1em;
	}

	.counter {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 1em;
		margin: 2em 0;
	}

	.count {
		font-size: 3em;
		font-weight: 200;
		min-width: 2em;
	}

	button {
		font-size: 1.5em;
		padding: 0.5em 1em;
		border: 2px solid #ff3e00;
		background: white;
		color: #ff3e00;
		border-radius: 4px;
		cursor: pointer;
		transition: all 0.2s;
	}

	button:hover:not(:disabled) {
		background: #ff3e00;
		color: white;
	}

	button:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.reset {
		font-size: 1em;
		margin-top: 1em;
	}

	p {
		font-size: 1.2em;
		color: #666;
	}
</style>