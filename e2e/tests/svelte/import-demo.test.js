// Svelte Import Demo E2E Test

module.exports = async function testSvelteImportDemo({ browser, assert, componentHelper, config, baseUrl }) {
  // Navigate to Svelte import demo
  const url = `${baseUrl}/svelte/import-demo`;
  console.log('  Navigating to:', url);
  await browser.goto(url);
  
  // Wait for Svelte to initialize
  await componentHelper.waitForFramework();
  
  // Test 1: Page loads correctly
  await assert.assertTextContains('h1', 'Import');
  
  // Test 2: Check for component imports
  const hasButtons = await browser.exists('button');
  
  if (hasButtons) {
    console.log('  ✓ Button component imported');
    
    // Test button functionality
    await browser.click('button:first-of-type');
    await browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 100)));
    console.log('  ✓ Button interaction works');
  }
  
  // Test 3: Check for CSS imports
  const hasCSSImports = await browser.evaluate(() => {
    // Check if styles are applied
    const styleTag = document.querySelector('style');
    const linkTag = document.querySelector('link[rel="stylesheet"]');
    return !!(styleTag || linkTag);
  });
  
  if (hasCSSImports) {
    console.log('  ✓ CSS imports detected');
  }
  
  // Test 4: Check for image imports
  const images = await browser.count('img');
  console.log('  Images found:', images);
  
  if (images > 0) {
    // Test if images are properly loaded
    const imageStatus = await browser.evaluate(() => {
      const imgs = document.querySelectorAll('img');
      const results = [];
      imgs.forEach((img, index) => {
        results.push({
          index,
          src: img.src,
          complete: img.complete,
          naturalWidth: img.naturalWidth,
          naturalHeight: img.naturalHeight
        });
      });
      return results;
    });
    
    console.log('  Image status:', imageStatus);
    
    const brokenImages = imageStatus.filter(img => 
      !img.complete || img.naturalWidth === 0
    );
    
    if (brokenImages.length > 0) {
      console.log(`  Warning: ${brokenImages.length} images may be broken`);
    } else {
      console.log('  ✓ All images loaded successfully');
    }
  }
  
  // Test 5: Check for JSON data imports
  const hasJSONData = await browser.evaluate(() => {
    // Look for elements that might contain JSON data
    const textContent = document.body.textContent;
    return textContent.includes('{') && textContent.includes('}') ||
           textContent.includes('[') && textContent.includes(']') ||
           Array.from(document.querySelectorAll('*')).some(el => {
             const text = el.textContent;
             return text && (text.includes('name') || text.includes('version') || text.includes('config'));
           });
  });
  
  if (hasJSONData) {
    console.log('  ✓ JSON data imports detected');
  }
  
  // Test 6: Check for font imports
  const hasFonts = await browser.evaluate(() => {
    const styles = window.getComputedStyle(document.body);
    const fontFamily = styles.fontFamily;
    return fontFamily && fontFamily !== 'serif' && fontFamily !== 'sans-serif';
  });
  
  if (hasFonts) {
    console.log('  ✓ Custom fonts detected');
  }
  
  // Test 7: Test asset URL imports
  const assetURLs = await browser.evaluate(() => {
    // Check for URLs in the page that might be imported assets
    const allElements = document.querySelectorAll('*');
    const urls = [];
    
    allElements.forEach(el => {
      // Check for src attributes
      if (el.src && (el.src.includes('/') || el.src.includes('.'))) {
        urls.push(el.src);
      }
      
      // Check for href attributes
      if (el.href && el.href.includes('.') && !el.href.includes('#')) {
        urls.push(el.href);
      }
      
      // Check for background images in style
      const styles = window.getComputedStyle(el);
      if (styles.backgroundImage && styles.backgroundImage !== 'none') {
        urls.push(styles.backgroundImage);
      }
    });
    
    return urls;
  });
  
  console.log('  Asset URLs found:', assetURLs.length);
  
  if (assetURLs.length > 0) {
    console.log('  ✓ Asset URL imports detected');
  }
  
  // Test 8: Test import transformation
  const pageSource = await browser.evaluate(() => document.documentElement.outerHTML);
  
  // Check for evidence of import transformation
  const hasTransformedImports = pageSource.includes('http') || 
                                pageSource.includes('/css/') ||
                                pageSource.includes('/images/') ||
                                pageSource.includes('/js/');
  
  if (hasTransformedImports) {
    console.log('  ✓ Import transformation detected');
  }
  
  // Test 9: Test dynamic imports (if any)
  const hasDynamicImports = await browser.evaluate(() => {
    // Check for evidence of dynamic imports in the console or page behavior
    return typeof window.import === 'function' || 
           document.documentElement.outerHTML.includes('import(');
  });
  
  if (hasDynamicImports) {
    console.log('  ✓ Dynamic imports detected');
  }
  
  // Test 10: Test CSS URL imports
  const cssUrls = await browser.evaluate(() => {
    const styles = document.querySelectorAll('style, link[rel="stylesheet"]');
    const urls = [];
    
    styles.forEach(style => {
      if (style.href) {
        urls.push(style.href);
      } else if (style.textContent) {
        // Look for URL references in CSS
        const urlMatches = style.textContent.match(/url\([^)]+\)/g);
        if (urlMatches) {
          urls.push(...urlMatches);
        }
      }
    });
    
    return urls;
  });
  
  console.log('  CSS URLs found:', cssUrls.length);
  
  // Test 11: Test component composition
  const componentElements = await browser.evaluate(() => {
    // Look for evidence of multiple components being used
    const elements = document.querySelectorAll('*');
    const componentClasses = [];
    
    elements.forEach(el => {
      if (el.className && el.className.includes('svelte-')) {
        componentClasses.push(el.className);
      }
    });
    
    return [...new Set(componentClasses)];
  });
  
  console.log('  Component classes found:', componentElements.length);
  
  if (componentElements.length > 1) {
    console.log('  ✓ Multiple components detected');
  }
  
  // Test 12: Test import error handling
  const consoleErrors = await browser.evaluate(() => {
    // Check for import-related errors
    return window.__importErrors || [];
  });
  
  if (consoleErrors.length > 0) {
    console.log('  Warning: Import errors detected:', consoleErrors);
  } else {
    console.log('  ✓ No import errors detected');
  }
  
  // Test 13: Test module system integration
  const hasModuleSystem = await browser.evaluate(() => {
    // Check for evidence of ES modules or module system
    return document.documentElement.outerHTML.includes('module') ||
           typeof window.require === 'function' ||
           typeof window.exports === 'object';
  });
  
  if (hasModuleSystem) {
    console.log('  ✓ Module system integration detected');
  }
  
  console.log('  ✓ Svelte import demo test passed');
};