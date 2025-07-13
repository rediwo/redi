// Component-specific test utilities
const puppeteerConfig = require('../config/puppeteer.config');

class ComponentHelper {
  constructor(browser, framework) {
    this.browser = browser;
    this.framework = framework;
  }

  // Get framework-specific selectors
  getSelectors() {
    // Common selectors that work across frameworks
    return {
      button: {
        primary: 'button:not([class*="secondary"]):not([class*="outline"])',
        secondary: 'button[class*="secondary"]',
        outline: 'button[class*="outline"]',
        byText: (text) => `button:has-text("${text}")`,
        disabled: 'button[disabled]'
      },
      counter: {
        display: '[data-testid="counter-display"], .counter-display, p:has-text("Count:")',
        incrementBtn: 'button:has-text("Increment"), button:has-text("+")',
        decrementBtn: 'button:has-text("Decrement"), button:has-text("-")',
        resetBtn: 'button:has-text("Reset")'
      },
      form: {
        input: (name) => `input[name="${name}"]`,
        textarea: (name) => `textarea[name="${name}"]`,
        select: (name) => `select[name="${name}"]`,
        submit: 'button[type="submit"], input[type="submit"]'
      },
      card: {
        container: '.card, [class*="card"]',
        title: '.card-title, .card h2, .card h3',
        content: '.card-content, .card p',
        image: '.card img'
      }
    };
  }

  // Wait for framework to be ready
  async waitForFramework() {
    switch (this.framework) {
      case 'react':
        await this.waitForReact();
        break;
      case 'vue':
        await this.waitForVue();
        break;
      case 'svelte':
        await this.waitForSvelte();
        break;
    }
  }

  // Wait for React to be ready
  async waitForReact() {
    await this.browser.evaluate(() => {
      return new Promise((resolve, reject) => {
        const timeout = setTimeout(() => {
          reject(new Error('React initialization timeout'));
        }, 10000);
        
        const checkReact = () => {
          if (window.React && window.ReactDOM) {
            // Check if React has rendered content
            const root = document.getElementById('root');
            if (root && root.children.length > 0) {
              clearTimeout(timeout);
              resolve();
            } else {
              setTimeout(checkReact, 100);
            }
          } else {
            setTimeout(checkReact, 100);
          }
        };
        
        checkReact();
      });
    });
  }

  // Wait for Vue to be ready
  async waitForVue() {
    await this.browser.evaluate(() => {
      return new Promise((resolve) => {
        if (window.Vue || window.__VUE__) {
          resolve();
        } else {
          const checkVue = setInterval(() => {
            if (window.Vue || window.__VUE__) {
              clearInterval(checkVue);
              resolve();
            }
          }, 100);
        }
      });
    });
  }

  // Wait for Svelte to be ready
  async waitForSvelte() {
    // Svelte doesn't have a global object, so we wait for the app to render
    await this.browser.waitForElement('body > *');
  }

  // Test button interaction
  async testButton(buttonSelector, expectedBehavior) {
    await this.browser.click(buttonSelector);
    
    if (expectedBehavior.stateChange) {
      await this.browser.waitForElement(expectedBehavior.stateChange.selector);
      const newValue = await this.browser.getText(expectedBehavior.stateChange.selector);
      return newValue;
    }
    
    return true;
  }

  // Test counter component
  async testCounter(options = {}) {
    const selectors = this.getSelectors().counter;
    const results = {};
    
    // Get initial value
    results.initialValue = await this.browser.getText(selectors.display);
    
    // Test increment
    if (options.testIncrement) {
      await this.browser.click(selectors.incrementBtn);
      await this.browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 100)));
      results.afterIncrement = await this.browser.getText(selectors.display);
    }
    
    // Test decrement
    if (options.testDecrement && await this.browser.exists(selectors.decrementBtn)) {
      await this.browser.click(selectors.decrementBtn);
      await this.browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 100)));
      results.afterDecrement = await this.browser.getText(selectors.display);
    }
    
    // Test reset
    if (options.testReset && await this.browser.exists(selectors.resetBtn)) {
      await this.browser.click(selectors.resetBtn);
      await this.browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 100)));
      results.afterReset = await this.browser.getText(selectors.display);
    }
    
    return results;
  }

  // Test form submission
  async testForm(formData) {
    const selectors = this.getSelectors().form;
    
    // Fill form fields
    for (const [fieldName, value] of Object.entries(formData)) {
      const inputSelector = selectors.input(fieldName);
      if (await this.browser.exists(inputSelector)) {
        await this.browser.type(inputSelector, value);
      }
    }
    
    // Submit form
    await this.browser.click(selectors.submit);
    
    // Wait for response (could be navigation or AJAX)
    await this.browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 500)));
  }

  // Test async component loading
  async testAsyncLoading(triggerSelector, componentSelector, timeout = 5000) {
    // Click trigger to load component
    await this.browser.click(triggerSelector);
    
    // Wait for component to appear
    const startTime = Date.now();
    while (Date.now() - startTime < timeout) {
      if (await this.browser.exists(componentSelector)) {
        return true;
      }
      await this.browser.evaluate(() => new Promise(resolve => setTimeout(resolve, 100)));
    }
    
    throw new Error(`Async component did not load within ${timeout}ms`);
  }

  // Get component state (framework-specific)
  async getComponentState(componentSelector) {
    switch (this.framework) {
      case 'react':
        return await this.getReactState(componentSelector);
      case 'vue':
        return await this.getVueState(componentSelector);
      case 'svelte':
        return await this.getSvelteState(componentSelector);
      default:
        return null;
    }
  }

  // Get React component state
  async getReactState(selector) {
    return await this.browser.evaluate((sel) => {
      const element = document.querySelector(sel);
      if (!element || !element._reactInternalFiber) return null;
      
      let fiber = element._reactInternalFiber;
      while (fiber && !fiber.stateNode?.state) {
        fiber = fiber.return;
      }
      
      return fiber?.stateNode?.state || null;
    }, selector);
  }

  // Get Vue component state
  async getVueState(selector) {
    return await this.browser.evaluate((sel) => {
      const element = document.querySelector(sel);
      if (!element || !element.__vue__) return null;
      return element.__vue__.$data;
    }, selector);
  }

  // Get Svelte component state
  async getSvelteState(selector) {
    // Svelte doesn't expose component state easily
    // We'd need to instrument the component during build
    return null;
  }

  // Compare component behavior across frameworks
  async compareAcrossFrameworks(testFn, frameworks = ['react', 'vue', 'svelte']) {
    const results = {};
    
    for (const fw of frameworks) {
      const url = `${puppeteerConfig.getBaseUrl()}/${fw}`;
      await this.browser.goto(url);
      await this.waitForFramework();
      
      results[fw] = await testFn(fw);
    }
    
    return results;
  }
}

module.exports = ComponentHelper;