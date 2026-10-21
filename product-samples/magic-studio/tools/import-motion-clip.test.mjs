import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtemp, mkdir, readFile, readdir, rm, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { pathToFileURL } from 'node:url';
import { importMotion } from './import-motion-clip.mjs';
import { run } from './generate-videos.mjs';
import { fakeMp4 } from './mp4-fixture.mjs';

const plan = [{ id: 'clip-pixel-night', source: 'pixel-night', motion: '计划里的说明', prompt: 'Reference image-to-video prompt.' }];
const recipes = { 'clip-pixel-night': { version: 3, source: 'pixel-night', art: 'pixel-night', width: 768, height: 512, fps: 30, seconds: 5, motion: '灯光、水波与星光动起来，主体不动。' } };
const meta = (extra = {}) => ({ recipe: 'clip-pixel-night', version: 3, width: 768, height: 512, fps: 30, frames: 150, seconds: 5, lateFrames: 0, userAgent: 'test', ...extra });
async function setup(video, info) {
 const dir = await mkdtemp(path.join(tmpdir(), 'mofa-motion-')), from = path.join(dir, 'staged'), root = path.join(dir, 'root');
 await mkdir(from, { recursive: true }); await mkdir(path.join(root, 'assets'), { recursive: true });
 await writeFile(path.join(from, 'clip-pixel-night.mp4'), video); await writeFile(path.join(from, 'clip-pixel-night.json'), JSON.stringify(info));
 const out = [], err = [];
 const go = argv => importMotion(['--id', 'clip-pixel-night', '--from', from, '--root', root, ...argv], { out: line => out.push(line), err: line => err.push(line), plan, recipes, now: () => Date.parse('2026-10-21T04:00:00Z') });
 return { dir, from, root, out, err, go, cleanup: () => rm(dir, { recursive: true, force: true }) };
}
test('程序动效入库：校验后按摘要写入，清单标 kind=motion，提示词用参考提示词并保留配方版本', async () => {
 const h = await setup(fakeMp4({ seconds: 5, width: 768, height: 512 }), meta());
 try {
  assert.equal(await h.go([]), 0, h.err.join('\n'));
  const files = (await readdir(path.join(h.root, 'assets'))).filter(name => name.endsWith('.mp4'));
  assert.equal(files.length, 1); assert.match(files[0], /^clip-pixel-night-[a-f0-9]{10}\.mp4$/);
  const { clips } = await import(pathToFileURL(path.join(h.root, 'videos.mjs')).href + '?t=' + Date.now());
  assert.equal(clips.length, 1); assert.equal(clips[0].kind, 'motion'); assert.equal(clips[0].model, null);
  assert.equal(clips[0].prompt, 'Reference image-to-video prompt.'); assert.equal(clips[0].motion, '灯光、水波与星光动起来，主体不动。');
  assert.equal(clips[0].recipe, 'clip-pixel-night'); assert.equal(clips[0].recipeVersion, 3); assert.equal(clips[0].resolution, '768×512');
  const record = JSON.parse(await readFile(path.join(h.root, 'assets', 'PROVENANCE-VIDEOS.json'), 'utf8')).clips['clip-pixel-night'];
  assert.equal(record.kind, 'motion'); assert.equal(record.lateFrames, 0); assert.match(record.tool, /record\.html/);
 } finally { await h.cleanup(); }
});
test('录制结果与配方不符时拒收：版本、尺寸、迟到帧、时长、非 MP4 都会拒绝且不写文件', async () => {
 for (const [video, info, message] of [
  [fakeMp4({ width: 768, height: 512 }), meta({ version: 2 }), /版本/],
  [fakeMp4({ width: 640, height: 480 }), meta(), /实际尺寸/],
  [fakeMp4({ width: 768, height: 512 }), meta({ lateFrames: 40 }), /迟到帧/],
  [fakeMp4({ width: 768, height: 512, seconds: 3 }), meta(), /时长/],
  [new TextEncoder().encode('not a video at all'), meta(), /不是 MP4/],
 ]) {
  const h = await setup(video, info);
  try {
   assert.equal(await h.go([]), 1); assert.match(h.err.join('\n'), message);
   assert.deepEqual((await readdir(path.join(h.root, 'assets'))), []);
  } finally { await h.cleanup(); }
 }
});
test('已有视频模型成片时不用程序动效覆盖；模型生成会把程序动效当作待替换', async () => {
 const h = await setup(fakeMp4({ width: 768, height: 512 }), meta());
 try {
  assert.equal(await h.go([]), 0);
  const lines = [];
  const code = await run(['--only', 'clip-pixel-night', '--root', h.root], { env: { MOFA_API_BASE: 'https://api.example.test/v1' }, out: line => lines.push(line), err: line => lines.push(line), plan, fetch: async () => { throw new Error('预览不应请求'); } });
  assert.equal(code, 0); assert.match(lines.join('\n'), /待生成（替换程序动效） clip-pixel-night/);
  const provenance = path.join(h.root, 'assets', 'PROVENANCE-VIDEOS.json'), data = JSON.parse(await readFile(provenance, 'utf8'));
  data.clips['clip-pixel-night'].kind = 'model'; await writeFile(provenance, JSON.stringify(data));
  assert.equal(await h.go([]), 1); assert.match(h.err.join('\n'), /已有视频模型成片/);
 } finally { await h.cleanup(); }
});
