import { items as builtinItems, plays, typeLabels, artSizes } from './catalog.mjs';
import { icon } from './icons.mjs';
import { STORE_KEY, makeStore, findItems, compile, exportText, itemFromHash, communityItem, masonryPositions, VIDEO_CONTENT_ENABLED, isItemAvailable, clipItem, formatBytes } from './core.mjs';
import { IMAGE_SIZES, GENERATE_TIMEOUT_MS, imageConnection, generateImage, suggestedSize, downloadName } from './generate.mjs';
import { clips } from './videos.mjs';
import { loadClip, isClipCached, pruneClips, clipCacheSummary, clearClips } from './video-cache.mjs';

const $ = id => document.getElementById(id);
const el = (tag, cls, text) => { const n=document.createElement(tag); if(cls)n.className=cls; if(text!==undefined)n.textContent=text; return n; };
const decorate = (n,name,text) => { if(n.flash){clearTimeout(n.flash);n.flash=0;n.flashOriginal=null;n.classList.remove('copied');} n.replaceChildren(icon(name));if(text)n.append(document.createTextNode(text));return n; };
const option = (label, value) => { const node=document.createElement('option'); node.textContent=label; node.value=value; return node; };
const reducedMotion = () => window.matchMedia('(prefers-reduced-motion: reduce)').matches;
// 复制成功后在按钮上短暂显示“已复制”，再恢复原内容；纯图标按钮只改颜色。
function flashCopied(button){
 if(!button?.isConnected)return;
 if(!button.flash)button.flashOriginal=[...button.childNodes];
 const iconOnly=!button.flashOriginal.some(node=>node.nodeType===3&&node.textContent.trim());
 clearTimeout(button.flash);button.classList.add('copied');
 if(!iconOnly){button.replaceChildren(icon('check'));button.append(document.createTextNode('已复制'));}
 button.flash=setTimeout(()=>{button.flash=0;button.classList.remove('copied');if(!iconOnly)button.replaceChildren(...button.flashOriginal);button.flashOriginal=null;},1600);
}
const gallery = $('gallery');
let galleryFrame = 0, galleryStopped = false, galleryWidth = 0, galleryColumns = 0;
function layoutGallery() {
 if (galleryStopped || gallery.hidden) return;
 const cards = [...gallery.children];
 if (!cards.length) { gallery.classList.remove('masonry'); return; }
 const style = window.getComputedStyle(gallery);
 const columns = Math.max(1, parseInt(style.getPropertyValue('--gallery-columns'), 10) || 1);
 if (columns !== galleryColumns) {
  cards.forEach((card, index) => { card.style.gridColumn = String(index % columns + 1); });
  galleryColumns = columns;
 }
 const heights = cards.map(card => card.getBoundingClientRect().height);
 if (heights.some(height => height <= 0)) return;
 const positions = masonryPositions(heights, columns, parseFloat(style.columnGap));
 gallery.classList.add('masonry');
 cards.forEach((card, index) => {
  const position = positions[index];
  card.style.gridColumn = String(position.column);
  card.style.gridRow = `${position.row} / span ${position.span}`;
 });
}
function queueGalleryLayout() {
 if (galleryStopped || galleryFrame) return;
 if (!window.requestAnimationFrame) { layoutGallery(); return; }
 galleryFrame = window.requestAnimationFrame(() => { galleryFrame = 0; layoutGallery(); });
}
const galleryObserver = window.ResizeObserver ? new window.ResizeObserver(entries => {
 let changed = false;
 for (const entry of entries) {
  if (entry.target !== gallery) changed = true;
  else if (entry.contentRect.width !== galleryWidth) { galleryWidth = entry.contentRect.width; changed = true; }
 }
 if (changed) queueGalleryLayout();
}) : null;
function observeGallery() {
 galleryObserver?.disconnect();
 if (!gallery.hidden) {
  galleryObserver?.observe(gallery);
  [...gallery.children].forEach(card => galleryObserver?.observe(card));
  layoutGallery();
 }
}
for (const event of ['load', 'error', 'loadedmetadata']) gallery.addEventListener(event, queueGalleryLayout, true);
window.addEventListener('resize', queueGalleryLayout);
window.addEventListener('pagehide', () => { galleryStopped = true; galleryObserver?.disconnect(); if (galleryFrame) window.cancelAnimationFrame(galleryFrame); galleryFrame = 0; });
window.addEventListener('pageshow', () => { galleryStopped = false; observeGallery(); });
let storage;try{storage=window.localStorage;}catch{storage={getItem(){throw Error('storage');},setItem(){throw Error('storage');}};}
// 已生成的短视频排在内置作品之后；清单为空时不新增任何条目。
const clipItems=clips.map(clip=>clipItem(clip,builtinItems.find(item=>item.id===clip.source&&item.media==='image'))).filter(Boolean);
const catalogItems=[...builtinItems,...clipItems];
// 概况文字：按种类计数，如“2 段程序动效”“1 段视频模型成片、1 段程序动效”。
function clipKinds(){const model=clipItems.filter(item=>!item.motionClip).length,motion=clipItems.length-model;return [model?model+' 段视频模型成片':'',motion?motion+' 段程序动效':''].filter(Boolean).join('、');}
let items=[...catalogItems];
const ids=items.map(x=>x.id),store=makeStore(storage,ids);
const host=window.MofaFamilyHost;
const integrated=document.querySelector('meta[name="mofa-api-site"]')?.content==='same-origin';
let previewReturn;
const state={section:'discover',media:'all',type:'all',category:'all',style:'all',query:'',sort:'featured'};
const requestedItem=itemFromHash(location.hash,items);
const drafts=new Map();let selected=requestedItem&&isItemAvailable(requestedItem)?requestedItem:items[0],language='zh',toastTimer;
const headings={discover:['灵感，从这里开始。','找一个喜欢的效果，做成自己的作品。'],skills:['把方法，接进工作流。','从官方技能和模板开始，先准备好环境。'],plays:['好玩的，不止一个提示词。','把画面、镜头和工具串成一次创作。'],saved:['留给下一次的灵感。','收藏保存在当前浏览器，随时回来继续。'],history:['刚刚看过的灵感。','最近浏览只记录条目，不记录你的材料。']};
function toast(text){clearTimeout(toastTimer);$('toast').textContent=text;$('toast').hidden=false;toastTimer=setTimeout(()=>{$('toast').hidden=true;},4200);}
function persist(next){const ok=store.write(next);if(!ok)toast('浏览器存储不可用，本次收藏与浏览记录只在当前页保留。');return ok;}
function applyTheme(){if(!host)document.documentElement.dataset.theme=store.data.theme;const dark=document.documentElement.dataset.theme==='dark';decorate($('theme-toggle'),dark?'sun':'moon');$('theme-toggle').setAttribute('aria-label',dark?'切换到浅色主题':'切换到深色主题');}
function saveButton(button,item){const saved=store.data.favorites.includes(item.id);decorate(button,'bookmark');button.classList.toggle('saved',saved);button.setAttribute('aria-label',`${saved?'取消收藏':'收藏'}：${item.title}`);button.setAttribute('aria-pressed',String(saved));button.dataset.saveId=item.id;}
function toggleFavorite(item){const favorites=store.data.favorites.includes(item.id)?store.data.favorites.filter(x=>x!==item.id):[...store.data.favorites,item.id];const ok=persist({...store.data,favorites});renderCollection();saveButton($('bench-save'),selected);const restore=[...document.querySelectorAll('.card-save')].find(b=>b.dataset.saveId===item.id);if(restore)restore.focus({preventScroll:true});else if(state.section==='saved')document.querySelector('.card-open')?.focus({preventScroll:true});if(ok)toast(favorites.includes(item.id)?'已收藏在当前浏览器。':'已取消收藏。');}
function makeImage(art,title,thumb=true){const image=el('img');image.decoding='async';image.alt=title;if(artSizes[art]){image.width=artSizes[art].width;image.height=artSizes[art].height;}
 // 先挂事件再设 src：图片加载前显示柔和的占位，加载完成或失败后移除。
 const settle=()=>{const p=image.parentElement;if(p?.classList.contains('card-media')){p.classList.remove('loading');p.classList.add('loaded');}};
 image.addEventListener('load',settle);
 image.addEventListener('error',()=>{settle();const p=image.parentElement;image.hidden=true;if(p&&!p.querySelector('.image-fallback'))p.append(el('span','image-fallback','预览暂时不可用，提示词仍可使用。'));});
 image.src=art.startsWith('/api/v1/studio/gallery/')?art:`assets/${art}${thumb?'-thumb':''}.webp`;return image;}
