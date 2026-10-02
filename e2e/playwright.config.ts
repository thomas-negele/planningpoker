import { defineConfig } from '@playwright/test';
import { origin, pooOrigin } from './servers';

const server = (url: string, env: Record<string, string> = {}) => ({
  command: 'go run -tags embedassets ./cmd/planningpoker',
  cwd: '..',
  env: { PLANNINGPOKER_LISTEN_ADDR: new URL(url).host, ...env },
  url,
  reuseExistingServer: false,
  timeout: 60_000,
});

export default defineConfig({
  testDir: './tests',
  use: { baseURL: origin, browserName: 'chromium' },
  webServer: [server(origin), server(pooOrigin, { PLANNINGPOKER_POO_THROWS: 'true' })],
});
