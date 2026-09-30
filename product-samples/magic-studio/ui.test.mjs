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
async function app(run, {data,hash='',denied=false,reduced=false}={}) {
 const dom = new JSDOM(html,{url:`http://localhost:4179/${hash}`});
 const {window:w} = dom;
 const clipboard=[], scrollTargets=[];
 w.matchMedia=()=>({matches:reduced});
 w.HTMLElement.prototype.scrollIntoView=function(){scrollTargets.push(this.id);};
 w.HTMLDialogElement.prototype.showModal=function(){this.setAttribute('open','');};
 w.HTMLDialogElement.prototype.close=function(){this.removeAttribute('open');};
 Object.defineProperty(w.navigator,'clipboard',{value:{writeText:async text=>clipboard.push(text)},configurable:true});
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
  await run({w,$,click,input,clipboard,scrollTargets,flush:async()=>{await Promise.resolve();await Promise.resolve();}});
 }finally{
  globalThis.setTimeout=originalTimer;timers.forEach(clearTimeout);dom.window.close();
  for(const name of names)if(originals[name])Object.defineProperty(globalThis,name,originals[name]);else delete globalThis[name];
 }
}
test('发现、视频、Skills与玩法之间能完成筛选和选择', async()=>app(({w,$,click,input})=>{
 assert.equal(w.document.querySelectorAll('.art-card').length,12);
 click('[data-media="video"]');assert.equal(w.document.querySelectorAll('.art-card').length,4);
 input('search','不存在');assert.equal($('empty-state').hidden,false);
 click('#empty-action');assert.equal(w.document.querySelectorAll('.art-card').length,12);
 click('[data-section="skills"]');assert.equal(w.document.querySelectorAll('.art-card').length,4);
 click('[data-item-id="remotion-caption"] .card-open');assert.ok($('skill-command').textContent.includes('npx skills add'));
 click('[data-section="plays"]');assert.equal(w.document.querySelectorAll('.play-story').length,4);
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
 w.dispatchEvent(new w.StorageEvent('storage',{key:STORE_KEY,newValue:JSON.stringify({favorites:['cloud-flight'],history:['jade-product'],theme:'light',private:'secret'})}));
 assert.equal(w.document.documentElement.dataset.theme,'light');
 assert.equal(w.document.querySelector('.art-card').dataset.itemId,'cloud-flight');
 click('#about-button');click('#clear-history');click('#close-about');click('#history-button');
 assert.equal(w.document.querySelectorAll('.art-card').length,0);assert.equal($('empty-state').hidden,false);
},{data:{favorites:['jade-product'],history:['cloud-flight'],theme:'dark'}}));
test('禁止存储时收藏仍在本页可用', async()=>app(({w,$,click})=>{
 click('#bench-save');click('[data-section="saved"]');
 assert.equal(w.document.querySelectorAll('.art-card').length,1);
 assert.ok($('toast').textContent.includes('当前页保留'));
},{denied:true}));
test('有效Skill深链接直达，未知条目回退安全示例', async()=>{
 await app(({w,$})=>{assert.equal($('selected-title').textContent,'给短片加上清晰字幕');assert.equal(w.document.querySelector('[data-section="skills"]').getAttribute('aria-current'),'page');},{hash:'#item=remotion-caption'});
 await app(({$})=>assert.equal($('selected-title').textContent,'一瓶清透的夏天'),{hash:'#item=unknown&subject=secret'});
});
test('运镜播放可以暂停，减少动画偏好明确提示', async()=>app(({w,$,click})=>{
 click('[data-item-id="product-push"] .card-open');
 assert.ok($('bench-preview').textContent.includes('二维运镜示意'));
 click('.motion-controls button');assert.equal(w.document.querySelector('.motion-stage').classList.contains('playing'),true);
 click('.motion-controls button');assert.equal(w.document.querySelector('.motion-stage').classList.contains('playing'),false);
 assert.ok($('bench-preview').textContent.includes('已减少动态效果'));
},{reduced:true}));
test('发现与Skills的资源类型选择进入可命中区域，导航同步', async()=>app(({w,$})=>{
 const choose=value=>{$('resource-type').value=value;$('resource-type').dispatchEvent(new w.Event('change',{bubbles:true}));};
 choose('skill');assert.equal(w.document.querySelectorAll('.art-card').length,2);assert.equal(w.document.querySelector('[data-section="skills"]').getAttribute('aria-current'),'page');
 choose('workflow');assert.equal(w.document.querySelectorAll('.art-card').length,2);
 choose('prompt');assert.equal(w.document.querySelectorAll('.art-card').length,12);assert.equal(w.document.querySelector('[data-section="discover"]').getAttribute('aria-current'),'page');
}));
test('返回灵感恢复列表焦点，筛选与当前材料保留', async()=>app(({w,$,click,input,scrollTargets})=>{
 click('[data-media="video"]');click('[data-item-id="product-push"] .card-open');input('material-subject','我的瓶子');
 click('#return-to-library');assert.equal(scrollTargets.at(-1),'collection');assert.equal(w.document.activeElement.id,'search');
 assert.equal(w.document.querySelectorAll('.art-card').length,4);assert.equal($('material-subject').value,'我的瓶子');
 click('[data-section="plays"]');click('#return-to-library');assert.equal(w.document.activeElement.id,'history-button');
}));
test('两个模态框均有有效且非空的可访问名称引用', async()=>app(({w,$,click})=>{
 click('.preview-button');assert.equal($('preview-dialog').open,true);assert.ok($( $('preview-dialog').getAttribute('aria-labelledby') ).textContent.includes('一瓶清透的夏天'));
 click('#close-preview');assert.equal($('preview-dialog').open,false);
 click('#about-button');assert.equal($('about-dialog').open,true);assert.equal($( $('about-dialog').getAttribute('aria-labelledby') ).textContent,'从灵感到自己的作品');
}));
