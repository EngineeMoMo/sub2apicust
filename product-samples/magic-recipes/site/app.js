(() => {
  'use strict';
  const recipes = JSON.parse(document.getElementById('recipe-data').textContent);
  const core = globalThis.MofaRecipeCore;
  const byId = id => document.getElementById(id);
  const states = new Map();
  let active = recipes.find(recipe => recipe.id === location.hash.slice(1)) || recipes[0];
  let category = 'all';
  let noticeTimer;
  function destinationFor(recipe) {
    if (recipe.target === 'image') return '支持生图的 AI 工具';
    if (recipe.target === 'video') return '支持视频生成的 AI 工具';
    return '自己的 AI 对话工具';
  }
  const confirmation = byId('confirm-dialog');
  let pendingConfirmation;
  function confirmAction({ title, message, label, destructive, trigger }) {
    if (pendingConfirmation) return Promise.resolve(false);
    if (typeof confirmation.showModal !== 'function') {
      announce('此浏览器无法打开确认框，材料未改动。请换用支持对话框的浏览器。');
      return Promise.resolve(false);
    }
    byId('confirm-title').textContent = title;
    byId('confirm-message').textContent = message;
    byId('confirm-accept').textContent = label;
    confirmation.dataset.destructive = String(Boolean(destructive));
    confirmation.returnValue = '';
    return new Promise(resolve => {
      pendingConfirmation = { resolve, trigger };
      confirmation.showModal();
      byId('confirm-cancel').focus();
    });
  }
  confirmation.addEventListener('close', () => {
    const pending = pendingConfirmation;
    pendingConfirmation = undefined;
    if (!pending) return;
    if (pending.trigger.isConnected) pending.trigger.focus({ preventScroll: true });
    pending.resolve(confirmation.returnValue === 'accept');
  });
  confirmation.addEventListener('cancel', event => {
    event.preventDefault();
    confirmation.close('cancel');
  });
  confirmation.addEventListener('keydown', event => {
    if (event.key !== 'Tab') return;
    const buttons = [...confirmation.querySelectorAll('button:not([disabled])')];
    const first = buttons[0];
    const last = buttons[buttons.length - 1];
    if (event.shiftKey && document.activeElement === first) {
      event.preventDefault();
      last.focus();
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault();
      first.focus();
    }
  });
  byId('confirm-cancel').addEventListener('click', () => confirmation.close('cancel'));
  byId('confirm-close').addEventListener('click', () => confirmation.close('cancel'));
  byId('confirm-accept').addEventListener('click', () => confirmation.close('accept'));
  function state() {
    if (!states.has(active.id)) states.set(active.id, { values: {}, example: false });
    return states.get(active.id);
  }
  function announce(message) {
    clearTimeout(noticeTimer);
    byId('status').textContent = message;
    noticeTimer = setTimeout(() => { byId('status').textContent = ''; }, 6500);
  }
  function showView(view) {
    const editing = view === 'edit';
    byId('editor').hidden = !editing;
    byId('guide').hidden = editing;
    byId('edit-view').setAttribute('aria-pressed', String(editing));
    byId('guide-view').setAttribute('aria-pressed', String(!editing));
  }
  function renderList() {
    const found = core.filter(recipes, category, byId('search').value);
    byId('count').textContent = found.length + ' 份';
    byId('list-status').textContent = '找到' + found.length + '份配方';
    byId('recipe-list').replaceChildren();
    for (const recipe of found) {
      const button = document.createElement('button');
      button.type = 'button';
      button.className = 'recipe-link';
      button.dataset.recipe = recipe.id;
      button.setAttribute('aria-current', String(active.id === recipe.id));
      const title = document.createElement('strong');
      title.textContent = recipe.title;
      const description = document.createElement('span');
      description.textContent = recipe.outcome;
      button.append(title, description);
      button.addEventListener('click', () => {
        activate(recipe);
        if (location.hash !== '#' + recipe.id) location.hash = recipe.id;
        if (matchMedia('(max-width: 680px)').matches) byId('workspace').scrollIntoView({ block: 'start' });
        byId('workspace').focus({ preventScroll: true });
      });
      byId('recipe-list').append(button);
    }
    byId('empty').hidden = found.length > 0;
  }
  function invalidate() {
    byId('result').hidden = true;
    byId('output').value = '';
    document.dispatchEvent(new CustomEvent('recipe-reset', { detail: { recipe: active } }));
  }
  function renderFields() {
    byId('fields').replaceChildren();
    byId('form-error').hidden = true;
    byId('example-notice').hidden = !state().example;
    for (const field of active.fields) {
      const wrapper = document.createElement('div');
      wrapper.className = 'field';
      const label = document.createElement('label');
      label.htmlFor = 'input-' + field.id;
      label.append(document.createTextNode(field.label));
      const requirement = document.createElement('span');
      requirement.textContent = field.required ? '必填' : '可选';
      label.append(requirement);
      const control = document.createElement(field.type === 'text' ? 'input' : 'textarea');
      if (field.type === 'text') control.type = 'text';
      else control.rows = field.id === 'material' ? 8 : 3;
      control.id = 'input-' + field.id;
      control.name = field.id;
      control.required = Boolean(field.required);
      control.maxLength = 12000;
      control.value = state().values[field.id] || '';
      control.spellcheck = false;
      const hint = document.createElement('p');
      hint.id = 'hint-' + field.id;
      hint.className = 'hint';
      hint.textContent = field.hint || (field.required ? '填写你已有的材料，不用补全未知事实。' : '没有信息可以留空。');
      const error = document.createElement('p');
      error.id = 'error-' + field.id;
      error.className = 'error';
      error.hidden = true;
      control.setAttribute('aria-describedby', hint.id + ' ' + error.id);
      control.addEventListener('input', () => {
        state().values[field.id] = control.value;
        control.removeAttribute('aria-invalid');
        error.hidden = true;
        byId('form-error').hidden = true;
        invalidate();
      });
      wrapper.append(label, control, hint, error);
      byId('fields').append(wrapper);
    }
    invalidate();
  }
  function activate(recipe) {
    if (confirmation.open) confirmation.close('cancel');
    active = recipe;
    clearTimeout(noticeTimer);
    byId('status').textContent = '';
    byId('category-label').textContent = { work: '工作配方', university: '大学配方', creative: '创作配方' }[recipe.category];
    byId('version').textContent = 'v' + recipe.version + ' · 待实测草稿';
    byId('recipe-title').textContent = recipe.title;
    byId('recipe-description').textContent = recipe.description;
    byId('outcome').textContent = recipe.outcome;
    byId('result-description').textContent = '已按配方模板组装你的材料，没有调用模型。请复制到' + destinationFor(recipe) + '中使用。';
    byId('result-next').textContent = recipe.target === 'image' || recipe.target === 'video'
      ? '下一步：打开' + destinationFor(recipe) + '，粘贴提示词，并按该工具支持的功能设置画幅、时长或参考素材。本页不会自动设置或上传。'
      : '下一步：到自己的 AI 工具新建对话，粘贴完整提示词并发送。请先检查材料是否已脱敏。';
    byId('scope').textContent = recipe.scope;
    byId('example').textContent = recipe.example.replace(/^\S+\n/, '').replace(/^~~~[^\n]*\n?/gm, '');
    byId('checks').textContent = recipe.checks;
    byId('repair').textContent = recipe.repair.replace(/^~~~text\n|\n~~~$/g, '');
    document.title = recipe.title + ' · 魔法配方';
    showView('edit');
    renderFields();
    renderList();
  }
  function hasValues() { return Object.values(state().values).some(value => value.trim()); }
  byId('load-example').addEventListener('click', async event => {
    const recipeId = active.id;
    if (hasValues() && !await confirmAction({
      title: '用示例替换当前材料？',
      message: '当前配方中已填写的内容会被虚构示例替换，替换后无法撤销。其他配方的材料不会改变。',
      label: '替换为示例', trigger: event.currentTarget
    })) return;
    if (active.id !== recipeId) return;
    state().values = Object.fromEntries(active.fields.map(field => [field.id, field.example]));
    state().example = true;
    renderFields();
    announce('已填入虚构示例。可编辑材料后生成提示词。');
  });
  byId('clear').addEventListener('click', async event => {
    const recipeId = active.id;
    if (hasValues() && !await confirmAction({
      title: '清空这份配方的材料？',
      message: '已填写的材料和生成的提示词都会清除，且无法恢复。其他配方的材料会保留。',
      label: '确认清空', destructive: true, trigger: event.currentTarget
    })) return;
    if (active.id !== recipeId) return;
    state().values = {};
    state().example = false;
    renderFields();
    announce('已清空当前配方。');
  });
  byId('recipe-form').addEventListener('submit', event => {
    event.preventDefault();
    for (const field of active.fields) state().values[field.id] = byId('input-' + field.id).value;
    const result = core.compile(active, state().values);
    for (const field of active.fields) { byId('input-' + field.id).removeAttribute('aria-invalid'); byId('error-' + field.id).hidden = true; }
    if (result.errors.length) {
      invalidate();
      byId('form-error').textContent = '还有' + result.errors.length + '项材料需要补充或调整。';
      byId('form-error').hidden = false;
      for (const error of result.errors) {
        byId('input-' + error.id).setAttribute('aria-invalid', 'true');
        byId('error-' + error.id).textContent = error.message;
        byId('error-' + error.id).hidden = false;
      }
      byId('input-' + result.errors[0].id).focus();
      return;
    }
    byId('form-error').hidden = true;
    byId('output').value = result.text;
    byId('result').hidden = false;
    document.dispatchEvent(new CustomEvent('recipe-compiled', { detail: { recipe: active } }));
    byId('result').scrollIntoView({ block: 'nearest', behavior: matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth' });
    byId('copy').focus({ preventScroll: true });
    announce('提示词已整理完成。这一步未调用模型，请复制到' + destinationFor(active) + '中使用。');
  });
  byId('copy').addEventListener('click', async () => {
    const text = byId('output').value;
    if (!text) return;
    const destination = destinationFor(active);
    byId('copy').disabled = true;
    try {
      if (!navigator.clipboard || !navigator.clipboard.writeText) throw new Error('Clipboard unavailable');
      await navigator.clipboard.writeText(text);
      announce('已复制。请到' + destination + '中粘贴并运行。');
    } catch {
      byId('output').focus();
      byId('output').select();
      announce('浏览器未允许自动复制。已选中提示词，请按 Ctrl+C，或在手机上长按复制。');
    } finally { byId('copy').disabled = false; }
  });
  byId('download').addEventListener('click', () => {
    if (!byId('output').value) return;
    const url = URL.createObjectURL(new Blob([byId('output').value], { type: 'text/plain;charset=utf-8' }));
    const link = document.createElement('a');
    link.href = url;
    link.download = '魔法配方-' + active.title + '.txt';
    document.body.append(link);
    link.click();
    link.remove();
    setTimeout(() => URL.revokeObjectURL(url), 1000);
    announce('已请求下载明文文件，请检查浏览器下载列表。');
  });
  for (const button of document.querySelectorAll('[data-category]')) button.addEventListener('click', () => {
    category = button.dataset.category;
    for (const item of document.querySelectorAll('[data-category]')) item.setAttribute('aria-pressed', String(item === button));
    const first = core.filter(recipes, category, byId('search').value)[0];
    if (first && category !== 'all' && active.category !== category) {
      activate(first);
      location.hash = first.id;
      return;
    }
    renderList();
  });
  byId('search').addEventListener('input', renderList);
  byId('reset-search').addEventListener('click', () => {
    category = 'all';
    byId('search').value = '';
    for (const button of document.querySelectorAll('[data-category]')) button.setAttribute('aria-pressed', String(button.dataset.category === 'all'));
    renderList();
    byId('search').focus();
  });
  byId('edit-view').addEventListener('click', () => showView('edit'));
  const tutorial = byId('usage-tutorial');
  function syncTutorial() {
    for (const button of document.querySelectorAll('[data-open-tutorial]')) button.setAttribute('aria-expanded', String(tutorial.open));
  }
  tutorial.addEventListener('toggle', syncTutorial);
  for (const button of document.querySelectorAll('[data-open-tutorial]')) button.addEventListener('click', () => {
    tutorial.open = true;
    syncTutorial();
    tutorial.scrollIntoView({ block: 'start' });
    byId('tutorial-summary').focus({ preventScroll: true });
  });
  byId('tutorial-start').addEventListener('click', () => {
    tutorial.open = false;
    syncTutorial();
    showView('edit');
    byId('workspace').scrollIntoView({ block: 'start' });
    byId('input-' + active.fields[0].id).focus({ preventScroll: true });
  });
  byId('guide-view').addEventListener('click', () => showView('guide'));
  byId('back-to-edit').addEventListener('click', () => { showView('edit'); byId('input-' + active.fields[0].id).focus(); });
  byId('theme').addEventListener('click', () => {
    const light = document.documentElement.dataset.theme !== 'light';
    document.documentElement.dataset.theme = light ? 'light' : 'dark';
    byId('theme').textContent = light ? '切换深色' : '切换浅色';
  });
  window.addEventListener('hashchange', () => {
    const recipe = recipes.find(item => item.id === location.hash.slice(1));
    if (recipe && recipe.id !== active.id) activate(recipe);
  });
  activate(active);
})();
