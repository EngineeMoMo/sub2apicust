import test from 'node:test';
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';
import { STORE_KEY } from './core.mjs';
// 复用本仓前端已经安装的测试依赖，不新增产品运行时依赖。
const require = createRequire(import.meta.url);
const { JSDOM } = require(require.resolve('jsdom',{paths:[fileURLToPath(new URL('../../frontend/',import.meta.url))]}));
const html = await readFile(new URL('./index.html',import.meta.url),'utf8');
let instance = 0;
test('视觉风格与题材分开筛选，未开放视频不可切换',async()=>app(({w,$,click,input,requests})=>{
 click('[data-style="像素"]');assert.equal(w.document.querySelectorAll('.art-card').length,1);
 click('[data-item-id="pixel-night"] .card-open');assert.match(w.document.querySelector('#preview-details .original-prompt').value,/pixel/);
 click('#close-preview');click('[data-style="水墨"]');assert.equal(w.document.querySelectorAll('.art-card').length,1);
 click('[data-media="video"]');assert.equal(w.document.querySelectorAll('.art-card').length,1);
 assert.equal(w.document.querySelector('[data-media="video"]').disabled,true);
 assert.equal(w.document.querySelector('[data-style="水墨"]').getAttribute('aria-pressed'),'true');
 click('[data-media="image"]');input('search','剪纸');assert.equal(w.document.querySelectorAll('.art-card').length,2);
 assert.equal($('submit-work').textContent,'分享作品与提示词');assert.equal(requests.length,0);
}));
test('同源只展示公开图片，视频投稿暂不展示，原始提示词仍可查看',async()=>{
 const id='b'.repeat(32),record={id,title:'投稿图片',author:'原创作者',category:'年轻人像',style:'电影感',media:'image',prompt_kind:'actual',prompt:'真实使用的人物场景与光线提示词',model:'作者模型',notes:'',media_url:`/api/v1/studio/gallery/${id}/media`};
 await app(async({w,$,click,requests,flush})=>{
  await flush();assert.equal(requests.length,1);assert.equal(requests[0].options.credentials,'omit');
  assert.equal(w.document.querySelectorAll('.art-card').length,48);
  click('[data-item-id="community-'+id+'"] .card-open');await flush();
  assert.equal($('preview-dialog').open,true);assert.ok(w.document.querySelector('#preview-media img'));
  assert.equal(w.document.querySelector('#preview-media video'),null);
  assert.equal(w.document.querySelector('#preview-details .original-prompt').value,record.prompt);
  assert.ok($('community-status').textContent.includes('1 份审核通过'));
 },{integrated:true,community:[record,{...record,id:'c'.repeat(32),status:'pending'},{...record,id:'d'.repeat(32),media:'video',media_url:'/api/v1/studio/gallery/'+ 'd'.repeat(32)+'/media'}]});
});
test('公开投稿接口失败不清空图库或草稿，也不显示假成功',async()=>app(async({w,$,input,flush})=>{
 input('material-subject','本页保留的材料');await flush();
 assert.equal(w.document.querySelectorAll('.art-card').length,47);
 assert.equal($('material-subject').value,'本页保留的材料');assert.match($('community-status').textContent,/暂不可用/);
 assert.equal($('refresh-community').disabled,false);
},{integrated:true,communityFailure:true}));
async function app(run, {data,hash='#item=jade-product',denied=false,reduced=false,integrated=false,community=[],communityFailure=false,setup=()=>{}}={}) {
 const dom = new JSDOM(html,{url:`http://localhost:4179/${hash}`});
 const {window:w} = dom;
 const clipboard=[], scrollTargets=[],requests=[];
 if(integrated){const meta=w.document.createElement('meta');meta.name='mofa-api-site';meta.content='same-origin';w.document.head.append(meta);}
 w.fetch=async(url,options)=>{requests.push({url,options});return {ok:!communityFailure,json:async()=>({code:0,data:community})};};
 w.matchMedia=()=>({matches:reduced});
 w.HTMLElement.prototype.scrollIntoView=function(){scrollTargets.push(this.id);};
 w.HTMLDialogElement.prototype.showModal=function(){this.setAttribute('open','');};
 w.HTMLDialogElement.prototype.close=function(){this.removeAttribute('open');this.dispatchEvent(new w.Event('close'));};
 w.HTMLMediaElement.prototype.play=async function(){this.dataset.played='true';this.dispatchEvent(new w.Event('play'));};
 w.HTMLMediaElement.prototype.pause=function(){this.dataset.paused='true';};
 Object.defineProperty(w.navigator,'clipboard',{value:{writeText:async text=>clipboard.push(text)},configurable:true});
 setup(w);
 if(data)w.localStorage.setItem(STORE_KEY,JSON.stringify(data));
 if(denied)Object.defineProperty(w,'localStorage',{get(){throw Error('denied');}});
 const names=['window','document','location','history','navigator'];
 const originals=Object.fromEntries(names.map(name=>[name,Object.getOwnPropertyDescriptor(globalThis,name)]));
 for(const name of names)Object.defineProperty(globalThis,name,{value:name==='window'?w:w[name],configurable:true,writable:true});
 const originalTimer=globalThis.setTimeout;
 const timers=[];
 globalThis.setTimeout=(fn,delay,...args)=>{const timer=originalTimer(fn,delay,...args);timer.unref();timers.push(timer);return timer;};
 const $=id=>w.document.getElementById(id);
 const click=selector=>{const node=w.document.querySelector(selector);assert.ok(node,selector);node.click();};
 const input=(id,value)=>{const node=$(id);node.value=value;node.dispatchEvent(new w.Event('input',{bubbles:true}));};
 try {
  await import(`./app.mjs?test=${++instance}`);
  await run({w,$,click,input,clipboard,scrollTargets,requests,flush:async()=>{for(let tick=0;tick<8;tick++)await Promise.resolve();}});
 }finally{
  globalThis.setTimeout=originalTimer;timers.forEach(clearTimeout);dom.window.close();
  for(const name of names)if(originals[name])Object.defineProperty(globalThis,name,originals[name]);else delete globalThis[name];
 }
}
test('瀑布流重算、媒体加载、筛选与观察器生命周期不打乱DOM或焦点', async () => {
 let columns = 3, shortHeight = 180, nextFrame = 0;
 const frames = new Map(), observers = [];
 const runFrames = () => { const pending = [...frames.values()]; frames.clear(); pending.forEach(callback => callback()); };
 await app(({w,$,click,input}) => {
  const gallery = $('gallery'), observer = observers[0];
  const ids = [...gallery.children].map(card => card.dataset.itemId);
  assert.equal(gallery.classList.contains('masonry'), true);
  assert.equal(gallery.children[3].style.gridRow, '199 / span 300');
  assert.equal(observer.targets.size, 48);
  observer.callback([{target:gallery,contentRect:{width:900}}]);
  observer.callback([{target:gallery,contentRect:{width:900}}]);
  assert.equal(frames.size, 1);runFrames();
  observer.callback([{target:gallery,contentRect:{width:900,height:5000}}]);
  assert.equal(frames.size, 0);
  shortHeight = 120;gallery.querySelector('img').dispatchEvent(new w.Event('load'));
  assert.equal(frames.size, 1);runFrames();
  assert.equal(gallery.children[3].style.gridRow, '139 / span 300');
  columns = 2;w.dispatchEvent(new w.Event('resize'));runFrames();
  assert.equal(gallery.children[2].style.gridColumn, '1');
  assert.ok([...gallery.children].every(card => Number(card.style.gridColumn) <= columns));
  assert.deepEqual([...gallery.children].map(card => card.dataset.itemId), ids);
  click('[data-style="像素"]');assert.equal(gallery.children.length, 1);assert.equal(observer.targets.size, 2);
  assert.equal(w.document.activeElement.dataset.style, '像素');
  assert.equal(gallery.children[0].style.gridRow, '1 / span 120');
  const image = gallery.querySelector('img');assert.equal(image.width,1536);assert.equal(image.height,1024);
  image.dispatchEvent(new w.Event('error'));runFrames();assert.ok(gallery.querySelector('.image-fallback'));
  input('search','不存在');assert.equal(gallery.classList.contains('masonry'), false);
  assert.equal(observer.targets.size, 1);
  click('[data-section="plays"]');assert.equal(observer.targets.size, 0);
  click('[data-section="discover"]');assert.equal(observer.targets.size, 48);
  click('[data-media="video"]');assert.equal(observer.targets.size, 48);
  assert.equal(gallery.querySelectorAll('.script-teaser').length,0);
  w.dispatchEvent(new w.Event('resize'));assert.equal(frames.size, 1);
  w.dispatchEvent(new w.Event('pagehide'));assert.equal(frames.size, 0);assert.equal(observer.targets.size, 0);
  w.dispatchEvent(new w.Event('pageshow'));assert.equal(observer.targets.size, 48);
 }, {setup:w => {
  w.ResizeObserver = class {
   constructor(callback) { this.callback=callback;this.targets=new Set();observers.push(this); }
   observe(target) { this.targets.add(target); }
   disconnect() { this.targets.clear(); }
  };
  w.requestAnimationFrame=callback=>{frames.set(++nextFrame,callback);return nextFrame;};
  w.cancelAnimationFrame=frame=>frames.delete(frame);
  const computed = w.getComputedStyle.bind(w);
  w.getComputedStyle=element=>element.id==='gallery'?{gridTemplateColumns:'0px 0px 300px 300px',columnGap:'18px',getPropertyValue:()=>String(columns)}:computed(element);
  w.HTMLElement.prototype.getBoundingClientRect=function(){return {height:this.classList.contains('art-card')?(this.dataset.itemId==='pixel-night'?shortHeight:300):0};};
 }});
});
test('没有ResizeObserver时仍可筛选并用窗口resize回退，不遮住内容', async () => app(({w,$,click}) => {
 assert.equal($('gallery').classList.contains('masonry'), false);
 click('[data-media="video"]');w.dispatchEvent(new w.Event('resize'));
 assert.equal($('gallery').children.length,47);
}));
test('发现、Skills与玩法保留，视频显示待开放且不可进入', async()=>app(({w,$,click,input})=>{
 assert.equal(w.document.querySelectorAll('.art-card').length,47);
 const video=w.document.querySelector('[data-media="video"]');assert.equal(video.disabled,true);assert.match(video.textContent,/待开放/);
 assert.equal(video.getAttribute('aria-describedby'),'video-availability');assert.equal($('video-availability').hidden,false);
 click('[data-media="video"]');assert.equal(w.document.querySelectorAll('.art-card').length,47);
 input('search','不存在');assert.equal($('empty-state').hidden,false);
 click('#empty-action');assert.equal(w.document.querySelectorAll('.art-card').length,47);
 click('[data-section="skills"]');assert.equal(w.document.querySelectorAll('.art-card').length,4);
 click('[data-item-id="remotion-caption"] .card-open');assert.ok($('skill-command').textContent.includes('npx skills add'));
 click('[data-section="plays"]');assert.equal(w.document.querySelectorAll('.play-story').length,4);
 assert.equal(w.document.querySelectorAll('.play-story button:disabled').length,4);
 click('.play-story button');assert.equal($('selected-title').textContent,'一瓶清透的夏天');
}));
test('私人材料切换后仍保留，但收藏与分享不包含材料', async()=>app(async({w,$,click,input,clipboard,flush})=>{
 input('material-subject','私人材料唯一标记');
 click('[data-item-id="paper-mascot"] .card-open');click('[data-item-id="jade-product"] .card-open');
 assert.ok($('prompt-output').value.includes('私人材料唯一标记'));
 click('#bench-save');
 const stored=w.localStorage.getItem(STORE_KEY);assert.ok(!stored.includes('私人材料唯一标记'));
 assert.deepEqual(Object.keys(JSON.parse(stored)),['favorites','history','theme']);
 click('#share-item');await flush();assert.equal(clipboard.at(-1),'http://localhost:4179/#item=jade-product');
 click('#copy-prompt');await flush();assert.ok(clipboard.at(-1).includes('私人材料唯一标记'));
 click('#language-en');assert.ok($('prompt-output').value.includes('私人材料唯一标记'));
}));
test('恢复材料需要确认，取消时内容不变', async()=>app(({w,$,click,input})=>{
 input('material-subject','用户修改');click('#reset-materials');
 const controls=w.document.querySelectorAll('.inline-confirm button');assert.equal(controls.length,2);
 controls[0].click();assert.equal($('material-subject').value,'用户修改');
 click('#reset-materials');w.document.querySelectorAll('.inline-confirm button')[1].click();
 assert.equal($('material-subject').value,'无品牌玉绿色磨砂护肤瓶');
}));
test('剪贴板拒绝时显示准确的手动复制文本', async()=>app(async({w,$,click,input,flush})=>{
 w.navigator.clipboard.writeText=async()=>{throw Error('denied');};
 input('material-subject','手动文本');click('#copy-prompt');await flush();
 assert.equal(w.document.querySelector('.manual-copy textarea').value,$('prompt-output').value);
 click('#share-item');await flush();
 assert.equal(w.document.querySelector('.manual-copy textarea').value,'http://localhost:4179/#item=jade-product');
}));
test('收藏恢复、跨窗口变更与清空历史只涉及条目ID', async()=>app(({w,$,click})=>{
 click('[data-section="saved"]');assert.equal(w.document.querySelectorAll('.art-card').length,1);
 w.dispatchEvent(new w.StorageEvent('storage',{key:STORE_KEY,newValue:JSON.stringify({favorites:['celestial-gate'],history:['jade-product'],theme:'light',private:'secret'})}));
 assert.equal(w.document.documentElement.dataset.theme,'light');
 assert.equal(w.document.querySelectorAll('.art-card').length,0);
 assert.equal($('saved-count').textContent,'0');assert.equal($('video-availability').hidden,false);
 click('#about-button');click('#clear-history');click('#close-about');click('#history-button');
 assert.equal(w.document.querySelectorAll('.art-card').length,0);assert.equal($('empty-state').hidden,false);
},{data:{favorites:['jade-product'],history:['celestial-gate'],theme:'dark'}}));
test('禁止存储时收藏仍在本页可用', async()=>app(({w,$,click})=>{
 click('#bench-save');click('[data-section="saved"]');
 assert.equal(w.document.querySelectorAll('.art-card').length,1);
 assert.ok($('toast').textContent.includes('当前页保留'));
},{denied:true}));
test('有效Skill深链接直达，未知条目回退安全示例', async()=>{
 await app(({w,$})=>{assert.equal($('selected-title').textContent,'给短片加上清晰字幕');assert.equal(w.document.querySelector('[data-section="skills"]').getAttribute('aria-current'),'page');},{hash:'#item=remotion-caption'});
 await app(({$})=>assert.equal($('selected-title').textContent,'像素港口 · 夜航开始'),{hash:'#item=unknown&subject=secret'});
});
test('旧分镜深链接回退图片并友好提示，不创建播放器或分镜入口', async()=>app(({w,$})=>{
 assert.equal($('selected-title').textContent,'像素港口 · 夜航开始');
 assert.match($('toast').textContent,/暂未开放/);
 assert.equal(w.document.querySelector('[data-item-id="elevator-secret"]'),null);
 assert.equal(w.document.querySelector('.script-preview'),null);
 assert.equal(w.document.querySelector('video'),null);
 assert.equal($('preview-dialog').open,false);
},{hash:'#item=elevator-secret',reduced:true}));
test('后退到未开放分镜不替换当前图片材料，暂存的收藏记录不删除', async()=>app(({w,$,input,click})=>{
 input('material-subject','当前图片的私人材料');
 w.history.pushState(null,'','#item=coast-reunion');w.dispatchEvent(new w.PopStateEvent('popstate'));
 assert.equal($('selected-title').textContent,'一瓶清透的夏天');
 assert.ok($('prompt-output').value.includes('当前图片的私人材料'));assert.match($('toast').textContent,/暂未开放/);
 click('[data-section="saved"]');assert.equal($('gallery').children.length,1);assert.equal($('saved-count').textContent,'1');
 const stored=JSON.parse(w.localStorage.getItem(STORE_KEY));
 assert.ok(stored.favorites.includes('coast-reunion'));assert.ok(!JSON.stringify(stored).includes('私人材料'));
},{data:{favorites:['coast-reunion','jade-product'],history:['coast-reunion']}}));
test('已审核视频投稿深链接也不能绕过前台暂停，图片入口正常', async()=>{
 const id='e'.repeat(32),record={id,title:'已审核视频',author:'作者',category:'年轻人像',style:'电影感',media:'video',prompt_kind:'actual',prompt:'真实使用的视频分镜与镜头提示词',model:'作者模型',notes:'12秒',media_url:`/api/v1/studio/gallery/${id}/media`};
 await app(async({w,$,flush})=>{
  await flush();assert.equal($('gallery').children.length,47);assert.equal($('selected-title').textContent,'像素港口 · 夜航开始');
  assert.equal(w.document.querySelector('video'),null);
  assert.equal(w.document.querySelector('[data-item-id="community-'+id+'"]'),null);
  assert.match($('community-status').textContent,/尚无可展示的图片/);
 },{integrated:true,community:[record],hash:'#item=community-'+id});
});
test('发现与Skills的资源类型选择进入可命中区域，导航同步', async()=>app(({w,$})=>{
 const choose=value=>{$('resource-type').value=value;$('resource-type').dispatchEvent(new w.Event('change',{bubbles:true}));};
 choose('skill');assert.equal(w.document.querySelectorAll('.art-card').length,2);assert.equal(w.document.querySelector('[data-section="skills"]').getAttribute('aria-current'),'page');
 choose('workflow');assert.equal(w.document.querySelectorAll('.art-card').length,2);
 choose('prompt');assert.equal(w.document.querySelectorAll('.art-card').length,47);assert.equal(w.document.querySelector('[data-section="discover"]').getAttribute('aria-current'),'page');
}));
test('返回灵感恢复列表焦点，筛选与当前材料保留', async()=>app(({w,$,click,input,scrollTargets})=>{
 click('[data-media="image"]');input('material-subject','我的瓶子');
 click('#return-to-library');assert.equal(scrollTargets.at(-1),'collection');assert.equal(w.document.activeElement.id,'search');
 assert.equal(w.document.querySelectorAll('.art-card').length,47);assert.equal($('material-subject').value,'我的瓶子');
 click('[data-section="plays"]');click('#return-to-library');assert.equal(w.document.activeElement.id,'history-button');
}));
test('两个模态框均有有效且非空的可访问名称引用', async()=>app(({w,$,click})=>{
 click('.preview-button');assert.equal($('preview-dialog').open,true);assert.ok($( $('preview-dialog').getAttribute('aria-labelledby') ).textContent.includes('一瓶清透的夏天'));
 click('#close-preview');assert.equal($('preview-dialog').open,false);
 click('#about-button');assert.equal($('about-dialog').open,true);assert.equal($( $('about-dialog').getAttribute('aria-labelledby') ).textContent,'从灵感到自己的作品');
}));