function makeStoryboard(item, compact=false){
 const board=el('div',compact?'script-teaser':'storyboard');
 board.append(el('p','script-status','视频分镜 · 样片待补'));
 const shots=el('ol','shot-list');
 for(const [timing,action,camera] of item.shots){
  const shot=el('li');shot.append(el('strong','shot-time',timing),el('p','',action));
  if(!compact)shot.append(el('p','shot-camera',camera));
  shots.append(shot);
 }
 board.append(shots);return board;
}
function card(item){
 const article=el('article','art-card '+(item.type!=='prompt'?'skill-card ':'')+(selected.id===item.id?'selected':''));article.setAttribute('role','listitem');article.dataset.itemId=item.id;
 const open=el('button','card-open');open.type='button';open.setAttribute('aria-label','查看：'+item.title);open.setAttribute('aria-pressed',String(selected.id===item.id));
 // 卡片只显示首帧图片和播放标记，不创建 <video>，浏览图库不会下载任何视频。
 if(item.art||item.video){const media=el('div','card-media '+item.ratio);if(item.art){media.classList.add('loading');const img=makeImage(item.art,'');img.loading='lazy';media.append(img);}else{const video=el('video');video.src=item.video;video.muted=true;video.preload='none';video.playsInline=true;media.append(video);}media.append(el('span','card-type',item.community?'用户投稿':item.clip?(item.clip.kind==='motion'?'程序动效':'短视频'):item.media==='video'?'视频作品':'原创图像'));if(item.video||item.clip){const play=el('span','card-play');play.append(icon('play'));media.append(play);if(item.clip)media.append(el('span','card-duration',item.clip.duration+' 秒'));}else media.append(decorate(el('span','card-hover'),'search','看大图和提示词'));open.append(media);}
 else if(item.shots)open.append(makeStoryboard(item,true));
 else{const visual=el('div','skill-visual');visual.append(icon(item.type==='skill'?'code':'sliders'),el('span','skill-tool',item.tool));open.append(visual);}
 const caption=el('div','card-caption');caption.append(el('h3','',item.title));if(!item.art&&!item.shots)caption.append(el('p','skill-summary',item.summary));caption.append(el('p','',item.category+' · '+(item.shots?'原创分镜':typeLabels[item.type])));open.append(caption);
 open.addEventListener('click',()=>{select(item);if(item.type==='prompt'){previewReturn=item.id;showLarge({play:Boolean(item.clip)});}else $('workbench').focus({preventScroll:true});});
 const favorite=el('button','icon-button card-save');favorite.type='button';saveButton(favorite,item);favorite.addEventListener('click',()=>toggleFavorite(item));article.append(open,favorite);
 if(item.type==='prompt'){
  // 短视频的提示词就是实际发送原文，没有可替换材料；按钮与提示随之改写。
  const quick=el('div','card-quick'),copy=el('button','card-copy',item.clip?(item.clip.kind==='motion'?'参考提示词':'复制提示词'):'复制模板'),use=el('button','card-use',item.clip?'看用法':'改成我的');
  copy.type=use.type='button';copy.setAttribute('aria-label',(item.clip?'复制视频提示词：':'复制模板：')+item.title);use.setAttribute('aria-label',(item.clip?'查看用法：':'改成我的：')+item.title);
  copy.addEventListener('click',()=>copyText(compile(item,drafts.get(item.id)||{},language),item.clip?(item.clip.kind==='motion'?'已复制图生视频参考提示词；这段动态不是用它生成的。':'已复制这段视频实际使用的提示词。'):'已复制复用模板；不是原始生成记录。',copy));
  use.addEventListener('click',()=>select(item,{focus:true}));quick.append(copy,use);article.append(quick);
 }
 return article;
}
function baseItems(){return findItems(items,{...state,query:'',category:'all',style:'all',media:'all',type:'all'},store.data);}
function renderFilters(){
 const base=baseItems();$('media-filters').replaceChildren();
 for(const [value,label,glyph]of[['all','全部',null],['image','图片','image'],['video',clipItems.length&&!VIDEO_CONTENT_ENABLED?'短视频':'视频与分镜','video']]){const b=el('button');const closed=value==='video'&&!VIDEO_CONTENT_ENABLED&&!clipItems.length;if(glyph)b.append(icon(glyph));b.append(document.createTextNode(label),el('span','filter-count',closed?'待开放':String(base.filter(x=>value==='all'||x.media===value).length)));b.setAttribute('aria-pressed',String(state.media===value));b.dataset.media=value;b.disabled=closed;if(closed){b.title='视频与分镜暂未开放';b.setAttribute('aria-describedby','video-availability');}b.addEventListener('click',()=>{if(closed)return;state.media=value;state.category='all';state.style='all';renderCollection();document.querySelector(`[data-media="${value}"]`)?.focus({preventScroll:true});});$('media-filters').append(b);}
 const categories=[...new Set(base.filter(x=>(state.media==='all'||x.media===state.media)&&(state.type==='all'||x.type===state.type)).map(x=>x.category))];if(state.category!=='all'&&!categories.includes(state.category))state.category='all';$('category-filters').replaceChildren();
 for(const value of['all',...categories]){const b=el('button','',value==='all'?'全部用途':value);b.setAttribute('aria-pressed',String(state.category===value));b.dataset.category=value;b.addEventListener('click',()=>{state.category=value;renderCollection();[...document.querySelectorAll('[data-category]')].find(n=>n.dataset.category===value)?.focus({preventScroll:true});});$('category-filters').append(b);}
 const styles=[...new Set(base.filter(item=>(state.media==='all'||item.media===state.media)&&(state.category==='all'||item.category===state.category)).flatMap(item=>item.styles||[]))];
 if(state.style!=='all'&&!styles.includes(state.style))state.style='all';
 $('style-filters').replaceChildren();$('style-filters').parentElement.hidden=state.section==='skills';
 for(const value of['all',...styles]){const button=el('button','',value==='all'?'全部风格':value);button.dataset.style=value;button.setAttribute('aria-pressed',String(state.style===value));button.addEventListener('click',()=>{state.style=value;renderCollection();[...document.querySelectorAll('[data-style]')].find(node=>node.dataset.style===value)?.focus({preventScroll:true});});$('style-filters').append(button);}
}
function renderPlays(){
 $('plays-list').replaceChildren();for(const play of plays){const article=el('article','play-story'),img=makeImage(play.art,'');img.loading='lazy';const content=el('div','play-story-content');const hasClosed=play.steps.some(step=>!isItemAvailable(items.find(item=>item.id===step.id)));content.append(el('h2','',play.title),el('p','',hasClosed?'视频与分镜待开放；当前可先使用图片和工具指引。':play.summary));const list=el('ol');play.steps.forEach((step,index)=>{const target=items.find(item=>item.id===step.id),closed=!isItemAvailable(target),li=el('li'),b=el('button'),label=el('span');label.append(el('span','step-number',String(index+1)),document.createTextNode(step.label));b.append(label,closed?el('span','fine-print','待开放'):icon('arrow'));b.disabled=closed;if(closed)b.title='视频与分镜暂未开放';else b.addEventListener('click',()=>select(target,{focus:true}));li.append(b);list.append(li);});content.append(list);article.append(img,content);$('plays-list').append(article);}
}
// 只有结果集合真正变化（切换区域、筛选、搜索、排序）时才播放卡片入场动画；收藏、选中等重绘不重复播放。
let revealSignature='';
function reveal(visible){
 const signature=[state.section,state.media,state.type,state.category,state.style,state.query,state.sort,visible.map(item=>item.id).join(',')].join('|');
 if(signature===revealSignature||reducedMotion()){revealSignature=signature;return;}
 revealSignature=signature;
 [...$('gallery').children].slice(0,18).forEach((node,index)=>{node.classList.add('reveal');node.style.setProperty('--reveal-delay',index*28+'ms');});
}
function updateChip(row){
 const max=row.scrollWidth-row.clientWidth;
 row.classList.toggle('fade-start',row.scrollLeft>2);
 row.classList.toggle('fade-end',max>2&&row.scrollLeft<max-2);
}
function updateChips(){document.querySelectorAll('.chip-row').forEach(updateChip);}
document.querySelectorAll('.chip-row').forEach(row=>{
 row.addEventListener('scroll',()=>updateChip(row),{passive:true});
 // 竖向滚轮在筛选条上改为横向滚动；已经到头时交还页面滚动。
 row.addEventListener('wheel',event=>{
  if(Math.abs(event.deltaY)<=Math.abs(event.deltaX))return;
  const before=row.scrollLeft;row.scrollLeft+=event.deltaY;
  if(row.scrollLeft!==before){event.preventDefault();updateChip(row);}
 },{passive:false});
});
function renderCollection(){
 const [title,description]=headings[state.section];$('page-title').textContent=title;$('page-description').textContent=description;$('hero-stats').hidden=state.section!=='discover';document.querySelectorAll('[data-section]').forEach(b=>{if(b.dataset.section===state.section)b.setAttribute('aria-current','page');else b.removeAttribute('aria-current');});$('saved-count').textContent=String(store.data.favorites.filter(id=>items.some(item=>item.id===id&&isItemAvailable(item))).length);$('video-availability').hidden=VIDEO_CONTENT_ENABLED||Boolean(clipItems.length);$('clip-cache').hidden=!clipItems.length||['skills','plays'].includes(state.section);
 const isPlays=state.section==='plays';$('catalog-controls').hidden=isPlays;$('gallery').hidden=isPlays;$('plays-list').hidden=!isPlays;$('empty-state').hidden=true;$('collection-note').textContent=isPlays?'组合方法 · 等待实测':state.section==='skills'?'官方出处 · 未安装运行':VIDEO_CONTENT_ENABLED?'原创图像与对应提示词 · 视频分镜待样片':host?.embedded?'原创图像与对应提示词 · 可用自己的模型生成':'原创图像与对应提示词';
 if(isPlays){observeGallery();$('result-count').textContent=`${plays.length} 个创作玩法`;renderPlays();revealSignature='plays';return;}
 renderFilters();updateChips();const visible=findItems(items,state,store.data);$('result-count').textContent=`${visible.length} 份${state.section==='skills'?'技能与工作流':'创作灵感'}`;$('gallery').replaceChildren(...visible.map(card));$('gallery').setAttribute('role','list');
 reveal(visible);
 observeGallery();
 if(!visible.length){$('empty-state').hidden=false;const emptyCollection=['saved','history'].includes(state.section)&&!state.query&&state.media==='all'&&state.category==='all'&&state.type==='all';$('empty-title').textContent=emptyCollection?(state.section==='saved'?'把喜欢的灵感留下来':'从一个灵感开始'):'暂时没有找到';$('empty-description').textContent=emptyCollection?(state.section==='saved'?'点击条目上的收藏按钮，它会出现在这里。':'打开任意条目，就能在这里继续找回。'):'试试更短的关键词，或者清除筛选看看全部内容。';$('empty-action').textContent=emptyCollection?'去发现灵感':'清除筛选';$('empty-action').onclick=()=>{if(emptyCollection)setSection('discover');else resetFilters();};}
}
function resetFilters(){state.media='all';state.category='all';state.style='all';state.type='all';state.query='';$('search').value='';$('resource-type').value='all';renderCollection();}
function setSection(section){state.section=section;resetFilters();document.querySelector(`[data-section="${section}"]`)?.focus({preventScroll:true});}
function values(){if(!drafts.has(selected.id))drafts.set(selected.id,Object.fromEntries(selected.fields.map(f=>[f.key,f.value])));return drafts.get(selected.id);}
function updatePrompt(){$('prompt-output').value=compile(selected,values(),language);if($('preview-prompt-output'))$('preview-prompt-output').value=$('prompt-output').value;}
function select(item,options={}){if(!item)return;if(!isItemAvailable(item)){toast('视频与分镜暂未开放，可以先从图片找灵感。');return;}pauseVideos();selected=item;persist({...store.data,history:[item.id,...store.data.history.filter(x=>x!==item.id)]});const hash=new URLSearchParams({item:item.id}).toString();if(location.hash!==`#${hash}`)history[options.replace?'replaceState':'pushState'](null,'',`#${hash}`);renderCollection();renderBench();if(options.focus){$('workbench').focus({preventScroll:true});if(window.innerWidth<=900)$('workbench').scrollIntoView({behavior:window.matchMedia('(prefers-reduced-motion: reduce)').matches?'instant':'smooth',block:'start'});}}
function pauseVideos(except){document.querySelectorAll('video').forEach(video=>{if(video!==except)video.pause();});}
function makeVideo(item){const video=el('video','film-player');video.src=item.video;if(item.art)video.poster=item.community?item.art:'assets/'+item.art+'.webp';video.controls=true;video.playsInline=true;video.preload='metadata';video.setAttribute('aria-label',item.title+'，'+(item.community?'用户投稿成片':'授权影片节选'));video.addEventListener('play',()=>pauseVideos(video));video.addEventListener('error',()=>toast('视频暂不可用，可能已下架；请刷新投稿目录。'));return video;}
function attribution(item){const line=el('p','film-credit');line.append(document.createTextNode(item.credit+' · '+item.license+(item.community?'':' · 节选已裁短、缩放、静音。')));return line;}
// 图生短视频：先只显示首帧，用户点击播放才下载；下载后存进当前浏览器，再看直接从本机读取，不再占用服务器流量。
let activeClip=null;
function stopClip(){if(!activeClip)return;const clip=activeClip;activeClip=null;clip.controller.abort();clip.video?.pause();clip.source?.dispose();}
function clipMeta(item){return item.clip.duration+' 秒 · '+formatBytes(item.clip.bytes);}
function clipPlayer(item,{autoplay=false}={}){
 const box=el('div','clip-player '+item.ratio),start=el('button','clip-start'),badge=el('span','clip-badge'),meta=el('span','clip-meta',clipMeta(item));
 start.type='button';start.setAttribute('aria-label','播放短视频：'+item.title+'（'+clipMeta(item)+'）');badge.append(icon('play'));
 start.append(makeImage(item.art,item.title,false),badge,meta);
 const progress=el('progress','clip-progress');progress.max=100;progress.value=0;progress.hidden=true;progress.setAttribute('aria-label','视频下载进度');
 const note=el('p','fine-print clip-note','点击播放才会下载，约 '+formatBytes(item.clip.bytes)+'；看过一次会保存在当前浏览器，再看不重复下载。');
 box.append(start,progress);
 isClipCached(item.clip.file,{win:window}).then(hit=>{if(hit&&box.isConnected&&!box.classList.contains('playing')){meta.textContent=item.clip.duration+' 秒 · 已在本机';note.textContent='这段视频已保存在当前浏览器，播放不会再下载。';}});
 const play=async()=>{
  if(box.classList.contains('loading')||box.classList.contains('playing'))return;
  stopClip();
  const record={controller:new AbortController(),box};activeClip=record;
  box.classList.add('loading');start.setAttribute('aria-busy','true');progress.hidden=false;meta.textContent='正在加载…';
  try{
   const source=await loadClip(item.clip.file,{win:window,signal:record.controller.signal,onProgress:(received,total)=>{if(total){progress.value=Math.min(100,Math.round(received/total*100));meta.textContent='正在加载 '+progress.value+'%';}else meta.textContent='正在加载 '+formatBytes(received);}});
   if(activeClip!==record||!box.isConnected){source.dispose();return;}
   record.source=source;
   const video=el('video','film-player clip-video');video.src=source.src;video.poster='assets/'+item.art+'.webp';video.controls=true;video.playsInline=true;video.preload='auto';video.loop=Boolean(item.clip.loop);
   video.setAttribute('aria-label',item.title+'，图生短视频');
   video.addEventListener('play',()=>pauseVideos(video));
   video.addEventListener('error',()=>toast('视频无法播放，请刷新后重试。'));
   record.video=video;start.replaceWith(video);progress.hidden=true;box.classList.remove('loading');box.classList.add('playing');
   note.textContent=source.fromCache?'从本机缓存播放，没有重新下载。':source.cached?'已下载并保存在当前浏览器，下次直接从本机播放。':'已加载。当前浏览器不支持本机缓存，下次可能需要重新下载。';
   video.focus({preventScroll:true});
   video.play()?.catch(()=>toast('点击视频上的播放按钮开始观看。'));
   updateClipCache();
  }catch(error){
   if(activeClip===record)activeClip=null;
   box.classList.remove('loading');start.removeAttribute('aria-busy');progress.hidden=true;meta.textContent=clipMeta(item);
   if(error?.name!=='AbortError')toast(error?.message||'视频暂时无法加载。');
  }
 };
 start.addEventListener('click',play);
 if(autoplay)queueMicrotask(play);
 return {box,note};
}
function updateClipCache(){
 if(!clipItems.length)return;
 const {count,bytes}=clipCacheSummary({win:window}),supported=Boolean(window.caches?.open);
 $('clip-cache-text').textContent=clipKinds()+'，点开才播放。'+(count?'本机已缓存 '+count+' 段（'+formatBytes(bytes)+'），再看不重复下载。':supported?'看过的会保存在当前浏览器，再看不重复下载。':'当前浏览器不支持本机缓存，每次播放可能重新下载。');
 $('clip-clear').hidden=!count;
}
function originalRecord(item, expanded=false){
 const record=el('details','original-record');record.open=expanded;
 const summary=el('summary','',item.rawPrompt?'原始提示词 · 实际生成记录':item.community?'反推 / 整理参考模板':'原始提示词记录未保存');record.append(summary);
 if(item.rawPrompt){
  const area=el('textarea','original-prompt');area.value=item.rawPrompt;area.readOnly=true;area.rows=7;area.setAttribute('aria-label',item.title+'原始提示词');
  const copy=el('button','secondary-button','复制原始提示词');copy.type='button';copy.dataset.copyKind='original';copy.addEventListener('click',()=>copyText(item.rawPrompt,'已复制这份作品的原始提示词。'));
  record.append(el('p','fine-print',item.recordNote||'保留实际发送的英文原文；下方复用模板是另写的版本，不保证重现完全相同的图。'),area,copy);
  if(item.generation)record.append(el('p','generation-note',item.generation.tool+' · '+item.generation.date+' · '+item.generation.framing+'。'+item.generation.notes));
 }else record.append(el('p','fine-print',item.recordNote||'这张旧图没有完整原始生成记录，只提供复用模板；不将后写的提示词标成原始提示词。'));
 return record;
}
// 大图翻页：沿当前筛选结果顺序浏览；不在结果里时（例如从玩法打开）改用全部可浏览作品。
function navItems(){
 const visible=state.section==='plays'?[]:findItems(items,state,store.data);
 const media=item=>isItemAvailable(item)&&item.type==='prompt'&&(item.art||item.video);
 const list=visible.filter(media);
 return list.some(item=>item.id===selected.id)?list:findItems(items,{...state,section:'discover',media:'all',type:'all',category:'all',style:'all',query:''},store.data).filter(media);
}
function stepPreview(offset){
 const list=navItems(),index=list.findIndex(item=>item.id===selected.id);
 if(list.length<2||index<0)return;
 const next=list[(index+offset+list.length)%list.length];
 select(next,{replace:true});previewReturn=next.id;showLarge({focus:offset<0?'preview-prev':'preview-next'});
}
function shuffle(){
 const pool=(state.section==='plays'?[]:findItems(items,state,store.data)).filter(item=>item.type==='prompt'&&item.art&&isItemAvailable(item));
 const candidates=(pool.length>1?pool:findItems(items,{...state,section:'discover',media:'image',type:'all',category:'all',style:'all',query:''},store.data)).filter(item=>item.art&&item.id!==selected.id);
 if(!candidates.length)return;
 const next=candidates[Math.floor(Math.random()*candidates.length)];
 select(next);previewReturn=$('shuffle-button');showLarge();
}
function showLarge(options={}){
 stopClip();pauseVideos();const dialog=$('preview-dialog'),media=$('preview-media'),details=$('preview-details');media.replaceChildren();details.replaceChildren();
 let clipNote=null;
 // 只有从卡片或工作区明确点开这段视频时才自动开始；大图里左右翻页只显示首帧，不会顺带下载。
 if(selected.clip){const player=clipPlayer(selected,{autoplay:Boolean(options.play)});media.append(player.box);clipNote=player.note;}
 else if(selected.video){const video=makeVideo(selected);media.append(video,attribution(selected));}
 else if(selected.art){const image=makeImage(selected.art,selected.title,false);image.id='preview-image';media.append(image);}
 else if(selected.shots)media.append(makeStoryboard(selected));
 $('preview-caption').textContent=selected.title+' · '+(selected.shots?'原创分镜，样片待补':selected.motionClip?'程序动效与参考提示词':selected.clip?'图生短视频与提示词':'作品与提示词');
 details.append(el('h3','',selected.title),el('p','fine-print',selected.clip?selected.summary:selected.shots||selected.community?selected.requirements:'原创AI图像；可以看原始记录，也可以替换内容复用。'));
 if(clipNote)details.append(clipNote);
 if(selected.media==='image'||selected.community)details.append(originalRecord(selected,true));
 const promptLabel=el('label','preview-prompt-label',selected.shots?'完整视频分镜提示词':selected.clip?(selected.motionClip?'图生视频参考提示词（这段动态未使用）':'视频提示词（实际使用）'):'可替换复用模板');promptLabel.htmlFor='preview-prompt-output';
 const prompt=el('textarea');prompt.id='preview-prompt-output';prompt.value=compile(selected,values(),language);prompt.readOnly=true;prompt.rows=8;
 const copy=el('button','primary-button',selected.shots?'复制分镜提示词':selected.clip?(selected.motionClip?'复制参考提示词':'复制视频提示词'):'复制我的模板');copy.type='button';copy.id='preview-copy';copy.addEventListener('click',()=>copyText(prompt.value,'提示词已复制；当前页面未运行生成模型。',copy));
 // 浏览器会在稍后派发 close 事件并按 previewReturn 还原焦点，所以离开大图前先改写返回目标。
 const use=el('button','secondary-button','替换内容，做成我的');use.type='button';use.addEventListener('click',()=>{previewReturn=$('workbench');dialog.close();$('workbench').focus({preventScroll:true});if(window.innerWidth<=900)$('workbench').scrollIntoView({behavior:'instant',block:'start'});});
 if(selected.clip){
  // 短视频没有可替换材料：提供复制视频提示词，并能回到首帧来源作品看图片提示词。
  const origin=items.find(item=>item.id===selected.sourceId),back=el('button','secondary-button','查看来源作品与图片提示词');back.type='button';back.hidden=!origin;
  back.addEventListener('click',()=>{if(!origin)return;select(origin,{replace:true});previewReturn=origin.id;showLarge({focus:'preview-copy'});});
  details.append(promptLabel,prompt,el('p','fine-print',selected.recordNote),copy,back);
 }else details.append(promptLabel,prompt,el('p','fine-print','模板使用当前右侧材料和语言；修改材料不改动原始生成记录。'),copy,use);
 if(canGenerate(selected)){
  const jump=decorate(el('button','text-button generate-jump'),'wand',host?.embedded?'用我的模型生成':'登录控制台后可直接生成');jump.type='button';
  jump.addEventListener('click',()=>{const wanted=!host?.embedded?$('generate-link'):gen.mode==='choosing'?$('generate-key'):$('generate-run'),target=wanted.disabled?$('workbench'):wanted;previewReturn=target;dialog.close();$('generate-panel').scrollIntoView?.(window.innerWidth<=900||reducedMotion()?{behavior:'instant',block:'start'}:{behavior:'smooth',block:'nearest'});target.focus({preventScroll:true});});
  details.append(jump);
 }
 const list=navItems(),index=list.findIndex(item=>item.id===selected.id),browsable=list.length>1&&index>=0;
 $('preview-prev').hidden=$('preview-next').hidden=!browsable;
 $('preview-position').textContent=browsable?`${index+1} / ${list.length}`:'';
 const opening=!dialog.open;
 if(opening)dialog.showModal();
 dialog.scrollTop=0;
 ($(options.focus)&&!$(options.focus).hidden?$(options.focus):$('close-preview')).focus({preventScroll:true});
 if(selected.video&&!selected.clip&&opening)media.querySelector('video').play()?.catch(()=>toast('点击视频播放按钮开始观看。'));
}
function renderBench(){
 pauseVideos();$('bench-content').scrollTop=0;$('selected-title').textContent=selected.title;$('selected-description').textContent=selected.summary;$('bench-badges').replaceChildren(el('span','',typeLabels[selected.type]),el('span','',selected.status));saveButton($('bench-save'),selected);$('bench-preview').replaceChildren();
 if(selected.shots&&!selected.video){const button=el('button','preview-button script-preview');button.type='button';button.setAttribute('aria-label','查看分镜：'+selected.title);button.append(makeStoryboard(selected));button.addEventListener('click',()=>{previewReturn=button;showLarge();});$('bench-preview').append(button);}
 else if(selected.clip){
  // 工作区只放首帧与时长，点击后在大图里播放；这里不创建 <video>。
  const button=el('button','preview-button clip-poster'),badge=el('span','clip-badge');button.type='button';badge.append(icon('play'));
  button.setAttribute('aria-label','播放短视频：'+selected.title+'（'+clipMeta(selected)+'）');
  button.append(makeImage(selected.art,selected.title,false),badge,el('span','clip-meta',clipMeta(selected)));
  button.addEventListener('click',()=>{previewReturn=button;showLarge({play:true});});
  $('bench-preview').append(button,el('p','fine-print',(selected.motionClip?'程序动效 · ':'图生短视频 · ')+'点击播放才下载，看过的保存在当前浏览器'));
 }
 else if(selected.art||selected.video){
  if(selected.media==='video'){$('bench-preview').append(makeVideo(selected),attribution(selected));}
  else{const button=el('button','preview-button');button.setAttribute('aria-label','放大预览：'+selected.title);button.append(makeImage(selected.art,selected.title,false));button.addEventListener('click',()=>{previewReturn=button;showLarge();});$('bench-preview').append(button,el('p','fine-print',selected.community?'用户授权投稿 · 点击查看完整作品':'原创AI视觉示例 · 点击查看完整大图'));}
 }else{const visual=el('div','bench-skill'),text=el('div');text.append(el('strong','',selected.tool),el('span','',selected.type==='skill'?'官方 Agent Skill':'官方模板使用指引'));visual.append(icon(selected.type==='skill'?'code':'sliders'),text);$('bench-preview').append(visual);}
 $('bench-origin').replaceChildren();if((selected.media==='image'||selected.community)&&selected.type==='prompt')$('bench-origin').append(originalRecord(selected));
 for(const id of ['language-zh','language-en']){$(id).disabled=!!selected.community||!!selected.clip;$(id).title=selected.community?'投稿提示词保持作者原文，不自动翻译':selected.clip?'视频提示词保持实际发送的英文原文':'';}
 $('material-form').hidden=!selected.fields.length;
 $('prompt-note').textContent=selected.community?'保持投稿人原文，不自动翻译或替换占位符。可复制后修改，模型效果不由人工审核保证。':selected.motionClip?'这段是程序动效，没有使用模型提示词；这里是参考提示词，可配合来源作品图片在图生视频工具里试。':selected.clip?'这是生成这段视频时实际发送的提示词；在图生视频工具里配合来源作品图片作为首帧使用。':'复用模板不等于原始生成记录。空白使用示例值；你的替换内容只留在当前页面。';
 document.querySelector('.source-note').textContent=selected.community?'作品、模型和提示词由投稿人提供；人工审核不等于模型效果实测。':selected.motionClip?'按工坊原创作品逐帧计算局部动态并录制，不是视频模型生成。':selected.clip?'首帧来自工坊原创作品；视频由模型生成，未做逐帧人工修饰。':'方法参考官方指南；原创模板尚未逐条实测。';
 const current=values();$('material-fields').replaceChildren();for(const f of selected.fields){const label=el('label','material-field',f.label),input=el('input');input.id=`material-${f.key}`;input.name=f.key;input.value=current[f.key];input.maxLength=300;input.autocomplete='off';input.addEventListener('input',()=>{current[f.key]=input.value;updatePrompt();});label.append(input);$('material-fields').append(label);}updatePrompt();decorate($('copy-prompt'),'copy',selected.type==='prompt'?'复制完整提示词':'复制使用提示词');$('usage-steps').replaceChildren(...selected.steps.map(s=>el('li','',s)));$('usage-requirements').textContent=selected.requirements;$('usage-tip').textContent=selected.tip;$('skill-command').replaceChildren();
 if(selected.command){const code=el('code','command',selected.command),b=el('button','text-button command-button','复制官方安装命令');b.type='button';b.addEventListener('click',()=>copyText(selected.command,'命令已复制，请在自己的环境检查后运行。'));$('skill-command').append(code,b);}
 $('source-link').href=selected.source.url;decorate($('source-link'),selected.clip?'arrow':'external',selected.source.label);
 // 短视频的来源是本页里的作品，在当前页切换；其他条目的官方出处仍在新窗口打开。
 if(selected.clip){$('source-link').removeAttribute('target');$('source-link').rel='noopener';}else{$('source-link').target='_blank';$('source-link').rel='noopener noreferrer';}$('workbench').querySelector('.instructions').open=false;document.querySelector('.inline-confirm')?.remove();document.querySelector('.manual-copy')?.remove();
 renderGenerate();
}
// —— 直接生图：控制台内选择本人密钥与模型，点击生成才调用；结果只留在本页内存。
const gen={mode:'idle',keys:[],models:[],connection:null,label:'',pending:null,results:[],version:0,leaving:false,source:'template',size:'',tick:0};
let generateLinkBase;
// 一体化部署指向同站控制台工作区；独立打开时保留官方站地址。
function generateBase(){if(generateLinkBase===undefined){const link=$('generate-link');if(integrated)link.setAttribute('href','/tools/studio');generateLinkBase=link.getAttribute('href');}return generateLinkBase;}
function canGenerate(item){return Boolean(item)&&item.type==='prompt'&&item.media==='image'&&Boolean(item.art)&&isItemAvailable(item);}
function generationPrompt(){return gen.source==='raw'&&selected.rawPrompt?selected.rawPrompt:compile(selected,values(),language);}
function renderGenerate(){
 const panel=$('generate-panel');panel.hidden=!canGenerate(selected);if(panel.hidden)return;
 const embedded=Boolean(host?.embedded),ready=embedded&&gen.mode==='ready'&&gen.connection;
 $('generate-setup').hidden=!(embedded&&gen.mode==='choosing');
 $('generate-ready').hidden=!embedded||gen.mode==='choosing';
 $('generate-link').hidden=embedded;
 $('generate-chip').hidden=!ready;$('generate-chip').textContent=ready?gen.label:'';
 if(!embedded){
  $('generate-intro').textContent='在魔法 API 控制台打开工坊，就能选择自己的生图密钥和模型，在这里直接生成。当前页替换的材料不会带过去。';
  const base=generateBase(),link=$('generate-link'),url=new URL(base,location.href),external=/^https?:/.test(base);url.searchParams.set('item',selected.id);
  link.href=external?url.href:url.pathname+url.search;
  if(external){link.target='_blank';link.rel='noopener noreferrer';link.title='在新窗口打开魔法 API 控制台';}else{link.removeAttribute('target');link.rel='noopener';link.removeAttribute('title');}
  decorate(link,'wand','登录控制台，用我的模型生成');
 } else {
  $('generate-intro').textContent=ready?'点击生成会把下方提示词发送到所选模型，按该密钥所属分组计费；不会自动生成或重试。':'选择自己的生图密钥和模型，点击生成才会调用，按该密钥所属分组计费。';
  const sources=[['template','我的模板（当前材料）']];if(selected.rawPrompt)sources.push(['raw','原始提示词（实际生成记录）']);
  $('generate-source').replaceChildren(...sources.map(([value,label])=>option(label,value)));
  if(!sources.some(([value])=>value===gen.source))gen.source='template';$('generate-source').value=gen.source;
  $('generate-size').replaceChildren(...IMAGE_SIZES.map(size=>option(size.label,size.value)));$('generate-size').value=gen.size;
  const suggestion=suggestedSize(selected.ratio);
  $('generate-size-hint').textContent=`原作是${selected.ratio==='portrait'?'竖幅':selected.ratio==='wide'?'横幅':'方形'}，想接近原作可选 ${suggestion.replace('x','×')}；不同模型支持的尺寸不同。`;
  const run=$('generate-run');run.disabled=Boolean(gen.pending);
  if(gen.pending)decorate(run,'spark',gen.pending.itemId===selected.id?`生成中 · ${gen.tick} 秒`:'另一张正在生成…');
  else decorate(run,ready?'wand':'key',ready?'生成一张':'选择生图模型');
  $('generate-stop').hidden=!gen.pending;
  $('generate-change').hidden=!ready||Boolean(gen.pending);
  $('generate-source').disabled=$('generate-size').disabled=Boolean(gen.pending)||!ready;
 }
 renderResults();
}
function renderResults(){
 const list=gen.results.filter(result=>result.itemId===selected.id);
 $('generate-results').replaceChildren(...list.map((result,index)=>{
  const figure=el('figure','result-card'),open=el('button','result-open'),image=el('img');
  open.type='button';open.setAttribute('aria-label',`查看生成结果 ${index+1}：${selected.title}`);
  image.src=result.src;image.alt=`${selected.title} 的生成结果 ${index+1}`;image.referrerPolicy='no-referrer';image.decoding='async';
  image.addEventListener('error',()=>{image.hidden=true;if(!open.querySelector('.image-fallback'))open.append(el('span','image-fallback','图片无法显示，远程地址可能已过期。'));});
  open.append(image);open.addEventListener('click',()=>showResult(result,index,open));
  const caption=el('figcaption','',`${result.model} · ${result.time}`);
  figure.append(open,caption);return figure;
 }));
 if(list.length)$('generate-results').append(el('p','fine-print','生成结果只留在当前页面，刷新或退出登录后清除，请及时下载。'));
}
let resultReturn;
function showResult(result,index,trigger){
 const dialog=$('result-dialog');if(typeof dialog.showModal!=='function'){toast('此浏览器不支持预览，请在图片上长按或右键保存。');return;}
 resultReturn=trigger;$('result-caption').textContent=`生成结果 ${index+1} · ${result.model}`;
 const image=$('result-image');image.src=result.src;image.alt=`${result.title} 的生成结果`;$('result-error').hidden=true;
 const name=downloadName(result.itemId,index,result.src),download=$('result-download');
 download.hidden=!name;$('result-save-note').hidden=Boolean(name);
 if(name){download.href=result.src;download.download=name;decorate(download,'download','下载原图');}else{download.removeAttribute('href');download.removeAttribute('download');}
 $('result-revised').hidden=!result.revised;$('result-revised').textContent=result.revised?'服务改写后的提示词：'+result.revised:'';
 dialog.showModal();$('close-result').focus({preventScroll:true});
}
function setGenerateStatus(text){$('generate-status').textContent=text;}
async function chooseModel(){
 const version=++gen.version;gen.mode='choosing';gen.leaving=false;decorate($('generate-manage'),'key','管理密钥');renderGenerate();
 $('generate-key').replaceChildren(option('正在读取可用密钥…',''));$('generate-model').replaceChildren(option('先选择密钥',''));
 $('generate-key').disabled=$('generate-model').disabled=$('generate-apply').disabled=true;setGenerateStatus('正在读取你的密钥…');
 try{
  const keys=await host.request('keys',{kind:'image'});if(version!==gen.version)return;
  gen.keys=Array.isArray(keys)?keys:[];
  $('generate-key').replaceChildren(option(gen.keys.length?'请选择密钥':'没有可用于生图的密钥',''),...gen.keys.map(key=>option(`${key.name} · ${key.group||'未命名分组'}`,String(key.id))));
  $('generate-key').disabled=!gen.keys.length;
  setGenerateStatus(gen.keys.length?'选择密钥后读取它的模型目录。':'没有可用于生图的有效密钥。请在密钥管理确认密钥启用、未过期、额度可用，且分组已开启生图权限。');
  if(gen.keys.length)$('generate-key').focus({preventScroll:true});else $('generate-manage').focus({preventScroll:true});
 }catch(error){if(version===gen.version){$('generate-key').replaceChildren(option('暂时无法读取密钥',''));setGenerateStatus(error.message);}}
}
async function readModels(){
 const version=++gen.version,keyID=Number($('generate-key').value);
 $('generate-model').replaceChildren(option('正在读取模型…',''));$('generate-model').disabled=$('generate-apply').disabled=true;
 if(!keyID){$('generate-model').replaceChildren(option('先选择密钥',''));setGenerateStatus('');return;}
 try{
  const result=await host.request('models',{kind:'image',keyID});if(version!==gen.version)return;
  gen.models=Array.isArray(result?.models)?result.models:[];
  const likely=gen.models.filter(model=>/image|dall|imagen|flux|seedream|kolors|midjourney|sd|diffusion/i.test(model)),others=gen.models.filter(model=>!likely.includes(model));
  const group=(label,list)=>{const node=document.createElement('optgroup');node.label=label;node.append(...list.map(model=>option(model,model)));return node;};
  $('generate-model').replaceChildren(option(gen.models.length?'请选择生图模型':'这个密钥没有返回模型',''));
  if(likely.length)$('generate-model').append(group('看起来支持生图',likely));
  if(others.length)$('generate-model').append(group(likely.length?'其他模型（不一定支持生图）':'目录中的模型（请确认支持生图）',others));
  $('generate-model').disabled=!gen.models.length;
  setGenerateStatus(gen.models.length?'模型目录来自所选密钥；目录不代表能力已验证。':'这个密钥没有返回模型，请检查分组或换一个密钥。');
 }catch(error){if(version===gen.version){$('generate-model').replaceChildren(option('目录读取失败',''));setGenerateStatus(error.message);}}
}
async function applyModel(){
 const keyID=Number($('generate-key').value),model=$('generate-model').value;if(!keyID||!model)return;
 const version=++gen.version;$('generate-apply').disabled=true;setGenerateStatus('正在确认所选模型…');
 try{
  const result=await host.request('apply',{kind:'image',keyID,model,protocol:'chat'});if(version!==gen.version)return;
  gen.connection=imageConnection(result);
  const key=gen.keys.find(entry=>entry.id===keyID);gen.label=`${model} · ${key?.group||key?.name||'所选密钥'}`;
  gen.mode='ready';renderGenerate();
  setGenerateStatus('已选择 '+model+'。密钥只留在本页内存；点击“生成一张”才会调用并计费。');
  $('generate-run').focus({preventScroll:true});
 }catch(error){if(version===gen.version){$('generate-apply').disabled=false;setGenerateStatus(error.message);}}
}
async function runGenerate(){
 if(!host?.embedded)return;
 if(gen.mode!=='ready'||!gen.connection){chooseModel();return;}
 if(gen.pending||!canGenerate(selected))return;
 const item=selected,prompt=generationPrompt();
 const operation={controller:new AbortController(),itemId:item.id,timedOut:false,started:Date.now()};
 operation.timer=setTimeout(()=>{operation.timedOut=true;operation.controller.abort();},GENERATE_TIMEOUT_MS);
 operation.clock=setInterval(()=>{gen.tick=Math.floor((Date.now()-operation.started)/1000);if(selected.id===operation.itemId)decorate($('generate-run'),'spark',`生成中 · ${gen.tick} 秒`);},1000);
 gen.pending=operation;gen.tick=0;renderGenerate();setGenerateStatus('已发送到 '+gen.connection.model+'，正在等待图片…');
 try{
  const images=await generateImage(gen.connection,prompt,gen.size,operation.controller.signal,(url,options)=>window.fetch(url,options));
  if(gen.pending!==operation)return;
  const time=new Date().toLocaleTimeString('zh-CN',{hour:'2-digit',minute:'2-digit'});
  gen.results=[...images.map(image=>({...image,itemId:item.id,title:item.title,model:gen.connection.model,time})),...gen.results].slice(0,12);
  if(selected.id===item.id)setGenerateStatus('已生成，可点开查看或下载。');else toast(`「${item.title}」已生成，回到该作品即可查看。`);
 }catch(error){
  if(gen.pending!==operation)return;
  const message=operation.controller.signal.aborted?(operation.timedOut?'等待超时。':'已停止等待。')+'服务端可能已开始处理并计费，请先检查用量记录。':error.message;
  if(selected.id===item.id)setGenerateStatus(message);else toast(`「${item.title}」生成失败：${message}`);
 }finally{
  clearTimeout(operation.timer);clearInterval(operation.clock);
  if(gen.pending===operation){gen.pending=null;renderGenerate();}
 }
}
function resetGeneration(message){
 gen.version++;gen.pending?.controller.abort();if(gen.pending){clearTimeout(gen.pending.timer);clearInterval(gen.pending.clock);}
 Object.assign(gen,{mode:'idle',keys:[],models:[],connection:null,label:'',pending:null,results:[],leaving:false});
 renderGenerate();setGenerateStatus(message||'');
}
// 与主站 useClipboard 一致：剪贴板接口被拒（未聚焦、非安全上下文、部分内置浏览器）时改用 execCommand；临时文本框放进已打开的模态框，避免被设为 inert。
async function writeClipboard(text){
 try{if(navigator.clipboard?.writeText){await navigator.clipboard.writeText(text);return true;}}catch{}
 const active=document.activeElement,area=document.createElement('textarea');
 area.value=text;area.readOnly=true;area.setAttribute('aria-hidden','true');area.style.cssText='position:fixed;left:0;top:0;width:1px;height:1px;opacity:0;pointer-events:none';
 (document.querySelector('dialog[open]')||document.body).append(area);
 try{area.focus({preventScroll:true});area.select();area.setSelectionRange(0,area.value.length);return document.execCommand('copy')===true;}catch{return false;}
 finally{area.remove();if(active?.isConnected&&typeof active.focus==='function')active.focus({preventScroll:true});}
}
async function copyText(text,success,button){if(await writeClipboard(text)){document.querySelector('.manual-copy')?.remove();toast(success);flashCopied(button);return true;}{document.querySelector('.manual-copy')?.remove();const box=el('div','manual-copy'),label=el('label','','手动复制文本'),area=el('textarea');area.value=text;area.readOnly=true;area.rows=4;label.append(area);box.append(label);($('preview-dialog').open?$('preview-details'):$('bench-content')).append(box);area.focus();area.select();toast('自动复制不可用。已选中文本，请按 Ctrl+C 或长按复制。');return false;}}
function resetMaterials(){if(selected.fields.every(f=>values()[f.key]===f.value)||document.querySelector('.inline-confirm'))return;const box=el('div','inline-confirm'),cancel=el('button','secondary-button','保留我的内容'),confirm=el('button','secondary-button','恢复示例');cancel.type=confirm.type='button';cancel.addEventListener('click',()=>{box.remove();$('reset-materials').focus();});confirm.addEventListener('click',()=>{drafts.delete(selected.id);renderBench();$('reset-materials').focus();toast('已恢复当前条目的示例材料。');});box.append(el('p','','恢复示例会覆盖当前条目的材料。'),cancel,confirm);$('material-form').append(box);cancel.focus();}
 for(const[id,glyph,label]of[['search-icon','search'],['empty-icon','search'],['more-filters','sliders','更多筛选'],['history-button','history','最近浏览'],['shuffle-button','shuffle','随机灵感'],['download-prompt','download'],['share-item','external'],['close-preview','close'],['preview-prev','prev'],['preview-next','next'],['close-result','close'],['close-about','close'],['return-to-library','arrow'],['generate-manage','key','管理密钥'],['generate-stop','stop','停止等待']])decorate($(id),glyph,label);
