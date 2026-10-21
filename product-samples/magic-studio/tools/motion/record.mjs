// 程序动效录制页逻辑：读取仓库里的原图 PNG，按配方逐帧渲染，用 MediaRecorder 录成 H.264 MP4，交给本机夹具暂存。
// 录制按真实时间推进；窗格不可见时浏览器会降低计时精度，页面会统计迟到帧并写进说明，入库工具据此拒收。
import { analyzeScene, renderFrame } from './effects.mjs';
import { motionRecipes } from './recipes.mjs';

const $ = id => document.getElementById(id);
let current = null, previewing = false;

export async function loadRecipe(id) {
 const recipe = motionRecipes[id];
 if (!recipe) throw new Error('没有这个配方：' + id);
 const image = new Image(); image.src = './source/' + recipe.art + '.png'; await image.decode();
 const canvas = $('stage'); canvas.width = recipe.width; canvas.height = recipe.height;
 const context = canvas.getContext('2d', { willReadFrequently: true });
 context.imageSmoothingQuality = 'high'; context.drawImage(image, 0, 0, recipe.width, recipe.height);
 const source = context.getImageData(0, 0, recipe.width, recipe.height);
 const scene = analyzeScene(source.data, recipe.width, recipe.height, recipe);
 current = { id, recipe, context, source, scene, out: new ImageData(recipe.width, recipe.height) };
 $('info').textContent = JSON.stringify({ id, version: recipe.version, size: recipe.width + '×' + recipe.height, fps: recipe.fps, seconds: recipe.seconds, ...scene.stats }, null, 1);
 return current;
}
function draw(u) { renderFrame(current.scene, current.source.data, current.out.data, u); current.context.putImageData(current.out, 0, 0); }
function preview() {
 if (previewing) { previewing = false; $('preview').textContent = '预览循环'; return; }
 previewing = true; $('preview').textContent = '停止预览';
 const start = performance.now(), period = current.recipe.seconds * 1000;
 const tick = now => { if (!previewing) return; draw(((now - start) % period) / period); requestAnimationFrame(tick); };
 requestAnimationFrame(tick);
}
// 返回 { blob, meta }；meta 记录配方版本、帧数、迟到帧和实际耗时，供入库核对。
export async function recordClip() {
 const { recipe } = current, frames = recipe.fps * recipe.seconds, interval = 1000 / recipe.fps;
 const type = ['video/mp4;codecs=avc1.42E01E', 'video/mp4;codecs=avc1', 'video/mp4'].find(item => MediaRecorder.isTypeSupported(item));
 if (!type) throw new Error('当前浏览器不能录制 MP4。');
 previewing = false; draw(0);
 const stream = $('stage').captureStream(0), track = stream.getVideoTracks()[0];
 const recorder = new MediaRecorder(stream, { mimeType: type, videoBitsPerSecond: recipe.bitrate });
 const chunks = []; recorder.ondataavailable = event => { if (event.data.size) chunks.push(event.data); };
 const stopped = new Promise(resolve => { recorder.onstop = resolve; });
 let late = 0, worst = 0;
 recorder.start(); track.requestFrame();
 const started = performance.now();
 for (let n = 1; n <= frames; n++) {
  const due = started + n * interval;
  await new Promise(resolve => setTimeout(resolve, Math.max(0, due - performance.now() - 2)));
  while (performance.now() < due) await new Promise(resolve => setTimeout(resolve, 0));
  const lag = performance.now() - due; if (lag > interval / 2) late++; worst = Math.max(worst, lag);
  if (n < frames) { draw(n / frames); track.requestFrame(); }
 }
 recorder.stop(); await stopped; track.stop();
 const blob = new Blob(chunks, { type: 'video/mp4' });
 return { blob, meta: { recipe: current.id, version: recipe.version, width: recipe.width, height: recipe.height, fps: recipe.fps, frames, seconds: recipe.seconds, mimeType: type, lateFrames: late, worstLagMs: Math.round(worst), elapsedMs: Math.round(performance.now() - started) } };
}
export async function recordAndSave() {
 const { blob, meta } = await recordClip();
 const response = await fetch('/fixture/motion-clip?id=' + encodeURIComponent(meta.recipe) + '&meta=' + encodeURIComponent(JSON.stringify(meta)), { method: 'POST', body: blob });
 if (!response.ok) throw new Error('保存失败（' + response.status + '）：' + await response.text());
 return { ...meta, bytes: blob.size };
}
async function boot() {
 const select = $('recipe');
 for (const id of Object.keys(motionRecipes)) select.append(new Option(id, id));
 const wanted = new URLSearchParams(location.search).get('recipe');
 if (wanted && motionRecipes[wanted]) select.value = wanted;
 await loadRecipe(select.value);
 select.addEventListener('change', () => { previewing = false; loadRecipe(select.value).catch(error => { $('status').textContent = error.message; }); });
 $('preview').addEventListener('click', preview);
 $('record').addEventListener('click', async () => {
  $('record').disabled = true; $('status').textContent = '录制中，请保持窗格可见…';
  try { const result = await recordAndSave(); $('status').textContent = '已保存 ' + (result.bytes / 1024).toFixed(0) + ' KB，迟到帧 ' + result.lateFrames; }
  catch (error) { $('status').textContent = error.message; }
  finally { $('record').disabled = false; }
 });
}
window.__motion = { loadRecipe, recordClip, recordAndSave, draw: u => draw(u), get current() { return current; } };
boot().catch(error => { $('status').textContent = error.message; });
