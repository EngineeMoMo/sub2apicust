import test from 'node:test';
import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { mkdtemp, readFile, readdir, rm, writeFile, mkdir } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { run, parseArgs, checkBase, planSelection, estimate, mp4Info, readKey, UsageError } from './generate-videos.mjs';

const KEY = 'sk-test-secret-0123456789';
const BASE = 'https://api.example.test/v1';
import { fakeMp4 } from './mp4-fixture.mjs';
const plan = [
 { id: 'clip-pixel-night', source: 'pixel-night', motion: '像素港口动起来', prompt: 'Animate the harbor.' },
 { id: 'clip-ink-koi', source: 'ink-koi', motion: '锦鲤游动', prompt: 'Animate the koi.' },
];
function gateway({ video = fakeMp4(), statuses = ['pending', 'done'], contentUrl = '/v1/videos/req-1/content', contentType = 'video/mp4' } = {}) {
 const calls = [];
 let polls = 0;
 const json = (status, body) => new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } });
 const fetch = async (url, init = {}) => {
  const { pathname, origin } = new URL(url), auth = init.headers?.Authorization;
  calls.push({ url, method: init.method || 'GET', auth, body: init.body ? JSON.parse(init.body) : null, redirect: init.redirect });
  if (pathname === '/v1/videos/generations' && init.method === 'POST') return json(200, { request_id: 'req-1' });
  if (pathname === '/v1/videos/req-1') { const state = statuses[Math.min(polls++, statuses.length - 1)]; return json(200, state === 'done' ? { status: 'done', video: { url: contentUrl, duration: 5 } } : { status: state }); }
  if (pathname.endsWith('/content') || origin !== new URL(BASE).origin) return new Response(video, { status: 200, headers: { 'Content-Type': contentType, 'Content-Length': String(video.byteLength) } });
  return json(404, { error: { message: 'not found' } });
 };
 return { fetch, calls };
}
async function harness(argv, options = {}) {
 const root = await mkdtemp(path.join(tmpdir(), 'mofa-videos-'));
 await mkdir(path.join(root, 'assets'), { recursive: true });
 for (const [name, value] of Object.entries(options.files || {})) await writeFile(path.join(root, name), value);
 const out = [], err = [], api = options.gateway || gateway();
 const code = await run([...argv, '--root', root, '--poll', '0'], { fetch: api.fetch, sleep: async () => {}, now: () => Date.parse('2026-10-19T08:00:00Z'), env: { MOFA_API_BASE: BASE, ...(options.noKey ? {} : { MOFA_API_KEY: KEY }) }, out: line => out.push(line), err: line => err.push(line), plan });
 return { root, code, out: out.join('\n'), err: err.join('\n'), calls: api.calls, cleanup: () => rm(root, { recursive: true, force: true }) };
}
async function allText(dir) { let text = ''; for (const entry of await readdir(dir, { withFileTypes: true })) { const full = path.join(dir, entry.name); text += entry.isDirectory() ? await allText(full) : await readFile(full, 'latin1'); } return text; }

