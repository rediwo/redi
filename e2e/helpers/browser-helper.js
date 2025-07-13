const puppeteer = require('puppeteer');
const puppeteerConfig = require('../config/puppeteer.config');
const fs = require('fs').promises;
const path = require('path');

class BrowserHelper {
  constructor() {
    this.browser = null;
    this.page = null;
  }

  // Launch browser
  async launch() {
    this.browser = await puppeteer.launch(puppeteerConfig.launch);
    this.page = await this.browser.newPage();
    await this.page.setViewport(puppeteerConfig.viewport);
    
    // Set up console logging
    this.page.on('console', msg => {
      const type = msg.type();
      if (type === 'error') {
        console.error('Browser console error:', msg.text());
      }
    });
    
    // Set up error handling
    this.page.on('pageerror', error => {
      console.error('Page error:', error.message);
    });
    
    return this.page;
  }

  // Navigate to URL
  async goto(url, options = {}) {
    const navigationOptions = {
      ...puppeteerConfig.navigationOptions,
      ...options
    };
    
    try {
      await this.page.goto(url, navigationOptions);
    } catch (error) {
      console.error(`Navigation failed to ${url}:`, error.message);
      throw error;
    }
  }

  // Wait for element
  async waitForElement(selector, options = {}) {
    const timeout = options.timeout || puppeteerConfig.timeouts.element;
    
    try {
      await this.page.waitForSelector(selector, { timeout, ...options });
    } catch (error) {
      console.error(`Element not found: ${selector}`);
      await this.screenshot(`element-not-found-${Date.now()}`);
      throw error;
    }
  }

  // Click element
  async click(selector, options = {}) {
    await this.waitForElement(selector, options);
    await this.page.click(selector);
  }

  // Type text
  async type(selector, text, options = {}) {
    await this.waitForElement(selector, options);
    await this.page.type(selector, text, { delay: 100, ...options });
  }

  // Get text content
  async getText(selector) {
    await this.waitForElement(selector);
    return await this.page.$eval(selector, el => el.textContent);
  }

  // Get attribute
  async getAttribute(selector, attribute) {
    await this.waitForElement(selector);
    return await this.page.$eval(selector, (el, attr) => el.getAttribute(attr), attribute);
  }

  // Check if element exists
  async exists(selector) {
    try {
      await this.page.waitForSelector(selector, { timeout: 1000 });
      return true;
    } catch {
      return false;
    }
  }

  // Count elements
  async count(selector) {
    return await this.page.$$eval(selector, elements => elements.length);
  }

  // Execute JavaScript
  async evaluate(fn, ...args) {
    return await this.page.evaluate(fn, ...args);
  }

  // Take screenshot
  async screenshot(name, options = {}) {
    const screenshotDir = path.join(__dirname, '../../screenshots');
    await fs.mkdir(screenshotDir, { recursive: true });
    
    const filename = `${name}-${Date.now()}.png`;
    const filepath = path.join(screenshotDir, filename);
    
    await this.page.screenshot({
      path: filepath,
      ...puppeteerConfig.screenshotOptions,
      ...options
    });
    
    console.log(`Screenshot saved: ${filename}`);
    return filepath;
  }

  // Wait for navigation
  async waitForNavigation(options = {}) {
    return await this.page.waitForNavigation({
      ...puppeteerConfig.navigationOptions,
      ...options
    });
  }

  // Reload page
  async reload() {
    await this.page.reload(puppeteerConfig.navigationOptions);
  }

  // Get current URL
  async getUrl() {
    return this.page.url();
  }

  // Close browser
  async close() {
    if (this.browser) {
      await this.browser.close();
      this.browser = null;
      this.page = null;
    }
  }
}

module.exports = BrowserHelper;