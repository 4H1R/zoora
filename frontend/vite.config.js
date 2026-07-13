import path from 'path'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { tanstackStart } from '@tanstack/react-start/plugin/vite'
import { nitro } from 'nitro/vite'
import tailwindcss from '@tailwindcss/vite'

const isBuild = process.argv.includes('build')

// https://vitejs.dev/config/
export default defineConfig({
  optimizeDeps: {
    // pdfjs-dist ships ESM workers; tell Vite not to bundle them so the
    // new URL(..., import.meta.url) worker pattern resolves at runtime.
    exclude: ['pdfjs-dist'],
  },
  plugins: [
    tailwindcss(),
    isBuild && nitro({ preset: 'bun' }),
    tanstackStart(),
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
    allowedHosts: ['.localhost', '.zoora.local'],
    fs: {
      allow: [path.resolve(__dirname, '..')],
    },
    proxy: {
      '/api': { target: 'http://localhost:8080', changeOrigin: false, ws: true },
      '/rtc': { target: 'http://localhost:7880', changeOrigin: true, ws: true },
    },
  },
  preview: {
    host: true,
    port: 3000,
    allowedHosts: ['.localhost', '.zoora.local'],
    proxy: {
      '/api': { target: 'http://localhost:8080', changeOrigin: false, ws: true },
      '/rtc': { target: 'http://localhost:7880', changeOrigin: true, ws: true },
    },
  },
})