test('默认只预览：列出计划与参考价，不发请求、不读 Key、不写任何文件', async () => {
 const h = await harness([], { noKey: true });
 try {
  assert.equal(h.code, 0); assert.equal(h.calls.length, 0);
  assert.match(h.out, /本次新提交 2 段，共 10 秒；按上游官方参考价约 \$0\.50/); assert.match(h.out, /没有发送任何请求/);
  assert.deepEqual(await readdir(h.root), ['assets']); assert.deepEqual(await readdir(path.join(h.root, 'assets')), []);
 } finally { await h.cleanup(); }
});
test('提交必须显式选择片段，且需要环境变量里的 Key；参数与地址校验拒绝不安全输入', async () => {
 assert.throws(() => parseArgs(['--yes']), UsageError);
 assert.throws(() => parseArgs(['--all', '--only', 'clip-ink-koi']), UsageError);
 assert.throws(() => parseArgs(['--duration', '20']), UsageError);
 assert.throws(() => parseArgs(['--resolution', '4k']), UsageError);
 assert.throws(() => checkBase('http://ai.example.com/v1'), /HTTPS/);
 assert.throws(() => checkBase('https://user:pass@ai.example.com/v1'), UsageError);
 assert.equal(checkBase('http://127.0.0.1:4181/v1/'), 'http://127.0.0.1:4181/v1');
 assert.throws(() => planSelection({ only: ['nope'], all: false }, plan), /计划里没有：nope/);
 assert.deepEqual(planSelection({ only: ['ink-koi'], all: false }, plan).map(entry => entry.id), ['clip-ink-koi']);
 assert.equal(estimate({ model: 'grok-imagine-video', resolution: '480p', duration: 5 }, 8), 2);
 assert.equal(estimate({ model: 'unknown-model', resolution: '480p', duration: 5 }, 8), null);
 const h = await harness(['--only', 'clip-pixel-night', '--yes'], { noKey: true });
 try { assert.equal(h.code, 2); assert.equal(h.calls.length, 0); assert.match(h.err, /MOFA_API_KEY/); } finally { await h.cleanup(); }
});
test('试一段：提交图生视频、轮询到完成、下载校验 MP4，按摘要命名并写入清单与记录', async () => {
 const video = fakeMp4({ seconds: 5, width: 854, height: 480 }), h = await harness(['--only', 'clip-pixel-night', '--yes'], { gateway: gateway({ video }) });
 try {
  assert.equal(h.code, 0, h.err);
  const post = h.calls.filter(call => call.method === 'POST');
  assert.equal(post.length, 1); assert.equal(post[0].url, BASE + '/videos/generations'); assert.equal(post[0].auth, 'Bearer ' + KEY);
  assert.equal(post[0].body.model, 'grok-imagine-video'); assert.equal(post[0].body.prompt, 'Animate the harbor.');
  assert.equal(post[0].body.duration, 5); assert.equal(post[0].body.resolution, '480p'); assert.match(post[0].body.image.url, /^data:image\/webp;base64,/);
  assert.equal(h.calls.filter(call => call.url === BASE + '/videos/req-1').length, 2);
  assert.ok(h.calls.every(call => call.redirect === 'manual'));
  const file = 'assets/clip-pixel-night-' + createHash('sha256').update(video).digest('hex').slice(0, 10) + '.mp4';
  assert.deepEqual(new Uint8Array(await readFile(path.join(h.root, file))), video);
  const { clips } = await import(pathToFileURL(path.join(h.root, 'videos.mjs')).href + '?t=' + Date.now());
  assert.equal(clips.length, 1); assert.equal(clips[0].file, file); assert.equal(clips[0].duration, 5); assert.equal(clips[0].width, 854); assert.equal(clips[0].height, 480);
  assert.equal(clips[0].requestId, undefined, '清单不公开任务号');
  const provenance = JSON.parse(await readFile(path.join(h.root, 'assets', 'PROVENANCE-VIDEOS.json'), 'utf8'));
  assert.equal(provenance.clips['clip-pixel-night'].requestId, 'req-1'); assert.equal(provenance.clips['clip-pixel-night'].prompt, 'Animate the harbor.');
  assert.deepEqual(JSON.parse(await readFile(path.join(h.root, '.video-jobs.json'), 'utf8')), {});
  assert.ok(!(await allText(h.root)).includes(KEY), 'Key 不写入任何文件'); assert.ok(!h.out.includes(KEY) && !h.err.includes(KEY));
  const again = await run(['--only', 'clip-pixel-night', '--yes', '--root', h.root, '--poll', '0'], { fetch: async () => { throw new Error('不应再请求'); }, sleep: async () => {}, env: { MOFA_API_BASE: BASE, MOFA_API_KEY: KEY }, out: () => {}, err: () => {}, plan });
  assert.equal(again, 0, '已完成的片段再次运行直接跳过');
 } finally { await h.cleanup(); }
});
test('中断后再运行继续查询原任务号，不重复提交；参数不同则拒绝以免重复计费', async () => {
 const jobs = JSON.stringify({ 'clip-pixel-night': { requestId: 'req-1', model: 'grok-imagine-video', resolution: '480p', duration: 5, prompt: 'Animate the harbor.', submittedAt: '2026-10-19T07:00:00Z' } });
 const h = await harness(['--only', 'clip-pixel-night', '--yes'], { files: { '.video-jobs.json': jobs } });
 try { assert.equal(h.code, 0, h.err); assert.equal(h.calls.filter(call => call.method === 'POST').length, 0); assert.match(h.out, /继续查询已提交的任务/); } finally { await h.cleanup(); }
 const mismatch = await harness(['--only', 'clip-pixel-night', '--yes', '--resolution', '720p'], { files: { '.video-jobs.json': jobs } });
 try { assert.equal(mismatch.code, 1); assert.equal(mismatch.calls.length, 0); assert.match(mismatch.err, /避免重复计费/); } finally { await mismatch.cleanup(); }
});
test('任务失败不写成片并清掉任务号；返回非 MP4 时丢弃；超时保留任务号便于继续', async () => {
 const failed = await harness(['--only', 'clip-pixel-night', '--yes'], { gateway: gateway({ statuses: ['pending', 'failed'] }) });
 try {
  assert.equal(failed.code, 1); assert.match(failed.err, /生成失败/);
  assert.deepEqual(JSON.parse(await readFile(path.join(failed.root, '.video-jobs.json'), 'utf8')), {});
  assert.ok(!(await readdir(path.join(failed.root, 'assets'))).some(name => name.endsWith('.mp4')));
 } finally { await failed.cleanup(); }
 const notVideo = await harness(['--only', 'clip-pixel-night', '--yes'], { gateway: gateway({ video: new TextEncoder().encode('<html>error</html>'), contentType: 'text/html' }) });
 try { assert.equal(notVideo.code, 1); assert.match(notVideo.err, /不是 MP4/); assert.ok(!(await readdir(path.join(notVideo.root, 'assets'))).some(name => name.endsWith('.mp4'))); } finally { await notVideo.cleanup(); }
 const root = await mkdtemp(path.join(tmpdir(), 'mofa-videos-')), api = gateway({ statuses: ['pending'] });
 try {
  let clock = 0;
  const code = await run(['--only', 'clip-pixel-night', '--yes', '--root', root, '--poll', '0', '--timeout', '1'], { fetch: api.fetch, sleep: async () => { clock += 30000; }, now: () => clock, env: { MOFA_API_BASE: BASE, MOFA_API_KEY: KEY }, out: () => {}, err: () => {}, plan });
  assert.equal(code, 1);
  assert.equal(JSON.parse(await readFile(path.join(root, '.video-jobs.json'), 'utf8'))['clip-pixel-night'].requestId, 'req-1');
 } finally { await rm(root, { recursive: true, force: true }); }
});
test('跨域的已签名成片地址不带 Key 下载，且只接受 HTTPS', async () => {
 const signed = await harness(['--only', 'clip-pixel-night', '--yes'], { gateway: gateway({ contentUrl: 'https://cdn.example.org/signed/clip.mp4' }) });
 try {
  assert.equal(signed.code, 0, signed.err);
  const cdn = signed.calls.find(call => call.url.startsWith('https://cdn.example.org/'));
  assert.ok(cdn); assert.equal(cdn.auth, undefined);
 } finally { await signed.cleanup(); }
 const insecure = await harness(['--only', 'clip-pixel-night', '--yes'], { gateway: gateway({ contentUrl: 'http://cdn.example.org/clip.mp4' }) });
 try { assert.equal(insecure.code, 1); assert.match(insecure.err, /不是 HTTPS/); } finally { await insecure.cleanup(); }
});
test('Key 文件：仓库外文件取第一行并去掉 BOM／换行；仓库内路径与异常内容拒绝且不回显', async () => {
 const dir = await mkdtemp(path.join(tmpdir(), 'mofa-key-'));
 try {
  const file = path.join(dir, 'key.txt');
  await writeFile(file, '﻿  ' + KEY + '  \r\n\r\n');
  assert.equal(await readKey({ MOFA_API_KEY_FILE: file }), KEY);
  assert.equal(await readKey({ MOFA_API_KEY: 'env-wins-key', MOFA_API_KEY_FILE: file }), 'env-wins-key');
  assert.equal(await readKey({}), '');
  await assert.rejects(readKey({ MOFA_API_KEY_FILE: path.join(dir, 'missing.txt') }), /读不到/);
  const inside = path.join(path.dirname(fileURLToPath(import.meta.url)), 'key.txt');
  await assert.rejects(readKey({ MOFA_API_KEY_FILE: inside }), /不能放在仓库目录里/);
  await writeFile(file, '这不是 key 但是很长的一行文字');
  await assert.rejects(readKey({ MOFA_API_KEY_FILE: file }), error => /不像 Key/.test(error.message) && !error.message.includes('这不是'));
  await writeFile(file, KEY + '\n');
  const root = await mkdtemp(path.join(tmpdir(), 'mofa-videos-')), api = gateway(), out = [], err = [];
  try {
   const code = await run(['--only', 'clip-pixel-night', '--yes', '--root', root, '--poll', '0'], { fetch: api.fetch, sleep: async () => {}, env: { MOFA_API_BASE: BASE, MOFA_API_KEY_FILE: file }, out: line => out.push(line), err: line => err.push(line), plan });
   assert.equal(code, 0, err.join('\n'));
   assert.equal(api.calls.find(call => call.method === 'POST').auth, 'Bearer ' + KEY);
   assert.ok(!out.join('\n').includes(KEY) && !(await allText(root)).includes(KEY));
  } finally { await rm(root, { recursive: true, force: true }); }
 } finally { await rm(dir, { recursive: true, force: true }); }
});
test('MP4 头部解析只读 ftyp/moov，越界或非 MP4 返回空', () => {
 assert.deepEqual(mp4Info(fakeMp4({ seconds: 6, width: 832, height: 480 })), { duration: 6, width: 832, height: 480 });
 assert.equal(mp4Info(new TextEncoder().encode('not an mp4 file at all')), null);
 const truncated = fakeMp4().slice(0, 40);
 assert.deepEqual(mp4Info(truncated), { duration: 0, width: 0, height: 0 });
});
