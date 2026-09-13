import { defineConfig } from '@playwright/test';

const origin = 'http://127.0.0.1:4324';

export default defineConfig({
  testDir: './tests',
  use: { baseURL: origin, browserName: 'chromium' },
  webServer: {
    command: 'go run -tags embedassets ./cmd/planningpoker',
    cwd: '..',
    env: { PLANNINGPOKER_LISTEN_ADDR: '127.0.0.1:4324' },
    url: origin,
    reuseExistingServer: false,
    timeout: 60_000,
  },
});
