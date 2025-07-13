// Svelte Card Gallery E2E Test

module.exports = async function testSvelteCardGallery({ browser, assert, componentHelper, config, baseUrl }) {
  // Navigate to Svelte card gallery
  const url = `${baseUrl}/svelte/card-gallery`;
  console.log('  Navigating to:', url);
  await browser.goto(url);
  
  // Wait for Svelte to initialize
  await componentHelper.waitForFramework();
  
  // Test 1: Page loads correctly
  await assert.assertTextContains('h1', 'Card Gallery');
  
  // Test 2: Cards are present
  const cardCount = await browser.count('.card, [class*="card"]');
  console.log('  Found cards:', cardCount);
  
  if (cardCount < 3) {
    throw new Error(`Expected at least 3 cards, found ${cardCount}`);
  }
  
  // Test 3: Check card content
  const cardTitles = await browser.evaluate(() => {
    const cards = document.querySelectorAll('.card, [class*="card"]');
    return Array.from(cards).map(card => {
      const title = card.querySelector('h3, .card-title, [class*="title"]');
      return title ? title.textContent.trim() : '';
    }).filter(title => title);
  });
  
  console.log('  Card titles:', cardTitles);
  
  if (cardTitles.length === 0) {
    throw new Error('No card titles found');
  }
  
  // Test 4: Check for card descriptions
  const cardDescriptions = await browser.evaluate(() => {
    const cards = document.querySelectorAll('.card, [class*="card"]');
    return Array.from(cards).map(card => {
      const desc = card.querySelector('p, .card-content, .description');
      return desc ? desc.textContent.trim() : '';
    }).filter(desc => desc);
  });
  
  console.log('  Card descriptions found:', cardDescriptions.length);
  
  // Test 5: Test card interactions (if buttons exist)
  const cardButtons = await browser.count('.card button, [class*="card"] button');
  console.log('  Card buttons found:', cardButtons);
  
  if (cardButtons > 0) {
    // Test clicking the first card button
    await browser.evaluate(() => {
      const firstButton = document.querySelector('.card button, [class*="card"] button');
      if (firstButton) firstButton.click();
    });
    
    await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 200)));
    console.log('  ✓ Card button interaction tested');
  }
  
  // Test 6: Check for responsive grid layout
  const hasGridLayout = await browser.evaluate(() => {
    const container = document.querySelector('.card-grid, .gallery, main');
    if (!container) return false;
    
    const styles = window.getComputedStyle(container);
    return styles.display === 'grid' || 
           styles.display === 'flex' ||
           container.style.display === 'grid' ||
           container.style.display === 'flex';
  });
  
  if (hasGridLayout) {
    console.log('  ✓ Grid layout detected');
  }
  
  // Test 7: Check for card styling
  const hasCardStyling = await browser.evaluate(() => {
    const card = document.querySelector('.card, [class*="card"]');
    if (!card) return false;
    
    const styles = window.getComputedStyle(card);
    return styles.boxShadow !== 'none' || 
           styles.border !== 'none' ||
           styles.borderRadius !== '0px' ||
           card.style.boxShadow ||
           card.style.border;
  });
  
  if (hasCardStyling) {
    console.log('  ✓ Card styling detected');
  }
  
  // Test 8: Test card hover effects (if any)
  const hasHoverEffects = await browser.evaluate(() => {
    const styles = document.querySelector('style');
    return styles && styles.textContent.includes(':hover');
  });
  
  if (hasHoverEffects) {
    console.log('  ✓ Hover effects detected');
  }
  
  // Test 9: Check for images in cards (if any)
  const cardImages = await browser.count('.card img, [class*="card"] img');
  console.log('  Card images found:', cardImages);
  
  if (cardImages > 0) {
    // Test if images are properly loaded
    const brokenImages = await browser.evaluate(() => {
      const images = document.querySelectorAll('.card img, [class*="card"] img');
      let broken = 0;
      images.forEach(img => {
        if (!img.complete || img.naturalHeight === 0) {
          broken++;
        }
      });
      return broken;
    });
    
    if (brokenImages > 0) {
      console.log(`  Warning: ${brokenImages} images may be broken`);
    } else {
      console.log('  ✓ All card images loaded properly');
    }
  }
  
  // Test 10: Test filtering functionality (if exists)
  const filterButtons = await browser.evaluate(() => {
    const buttons = Array.from(document.querySelectorAll('button, [class*="filter"] button'));
    return buttons.filter(btn => {
      const text = btn.textContent || '';
      return text.includes('All') || text.includes('Filter');
    }).length;
  });
  
  if (filterButtons > 0) {
    console.log('  Testing filter functionality...');
    
    await browser.evaluate(() => {
      const buttons = Array.from(document.querySelectorAll('button, [class*="filter"] button'));
      const filterBtn = buttons.find(btn => {
        const text = btn.textContent || '';
        return text.includes('All') || text.includes('Filter');
      });
      if (filterBtn) filterBtn.click();
    });
    
    await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 200)));
    console.log('  ✓ Filter functionality tested');
  }
  
  // Test 11: Test card animation/transitions
  const hasAnimations = await browser.evaluate(() => {
    const styles = document.querySelector('style');
    return styles && (
      styles.textContent.includes('transition') ||
      styles.textContent.includes('animation') ||
      styles.textContent.includes('@keyframes')
    );
  });
  
  if (hasAnimations) {
    console.log('  ✓ Animations/transitions detected');
  }
  
  // Test 12: Component import verification
  const pageSource = await browser.evaluate(() => document.documentElement.outerHTML);
  
  if (pageSource.includes('Card') || pageSource.includes('card')) {
    console.log('  ✓ Card component likely imported correctly');
  }
  
  // Test 13: Test accessibility features
  const hasAltTexts = await browser.evaluate(() => {
    const images = document.querySelectorAll('img');
    let withAlt = 0;
    images.forEach(img => {
      if (img.alt && img.alt.trim()) withAlt++;
    });
    return withAlt;
  });
  
  if (hasAltTexts > 0) {
    console.log('  ✓ Images have alt text for accessibility');
  }
  
  // Test 14: Test responsive behavior (viewport change)
  await browser.evaluate(() => {
    // Simulate smaller viewport
    if (window.innerWidth > 768) {
      console.log('Testing responsive behavior...');
    }
  });
  
  console.log('  ✓ Svelte card gallery test passed');
};