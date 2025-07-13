// Puppeteer configuration for E2E tests
module.exports = {
  // Launch options
  launch: {
    headless: process.env.PUPPETEER_HEADLESS !== 'false',
    slowMo: process.env.PUPPETEER_SLOWMO ? parseInt(process.env.PUPPETEER_SLOWMO) : 0,
    devtools: process.env.PUPPETEER_DEVTOOLS === 'true',
    args: [
      '--no-sandbox',
      '--disable-setuid-sandbox',
      '--disable-dev-shm-usage',
      '--disable-accelerated-2d-canvas',
      '--disable-gpu'
    ]
  },

  // Viewport settings
  viewport: {
    width: 1280,
    height: 720
  },

  // Navigation options
  navigationOptions: {
    waitUntil: 'domcontentloaded',
    timeout: 30000
  },

  // Screenshot options
  screenshotOptions: {
    type: 'png',
    fullPage: true
  },

  // Test timeouts
  timeouts: {
    test: 60000,
    navigation: 30000,
    element: 10000
  },

  // Server configuration
  server: {
    port: process.env.TEST_PORT || 8080,
    host: process.env.TEST_HOST || 'localhost',
    protocol: process.env.TEST_PROTOCOL || 'http'
  },

  // Get base URL
  getBaseUrl() {
    const { protocol, host, port } = this.server;
    return `${protocol}://${host}:${port}`;
  }
};