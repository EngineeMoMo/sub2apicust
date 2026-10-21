#!/usr/bin/env node
// 把 tools/motion/record.html 录好的程序动效 MP4 入库：核对配方版本、尺寸、时长与迟到帧，按内容摘要写入 assets，
// 并更新 videos.mjs 与 assets/PROVENANCE-VIDEOS.json（kind=motion）。不联网、不读取任何 Key。
// 用法：node tools/import-motion-clip.mjs --id clip-pixel-night [--from fixtures/.video-run/motion] [--root .] [--max-late 0.1]
import { readFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { videoPlan } from '../video-plan.mjs';
import { motionRecipes } from './motion/recipes.mjs';
import { storeClip, writeManifest, clipFiles, loadProvenance, mp4Info, UsageError } from './generate-videos.mjs';

const product = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
export function parseImportArgs(argv) {
 const options = { id: '', from: path.join(product, 'fixtures', '.video-run', 'motion'), root: product, maxLate: 0.1 };
 for (let index = 0; index < argv.length; index++) {
  const arg = argv[index], value = () => { const next = argv[++index]; if (next === undefined || next.startsWith('--')) throw new UsageError('参数 ' + arg + ' 缺少取值。'); return next; };
  if (arg === '--id') options.id = value();
  else if (arg === '--from') options.from = path.resolve(value());
  else if (arg === '--root') options.root = path.resolve(value());
  else if (arg === '--max-late') options.maxLate = Number(value());
  else throw new UsageError('不认识的参数：' + arg);
 }
 if (!/^clip-[a-z0-9-]+$/.test(options.id)) throw new UsageError('需要 --id，例如 --id clip-pixel-night。');
 if (!(options.maxLate >= 0 && options.maxLate <= 1)) throw new UsageError('--max-late 需要 0 到 1 之间。');
 return options;
}
export async function importMotion(argv, io = {}) {
 const out = io.out || console.log, err = io.err || console.error, plan = io.plan || videoPlan, recipes = io.recipes || motionRecipes, now = io.now || Date.now;
 let options;
 try { options = parseImportArgs(argv); } catch (error) { if (error instanceof UsageError) { err(error.message); return 2; } throw error; }
 const entry = plan.find(item => item.id === options.id), recipe = recipes[options.id];
 if (!entry || !recipe) { err('计划或动效配方里没有 ' + options.id + '。'); return 2; }
 let bytes, meta;
 try { bytes = new Uint8Array(await readFile(path.join(options.from, options.id + '.mp4'))); meta = JSON.parse(await readFile(path.join(options.from, options.id + '.json'), 'utf8')); }
 catch { err('读不到录制结果：' + path.join(options.from, options.id) + '.mp4／.json。请先在录制页录制。'); return 1; }
 const info = mp4Info(bytes), problems = [];
 if (!info) problems.push('不是 MP4');
 if (meta.recipe !== options.id) problems.push('说明里的配方是 ' + meta.recipe);
 if (meta.version !== recipe.version) problems.push('录制时配方版本 ' + meta.version + '，当前 ' + recipe.version + '，请重新录制');
 if (meta.width !== recipe.width || meta.height !== recipe.height) problems.push('尺寸 ' + meta.width + '×' + meta.height + ' 与配方不符');
 if (info && info.width && (info.width !== recipe.width || info.height !== recipe.height)) problems.push('视频实际尺寸 ' + info.width + '×' + info.height + ' 与配方不符');
 if (!(meta.frames > 0) || meta.lateFrames / meta.frames > options.maxLate) problems.push('迟到帧 ' + meta.lateFrames + '／' + meta.frames + ' 超过上限，录制时窗格可能不可见');
 const duration = info?.duration || meta.seconds;
 if (Math.abs(duration - recipe.seconds) > 0.6) problems.push('时长 ' + duration.toFixed(2) + ' 秒与配方 ' + recipe.seconds + ' 秒相差太多');
 if (problems.length) { err('拒绝入库：' + problems.join('；') + '。'); return 1; }
 const files = clipFiles(options.root), provenance = await loadProvenance(files.provenance);
 if (provenance.clips[options.id] && provenance.clips[options.id].kind !== 'motion') { err(options.id + ' 已有视频模型成片，不用程序动效覆盖。'); return 1; }
 const record = await storeClip({ root: options.root, files, provenance, entry, bytes, now, fallbackDuration: meta.seconds,
  fields: { kind: 'motion', motion: recipe.motion, model: null, resolution: recipe.width + '×' + recipe.height, recipe: options.id, recipeVersion: recipe.version, fps: recipe.fps, recordedWith: String(meta.userAgent || '').slice(0, 160), lateFrames: meta.lateFrames, tool: 'tools/motion/record.html + tools/import-motion-clip.mjs' } });
 const clips = await writeManifest(options.root, files, provenance, plan);
 out('已入库 ' + record.file + '（' + (record.bytes / 1024).toFixed(0) + ' KB，' + record.duration + ' 秒，' + record.width + '×' + record.height + '，程序动效）；清单共 ' + clips.length + ' 段。');
 return 0;
}
if (process.argv[1] && fileURLToPath(import.meta.url) === path.resolve(process.argv[1])) {
 importMotion(process.argv.slice(2)).then(code => { process.exitCode = code; }, error => { console.error(error?.message || error); process.exitCode = 1; });
}
