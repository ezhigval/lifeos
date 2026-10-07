import fs from 'fs'
import path from 'path'
import { defineConfig, type Plugin } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// One classic script. Thirty-seven XHR slices added a round trip each, and any
// stalled slice aborted the boot. The gzipped bundle is about 126KB.

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
      const tag = html.match(/<script type="module"(?: crossorigin)? src="([^"]+)"><\/script>/)
      let out = html.replace(/<link rel="stylesheet" crossorigin href=/g, '<link rel="stylesheet" href=')
      if (!tag) return out
      out = out.replace(tag[0], '')
      const href = tag[1]
      const preload = `<link rel="preload" href="${href}" as="script">`
      if (out.includes('</head>')) out = out.replace('</head>', `${preload}\n  </head>`)
      const classic = `<script src="${href}"></script>`
      return out.includes('</body>') ? out.replace('</body>', `${classic}\n  </body>`) : out + classic
    },
    closeBundle() {
      const assetsDir = path.resolve(__dirname, 'dist/assets')
      for (const name of fs.readdirSync(assetsDir)) {
        const filePath = path.join(assetsDir, name)
        if (name.endsWith('.map')) {
          fs.unlinkSync(filePath)
          continue
        }
        if (!name.endsWith('.js')) continue
        const code = fs.readFileSync(filePath, 'utf8')
        const stripped = code.replace(/\n\/\/# sourceMappingURL=\S+\s*$/, '')
        if (stripped !== code) fs.writeFileSync(filePath, stripped)
      }
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
