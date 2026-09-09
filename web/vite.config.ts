import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';

// Matches the default Go backend address.
const backend = 'http://127.0.0.1:8080';

export default defineConfig({
  plugins: [svelte()],
  server: {
    // Preserve Host (changeOrigin defaults to false) so it matches the browser's
    // Origin during WebSocket validation. Vite serves the frontend itself.
    proxy: {
      '/api': { target: backend },
      '/ws': { target: backend, ws: true },
      // The optional legal notices are read by the Go process from the operator's
      // files, so development reaches them the same way production does — with the
      // same PLANNINGPOKER_LEGAL_DIR and no frontend rebuild.
      '/legal': { target: backend },
    },
  },
  build: {
    // Write inside the Go package because go:embed cannot include parent directories.
    outDir: '../internal/webassets/dist',
    emptyOutDir: true,
  },
});
