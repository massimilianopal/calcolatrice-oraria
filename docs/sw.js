// Tiene in cache i file della pagina, così funziona anche senza rete.
// Prima prova la rete (per avere sempre l'ultima versione), poi la cache.
const CACHE = "calcolatrice-ore";
const FILE = ["./", "index.html", "app.js", "ore.js", "manifest.webmanifest", "favicon.ico", "icona-192.png", "icona-512.png"];

self.addEventListener("install", (e) => {
  e.waitUntil(caches.open(CACHE).then((c) => c.addAll(FILE)));
  self.skipWaiting();
});

self.addEventListener("activate", (e) => e.waitUntil(self.clients.claim()));

self.addEventListener("fetch", (e) => {
  if (e.request.method !== "GET" || new URL(e.request.url).origin !== location.origin) return;
  e.respondWith(
    fetch(e.request)
      .then((r) => {
        const copia = r.clone();
        caches.open(CACHE).then((c) => c.put(e.request, copia));
        return r;
      })
      .catch(() => caches.match(e.request, { ignoreSearch: true })),
  );
});
