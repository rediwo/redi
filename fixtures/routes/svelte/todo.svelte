<script>
	let todos = [];
	let newTodo = '';
	let filter = 'all';
	
	$: filteredTodos = filter === 'all' 
		? todos
		: filter === 'active'
		? todos.filter(t => !t.done)
		: todos.filter(t => t.done);
	
	$: remaining = todos.filter(t => !t.done).length;
	
	function addTodo() {
		if (newTodo.trim()) {
			todos = [...todos, { 
				id: Date.now(), 
				text: newTodo.trim(), 
				done: false 
			}];
			newTodo = '';
		}
	}
	
	function toggleTodo(id) {
		todos = todos.map(todo =>
			todo.id === id ? { ...todo, done: !todo.done } : todo
		);
	}
	
	function removeTodo(id) {
		todos = todos.filter(todo => todo.id !== id);
	}
	
	function clearCompleted() {
		todos = todos.filter(todo => !todo.done);
	}
</script>

<main>
	<h1>Svelte Todo App</h1>
	
	<div class="input-container">
		<input
			bind:value={newTodo}
			on:keydown={e => e.key === 'Enter' && addTodo()}
			placeholder="What needs to be done?"
		/>
		<button on:click={addTodo}>Add</button>
	</div>
	
	<div class="filters">
		<button 
			class:active={filter === 'all'} 
			on:click={() => filter = 'all'}>
			All
		</button>
		<button 
			class:active={filter === 'active'} 
			on:click={() => filter = 'active'}>
			Active
		</button>
		<button 
			class:active={filter === 'completed'} 
			on:click={() => filter = 'completed'}>
			Completed
		</button>
	</div>
	
	<ul class="todo-list">
		{#each filteredTodos as todo (todo.id)}
			<li class:done={todo.done}>
				<input
					type="checkbox"
					checked={todo.done}
					on:change={() => toggleTodo(todo.id)}
				/>
				<span>{todo.text}</span>
				<button class="remove" on:click={() => removeTodo(todo.id)}>×</button>
			</li>
		{:else}
			<li class="empty">No todos yet!</li>
		{/each}
	</ul>
	
	{#if todos.length > 0}
		<div class="footer">
			<span>{remaining} {remaining === 1 ? 'item' : 'items'} left</span>
			{#if todos.some(t => t.done)}
				<button on:click={clearCompleted}>Clear completed</button>
			{/if}
		</div>
	{/if}
</main>

<style>
	main {
		max-width: 600px;
		margin: 0 auto;
		padding: 2em;
		font-family: system-ui, -apple-system, sans-serif;
	}
	
	h1 {
		color: #ff3e00;
		text-align: center;
		margin-bottom: 1em;
	}
	
	.input-container {
		display: flex;
		gap: 0.5em;
		margin-bottom: 1em;
	}
	
	input[type="text"] {
		flex: 1;
		padding: 0.5em;
		font-size: 1.2em;
		border: 2px solid #ddd;
		border-radius: 4px;
	}
	
	button {
		padding: 0.5em 1em;
		font-size: 1em;
		background: #ff3e00;
		color: white;
		border: none;
		border-radius: 4px;
		cursor: pointer;
	}
	
	button:hover {
		background: #e62e00;
	}
	
	.filters {
		display: flex;
		gap: 0.5em;
		margin-bottom: 1em;
		justify-content: center;
	}
	
	.filters button {
		background: white;
		color: #333;
		border: 2px solid #ddd;
	}
	
	.filters button.active {
		background: #ff3e00;
		color: white;
		border-color: #ff3e00;
	}
	
	.todo-list {
		list-style: none;
		padding: 0;
		margin: 0;
		border: 2px solid #ddd;
		border-radius: 4px;
	}
	
	.todo-list li {
		display: flex;
		align-items: center;
		gap: 0.5em;
		padding: 1em;
		border-bottom: 1px solid #eee;
	}
	
	.todo-list li:last-child {
		border-bottom: none;
	}
	
	.todo-list li.done span {
		text-decoration: line-through;
		opacity: 0.5;
	}
	
	.todo-list li.empty {
		text-align: center;
		color: #999;
		font-style: italic;
	}
	
	.todo-list span {
		flex: 1;
	}
	
	.remove {
		background: #dc3545;
		color: white;
		width: 30px;
		height: 30px;
		padding: 0;
		font-size: 1.5em;
		line-height: 1;
	}
	
	.footer {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-top: 1em;
		padding: 0.5em;
		color: #666;
	}
	
	input[type="checkbox"] {
		width: 20px;
		height: 20px;
		cursor: pointer;
	}
</style>