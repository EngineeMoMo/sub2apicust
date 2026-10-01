import test from 'node:test';
import { readFile } from 'node:fs/promises';
import assert from 'node:assert/strict';
import { items, plays, artSizes } from './catalog.mjs';
import { findItems, cleanSaved, makeStore, compile, exportText, itemFromHash, STORE_KEY, communityItem, masonryPositions, VIDEO_CONTENT_ENABLED, isItemAvailable } from './core.mjs';
import { createAppServer } from './serve.mjs';
const ids = items.map(x => x.id);
const state = { section:'discover', media:'all', type:'all', category:'all', query:'', sort:'featured' };
const saved = { favorites:[], history:[] };
test('五类各补两图与首页人像具有准确提示词和完整资产，移除项不能再直达', async () => {
 const { curatedImages, homepageImage } = await import('./new-gallery.mjs');
 const { createHash } = await import('node:crypto');
 const provenance = JSON.parse(await readFile(new URL('assets/PROVENANCE-CURATED-20261001.json',import.meta.url),'utf8'));
 assert.equal(curatedImages.length,10);
 for(const category of ['Cosplay','动物自然','海报社媒','动漫二次元','空间设计']) assert.equal(curatedImages.filter(item=>item.category===category).length,2);
 for(const item of [...curatedImages,homepageImage]) {
  const record = provenance.assets.find(asset=>asset.id===item.id);
  assert.equal(item.rawPrompt,record.prompt);
  assert.ok(findItems(items,{...state,category:item.category},saved).some(found=>found.id===item.id));
  for(const suffix of ['.png','.webp','-thumb.webp']) {
   const bytes = await readFile(new URL('assets/'+item.art+suffix,import.meta.url));
   assert.equal(createHash('sha256').update(bytes).digest('hex'),record.files[suffix].sha256);
   assert.ok(Math.abs(record.files[suffix].width/record.files[suffix].height-record.files['.png'].width/record.files['.png'].height)<0.002);
  }
 }
 assert.deepEqual(await readFile(new URL('assets/window-portrait.webp',import.meta.url)),await readFile(new URL('../../frontend/src/custom/assets/studio-reference-portrait.webp',import.meta.url)));
 for(const id of ['midnight-editorial','coral-sneaker']) {
  assert.ok(!ids.includes(id));
  assert.equal(itemFromHash('#item='+id,items),null);
 }
 assert.equal(findItems(items,{...state,category:'时尚肖像'},saved).length,1);
 assert.equal(findItems(items,{...state,category:'时尚肖像'},saved)[0].id,'fashion-editorial');
});
test('公开投稿收藏保存的只是规范ID，加载目录前可恢复，不接受外链和私人材料',()=>{
 const id='community-'+'d'.repeat(32);
 const restored=cleanSaved({favorites:[id,'https://evil.test','community-../private'],history:[id],prompt:'不保存'},ids);
 assert.deepEqual(restored,{favorites:[id],history:[id],theme:'light'});
});
test('六种新画风原始记录和资产逐字对应，独立筛选', async () => {
 const { readFile } = await import('node:fs/promises');
 const { createHash } = await import('node:crypto');
 const { diverseImages } = await import('./new-gallery.mjs');
 const provenance = JSON.parse(await readFile(new URL('assets/PROVENANCE-STYLES-20261001.json',import.meta.url),'utf8'));
 assert.equal(diverseImages.length,6);
 assert.deepEqual(new Set(diverseImages.flatMap(item=>item.styles)),new Set(['像素','水墨','美漫','水彩','剪纸','复古未来']));
 for(const item of diverseImages){
  const record=provenance.assets.find(record=>record.id===item.id);
  assert.equal(item.rawPrompt,record.prompt);
  assert.equal(findItems(items,{...state,style:item.styles[0]},saved).length,1);
  for(const suffix of ['.png','.webp','-thumb.webp']){
   const bytes=await readFile(new URL('assets/'+item.id+suffix,import.meta.url));
   assert.equal(createHash('sha256').update(bytes).digest('hex'),record.files[suffix].sha256);
   assert.ok(Math.abs(record.files[suffix].width/record.files[suffix].height-record.width/record.height)<0.002);
  }
 }
});
test('公开投稿适配拒绝外链、无效编号、待审核与虚假来源，视频保留提示词', () => {
 const id='a'.repeat(32),record={id,title:'测试作品',author:'作者',category:'奇幻风景',style:'水彩',media:'video',prompt_kind:'actual',prompt:'完整的视频镜头动作提示词',model:'投稿模型',notes:'',media_url:`/api/v1/studio/gallery/${id}/media`};
 const item=communityItem(record);
 assert.equal(item.video,record.media_url);assert.equal(item.art,undefined);assert.equal(item.rawPrompt,record.prompt);
 assert.ok(item.requirements.includes('不保证'));
 const reference=communityItem({...record,prompt_kind:'reference'});
 const withPlaceholder=communityItem({...record,prompt:'完整的 {{character}} 分镜与 {{camera}} 镜头提示词'});
 assert.equal(compile(withPlaceholder,{},'zh'),withPlaceholder.rawPrompt);
 assert.equal(reference.rawPrompt,undefined);assert.match(reference.recordNote,/不是原始/);
 for(const mutation of [{media_url:'https://other.test/media'},{id:'../private'},{status:'pending'},{prompt_kind:'made-up'},{prompt:''},{notes:null},{media:'html'}])assert.equal(communityItem({...record,...mutation}),null);
});