// 首屏概况只统计内置目录，不虚构浏览量或成功率。
(()=>{const images=builtinItems.filter(item=>item.type==='prompt'&&item.media==='image'&&item.art);const styles=new Set(images.flatMap(item=>item.styles||[]));
 $('hero-stats').replaceChildren(...[[images.length,'张原创 AI 作品'],[images.filter(item=>item.rawPrompt).length,'份实际生成记录'],[styles.size,'种视觉风格'],...(clipItems.length?[[clipItems.length,'段短视频']]:[]),[builtinItems.filter(item=>item.type!=='prompt').length,'份 Skills·工作流']].map(([count,label])=>{const li=el('li');li.append(el('strong','',String(count)),document.createTextNode(label));return li;}));})();
// 有图生短视频时同步改写说明文字；封存的分镜与旧视频仍不开放。
if(clipItems.length){
 const images=builtinItems.filter(item=>item.type==='prompt'&&item.media==='image'&&item.art).length;
 $('about-scope').textContent='目前可浏览'+images+'张AI图像和'+clipItems.length+'段由其中作品做成的短视频（'+clipKinds()+'）。程序动效只让灯光、水面等局部动起来，主体不动；短视频点开才下载，看过的保存在当前浏览器；原创分镜仍待开放。';
 $('advanced-scope').textContent='图片、短视频、Skills与工作流可浏览；原创分镜待开放。选择Skill或工作流会进入对应资源区。';
}
$('shuffle-button').addEventListener('click',shuffle);
$('clip-only').addEventListener('click',()=>{state.section='discover';state.media='video';state.category='all';state.style='all';state.type='all';state.query='';$('search').value='';renderCollection();document.querySelector('[data-media="video"]')?.focus({preventScroll:true});});
$('clip-clear').addEventListener('click',async()=>{stopClip();await clearClips({win:window});updateClipCache();toast('已清除本机缓存的短视频；再次播放会重新下载。');$('clip-only').focus({preventScroll:true});});
window.addEventListener('pagehide',stopClip);
// 启动时把已从清单移除（重新生成或下架）的旧片段从本机缓存删掉，再显示缓存概况。
if(clipItems.length)pruneClips(clipItems.map(item=>item.clip.file),{win:window}).finally(updateClipCache);
$('preview-prev').addEventListener('click',()=>stepPreview(-1));
$('preview-next').addEventListener('click',()=>stepPreview(1));
$('preview-dialog').addEventListener('keydown',event=>{if((event.key==='ArrowLeft'||event.key==='ArrowRight')&&!/INPUT|TEXTAREA|SELECT/.test(event.target.tagName)){event.preventDefault();stepPreview(event.key==='ArrowLeft'?-1:1);}});
$('generate-run').addEventListener('click',runGenerate);
$('generate-stop').addEventListener('click',()=>{gen.pending?.controller.abort();});
$('generate-change').addEventListener('click',chooseModel);
$('generate-key').addEventListener('change',readModels);
$('generate-model').addEventListener('change',()=>{$('generate-apply').disabled=!$('generate-model').value;});
$('generate-apply').addEventListener('click',applyModel);
$('generate-cancel').addEventListener('click',()=>{gen.version++;gen.leaving=false;decorate($('generate-manage'),'key','管理密钥');gen.mode=gen.connection?'ready':'idle';renderGenerate();setGenerateStatus('');$('generate-run').focus({preventScroll:true});});
$('generate-manage').addEventListener('click',()=>{
 if(!gen.leaving){gen.leaving=true;decorate($('generate-manage'),'key','确认离开，管理密钥');setGenerateStatus('前往密钥管理会离开工坊，本页材料和生成结果会清除。');return;}
 host?.request('manage-keys').catch(error=>setGenerateStatus(error.message));
});
$('generate-source').addEventListener('change',()=>{gen.source=$('generate-source').value;});
$('generate-size').addEventListener('change',()=>{gen.size=$('generate-size').value;});
$('close-result').addEventListener('click',()=>$('result-dialog').close());
$('result-image').addEventListener('error',()=>{if($('result-dialog').open)$('result-error').hidden=false;});
$('result-dialog').addEventListener('close',()=>{$('result-image').removeAttribute('src');$('result-download').removeAttribute('href');if(resultReturn?.isConnected)resultReturn.focus({preventScroll:true});resultReturn=undefined;});
$('result-dialog').addEventListener('click',event=>{if(event.target===$('result-dialog')){const r=event.target.getBoundingClientRect();if(event.clientX<r.left||event.clientX>r.right||event.clientY<r.top||event.clientY>r.bottom)event.target.close();}});
document.addEventListener('mofa-session-ended',()=>{if($('result-dialog').open)$('result-dialog').close();resetGeneration('登录已失效，本页的模型连接与生成结果已清除。');});
window.addEventListener('resize',updateChips);
$('theme-toggle').addEventListener('click',()=>{const next=document.documentElement.dataset.theme==='dark'?'light':'dark';persist({...store.data,theme:next});host?.setTheme(next);applyTheme();});document.querySelectorAll('[data-section]').forEach(b=>b.addEventListener('click',()=>setSection(b.dataset.section)));$('history-button').addEventListener('click',()=>setSection('history'));
$('search').addEventListener('input',()=>{state.query=$('search').value;renderCollection();});$('sort').addEventListener('change',()=>{state.sort=$('sort').value;renderCollection();});$('resource-type').addEventListener('change',()=>{state.type=$('resource-type').value;state.category='all';if(state.section==='discover'&&['skill','workflow'].includes(state.type))state.section='skills';else if(state.section==='skills'&&state.type==='prompt')state.section='discover';renderCollection();});$('more-filters').addEventListener('click',()=>{const open=$('advanced-filters').hidden;$('advanced-filters').hidden=!open;$('more-filters').setAttribute('aria-expanded',String(open));});
$('return-to-library').addEventListener('click',()=>{$('collection').scrollIntoView({behavior:window.matchMedia('(prefers-reduced-motion: reduce)').matches?'instant':'smooth',block:'start'});(state.section==='plays'?$('history-button'):$('search')).focus({preventScroll:true});});
$('bench-save').addEventListener('click',()=>toggleFavorite(selected));$('material-form').addEventListener('submit',event=>event.preventDefault());$('reset-materials').addEventListener('click',resetMaterials);
for(const lang of['zh','en'])$('language-'+lang).addEventListener('click',()=>{language=lang;$('language-zh').setAttribute('aria-pressed',String(lang==='zh'));$('language-en').setAttribute('aria-pressed',String(lang==='en'));updatePrompt();});
// 复制期间只防重复点击，不禁用按钮：禁用已聚焦按钮会让焦点掉到页面开头。
$('copy-prompt').addEventListener('click',async()=>{const b=$('copy-prompt');if(b.dataset.busy)return;b.dataset.busy='true';b.setAttribute('aria-busy','true');try{await copyText($('prompt-output').value,'提示词已复制，可以粘贴到目标工具。',b);}finally{delete b.dataset.busy;b.removeAttribute('aria-busy');}});
$('download-prompt').addEventListener('click',()=>{const blob=new Blob([exportText(selected,$('prompt-output').value)],{type:'text/plain;charset=utf-8'}),url=URL.createObjectURL(blob),link=el('a');link.href=url;link.download=`魔法工坊-${selected.id}.txt`;document.body.append(link);link.click();link.remove();setTimeout(()=>URL.revokeObjectURL(url),1000);toast('已请求下载TXT，请在浏览器下载记录中查看。');});
$('share-item').addEventListener('click',()=>{const url=new URL(location.href);url.search='';url.hash=new URLSearchParams({item:selected.id}).toString();copyText(url.toString(),'条目链接已复制，不含你的替换材料。');});
$('about-button').addEventListener('click',()=>{$('storage-status').textContent=store.available?'本地收藏和浏览记录可用；清除浏览器数据会删除这些记录。':'浏览器存储不可用，收藏与浏览记录仅本页有效。';$('about-dialog').showModal();});$('clear-history').addEventListener('click',()=>{persist({...store.data,history:[]});renderCollection();toast('已清除最近浏览，收藏保留。');});
for(const name of['preview','about']){const dialog=$(name+'-dialog');$('close-'+name).addEventListener('click',()=>dialog.close());dialog.addEventListener('click',event=>{if(event.target===dialog){const r=dialog.getBoundingClientRect();if(event.clientX<r.left||event.clientX>r.right||event.clientY<r.top||event.clientY>r.bottom)dialog.close();}});}
function fromLocation(){const linked=itemFromHash(location.hash,items);if(linked&&!isItemAvailable(linked)){toast('视频与分镜暂未开放，可以先从图片找灵感。');return;}selected=linked||items[0];renderCollection();renderBench();}window.addEventListener('hashchange',fromLocation);window.addEventListener('popstate',fromLocation);window.addEventListener('storage',event=>{if(event.key===STORE_KEY){store.refresh(event.newValue);applyTheme();renderCollection();saveButton($('bench-save'),selected);}});
document.addEventListener('keydown',event=>{if(event.key==='/'&&!/INPUT|TEXTAREA|SELECT/.test(event.target.tagName)&&!event.target.isContentEditable&&!document.querySelector('dialog[open]')){event.preventDefault();if(state.section==='plays')setSection('discover');$('search').focus();}});
document.addEventListener('visibilitychange',()=>{if(document.hidden)pauseVideos();});
$('preview-dialog').addEventListener('close',()=>{stopClip();pauseVideos();$('preview-media').replaceChildren();$('preview-details').replaceChildren();if(typeof previewReturn==='string')document.querySelector('[data-item-id="'+previewReturn+'"] .card-open')?.focus({preventScroll:true});else previewReturn?.focus({preventScroll:true});});
document.addEventListener('mofa-theme',applyTheme);
if(host?.embedded){const home=document.querySelector('.library-footer a');home.removeAttribute('target');home.removeAttribute('aria-label');home.addEventListener('click',event=>{event.preventDefault();host.request('family').catch(()=>{});});}

