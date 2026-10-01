export const STORE_KEY = 'mofa-studio:v1';
export const VIDEO_CONTENT_ENABLED = false;
export function isItemAvailable(item) {
 return Boolean(item) && (VIDEO_CONTENT_ENABLED || item.media !== 'video' || item.type !== 'prompt');
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
