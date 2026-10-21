// 本机验收夹具：模拟魔法 API 控制台宿主，同源提供内嵌工坊／配方和假的 Images 与 Chat Completions 接口。
// 只监听 127.0.0.1:4181；生图返回图库里已有的图片，文字分段流式返回固定 Markdown，不调用任何真实模型，也不读取真实密钥。
import { createServer } from 'node:http';
import { execFileSync } from 'node:child_process';
import { readFile, mkdir, writeFile, access } from 'node:fs/promises';
import { extname } from 'node:path';
import { fileURLToPath } from 'node:url';

const studio = new URL('../', import.meta.url);
const recipesSite = new URL('../../magic-recipes/site/', import.meta.url);
const recipesBuilt = new URL('./.embedded-recipes/', import.meta.url);
const runtime = new URL('../../family-runtime/host-client.js', import.meta.url);
const sample = await readFile(new URL('../assets/paper-fox.webp', import.meta.url));
const types = { '.html': 'text/html; charset=utf-8', '.css': 'text/css; charset=utf-8', '.js': 'text/javascript; charset=utf-8', '.mjs': 'text/javascript; charset=utf-8', '.webp': 'image/webp', '.png': 'image/png' };
// 与后端 familyStudioEmbeddedCSP／familyRecipeCSP 保持一致。
// 浏览器不接受 CSP 里的 IPv6 字面地址（http://[::1]:* 会被忽略并报错），本机只放行 localhost 与 127.0.0.1。
const local = 'http://localhost:* http://127.0.0.1:*';
const studioCSP = `default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self' data: https: ${local}; media-src 'self' blob:; connect-src 'self' https: ${local}; object-src 'none'; base-uri 'none'; form-action 'none'; frame-ancestors 'self'`;
const recipeCSP = `default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self' data: https: ${local}; connect-src 'self' https: ${local}; object-src 'none'; base-uri 'none'; form-action 'none'; frame-ancestors 'self'`;
const hostCSP = "default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self'; frame-src 'self'; connect-src 'self'; base-uri 'none'; form-action 'none'";
const studioFiles = new Set(['index.html', 'theme.css', 'app.mjs', 'core.mjs', 'icons.mjs', 'catalog.mjs', 'new-gallery.mjs', 'generate.mjs', 'videos.mjs', 'video-cache.mjs', 'host-client.js']);
const clipFile = /^assets\/clip-[a-z0-9-]+-[a-f0-9]{10}\.mp4$/;
// 夹具运行目录（Git 忽略）：生成脚本对着本夹具跑时把清单和成片写在这里，不改动产品目录。
const videoRun = new URL('./.video-run/', import.meta.url);
const videoJobs = new Map();
let videoCounter = 0;
const chatPieces = ['## 审查结论\n\n', '夹具回复：**未调用真实模型**，', '只用于检查内嵌配方的流式排版。\n\n', '| 级别 | 位置 | 问题 |\n| --- | --- | --- |\n', '| 中 | `api.js:12` | 缺少超时处理 |\n\n', '- [ ] 补充失败重试的测试\n', '- [x] 已确认输入校验\n'];

await mkdir(recipesBuilt, { recursive: true });
execFileSync(process.execPath, [fileURLToPath(new URL('build.mjs', recipesSite)), '--embedded', '--output', fileURLToPath(recipesBuilt)], { stdio: 'pipe' });

