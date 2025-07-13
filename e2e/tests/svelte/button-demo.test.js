// Svelte Button Demo E2E Test

module.exports = async function testSvelteButtonDemo({ browser, assert, componentHelper, config, baseUrl }) {
  // Navigate to Svelte button demo
  const url = `${baseUrl}/svelte/button-demo`;
  console.log('  Navigating to:', url);
  await browser.goto(url);
  
  // Wait for Svelte to initialize
  await componentHelper.waitForFramework();
  
  // Test 1: Page loads correctly
  await assert.assertTextContains('h1', 'Svelte Component Import Demo');
  await assert.assertTextContains('p', 'This demonstrates importing and using another Svelte component');
  
  // Test 2: All button variants are present
  const buttonTexts = await browser.evaluate(() => {
    return Array.from(document.querySelectorAll('button')).map(btn => btn.textContent.trim());
  });
  console.log('  Found buttons:', buttonTexts);
  
  const expectedButtons = ['Primary Button', 'Secondary Button', 'Outline Button', 'Small', 'Medium', 'Large'];
  for (const expected of expectedButtons) {
    if (!buttonTexts.includes(expected)) {
      throw new Error(`Missing button: ${expected}`);
    }
  }
  
  // Test 3: Disabled buttons
  const disabledButtons = await browser.count('button[disabled]');
  console.log('  Disabled buttons count:', disabledButtons);
  
  if (disabledButtons !== 3) {
    throw new Error(`Expected 3 disabled buttons, found ${disabledButtons}`);
  }
  
  // Test 4: Click counter functionality - initial state
  await assert.assertTextContains('.stats', 'Total clicks: 0');
  
  // Test 5: Click primary button
  await browser.evaluate(() => {
    const btn = Array.from(document.querySelectorAll('button')).find(b => 
      b.textContent.trim() === 'Primary Button' && !b.disabled
    );
    if (btn) btn.click();
  });
  await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 200)));
  
  const stats1 = await browser.getText('.stats');
  console.log('  After primary click:', stats1);
  
  if (!stats1.includes('Total clicks: 1')) {
    throw new Error('Primary button click should increment counter to 1');
  }
  
  if (!stats1.includes('Last clicked: primary')) {
    throw new Error('Should show last clicked variant as primary');
  }
  
  // Test 6: Click secondary button
  await browser.evaluate(() => {
    const btn = Array.from(document.querySelectorAll('button')).find(b => 
      b.textContent.trim() === 'Secondary Button' && !b.disabled
    );
    if (btn) btn.click();
  });
  await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 200)));
  
  const stats2 = await browser.getText('.stats');
  console.log('  After secondary click:', stats2);
  
  if (!stats2.includes('Total clicks: 2')) {
    throw new Error('Secondary button click should increment counter to 2');
  }
  
  if (!stats2.includes('Last clicked: secondary')) {
    throw new Error('Should show last clicked variant as secondary');
  }
  
  // Test 7: Click outline button
  await browser.evaluate(() => {
    const btn = Array.from(document.querySelectorAll('button')).find(b => 
      b.textContent.trim() === 'Outline Button' && !b.disabled
    );
    if (btn) btn.click();
  });
  await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 200)));
  
  const stats3 = await browser.getText('.stats');
  console.log('  After outline click:', stats3);
  
  if (!stats3.includes('Total clicks: 3')) {
    throw new Error('Outline button click should increment counter to 3');
  }
  
  if (!stats3.includes('Last clicked: outline')) {
    throw new Error('Should show last clicked variant as outline');
  }
  
  // Test 8: Reset counter
  const hasResetButton = await browser.evaluate(() => {
    return Array.from(document.querySelectorAll('button')).some(btn => 
      btn.textContent.includes('Reset Counter')
    );
  });
  
  if (hasResetButton) {
    await browser.evaluate(() => {
      const btn = Array.from(document.querySelectorAll('button')).find(b => 
        b.textContent.includes('Reset Counter')
      );
      if (btn) btn.click();
    });
    await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 200)));
    
    const statsAfterReset = await browser.getText('.stats');
    console.log('  After reset:', statsAfterReset);
    
    if (!statsAfterReset.includes('Total clicks: 0')) {
      throw new Error('Reset should set counter back to 0');
    }
  }
  
  // Test 9: Button styling - check CSS classes
  const hasButtonStyles = await browser.evaluate(() => {
    const buttons = document.querySelectorAll('button');
    return Array.from(buttons).some(btn => {
      const classes = btn.className;
      return classes.includes('btn') || btn.style.padding || btn.style.backgroundColor;
    });
  });
  
  if (!hasButtonStyles) {
    console.log('  Warning: Buttons may not have proper styling');
  }
  
  // Test 10: Test disabled button interaction
  const currentStats = await browser.getText('.stats');
  await browser.evaluate(() => {
    const disabledBtn = document.querySelector('button[disabled]');
    if (disabledBtn) {
      disabledBtn.click();
    }
  });
  await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 100)));
  
  const statsAfterDisabledClick = await browser.getText('.stats');
  if (currentStats !== statsAfterDisabledClick) {
    throw new Error('Disabled button should not affect counter');
  }
  
  // Test 11: Component import verification
  const pageSource = await browser.evaluate(() => document.documentElement.outerHTML);
  if (!pageSource.includes('Button') && !pageSource.includes('btn')) {
    throw new Error('Button component may not be properly imported');
  }
  
  console.log('  ✓ Svelte button demo test passed');
};