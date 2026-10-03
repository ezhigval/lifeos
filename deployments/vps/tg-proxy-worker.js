// tg-proxy worker — deploy on Cloudflare Workers (free tier, 100k req/day).
// Proxies api.telegram.org calls for hosts where Yandex Cloud blocks direct egress.
export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    if (url.pathname !== "/fetch") return new Response("not found", { status: 404 });
    const path = url.searchParams.get("path"); // e.g. /bot<token>/sendMessage
    if (!path || !path.startsWith("/bot")) return new Response("bad path", { status: 400 });
    // Optional shared secret to prevent open-proxy abuse:
    if (env.PROXY_SECRET && request.headers.get("x-proxy-secret") !== env.PROXY_SECRET)
      return new Response("unauthorized", { status: 401 });

    const init = { method: request.method, headers: { "content-type": request.headers.get("content-type") || "application/json" } };
    if (request.method !== "GET") init.body = await request.arrayBuffer();
    const resp = await fetch("https://api.telegram.org" + path, init);
    return new Response(resp.body, { status: resp.status, headers: { "content-type": "application/json" } });
  },
};
