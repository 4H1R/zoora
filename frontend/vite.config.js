import path from 'path'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { tanstackRouter } from '@tanstack/router-plugin/vite'
import tailwindcss from '@tailwindcss/vite'
import { VitePWA } from 'vite-plugin-pwa'

// https://vitejs.dev/config/
export default defineConfig({
  optimizeDeps: {
    // pdfjs-dist ships ESM workers; tell Vite not to bundle them so the
    // new URL(..., import.meta.url) worker pattern resolves at runtime.
    exclude: ['pdfjs-dist'],
  },
  plugins: [
    tailwindcss(),
    tanstackRouter({ target: 'react', autoCodeSplitting: true }),
    react({
      babel: {
        plugins: ['babel-plugin-react-compiler'],
      },
    }),
    VitePWA({
      // 'prompt' — we surface an in-app toast (see src/pwa.ts) instead of
      // silently reloading, so a live meeting/recording is never interrupted.
      registerType: 'prompt',
      // Registration is wired manually via virtual:pwa-register in src/pwa.ts.
      injectRegister: null,
      // Cache the app shell + hashed static assets. Fonts (woff2) included so
      // Geist/Vazirmatn render offline.
      workbox: {
        globPatterns: ['**/*.{js,css,html,ico,png,svg,webp,woff2}'],
        // The live-room bundle (livekit + tldraw + pdfjs) is ~2.7 MB and lazily
        // loaded only when entering a room. Keep it out of the install-time
        // precache — the runtime rule below caches it on first use instead.
        globIgnores: ['**/_liveId-*.js'],
        // SPA fallback: unmatched navigations serve index.html so deep links
        // work offline. Never intercept the API or LiveKit signalling paths.
        navigateFallback: '/index.html',
        // Also exclude the dedicated FCM background worker so it is served
        // as-is on its own scope instead of falling back to index.html.
        navigateFallbackDenylist: [/^\/api/, /^\/rtc/, /^\/firebase-messaging-sw\.js$/],
        // Never cache auth'd API responses. Hashed JS/CSS chunks are
        // immutable, so cache them stale-while-revalidate as they're requested
        // — covers the excluded live-room chunk and any future big splits.
        runtimeCaching: [
          {
            urlPattern: ({ request }) =>
              request.destination === "script" ||
              request.destination === "style" ||
              request.destination === "worker",
            handler: "StaleWhileRevalidate",
            options: { cacheName: "app-assets" },
          },
        ],
        cleanupOutdatedCaches: true,
        clientsClaim: true,
      },
      includeAssets: ['favicon.svg', 'favicon-32.png', 'apple-touch-icon.png'],
      manifest: {
        name: 'Zoora — Virtual Classrooms & Video Meetings',
        short_name: 'Zoora',
        description:
          'Run live online classes, video meetings, and recordings on one secure multi-tenant platform.',
        theme_color: '#16a34a',
        background_color: '#ffffff',
        display: 'standalone',
        orientation: 'any',
        start_url: '/',
        scope: '/',
        categories: ['education', 'productivity', 'business'],
        icons: [
          { src: '/pwa-192.png', sizes: '192x192', type: 'image/png', purpose: 'any' },
          { src: '/pwa-512.png', sizes: '512x512', type: 'image/png', purpose: 'any' },
          { src: '/maskable-512.png', sizes: '512x512', type: 'image/png', purpose: 'maskable' },
        ],
      },
      // Keep the SW out of dev so it never shadows Vite HMR or the /api and
      // /rtc proxies.
      devOptions: {
        enabled: false,
      },
    }),
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    host: true,
    port: 3000,
    allowedHosts: ['.localhost', '.zoora.local'],
    proxy: {
      // changeOrigin: false preserves Host: <slug>.localhost so the Go tenant
      // middleware resolves the right org from the subdomain.
      // ws:true so the conversations hub upgrade (/api/v1/ws) is proxied too —
      // the client uses a host-relative WS URL now, not a hardcoded :8080.
      '/api': { target: 'http://localhost:8080', changeOrigin: false, ws: true },
      // LiveKit signal WS. Browser->Docker-Desktop port 7880 forwards plain HTTP
      // but drops the WebSocket upgrade; route it through Vite (proven WS path)
      // so the upgrade actually reaches the container. ws:true enables upgrade.
      '/rtc': { target: 'http://localhost:7880', changeOrigin: true, ws: true },
    },
  },
  preview: {
    host: true,
    port: 3000,
    allowedHosts: ['.localhost', '.zoora.local'],
    proxy: {
      // ws:true so the conversations hub upgrade (/api/v1/ws) is proxied too —
      // the client uses a host-relative WS URL now, not a hardcoded :8080.
      '/api': { target: 'http://localhost:8080', changeOrigin: false, ws: true },
      '/rtc': { target: 'http://localhost:7880', changeOrigin: true, ws: true },
    },
  },
})
