import { fileURLToPath, URL } from 'node:url'

import tailwindcss from '@tailwindcss/vite'
import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vitest/config'

// cmd/frontend embeds dist/ with a go:embed directive that resolves next to
// the Go source file, which is why the build writes there directly rather
// than into web/dist and leaving a copy step for the Taskfile to do. The
// prerender step writes the pages into the same directory afterwards.
export default defineConfig({
  plugins: [vue(), tailwindcss()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  build: {
    outDir: '../cmd/frontend/dist',
    emptyOutDir: true,
    // The prerender script reads this to find the content hashed entry
    // script and its stylesheet, so it can write them into the HTML it
    // generates without hardcoding a filename vite is free to change.
    manifest: true,
    // Every route is prerendered and served as static HTML, and that
    // HTML has to carry a stylesheet link for everything the page needs
    // at first paint: a split view stylesheet arrives with the route
    // chunk, which is after first paint, and never at all for a visitor
    // with no JavaScript. The prerenderer therefore links every
    // stylesheet the build emits; splitting it would buy a second page
    // nothing and cost it a second request.
    cssCodeSplit: false,
  },
  test: {
    environment: 'jsdom',
    setupFiles: ['./vitest.setup.ts'],
    css: false,
  },
})
