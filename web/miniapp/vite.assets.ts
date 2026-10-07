import { createHash } from 'node:crypto'
import fs from 'node:fs'
import path from 'node:path'
import { brotliCompressSync, constants, gzipSync } from 'node:zlib'
import type { Plugin } from 'vite'

const SDK_QUERY = 'telegram-web-app.js?v=20261007'
const COMPRESS_EXT = new Set(['.js', '.css', '.html', '.svg'])

function sdkHashedName(root: string): string {
  const bytes = fs.readFileSync(path.join(root, 'public', 'telegram-web-app.js'))
  const hash = createHash('sha256').update(bytes).digest('hex').slice(0, 10)
  return `telegram-web-app.${hash}.js`
}

function precompress(file: string) {
  const buf = fs.readFileSync(file)
  if (buf.length < 1024) return
  const gz = gzipSync(buf, { level: 9 })
  const br = brotliCompressSync(buf, {
    params: { [constants.BROTLI_PARAM_QUALITY]: 11 },
  })
  if (gz.length + 32 < buf.length) fs.writeFileSync(`${file}.gz`, gz)
  if (br.length + 32 < buf.length) fs.writeFileSync(`${file}.br`, br)
}

function compressTree(dir: string) {
  for (const ent of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, ent.name)
    if (ent.isDirectory()) {
      compressTree(full)
      continue
    }
    if (!COMPRESS_EXT.has(path.extname(ent.name))) continue
    precompress(full)
  }
}

/** Fingerprint the Telegram SDK and emit .br/.gz next to static files. */
export function lifeosStaticCache(): Plugin {
  let outDir = ''
  let hashed = ''
  return {
    name: 'lifeos-static-cache',
    configResolved(config) {
      outDir = path.resolve(config.root, config.build.outDir)
      hashed = sdkHashedName(config.root)
    },
    transformIndexHtml: {
      order: 'post',
      handler(html, ctx) {
        if (ctx.server) return html
        return html
          .replaceAll(SDK_QUERY, hashed)
          .replace('<script type="module"', '<script type="module" fetchpriority="high"')
      },
    },
    closeBundle() {
      const from = path.join(outDir, 'telegram-web-app.js')
      if (fs.existsSync(from)) fs.renameSync(from, path.join(outDir, hashed))
      if (fs.existsSync(outDir)) compressTree(outDir)
    },
  }
}
