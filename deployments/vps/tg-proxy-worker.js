// tg-proxy worker — deploy on Cloudflare Workers (free tier, 100k req/day).
// Proxies api.telegram.org for hosts whose egress to Telegram is blocked.
// Redeploy after pulling this file so /file/bot downloads are allowed.
export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    if (url.pathname !== "/fetch") return new Response("not found", { status: 404 });

    const path = url.searchParams.get("path"); // /bot<token>/method or /file/bot<token>/<file>
    if (!allowedPath(path)) return new Response("bad path", { status: 400 });

    // Optional shared secret. Set with: npx wrangler secret put PROXY_SECRET
    // The same value goes to /opt/lifeos/secrets/tg-proxy.env on the VM, never to git.
    if (env.PROXY_SECRET && request.headers.get("x-proxy-secret") !== env.PROXY_SECRET) {
      return new Response("unauthorized", { status: 401 });
    }

    const headers = {};
    const ct = request.headers.get("content-type");
    if (ct) headers["content-type"] = ct;

    const init = { method: request.method, headers };
    if (request.method !== "GET" && request.method !== "HEAD") {
      init.body = await request.arrayBuffer();
    }

    const resp = await fetch("https://api.telegram.org" + path, init);
    const outHeaders = new Headers();
    const respType = resp.headers.get("content-type");
    if (respType) outHeaders.set("content-type", respType);
    return new Response(resp.body, { status: resp.status, headers: outHeaders });
  },
};

function allowedPath(path) {
  if (typeof path !== "string" || path.length === 0 || path.length > 4096) return false;
  if (path.includes("..") || path.includes("\\") || path.includes("\0") || path.includes("@")) return false;
  return path.startsWith("/bot") || path.startsWith("/file/bot");
}
