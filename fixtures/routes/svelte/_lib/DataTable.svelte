<script>
    export let data = [];
    export let title = "Data Table";
    export let pageSize = 10;
    
    let currentPage = 0;
    let sortColumn = null;
    let sortDirection = 'asc';
    let filterText = '';
    let isLoading = true;
    
    // Simulate data loading
    setTimeout(() => {
        isLoading = false;
    }, 800);
    
    // Generate sample data if none provided
    if (data.length === 0) {
        data = Array.from({ length: 50 }, (_, i) => ({
            id: i + 1,
            name: `User ${i + 1}`,
            email: `user${i + 1}@example.com`,
            role: ['Admin', 'User', 'Editor'][i % 3],
            status: ['Active', 'Inactive'][i % 2],
            created: new Date(2024, 0, 1 + i).toLocaleDateString()
        }));
    }
    
    // Computed values
    $: filteredData = data.filter(item => 
        Object.values(item).some(value => 
            value.toString().toLowerCase().includes(filterText.toLowerCase())
        )
    );
    
    $: sortedData = [...filteredData].sort((a, b) => {
        if (!sortColumn) return 0;
        
        const aVal = a[sortColumn];
        const bVal = b[sortColumn];
        
        if (sortDirection === 'asc') {
            return aVal > bVal ? 1 : -1;
        } else {
            return aVal < bVal ? 1 : -1;
        }
    });
    
    $: pageCount = Math.ceil(sortedData.length / pageSize);
    $: paginatedData = sortedData.slice(currentPage * pageSize, (currentPage + 1) * pageSize);
    
    function handleSort(column) {
        if (sortColumn === column) {
            sortDirection = sortDirection === 'asc' ? 'desc' : 'asc';
        } else {
            sortColumn = column;
            sortDirection = 'asc';
        }
    }
    
    function nextPage() {
        if (currentPage < pageCount - 1) {
            currentPage++;
        }
    }
    
    function prevPage() {
        if (currentPage > 0) {
            currentPage--;
        }
    }
</script>

<div class="data-table">
    <h3>{title}</h3>
    
    {#if isLoading}
        <div class="loading">
            <div class="spinner"></div>
            <p>Loading table data...</p>
        </div>
    {:else}
        <div class="table-controls">
            <input 
                type="text" 
                placeholder="Filter data..." 
                bind:value={filterText}
                class="filter-input"
            />
            <span class="row-count">{filteredData.length} rows</span>
        </div>
        
        <div class="table-container">
            <table>
                <thead>
                    <tr>
                        <th on:click={() => handleSort('id')} class="sortable">
                            ID {sortColumn === 'id' ? (sortDirection === 'asc' ? '↑' : '↓') : ''}
                        </th>
                        <th on:click={() => handleSort('name')} class="sortable">
                            Name {sortColumn === 'name' ? (sortDirection === 'asc' ? '↑' : '↓') : ''}
                        </th>
                        <th on:click={() => handleSort('email')} class="sortable">
                            Email {sortColumn === 'email' ? (sortDirection === 'asc' ? '↑' : '↓') : ''}
                        </th>
                        <th on:click={() => handleSort('role')} class="sortable">
                            Role {sortColumn === 'role' ? (sortDirection === 'asc' ? '↑' : '↓') : ''}
                        </th>
                        <th on:click={() => handleSort('status')} class="sortable">
                            Status {sortColumn === 'status' ? (sortDirection === 'asc' ? '↑' : '↓') : ''}
                        </th>
                        <th on:click={() => handleSort('created')} class="sortable">
                            Created {sortColumn === 'created' ? (sortDirection === 'asc' ? '↑' : '↓') : ''}
                        </th>
                    </tr>
                </thead>
                <tbody>
                    {#each paginatedData as row}
                        <tr>
                            <td>{row.id}</td>
                            <td>{row.name}</td>
                            <td>{row.email}</td>
                            <td>
                                <span class="role-badge {row.role.toLowerCase()}">{row.role}</span>
                            </td>
                            <td>
                                <span class="status-badge {row.status.toLowerCase()}">{row.status}</span>
                            </td>
                            <td>{row.created}</td>
                        </tr>
                    {/each}
                </tbody>
            </table>
        </div>
        
        <div class="pagination">
            <button on:click={prevPage} disabled={currentPage === 0}>
                Previous
            </button>
            <span class="page-info">
                Page {currentPage + 1} of {pageCount}
            </span>
            <button on:click={nextPage} disabled={currentPage >= pageCount - 1}>
                Next
            </button>
        </div>
    {/if}
</div>

<style>
    .data-table {
        border: 2px solid #ddd;
        border-radius: 8px;
        padding: 1rem;
        margin: 1rem 0;
        background: white;
    }
    
    .data-table h3 {
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
    
    .table-controls {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 1rem;
    }
    
    .filter-input {
        padding: 0.5rem;
        border: 1px solid #ddd;
        border-radius: 4px;
        font-size: 0.9rem;
    }
    
    .row-count {
        color: #666;
        font-size: 0.9rem;
    }
    
    .table-container {
        overflow-x: auto;
    }
    
    table {
        width: 100%;
        border-collapse: collapse;
        font-size: 0.9rem;
    }
    
    th, td {
        text-align: left;
        padding: 0.75rem;
        border-bottom: 1px solid #eee;
    }
    
    th {
        background-color: #f5f5f5;
        font-weight: 600;
    }
    
    .sortable {
        cursor: pointer;
        user-select: none;
    }
    
    .sortable:hover {
        background-color: #e0e0e0;
    }
    
    .role-badge, .status-badge {
        padding: 0.25rem 0.5rem;
        border-radius: 12px;
        font-size: 0.8rem;
        font-weight: 500;
    }
    
    .role-badge.admin {
        background-color: #ffebee;
        color: #d32f2f;
    }
    
    .role-badge.editor {
        background-color: #e8f5e8;
        color: #2e7d32;
    }
    
    .role-badge.user {
        background-color: #e3f2fd;
        color: #1976d2;
    }
    
    .status-badge.active {
        background-color: #e8f5e8;
        color: #2e7d32;
    }
    
    .status-badge.inactive {
        background-color: #fafafa;
        color: #757575;
    }
    
    .pagination {
        display: flex;
        justify-content: center;
        align-items: center;
        gap: 1rem;
        margin-top: 1rem;
    }
    
    .pagination button {
        padding: 0.5rem 1rem;
        border: 1px solid #ddd;
        border-radius: 4px;
        background: white;
        cursor: pointer;
    }
    
    .pagination button:disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }
    
    .pagination button:hover:not(:disabled) {
        background-color: #f5f5f5;
    }
    
    .page-info {
        color: #666;
        font-size: 0.9rem;
    }
</style>