test('左侧图片点开完整图，关闭恢复条目焦点，右侧也能放大', async()=>app(({w,$,click})=>{
 click('[data-item-id="sweet-man"] .card-open');
 assert.equal($('preview-dialog').open,true);
 assert.ok($('preview-image').src.endsWith('/assets/sweet-man.webp'));
 click('#close-preview');assert.equal(w.document.activeElement.closest('[data-item-id]').dataset.itemId,'sweet-man');
 click('.preview-button');assert.equal($('preview-dialog').open,true);
 click('#close-preview');assert.equal(w.document.activeElement.className,'preview-button');
}));

test('10套视频分镜有独立题材与完整双语提示词，明确无样片', async()=>{
 const {items}=await import('./catalog.mjs');
 const plans=items.filter(item=>item.media==='video'&&item.type==='prompt');
 assert.equal(plans.length,10);
 assert.deepEqual(['美剧感','韩剧感','修仙动漫'].map(category=>plans.filter(item=>item.category===category).length),[4,3,3]);
 for(const item of plans){
  assert.equal(item.art,undefined);assert.equal(item.video,undefined);
  assert.match(item.status,/样片待补/);
  assert.equal(item.shots.length,3);
  assert.ok(item.zh.includes('12秒')&&item.en.includes('12-second'));
  assert.ok(item.requirements.includes('没有对应可播放样片'));
 }
});
test('图片原始提示词与保存的实际生成记录逐字一致', async()=>{
 const {items}=await import('./catalog.mjs');
 const fresh=JSON.parse(await readFile(new URL('assets/PROVENANCE-VIBRANT-20261001.json',import.meta.url),'utf8'));
 const expanded=JSON.parse(await readFile(new URL('assets/PROVENANCE-EXPANDED-20261001.json',import.meta.url),'utf8'));
 const older=JSON.parse(await readFile(new URL('assets/PROVENANCE.json',import.meta.url),'utf8'));
 const curated=JSON.parse(await readFile(new URL('assets/PROVENANCE-CURATED-20261001.json',import.meta.url),'utf8'));
 for(const asset of [...curated.assets,...expanded.assets,...fresh.assets,...older.assets].filter(asset=>!['midnight-editorial','coral-sneaker'].includes(asset.id))){
  const item=items.find(item=>item.art===asset.id);
  assert.ok(item,asset.id);assert.equal(item.rawPrompt,asset.prompt);
  assert.ok((await readFile(new URL('assets/'+asset.id+'.webp',import.meta.url))).length>1000);
 }
 assert.equal(items.filter(item=>item.rawPrompt).length,39);
});
test('弹窗分别复制原始提示词与我的模板，替换材料不改原始记录', async()=>app(async({w,$,click,input,clipboard,flush})=>{
 click('[data-item-id="festival-friends"] .card-use');
 input('material-subject','原创测试主体');
 click('[data-item-id="festival-friends"] .card-open');
 const original=w.document.querySelector('#preview-details .original-prompt').value;
 assert.ok(original.includes('Two fictional East Asian adult women'));
 assert.ok($('preview-prompt-output').value.includes('原创测试主体'));
 click('#preview-details [data-copy-kind="original"]');await flush();assert.equal(clipboard.at(-1),original);
 click('#preview-copy');await flush();assert.equal(clipboard.at(-1),$('preview-prompt-output').value);
 click('#preview-details > .secondary-button');
 assert.equal($('preview-dialog').open,false);assert.equal(w.document.activeElement.id,'workbench');
 assert.equal($('material-subject').value,'原创测试主体');
}));
test('保留扩展作品均可打开对应原始提示词、替换和分别复制', async()=>app(async({w,$,click,input,clipboard,flush})=>{
 const { expansionImages } = await import('./new-gallery.mjs');
 click('[data-media="image"]');
 assert.equal(w.document.querySelectorAll('.art-card').length,47);
 for(const item of expansionImages){
  click('[data-item-id="'+item.id+'"] .card-use');
  const replacement='自己的主体-'+item.id;
  input('material-subject',replacement);
  click('[data-item-id="'+item.id+'"] .card-open');
  assert.equal(w.document.querySelector('#preview-details .original-prompt').value,item.rawPrompt);
  assert.ok($('preview-prompt-output').value.includes(replacement));
  click('#preview-details [data-copy-kind="original"]');await flush();
  assert.equal(clipboard.at(-1),item.rawPrompt);
  click('#preview-copy');await flush();
  assert.equal(clipboard.at(-1),$('preview-prompt-output').value);
  click('#close-preview');
 }
}));

