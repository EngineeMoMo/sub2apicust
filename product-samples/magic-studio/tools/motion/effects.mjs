// 程序动效：按原图像素逐帧生成局部动态（灯光明暗、水面波纹、星光闪烁、月光呼吸），主体与构图不变，不是视频模型生成。
// 纯函数，浏览器录制页与 node 测试共用。所有时间项都用“整段循环的整数周期”，第 0 帧与第 N 帧相同，可无缝循环。
const TAU = Math.PI * 2;

// 可复现的伪随机（mulberry32）：同一配方与种子得到同样的闪烁相位。
export function seeded(seed) {
 let state = seed >>> 0;
 return () => { state = (state + 0x6D2B79F5) >>> 0; let t = state; t = Math.imul(t ^ (t >>> 15), t | 1); t ^= t + Math.imul(t ^ (t >>> 7), t | 61); return ((t ^ (t >>> 14)) >>> 0) / 4294967296; };
}
export function inPolygon(x, y, polygon) {
 let inside = false;
 for (let i = 0, j = polygon.length - 1; i < polygon.length; j = i++) {
  const [xi, yi] = polygon[i], [xj, yj] = polygon[j];
  if ((yi > y) !== (yj > y) && x < (xj - xi) * (y - yi) / (yj - yi) + xi) inside = !inside;
 }
 return inside;
}
// 区域用 0–1 归一化坐标，与输出尺寸无关。
const inAny = (x, y, polygons = []) => polygons.some(polygon => inPolygon(x, y, polygon));
export const classify = {
 warm: (r, g, b) => r > 190 && g > 120 && b < 150 && r > b + 60,
 water: (r, g, b) => b > r + 10 && b >= g - 10,
 star: (r, g, b, minLum = 175) => 0.299 * r + 0.587 * g + 0.114 * b > minLum && Math.abs(r - b) < 70,
 // 水面区域里的木桩、阴影：暗且不偏蓝，保持静止。
 solid: (r, g, b) => 0.299 * r + 0.587 * g + 0.114 * b < 70 && Math.max(r, g, b) - Math.min(r, g, b) < 60 && !(b > r + 10),
};
function components(mask, width, height) {
 const label = new Int32Array(width * height).fill(-1), groups = [], stack = [];
 for (let start = 0; start < mask.length; start++) {
  if (!mask[start] || label[start] >= 0) continue;
  const id = groups.length, members = [];
  label[start] = id; stack.push(start);
  while (stack.length) {
   const p = stack.pop(), x = p % width;
   members.push(p);
   for (const q of [x > 0 ? p - 1 : -1, x < width - 1 ? p + 1 : -1, p >= width ? p - width : -1, p + width < mask.length ? p + width : -1]) {
    if (q >= 0 && mask[q] && label[q] < 0) { label[q] = id; stack.push(q); }
   }
  }
  groups.push(members);
 }
 return groups;
}
const centroid = (members, width) => { let sx = 0, sy = 0; for (const p of members) { sx += p % width; sy += Math.floor(p / width); } return [sx / members.length, sy / members.length]; };

