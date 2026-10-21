import test from 'node:test';
import assert from 'node:assert/strict';
import { analyzeScene, renderFrame, inPolygon, classify, seeded } from './effects.mjs';
import { motionRecipes } from './recipes.mjs';
import { videoPlan } from '../../video-plan.mjs';
import { items } from '../../catalog.mjs';

const W = 64, H = 40;
const rect = (x1, y1, x2, y2) => [[x1, y1], [x2, y1], [x2, y2], [x1, y2]];
// 合成场景：上半夜空（星星、月亮、两盏暖灯、一盏灯塔），下半水面（暗木桩、暖色倒影条），左下角在所有区域之外。
function scene() {
 const px = new Uint8ClampedArray(W * H * 4);
 const set = (x, y, [r, g, b]) => { const i = (y * W + x) * 4; px[i] = r; px[i + 1] = g; px[i + 2] = b; px[i + 3] = 255; };
 for (let y = 0; y < H; y++) for (let x = 0; x < W; x++) set(x, y, y < 20 ? [18, 22, 60] : [20, 40, 90]);
 for (const [x, y] of [[5, 3], [20, 5], [40, 2]]) set(x, y, [230, 230, 240]);
 for (let y = 2; y < 8; y++) for (let x = 50; x < 56; x++) set(x, y, [235, 230, 200]);
 set(52, 4, [240, 170, 90]);
 for (const [x0, y0] of [[10, 12], [30, 12], [60, 10]]) for (let y = y0; y < y0 + 3; y++) for (let x = x0; x < x0 + 3; x++) set(x, y, [250, 190, 80]);
 for (let y = 24; y < 36; y++) for (let x = 30; x < 33; x++) set(x, y, [40, 32, 28]);
 for (let y = 22; y < 38; y += 2) for (let x = 10; x < 14; x++) set(x, y, [200, 140, 60]);
 return px;
}
const recipe = {
 version: 1, source: 'synthetic', width: W, height: H, fps: 30, seconds: 5, seed: 7,
 water: { regions: [rect(0.1, 0.52, 1, 1)], exclude: [], edge: 2, amplitude: 2, wavelength: 4, cycles: 4, span: 30, drift: 1, shimmer: 0.14, shimmerCycles: 5 },
 lights: { regions: [rect(0, 0, 1, 0.5)], amplitude: [0.08, 0.18], spill: 2, spillStrength: 0.5 },
 beacons: [{ x: 61.5 / 64, y: 11.5 / 40, radius: 0.05, cycles: 2, amplitude: 0.45 }],
 stars: { regions: [rect(0, 0, 0.75, 0.3)], minLum: 150, maxSize: 4, amplitude: [0.3, 0.55] },
 moon: { x: 53 / 64, y: 5 / 40, radius: 4 / 64, amplitude: 0.06 },
};
const frame = (s, src, u) => renderFrame(s, src, new Uint8ClampedArray(src.length), u);
const maxDiff = (a, b) => { let m = 0; for (let i = 0; i < a.length; i++) m = Math.max(m, Math.abs(a[i] - b[i])); return m; };

test('场景分析：暖灯与灯塔分组，月亮圆盘不当作灯或星星，暗木桩不算水面', () => {
 const src = scene(), s = analyzeScene(src, W, H, recipe);
 assert.equal(s.stats.lights, 3); assert.equal(s.stats.beacons, 1); assert.equal(s.stats.stars, 3);
 assert.ok(s.stats.moon > 0);
 const post = 30 * W + 31;
 assert.equal(s.water[post], 0); assert.equal(s.water[30 * W + 20], 1); assert.equal(s.water[30 * W + 2], 0, '区域外');
 assert.ok(s.taper[30 * W + 29] < s.taper[30 * W + 20], '靠近木桩的波纹减弱');
 assert.ok(classify.solid(40, 32, 28)); assert.ok(!classify.solid(20, 40, 90)); assert.ok(inPolygon(0.5, 0.5, rect(0, 0, 1, 1)));
 const a = seeded(3), b = seeded(3); assert.equal(a(), b());
});
test('渲染：第 0 帧与第 N 帧一致可无缝循环；同一种子结果相同；木桩与区域外像素始终不动', () => {
 const src = scene(), s = analyzeScene(src, W, H, recipe);
 assert.ok(maxDiff(frame(s, src, 0), frame(s, src, 1)) <= 1);
 assert.deepEqual(frame(s, src, 0.37), frame(analyzeScene(src, W, H, recipe), src, 0.37));
 let moved = false;
 for (const u of [0.1, 0.25, 0.5, 0.75, 0.9]) {
  const out = frame(s, src, u);
  for (const p of [30 * W + 31, 25 * W + 30, 38 * W + 2, 35 * W + 1]) for (let c = 0; c < 4; c++) assert.equal(out[p * 4 + c], src[p * 4 + c], 'p=' + p + ' u=' + u);
  if (maxDiff(out, src) > 0) moved = true;
 }
 assert.ok(moved, '有动态');
});
test('动态幅度：灯光柔和起伏、灯塔脉冲更强，水面位移不超过振幅', () => {
 const src = scene(), s = analyzeScene(src, W, H, recipe);
 const lamp = (12 * W + 31) * 4, beacon = (11 * W + 61) * 4, values = { lamp: [], beacon: [] };
 for (let n = 0; n < 60; n++) { const out = frame(s, src, n / 60); values.lamp.push(out[lamp + 2]); values.beacon.push(out[beacon + 2]); }
 const range = list => Math.max(...list) - Math.min(...list);
 assert.ok(range(values.lamp) > 0 && range(values.lamp) < 60, '暖灯起伏 ' + range(values.lamp));
 assert.ok(range(values.beacon) > range(values.lamp), '灯塔脉冲更明显');
 const out = frame(s, src, 0.2);
 for (const p of s.waterPixels) {
  const i = p * 4, x = p % W, y = (p - x) / W;
  let found = false;
  for (let dx = -recipe.water.amplitude; dx <= recipe.water.amplitude && !found; dx++) {
   const q = (y * W + x + dx) * 4;
   if (x + dx >= 0 && x + dx < W && Math.abs(out[i] - src[q]) <= Math.ceil(src[q] * 0.15) + 1) found = true;
  }
  assert.ok(found, '水面像素只取自振幅内的同行水面');
 }
});
test('配方与计划一致：来源、编号与动效说明都对得上，输出尺寸是 16 的倍数', () => {
 for (const [id, item] of Object.entries(motionRecipes)) {
  const plan = videoPlan.find(entry => entry.id === id);
  assert.ok(plan, id); assert.equal(item.source, plan.source);
  assert.equal(items.find(entry => entry.id === item.source && entry.media === 'image')?.art, item.art, id + ' 的原图文件名');
  assert.equal(item.width % 16, 0); assert.equal(item.height % 16, 0); assert.ok(item.fps >= 24 && item.seconds >= 3 && item.seconds <= 8);
  assert.ok(item.motion.length > 10); assert.ok(Number.isInteger(item.version));
 }
});