const hostPage = `<!doctype html><html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>本机控制台夹具</title><link rel="stylesheet" href="/fixture.css"><script src="/fixture.js" defer></script></head>
<body><header class="bar"><strong>本机控制台夹具（模拟数据，不调用真实模型）</strong><button id="studio" type="button">魔法工坊</button><button id="recipes" type="button">魔法配方</button><button id="theme" type="button">切换主题</button><button id="logout" type="button">模拟退出登录</button><span id="log" role="status"></span></header>
<iframe id="frame" title="内嵌产品" src="/studio/?embedded=1"></iframe></body></html>`;
const hostStyle = `html,body{margin:0;height:100%;font:14px/1.5 'Segoe UI','Microsoft YaHei',sans-serif;background:#f5f8f8;color:#1d3033}html.dark,html.dark body{background:#162124;color:#e9f1f0}.bar{display:flex;gap:8px;align-items:center;flex-wrap:wrap;padding:8px 12px;border-bottom:1px solid #9993}.bar button{min-height:32px;padding:4px 10px;border-radius:8px;border:1px solid #9996;background:transparent;color:inherit;cursor:pointer}#log{margin-left:auto;font-size:12px;opacity:.75}iframe{display:block;width:100%;height:calc(100% - 50px);border:0}`;
const hostScript = `(() => {
  const frame = document.getElementById('frame'), log = document.getElementById('log');
  const key = { id: 7, name: '模拟生图密钥', group: '生图分组', protocol: 'chat' };
  let nonce = '', dark = false;
  const theme = () => dark ? 'dark' : 'light';
  const send = (type, details = {}) => frame.contentWindow.postMessage({ type, nonce, ...details }, location.origin);
  frame.addEventListener('load', () => {
    nonce = Array.from(crypto.getRandomValues(new Uint8Array(32)), byte => byte.toString(16).padStart(2, '0')).join('');
    send('mofa-host-init', { theme: theme() });
  });
  window.addEventListener('message', event => {
    const data = event.data;
    if (event.origin !== location.origin || event.source !== frame.contentWindow || data?.type !== 'mofa-host-request' || data.nonce !== nonce) return;
    log.textContent = '收到：' + data.action + (data.payload?.kind ? '（' + data.payload.kind + '）' : '');
    const reply = (result, error) => setTimeout(() => send('mofa-host-response', error ? { id: data.id, error } : { id: data.id, result }), 350);
    if (data.action === 'theme') { dark = data.payload.theme === 'dark'; document.documentElement.classList.toggle('dark', dark); reply(null); return; }
    if (['family', 'studio-submit', 'manage-keys'].includes(data.action)) { reply(null); log.textContent = '宿主会跳转：' + data.action; return; }
    if (frame.src.includes('/studio/') && data.payload?.kind !== 'image') { reply(null, '魔法工坊只支持生图连接。'); return; }
    if (data.action === 'keys') reply([key]);
    else if (data.action === 'models') reply({ models: ['gpt-image-1', 'mock-image-model', 'text-model'], protocol: 'chat' });
    else if (data.action === 'apply') reply({ kind: data.payload.kind, protocol: data.payload.kind === 'image' ? 'images' : data.payload.protocol, base: location.origin + '/v1', key: 'fixture-fake-key', model: data.payload.model, size: '' });
    else reply(null, '不支持此配置操作。');
  });
  document.getElementById('theme').addEventListener('click', () => { dark = !dark; document.documentElement.classList.toggle('dark', dark); send('mofa-host-theme', { theme: theme() }); });
  document.getElementById('logout').addEventListener('click', () => { send('mofa-host-session-ended'); log.textContent = '已通知子页退出'; });
  document.getElementById('studio').addEventListener('click', () => { frame.src = '/studio/?embedded=1'; });
  document.getElementById('recipes').addEventListener('click', () => { frame.src = '/recipes/?embedded=1'; });
})();`;

