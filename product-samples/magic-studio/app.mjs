import { items as builtinItems, plays, typeLabels, artSizes } from './catalog.mjs';
import { icon } from './icons.mjs';
import { STORE_KEY, makeStore, findItems, compile, exportText, itemFromHash, communityItem, masonryPositions, VIDEO_CONTENT_ENABLED, isItemAvailable } from './core.mjs';

const $ = id => document.getElementById(id);
const el = (tag, cls, text) => { const n=document.createElement(tag); if(cls)n.className=cls; if(text!==undefined)n.textContent=text; return n; };
const decorate = (n,name,text) => { n.replaceChildren(icon(name));if(text)n.append(document.createTextNode(text));return n; };
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
let items=[...builtinItems];
const ids=items.map(x=>x.id),store=makeStore(storage,ids);
const host=window.MofaFamilyHost;
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
function makeImage(art,title,thumb=true){const image=el('img');image.src=art.startsWith('/api/v1/studio/gallery/')?art:`assets/${art}${thumb?'-thumb':''}.webp`;image.alt=title;image.decoding='async';if(artSizes[art]){image.width=artSizes[art].width;image.height=artSizes[art].height;}image.addEventListener('error',()=>{const p=image.parentElement;image.hidden=true;if(p&&!p.querySelector('.image-fallback'))p.append(el('span','image-fallback','预览暂时不可用，提示词仍可使用。'));});return image;}
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
 if(item.art||item.video){const media=el('div','card-media '+item.ratio);if(item.art){const img=makeImage(item.art,'');img.loading='lazy';media.append(img);}else{const video=el('video');video.src=item.video;video.muted=true;video.preload='metadata';video.playsInline=true;media.append(video);}media.append(el('span','card-type',item.community?'用户投稿':item.media==='video'?'视频作品':'原创图像'));if(item.video){const play=el('span','card-play');play.append(icon('play'));media.append(play);}open.append(media);}
 else if(item.shots)open.append(makeStoryboard(item,true));
 else{const visual=el('div','skill-visual');visual.append(icon(item.type==='skill'?'code':'sliders'),el('span','skill-tool',item.tool));open.append(visual);}
 const caption=el('div','card-caption');caption.append(el('h3','',item.title));if(!item.art&&!item.shots)caption.append(el('p','skill-summary',item.summary));caption.append(el('p','',item.category+' · '+(item.shots?'原创分镜':typeLabels[item.type])));open.append(caption);
 open.addEventListener('click',()=>{select(item);if(item.type==='prompt'){previewReturn=item.id;showLarge();}else $('workbench').focus({preventScroll:true});});
 const favorite=el('button','icon-button card-save');favorite.type='button';saveButton(favorite,item);favorite.addEventListener('click',()=>toggleFavorite(item));article.append(open,favorite);
 if(item.type==='prompt'){
  const quick=el('div','card-quick'),copy=el('button','card-copy','复制模板'),use=el('button','card-use','改成我的');
  copy.type=use.type='button';copy.setAttribute('aria-label','复制模板：'+item.title);use.setAttribute('aria-label','改成我的：'+item.title);
  copy.addEventListener('click',()=>copyText(compile(item,drafts.get(item.id)||{},language),'已复制复用模板；不是原始生成记录。'));
  use.addEventListener('click',()=>select(item,{focus:true}));quick.append(copy,use);article.append(quick);
 }
 return article;
}
function baseItems(){return findItems(items,{...state,query:'',category:'all',style:'all',media:'all',type:'all'},store.data);}
function renderFilters(){
 const base=baseItems();$('media-filters').replaceChildren();
 for(const [value,label,glyph]of[['all','全部',null],['image','图片','image'],['video','视频与分镜','video']]){const b=el('button');const closed=value==='video'&&!VIDEO_CONTENT_ENABLED;if(glyph)b.append(icon(glyph));b.append(document.createTextNode(label),el('span','filter-count',closed?'待开放':String(base.filter(x=>value==='all'||x.media===value).length)));b.setAttribute('aria-pressed',String(state.media===value));b.dataset.media=value;b.disabled=closed;if(closed){b.title='视频与分镜暂未开放';b.setAttribute('aria-describedby','video-availability');}b.addEventListener('click',()=>{if(closed)return;state.media=value;state.category='all';state.style='all';renderCollection();document.querySelector(`[data-media="${value}"]`)?.focus({preventScroll:true});});$('media-filters').append(b);}
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
function renderCollection(){
 const [title,description]=headings[state.section];$('page-title').textContent=title;$('page-description').textContent=description;document.querySelectorAll('[data-section]').forEach(b=>{if(b.dataset.section===state.section)b.setAttribute('aria-current','page');else b.removeAttribute('aria-current');});$('saved-count').textContent=String(store.data.favorites.filter(id=>items.some(item=>item.id===id&&isItemAvailable(item))).length);$('video-availability').hidden=VIDEO_CONTENT_ENABLED;
 const isPlays=state.section==='plays';$('catalog-controls').hidden=isPlays;$('gallery').hidden=isPlays;$('plays-list').hidden=!isPlays;$('empty-state').hidden=true;$('collection-note').textContent=isPlays?'组合方法 · 等待实测':state.section==='skills'?'官方出处 · 未安装运行':VIDEO_CONTENT_ENABLED?'原创图像与对应提示词 · 视频分镜待样片':'原创图像与对应提示词';
 if(isPlays){observeGallery();$('result-count').textContent=`${plays.length} 个创作玩法`;renderPlays();return;}
 renderFilters();const visible=findItems(items,state,store.data);$('result-count').textContent=`${visible.length} 份${state.section==='skills'?'技能与工作流':'创作灵感'}`;$('gallery').replaceChildren(...visible.map(card));$('gallery').setAttribute('role','list');
 observeGallery();
 if(!visible.length){$('empty-state').hidden=false;const emptyCollection=['saved','history'].includes(state.section)&&!state.query&&state.media==='all'&&state.category==='all'&&state.type==='all';$('empty-title').textContent=emptyCollection?(state.section==='saved'?'把喜欢的灵感留下来':'从一个灵感开始'):'暂时没有找到';$('empty-description').textContent=emptyCollection?(state.section==='saved'?'点击条目上的收藏按钮，它会出现在这里。':'打开任意条目，就能在这里继续找回。'):'试试更短的关键词，或者清除筛选看看全部内容。';$('empty-action').textContent=emptyCollection?'去发现灵感':'清除筛选';$('empty-action').onclick=()=>{if(emptyCollection)setSection('discover');else resetFilters();};}
}
function resetFilters(){state.media='all';state.category='all';state.style='all';state.type='all';state.query='';$('search').value='';$('resource-type').value='all';renderCollection();}
function setSection(section){state.section=section;resetFilters();document.querySelector(`[data-section="${section}"]`)?.focus({preventScroll:true});}
function values(){if(!drafts.has(selected.id))drafts.set(selected.id,Object.fromEntries(selected.fields.map(f=>[f.key,f.value])));return drafts.get(selected.id);}
function updatePrompt(){$('prompt-output').value=compile(selected,values(),language);if($('preview-prompt-output'))$('preview-prompt-output').value=$('prompt-output').value;}
function select(item,options={}){if(!item)return;if(!isItemAvailable(item)){toast('视频与分镜暂未开放，可以先从图片找灵感。');return;}pauseVideos();selected=item;persist({...store.data,history:[item.id,...store.data.history.filter(x=>x!==item.id)]});const hash=new URLSearchParams({item:item.id}).toString();if(location.hash!==`#${hash}`)history.pushState(null,'',`#${hash}`);renderCollection();renderBench();if(options.focus){$('workbench').focus({preventScroll:true});if(window.innerWidth<=900)$('workbench').scrollIntoView({behavior:window.matchMedia('(prefers-reduced-motion: reduce)').matches?'instant':'smooth',block:'start'});}}
function pauseVideos(except){document.querySelectorAll('video').forEach(video=>{if(video!==except)video.pause();});}
function makeVideo(item){const video=el('video','film-player');video.src=item.video;if(item.art)video.poster=item.community?item.art:'assets/'+item.art+'.webp';video.controls=true;video.playsInline=true;video.preload='metadata';video.setAttribute('aria-label',item.title+'，'+(item.community?'用户投稿成片':'授权影片节选'));video.addEventListener('play',()=>pauseVideos(video));video.addEventListener('error',()=>toast('视频暂不可用，可能已下架；请刷新投稿目录。'));return video;}
function attribution(item){const line=el('p','film-credit');line.append(document.createTextNode(item.credit+' · '+item.license+(item.community?'':' · 节选已裁短、缩放、静音。')));return line;}
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
function showLarge(){
 pauseVideos();const dialog=$('preview-dialog'),media=$('preview-media'),details=$('preview-details');media.replaceChildren();details.replaceChildren();
 if(selected.video){const video=makeVideo(selected);media.append(video,attribution(selected));}
 else if(selected.art){const image=makeImage(selected.art,selected.title,false);image.id='preview-image';media.append(image);}
 else if(selected.shots)media.append(makeStoryboard(selected));
 $('preview-caption').textContent=selected.title+' · '+(selected.shots?'原创分镜，样片待补':'作品与提示词');
 details.append(el('h3','',selected.title),el('p','fine-print',selected.shots||selected.community?selected.requirements:'原创AI图像；可以看原始记录，也可以替换内容复用。'));
 if(selected.media==='image'||selected.community)details.append(originalRecord(selected,true));
 const promptLabel=el('label','preview-prompt-label',selected.shots?'完整视频分镜提示词':'可替换复用模板');promptLabel.htmlFor='preview-prompt-output';
 const prompt=el('textarea');prompt.id='preview-prompt-output';prompt.value=compile(selected,values(),language);prompt.readOnly=true;prompt.rows=8;
 const copy=el('button','primary-button',selected.shots?'复制分镜提示词':'复制我的模板');copy.type='button';copy.id='preview-copy';copy.addEventListener('click',()=>copyText(prompt.value,'提示词已复制；当前页面未运行生成模型。'));
 const use=el('button','secondary-button','替换内容，做成我的');use.type='button';use.addEventListener('click',()=>{dialog.close();$('workbench').focus({preventScroll:true});if(window.innerWidth<=900)$('workbench').scrollIntoView({behavior:'instant',block:'start'});});
 details.append(promptLabel,prompt,el('p','fine-print','模板使用当前右侧材料和语言；修改材料不改动原始生成记录。'),copy,use);
 dialog.showModal();dialog.scrollTop=0;$('close-preview').focus({preventScroll:true});if(selected.video)media.querySelector('video').play()?.catch(()=>toast('点击视频播放按钮开始观看。'));
}
function renderBench(){
 pauseVideos();$('bench-content').scrollTop=0;$('selected-title').textContent=selected.title;$('selected-description').textContent=selected.summary;$('bench-badges').replaceChildren(el('span','',typeLabels[selected.type]),el('span','',selected.status));saveButton($('bench-save'),selected);$('bench-preview').replaceChildren();
 if(selected.shots&&!selected.video){const button=el('button','preview-button script-preview');button.type='button';button.setAttribute('aria-label','查看分镜：'+selected.title);button.append(makeStoryboard(selected));button.addEventListener('click',()=>{previewReturn=button;showLarge();});$('bench-preview').append(button);}
 else if(selected.art||selected.video){
  if(selected.media==='video'){$('bench-preview').append(makeVideo(selected),attribution(selected));}
  else{const button=el('button','preview-button');button.setAttribute('aria-label','放大预览：'+selected.title);button.append(makeImage(selected.art,selected.title,false));button.addEventListener('click',()=>{previewReturn=button;showLarge();});$('bench-preview').append(button,el('p','fine-print',selected.community?'用户授权投稿 · 点击查看完整作品':'原创AI视觉示例 · 点击查看完整大图'));}
 }else{const visual=el('div','bench-skill'),text=el('div');text.append(el('strong','',selected.tool),el('span','',selected.type==='skill'?'官方 Agent Skill':'官方模板使用指引'));visual.append(icon(selected.type==='skill'?'code':'sliders'),text);$('bench-preview').append(visual);}
 $('bench-origin').replaceChildren();if((selected.media==='image'||selected.community)&&selected.type==='prompt')$('bench-origin').append(originalRecord(selected));
 for(const id of ['language-zh','language-en']){$(id).disabled=!!selected.community;$(id).title=selected.community?'投稿提示词保持作者原文，不自动翻译':'';}
 $('prompt-note').textContent=selected.community?'保持投稿人原文，不自动翻译或替换占位符。可复制后修改，模型效果不由人工审核保证。':'复用模板不等于原始生成记录。空白使用示例值；你的替换内容只留在当前页面。';
 document.querySelector('.source-note').textContent=selected.community?'作品、模型和提示词由投稿人提供；人工审核不等于模型效果实测。':'方法参考官方指南；原创模板尚未逐条实测。';
 const current=values();$('material-fields').replaceChildren();for(const f of selected.fields){const label=el('label','material-field',f.label),input=el('input');input.id=`material-${f.key}`;input.name=f.key;input.value=current[f.key];input.maxLength=300;input.autocomplete='off';input.addEventListener('input',()=>{current[f.key]=input.value;updatePrompt();});label.append(input);$('material-fields').append(label);}updatePrompt();decorate($('copy-prompt'),'copy',selected.type==='prompt'?'复制完整提示词':'复制使用提示词');$('usage-steps').replaceChildren(...selected.steps.map(s=>el('li','',s)));$('usage-requirements').textContent=selected.requirements;$('usage-tip').textContent=selected.tip;$('skill-command').replaceChildren();
 if(selected.command){const code=el('code','command',selected.command),b=el('button','text-button command-button','复制官方安装命令');b.type='button';b.addEventListener('click',()=>copyText(selected.command,'命令已复制，请在自己的环境检查后运行。'));$('skill-command').append(code,b);}
 $('source-link').href=selected.source.url;decorate($('source-link'),'external',selected.source.label);$('workbench').querySelector('.instructions').open=false;document.querySelector('.inline-confirm')?.remove();document.querySelector('.manual-copy')?.remove();
}
async function copyText(text,success){try{if(!navigator.clipboard?.writeText)throw Error('clipboard');await navigator.clipboard.writeText(text);document.querySelector('.manual-copy')?.remove();toast(success);return true;}catch{document.querySelector('.manual-copy')?.remove();const box=el('div','manual-copy'),label=el('label','','手动复制文本'),area=el('textarea');area.value=text;area.readOnly=true;area.rows=4;label.append(area);box.append(label);($('preview-dialog').open?$('preview-details'):$('bench-content')).append(box);area.focus();area.select();toast('自动复制不可用。已选中文本，请按 Ctrl+C 或长按复制。');return false;}}
function resetMaterials(){if(selected.fields.every(f=>values()[f.key]===f.value)||document.querySelector('.inline-confirm'))return;const box=el('div','inline-confirm'),cancel=el('button','secondary-button','保留我的内容'),confirm=el('button','secondary-button','恢复示例');cancel.type=confirm.type='button';cancel.addEventListener('click',()=>{box.remove();$('reset-materials').focus();});confirm.addEventListener('click',()=>{drafts.delete(selected.id);renderBench();$('reset-materials').focus();toast('已恢复当前条目的示例材料。');});box.append(el('p','','恢复示例会覆盖当前条目的材料。'),cancel,confirm);$('material-form').append(box);cancel.focus();}
 for(const[id,glyph,label]of[['search-icon','search'],['empty-icon','search'],['more-filters','sliders','更多筛选'],['history-button','history','最近浏览'],['download-prompt','download'],['share-item','external'],['close-preview','close'],['close-about','close'],['return-to-library','arrow']])decorate($(id),glyph,label);
$('theme-toggle').addEventListener('click',()=>{const next=document.documentElement.dataset.theme==='dark'?'light':'dark';persist({...store.data,theme:next});host?.setTheme(next);applyTheme();});document.querySelectorAll('[data-section]').forEach(b=>b.addEventListener('click',()=>setSection(b.dataset.section)));$('history-button').addEventListener('click',()=>setSection('history'));
$('search').addEventListener('input',()=>{state.query=$('search').value;renderCollection();});$('sort').addEventListener('change',()=>{state.sort=$('sort').value;renderCollection();});$('resource-type').addEventListener('change',()=>{state.type=$('resource-type').value;state.category='all';if(state.section==='discover'&&['skill','workflow'].includes(state.type))state.section='skills';else if(state.section==='skills'&&state.type==='prompt')state.section='discover';renderCollection();});$('more-filters').addEventListener('click',()=>{const open=$('advanced-filters').hidden;$('advanced-filters').hidden=!open;$('more-filters').setAttribute('aria-expanded',String(open));});
$('return-to-library').addEventListener('click',()=>{$('collection').scrollIntoView({behavior:window.matchMedia('(prefers-reduced-motion: reduce)').matches?'instant':'smooth',block:'start'});(state.section==='plays'?$('history-button'):$('search')).focus({preventScroll:true});});
$('bench-save').addEventListener('click',()=>toggleFavorite(selected));$('material-form').addEventListener('submit',event=>event.preventDefault());$('reset-materials').addEventListener('click',resetMaterials);
for(const lang of['zh','en'])$('language-'+lang).addEventListener('click',()=>{language=lang;$('language-zh').setAttribute('aria-pressed',String(lang==='zh'));$('language-en').setAttribute('aria-pressed',String(lang==='en'));updatePrompt();});
$('copy-prompt').addEventListener('click',async()=>{const b=$('copy-prompt');b.disabled=true;await copyText($('prompt-output').value,'提示词已复制，可以粘贴到目标工具。');b.disabled=false;});
$('download-prompt').addEventListener('click',()=>{const blob=new Blob([exportText(selected,$('prompt-output').value)],{type:'text/plain;charset=utf-8'}),url=URL.createObjectURL(blob),link=el('a');link.href=url;link.download=`魔法工坊-${selected.id}.txt`;document.body.append(link);link.click();link.remove();setTimeout(()=>URL.revokeObjectURL(url),1000);toast('已请求下载TXT，请在浏览器下载记录中查看。');});
$('share-item').addEventListener('click',()=>{const url=new URL(location.href);url.search='';url.hash=new URLSearchParams({item:selected.id}).toString();copyText(url.toString(),'条目链接已复制，不含你的替换材料。');});
$('about-button').addEventListener('click',()=>{$('storage-status').textContent=store.available?'本地收藏和浏览记录可用；清除浏览器数据会删除这些记录。':'浏览器存储不可用，收藏与浏览记录仅本页有效。';$('about-dialog').showModal();});$('clear-history').addEventListener('click',()=>{persist({...store.data,history:[]});renderCollection();toast('已清除最近浏览，收藏保留。');});
for(const name of['preview','about']){const dialog=$(name+'-dialog');$('close-'+name).addEventListener('click',()=>dialog.close());dialog.addEventListener('click',event=>{if(event.target===dialog){const r=dialog.getBoundingClientRect();if(event.clientX<r.left||event.clientX>r.right||event.clientY<r.top||event.clientY>r.bottom)dialog.close();}});}
function fromLocation(){const linked=itemFromHash(location.hash,items);if(linked&&!isItemAvailable(linked)){toast('视频与分镜暂未开放，可以先从图片找灵感。');return;}selected=linked||items[0];renderCollection();renderBench();}window.addEventListener('hashchange',fromLocation);window.addEventListener('popstate',fromLocation);window.addEventListener('storage',event=>{if(event.key===STORE_KEY){store.refresh(event.newValue);applyTheme();renderCollection();saveButton($('bench-save'),selected);}});
document.addEventListener('keydown',event=>{if(event.key==='/'&&!/INPUT|TEXTAREA|SELECT/.test(event.target.tagName)&&!event.target.isContentEditable&&!document.querySelector('dialog[open]')){event.preventDefault();if(state.section==='plays')setSection('discover');$('search').focus();}});
document.addEventListener('visibilitychange',()=>{if(document.hidden)pauseVideos();});
$('preview-dialog').addEventListener('close',()=>{pauseVideos();$('preview-media').replaceChildren();$('preview-details').replaceChildren();if(typeof previewReturn==='string')document.querySelector('[data-item-id="'+previewReturn+'"] .card-open')?.focus({preventScroll:true});else previewReturn?.focus({preventScroll:true});});
document.addEventListener('mofa-theme',applyTheme);
if(host?.embedded){const home=document.querySelector('.library-footer a');home.removeAttribute('target');home.removeAttribute('aria-label');home.addEventListener('click',event=>{event.preventDefault();host.request('family').catch(()=>{});});}

if(selected.type!=='prompt')state.section='skills';applyTheme();renderCollection();renderBench();
if(requestedItem&&!isItemAvailable(requestedItem))toast('视频与分镜暂未开放，先为你打开图片灵感。');
const integrated=document.querySelector('meta[name="mofa-api-site"]')?.content==='same-origin';
if(host?.embedded)$('submit-work').addEventListener('click',event=>{event.preventDefault();host.request('studio-submit').catch(()=>toast('投稿入口暂不可用，请通过控制台打开。'));});
let communityLoading=false;
async function refreshCommunity(){
 if(communityLoading)return;communityLoading=true;$('refresh-community').disabled=true;$('community-status').textContent='正在读取审核通过的投稿…';
 try{
  const response=await window.fetch('/api/v1/studio/gallery',{credentials:'omit',cache:'no-store'});
  if(!response.ok)throw Error('gallery');const body=await response.json();
  if(body.code!==0||!Array.isArray(body.data)||body.data.length>1000)throw Error('gallery');
  const published=[...new Map(body.data.map(communityItem).filter(Boolean).map(item=>[item.id,item])).values()];
  items=[...builtinItems,...published];ids.splice(0,ids.length,...items.map(item=>item.id));
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
