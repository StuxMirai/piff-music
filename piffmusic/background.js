// Keep the extension awake and handle cross-origin POSTs
try { console.log('PiffMusic background loaded'); } catch (_) {}

(typeof browser !== 'undefined' ? browser : chrome).runtime.onMessage.addListener((msg) => {
  if (!msg || msg.type !== 'postNowPlaying' || !msg.payload) return;
  const payload = msg.payload;
  return fetch('http://localhost:8080/webhook', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  }).then(() => ({ ok: true })).catch((err) => ({ ok: false, error: String(err) }));
});


