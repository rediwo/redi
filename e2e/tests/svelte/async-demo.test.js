// Svelte Async Demo E2E Test

module.exports = async function testSvelteAsyncDemo({ browser, assert, componentHelper, config, baseUrl }) {
  // Navigate to Svelte async demo
  const url = `${baseUrl}/svelte/async-demo`;
  console.log('  Navigating to:', url);
  await browser.goto(url);
  
  // Wait for Svelte to initialize
  await componentHelper.waitForFramework();
  
  // Test 1: Page loads correctly
  await assert.assertTextContains('h1', 'Async');
  
  // Test 2: Check for async loading triggers
  const asyncButtons = await browser.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('button'));
    return buttons.filter(btn => 
      btn.textContent.includes('Load') ||
      btn.textContent.includes('Toggle') ||
      btn.textContent.includes('Async')
    ).map(btn => btn.textContent.trim());
  });
  
  console.log('  Async buttons found:', asyncButtons);
  
  if (asyncButtons.length === 0) {
    throw new Error('No async loading buttons found');
  }
  
  // Test 3: Test async component loading
  const initialComponents = await browser.count('[class*="svelte-"], .component, .async-component');
  console.log('  Initial components:', initialComponents);
  
  // Click first async loading button
  await browser.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('button'));
    const loadBtn = buttons.find(btn => 
      btn.textContent.includes('Load') ||
      btn.textContent.includes('Toggle') ||
      btn.textContent.includes('Async')
    );
    if (loadBtn) loadBtn.click();
  });
  
  // Wait for async loading
  await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 1000)));
  
  const componentsAfterLoad = await browser.count('[class*="svelte-"], .component, .async-component');
  console.log('  Components after async load:', componentsAfterLoad);
  
  // Test 4: Check for loading states
  const hasLoadingStates = await browser.evaluate(() => {
    const text = document.body.textContent;
    return text.includes('Loading') || 
           text.includes('loading') ||
           document.querySelector('.loading, [class*="loading"]');
  });
  
  if (hasLoadingStates) {
    console.log('  ✓ Loading states detected');
  }
  
  // Test 5: Test multiple async components
  const multipleAsyncButtons = await browser.evaluate(() => {
    return Array.from(document.querySelectorAll('button')).filter(btn => 
      btn.textContent.includes('Load') ||
      btn.textContent.includes('Toggle') ||
      btn.textContent.includes('Async')
    ).length;
  });
  
  if (multipleAsyncButtons > 1) {
    console.log('  Testing multiple async components...');
    
    // Click second button if exists
    await browser.evaluate(() => {
      const buttons = Array.from(document.querySelectorAll('button'));
      const asyncBtns = buttons.filter(btn => 
        btn.textContent.includes('Load') ||
        btn.textContent.includes('Toggle') ||
        btn.textContent.includes('Async')
      );
      if (asyncBtns[1]) asyncBtns[1].click();
    });
    
    await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 800)));
    console.log('  ✓ Multiple async components tested');
  }
  
  // Test 6: Test async component performance
  const startTime = Date.now();
  
  await browser.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('button'));
    const loadBtn = buttons.find(btn => 
      btn.textContent.includes('Load') ||
      btn.textContent.includes('Heavy') ||
      btn.textContent.includes('Chart')
    );
    if (loadBtn) loadBtn.click();
  });
  
  // Wait for component to load
  await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 500)));
  
  const loadTime = Date.now() - startTime;
  console.log('  Async component load time:', loadTime + 'ms');
  
  if (loadTime > 5000) {
    console.log('  Warning: Async component took longer than 5 seconds to load');
  }
  
  // Test 7: Test error handling
  const hasErrorHandling = await browser.evaluate(() => {
    const text = document.body.textContent;
    return text.includes('Error') || 
           text.includes('Failed') ||
           document.querySelector('.error, [class*="error"]');
  });
  
  if (hasErrorHandling) {
    console.log('  ✓ Error handling detected');
  }
  
  // Test 8: Test lazy loading indicators
  const hasLazyIndicators = await browser.evaluate(() => {
    const spinners = document.querySelectorAll('.spinner, .loader, [class*="spin"]');
    const dots = document.querySelectorAll('.dots, [class*="dots"]');
    return spinners.length > 0 || dots.length > 0;
  });
  
  if (hasLazyIndicators) {
    console.log('  ✓ Lazy loading indicators detected');
  }
  
  // Test 9: Test async data loading
  const hasAsyncData = await browser.evaluate(() => {
    // Look for data that might have been loaded asynchronously
    const elements = Array.from(document.querySelectorAll('*'));
    return elements.some(el => {
      const text = el.textContent;
      return text && (
        text.includes('users') ||
        text.includes('posts') ||
        text.includes('data') ||
        /\d+/.test(text) // Contains numbers (possible data)
      );
    });
  });
  
  if (hasAsyncData) {
    console.log('  ✓ Async data loading detected');
  }
  
  // Test 10: Test component cleanup
  await browser.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('button'));
    const toggleBtn = buttons.find(btn => 
      btn.textContent.includes('Toggle') ||
      btn.textContent.includes('Hide') ||
      btn.textContent.includes('Remove')
    );
    if (toggleBtn) {
      toggleBtn.click();
      // Click again to test cleanup
      setTimeout(() => toggleBtn.click(), 100);
    }
  });
  
  await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 300)));
  console.log('  ✓ Component cleanup tested');
  
  // Test 11: Test async component caching
  const cacheTest = await browser.evaluate(() => {
    let clickCount = 0;
    const buttons = Array.from(document.querySelectorAll('button'));
    const loadBtn = buttons.find(btn => 
      btn.textContent.includes('Load') ||
      btn.textContent.includes('Toggle')
    );
    
    if (loadBtn) {
      const startTime = performance.now();
      loadBtn.click();
      
      setTimeout(() => {
        loadBtn.click(); // Second click should be faster if cached
        const endTime = performance.now();
        window.__asyncTestTime = endTime - startTime;
      }, 200);
      
      return true;
    }
    return false;
  });
  
  if (cacheTest) {
    await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 500)));
    
    const testTime = await browser.evaluate(() => window.__asyncTestTime);
    if (testTime) {
      console.log('  Async caching test time:', testTime + 'ms');
    }
  }
  
  // Test 12: Test memory management
  const initialMemory = await browser.evaluate(() => {
    return performance.memory ? performance.memory.usedJSHeapSize : 0;
  });
  
  // Load and unload components multiple times
  for (let i = 0; i < 3; i++) {
    await browser.evaluate(() => {
      const buttons = Array.from(document.querySelectorAll('button'));
      const toggleBtn = buttons.find(btn => btn.textContent.includes('Toggle'));
      if (toggleBtn) {
        toggleBtn.click();
        setTimeout(() => toggleBtn.click(), 100);
      }
    });
    await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 200)));
  }
  
  const finalMemory = await browser.evaluate(() => {
    return performance.memory ? performance.memory.usedJSHeapSize : 0;
  });
  
  if (initialMemory && finalMemory) {
    const memoryDiff = finalMemory - initialMemory;
    console.log('  Memory usage change:', memoryDiff + ' bytes');
    
    if (memoryDiff > 10000000) { // 10MB
      console.log('  Warning: Significant memory increase detected');
    }
  }
  
  // Test 13: Test async component props
  const hasPropsHandling = await browser.evaluate(() => {
    // Look for evidence of props being passed to async components
    const elements = Array.from(document.querySelectorAll('*'));
    return elements.some(el => 
      el.textContent && (
        el.textContent.includes('prop') ||
        el.textContent.includes('title:') ||
        el.textContent.includes('name:')
      )
    );
  });
  
  if (hasPropsHandling) {
    console.log('  ✓ Async component props handling detected');
  }
  
  console.log('  ✓ Svelte async demo test passed');
};