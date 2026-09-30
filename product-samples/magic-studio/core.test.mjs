import test from 'node:test';
import assert from 'node:assert/strict';
import { items, plays } from './catalog.mjs';
import { findItems, cleanSaved, makeStore, compile, exportText, itemFromHash, STORE_KEY } from './core.mjs';
import { createAppServer } from './serve.mjs';
const ids = items.map(x => x.id);
const state = { section:'discover', media:'all', type:'all', category:'all', query:'', sort:'featured' };
const saved = { favorites:[], history:[] };

test('目录完整：16条唯一资源，所有玩法指向已有条目', () => {
 assert.equal(new Set(ids).size, 16);
 assert.equal(items.filter(x => x.media === 'image' && x.type === 'prompt').length, 8);
 assert.equal(items.filter(x => x.media === 'video' && x.type === 'prompt').length, 4);
 assert.equal(items.filter(x => x.type !== 'prompt').length, 4);
 assert.equal(plays.length, 4);
 for (const play of plays) for (const step of play.steps) assert.ok(ids.includes(step.id));
 for (const item of items) {
  assert.ok(item.source.url.startsWith('https://'));
  for (const language of ['zh','en']) assert.doesNotMatch(compile(item, {}, language), /\{\{\w+\}\}/);
 }
});
test('媒介、资源类型与用途可以同时筛选', () => {
 const result = findItems(items, {...state,section:'skills',media:'video',type:'skill'}, saved);
 assert.equal(result.length, 2);
 assert.ok(result.every(x => x.media === 'video' && x.type === 'skill'));
 assert.equal(findItems(items, {...state,media:'image',category:'电商产品'}, saved).length, 2);
});
test('多关键词搜索同时匹配中文、英文、工具与标签', () => {
 assert.equal(findItems(items, {...state,query:'产品 玻璃'}, saved)[0].id, 'jade-product');
 assert.equal(findItems(items, {...state,section:'skills',query:'ＲＥＭＯＴＩＯＮ'}, saved).length, 2);
 assert.equal(findItems(items, {...state,query:'确实不存在的名字'}, saved).length, 0);
});
test('收藏包含跨媒介资源，浏览记录按最新顺序显示', () => {
 const selected = {favorites:['remotion-caption','jade-product'],history:['cloud-flight','jade-product']};
 assert.equal(findItems(items, {...state,section:'saved'}, selected).length, 2);
 assert.deepEqual(findItems(items, {...state,section:'history'}, selected).map(x=>x.id), selected.history);
 assert.equal(findItems(items, {...state,section:'saved',media:'image'}, selected).length, 1);
});
test('存储只保留白名单ID与主题，不保存材料或未知内容', () => {
 assert.deepEqual(cleanSaved({favorites:['jade-product','jade-product','unknown',null],history:['cloud-flight'],theme:'light',secret:'材料'},ids), {favorites:['jade-product'],history:['cloud-flight'],theme:'light'});
 assert.deepEqual(cleanSaved({favorites:'wrong',history:[{}],theme:'wrong'},ids), {favorites:[],history:[],theme:'dark'});
});
test('存储损坏或禁止时保留本页收藏，刷新事件不会注入其他字段', () => {
 const denied = makeStore({getItem(){throw Error('denied');},setItem(){throw Error('denied');}}, ids);
 assert.equal(denied.available, false);
 assert.equal(denied.write({favorites:['jade-product']}), false);
 assert.deepEqual(denied.data.favorites, ['jade-product']);
 denied.refresh('{broken');
 assert.deepEqual(denied.data.favorites, ['jade-product']);
 denied.refresh(JSON.stringify({favorites:['cloud-flight'],private:'secret'}));
 assert.deepEqual(Object.keys(denied.data), ['favorites','history','theme']);
 const corrupt = makeStore({getItem(){return '{broken';},setItem(){}},ids);
 assert.equal(corrupt.available,false);
});
test('模板替换为纯文本，空白回退示例，不递归解释用户材料', () => {
 const prompt = compile(items[0], {subject:'<script>材料</script> {{scene}}',scene:'  '}, 'zh');
 assert.ok(prompt.includes('<script>材料</script> {{scene}}'));
 assert.ok(prompt.includes(items[0].fields[1].value));
 assert.ok(compile(items[0], {subject:'My bottle'}, 'en').includes('My bottle'));
});
test('文本导出保留实际材料、来源及验证状态', () => {
 const text = exportText(items[0], '实际替换的主体');
 assert.ok(text.includes('实际替换的主体'));
 assert.ok(text.includes(items[0].source.url));
 assert.ok(text.includes('未实测'));
});
test('分享仅识别有效条目ID，不从链接导入私人材料', () => {
 assert.equal(itemFromHash('#item=jade-product&subject=secret',items).id, 'jade-product');
 assert.equal(itemFromHash('#item=unknown',items), null);
 assert.equal(itemFromHash('#subject=secret',items), null);
});
test('静态服务限制可读文件、方法与联网能力', async () => {
 const server = createAppServer();
 await new Promise(resolve => server.listen(0,'127.0.0.1',resolve));
 const origin = `http://127.0.0.1:${server.address().port}`;
 try {
  const page = await fetch(origin);
  assert.equal(page.status,200);
  assert.match(page.headers.get('content-security-policy'), /connect-src 'none'/);
  assert.ok((await page.text()).includes('魔法工坊'));
  const head = await fetch(origin+'/assets/jade-bottle-thumb.webp',{method:'HEAD'});
  assert.equal(head.status,200); assert.equal((await head.arrayBuffer()).byteLength,0);
  for (const path of ['/PLAN.md','/serve.mjs','/assets/PROVENANCE.json','/%2e%2e/HANDOFF.md','/assets/../../HANDOFF.md']) assert.equal((await fetch(origin+path)).status,404);
  assert.equal((await fetch(origin,{method:'POST'})).status,405);
 } finally { server.closeAllConnections(); await new Promise(resolve => server.close(resolve)); }
});
