// Akrasia lives on the computer, so the phone keeps what it needs to open
// without it: the assets and the last copy of each tab. Writes made
// offline wait in queue.js, not here.
// The server replaces __BUILD__ with the id of the binary that serves this
// file, so a new binary starts new caches and the old ones are deleted.
const BUILD = "__BUILD__";
const SHELL = "shell-" + BUILD;
const PAGES = "pages-" + BUILD;

const ASSETS = ["/app.css", "/app.js", "/queue.js", "/htmx.min.js", "/manifest.json", "/icon-192.png", "/icon-512.png"];
const TABS = ["/today", "/all", "/dump"];
const FRAGMENT = "/__fragment__";

self.addEventListener("install", (e) => {
  e.waitUntil(
    (async () => {
      await (await caches.open(SHELL)).addAll(ASSETS);
      // the tabs need the login cookie: if there is none yet, the next visit stores them
      const pages = await caches.open(PAGES);
      await Promise.all(TABS.map((tab) => store(pages, tab, fetch(tab)).catch(() => {})));
      await self.skipWaiting();
    })(),
  );
});

self.addEventListener("activate", (e) => {
  e.waitUntil(
    (async () => {
      for (const name of await caches.keys()) {
        if (name !== SHELL && name !== PAGES) await caches.delete(name);
      }
      await self.clients.claim();
    })(),
  );
});

self.addEventListener("fetch", (e) => {
  const req = e.request;
  const url = new URL(req.url);
  if (req.method !== "GET" || url.origin !== self.location.origin) return;

  if (ASSETS.includes(url.pathname)) {
    e.respondWith(
      caches.match(req, { cacheName: SHELL }).then((hit) => hit || fetch(req)),
    );
  } else if (TABS.includes(url.pathname)) {
    e.respondWith(tab(e, req, url));
  }
});

// store keeps a successful answer of the server; the login page it is sent
// to without a cookie is not a tab.
async function store(cache, key, pending) {
  const res = await pending;
  if (res.ok && !res.redirected) await cache.put(key, res.clone());
  return res;
}

// tab: network first, the last copy when the server can't be reached.
// htmx asks for the app alone and a visit for the whole page; they are kept
// apart, and a fragment fetched online also refreshes the whole page, so
// reloading offline never shows something older than the last click.
async function tab(e, req, url) {
  const pages = await caches.open(PAGES);
  const fragment = req.headers.get("HX-Request") === "true" && req.headers.get("HX-History-Restore-Request") !== "true";
  const key = fragment ? FRAGMENT + url.pathname : url.pathname;

  try {
    const res = await store(pages, key, fetch(req));
    if (fragment && res.ok && !res.redirected) {
      e.waitUntil(store(pages, url.pathname, fetch(url.pathname)).catch(() => {}));
    }
    return res;
  } catch (err) {
    const hit = await pages.match(key);
    if (hit) return hit;
    // no copy of this tab: htmx fails and queue.js opens the whole page
    if (fragment) return Response.error();
    return (await pages.match("/today")) || new Response(OFFLINE, { status: 503, headers: { "Content-Type": "text/html; charset=utf-8" } });
  }
}

const OFFLINE =
  '<!doctype html><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">' +
  "<title>Akrasia</title><p style=\"font:16px system-ui;padding:2rem\">Offline. Open Akrasia once with the computer on, and it will open from here next time.</p>";
