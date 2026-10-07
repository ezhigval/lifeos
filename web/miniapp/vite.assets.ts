import fs from 'node:fs'
import path from 'node:path'
import { brotliCompressSync, constants, gzipSync } from 'node:zlib'
import type { Plugin } from 'vite'

const COMPRESS_EXT = new Set(['.js', '.css', '.html', '.svg'])

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

/** Emit .br/.gz next to built static files. The Telegram SDK stays at
 *  /app/telegram-web-app.js: main.tsx injects that exact URL after React mounts. */
export function lifeosStaticCache(): Plugin {
  let outDir = ''
  return {
    name: 'lifeos-static-cache',
    apply: 'build',
    enforce: 'post',
    configResolved(config) {
      outDir = path.resolve(config.root, config.build.outDir)
    },
    closeBundle() {
      if (fs.existsSync(outDir)) compressTree(outDir)
    },
  }
}
