export const STORE_KEY = 'mofa-studio:v1';
const normal = value => String(value || '').normalize('NFKC').toLowerCase().trim();
export function findItems(items, state, saved) {
 let result = items.filter(item => {
  if (state.section === 'skills' && item.type === 'prompt') return false;
  if (state.section === 'discover' && item.type !== 'prompt') return false;
  if (state.section === 'saved' && !saved.favorites.includes(item.id)) return false;
  if (state.section === 'history' && !saved.history.includes(item.id)) return false;
  if (state.media !== 'all' && item.media !== state.media) return false;
  if (state.type !== 'all' && item.type !== state.type) return false;
  if (state.category !== 'all' && item.category !== state.category) return false;
  const hay = normal([item.title, item.summary, item.tool, item.category, item.zh, item.en, ...item.tags].join(' '));
  return normal(state.query).split(/\s+/).every(part => hay.includes(part));
 });
 if (state.section === 'history') result.sort((a,b) => saved.history.indexOf(a.id) - saved.history.indexOf(b.id));
 else if (state.sort === 'title') result.sort((a,b) => a.title.localeCompare(b.title, 'zh-CN'));
 else if (state.sort === 'reverse') result.reverse();
 return result;
}
export function cleanSaved(raw, ids) {
 const allow = new Set(ids);
 const clean = list => Array.isArray(list) ? [...new Set(list.filter(x => typeof x === 'string' && allow.has(x)))].slice(0,100) : [];
 return { favorites: clean(raw?.favorites), history: clean(raw?.history), theme: raw?.theme === 'light' ? 'light' : 'dark' };
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
 return item[language === 'en' ? 'en' : 'zh'].replace(/\{\{(\w+)\}\}/g, (_,key) => {
  const value = typeof values[key] === 'string' ? values[key].trim() : '';
  return value || defaults[key] || '';
 });
}
export function exportText(item, prompt) {
 return `${item.title}\n\n${prompt}\n\n使用条件\n${item.requirements}\n\n来源\n${item.source.label}\n${item.source.url}\n\n验证状态：${item.status}\n`;
}
export function itemFromHash(hash, items) {
 const id = new URLSearchParams(hash.replace(/^#/, '')).get('item');
 return items.find(x => x.id === id) || null;
}
