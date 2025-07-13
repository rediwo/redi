// Svelte Counter Component E2E Test

module.exports = async function testSvelteCounter({ browser, assert, componentHelper, config, baseUrl }) {
  // Navigate to Svelte counter
  const url = `${baseUrl}/svelte/counter`;
  console.log('  Navigating to:', url);
  await browser.goto(url);
  
  // Wait for Svelte to initialize
  await componentHelper.waitForFramework();
  
  // Test 1: Page loads with counter
  await assert.assertTextContains('h1', 'Svelte Counter Demo');
  
  // Test 2: Initial counter value
  const initialCount = await browser.getText('.count');
  console.log('  Initial counter value:', initialCount);
  
  if (initialCount !== '0') {
    throw new Error(`Expected initial count to be 0, got ${initialCount}`);
  }
  
  // Test 3: Initial doubled value
  await assert.assertTextContains('p', 'Doubled: 0');
  
  // Test 4: Decrement button should be disabled initially
  const decrementButton = 'button:first-of-type';
  await assert.assertDisabled(decrementButton);
  
  // Test 5: Reset button should be disabled initially
  const resetButton = '.reset';
  await assert.assertDisabled(resetButton);
  
  // Test 6: Increment functionality
  const incrementButton = 'button:nth-of-type(2)';
  await browser.click(incrementButton);
  await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 100)));
  
  const countAfterIncrement = await browser.getText('.count');
  console.log('  Count after increment:', countAfterIncrement);
  
  if (countAfterIncrement !== '1') {
    throw new Error(`Expected count to be 1 after increment, got ${countAfterIncrement}`);
  }
  
  // Test 7: Doubled value updates reactively
  await assert.assertTextContains('p', 'Doubled: 2');
  
  // Test 8: Decrement button should be enabled now
  const isDecrementDisabled = await browser.evaluate((selector) => {
    return document.querySelector(selector).disabled;
  }, decrementButton);
  
  if (isDecrementDisabled) {
    throw new Error('Decrement button should be enabled when count > 0');
  }
  
  // Test 9: Reset button should be enabled now
  const isResetDisabled = await browser.evaluate((selector) => {
    return document.querySelector(selector).disabled;
  }, resetButton);
  
  if (isResetDisabled) {
    throw new Error('Reset button should be enabled when count > 0');
  }
  
  // Test 10: Multiple increments
  for (let i = 0; i < 3; i++) {
    await browser.click(incrementButton);
    await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 50)));
  }
  
  const countAfterMultiple = await browser.getText('.count');
  console.log('  Count after multiple increments:', countAfterMultiple);
  
  if (countAfterMultiple !== '4') {
    throw new Error(`Expected count to be 4, got ${countAfterMultiple}`);
  }
  
  await assert.assertTextContains('p', 'Doubled: 8');
  
  // Test 11: Decrement functionality
  await browser.click(decrementButton);
  await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 100)));
  
  const countAfterDecrement = await browser.getText('.count');
  console.log('  Count after decrement:', countAfterDecrement);
  
  if (countAfterDecrement !== '3') {
    throw new Error(`Expected count to be 3 after decrement, got ${countAfterDecrement}`);
  }
  
  await assert.assertTextContains('p', 'Doubled: 6');
  
  // Test 12: Reset functionality
  await browser.click(resetButton);
  await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 100)));
  
  const countAfterReset = await browser.getText('.count');
  console.log('  Count after reset:', countAfterReset);
  
  if (countAfterReset !== '0') {
    throw new Error(`Expected count to be 0 after reset, got ${countAfterReset}`);
  }
  
  await assert.assertTextContains('p', 'Doubled: 0');
  
  // Test 13: Buttons should be disabled again after reset
  await assert.assertDisabled(decrementButton);
  await assert.assertDisabled(resetButton);
  
  // Test 14: Rapid clicking stress test
  console.log('  Testing rapid increments...');
  for (let i = 0; i < 10; i++) {
    await browser.click(incrementButton);
  }
  
  await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 300)));
  const finalCount = await browser.getText('.count');
  console.log('  Final count after rapid clicks:', finalCount);
  
  if (finalCount !== '10') {
    throw new Error(`Expected final count to be 10, got ${finalCount}`);
  }
  
  await assert.assertTextContains('p', 'Doubled: 20');
  
  console.log('  ✓ Svelte counter test passed');
};