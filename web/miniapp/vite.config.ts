import crypto from 'crypto'
import fs from 'fs'
import path from 'path'
import { defineConfig, type Plugin } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// Telegram's webview finishes small files (the page, the stylesheet) and then
// stalls on a large body. A 441KB inline script never reaches its closing tag,
// so the marker after it stays "no-end". Ship the bundle as small slices and
// assemble them with an inline loader.
const PART_CHARS = 12_000

function splitParts(code: string): string[] {
  const parts: string[] = []
  let i = 0
  while (i < code.length) {
    let end = Math.min(i + PART_CHARS, code.length)
    if (end < code.length) {
      const nl = code.lastIndexOf('\n', end)
      if (nl > i + PART_CHARS / 2) end = nl + 1
    }
    parts.push(code.slice(i, end))
    i = end
  }
  return parts
}

function partLoader(urls: string[]): string {
  const list = JSON.stringify(urls)
  return `<script>
(function () {
  var urls = ${list}
  window.__LIFEOS_LOAD = '0/' + urls.length
  var got = new Array(urls.length)
  var next = 0
  var inflight = 0
  var done = 0
  var failed = false
  function finish() {
    window.__LIFEOS_LOAD = 'run'
    var s = document.createElement('script')
    s.text = got.join('')
    document.body.appendChild(s)
  }
  function fail(i, msg) {
    if (failed) return
    failed = true
    window.__LIFEOS_LOAD = 'fail ' + i
    lifeosBoot('Не удалось загрузить часть ' + (i + 1) + ': ' + msg)
  }
  function kick() {
    if (failed) return
    while (inflight < 4 && next < urls.length) {
      (function (i) {
        inflight++
        var x = new XMLHttpRequest()
        x.open('GET', urls[i], true)
        x.overrideMimeType('text/plain; charset=utf-8')
        x.onload = function () {
          inflight--
          if (failed) return
          if (x.status < 200 || x.status >= 300) { fail(i, 'http ' + x.status); return }
          var type = x.getResponseHeader('Content-Type') || ''
          if (type.indexOf('javascript') < 0 && type.indexOf('ecmascript') < 0 && type.indexOf('text/plain') < 0) {
            fail(i, type || 'не js')
            return
          }
          got[i] = x.responseText
          done++
          window.__LIFEOS_LOAD = done + '/' + urls.length
          lifeosBoot('Загрузка LifeOS… ' + done + '/' + urls.length)
          if (done === urls.length) finish()
          else kick()
        }
        x.onerror = function () { inflight--; fail(i, 'сеть') }
        x.ontimeout = function () { inflight--; fail(i, 'таймаут') }
        x.timeout = 20000
        x.send()
      })(next)
      next++
    }
  }
  kick()
})()
</script>`
}

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
      const assetsDir = path.join(dist, 'assets')
      let html = fs.readFileSync(htmlPath, 'utf8')
      const tag = html.match(/<script src="(\/app\/assets\/[^"]+\.js)"><\/script>/)
      if (!tag) throw new Error('telegram-classic-bundle: classic script tag missing')
      const jsPath = path.join(dist, tag[1].replace(/^\/app\//, ''))
      let code = fs.readFileSync(jsPath, 'utf8')
      code = code.replace(/\n\/\/# sourceMappingURL=\S+\s*$/, '')
      const parts = splitParts(code)
      if (parts.join('') !== code) throw new Error('telegram-classic-bundle: part join mismatch')
      const hash = crypto.createHash('sha256').update(code).digest('hex').slice(0, 10)
      for (const name of fs.readdirSync(assetsDir)) {
        if (/^lifeos-[0-9a-f]+-\d+\.js$/.test(name)) fs.unlinkSync(path.join(assetsDir, name))
      }
      const urls: string[] = []
      parts.forEach((part, i) => {
        const name = `lifeos-${hash}-${i}.js`
        fs.writeFileSync(path.join(assetsDir, name), part)
        urls.push(`/app/assets/${name}`)
      })
      html = html.replace(tag[0], () => partLoader(urls))
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
