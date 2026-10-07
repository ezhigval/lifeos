import path from 'path'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// Desktop window: real ES modules, not the Telegram classic-script rewrite.
// The UI is local. API calls stay same-origin via the dev proxy so the
// browser does not need CORS. Point LIFEOS_API_ORIGIN at the VM when the
// API is not on this machine.
const apiOrigin = process.env.LIFEOS_API_ORIGIN || 'http://127.0.0.1:8080'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  root: path.resolve(__dirname, './desktop'),
  base: './',
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 5174,
    proxy: {
      '/api': {
        target: apiOrigin,
        changeOrigin: true,
      },
    },
  },
  preview: {
    port: 4174,
    proxy: {
      '/api': {
        target: apiOrigin,
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: path.resolve(__dirname, './desktop/dist'),
    emptyOutDir: true,
    sourcemap: false,
  },
})
