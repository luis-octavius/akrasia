// Akrasia needs the server (it lives on the computer), so nothing is
// cached: this worker only exists so the page can be installed as an app.
// The server replaces __BUILD__ with the id of the binary that serves it.
const BUILD = "__BUILD__";
self.addEventListener("install", () => self.skipWaiting());
self.addEventListener("activate", (e) => e.waitUntil(self.clients.claim()));
self.addEventListener("fetch", () => {});
