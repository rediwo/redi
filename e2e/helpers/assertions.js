// Custom assertions for E2E tests

class Assertions {
  constructor(browser) {
    this.browser = browser;
  }

  // Assert element exists
  async assertExists(selector, message) {
    const exists = await this.browser.exists(selector);
    if (!exists) {
      throw new Error(message || `Element not found: ${selector}`);
    }
  }

  // Assert element does not exist
  async assertNotExists(selector, message) {
    const exists = await this.browser.exists(selector);
    if (exists) {
      throw new Error(message || `Element should not exist: ${selector}`);
    }
  }

  // Assert text content
  async assertTextEquals(selector, expectedText, message) {
    const actualText = await this.browser.getText(selector);
    if (actualText.trim() !== expectedText.trim()) {
      throw new Error(
        message || 
        `Text mismatch for ${selector}\nExpected: "${expectedText}"\nActual: "${actualText}"`
      );
    }
  }

  // Assert text contains
  async assertTextContains(selector, substring, message) {
    const text = await this.browser.getText(selector);
    if (!text.includes(substring)) {
      throw new Error(
        message || 
        `Text does not contain substring in ${selector}\nText: "${text}"\nExpected substring: "${substring}"`
      );
    }
  }

  // Assert attribute value
  async assertAttributeEquals(selector, attribute, expectedValue, message) {
    const actualValue = await this.browser.getAttribute(selector, attribute);
    if (actualValue !== expectedValue) {
      throw new Error(
        message || 
        `Attribute mismatch for ${selector}[${attribute}]\nExpected: "${expectedValue}"\nActual: "${actualValue}"`
      );
    }
  }

  // Assert element count
  async assertCount(selector, expectedCount, message) {
    const actualCount = await this.browser.count(selector);
    if (actualCount !== expectedCount) {
      throw new Error(
        message || 
        `Element count mismatch for ${selector}\nExpected: ${expectedCount}\nActual: ${actualCount}`
      );
    }
  }

  // Assert URL
  async assertUrlEquals(expectedUrl, message) {
    const actualUrl = await this.browser.getUrl();
    if (actualUrl !== expectedUrl) {
      throw new Error(
        message || 
        `URL mismatch\nExpected: "${expectedUrl}"\nActual: "${actualUrl}"`
      );
    }
  }

  // Assert URL contains
  async assertUrlContains(substring, message) {
    const url = await this.browser.getUrl();
    if (!url.includes(substring)) {
      throw new Error(
        message || 
        `URL does not contain substring\nURL: "${url}"\nExpected substring: "${substring}"`
      );
    }
  }

  // Assert class exists
  async assertHasClass(selector, className, message) {
    const classAttr = await this.browser.getAttribute(selector, 'class');
    const classes = classAttr ? classAttr.split(' ') : [];
    if (!classes.includes(className)) {
      throw new Error(
        message || 
        `Element ${selector} does not have class "${className}"\nClasses: ${classAttr || '(none)'}`
      );
    }
  }

  // Assert element is visible
  async assertVisible(selector, message) {
    await this.assertExists(selector, message);
    const isVisible = await this.browser.evaluate((sel) => {
      const element = document.querySelector(sel);
      if (!element) return false;
      const style = window.getComputedStyle(element);
      return style.display !== 'none' && 
             style.visibility !== 'hidden' && 
             style.opacity !== '0';
    }, selector);
    
    if (!isVisible) {
      throw new Error(message || `Element is not visible: ${selector}`);
    }
  }

  // Assert element is disabled
  async assertDisabled(selector, message) {
    const isDisabled = await this.browser.evaluate((sel) => {
      const element = document.querySelector(sel);
      return element && element.disabled;
    }, selector);
    
    if (!isDisabled) {
      throw new Error(message || `Element is not disabled: ${selector}`);
    }
  }

  // Assert JavaScript value
  async assertJsEquals(expression, expectedValue, message) {
    const actualValue = await this.browser.evaluate((expr) => {
      return eval(expr);
    }, expression);
    
    if (JSON.stringify(actualValue) !== JSON.stringify(expectedValue)) {
      throw new Error(
        message || 
        `JavaScript value mismatch\nExpression: ${expression}\nExpected: ${JSON.stringify(expectedValue)}\nActual: ${JSON.stringify(actualValue)}`
      );
    }
  }
}

module.exports = Assertions;