if(selected.type!=='prompt')state.section='skills';applyTheme();renderCollection();renderBench();
if(requestedItem&&!isItemAvailable(requestedItem))toast('视频与分镜暂未开放，先为你打开图片灵感。');
if(host?.embedded)$('submit-work').addEventListener('click',event=>{event.preventDefault();host.request('studio-submit').catch(()=>toast('投稿入口暂不可用，请通过控制台打开。'));});
let communityLoading=false;
async function refreshCommunity(){
 if(communityLoading)return;communityLoading=true;$('refresh-community').disabled=true;$('community-status').textContent='正在读取审核通过的投稿…';
 try{
  const response=await window.fetch('/api/v1/studio/gallery',{credentials:'omit',cache:'no-store'});
  if(!response.ok)throw Error('gallery');const body=await response.json();
  if(body.code!==0||!Array.isArray(body.data)||body.data.length>1000)throw Error('gallery');
  const published=[...new Map(body.data.map(communityItem).filter(Boolean).map(item=>[item.id,item])).values()];
  items=[...catalogItems,...published];ids.splice(0,ids.length,...items.map(item=>item.id));
  const available=new Set(ids);store.write({...store.data,favorites:store.data.favorites.filter(id=>available.has(id)),history:store.data.history.filter(id=>available.has(id))});
  const stillSelected=items.find(item=>item.id===selected.id);
  if(selected.community&&!stillSelected){pauseVideos();$('preview-dialog').close();selected=items[0];renderBench();}
  const linked=itemFromHash(location.hash,items);if(linked&&isItemAvailable(linked)&&linked.id!==selected.id){selected=linked;renderBench();}
  const availablePublished=published.filter(isItemAvailable);
  renderCollection();$('community-status').textContent=availablePublished.length?`${availablePublished.length} 份审核通过的投稿 · 提示词由作者提供`:'尚无可展示的图片投稿，欢迎分享原创作品。';
 }catch{$('community-status').textContent='投稿目录暂不可用；内置图库仍可使用。';}
 finally{communityLoading=false;$('refresh-community').disabled=false;}
}
if(integrated){$('refresh-community').hidden=false;$('refresh-community').addEventListener('click',refreshCommunity);void refreshCommunity();}
