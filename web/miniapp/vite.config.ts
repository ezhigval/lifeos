import { defineConfig, type Plugin } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import path from 'path'

// Telegram Desktop's webview runs inline scripts and then ignores type=module.
// The boot line stays until the 12s watchdog. Ship one classic script instead.
function telegramClassicBundle(): Plugin {
  return {
    name: 'telegram-classic-bundle',
    apply: 'build',
    enforce: 'post',
    generateBundle(_options, bundle) {
      for (const file of Object.values(bundle)) {
        if (file.type !== 'chunk' || !file.fileName.endsWith('.js')) continue
        let code = file.code
        code = code.replace(/import\.meta\.url/g, '""')
        code = code.replace(/import\.meta/g, '({})')
        if (code.includes('import(')) {
          code =
            'function __lifeosImport(){return Promise.reject(new Error("dynamic import"))}\n' +
            code.replace(/import\(/g, '__lifeosImport(')
        }
        file.code = 'window.__LIFEOS_JS=1;\n' + code
      }
    },
    transformIndexHtml(html) {
      return html
        .replace(
          /<script type="module"(?: crossorigin)? src="([^"]+)"><\/script>/g,
          '<script defer src="$1"></script>',
        )
        .replace(/<link rel="stylesheet" crossorigin href=/g, '<link rel="stylesheet" href=')
    },
  }
}

export default defineConfig({
  plugins: [react(), tailwindcss(), telegramClassicBundle()],
  base: '/app/',
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
    target: 'es2020',
    modulePreload: false,
  },
})
