import test from 'node:test';
import assert from 'node:assert/strict';
import { loadClip, isClipCached, pruneClips, clipCacheSummary, clearClips, VIDEO_CACHE_LIMIT } from './video-cache.mjs';

const ORIGIN = 'https://ai.example.test';
function fakeWindow({ caches = true, putError = null, respond } = {}) {
 const store = new Map(), storage = new Map(), revoked = [], calls = [];
 let objects = 0;
 const cache = {
  match: async url => store.get(String(url))?.clone(),
  put: async (url, response) => { if (putError) throw putError; store.set(String(url), response); },
  delete: async key => store.delete(typeof key === 'string' ? key : key.url),
  keys: async () => [...store.keys()].map(url => ({ url })),
 };
 const win = {
  location: { href: ORIGIN + '/studio/' },
  localStorage: { getItem: key => storage.has(key) ? storage.get(key) : null, setItem: (key, value) => storage.set(key, String(value)) },
  URL: { createObjectURL: () => 'blob:' + ORIGIN + '/' + (++objects), revokeObjectURL: url => revoked.push(url) },
  fetch: async (url, init) => { calls.push({ url, init }); return respond ? respond(url, init) : video(1024); },
 };
 if (caches) win.caches = { open: async () => cache, delete: async () => { store.clear(); return true; } };
 return { win, store, storage, revoked, calls };
}
const video = (bytes, headers = {}) => new Response(new Uint8Array(bytes).fill(9), { headers: { 'Content-Type': 'video/mp4', 'Content-Length': String(bytes), ...headers } });
const clip = name => 'assets/clip-' + name + '-0123456789.mp4';

test('首次点击下载一次并存进本机缓存；再看直接读缓存不再请求；对象地址可重复释放', async () => {
 const env = fakeWindow(), progress = [];
 const first = await loadClip(clip('koi'), { win: env.win, onProgress: (received, total) => progress.push([received, total]) });
 assert.equal(env.calls.length, 1); assert.equal(env.calls[0].url, ORIGIN + '/studio/' + clip('koi'));
 assert.equal(env.calls[0].init.credentials, 'same-origin');
 assert.equal(first.cached, true); assert.equal(first.fromCache, false); assert.match(first.src, /^blob:/);
 assert.deepEqual(progress.at(-1), [1024, 1024]);
 assert.equal(await isClipCached(clip('koi'), { win: env.win }), true);
 assert.deepEqual(clipCacheSummary({ win: env.win }), { count: 1, bytes: 1024 });
 const second = await loadClip(clip('koi'), { win: env.win });
 assert.equal(env.calls.length, 1, '第二次不再访问服务器'); assert.equal(second.fromCache, true); assert.equal(second.cached, true);
 first.dispose(); first.dispose(); second.dispose();
 assert.deepEqual(env.revoked, [first.src, second.src]);
 assert.equal(env.store.size, 1, '释放对象地址不删除缓存');
});
test('超出条数上限时淘汰最久没看的片段，刚看过的保留', async () => {
 const env = fakeWindow(), names = Array.from({ length: VIDEO_CACHE_LIMIT.entries }, (_, index) => 'n' + index);
 for (const name of names) await loadClip(clip(name), { win: env.win });
 await loadClip(clip('n0'), { win: env.win });
 await loadClip(clip('fresh'), { win: env.win });
 assert.equal(env.calls.length, VIDEO_CACHE_LIMIT.entries + 1);
 assert.ok(env.store.has(ORIGIN + '/studio/' + clip('n0')), '刚看过的 n0 保留');
 assert.ok(!env.store.has(ORIGIN + '/studio/' + clip('n1')), '最久没看的 n1 被淘汰');
 assert.equal(clipCacheSummary({ win: env.win }).count, VIDEO_CACHE_LIMIT.entries);
});
test('服务返回错误、非视频或超出单段上限时报错且不写缓存', async () => {
 for (const [respond, message] of [
  [() => new Response('missing', { status: 404 }), /无法加载（404）/],
  [() => new Response('<html></html>', { headers: { 'Content-Type': 'text/html' } }), /不是视频/],
  [() => video(16, { 'Content-Length': String(VIDEO_CACHE_LIMIT.singleBytes + 1) }), /超出单段上限/],
  [() => new Response(new Uint8Array(0), { headers: { 'Content-Type': 'video/mp4' } }), /为空/],
 ]) {
  const env = fakeWindow({ respond });
  await assert.rejects(loadClip(clip('bad'), { win: env.win }), message);
  assert.equal(env.store.size, 0); assert.equal(clipCacheSummary({ win: env.win }).count, 0);
 }
});
test('下载途中停止：返回 AbortError，不留下缓存', async () => {
 const controller = new AbortController();
 const env = fakeWindow({ respond: () => new Response(new ReadableStream({ start(stream) { stream.enqueue(new Uint8Array(512)); } }), { headers: { 'Content-Type': 'video/mp4', 'Content-Length': '4096' } }) });
 await assert.rejects(loadClip(clip('stop'), { win: env.win, signal: controller.signal, onProgress: () => controller.abort() }), error => error.name === 'AbortError');
 assert.equal(env.store.size, 0);
});
test('浏览器不支持 Cache Storage 时直接返回原地址，不预先下载', async () => {
 const env = fakeWindow({ caches: false });
 const source = await loadClip(clip('plain'), { win: env.win });
 assert.equal(source.src, ORIGIN + '/studio/' + clip('plain')); assert.equal(source.cached, false); assert.equal(env.calls.length, 0);
 assert.equal(await isClipCached(clip('plain'), { win: env.win }), false);
 await pruneClips([clip('plain')], { win: env.win }); await clearClips({ win: env.win });
 source.dispose();
});
test('存储空间不足时仍可播放，只是不记入缓存', async () => {
 const env = fakeWindow({ putError: new DOMException('full', 'QuotaExceededError') });
 const source = await loadClip(clip('full'), { win: env.win });
 assert.match(source.src, /^blob:/); assert.equal(source.cached, false); assert.equal(clipCacheSummary({ win: env.win }).count, 0);
});
test('启动时删掉清单外的旧片段，并去掉浏览器已清掉的索引；清除按钮全部清空', async () => {
 const env = fakeWindow();
 await loadClip(clip('keep'), { win: env.win }); await loadClip(clip('old'), { win: env.win });
 await pruneClips([clip('keep')], { win: env.win });
 assert.deepEqual([...env.store.keys()], [ORIGIN + '/studio/' + clip('keep')]);
 assert.equal(clipCacheSummary({ win: env.win }).count, 1);
 env.store.clear();
 await pruneClips([clip('keep')], { win: env.win });
 assert.equal(clipCacheSummary({ win: env.win }).count, 0, '浏览器自行清掉的条目不再计入');
 await loadClip(clip('keep'), { win: env.win });
 await clearClips({ win: env.win });
 assert.equal(env.store.size, 0); assert.deepEqual(clipCacheSummary({ win: env.win }), { count: 0, bytes: 0 });
});
