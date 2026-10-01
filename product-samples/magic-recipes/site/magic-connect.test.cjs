const test = require('node:test');
const assert = require('node:assert/strict');
const connector = require('./magic-connect.cjs');
const { webcrypto } = require('node:crypto');

test('一体化部署仅信任当前完整来源，不借此允许第三方登录站点', () => {
  const own = 'https://family.example.test';
  assert.equal(connector.portalURL(own, own).origin, own);
  const connection = connector.begin(own, own, webcrypto, true);
  assert.equal(new URL(connector.loginURL(connection, 'text')).origin, own);
  assert.throws(() => connector.portalURL('https://evil.test', own));
  assert.throws(() => connector.portalURL(own, own + '/recipes/'));
  assert.throws(() => connector.begin(own, own, webcrypto));
});

test('登录仅接受官方域名和明确本机 HTTP，不接受仿冒、凭据或路径', () => {
  assert.equal(connector.portalURL('https://ai.mofamilys.com/').origin, 'https://ai.mofamilys.com');
  assert.equal(connector.portalURL('http://127.0.0.1:4175').origin, 'http://127.0.0.1:4175');
  for (const portal of ['https://ai.mofamilys.com.evil.test', 'https://evil.test', 'http://ai.mofamilys.com', 'https://me:secret@ai.mofamilys.com', 'https://ai.mofamilys.com/?key=secret', 'https://ai.mofamilys.com/#secret', 'https://ai.mofamilys.com/login', 'javascript:alert(1)']) assert.throws(() => connector.portalURL(portal));
});

test('URL仅包含页面来源、随机nonce与用途，不携带密钥或材料', () => {
  const connection = connector.begin('https://ai.mofamilys.com', 'http://127.0.0.1:4178', webcrypto);
  assert.match(connection.nonce, /^[a-f0-9]{64}$/);
  assert.notEqual(connection.nonce, connector.begin('https://ai.mofamilys.com', connection.origin, webcrypto).nonce);
  const url = new URL(connector.loginURL(connection, 'image'));
  assert.equal(url.pathname, '/connect/recipes');
  assert.deepEqual([...url.searchParams.keys()], ['origin', 'nonce', 'kind']);
  assert.equal(url.searchParams.get('kind'), 'image');
  assert.throws(() => connector.begin(connection.portal, 'null', webcrypto));
  assert.throws(() => connector.begin(connection.portal, 'http://remote.test', webcrypto));
});

test('回传必须精确匹配窗口、来源、nonce、用途、类型与期限', () => {
  const popup = {};
  const connection = { popup, portal: 'https://ai.mofamilys.com', nonce: 'a'.repeat(64), kind: 'text', expires: 300 };
  const event = { source: popup, origin: connection.portal, data: { type: 'mofa-recipes-connection', nonce: connection.nonce, config: { kind: 'text' } } };
  assert.equal(connector.accepts(event, connection, 299), true);
  assert.equal(connector.accepts(event, connection, 300), false);
  assert.equal(connector.accepts(event, undefined, 200), false);
  for (const forged of [{ ...event, source: {} }, { ...event, origin: 'https://evil.test' }, { ...event, data: {} }, { ...event, data: { ...event.data, nonce: 'b'.repeat(64) } }, { ...event, data: { ...event.data, config: { kind: 'image' } } }]) assert.equal(connector.accepts(forged, connection, 200), false);
});