test('目录完整：61条唯一资源，所有玩法指向已有条目', () => {
 assert.equal(new Set(ids).size, 61);
 assert.equal(items.filter(x => x.media === 'image' && x.type === 'prompt').length, 47);
 assert.equal(items.filter(x => x.media === 'video' && x.type === 'prompt').length, 10);
 assert.equal(items.filter(x => x.type !== 'prompt').length, 4);
 assert.equal(plays.length, 4);
 for (const play of plays) for (const step of play.steps) assert.ok(ids.includes(step.id));
 for (const item of items) {
  assert.ok(item.source.url.startsWith('https://'));
  for (const language of ['zh','en']) assert.doesNotMatch(compile(item, {}, language), /\{\{\w+\}\}/);
 }
});
test('保留扩展图有逐字原始记录、三种资产与无裁切尺寸证据', async () => {
 const { readFile } = await import('node:fs/promises');
 const { createHash } = await import('node:crypto');
 const { expansionImages } = await import('./new-gallery.mjs');
 const provenance = JSON.parse(await readFile(new URL('assets/PROVENANCE-EXPANDED-20261001.json', import.meta.url), 'utf8'));
 assert.equal(expansionImages.length, 8);
 assert.equal(provenance.assets.length, 10);
 assert.equal(new Set(expansionImages.map(item => item.category)).size, 7);
 assert.deepEqual(provenance.referenceImages, []);
 for (const item of expansionImages) {
  const asset = provenance.assets.find(asset => asset.id === item.art);
  assert.ok(asset, item.id);
  assert.equal(item.rawPrompt, asset.prompt);
  assert.equal(item.generation.tool, provenance.tool);
  assert.deepEqual(item.fields.map(field => field.key), ['subject', 'scene', 'light']);
  for (const language of ['zh', 'en']) {
   assert.doesNotMatch(compile(item, {}, language), /\{\{\w+\}\}/);
   assert.ok(compile(item, { subject: '本轮替换主体' }, language).includes('本轮替换主体'));
  }
  assert.ok(exportText(item, compile(item, {})).includes(asset.prompt));
  for (const suffix of ['.png', '.webp', '-thumb.webp']) {
   const bytes = await readFile(new URL('assets/' + item.art + suffix, import.meta.url));
   const metadata = asset.files[suffix];
   assert.equal(bytes.length, metadata.bytes);
   assert.equal(createHash('sha256').update(bytes).digest('hex'), metadata.sha256);
   const original = asset.files['.png'];
   assert.ok(Math.abs(metadata.width / metadata.height - original.width / original.height) < 0.002);
  }
 }
});

