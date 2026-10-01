(() => {
  'use strict';
  const embedded = window.parent !== window && new URLSearchParams(location.search).get('embedded') === '1';
  const pending = new Map();
  let nonce = '', counter = 0;
  function theme(value) {
    const next = value === 'dark' ? 'dark' : 'light';
    document.documentElement.dataset.theme = next;
    document.dispatchEvent(new CustomEvent('mofa-theme', { detail: next }));
  }
  function savedTheme() {
    try { return localStorage.getItem('theme'); } catch { return 'light'; }
  }
  theme(savedTheme());
  if (embedded) document.documentElement.classList.add('mofa-embedded');
  function request(action, payload = {}) {
    if (!embedded || !nonce) return Promise.reject(new Error('控制台连接尚未就绪，请稍后再试。'));
    const id = String(++counter);
    return new Promise((resolve, reject) => {
      const timer = setTimeout(() => { pending.delete(id); reject(new Error('读取配置超时，请重试。')); }, 35000);
      pending.set(id, { resolve, reject, timer });
      window.parent.postMessage({ type: 'mofa-host-request', nonce, id, action, payload }, location.origin);
    });
  }
  window.addEventListener('message', event => {
    if (!embedded || event.origin !== location.origin || event.source !== window.parent) return;
    const data = event.data;
    if (!data || typeof data !== 'object') return;
    if (data.type === 'mofa-host-init' && /^[a-f0-9]{64}$/.test(data.nonce) && (!nonce || nonce === data.nonce)) {
      nonce = data.nonce;
      theme(data.theme);
      document.dispatchEvent(new CustomEvent('mofa-host-ready'));
    } else if (nonce && data.nonce === nonce) {
      if (data.type === 'mofa-host-theme') theme(data.theme);
      if (data.type === 'mofa-host-session-ended') {
        for (const entry of pending.values()) { clearTimeout(entry.timer); entry.reject(new Error('登录已失效，请重新登录。')); }
        pending.clear();
        nonce = '';
        document.dispatchEvent(new CustomEvent('mofa-session-ended'));
      }
      if (data.type === 'mofa-host-response' && pending.has(data.id)) {
        const entry = pending.get(data.id);
        pending.delete(data.id); clearTimeout(entry.timer);
        if (data.error) entry.reject(new Error(data.error)); else entry.resolve(data.result);
      }
    }
  });
  function setTheme(value) {
    theme(value);
    try { localStorage.setItem('theme', value); } catch {}
    if (embedded && nonce) window.parent.postMessage({ type: 'mofa-host-request', nonce, id: String(++counter), action: 'theme', payload: { theme: value } }, location.origin);
  }
  window.addEventListener('storage', event => { if (event.key === 'theme') theme(event.newValue); });
  globalThis.MofaFamilyHost = { embedded, request, setTheme, get ready() { return Boolean(nonce); } };
})();
