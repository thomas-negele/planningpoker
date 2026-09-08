import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

export default {
  // vitePreprocess lets <script lang="ts"> blocks inside components be type-checked
  // and compiled. Without it, TypeScript in a component is a syntax error.
  preprocess: vitePreprocess(),
};
