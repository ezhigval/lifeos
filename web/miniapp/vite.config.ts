import fs from 'fs'
import path from 'path'
import { defineConfig, type Plugin } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { lifeosStaticCache } from './vite.assets.ts'

// Telegram's webview runs classic scripts, not ES modules. The entry stays one
// script. Other screens are separate classic scripts and load on navigation.

const LOADER = `window.__lifeosMods=window.__lifeosMods||{};function __lifeosImport(spec){if(typeof spec!="string")return Promise.reject(new Error("dynamic"));var file=spec.split("/").pop();var hit=window.__lifeosMods[file];if(hit)return Promise.resolve(hit);return new Promise(function(resolve,reject){var s=document.createElement("script");s.src="/app/assets/"+file;s.onload=function(){var mod=window.__lifeosMods[file];mod?resolve(mod):reject(new Error("empty"))};s.onerror=function(){reject(new Error("chunk"))};document.head.appendChild(s)})}
`

function rewriteStaticImports(code: string): string {
  return code.replace(/import\{([^}]+)\}from"(\.\/[^"]+)"/g, (_match, clause: string, spec: string) => {
    const file = spec.slice(2)
    const bindings = clause.split(',').map((part) => {
      const item = part.trim()
      const named = item.match(/^([A-Za-z0-9_$]+) as ([A-Za-z0-9_$]+)$/)
      if (named) return `${named[2]}=window.__lifeosMods["${file}"].${named[1]}`
      if (!/^[A-Za-z0-9_$]+$/.test(item)) throw new Error(`telegram-classic-bundle: bad import "${item}"`)
      return `${item}=window.__lifeosMods["${file}"].${item}`
    })
    return `var ${bindings.join(',')}`
  })
}

function rewriteExports(code: string, fileName: string): string {
  return code.replace(/export\{([^}]+)\}/g, (_match, clause: string) => {
    const fields = clause.split(',').map((part) => {
      const item = part.trim()
      const named = item.match(/^([A-Za-z0-9_$]+) as ([A-Za-z0-9_$]+)$/)
      if (named) return `${named[2]}:${named[1]}`
      if (!/^[A-Za-z0-9_$]+$/.test(item)) throw new Error(`telegram-classic-bundle: bad export "${item}"`)
      return `${item}:${item}`
    })
    return `window.__lifeosMods["${fileName}"]={${fields.join(',')}}`
  })
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
        code = rewriteStaticImports(code)
        code = rewriteExports(code, path.posix.basename(file.fileName))
        code = code.replace(/import\(/g, '__lifeosImport(')
        if (/[(;\n]import[\s{*]/.test(code) || code.includes('export{') || code.includes('export ')) {
          throw new Error(`telegram-classic-bundle: module syntax left in ${file.fileName}`)
        }
        // Classic scripts share one global scope. A section chunk's `var i`
        // overwrites the entry helper (Object.getPrototypeOf) and the screen
        // dies with "empty" or "useContext is not a function".
        code = `(function(){\n${code}\n})();\n`
        if (file.isEntry) code = `window.__LIFEOS_JS=1;\n${LOADER}${code}`
        file.code = code
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
      // Last close tag: an earlier </body> inside a comment must not swallow the script.
      const bodyClose = out.lastIndexOf('</body>')
      if (bodyClose < 0) return out + classic
      return `${out.slice(0, bodyClose)}${classic}\n  ${out.slice(bodyClose)}`
    },
    closeBundle() {
      const assetsDir = path.resolve(__dirname, 'dist/assets')
      if (!fs.existsSync(assetsDir)) return
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
  plugins: [react(), tailwindcss(), telegramClassicBundle(), lifeosStaticCache()],
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
    rollupOptions: {
      output: {
        manualChunks(id: string) {
          // Pages other than Home stay async. Everything they share stays in the
          // entry chunk so the first script has no static imports.
          if (/\/pages\/(?!HomePage\.tsx$)/.test(id)) return undefined
          return 'app'
        },
      },
    },
  },
})