test('媒介、资源类型与用途可以同时筛选', () => {
 const result = findItems(items, {...state,section:'skills',media:'video',type:'skill'}, saved);
 assert.equal(result.length, 2);
 assert.ok(result.every(x => x.media === 'video' && x.type === 'skill'));
 assert.equal(findItems(items, {...state,media:'image',category:'电商产品'}, saved).length, 4);
});
test('多关键词搜索同时匹配中文、英文、工具与标签', () => {
 assert.equal(findItems(items, {...state,query:'产品 玻璃'}, saved)[0].id, 'jade-product');
 assert.equal(findItems(items, {...state,section:'skills',query:'ＲＥＭＯＴＩＯＮ'}, saved).length, 2);
 assert.equal(findItems(items, {...state,query:'确实不存在的名字'}, saved).length, 0);
});
test('收藏包含跨媒介资源，浏览记录按最新顺序显示', () => {
 const selected = {favorites:['remotion-caption','jade-product'],history:['celestial-gate','jade-product']};
 assert.equal(findItems(items, {...state,section:'saved'}, selected).length, 2);
 assert.deepEqual(findItems(items, {...state,section:'history'}, selected).map(x=>x.id), ['jade-product']);
 assert.equal(findItems(items, {...state,section:'saved',media:'image'}, selected).length, 1);
});
test('视频与分镜暂不开放，检索收藏历史均不展示但保留原始记录', () => {
 assert.equal(VIDEO_CONTENT_ENABLED, false);
 assert.equal(findItems(items, state, saved).length,47);
 assert.equal(findItems(items, {...state,media:'video'}, saved).length,0);
 for (const section of ['discover','saved','history']) {
  const result=findItems(items,{...state,section},{favorites:ids,history:ids});
  assert.ok(result.every(isItemAvailable));
  assert.ok(!result.some(item=>item.type==='prompt'&&item.media==='video'));
 }
 assert.equal(items.filter(item=>!isItemAvailable(item)).length,10);
 assert.equal(findItems(items,{...state,section:'skills'},saved).length,4);
 assert.deepEqual(cleanSaved({favorites:['celestial-gate'],history:['coast-reunion']},ids).favorites,['celestial-gate']);
});
test('存储只保留白名单ID与主题，不保存材料或未知内容', () => {
 assert.deepEqual(cleanSaved({favorites:['jade-product','jade-product','unknown',null],history:['celestial-gate'],theme:'light',secret:'材料'},ids), {favorites:['jade-product'],history:['celestial-gate'],theme:'light'});
 assert.deepEqual(cleanSaved({favorites:'wrong',history:[{}],theme:'wrong'},ids), {favorites:[],history:[],theme:'light'});
});
test('存储损坏或禁止时保留本页收藏，刷新事件不会注入其他字段', () => {
 const denied = makeStore({getItem(){throw Error('denied');},setItem(){throw Error('denied');}}, ids);
 assert.equal(denied.available, false);
 assert.equal(denied.write({favorites:['jade-product']}), false);
 assert.deepEqual(denied.data.favorites, ['jade-product']);
 denied.refresh('{broken');
 assert.deepEqual(denied.data.favorites, ['jade-product']);
 denied.refresh(JSON.stringify({favorites:['celestial-gate'],private:'secret'}));
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
 assert.ok(text.includes(items[0].status));
 assert.ok(text.includes('复用模板（当前材料）'));
 assert.ok(text.includes(items[0].rawPrompt));
 assert.ok(text.includes('实际生成记录，不含替换材料'));
});
test('分享仅识别有效条目ID，不从链接导入私人材料', () => {
 assert.equal(itemFromHash('#item=jade-product&subject=secret',items).id, 'jade-product');
 assert.equal(itemFromHash('#item=unknown',items), null);
 assert.equal(itemFromHash('#subject=secret',items), null);
});
test('瀑布流优先填最短列，原顺序按顶部从左到右延续', () => {
 const heights = [180, 320, 260, 140, 200, 300, 170];
 const positions = masonryPositions(heights, 3, 18);
 assert.deepEqual(positions.slice(0, 4), [
  {column:1,row:1,span:180}, {column:2,row:1,span:320},
  {column:3,row:1,span:260}, {column:1,row:199,span:140}
 ]);
 positions.forEach((position, index) => {
  const previous = positions.slice(0,index).filter(other => other.column === position.column).at(-1);
  if (previous) assert.equal(position.row - previous.row - previous.span, 18);
  if (index) assert.ok(position.row >= positions[index-1].row);
 });
 assert.equal(Math.max(...positions.map(position => position.row-1+position.span)), 638);
});
test('瀑布流支持单列、分数高度、空结果与列数变化', () => {
 assert.deepEqual(masonryPositions([], 2, 12), []);
 assert.deepEqual(masonryPositions([120.2, 90.5], 1, 12), [
  {column:1,row:1,span:121}, {column:1,row:134,span:91}
 ]);
 assert.deepEqual(masonryPositions([200, 100, 200], 2, 12).map(position => position.column), [1,2,2]);
 assert.deepEqual(masonryPositions([200, 100, 200], 3, 18).map(position => position.row), [1,1,1]);
});
test('内置画幅预留值与47份原始PNG尺寸一致', async () => {
 const images = items.filter(item => item.media === 'image' && item.art);
 assert.equal(Object.keys(artSizes).length, 47);
 for (const item of images) {
  const image = await readFile(new URL(`./assets/${item.art}.png`, import.meta.url));
  assert.deepEqual(artSizes[item.art], {width:image.readUInt32BE(16),height:image.readUInt32BE(20)});
 }
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
