export const STORE_KEY = 'mofa-studio:v1';
export const VIDEO_CONTENT_ENABLED = false;
export function isItemAvailable(item) {
 // 由工坊图片生成、已在清单里的短视频单独开放；封存的分镜与旧视频仍受 VIDEO_CONTENT_ENABLED 控制。
 return Boolean(item) && (VIDEO_CONTENT_ENABLED || item.media !== 'video' || item.type !== 'prompt' || Boolean(item.clip));
}
const formatBytes = bytes => bytes >= 1024 * 1024 ? (bytes / 1024 / 1024).toFixed(1) + ' MB' : Math.max(1, Math.round(bytes / 1024)) + ' KB';
export { formatBytes };
// 把清单里的一段短视频变成可浏览条目：首帧沿用来源图片。kind=model 是视频模型成片，提示词为实际发送原文；
// kind=motion 是程序动效（按原图逐帧渲染局部动态，主体不动），提示词只是给图生视频工具的参考，不能说成生成记录。
export function clipItem(clip, source) {
 if (!clip || !source || !/^clip-[a-z0-9-]+$/.test(clip.id) || !/^assets\/clip-[a-z0-9-]+-[a-f0-9]{10}\.mp4$/.test(clip.file)) return null;
 const seconds = Math.round(Number(clip.duration) || 0), size = formatBytes(Number(clip.bytes) || 0), motion = clip.kind === 'motion';
 const common = {
  id: clip.id, clip: { file: clip.file, bytes: Number(clip.bytes) || 0, duration: seconds, resolution: clip.resolution, sha256: clip.sha256, kind: motion ? 'motion' : 'model', loop: motion },
  category: source.category, styles: source.styles || [], ratio: source.ratio, art: source.art, media: 'video', type: 'prompt',
  summary: clip.motion, fields: [], zh: clip.prompt, en: clip.prompt,
  steps: ['在支持“图生视频”的工具里上传来源作品图片作为首帧。', '粘贴下方提示词，时长设为 ' + seconds + ' 秒左右。', '对照原图检查构图、配色和主体有没有变形，再决定是否重试。'],
  tip: '视频只在点击播放时下载一次，之后保存在当前浏览器，再看不会重复下载；可在列表上方清除。',
  source: { label: '来源作品：' + source.title, url: '#' + new URLSearchParams({ item: source.id }).toString() },
  sourceId: source.id,
 };
 if (motion) return {
  ...common, title: source.title + ' · 动态版', tool: '程序动效', tags: [...(source.tags || []), '短视频', '程序动效'],
  status: '程序动效 · 主体不动', previewLabel: '程序动效 · ' + seconds + ' 秒', motionClip: true,
  recordNote: '这段动态是按原图「' + source.title + '」逐帧计算的灯光、水波与星光，不是视频模型生成，人物和主体不会动。下方是给图生视频工具的参考提示词，并未用于这段视频。',
  requirements: '在浏览器里按配方逐帧渲染并录制，' + clip.resolution + '，约 ' + size + '。参考提示词需要目标工具支持以图片作首帧。',
  generation: { tool: '程序动效', date: String(clip.generatedAt || '').slice(0, 10), framing: clip.resolution + ' · ' + seconds + ' 秒 · 循环播放', notes: '只改变灯光、水面与星光等局部，构图与主体保持原图。' },
 };
 return {
  ...common, title: source.title + ' · 动起来', tool: '视频生成', tags: [...(source.tags || []), '短视频', '图生视频'],
  status: '已生成 · 效果因模型而异', previewLabel: '图生视频 · ' + seconds + ' 秒',
  rawPrompt: clip.prompt, rawPromptLabel: '这段视频实际使用的提示词',
  recordNote: '这是生成这段视频时实际发送的提示词，首帧是来源作品「' + source.title + '」；同一提示词再次生成也会有差异。',
  requirements: '由 ' + clip.model + ' 按 ' + clip.resolution + ' 生成，约 ' + size + '。需要目标工具支持以图片作首帧；不同模型的动作幅度与稳定性不同。',
  generation: { tool: clip.model, date: String(clip.generatedAt || '').slice(0, 10), framing: clip.resolution + ' · ' + seconds + ' 秒 · 以来源作品为首帧', notes: '未做逐帧人工修饰。' },
 };
}
const normal = value => String(value || '').normalize('NFKC').toLowerCase().trim();
export function findItems(items, state, saved) {
 let result = items.filter(item => {
  if (!isItemAvailable(item)) return false;
  if (state.section === 'skills' && item.type === 'prompt') return false;
  if (state.section === 'discover' && item.type !== 'prompt') return false;
  if (state.section === 'saved' && !saved.favorites.includes(item.id)) return false;
  if (state.section === 'history' && !saved.history.includes(item.id)) return false;
  if (state.media !== 'all' && item.media !== state.media) return false;
  if (state.type !== 'all' && item.type !== state.type) return false;
  if (state.category !== 'all' && item.category !== state.category) return false;
  if (state.style && state.style !== 'all' && !item.styles?.includes(state.style)) return false;
  const hay = normal([item.title, item.summary, item.tool, item.category, item.zh, item.en, ...item.tags, ...(item.styles || [])].join(' '));
  return normal(state.query).split(/\s+/).every(part => hay.includes(part));
 });
 if (state.section === 'history') result.sort((a,b) => saved.history.indexOf(a.id) - saved.history.indexOf(b.id));
 else if (state.sort === 'title') result.sort((a,b) => a.title.localeCompare(b.title, 'zh-CN'));
 else if (state.sort === 'reverse') result.reverse();
 return result;
}
export function cleanSaved(raw, ids) {
 const allow = new Set(ids);
 const clean = list => Array.isArray(list) ? [...new Set(list.filter(x => typeof x === 'string' && (allow.has(x) || /^community-[a-f0-9]{32}$/.test(x))))].slice(0,100) : [];
 return { favorites: clean(raw?.favorites), history: clean(raw?.history), theme: raw?.theme === 'dark' ? 'dark' : 'light' };
}
export function makeStore(storage, ids) {
 let data = cleanSaved(null, ids), available = true;
 try { data = cleanSaved(JSON.parse(storage.getItem(STORE_KEY) || 'null'), ids); } catch { available = false; }
 return {
  get data() { return data; }, get available() { return available; },
  write(next) { data = cleanSaved(next, ids); try { storage.setItem(STORE_KEY, JSON.stringify(data)); } catch { available = false; } return available; },
  refresh(raw) { try { data = cleanSaved(JSON.parse(raw), ids); } catch { /* 保留当前状态 */ } return data; }
 };
}
export function compile(item, values, language) {
 const defaults = Object.fromEntries(item.fields.map(f => [f.key, f.value]));
 return item[language === 'en' ? 'en' : 'zh'].replace(/\{\{(\w+)\}\}/g, (match,key) => {
  if (!Object.hasOwn(defaults,key)) return match;
  const value = typeof values[key] === 'string' ? values[key].trim() : '';
  return value || defaults[key] || '';
 });
}
export function exportText(item, prompt) {
 const record = item.rawPrompt ? `\n原始提示词（实际生成记录，不含替换材料）\n${item.rawPrompt}\n` : item.community ? `\n${item.recordNote}\n` : item.media === 'image' && item.type === 'prompt' ? '\n原始提示词：未保存完整记录，仅提供复用模板。\n' : '';
 return `${item.title}\n\n复用模板（当前材料）\n${prompt}\n${record}\n使用条件\n${item.requirements}\n\n来源\n${item.source.label}\n${item.source.url}\n\n验证状态：${item.status}\n`;
}
export function communityItem(record) {
 if (record?.status && record.status !== 'published') return null;
 if (!record || !/^[a-f0-9]{32}$/.test(record.id) || !['image','video'].includes(record.media) ||
  !['actual','reference'].includes(record.prompt_kind) || record.media_url !== `/api/v1/studio/gallery/${record.id}/media`) return null;
 for (const [key, minimum, maximum] of [['title',2,80],['author',1,40],['category',1,40],['style',1,40],['prompt',10,12000],['model',1,80],['notes',0,1000]]) {
  if (typeof record[key] !== 'string' || [...record[key]].length < minimum || [...record[key]].length > maximum) return null;
 }
 const actual = record.prompt_kind === 'actual';
 return {
  id:'community-'+record.id, title:record.title, category:record.category, styles:[record.style],
  media:record.media, type:'prompt', ratio:'portrait', tags:[record.category,record.style,'用户投稿'],
  art:record.media === 'image' ? record.media_url : undefined, video:record.media === 'video' ? record.media_url : undefined,
  fields:[], zh:record.prompt, en:record.prompt, rawPrompt:actual ? record.prompt : undefined,
  recordNote:actual ? '投稿人提供的实际提示词；人工审核不等于模型效果验证。' : '投稿人提供的反推或整理参考模板，不是原始生成记录。',
  tool:record.model, status:'用户投稿 · 人工审核', community:true,
  summary:record.author+' · '+record.model+(record.notes ? ' · '+record.notes : ''),
  credit:record.author, license:'投稿人授权本站展示，不含素材商用许可',
  requirements:'提示词与制作工具由投稿人提供；审核通过不保证可重现效果，不自动授予素材商用或肖像使用权。',
  tip:'先检查目标模型支持的参数，制作自己的原创作品。',
  source:{label:'用户授权展示的作品',url:record.media_url},
  steps:['查看完整作品与投稿人提供的提示词。','检查目标工具与参数，自己补充所需参考素材。','在自己的工具中创作；当前页不调用模型。']
 };
}
export function masonryPositions(heights, columns, gap) {
 const columnCount = Math.max(1, Math.floor(columns) || 1);
 const spacing = Math.max(0, Number(gap) || 0);
 const bottoms = Array(columnCount).fill(0);
 return heights.map(height => {
  const column = bottoms.indexOf(Math.min(...bottoms));
  const row = Math.ceil(bottoms[column]) + 1;
  const span = Math.max(1, Math.ceil(height) || 1);
  bottoms[column] = row - 1 + span + spacing;
  return { column: column + 1, row, span };
 });
}
export function itemFromHash(hash, items) {
 const id = new URLSearchParams(hash.replace(/^#/, '')).get('item');
 return items.find(x => x.id === id) || null;
}
