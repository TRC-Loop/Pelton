import {defineConfig} from 'vite'
import {svelte} from '@sveltejs/vite-plugin-svelte'
import {fileURLToPath} from 'node:url'

// @tabler/icons ships its full icon geometry as tabler-nodes-outline.json, but
// the package's exports map only exposes ./icons/*, so the file is aliased here
// to a stable specifier. The icon picker lazy-imports it, keeping the ~2MB blob
// in its own chunk and out of the main bundle.
const tablerNodesOutline = fileURLToPath(
  new URL('./node_modules/@tabler/icons/tabler-nodes-outline.json', import.meta.url),
)

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [svelte()],
  resolve: {
    alias: {
      'tabler-nodes-outline': tablerNodesOutline,
    },
  },
  build: {
    // the flag set is globbed whole so a new locale needs no asset work, and
    // most of its svgs are small enough that vite would inline every one of
    // them into the picker's chunk. They stay files, so only the handful
    // actually rendered is ever read.
    assetsInlineLimit: (file) => (file.includes('flag-icons/flags/') ? false : undefined),
  },
})
