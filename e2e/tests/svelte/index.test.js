// Svelte Index Page E2E Test

module.exports = async function testSvelteIndex({ browser, assert, componentHelper, config, baseUrl }) {
  // Navigate to Svelte index page
  const url = `${baseUrl}/svelte`;
  console.log('  Navigating to:', url);
  await browser.goto(url);
  
  // Wait for Svelte to initialize
  await componentHelper.waitForFramework();
  
  // Test 1: Page loads correctly
  await assert.assertTextContains('h1', 'Svelte Framework Demos');
  await assert.assertTextContains('p', 'Explore various Svelte features');
  
  // Test 2: Check demo cards are present
  const demoCards = await browser.count('.demo-card');
  console.log('  Found demo cards:', demoCards);
  
  if (demoCards < 5) {
    throw new Error(`Expected at least 5 demo cards, found ${demoCards}`);
  }
  
  // Test 3: Check specific demo cards
  const cardTitles = await browser.evaluate(() => {
    return Array.from(document.querySelectorAll('.demo-card h3')).map(h3 => h3.textContent);
  });
  console.log('  Demo card titles:', cardTitles);
  
  const expectedTitles = ['Counter Demo', 'Button Demo', 'Card Gallery', 'Component Library'];
  for (const title of expectedTitles) {
    if (!cardTitles.some(cardTitle => cardTitle.includes(title))) {
      throw new Error(`Missing demo card: ${title}`);
    }
  }
  
  // Test 4: Test navigation links
  const firstCard = '.demo-card:first-child';
  await assert.assertExists(firstCard);
  
  const href = await browser.getAttribute(firstCard, 'href');
  console.log('  First card href:', href);
  
  if (!href || !href.includes('/svelte/')) {
    throw new Error('Demo card links are not properly formatted');
  }
  
  // Test 5: Test framework info section
  await assert.assertExists('.framework-info');
  await assert.assertTextContains('.framework-info h2', 'About Svelte');
  
  const features = await browser.count('.framework-info li');
  console.log('  Svelte features listed:', features);
  
  if (features < 4) {
    throw new Error(`Expected at least 4 Svelte features, found ${features}`);
  }
  
  // Test 6: Test hover effects (CSS)
  const hasHoverStyle = await browser.evaluate(() => {
    const style = document.querySelector('style');
    return style && style.textContent.includes(':hover');
  });
  
  if (!hasHoverStyle) {
    throw new Error('Demo cards should have hover effects');
  }
  
  console.log('  ✓ Svelte index page test passed');
};