test('未保存原始提示词的旧图只显示明确的缺失说明，不编造记录', async()=>app(({w,$,click})=>{
 click('[data-item-id="sweet-man"] .card-open');
 assert.ok($('preview-details').textContent.includes('原始提示词记录未保存'));
 assert.equal(w.document.querySelector('#preview-details .original-prompt'),null);
 assert.ok($('preview-prompt-output').value.length>100);
}));
test('弹窗复制失败时手动文本留在弹窗内，而不是被模态遮住', async()=>app(async({w,$,click,flush})=>{
 w.navigator.clipboard.writeText=async()=>{throw Error('denied');};
 click('[data-item-id="festival-friends"] .card-open');click('#preview-copy');await flush();
 assert.equal(w.document.querySelector('#preview-details .manual-copy textarea').value,$('preview-prompt-output').value);
 assert.equal(w.document.activeElement,w.document.querySelector('#preview-details .manual-copy textarea'));
}));
test('卡片快速复制使用该条目的当前草稿，不更改选中条目', async()=>app(async({$ ,click,input,clipboard,flush})=>{
 click('[data-item-id="festival-friends"] .card-use');input('material-subject','保存当前页草稿');
 click('[data-item-id="city-skater"] .card-use');
 const selectedTitle=$('selected-title').textContent;
 click('[data-item-id="festival-friends"] .card-copy');await flush();
 assert.ok(clipboard.at(-1).includes('保存当前页草稿'));
 assert.equal($('selected-title').textContent,selectedTitle);
}));