const server = createServer(async (request, response) => {
  const url = new URL(request.url, 'http://127.0.0.1');
  const send = (status, body, type, csp) => { response.writeHead(status, { 'Content-Type': type, 'Cache-Control': 'no-store', 'X-Content-Type-Options': 'nosniff', ...(csp ? { 'Content-Security-Policy': csp } : {}) }); response.end(request.method === 'HEAD' ? undefined : body); };
  try {
    if (url.pathname === '/v1/images/generations' && request.method === 'POST') {
      let raw = '';
      for await (const chunk of request) raw += chunk;
      const body = JSON.parse(raw || '{}');
      if (request.headers.authorization !== 'Bearer fixture-fake-key') { send(401, '{}', 'application/json'); return; }
      if (!body.prompt) { send(400, '{}', 'application/json'); return; }
      await new Promise(resolve => setTimeout(resolve, 1600));
      send(200, JSON.stringify({ output_format: 'webp', data: [{ b64_json: sample.toString('base64'), revised_prompt: '夹具返回图库中的纸艺狐狸，未调用真实模型。' }] }), 'application/json');
      return;
    }
    if (url.pathname === '/v1/chat/completions' && request.method === 'POST') {
      let raw = '';
      for await (const chunk of request) raw += chunk;
      const body = JSON.parse(raw || '{}');
      if (request.headers.authorization !== 'Bearer fixture-fake-key') { send(401, '{}', 'application/json'); return; }
      if (!Array.isArray(body.messages) || !body.messages.length) { send(400, '{}', 'application/json'); return; }
      if (!body.stream) { send(200, JSON.stringify({ choices: [{ message: { role: 'assistant', content: chatPieces.join('') } }] }), 'application/json'); return; }
      let closed = false;
      response.on('close', () => { closed = true; });
      response.writeHead(200, { 'Content-Type': 'text/event-stream; charset=utf-8', 'Cache-Control': 'no-store', 'X-Content-Type-Options': 'nosniff' });
      for (const piece of chatPieces) {
        if (closed) return;
        response.write('data: ' + JSON.stringify({ choices: [{ delta: { content: piece } }] }) + '\n\n');
        await new Promise(resolve => setTimeout(resolve, 260));
      }
      if (!closed) response.end('data: [DONE]\n\n');
      return;
    }
    // 假的图生视频接口（形状同本站 /v1/videos/generations）：第一次查询返回处理中，之后返回完成；成片为本机录制的测试样片。
    if (url.pathname === '/v1/videos/generations' && request.method === 'POST') {
      let raw = '';
      for await (const chunk of request) raw += chunk;
      if (request.headers.authorization !== 'Bearer fixture-fake-key') { send(401, JSON.stringify({ error: { message: 'invalid key' } }), 'application/json'); return; }
      const body = JSON.parse(raw || '{}');
      if (!body.prompt || !/^data:image\/(webp|png);base64,/.test(body.image?.url || '')) { send(400, JSON.stringify({ error: { message: 'image and prompt are required' } }), 'application/json'); return; }
      const id = 'fixture-' + (++videoCounter);
      videoJobs.set(id, { polls: 0 });
      send(200, JSON.stringify({ request_id: id }), 'application/json');
      return;
    }
    const videoMatch = url.pathname.match(/^\/v1\/videos\/(fixture-\d+)(\/content)?$/);
    if (videoMatch && request.method === 'GET') {
      if (request.headers.authorization !== 'Bearer fixture-fake-key') { send(401, '{}', 'application/json'); return; }
      const job = videoJobs.get(videoMatch[1]);
      if (!job) { send(404, JSON.stringify({ error: { message: 'unknown task' } }), 'application/json'); return; }
      if (!videoMatch[2]) { job.polls++; send(200, JSON.stringify(job.polls < 2 ? { status: 'pending' } : { status: 'done', video: { url: '/v1/videos/' + videoMatch[1] + '/content', duration: 5 } }), 'application/json'); return; }
      try { send(200, await readFile(new URL('sample-clip.mp4', videoRun)), 'video/mp4'); }
      catch { send(404, JSON.stringify({ error: { message: '先上传测试样片 fixtures/.video-run/sample-clip.mp4' } }), 'application/json'); }
      return;
    }
    // 程序动效录制结果：tools/motion/record.html 录好的 MP4 与说明暂存在 .video-run/motion，再由 tools/import-motion-clip.mjs 校验入库。
    if (url.pathname === '/fixture/motion-clip' && request.method === 'POST') {
      if (request.headers.origin !== 'http://127.0.0.1:4181' && request.headers.origin !== 'http://localhost:4181') { send(403, '', 'text/plain'); return; }
      const id = url.searchParams.get('id') || '', meta = url.searchParams.get('meta') || '';
      if (!/^clip-[a-z0-9-]+$/.test(id)) { send(400, 'bad id', 'text/plain'); return; }
      const parts = []; let size = 0;
      for await (const chunk of request) { size += chunk.length; if (size > 20 * 1024 * 1024) { send(413, '', 'text/plain'); return; } parts.push(chunk); }
      const data = Buffer.concat(parts);
      if (data.subarray(4, 8).toString('latin1') !== 'ftyp') { send(415, 'not mp4', 'text/plain'); return; }
      let info;
      try { info = JSON.parse(meta); } catch { send(400, 'bad meta', 'text/plain'); return; }
      const dir = new URL('motion/', videoRun);
      await mkdir(dir, { recursive: true });
      await writeFile(new URL(id + '.mp4', dir), data);
      await writeFile(new URL(id + '.json', dir), JSON.stringify({ ...info, id, bytes: data.length, recordedAt: new Date().toISOString(), userAgent: String(request.headers['user-agent'] || '').slice(0, 200) }, null, 1));
      send(200, JSON.stringify({ bytes: data.length }), 'application/json');
      return;
    }
    // 仅供本机验收：页面录制的测试样片（MP4，≤8MB）存到 .video-run，不写入产品目录。
    if (url.pathname === '/fixture/sample-clip' && request.method === 'POST') {
      if (request.headers.origin !== 'http://127.0.0.1:4181' && request.headers.origin !== 'http://localhost:4181') { send(403, '', 'text/plain'); return; }
      const parts = []; let size = 0;
      for await (const chunk of request) { size += chunk.length; if (size > 8 * 1024 * 1024) { send(413, '', 'text/plain'); return; } parts.push(chunk); }
      const data = Buffer.concat(parts);
      if (data.subarray(4, 8).toString('latin1') !== 'ftyp') { send(415, 'not mp4', 'text/plain'); return; }
      await mkdir(videoRun, { recursive: true }); await writeFile(new URL('sample-clip.mp4', videoRun), data);
      send(200, JSON.stringify({ bytes: data.length }), 'application/json');
      return;
    }
    if (!['GET', 'HEAD'].includes(request.method)) { send(405, '', 'text/plain'); return; }
    if (url.pathname === '/') { send(200, hostPage, types['.html'], hostCSP); return; }
    if (url.pathname === '/fixture.js') { send(200, hostScript, types['.js']); return; }
    if (url.pathname === '/fixture.css') { send(200, hostStyle, types['.css']); return; }
    if (url.pathname === '/api/v1/studio/gallery') { send(200, JSON.stringify({ code: 0, data: [] }), 'application/json'); return; }
    // 程序动效录制页：只在本机夹具提供；读取仓库里原图 PNG（无损）作为首帧。
    if (url.pathname.startsWith('/motion/')) {
      const file = url.pathname.slice(8) || 'record.html';
      const motionCSP = "default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self'; connect-src 'self'; object-src 'none'; base-uri 'none'; form-action 'none'";
      if (['record.html', 'record.mjs', 'record.css', 'effects.mjs', 'recipes.mjs'].includes(file)) { send(200, await readFile(new URL('../tools/motion/' + file, import.meta.url)), types[extname(file)] || 'text/plain', motionCSP); return; }
      const source = file.match(/^source\/([a-z0-9-]+)\.png$/);
      if (source) { send(200, await readFile(new URL('assets/' + source[1] + '.png', studio)), 'image/png', motionCSP); return; }
      send(404, '', 'text/plain');
      return;
    }
    if (url.pathname === '/studio' || url.pathname === '/recipes') { response.writeHead(308, { Location: url.pathname + '/' + url.search }); response.end(); return; }
    if (url.pathname.startsWith('/studio/')) {
      const file = url.pathname.slice(8) || 'index.html';
      if (!studioFiles.has(file) && !/^assets\/[a-z0-9-]+\.webp$/.test(file) && file !== 'assets/mofa-mark-flat.png' && !clipFile.test(file)) { send(404, '', 'text/plain'); return; }
      // 生成脚本对着本夹具运行后，清单与成片在 .video-run；存在时优先使用，产品目录保持不变。
      const shadow = (file === 'videos.mjs' || clipFile.test(file)) && await access(new URL(file, videoRun)).then(() => true, () => false);
      let data = await readFile(file === 'host-client.js' ? runtime : new URL(file, shadow ? videoRun : studio));
      if (file === 'index.html') data = Buffer.from(String(data).replace('<head>', '<head>\n<meta name="mofa-api-site" content="same-origin">'));
      if (clipFile.test(file)) {
        response.writeHead(200, { 'Content-Type': 'video/mp4', 'Content-Length': data.length, 'Cache-Control': 'public, max-age=31536000, immutable', 'X-Content-Type-Options': 'nosniff', 'Content-Security-Policy': studioCSP });
        response.end(request.method === 'HEAD' ? undefined : data);
        return;
      }
      send(200, data, types[extname(file)] || 'application/octet-stream', studioCSP);
      return;
    }
    if (url.pathname.startsWith('/recipes/')) {
      const file = url.pathname.slice(9) || 'index.html';
      if (!['index.html', 'theme.css', 'bundle.js'].includes(file)) { send(404, '', 'text/plain'); return; }
      send(200, await readFile(new URL(file, recipesBuilt)), types[extname(file)], recipeCSP);
      return;
    }
    send(404, '', 'text/plain');
  } catch { send(404, '', 'text/plain'); }
});
server.on('error', error => { console.error(error.message); process.exitCode = 1; });
server.listen(4181, '127.0.0.1', () => console.log('本机控制台夹具：http://127.0.0.1:4181 （模拟数据，不调用真实模型）'));
