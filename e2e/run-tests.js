#!/usr/bin/env node

const { spawn } = require('child_process');
const path = require('path');
const fs = require('fs').promises;
const BrowserHelper = require('./helpers/browser-helper');
const Assertions = require('./helpers/assertions');
const ComponentHelper = require('./helpers/component-helper');
const testConfig = require('./config/test.config');
const puppeteerConfig = require('./config/puppeteer.config');

// ANSI color codes
const colors = {
  reset: '\x1b[0m',
  bright: '\x1b[1m',
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  cyan: '\x1b[36m'
};

// Test runner class
class TestRunner {
  constructor() {
    this.server = null;
    this.browser = null;
    this.assertions = null;
    this.componentHelper = null;
    this.results = {
      passed: 0,
      failed: 0,
      skipped: 0,
      errors: []
    };
  }

  // Start the server
  async startServer() {
    console.log(`${colors.cyan}Starting Redi server...${colors.reset}`);
    
    return new Promise((resolve, reject) => {
      const serverPath = path.join(__dirname, '../redi');
      const args = [
        '--root=../fixtures',
        `--port=${puppeteerConfig.server.port}`,
        '--log=e2e-server.log'
      ];
      
      this.server = spawn(serverPath, args, {
        detached: true,
        stdio: ['ignore', 'pipe', 'pipe']
      });
      
      this.server.on('error', reject);
      
      // Log server output for debugging
      this.server.stdout.on('data', (data) => {
        console.log(`Server stdout: ${data}`);
      });
      
      this.server.stderr.on('data', (data) => {
        console.error(`Server stderr: ${data}`);
      });
      
      // Wait for server to start - increased timeout
      setTimeout(() => {
        console.log(`${colors.green}Server started on port ${puppeteerConfig.server.port}${colors.reset}`);
        resolve();
      }, 5000);
    });
  }

  // Stop the server
  async stopServer() {
    if (this.server) {
      console.log(`${colors.cyan}Stopping server...${colors.reset}`);
      try {
        process.kill(-this.server.pid);
      } catch (error) {
        // Server may have already stopped
        if (error.code !== 'ESRCH') {
          console.error('Error stopping server:', error);
        }
      }
      this.server = null;
    }
  }

  // Initialize browser and helpers
  async initBrowser() {
    this.browser = new BrowserHelper();
    await this.browser.launch();
    this.assertions = new Assertions(this.browser);
  }

  // Run a single test
  async runTest(testFile, framework) {
    const testName = path.basename(testFile, '.test.js');
    console.log(`\n${colors.bright}Running test: ${framework}/${testName}${colors.reset}`);
    
    try {
      // Clear require cache for fresh test load
      delete require.cache[require.resolve(testFile)];
      const test = require(testFile);
      
      // Create component helper for this framework
      this.componentHelper = new ComponentHelper(this.browser, framework);
      
      // Run test
      const startTime = Date.now();
      await test({
        browser: this.browser,
        assert: this.assertions,
        componentHelper: this.componentHelper,
        config: testConfig,
        baseUrl: puppeteerConfig.getBaseUrl()
      });
      
      const duration = Date.now() - startTime;
      console.log(`${colors.green}✓ ${testName} passed (${duration}ms)${colors.reset}`);
      this.results.passed++;
      
    } catch (error) {
      console.log(`${colors.red}✗ ${testName} failed${colors.reset}`);
      console.error(`  ${error.message}`);
      this.results.failed++;
      this.results.errors.push({
        test: `${framework}/${testName}`,
        error: error.message,
        stack: error.stack
      });
      
      // Take screenshot on failure
      if (testConfig.screenshots.onFailure) {
        const screenshotName = `${framework}-${testName}-failure`;
        await this.browser.screenshot(screenshotName);
      }
    }
  }

  // Run all tests for a framework
  async runFrameworkTests(framework) {
    console.log(`\n${colors.blue}${colors.bright}Testing ${framework.toUpperCase()} components${colors.reset}`);
    
    const testDir = path.join(testConfig.testDir, framework);
    
    try {
      const files = await fs.readdir(testDir);
      const testFiles = files
        .filter(file => file.endsWith('.test.js'))
        .map(file => path.join(testDir, file));
      
      for (const testFile of testFiles) {
        await this.runTest(testFile, framework);
      }
    } catch (error) {
      console.log(`${colors.yellow}No tests found for ${framework}${colors.reset}`);
    }
  }
  
