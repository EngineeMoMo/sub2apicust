#!/usr/bin/env node
// 工坊图生短视频：在你自己的电脑上、用你自己的魔法 API Key 运行。默认只打印计划，不发请求、不计费。
// Key 只从环境变量 MOFA_API_KEY 读取，不写入文件、不打印；接口默认 https://ai.mofamilys.com/v1，可用 MOFA_API_BASE 修改。
// 成片下载后校验 MP4、按内容摘要命名，写入 assets/、videos.mjs 与 assets/PROVENANCE-VIDEOS.json。
// 已提交的任务号记在 .video-jobs.json：中断后再运行会继续查询原任务，不会重复提交（重复提交会重复计费）。
import { createHash } from 'node:crypto';
import { access, mkdir, readFile, rename, writeFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { items } from '../catalog.mjs';
import { videoPlan, VIDEO_DEFAULTS } from '../video-plan.mjs';

export class UsageError extends Error {}
const product = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
// 上游官方参考价（xAI 文档 2026-09-21，美元/秒，不含输入图片费）；本站实际按你所在分组的单价计费，以用量记录为准。
export const UPSTREAM_PRICE = Object.freeze({ 'grok-imagine-video': { '480p': 0.05, '720p': 0.07 }, 'grok-imagine-video-1.5': { '480p': 0.08, '720p': 0.14, '1080p': 0.25 } });
const MAX_VIDEO_BYTES = 50 * 1024 * 1024;
const MANIFEST_HEADER = '// 已生成的工坊短视频清单，由 tools/generate-videos.mjs 在下载并校验成片后改写；请勿手工编辑。\n// 为空时页面保持“视频待开放”，不显示任何视频入口，也不会请求视频文件。\n';
const exists = file => access(file).then(() => true, () => false);
const loadJSON = async (file, fallback) => { try { return JSON.parse(await readFile(file, 'utf8')); } catch { return fallback; } };
async function atomicWrite(file, data) { await mkdir(path.dirname(file), { recursive: true }); const temp = file + '.tmp-' + process.pid; await writeFile(temp, data); await rename(temp, file); }
const saveJSON = (file, value) => atomicWrite(file, JSON.stringify(value, null, 1) + '\n');

export function usage() {
 return ['用法（在 product-samples/magic-studio 目录下运行）：',
  '  node tools/generate-videos.mjs                                 预览计划与参考费用，不发请求',
  '  node tools/generate-videos.mjs --only clip-pixel-night --yes   先试 1 段',
  '  node tools/generate-videos.mjs --all --yes                     生成全部（已完成的自动跳过）',
  '环境变量：MOFA_API_KEY（提交时才读取，需所在分组已开通视频生成）；MOFA_API_BASE（默认 https://ai.mofamilys.com/v1）',
  '参数：--model grok-imagine-video|grok-imagine-video-1.5  --resolution 480p|720p|1080p  --duration 1-15',
  '      --image webp|png（来源图格式，默认 webp）  --force（重新生成已完成的）  --poll 秒  --timeout 分钟  --root 输出目录'].join('\n');
}
export function parseArgs(argv) {
 const options = { only: [], all: false, yes: false, force: false, help: false, model: VIDEO_DEFAULTS.model, duration: VIDEO_DEFAULTS.duration, resolution: VIDEO_DEFAULTS.resolution, image: 'webp', root: product, poll: 5, timeout: 10 };
 for (let index = 0; index < argv.length; index++) {
  const arg = argv[index], value = () => { const next = argv[++index]; if (next === undefined || next.startsWith('--')) throw new UsageError('参数 ' + arg + ' 缺少取值。'); return next; };
  if (arg === '--help' || arg === '-h') options.help = true;
  else if (arg === '--all') options.all = true;
  else if (arg === '--yes') options.yes = true;
  else if (arg === '--force') options.force = true;
  else if (arg === '--only') options.only.push(...value().split(',').map(part => part.trim()).filter(Boolean));
  else if (arg === '--model') options.model = value();
  else if (arg === '--duration') options.duration = Number(value());
  else if (arg === '--resolution') options.resolution = value();
  else if (arg === '--image') options.image = value();
  else if (arg === '--root') options.root = path.resolve(value());
  else if (arg === '--poll') options.poll = Number(value());
  else if (arg === '--timeout') options.timeout = Number(value());
  else throw new UsageError('不认识的参数：' + arg);
 }
 if (!Number.isInteger(options.duration) || options.duration < 1 || options.duration > 15) throw new UsageError('--duration 需要 1 到 15 之间的整数秒。');
 if (!['480p', '720p', '1080p'].includes(options.resolution)) throw new UsageError('--resolution 只能是 480p、720p 或 1080p。');
 if (!['webp', 'png'].includes(options.image)) throw new UsageError('--image 只能是 webp 或 png。');
 if (!/^[a-z0-9][a-z0-9.-]{2,79}$/i.test(options.model)) throw new UsageError('--model 格式不对。');
 if (!(options.poll >= 0 && options.poll <= 60)) throw new UsageError('--poll 需要 0 到 60 秒。');
 if (!(options.timeout > 0 && options.timeout <= 60)) throw new UsageError('--timeout 需要大于 0、不超过 60 分钟。');
 if (options.all && options.only.length) throw new UsageError('--all 与 --only 只能选一个。');
 if (options.yes && !options.all && !options.only.length) throw new UsageError('实际提交需要用 --only 指定片段或用 --all；建议先用 --only 试一段。');
 return options;
}
export function checkBase(raw) {
 let url;
 try { url = new URL(String(raw || '').trim()); } catch { throw new UsageError('MOFA_API_BASE 不是有效地址。'); }
 const local = ['localhost', '127.0.0.1', '[::1]'].includes(url.hostname);
 if (url.username || url.password || url.search || url.hash) throw new UsageError('MOFA_API_BASE 不能包含账号、查询参数或锚点。');
 if (url.protocol !== 'https:' && !(url.protocol === 'http:' && local)) throw new UsageError('远程接口必须使用 HTTPS，避免 Key 明文传输；本机 localhost 可用 HTTP。');
 return url.href.replace(/\/+$/, '');
}
export function planSelection(options, plan = videoPlan) {
 const wanted = options.all || !options.only.length ? plan : options.only.map(key => plan.find(entry => entry.id === key || entry.source === key) || key);
 const unknown = wanted.filter(entry => typeof entry === 'string');
 if (unknown.length) throw new UsageError('计划里没有：' + unknown.join('、') + '。可选：' + plan.map(entry => entry.id).join('、'));
 return [...new Map(wanted.map(entry => [entry.id, entry])).values()];
}
// Key 来源：优先环境变量 MOFA_API_KEY；否则读取 MOFA_API_KEY_FILE 指向的文件（第一行，去掉 BOM 与空白）。
// 文件必须在仓库之外，避免被提交或打进镜像构建上下文；读取失败只报路径问题，不回显内容。
export async function readKey(env) {
 const direct = String(env.MOFA_API_KEY || '').trim();
 if (direct) return direct;
 const file = String(env.MOFA_API_KEY_FILE || '').trim();
 if (!file) return '';
 const resolved = path.resolve(file), repo = path.resolve(product, '..', '..');
 if (resolved === repo || resolved.startsWith(repo + path.sep)) throw new Error('MOFA_API_KEY_FILE 不能放在仓库目录里，请改到仓库之外（例如用户主目录）。');
 let text;
 try { text = await readFile(resolved, 'utf8'); } catch { throw new Error('读不到 MOFA_API_KEY_FILE 指向的文件，请确认路径。'); }
 const key = text.replace(/^﻿/, '').split(/\r?\n/).map(line => line.trim()).find(Boolean) || '';
 if (key && !/^[\x21-\x7e]{8,512}$/.test(key)) throw new Error('MOFA_API_KEY_FILE 第一行不像 Key（含空格、中文或长度不对），请确认文件里只有 Key。');
 return key;
}
export function estimate(options, count) {
 const rate = UPSTREAM_PRICE[options.model]?.[options.resolution];
 return rate === undefined ? null : Math.round(rate * options.duration * count * 100) / 100;
}
// 只读 MP4 头部：确认 ftyp，并在 moov 里找 mvhd（时长）与第一条有画面的 tkhd（宽高）；越界立即停止。
export function mp4Info(input) {
 const bytes = input instanceof Uint8Array ? input : new Uint8Array(input || []);
 if (bytes.length < 16) return null;
 const view = new DataView(bytes.buffer, bytes.byteOffset, bytes.byteLength), type = offset => String.fromCharCode(...bytes.subarray(offset + 4, offset + 8));
 if (type(0) !== 'ftyp') return null;
 const info = { duration: 0, width: 0, height: 0 };
 const walk = (start, end, depth) => {
  for (let offset = start; offset + 8 <= end && depth < 6;) {
   let size = view.getUint32(offset), header = 8;
   if (size === 1) { if (offset + 16 > end) return; size = Number(view.getBigUint64(offset + 8)); header = 16; }
   else if (size === 0) size = end - offset;
   if (size < header || offset + size > end) return;
   const kind = type(offset), body = offset + header;
   if (kind === 'moov' || kind === 'trak') walk(body, offset + size, depth + 1);
   else if (kind === 'mvhd' && size >= header + 32) {
    const v1 = bytes[body] === 1, scale = view.getUint32(body + (v1 ? 20 : 12)), length = v1 ? Number(view.getBigUint64(body + 24)) : view.getUint32(body + 16);
    if (scale) info.duration = length / scale;
   } else if (kind === 'tkhd' && size >= header + 84 && !info.width) {
    const width = view.getUint32(offset + size - 8) / 65536, height = view.getUint32(offset + size - 4) / 65536;
    if (width && height) { info.width = Math.round(width); info.height = Math.round(height); }
   }
   offset += size;
  }
 };
 walk(0, bytes.length, 0);
 return info;
}

async function call(io, url, init, ms) {
 const controller = new AbortController(), timer = setTimeout(() => controller.abort(), ms);
 try { return await io.fetch(url, { ...init, signal: controller.signal, redirect: 'manual' }); }
 catch (error) { throw new Error(error?.name === 'AbortError' ? '请求超时。' : '网络请求失败：' + (error?.message || error)); }
 finally { clearTimeout(timer); }
}
const hints = { 400: '检查模型是否支持图生视频；也可以加 --image png 换用原图格式', 401: 'Key 无效或已过期', 403: '这个 Key 所在分组没有视频生成权限或额度不足', 404: '接口地址不对，或该分组没有可用的视频账号', 413: '图片太大被拒，换用默认的 --image webp', 429: '请求太频繁或额度用完，稍后再试' };
async function failure(response, what) {
 let detail = '';
 try { const text = await response.text(); try { const body = JSON.parse(text); detail = String(body?.error?.message || body?.message || body?.error || ''); } catch { detail = text; } } catch {}
 detail = detail.replace(/\s+/g, ' ').trim().slice(0, 240);
 return new Error(what + '失败（HTTP ' + response.status + '）' + (detail ? '：' + detail : '') + (hints[response.status] ? '。' + hints[response.status] + '。' : ''));
}
// 成片地址：本站相对地址或同源地址带 Key；跨域的已签名地址不带 Key，避免把 Key 发给第三方。
async function download(io, base, key, raw) {
 const origin = new URL(base).origin;
 let target = new URL(raw, base + '/');
 for (let hop = 0; hop < 4; hop++) {
  const sameOrigin = target.origin === origin;
  if (!sameOrigin && target.protocol !== 'https:') throw new Error('成片地址不是 HTTPS，已停止下载。');
  const response = await call(io, target.href, { headers: sameOrigin ? { Authorization: 'Bearer ' + key } : {} }, 120000);
  if ([301, 302, 303, 307, 308].includes(response.status) && response.headers.get('location')) { target = new URL(response.headers.get('location'), target); continue; }
  if (!response.ok) throw await failure(response, '下载成片');
  if ((Number(response.headers.get('content-length')) || 0) > MAX_VIDEO_BYTES) throw new Error('成片超过 50MB，已停止下载。');
  const bytes = new Uint8Array(await response.arrayBuffer());
  if (bytes.byteLength > MAX_VIDEO_BYTES) throw new Error('成片超过 50MB，已停止下载。');
  return bytes;
 }
 throw new Error('成片地址重定向次数过多。');
}
async function sourceImage(entry, options, catalog) {
 const item = catalog.find(candidate => candidate.id === entry.source && candidate.media === 'image' && candidate.art);
 if (!item) throw new Error('找不到来源作品 ' + entry.source + '。');
 const file = path.join(product, 'assets', item.art + '.' + options.image);
 try { return 'data:image/' + options.image + ';base64,' + (await readFile(file)).toString('base64'); }
 catch { throw new Error('读不到来源图片 ' + path.relative(product, file) + (options.image === 'png' ? '；没有原图 PNG 时改用 --image webp。' : '。')); }
}
const videoUrl = status => [status?.video?.url, status?.download_url, status?.url, status?.video_url].find(value => typeof value === 'string' && value.trim());
export async function writeManifest(root, files, provenance, plan) {
 const clips = [];
 for (const entry of plan) {
  const record = provenance.clips[entry.id];
  if (!record || !await exists(path.join(root, record.file))) continue;
  const kind = record.kind === 'motion' ? 'motion' : 'model';
  // 程序动效没有模型提示词，清单里的 prompt 用计划里的参考提示词，供用户拿去图生视频工具使用。
  clips.push({ id: record.id, kind, source: record.source, motion: record.motion, prompt: kind === 'motion' ? entry.prompt : record.prompt, file: record.file, bytes: record.bytes, sha256: record.sha256, duration: record.duration, width: record.width, height: record.height, model: record.model || null, resolution: record.resolution, ...(kind === 'motion' ? { recipe: record.recipe, recipeVersion: record.recipeVersion } : {}), generatedAt: record.generatedAt });
 }
 await atomicWrite(files.manifest, MANIFEST_HEADER + 'export const clips = ' + JSON.stringify(clips, null, 1) + ';\n');
 return clips;
}

async function generateOne(entry, context) {
 const { options, base, key, io, log, files, jobs, provenance, catalog } = context;
 let job = jobs[entry.id];
 if (job && (job.model !== options.model || job.resolution !== options.resolution || job.duration !== options.duration)) {
  throw new Error('已有未完成的任务（' + job.model + ' / ' + job.resolution + ' / ' + job.duration + ' 秒），与本次参数不同。为避免重复计费，请用原参数继续，或删除 .video-jobs.json 里的这一条后再提交。');
 }
 if (job) log('  继续查询已提交的任务，不重复提交。');
 else {
  const image = await sourceImage(entry, options, catalog);
  const response = await call(io, base + '/videos/generations', { method: 'POST', headers: { Authorization: 'Bearer ' + key, 'Content-Type': 'application/json' }, body: JSON.stringify({ model: options.model, prompt: entry.prompt, image: { url: image }, duration: options.duration, resolution: options.resolution }) }, 60000);
  if (!response.ok) throw await failure(response, '提交');
  const body = await response.json().catch(() => ({}));
  const requestId = String(body?.request_id || body?.id || '');
  if (!/^[A-Za-z0-9._:-]{1,128}$/.test(requestId)) throw new Error('提交后没有拿到有效的任务号。');
  job = { requestId, model: options.model, resolution: options.resolution, duration: options.duration, prompt: entry.prompt, submittedAt: new Date(io.now()).toISOString() };
  jobs[entry.id] = job; await saveJSON(files.jobs, jobs);
  log('  已提交，任务号 ' + requestId + '，开始等待成片。');
 }
 const deadline = io.now() + options.timeout * 60000;
 let status;
 for (;;) {
  const response = await call(io, base + '/videos/' + encodeURIComponent(job.requestId), { headers: { Authorization: 'Bearer ' + key } }, 30000);
  if (!response.ok) throw await failure(response, '查询任务');
  status = await response.json().catch(() => ({}));
  const state = String(status?.status || '').toLowerCase();
  if (state === 'done' || state === 'completed') break;
  if (state === 'failed' || state === 'expired') {
   delete jobs[entry.id]; await saveJSON(files.jobs, jobs);
   throw new Error('任务' + (state === 'failed' ? '生成失败' : '已过期') + (status?.error?.message ? '：' + String(status.error.message).slice(0, 200) : '') + '。是否计费以用量记录为准；需要时可重新提交。');
  }
  if (io.now() >= deadline) throw new Error('等待超时，任务仍在处理。稍后用同样参数再运行会继续查询，不会重复提交。');
  await io.sleep(options.poll * 1000);
 }
 const location = videoUrl(status);
 if (!location) throw new Error('任务已完成，但没有返回成片地址。');
 const bytes = await download(io, base, key, location);
 const record = await storeClip({ root: options.root, files, provenance, entry, bytes, now: io.now, fallbackDuration: Number(status?.video?.duration) || job.duration,
  fields: { kind: 'model', prompt: job.prompt || entry.prompt, model: job.model, resolution: job.resolution, requestedDuration: job.duration, requestId: job.requestId, gateway: new URL(base).origin, submittedAt: job.submittedAt, tool: 'tools/generate-videos.mjs' } });
 delete jobs[entry.id]; await saveJSON(files.jobs, jobs);
 return record;
}
// 模型成片与程序动效共用：校验 MP4、按内容摘要命名写入 assets、旧文件移到被忽略的 .video-archive、更新生成记录（清单由调用方重写）。
export async function storeClip({ root, files, provenance, entry, bytes, fields, now = Date.now, fallbackDuration = 0 }) {
 const info = mp4Info(bytes);
 if (!info) throw new Error('下载的文件不是 MP4，已丢弃。');
 const sha256 = createHash('sha256').update(bytes).digest('hex'), file = 'assets/' + entry.id + '-' + sha256.slice(0, 10) + '.mp4';
 await atomicWrite(path.join(root, file), bytes);
 const previous = provenance.clips[entry.id];
 if (previous?.file && previous.file !== file && await exists(path.join(root, previous.file))) {
  // 旧成片移出 assets，不再被打包；保留在被 Git 忽略的 .video-archive 里，需要时可找回。
  await mkdir(path.join(root, '.video-archive'), { recursive: true });
  await rename(path.join(root, previous.file), path.join(root, '.video-archive', path.basename(previous.file)));
 }
 const duration = info.duration ? Math.round(info.duration * 10) / 10 : fallbackDuration;
 provenance.clips[entry.id] = { id: entry.id, source: entry.source, motion: fields.motion || entry.motion, file, bytes: bytes.byteLength, sha256, duration, width: info.width || null, height: info.height || null, ...fields, generatedAt: new Date(now()).toISOString() };
 await saveJSON(files.provenance, provenance);
 return provenance.clips[entry.id];
}
export function clipFiles(root) { return { manifest: path.join(root, 'videos.mjs'), provenance: path.join(root, 'assets', 'PROVENANCE-VIDEOS.json'), jobs: path.join(root, '.video-jobs.json') }; }
export async function loadProvenance(file) {
 const provenance = await loadJSON(file, { note: '工坊短视频生成记录：首帧为工坊原创作品；kind=model 为视频模型成片（prompt 为实际发送原文），kind=motion 为程序动效（按 tools/motion 配方逐帧渲染，不是模型生成）。', clips: {} });
 if (!provenance.clips || typeof provenance.clips !== 'object') provenance.clips = {};
 return provenance;
}

export async function run(argv, io = {}) {
 const deps = { fetch: io.fetch || globalThis.fetch, sleep: io.sleep || (ms => new Promise(resolve => setTimeout(resolve, ms))), now: io.now || Date.now, env: io.env || process.env, out: io.out || console.log, err: io.err || console.error, plan: io.plan || videoPlan, catalog: io.catalog || items };
 let options;
 try { options = parseArgs(argv); } catch (error) { if (error instanceof UsageError) { deps.err(error.message + '\n\n' + usage()); return 2; } throw error; }
 if (options.help) { deps.out(usage()); return 0; }
 let base, selection;
 try { base = checkBase(deps.env.MOFA_API_BASE || 'https://ai.mofamilys.com/v1'); selection = planSelection(options, deps.plan); }
 catch (error) { if (error instanceof UsageError) { deps.err(error.message); return 2; } throw error; }
 const files = clipFiles(options.root), provenance = await loadProvenance(files.provenance);
 const jobs = await loadJSON(files.jobs, {});
 // 程序动效只是占位，视频模型成片会替换它；所以只有模型成片算“已完成”。
 const done = async entry => provenance.clips[entry.id]?.kind !== 'motion' && Boolean(provenance.clips[entry.id]?.file) && await exists(path.join(options.root, provenance.clips[entry.id].file));
 const pending = [];
 for (const entry of selection) if (options.force || !await done(entry)) pending.push(entry);
 const fresh = pending.filter(entry => !jobs[entry.id]).length, cost = estimate(options, fresh);
 deps.out('接口：' + base + '　模型：' + options.model + '　清晰度：' + options.resolution + '　时长：' + options.duration + ' 秒　来源图：' + options.image);
 for (const entry of selection) deps.out((pending.includes(entry) ? (jobs[entry.id] ? '  · 继续 ' : provenance.clips[entry.id]?.kind === 'motion' ? '  · 待生成（替换程序动效） ' : '  · 待生成 ') : '  · 已完成 ') + entry.id + '　' + entry.motion);
 deps.out('本次新提交 ' + fresh + ' 段，共 ' + fresh * options.duration + ' 秒' + (cost === null ? '。' : '；按上游官方参考价约 $' + cost.toFixed(2) + '（不含输入图片费）。本站实际按你所在分组的单价扣费，以用量记录为准。'));
 if (!pending.length) { if (options.yes) await writeManifest(options.root, files, provenance, deps.plan); deps.out('没有需要生成的片段。'); return 0; }
 for (const entry of pending) if (!jobs[entry.id]) await sourceImage(entry, options, deps.catalog);
 if (!options.yes) { deps.out('这只是预览，没有发送任何请求。确认后加 --yes 提交；建议先用 --only ' + pending[0].id + ' --yes 试一段。'); return 0; }
 let key;
 try { key = await readKey(deps.env); } catch (error) { deps.err(error.message); return 2; }
 if (!key) { deps.err('没有读取到 Key。请设置环境变量 MOFA_API_KEY，或用 MOFA_API_KEY_FILE 指向仓库外只含 Key 的文本文件；Key 不会被写入其他文件或打印。'); return 2; }
 const scrub = text => String(text).split(key).join('[已隐藏]'), log = text => deps.out(scrub(text));
 let failed = 0;
 for (const entry of pending) {
  log('▶ ' + entry.id);
  try { const record = await generateOne(entry, { options, base, key, io: deps, log, files, jobs, provenance, catalog: deps.catalog }); log('  ✓ ' + record.file + '（' + (record.bytes / 1024 / 1024).toFixed(2) + ' MB，' + record.duration + ' 秒' + (record.width ? '，' + record.width + '×' + record.height : '') + '）'); }
  catch (error) { failed++; deps.err(scrub('  ✗ ' + entry.id + '：' + (error?.message || error))); }
 }
 const clips = await writeManifest(options.root, files, provenance, deps.plan);
 log('清单已更新：' + clips.length + ' 段可在工坊播放' + (failed ? '，' + failed + ' 段失败' : '') + '。');
 return failed ? 1 : 0;
}

if (process.argv[1] && fileURLToPath(import.meta.url) === path.resolve(process.argv[1])) {
 run(process.argv.slice(2)).then(code => { process.exitCode = code; }, error => {
  const key = String(process.env.MOFA_API_KEY || '').trim();
  console.error((key ? String(error?.message || error).split(key).join('[已隐藏]') : String(error?.message || error)));
  process.exitCode = 1;
 });
}
