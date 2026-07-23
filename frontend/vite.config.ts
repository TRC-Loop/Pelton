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
})
