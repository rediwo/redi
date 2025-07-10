<script>
  let tasks = [
    { id: 1, text: 'Learn Svelte', done: false },
    { id: 2, text: 'Use Vimesh Style', done: true },
    { id: 3, text: 'Build awesome apps', done: false }
  ];
  
  function addTask() {
    const text = prompt('Enter a new task:');
    if (text) {
      tasks = [...tasks, { id: Date.now(), text, done: false }];
    }
  }
  
  function toggleTask(id) {
    tasks = tasks.map(task => 
      task.id === id ? { ...task, done: !task.done } : task
    );
  }
  
  function deleteTask(id) {
    tasks = tasks.filter(task => task.id !== id);
  }
</script>

<div class="min-h-screen bg-gray-100 py-6 flex flex-col justify-center sm:py-12">
  <div class="relative py-3 sm:max-w-xl sm:mx-auto">
    <div class="absolute inset-0 bg-gradient-to-r from-cyan-400 to-light-blue-500 shadow-lg transform -skew-y-6 sm:skew-y-0 sm:-rotate-6 sm:rounded-3xl"></div>
    <div class="relative px-4 py-10 bg-white shadow-lg sm:rounded-3xl sm:p-20">
      <div class="max-w-md mx-auto">
        <div class="divide-y divide-gray-200">
          <div class="py-8 text-base leading-6 space-y-4 text-gray-700 sm:text-lg sm:leading-7">
            <h1 class="text-3xl font-bold text-gray-900 mb-8">Todo List with Vimesh Style</h1>
            
            <button 
              on:click={addTask}
              class="w-full bg-blue-500 hover:bg-blue-600 text-white font-bold py-2 px-4 rounded-lg transition-colors duration-200 mb-6"
            >
              Add New Task
            </button>
            
            <ul class="space-y-2">
              {#each tasks as task (task.id)}
                <li class="flex items-center p-3 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors">
                  <input 
                    type="checkbox" 
                    checked={task.done}
                    on:change={() => toggleTask(task.id)}
                    class="h-5 w-5 text-blue-600 rounded focus:ring-blue-500"
                  />
                  <span class="ml-3 flex-grow {task.done ? 'line-through text-gray-500' : 'text-gray-900'}">
                    {task.text}
                  </span>
                  <button
                    on:click={() => deleteTask(task.id)}
                    class="ml-3 text-red-500 hover:text-red-700 transition-colors"
                  >
                    <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
                    </svg>
                  </button>
                </li>
              {/each}
            </ul>
            
            {#if tasks.length === 0}
              <p class="text-center text-gray-500 py-8">No tasks yet. Add one above!</p>
            {/if}
          </div>
        </div>
      </div>
    </div>
  </div>
</div>

<style>
  /* Component-specific animations */
  li {
    animation: slideIn 0.3s ease-out;
  }
  
  @keyframes slideIn {
    from {
      opacity: 0;
      transform: translateX(-10px);
    }
    to {
      opacity: 1;
      transform: translateX(0);
    }
  }
</style>