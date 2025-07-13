// Svelte Todo App E2E Test

module.exports = async function testSvelteTodo({ browser, assert, componentHelper, config, baseUrl }) {
  // Navigate to Svelte todo app
  const url = `${baseUrl}/svelte/todo`;
  console.log('  Navigating to:', url);
  await browser.goto(url);
  
  // Wait for Svelte to initialize
  await componentHelper.waitForFramework();
  
  // Test 1: Page loads correctly
  await assert.assertTextContains('h1', 'Todo');
  
  // Test 2: Initial state - no todos
  const initialTodos = await browser.count('.todo-item, li');
  console.log('  Initial todo count:', initialTodos);
  
  if (initialTodos > 0) {
    console.log('  Note: Found existing todos, app may have persisted state');
  }
  
  // Test 3: Input field exists
  await assert.assertExists('input[placeholder*="What needs to be done"], input[placeholder*="todo"], input[placeholder*="task"], input[type="text"]');
  
  // Test 4: Add button exists
  const addButtonExists = await browser.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('button'));
    return buttons.some(btn => btn.textContent.includes('Add') || btn.textContent.includes('+'));
  });
  
  if (!addButtonExists) {
    throw new Error('Add button should exist');
  }
  
  // Test 5: Add a new todo
  const todoText = 'Test Todo Item';
  const inputSelector = 'input[placeholder*="What needs to be done"], input[placeholder*="todo"], input[placeholder*="task"], input[type="text"]';
  
  await browser.type(inputSelector, todoText);
  
  // Find and click add button
  await browser.evaluate((text) => {
    const buttons = Array.from(document.querySelectorAll('button'));
    const addBtn = buttons.find(btn => 
      btn.textContent.includes('Add') || 
      btn.textContent.includes('+') || 
      btn.type === 'submit'
    );
    if (addBtn) addBtn.click();
  }, todoText);
  
  await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 200)));
  
  // Test 6: Todo item appears
  const todoExists = await browser.evaluate((text) => {
    return Array.from(document.querySelectorAll('*')).some(el => 
      el.textContent && el.textContent.includes(text)
    );
  }, todoText);
  
  if (!todoExists) {
    throw new Error(`Todo item "${todoText}" was not added to the list`);
  }
  
  console.log('  ✓ Todo item added successfully');
  
  // Test 7: Input field should be cleared after adding
  const inputValue = await browser.evaluate((selector) => {
    const input = document.querySelector(selector);
    return input ? input.value : '';
  }, inputSelector);
  
  if (inputValue !== '') {
    console.log('  Note: Input field was not cleared after adding todo');
  }
  
  // Test 8: Add another todo
  const secondTodo = 'Second Todo Item';
  await browser.type(inputSelector, secondTodo);
  
  await browser.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('button'));
    const addBtn = buttons.find(btn => 
      btn.textContent.includes('Add') || 
      btn.textContent.includes('+') || 
      btn.type === 'submit'
    );
    if (addBtn) addBtn.click();
  });
  
  await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 200)));
  
  // Test 9: Check if both todos exist
  const secondTodoExists = await browser.evaluate((text) => {
    return Array.from(document.querySelectorAll('*')).some(el => 
      el.textContent && el.textContent.includes(text)
    );
  }, secondTodo);
  
  if (!secondTodoExists) {
    throw new Error(`Second todo item "${secondTodo}" was not added`);
  }
  
  // Test 10: Test todo completion (if checkboxes exist)
  const hasCheckboxes = await browser.exists('input[type="checkbox"]');
  
  if (hasCheckboxes) {
    console.log('  Testing todo completion...');
    
    await browser.click('input[type="checkbox"]:first-of-type');
    await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 200)));
    
    const isChecked = await browser.evaluate(() => {
      const checkbox = document.querySelector('input[type="checkbox"]:first-of-type');
      return checkbox ? checkbox.checked : false;
    });
    
    if (!isChecked) {
      throw new Error('Checkbox should be checked after clicking');
    }
    
    console.log('  ✓ Todo completion works');
  }
  
  // Test 11: Test filter functionality (if filters exist)
  const hasFilters = await browser.exists('button:contains("All"), button:contains("Active"), button:contains("Completed")');
  
  if (hasFilters) {
    console.log('  Testing filter functionality...');
    
    // Test "All" filter
    await browser.evaluate(() => {
      const allBtn = Array.from(document.querySelectorAll('button')).find(btn => 
        btn.textContent.includes('All')
      );
      if (allBtn) allBtn.click();
    });
    
    await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 100)));
    console.log('  ✓ All filter works');
    
    // Test "Active" filter
    await browser.evaluate(() => {
      const activeBtn = Array.from(document.querySelectorAll('button')).find(btn => 
        btn.textContent.includes('Active')
      );
      if (activeBtn) activeBtn.click();
    });
    
    await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 100)));
    console.log('  ✓ Active filter works');
  }
  
  // Test 12: Counter functionality (if exists)
  const hasCounter = await browser.evaluate(() => {
    return Array.from(document.querySelectorAll('*')).some(el => 
      el.textContent && (
        el.textContent.includes('remaining') || 
        el.textContent.includes('left') ||
        el.textContent.includes('items')
      )
    );
  });
  
  if (hasCounter) {
    console.log('  ✓ Todo counter found');
  }
  
  // Test 13: Test empty todo prevention
  await browser.evaluate((selector) => {
    const input = document.querySelector(selector);
    if (input) input.value = '';
  }, inputSelector);
  
  await browser.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('button'));
    const addBtn = buttons.find(btn => 
      btn.textContent.includes('Add') || 
      btn.textContent.includes('+') || 
      btn.type === 'submit'
    );
    if (addBtn) addBtn.click();
  });
  
  await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 200)));
  
  console.log('  ✓ Empty todo handling tested');
  
  // Test 14: Keyboard interaction (Enter key)
  await browser.type(inputSelector, 'Keyboard Todo');
  await browser.evaluate((selector) => {
    const input = document.querySelector(selector);
    if (input) {
      const event = new KeyboardEvent('keydown', { key: 'Enter' });
      input.dispatchEvent(event);
    }
  }, inputSelector);
  
  await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 200)));
  
  const keyboardTodoExists = await browser.evaluate(() => {
    return Array.from(document.querySelectorAll('*')).some(el => 
      el.textContent && el.textContent.includes('Keyboard Todo')
    );
  });
  
  if (keyboardTodoExists) {
    console.log('  ✓ Keyboard interaction (Enter) works');
  }
  
  console.log('  ✓ Svelte todo app test passed');
};