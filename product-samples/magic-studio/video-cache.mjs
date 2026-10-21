// 工坊短视频的浏览器缓存：只在用户点击播放时下载一次，存进 Cache Storage，下次直接从本机读取，不再占用服务器流量。
// 文件名带内容摘要，内容变了地址就变，旧缓存不会被误用；总量超出上限时先淘汰最久未看的片段。
// 浏览器不支持 Cache Storage（如部分隐私模式）时返回原地址，交给浏览器普通 HTTP 缓存，仍然只在点击后加载。
export const VIDEO_CACHE_NAME = 'mofa-studio-videos-v1';
export const VIDEO_CACHE_LIMIT = Object.freeze({ bytes: 80 * 1024 * 1024, entries: 16, singleBytes: 30 * 1024 * 1024 });
const INDEX_KEY = 'mofa-studio:video-cache';

function absolute(src, win) { return new URL(src, win.location.href).href; }
function readIndex(win) {
 try {
  const raw = JSON.parse(win.localStorage.getItem(INDEX_KEY) || '[]');
  return Array.isArray(raw) ? raw.filter(entry => entry && typeof entry.url === 'string' && Number.isFinite(entry.bytes) && entry.bytes >= 0).slice(-64) : [];
 } catch { return []; }
}
function writeIndex(win, list) { try { win.localStorage.setItem(INDEX_KEY, JSON.stringify(list)); } catch {} }
function touch(win, url, bytes) {
 const list = readIndex(win).filter(entry => entry.url !== url);
 list.push({ url, bytes, at: Date.now() });
 writeIndex(win, list);
 return list;
}
async function openCache(win) {
 if (!win.caches?.open) return null;
 try { return await win.caches.open(VIDEO_CACHE_NAME); } catch { return null; }
}
async function enforce(win, cache, keepUrl) {
 let list = readIndex(win), total = list.reduce((sum, entry) => sum + entry.bytes, 0);
 while (list.length && (total > VIDEO_CACHE_LIMIT.bytes || list.length > VIDEO_CACHE_LIMIT.entries)) {
  const index = list.findIndex(entry => entry.url !== keepUrl);
  if (index < 0) break;
  const [oldest] = list.splice(index, 1);
  total -= oldest.bytes;
  try { await cache.delete(oldest.url); } catch {}
 }
 writeIndex(win, list);
}
async function readBody(response, total, onProgress, signal) {
 if (!response.body?.getReader) return new Uint8Array(await response.arrayBuffer());
 const reader = response.body.getReader(), parts = [];
 let received = 0;
 const abort = () => { reader.cancel().catch(() => {}); };
 signal?.addEventListener('abort', abort, { once: true });
 try {
  for (;;) {
   const { done, value } = await reader.read();
   if (signal?.aborted) throw new DOMException('已停止加载', 'AbortError');
   if (done) break;
   parts.push(value); received += value.byteLength;
   if (received > VIDEO_CACHE_LIMIT.singleBytes) throw new Error('视频文件超出单段上限。');
   onProgress?.(received, total);
  }
 } finally { signal?.removeEventListener('abort', abort); }
 const bytes = new Uint8Array(received);
 let offset = 0;
 for (const part of parts) { bytes.set(part, offset); offset += part.byteLength; }
 return bytes;
}

// 返回 { src, cached, fromCache, dispose }：src 可直接给 <video>；dispose 释放本次创建的对象地址，不删除缓存。
export async function loadClip(src, { win = globalThis, signal, onProgress, type = 'video/mp4' } = {}) {
 const url = absolute(src, win), cache = await openCache(win);
 if (!cache || typeof win.URL?.createObjectURL !== 'function') return { src: url, cached: false, fromCache: false, dispose() {} };
 const objectSource = (blob, cached, fromCache) => {
  const objectUrl = win.URL.createObjectURL(blob);
  let released = false;
  return { src: objectUrl, cached, fromCache, dispose() { if (!released) { released = true; win.URL.revokeObjectURL(objectUrl); } } };
 };
 try {
  const hit = await cache.match(url);
  if (hit) {
   const blob = await hit.blob();
   if (blob.size > 0) {
    touch(win, url, blob.size);
    return objectSource(blob.type ? blob : new Blob([blob], { type }), true, true);
   }
   await cache.delete(url);
  }
 } catch {}
 // 默认缓存模式：Cache Storage 被浏览器清掉时，服务端的长期缓存头仍可能直接命中，减少重复下载。
 const response = await win.fetch(url, { signal, credentials: 'same-origin' });
 if (!response.ok) throw new Error('视频暂时无法加载（' + response.status + '）。');
 const contentType = response.headers.get('content-type') || type;
 if (!/^video\//i.test(contentType)) throw new Error('服务返回的不是视频文件。');
 const length = Number(response.headers.get('content-length')) || 0;
 if (length > VIDEO_CACHE_LIMIT.singleBytes) throw new Error('视频文件超出单段上限。');
 const bytes = await readBody(response, length, onProgress, signal);
 if (!bytes.byteLength) throw new Error('视频文件为空。');
 const blob = new Blob([bytes], { type: contentType });
 let cached = false;
 try {
  await cache.put(url, new Response(blob, { headers: { 'Content-Type': contentType, 'Content-Length': String(blob.size) } }));
  touch(win, url, blob.size);
  await enforce(win, cache, url);
  cached = true;
 } catch {}
 return objectSource(blob, cached, false);
}

export async function isClipCached(src, { win = globalThis } = {}) {
 const cache = await openCache(win);
 if (!cache) return false;
 try { return Boolean(await cache.match(absolute(src, win))); } catch { return false; }
}
// 清单里已不存在的片段（重新生成或下架）连同索引一起删除；浏览器自行清掉的条目从索引里去掉。
export async function pruneClips(validSrcs, { win = globalThis } = {}) {
 const cache = await openCache(win);
 if (!cache) return;
 const valid = new Set(validSrcs.map(src => absolute(src, win)));
 try { for (const request of await cache.keys()) if (!valid.has(request.url)) await cache.delete(request); } catch {}
 const present = new Set();
 try { for (const request of await cache.keys()) present.add(request.url); } catch {}
 writeIndex(win, readIndex(win).filter(entry => valid.has(entry.url) && present.has(entry.url)));
}
export function clipCacheSummary({ win = globalThis } = {}) {
 const list = readIndex(win);
 return { count: list.length, bytes: list.reduce((sum, entry) => sum + entry.bytes, 0) };
}
export async function clearClips({ win = globalThis } = {}) {
 try { await win.caches?.delete(VIDEO_CACHE_NAME); } catch {}
 writeIndex(win, []);
}
