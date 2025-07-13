<script>
    import DataTable from './_lib/DataTable.svelte';
    
    // Sample data for the table
    const users = [
        { id: 1, name: 'Alice Johnson', email: 'alice@example.com', role: 'Admin', status: 'Active', joinDate: '2023-01-15' },
        { id: 2, name: 'Bob Smith', email: 'bob@example.com', role: 'User', status: 'Active', joinDate: '2023-02-20' },
        { id: 3, name: 'Carol Davis', email: 'carol@example.com', role: 'Editor', status: 'Inactive', joinDate: '2023-03-10' },
        { id: 4, name: 'David Wilson', email: 'david@example.com', role: 'User', status: 'Active', joinDate: '2023-04-05' },
        { id: 5, name: 'Eve Brown', email: 'eve@example.com', role: 'Admin', status: 'Active', joinDate: '2023-05-12' },
        { id: 6, name: 'Frank Miller', email: 'frank@example.com', role: 'User', status: 'Pending', joinDate: '2023-06-18' },
        { id: 7, name: 'Grace Lee', email: 'grace@example.com', role: 'Editor', status: 'Active', joinDate: '2023-07-03' },
        { id: 8, name: 'Henry Taylor', email: 'henry@example.com', role: 'User', status: 'Inactive', joinDate: '2023-08-22' },
        { id: 9, name: 'Ivy Chen', email: 'ivy@example.com', role: 'Admin', status: 'Active', joinDate: '2023-09-14' },
        { id: 10, name: 'Jack Rodriguez', email: 'jack@example.com', role: 'User', status: 'Active', joinDate: '2023-10-08' }
    ];
    
    const columns = [
        { key: 'id', title: 'ID', sortable: true, width: '60px' },
        { key: 'name', title: 'Name', sortable: true },
        { key: 'email', title: 'Email', sortable: true },
        { key: 'role', title: 'Role', sortable: true, filterable: true },
        { key: 'status', title: 'Status', sortable: true, filterable: true },
        { key: 'joinDate', title: 'Join Date', sortable: true, type: 'date' }
    ];
    
    let selectedRows = [];
    
    function handleRowSelect(event) {
        selectedRows = event.detail.selectedRows;
    }
    
    function handleSort(event) {
        console.log('Sort:', event.detail);
    }
    
    function handleFilter(event) {
        console.log('Filter:', event.detail);
    }
</script>

<main>
    <h1>Data Table Demo</h1>
    <p>Advanced data display with sorting, filtering, and selection capabilities.</p>
    
    <section class="demo-section">
        <h2>User Management Table</h2>
        <p>This table demonstrates various data table features:</p>
        <ul>
            <li>Sortable columns (click column headers)</li>
            <li>Filterable columns (Role and Status)</li>
            <li>Row selection with checkboxes</li>
            <li>Responsive design</li>
            <li>Custom cell formatting</li>
        </ul>
        
        <div class="table-container">
            <DataTable 
                data={users}
                columns={columns}
                selectable={true}
                searchable={true}
                pageSize={5}
                on:rowSelect={handleRowSelect}
                on:sort={handleSort}
                on:filter={handleFilter}
            />
        </div>
        
        {#if selectedRows.length > 0}
        <div class="selection-info">
            <h3>Selected Rows</h3>
            <p>{selectedRows.length} row(s) selected:</p>
            <ul>
                {#each selectedRows as row}
                <li>{row.name} ({row.email})</li>
                {/each}
            </ul>
        </div>
        {/if}
    </section>
    
    <section class="features-section">
        <h2>Data Table Features</h2>
        <div class="features-grid">
            <div class="feature">
                <h3>Sorting</h3>
                <p>Click any column header to sort the data. Click again to reverse the sort order.</p>
            </div>
            <div class="feature">
                <h3>Filtering</h3>
                <p>Some columns support filtering. Look for the filter icon in the header.</p>
            </div>
            <div class="feature">
                <h3>Selection</h3>
                <p>Use checkboxes to select individual rows or select all with the header checkbox.</p>
            </div>
            <div class="feature">
                <h3>Search</h3>
                <p>Use the search box to find specific records across all columns.</p>
            </div>
            <div class="feature">
                <h3>Pagination</h3>
                <p>Large datasets are automatically paginated for better performance.</p>
            </div>
            <div class="feature">
                <h3>Responsive</h3>
                <p>The table adapts to different screen sizes and devices.</p>
            </div>
        </div>
    </section>
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
        margin-bottom: 1rem;
    }
    
    h2 {
        color: #333;
        margin: 2rem 0 1rem 0;
        border-bottom: 2px solid #ff3e00;
        padding-bottom: 0.5rem;
    }
    
    h3 {
        color: #555;
        margin: 1rem 0 0.5rem 0;
    }
    
    .demo-section {
        margin: 2rem 0;
    }
    
    .table-container {
        margin: 2rem 0;
        background: white;
        border-radius: 8px;
        box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        overflow: hidden;
    }
    
    .selection-info {
        margin: 1rem 0;
        padding: 1rem;
        background: #e8f4fd;
        border-radius: 8px;
        border: 1px solid #b3d9f2;
    }
    
    .selection-info ul {
        margin: 0.5rem 0;
        padding-left: 1.5rem;
    }
    
    .features-section {
        margin: 3rem 0;
    }
    
    .features-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
        gap: 1.5rem;
        margin: 2rem 0;
    }
    
    .feature {
        padding: 1.5rem;
        background: #f9f9f9;
        border-radius: 8px;
        border: 1px solid #ddd;
    }
    
    .feature h3 {
        color: #ff3e00;
        margin-top: 0;
    }
    
    ul {
        margin: 1rem 0;
        padding-left: 2rem;
    }
    
    li {
        margin: 0.5rem 0;
        line-height: 1.6;
    }
</style>