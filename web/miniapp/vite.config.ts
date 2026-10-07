import fs from 'fs'
import path from 'path'
import { defineConfig, type Plugin } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

function escapeInlineScript(code: string): string {
  return code
    .replace(/\u2028/g, '\\u2028')
    .replace(/\u2029/g, '\\u2029')
    .replace(/<!--/g, '<\\!--')
    .replace(/<\/script/gi, '<\\/script')
}

// Telegram Desktop's webview runs inline scripts and does not request an
// external bundle: the watchdog reports "js не запрошен" while CSS loads.
// Inline one classic script so the parser executes it with the document.
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
      const classic = `<script src="${tag[1]}"></script>`
      return out.includes('</body>') ? out.replace('</body>', `${classic}\n  </body>`) : out + classic
    },
    closeBundle() {
      const dist = path.resolve(__dirname, 'dist')
      const htmlPath = path.join(dist, 'index.html')
      let html = fs.readFileSync(htmlPath, 'utf8')
      const tag = html.match(/<script src="(\/app\/assets\/[^"]+\.js)"><\/script>/)
      if (!tag) throw new Error('telegram-classic-bundle: classic script tag missing')
      const jsPath = path.join(dist, tag[1].replace(/^\/app\//, ''))
      let code = fs.readFileSync(jsPath, 'utf8')
      code = code.replace(/\n\/\/# sourceMappingURL=\S+\s*$/, '')
      code = escapeInlineScript(code)
      const inline =
        `<script>\n${code}\n</script>\n  <script>window.__LIFEOS_JS_END=1</script>`
      // Replacement strings treat $&, $` and $' as match fragments. The bundle
      // uses those as ordinary characters, so insert the script as a function.
      html = html.replace(tag[0], () => inline)
      fs.writeFileSync(htmlPath, html)
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
