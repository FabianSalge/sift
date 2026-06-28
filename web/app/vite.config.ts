import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

// Relative base so the built site works served from any path (root, subpath, or
// the distroless static server in Phase 4).
export default defineConfig({
  base: './',
  plugins: [svelte()],
})