  // Run framework tests standalone (with server lifecycle)
  async runFrameworkTestsStandalone(framework) {
    console.log(`${colors.bright}${colors.cyan}Running ${framework.toUpperCase()} tests${colors.reset}`);
    console.log(`Base URL: ${puppeteerConfig.getBaseUrl()}`);
    
    const startTime = Date.now();
    
    try {
      // Start server
      await this.startServer();
      
      // Initialize browser
      await this.initBrowser();
      
      // Run the framework tests
      await this.runFrameworkTests(framework);
      
    } catch (error) {
      console.error(`${colors.red}Test runner error:${colors.reset}`, error);
      this.results.errors.push({
        test: 'Test Runner',
        error: error.message,
        stack: error.stack
      });
    } finally {
      // Clean up
      if (this.browser) {
        await this.browser.close();
      }
      await this.stopServer();
    }
    
    const duration = Math.round((Date.now() - startTime) / 1000);
    this.printSummary(duration);
  }

  // Run a single test
  async runSingleTest(framework, testName) {
    console.log(`${colors.bright}${colors.cyan}Running single test: ${framework}/${testName}${colors.reset}`);
    console.log(`Base URL: ${puppeteerConfig.getBaseUrl()}`);
    
    const startTime = Date.now();
    
    try {
      // Start server
      await this.startServer();
      
      // Initialize browser
      await this.initBrowser();
      
      // Run the specific test
      const testFile = path.join(testConfig.testDir, framework, `${testName}.test.js`);
      await this.runTest(testFile, framework);
      
    } catch (error) {
      console.error(`${colors.red}Test runner error:${colors.reset}`, error);
      this.results.errors.push({
        test: 'Test Runner',
        error: error.message,
        stack: error.stack
      });
    } finally {
      // Clean up
      if (this.browser) {
        await this.browser.close();
      }
      await this.stopServer();
    }
    
    const duration = Math.round((Date.now() - startTime) / 1000);
    this.printSummary(duration);
  }

  // Print test summary
  printSummary(duration) {
    // Print summary
    console.log(`\n${colors.bright}Test Summary${colors.reset}`);
    console.log(`${colors.green}Passed: ${this.results.passed}${colors.reset}`);
    console.log(`${colors.red}Failed: ${this.results.failed}${colors.reset}`);
    console.log(`${colors.yellow}Skipped: ${this.results.skipped}${colors.reset}`);
    console.log(`Total time: ${duration}s`);
    
    // Print errors if any
    if (this.results.errors.length > 0) {
      console.log(`\n${colors.red}${colors.bright}Errors:${colors.reset}`);
      this.results.errors.forEach((err, index) => {
        console.log(`\n${index + 1}. ${err.test}`);
        console.log(`   ${err.error}`);
        if (process.env.VERBOSE) {
          console.log(`   ${err.stack}`);
        }
      });
    }
    
    // Exit with appropriate code
    process.exit(this.results.failed > 0 ? 1 : 0);
  }

  // Run all tests
  async runAllTests() {
    console.log(`${colors.bright}${colors.cyan}Starting E2E Test Suite${colors.reset}`);
    console.log(`Base URL: ${puppeteerConfig.getBaseUrl()}`);
    
    const startTime = Date.now();
    
    try {
      // Start server
      await this.startServer();
      
      // Initialize browser
      await this.initBrowser();
      
      // Run tests for each framework
      for (const framework of testConfig.frameworks) {
        await this.runFrameworkTests(framework);
      }
      
    } catch (error) {
      console.error(`${colors.red}Test runner error:${colors.reset}`, error);
      this.results.errors.push({
        test: 'Test Runner',
        error: error.message,
        stack: error.stack
      });
    } finally {
      // Clean up
      if (this.browser) {
        await this.browser.close();
      }
      await this.stopServer();
    }
    
    const duration = Math.round((Date.now() - startTime) / 1000);
    this.printSummary(duration);
  }
}

// Main execution
async function main() {
  const runner = new TestRunner();
  
  // Handle process termination
  process.on('SIGINT', async () => {
    console.log('\nTest interrupted, cleaning up...');
    await runner.stopServer();
    process.exit(1);
  });
  
  // Parse command line arguments
  const args = process.argv.slice(2);
  const watch = args.includes('--watch');
  const frameworkArg = args.find(arg => arg.startsWith('--framework='));
  const testArg = args.find(arg => arg.startsWith('--test='));
  
  if (watch) {
    console.log('Watch mode not yet implemented');
    process.exit(0);
  }
  
  // Run specific test if requested
  if (frameworkArg && testArg) {
    const framework = frameworkArg.split('=')[1];
    const testName = testArg.split('=')[1];
    await runner.runSingleTest(framework, testName);
  } else if (frameworkArg) {
    const framework = frameworkArg.split('=')[1];
    await runner.runFrameworkTestsStandalone(framework);
  } else {
    // Run all tests
    await runner.runAllTests();
  }
}

// Run if called directly
if (require.main === module) {
  main().catch(console.error);
}

module.exports = { TestRunner };