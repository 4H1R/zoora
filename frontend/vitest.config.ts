import path from "path"
import react from "@vitejs/plugin-react"
import { defineConfig } from "vitest/config"

// Minimal Vitest config — deliberately does NOT include the TanStack Router
// plugin or vite-plugin-pwa from vite.config.js. Those plugins expect a full
// dev/build pipeline (route tree generation, service worker manifest) and
// break under Vitest's unit-test transform. The `@` alias mirrors
// vite.config.js's resolve.alias so imports resolve the same way in tests.
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: ["./src/test/setup.ts"],
    // Scope to unit tests under src/ only — tests/e2e/**/*.spec.ts belong to
    // Playwright (see playwright.config.ts) and importing @playwright/test's
    // test() alongside Vitest's global test() throws at collection time.
    include: ["src/**/*.{test,spec}.{ts,tsx}"],
  },
})
