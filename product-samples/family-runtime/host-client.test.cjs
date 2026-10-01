const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { createRequire } = require('node:module');
const { JSDOM } = createRequire(path.resolve(__dirname, '../../frontend/package.json'))('jsdom');
const script = fs.readFileSync(path.join(__dirname, 'host-client.js'), 'utf8');
test('宿主消息只接受当前父窗口、同源和nonce；默认浅色、同步主题及退出取消请求', async () => {
  const parent = { postMessage() {} }, dom = new JSDOM('<html><body></body></html>', { runScripts: 'outside-only', url: 'https://api.test/recipes/?embedded=1' });
  const window = dom.window;
  Object.defineProperty(window, 'parent', { value: parent }); window.eval(script);
  const send = (data, origin = 'https://api.test', source = parent) => window.dispatchEvent(new window.MessageEvent('message', { origin, source, data }));
  const nonce = 'a'.repeat(64);
  try {
    assert.equal(window.document.documentElement.dataset.theme, 'light');
    send({ type: 'mofa-host-init', nonce, theme: 'dark' }, 'https://evil.test'); assert.equal(window.MofaFamilyHost.ready, false);
    send({ type: 'mofa-host-init', nonce, theme: 'dark' }, 'https://api.test', {}); assert.equal(window.MofaFamilyHost.ready, false);
    send({ type: 'mofa-host-init', nonce, theme: 'dark' }); assert.equal(window.MofaFamilyHost.ready, true);
    send({ type: 'mofa-host-theme', nonce: 'wrong', theme: 'light' }); assert.equal(window.document.documentElement.dataset.theme, 'dark');
    send({ type: 'mofa-host-theme', nonce, theme: 'light' }); assert.equal(window.document.documentElement.dataset.theme, 'light');
    const pending = window.MofaFamilyHost.request('keys', { kind: 'text' });
    const assertion = assert.rejects(pending, /登录已失效/);
    send({ type: 'mofa-host-session-ended', nonce }); await assertion;
    assert.equal(window.MofaFamilyHost.ready, false);
  } finally { window.close(); }
});