// 预先算好每种动态影响的像素，逐帧只做查表与乘法。
export function analyzeScene(rgba, width, height, recipe) {
 const random = seeded(recipe.seed || 1), count = width * height;
 const water = new Uint8Array(count), warm = new Uint8Array(count), light = new Uint8Array(count), star = new Uint8Array(count);
 // 月亮圆盘内不当作灯或星星，只做月光呼吸。
 const inMoon = (x, y) => recipe.moon && Math.hypot(x - recipe.moon.x * width, y - recipe.moon.y * height) <= recipe.moon.radius * width;
 for (let p = 0; p < count; p++) {
  const x = p % width, y = (p - x) / width, nx = (x + 0.5) / width, ny = (y + 0.5) / height, i = p * 4, r = rgba[i], g = rgba[i + 1], b = rgba[i + 2];
  const isWarm = classify.warm(r, g, b);
  warm[p] = isWarm ? 1 : 0;
  // 水面：区域内除人工排除的物体（小船、鹈鹕等）与暗色实体（木桩、阴影）外都随波纹晃动，含各种亮度的倒影。
  if (recipe.water && inAny(nx, ny, recipe.water.regions) && !inAny(nx, ny, recipe.water.exclude) && !classify.solid(r, g, b)) water[p] = 1;
  else if (inMoon(x, y)) continue;
  else if (recipe.lights && isWarm && inAny(nx, ny, recipe.lights.regions)) light[p] = 1;
  else if (recipe.stars && !isWarm && inAny(nx, ny, recipe.stars.regions) && classify.star(r, g, b, recipe.stars.minLum)) star[p] = 1;
 }
 // 灯光：连通的暖色像素为一盏灯，各自相位；外溢几像素的柔光同步明暗（不进入水面）。
 const lightGroups = recipe.lights ? components(light, width, height) : [];
 const lightParams = lightGroups.map(members => {
  const [cx, cy] = centroid(members, width), beacon = (recipe.beacons || []).find(item => Math.hypot(cx / width - item.x, (cy / height - item.y) * height / width) <= item.radius);
  const [low, high] = recipe.lights.amplitude;
  return beacon ? { beacon: true, k1: beacon.cycles, phase: random(), amplitude: beacon.amplitude }
   : { beacon: false, k1: 2 + Math.floor(random() * 4), k2: 7 + Math.floor(random() * 7), p1: random(), p2: random(), amplitude: low + random() * (high - low) };
 });
 const radius = recipe.lights?.spill || 0, owner = new Int32Array(count).fill(-1), distance = new Uint8Array(count).fill(255);
 let frontier = [];
 lightGroups.forEach((members, id) => { for (const p of members) { owner[p] = id; distance[p] = 0; frontier.push(p); } });
 for (let d = 1; d <= radius; d++) {
  const next = [];
  for (const p of frontier) {
   const x = p % width;
   for (const q of [x > 0 ? p - 1 : -1, x < width - 1 ? p + 1 : -1, p >= width ? p - width : -1, p + width < count ? p + width : -1]) {
    if (q >= 0 && distance[q] === 255 && !water[q]) { distance[q] = d; owner[q] = owner[p]; next.push(q); }
   }
  }
  frontier = next;
 }
 const glow = [];
 for (let p = 0; p < count; p++) if (owner[p] >= 0) glow.push(p);
 const glowPixels = Int32Array.from(glow), glowOwner = Int32Array.from(glow, p => owner[p]), glowWeight = Float32Array.from(glow, p => distance[p] === 0 ? 1 : (recipe.lights.spillStrength ?? 0.5) * (1 - distance[p] / (radius + 1)) ** 2);
 // 星星：小的亮点各自闪；大的亮块按月亮处理（只轻微呼吸）。
 const starGroups = recipe.stars ? components(star, width, height).filter(members => members.length <= recipe.stars.maxSize) : [];
 const starParams = starGroups.map(() => ({ k: 1 + Math.floor(random() * 3), phase: random(), amplitude: recipe.stars.amplitude[0] + random() * (recipe.stars.amplitude[1] - recipe.stars.amplitude[0]) }));
 const moon = [];
 if (recipe.moon) {
  const cx = recipe.moon.x * width, cy = recipe.moon.y * height, r = recipe.moon.radius * width;
  for (let y = Math.max(0, Math.floor(cy - r)); y < Math.min(height, Math.ceil(cy + r)); y++) for (let x = Math.max(0, Math.floor(cx - r)); x < Math.min(width, Math.ceil(cx + r)); x++) {
   const d = Math.hypot(x - cx, y - cy), p = y * width + x;
   if (d <= r && !water[p] && owner[p] < 0) moon.push([p, (1 - d / r) ** 2]);
  }
 }
 const waterPixels = [];
 for (let p = 0; p < count; p++) if (water[p]) waterPixels.push(p);
 // 波纹在靠近水面边界（木桩、船、区域边缘）的几个像素内逐渐减弱，避免静止与晃动之间出现硬接缝。
 const edge = recipe.water?.edge || 0, taper = new Float32Array(count).fill(1);
 if (edge > 0) {
  const steps = new Uint8Array(count).fill(255);
  let ring = [];
  for (let p = 0; p < count; p++) if (!water[p]) { steps[p] = 0; ring.push(p); }
  for (let d = 1; d <= edge && ring.length; d++) {
   const next = [];
   for (const p of ring) { const x = p % width; for (const q of [x > 0 ? p - 1 : -1, x < width - 1 ? p + 1 : -1, p >= width ? p - width : -1, p + width < count ? p + width : -1]) if (q >= 0 && steps[q] === 255) { steps[q] = d; next.push(q); } }
   ring = next;
  }
  for (const p of waterPixels) if (steps[p] <= edge) taper[p] = steps[p] / (edge + 1);
 }
 return {
  width, height, recipe, water, warm, taper, waterPixels: Int32Array.from(waterPixels),
  glowPixels, glowOwner, glowWeight, lightParams,
  starPixels: starGroups.map(members => Int32Array.from(members)), starParams,
  moonPixels: Int32Array.from(moon, item => item[0]), moonWeight: Float32Array.from(moon, item => item[1]), moonPhase: random(),
  stats: { lights: lightGroups.length, beacons: lightParams.filter(item => item.beacon).length, water: waterPixels.length, stars: starGroups.length, moon: moon.length },
 };
}
const scale = (out, i, m) => { out[i] = Math.min(255, out[i] * m); out[i + 1] = Math.min(255, out[i + 1] * m); out[i + 2] = Math.min(255, out[i + 2] * m); };
// u 为循环进度 [0, 1)：u=0 与 u=1 完全相同。source 与 out 为同尺寸 RGBA。
export function renderFrame(scene, source, out, u) {
 const { width, height, recipe } = scene;
 out.set(source);
 if (recipe.water) {
  const w = recipe.water, rows = new Float32Array(height), cols = new Float32Array(width);
  for (let y = 0; y < height; y++) rows[y] = Math.sin(TAU * (y / w.wavelength - w.cycles * u));
  for (let x = 0; x < width; x++) cols[x] = 0.65 + 0.35 * Math.sin(TAU * (x / w.span + w.drift * u));
  for (const p of scene.waterPixels) {
   const x = p % width, y = (p - x) / width, dx = Math.round(w.amplitude * rows[y] * cols[x] * scene.taper[p]);
   let q = p;
   if (dx) { const sx = x + dx; if (sx >= 0 && sx < width && scene.water[y * width + sx]) q = y * width + sx; }
   const i = p * 4, j = q * 4;
   out[i] = source[j]; out[i + 1] = source[j + 1]; out[i + 2] = source[j + 2];
   if (scene.warm[q]) scale(out, i, 1 + w.shimmer * Math.sin(TAU * (y / 6 + x / 37 - w.shimmerCycles * u)));
  }
 }
 if (scene.glowPixels.length) {
  const levels = scene.lightParams.map(item => item.beacon
   ? 1 - 0.35 * item.amplitude + 1.35 * item.amplitude * (0.5 + 0.5 * Math.sin(TAU * (item.k1 * u + item.phase))) ** 4
   : 1 + item.amplitude * (0.6 * Math.sin(TAU * (item.k1 * u + item.p1)) + 0.4 * Math.sin(TAU * (item.k2 * u + item.p2))));
  for (let n = 0; n < scene.glowPixels.length; n++) scale(out, scene.glowPixels[n] * 4, 1 + (levels[scene.glowOwner[n]] - 1) * scene.glowWeight[n]);
 }
 scene.starPixels.forEach((pixels, n) => { const item = scene.starParams[n], m = 1 + item.amplitude * Math.sin(TAU * (item.k * u + item.phase)); for (const p of pixels) scale(out, p * 4, m); });
 if (scene.moonPixels.length) {
  const pulse = recipe.moon.amplitude * (0.5 + 0.5 * Math.sin(TAU * (u + scene.moonPhase)));
  for (let n = 0; n < scene.moonPixels.length; n++) scale(out, scene.moonPixels[n] * 4, 1 + pulse * scene.moonWeight[n]);
 }
 return out;
}
