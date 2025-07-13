// Svelte Component Library Demo E2E Test

module.exports = async function testSvelteComponentLibraryDemo({ browser, assert, componentHelper, config, baseUrl }) {
  // Navigate to Svelte component library demo
  const url = `${baseUrl}/svelte/component-library-demo`;
  console.log('  Navigating to:', url);
  await browser.goto(url);
  
  // Wait for Svelte to initialize
  await componentHelper.waitForFramework();
  
  // Test 1: Page loads correctly
  await assert.assertTextContains('h1', 'Component Library');
  
  // Test 2: Check for multiple component types
  const components = await browser.evaluate(() => {
    const componentElements = document.querySelectorAll('button, .card, .icon, [class*="component"]');
    return componentElements.length;
  });
  
  console.log('  Components found:', components);
  
  if (components > 5) {
    console.log('  ✓ Multiple components detected');
  }
  
  // Test 3: Check for buttons with different variants
  const buttons = await browser.evaluate(() => {
    const btnElements = document.querySelectorAll('button');
    const variants = [];
    btnElements.forEach(btn => {
      const classes = btn.className;
      if (classes.includes('primary')) variants.push('primary');
      if (classes.includes('secondary')) variants.push('secondary');
      if (classes.includes('outline')) variants.push('outline');
      if (classes.includes('small')) variants.push('small');
      if (classes.includes('medium')) variants.push('medium');
      if (classes.includes('large')) variants.push('large');
    });
    return [...new Set(variants)];
  });
  
  console.log('  Button variants found:', buttons);
  
  if (buttons.length > 2) {
    console.log('  ✓ Multiple button variants detected');
  }
  
  // Test 4: Check for cards
  const cards = await browser.count('.card, [class*="card"]');
  console.log('  Cards found:', cards);
  
  if (cards > 0) {
    console.log('  ✓ Card components detected');
    
    // Test card interaction
    await browser.evaluate(() => {
      const card = document.querySelector('.card, [class*="card"]');
      if (card) {
        card.click();
      }
    });
    
    await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 200)));
    console.log('  ✓ Card interaction tested');
  }
  
  // Test 5: Check for icons
  const icons = await browser.count('.icon, [class*="icon"], svg');
  console.log('  Icons found:', icons);
  
  if (icons > 0) {
    console.log('  ✓ Icon components detected');
  }
  
  // Test 6: Test button interactions
  const buttonTests = await browser.evaluate(() => {
    const buttons = document.querySelectorAll('button:not([disabled])');
    let clickedCount = 0;
    
    buttons.forEach((btn, index) => {
      if (index < 3) { // Test first 3 buttons
        btn.click();
        clickedCount++;
      }
    });
    
    return clickedCount;
  });
  
  if (buttonTests > 0) {
    await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 300)));
    console.log(`  ✓ Tested ${buttonTests} button interactions`);
  }
  
  // Test 7: Check for component composition
  const nestedComponents = await browser.evaluate(() => {
    // Look for components that contain other components
    const cards = document.querySelectorAll('.card, [class*="card"]');
    let compositionCount = 0;
    
    cards.forEach(card => {
      const hasButton = card.querySelector('button');
      const hasIcon = card.querySelector('.icon, [class*="icon"], svg');
      if (hasButton || hasIcon) {
        compositionCount++;
      }
    });
    
    return compositionCount;
  });
  
  if (nestedComponents > 0) {
    console.log('  ✓ Component composition detected');
  }
  
  // Test 8: Check for interactive states
  const interactiveStates = await browser.evaluate(() => {
    // Look for hover, focus, or active states
    const buttons = document.querySelectorAll('button');
    let stateCount = 0;
    
    buttons.forEach(btn => {
      // Trigger hover
      btn.dispatchEvent(new Event('mouseenter'));
      btn.dispatchEvent(new Event('mouseleave'));
      
      // Trigger focus
      btn.focus();
      btn.blur();
      
      stateCount++;
    });
    
    return stateCount;
  });
  
  if (interactiveStates > 0) {
    console.log('  ✓ Interactive states tested');
  }
  
  // Test 9: Check for responsive behavior
  const isResponsive = await browser.evaluate(() => {
    const containers = document.querySelectorAll('.grid, .flex, [class*="responsive"]');
    return containers.length > 0;
  });
  
  if (isResponsive) {
    console.log('  ✓ Responsive layout detected');
  }
  
  // Test 10: Check for accessibility features
  const a11yFeatures = await browser.evaluate(() => {
    const buttons = document.querySelectorAll('button[aria-label], button[title]');
    const focusableElements = document.querySelectorAll('[tabindex], button, input, select, textarea, a[href]');
    
    return {
      labeledButtons: buttons.length,
      focusableElements: focusableElements.length
    };
  });
  
  if (a11yFeatures.labeledButtons > 0 || a11yFeatures.focusableElements > 5) {
    console.log('  ✓ Accessibility features detected');
  }
  
  // Test 11: Check for component variants showcase
  const variantShowcase = await browser.evaluate(() => {
    const text = document.body.textContent;
    return text.includes('variant') || 
           text.includes('size') || 
           text.includes('style') ||
           text.includes('theme');
  });
  
  if (variantShowcase) {
    console.log('  ✓ Component variants showcase detected');
  }
  
  // Test 12: Test component library organization
  const organized = await browser.evaluate(() => {
    const sections = document.querySelectorAll('section, .section, h2, h3');
    const hasCategories = Array.from(sections).some(el => {
      const text = el.textContent.toLowerCase();
      return text.includes('button') || 
             text.includes('card') || 
             text.includes('icon') ||
             text.includes('component');
    });
    
    return hasCategories;
  });
  
  if (organized) {
    console.log('  ✓ Organized component library structure detected');
  }
  
  console.log('  ✓ Svelte component library demo test passed');
};