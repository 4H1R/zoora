import path from 'path'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { tanstackRouter } from '@tanstack/router-plugin/vite'
import tailwindcss from '@tailwindcss/vite'

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
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    host: true,
    port: 3000,
    allowedHosts: ['.localhost'],
    proxy: {
      // changeOrigin: false preserves Host: <slug>.localhost so the Go tenant
      // middleware resolves the right org from the subdomain.
      '/api': { target: 'http://localhost:8080', changeOrigin: false },
      // LiveKit signal WS. Browser->Docker-Desktop port 7880 forwards plain HTTP
      // but drops the WebSocket upgrade; route it through Vite (proven WS path)
      // so the upgrade actually reaches the container. ws:true enables upgrade.
      '/rtc': { target: 'http://localhost:7880', changeOrigin: true, ws: true },
    },
  },
})
