/* Cache only fingerprinted Mini App files. Never cache index.html or /api:
   a stale shell points at deleted chunks and the module never starts. */
const CACHE = 'lifeos-static-v1'

self.addEventListener('install', (event) => {
  event.waitUntil(self.skipWaiting())
})

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches
      .keys()
      .then((keys) => Promise.all(keys.filter((key) => key !== CACHE).map((key) => caches.delete(key))))
      .then(() => self.clients.claim()),
  )
})

function isImmutableAsset(url) {
  if (url.origin !== self.location.origin) return false
  const path = url.pathname
  if (path.startsWith('/app/assets/') && !path.endsWith('.map')) return true
  return /^\/app\/telegram-web-app\.[a-f0-9]{8,}\.js$/.test(path)
}

self.addEventListener('fetch', (event) => {
  const req = event.request
  if (req.method !== 'GET') return
  let url
  try {
    url = new URL(req.url)
  } catch (e) {
    return
  }
  if (!isImmutableAsset(url)) return
  event.respondWith(
    caches.open(CACHE).then(async (cache) => {
      const hit = await cache.match(req)
      if (hit) return hit
      const res = await fetch(req)
      if (res.ok && res.type === 'basic') cache.put(req, res.clone())
      return res
    }),
  )
})
