// Test configuration
const path = require('path');

module.exports = {
  // Test directories
  testDir: path.join(__dirname, '../tests'),
  
  // Test patterns
  testMatch: ['**/*.test.js'],
  
  // Frameworks to test
  frameworks: ['svelte'],
  
  // Common test routes
  routes: {
    svelte: {
      index: '/svelte',
      buttonDemo: '/svelte/button-demo',
      counter: '/svelte/counter',
      asyncDemo: '/svelte/async-demo',
      cardGallery: '/svelte/card-gallery',
      componentLibrary: '/svelte/component-library-demo'
    }
  },
  
  // Test data
  testData: {
    button: {
      primaryText: 'Primary Button',
      secondaryText: 'Secondary Button',
      outlineText: 'Outline Button'
    },
    counter: {
      initialValue: 0,
      incrementValue: 1
    }
  },
  
  // Error messages
  errorMessages: {
    timeout: 'Test timed out',
    elementNotFound: 'Element not found',
    navigationFailed: 'Navigation failed'
  },
  
  // Screenshot settings
  screenshots: {
    onFailure: true,
    path: path.join(__dirname, '../../screenshots'),
    namePattern: '{framework}-{test}-{timestamp}.png'
  }
};