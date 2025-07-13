// Svelte Data Table Demo E2E Test

module.exports = async function testSvelteDataTableDemo({ browser, assert, componentHelper, config, baseUrl }) {
  // Navigate to Svelte data table demo
  const url = `${baseUrl}/svelte/data-table-demo`;
  console.log('  Navigating to:', url);
  await browser.goto(url);
  
  // Wait for Svelte to initialize
  await componentHelper.waitForFramework();
  
  // Test 1: Page loads correctly
  await assert.assertTextContains('h1', 'Data Table');
  
  // Test 2: Check for data table presence
  const hasTable = await browser.exists('table, .data-table, [class*="table"]');
  
  if (hasTable) {
    console.log('  ✓ Data table found');
    
    // Test table headers
    const headers = await browser.evaluate(() => {
      const headerElements = document.querySelectorAll('th, [class*="header"], .table-header');
      return Array.from(headerElements).map(el => el.textContent.trim()).filter(text => text.length > 0);
    });
    
    console.log('  Table headers:', headers);
    
    if (headers.length > 0) {
      console.log('  ✓ Table headers detected');
    }
  }
  
  // Test 3: Check for table rows/data
  const tableRows = await browser.count('tr, .table-row, [class*="row"]');
  console.log('  Table rows found:', tableRows);
  
  if (tableRows > 1) { // More than just header row
    console.log('  ✓ Table data detected');
  }
  
  // Test 4: Check for sorting functionality
  const sortableElements = await browser.evaluate(() => {
    // Look for sortable indicators
    const sortElements = document.querySelectorAll('[class*="sort"], .sortable, th[data-sort]');
    return sortElements.length;
  });
  
  if (sortableElements > 0) {
    console.log('  ✓ Sortable elements detected');
    
    // Test clicking on a sortable header
    const sortClicked = await browser.evaluate(() => {
      const sortableHeaders = document.querySelectorAll('th, [class*="header"]');
      if (sortableHeaders.length > 0) {
        sortableHeaders[0].click();
        return true;
      }
      return false;
    });
    
    if (sortClicked) {
      await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 200)));
      console.log('  ✓ Sort interaction tested');
    }
  }
  
  // Test 5: Check for filtering functionality
  const filterElements = await browser.count('input[type="search"], input[placeholder*="filter"], input[placeholder*="search"], .filter, [class*="filter"]');
  console.log('  Filter elements found:', filterElements);
  
  if (filterElements > 0) {
    console.log('  ✓ Filter functionality detected');
    
    // Test search/filter input
    const searchInput = await browser.evaluate(() => {
      const inputs = document.querySelectorAll('input[type="search"], input[placeholder*="filter"], input[placeholder*="search"]');
      if (inputs.length > 0) {
        inputs[0].value = 'test';
        inputs[0].dispatchEvent(new Event('input', { bubbles: true }));
        return true;
      }
      return false;
    });
    
    if (searchInput) {
      await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 300)));
      console.log('  ✓ Search/filter tested');
    }
  }
  
  // Test 6: Check for selection functionality
  const checkboxes = await browser.count('input[type="checkbox"]');
  console.log('  Checkboxes found:', checkboxes);
  
  if (checkboxes > 0) {
    console.log('  ✓ Row selection functionality detected');
    
    // Test checkbox interaction
    await browser.evaluate(() => {
      const checkbox = document.querySelector('input[type="checkbox"]');
      if (checkbox) {
        checkbox.click();
      }
    });
    
    await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 200)));
    console.log('  ✓ Checkbox interaction tested');
  }
  
  // Test 7: Check for pagination
  const paginationElements = await browser.count('.pagination, [class*="page"], button[class*="page"]');
  
  if (paginationElements > 0) {
    console.log('  ✓ Pagination detected');
    
    // Test pagination interaction
    const paginationClicked = await browser.evaluate(() => {
      const pageButtons = document.querySelectorAll('button[class*="page"], .pagination button');
      if (pageButtons.length > 0) {
        // Find a clickable page button
        for (const btn of pageButtons) {
          if (!btn.disabled && btn.textContent.match(/\d|next|prev/i)) {
            btn.click();
            return true;
          }
        }
      }
      return false;
    });
    
    if (paginationClicked) {
      await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 300)));
      console.log('  ✓ Pagination interaction tested');
    }
  }
  
  // Test 8: Check for responsive design
  const isResponsive = await browser.evaluate(() => {
    const table = document.querySelector('table, .data-table, [class*="table"]');
    if (table) {
      const styles = window.getComputedStyle(table);
      return styles.overflowX === 'auto' || styles.overflowX === 'scroll' || 
             table.closest('.table-responsive') !== null;
    }
    return false;
  });
  
  if (isResponsive) {
    console.log('  ✓ Responsive design detected');
  }
  
  // Test 9: Check for data formatting
  const hasFormattedData = await browser.evaluate(() => {
    const cells = document.querySelectorAll('td, .table-cell, [class*="cell"]');
    return Array.from(cells).some(cell => {
      const text = cell.textContent;
      // Look for formatted dates, emails, or other patterns
      return text.match(/\d{4}-\d{2}-\d{2}/) || 
             text.match(/\w+@\w+\.\w+/) || 
             text.match(/Active|Inactive|Pending/i);
    });
  });
  
  if (hasFormattedData) {
    console.log('  ✓ Formatted data detected');
  }
  
  // Test 10: Check for loading states or empty states
  const hasStates = await browser.evaluate(() => {
    const text = document.body.textContent;
    return text.includes('Loading') || 
           text.includes('No data') || 
           text.includes('Empty') ||
           document.querySelector('.loading, .empty, [class*="loading"], [class*="empty"]');
  });
  
  if (hasStates) {
    console.log('  ✓ Loading/empty states detected');
  }
  
  // Test 11: Test bulk actions if available
  const bulkActions = await browser.evaluate(() => {
    // Look for bulk action buttons or menus
    const actions = document.querySelectorAll('button[class*="bulk"], .bulk-actions, [class*="action"]');
    return actions.length;
  });
  
  if (bulkActions > 0) {
    console.log('  ✓ Bulk actions detected');
  }
  
  // Test 12: Check for column customization
  const columnControls = await browser.evaluate(() => {
    // Look for column show/hide controls
    const controls = document.querySelectorAll('[class*="column"], button[class*="col"]');
    return controls.length;
  });
  
  if (columnControls > 0) {
    console.log('  ✓ Column controls detected');
  }
  
  console.log('  ✓ Svelte data table demo test passed');
};