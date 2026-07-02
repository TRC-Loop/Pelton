import { vitePreprocess } from '@sveltejs/vite-plugin-svelte'

// vitePreprocess (oxc/esbuild-based) replaces svelte-preprocess here:
// svelte-preprocess's ts-compiler-based transform breaks on every file under
// TypeScript 6, and vitePreprocess is the path the Svelte/Vite tooling itself
// recommends going forward.
export default {
  preprocess: vitePreprocess()
}
