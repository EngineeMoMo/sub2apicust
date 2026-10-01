(() => {
  'use strict';
  const byId = id => document.getElementById(id);
  const client = globalThis.MofaRecipeModel;
  const connections = new Map();
  const drafts = new Map();
  let editing = 'text';
  let recipe;
  let compiled = false;
  let pending;
  let history = [];
  let latest = '';
  let login;
  const connector = globalThis.MofaRecipeConnect;
  const host = globalThis.MofaFamilyHost;
  let inlineVersion = 0, inlineBusy = false;
  function inlineSelection() {
    return { kind: editing, keyID: Number(byId('magic-key-select').value), model: byId('magic-model-select').value, protocol: byId('magic-protocol').value };
  }
  function inlineControls() {
    byId('connect-magic-api').disabled = inlineBusy;
    byId('magic-key-select').disabled = inlineBusy || byId('magic-key-select').options.length < 2;
    byId('magic-model-select').disabled = inlineBusy || byId('magic-model-select').options.length < 2;
    byId('apply-magic-config').disabled = inlineBusy || !byId('magic-model-select').value;
  }
  function options(id, list, placeholder) {
    const select = byId(id);
    select.replaceChildren(new Option(placeholder, ''));
    list.forEach(entry => select.append(new Option(entry.label, entry.value)));
  }
  async function inlineRequest(action) {
    const version = ++inlineVersion;
    if (action === 'keys') {
      options('magic-key-select', [], '正在读取可用密钥');
      options('magic-model-select', [], '先选择密钥');
    }
    inlineBusy = true; inlineControls();
    byId('magic-connect-status').textContent = action === 'apply' ? '正在应用所选连接…' : '正在读取配置…';
    try {
      const result = await host.request(action, inlineSelection());
      if (version !== inlineVersion) return;
      if (action === 'keys') {
        options('magic-key-select', result.map(entry => ({ value: String(entry.id), label: entry.name + ' · ' + (entry.group || '未命名分组') })), result.length ? '请选择可用密钥' : '没有可用密钥');
        options('magic-model-select', [], '先选择密钥');
        byId('magic-connect-status').textContent = result.length ? '选择密钥后读取其模型目录。' : editing === 'image'
          ? '没有可用于生图的有效密钥。请在 API 密钥管理确认密钥启用、未过期、额度可用且绑定已开启生图权限的分组。'
          : '没有有效密钥。请在 API 密钥管理确认密钥启用、未过期、额度可用且已绑定启用的分组。';
      } else if (action === 'models') {
        options('magic-model-select', result.models.map(model => ({ value: model, label: model })), '请选择模型');
        byId('magic-protocol').value = result.protocol;
        byId('magic-connect-status').textContent = result.models.length ? '目录来自所选密钥；请确认模型能力与接口格式后应用。' : '此密钥没有返回模型，请检查分组或更换密钥。';
      } else {
        const config = client.configure(result);
        connections.set(config.kind, config); drafts.set(config.kind, config);
        loadDraft(editing); if (kind() === editing) resetRun(); refreshConnection();
        byId('magic-connect-status').textContent = '已应用 ' + config.model + '，密钥只留本页内存，尚未调用生成模型。';
        byId('settings-status').textContent = '魔法 API 连接已应用。检查提示词后再手动运行。';
      }
    } catch (error) {
      if (version === inlineVersion) {
        if (action === 'keys') {
          options('magic-key-select', [], '暂无可选密钥，请查看下方原因');
          options('magic-model-select', [], '先选择可用密钥');
        }
        if (action === 'models') options('magic-model-select', [], '目录读取失败，请重选密钥');
        byId('magic-connect-status').textContent = error.message;
      }
    } finally { if (version === inlineVersion) { inlineBusy = false; inlineControls(); } }
  }
  if (host?.embedded) {
    byId('inline-magic-config').hidden = false;
    byId('magic-connect-hint').textContent = '已在魔法 API 控制台中打开。直接读取当前账户的可用密钥与模型，也可在下方手动填写接口。';
    byId('magic-site').closest('.field').hidden = true;
    byId('connect-magic-api').textContent = '读取／刷新可用密钥';
    byId('magic-family-home').removeAttribute('target');
    byId('magic-family-home').textContent = '家族首页';
    byId('magic-family-home').addEventListener('click', event => { event.preventDefault(); host.request('family').catch(() => {}); });
    byId('magic-key-select').addEventListener('change', () => {
      options('magic-model-select', [], '先选择密钥');
      if (byId('magic-key-select').value) inlineRequest('models'); else inlineControls();
    });
    byId('magic-model-select').addEventListener('change', inlineControls);
    byId('apply-magic-config').addEventListener('click', () => inlineRequest('apply'));
    byId('manage-magic-keys').addEventListener('click', () => {
      byId('magic-connect-status').textContent = '前往密钥管理会清除当前页材料与模型连接。';
      const button = byId('manage-magic-keys');
      if (button.dataset.confirmed) host.request('manage-keys').catch(() => {});
      else { button.dataset.confirmed = 'true'; button.textContent = '确认离开，管理密钥'; }
    });
    document.addEventListener('mofa-host-ready', () => { if (!byId('model-settings').hidden) inlineRequest('keys'); });
    document.addEventListener('mofa-session-ended', () => {
      inlineVersion++; inlineBusy = false; connections.clear(); drafts.clear(); resetRun(); loadDraft(editing); refreshConnection();
      options('magic-key-select', [], '登录已失效'); options('magic-model-select', [], '请重新登录'); inlineControls();
      byId('magic-connect-status').textContent = '登录已失效，本页连接与密钥已清除。';
    });
  }
  const sameSite = document.querySelector('meta[name="mofa-api-site"]')?.content === 'same-origin';
  const portalHint = new URLSearchParams(location.search).get('api_site');
  if (sameSite && !host?.embedded) {
    byId('magic-site').value = location.origin;
    byId('magic-site').readOnly = true;
    byId('magic-site-hint').textContent = '已使用当前站点。已有登录有效时可直接选择配置，仍需你明确授权密钥。';
    byId('magic-family-home').href = new URL('/family', location.origin).href;
    byId('connect-magic-api').textContent = '选择魔法 API 配置';
  } else if (portalHint) {
    try {
      const portal = connector.portalURL(portalHint).origin;
      const localPreview = location.protocol === 'http:' && ['localhost', '127.0.0.1', '[::1]'].includes(location.hostname);
      if (new URL(portal).protocol === 'http:' && !localPreview) throw new Error('本机来源提示仅供本机预览');
      byId('magic-site').value = portal;
      byId('magic-family-home').href = new URL('/family', portal).href;
    }
    catch { byId('magic-connect-status').textContent = '来源链接中的网站地址不受支持，已保留魔法 API 官方地址；你也可以手动配置接口。'; }
  }
  function endLogin(message) {
    if (login) clearInterval(login.timer);
    login = undefined;
    byId('connect-magic-api').disabled = false;
    byId('cancel-magic-connect').hidden = true;
    if (message) byId('magic-connect-status').textContent = message;
  }
  function startLogin() {
    if (host?.embedded) { inlineRequest('keys'); return; }
    endLogin();
    byId('magic-site').removeAttribute('aria-invalid');
    try {
      const connection = connector.begin(byId('magic-site').value.trim(), location.origin, crypto, sameSite);
      connection.kind = editing;
      connection.expires = Date.now() + 300000;
      connection.popup = window.open(connector.loginURL(connection, editing), '_blank', 'popup,width=640,height=800');
      if (!connection.popup) throw new Error('浏览器阻止了登录窗口，请允许本页弹出窗口后重试。');
      login = connection;
      connection.timer = setInterval(() => {
        if (Date.now() >= connection.expires) endLogin('接入等待超过五分钟，请重新登录选择配置。');
        else if (connection.popup.closed) endLogin('登录窗口已关闭或被浏览器安全策略隔离，请重新接入或手动配置。');
      }, 1000);
      byId('connect-magic-api').disabled = true;
      byId('cancel-magic-connect').hidden = false;
      byId('magic-connect-status').textContent = '请在魔法 API 窗口登录，选择' + (editing === 'image' ? '生图' : '文字') + '密钥与模型，并点击授权带回。五分钟内有效；不会自动运行模型。';
    } catch (error) {
      endLogin(error.message);
      byId('magic-site').setAttribute('aria-invalid', 'true');
    }
  }
  function receiveLogin(event) {
    if (!connector.accepts(event, login)) return;
    const connection = login;
    try {
      const config = client.configure(event.data.config);
      connections.set(config.kind, config);
      drafts.set(config.kind, config);
      if (config.kind === editing) loadDraft(editing);
      if (config.kind === kind()) resetRun();
      refreshConnection();
      endLogin('已从魔法 API 导入' + (config.kind === 'image' ? '生图' : '文字') + '连接：' + config.model + '。密钥仅留本页，检查提示词后再手动运行。');
      byId('settings-status').textContent = '连接已应用，尚未调用生成模型。清除连接不会退出魔法 API 账号。';
      connection.popup.postMessage({ type: 'mofa-recipes-applied', nonce: connection.nonce }, connection.portal);
    } catch {
      endLogin('所选配置无法应用，请核对原站 API 地址与模型格式后重新选择，或手动配置。');
    }
  }
  function kind() { return recipe?.target === 'image' ? 'image' : 'text'; }
  function defaults(target) { return { kind: target, protocol: 'chat', base: '', key: '', model: '', size: '' }; }
  function collect() {
    return { kind: editing, protocol: byId('model-protocol').value, base: byId('model-base').value, key: byId('model-key').value, model: byId('model-name').value, size: byId('model-size').value };
  }
  function settingsError(error) {
    for (const id of ['base', 'key', 'name', 'size', 'protocol']) byId('model-' + id).removeAttribute('aria-invalid');
    byId('settings-error').hidden = !error;
    byId('settings-error').textContent = error?.message || '';
    if (error?.field) {
      const field = byId('model-' + (error.field === 'model' ? 'name' : error.field));
      field?.setAttribute('aria-invalid', 'true');
      field?.focus();
    }
  }
  function loadDraft(target) {
    editing = target;
    const draft = drafts.get(target) || connections.get(target) || defaults(target);
    byId('model-kind').value = target;
    byId('model-protocol').value = draft.protocol === 'images' ? 'chat' : draft.protocol;
    for (const field of ['base', 'key', 'size']) byId('model-' + field).value = draft[field];
    byId('model-name').value = draft.model;
    byId('model-key').type = 'password';
    byId('show-model-key').textContent = '显示密钥';
    byId('show-model-key').setAttribute('aria-pressed', 'false');
    byId('protocol-field').hidden = target === 'image';
    byId('image-size-field').hidden = target !== 'image';
    settingsError();
  }
  function openSettings(target = kind()) {
    drafts.set(editing, collect());
    loadDraft(target);
    byId('model-settings').hidden = false;
    byId('open-model-settings').setAttribute('aria-expanded', 'true');
    byId('model-settings').scrollIntoView({ block: 'start' });
    byId('model-kind').focus({ preventScroll: true });
    if (host?.embedded) { byId('magic-protocol-field').hidden = editing === 'image'; inlineRequest('keys'); }
  }
  function closeSettings() {
    drafts.set(editing, collect());
    byId('model-settings').hidden = true;
    byId('open-model-settings').setAttribute('aria-expanded', 'false');
    byId('open-model-settings').focus({ preventScroll: true });
  }
  function refreshConnection() {
    const video = recipe?.target === 'video';
    const config = connections.get(kind());
    byId('model-run').hidden = !compiled;
    byId('run-model').disabled = Boolean(pending) || video;
    byId('run-model').textContent = pending ? '正在等待模型…' : kind() === 'image' ? '用我的模型生图' : '用我的模型运行';
    byId('configure-model').hidden = video;
    byId('connection-summary').textContent = video
      ? '视频接口尚未接入，请复制提示词到自己的视频工具。'
      : config ? config.model + ' · ' + new URL(config.endpoint).host + ' · ' + (config.kind === 'image' ? 'Images API' : config.protocol === 'responses' ? 'Responses' : 'Chat Completions')
        : (kind() === 'image' ? '生图' : '文字') + '模型尚未设置。可以先设置连接，或继续复制提示词使用。';
    byId('run-note').textContent = video ? '视频模型的任务提交与结果查询需要按服务商格式适配。'
      : '点击后会把上方完整提示词发给你的接口，费用由该服务收取。重新运行会开始新对话。';
    byId('stop-model').hidden = !pending;
    byId('stop-model').disabled = Boolean(pending?.controller.signal.aborted);
    byId('model-run').setAttribute('aria-busy', String(Boolean(pending)));
    byId('send-followup').disabled = Boolean(pending);
  }
  function clearAnswer() {
    history = [];
    latest = '';
    byId('conversation').replaceChildren();
    byId('generated-images').replaceChildren();
    byId('model-answer').hidden = true;
    byId('copy-answer').hidden = true;
    byId('followup-form').hidden = true;
    byId('followup').value = '';
    byId('model-run-error').hidden = true;
    byId('model-run-error').textContent = '';
    byId('model-run-status').textContent = '';
  }
  function resetRun() {
    if (pending) {
      const stopped = pending;
      pending = undefined;
      clearTimeout(stopped.timer);
      stopped.controller.abort();
    }
    clearAnswer();
    refreshConnection();
  }
  function appendMessage(label, text) {
    const entry = document.createElement('article');
    entry.className = 'conversation-entry';
    const heading = document.createElement('h4');
    heading.textContent = label;
    const content = document.createElement('pre');
    content.className = 'prose model-text';
    content.textContent = text;
    entry.append(heading, content);
    byId('conversation').append(entry);
  }
  function renderAnswer(answer, question, continued) {
    byId('model-answer').hidden = false;
    if (answer.text) {
      if (continued) appendMessage('你的追问', question);
      appendMessage('模型回答', answer.text);
      history = [...answer.messages, { role: 'assistant', content: answer.text }];
      latest = answer.text;
      byId('copy-answer').hidden = false;
      byId('followup-form').hidden = false;
      byId('followup').value = '';
      byId('model-run-status').textContent = answer.warning || '回答已返回，可以继续追问。';
    } else {
      for (const [index, source] of answer.images.entries()) {
        const image = document.createElement('img');
        image.alt = '模型生成的图片 ' + (index + 1);
        image.referrerPolicy = 'no-referrer';
        image.src = source;
        image.addEventListener('error', () => { if (image.isConnected) byId('model-run-status').textContent = '接口已返回图片，但浏览器无法显示；远程图片可能已过期。'; });
        const link = document.createElement('a');
        link.href = source;
        link.target = '_blank';
        link.rel = 'noopener noreferrer';
        link.textContent = '打开图片 ' + (index + 1);
        byId('generated-images').append(image, link);
      }
      byId('model-run-status').textContent = '图片已返回。画幅与风格请对照提示词检查。';
    }
  }
  async function run(question, continued = false) {
    if (pending || !compiled || recipe.target === 'video') return;
    const config = connections.get(kind());
    if (!config) { openSettings(); byId('settings-status').textContent = '请先填写并应用当前连接，再返回配方运行。'; return; }
    if (!question.trim() || (continued && question.length > 12000)) {
      byId('model-run-error').textContent = '请填写追问，最多 12000 字。';
      byId('model-run-error').hidden = false;
      byId('followup').focus();
      return;
    }
    if (!continued) clearAnswer();
    byId('model-run-error').hidden = true;
    const operation = { controller: new AbortController(), timedOut: false };
    pending = operation;
    operation.timer = setTimeout(() => { operation.timedOut = true; operation.controller.abort(); }, config.kind === 'image' ? 180000 : 120000);
    refreshConnection();
    byId('model-run-status').textContent = '请求已发送到 ' + new URL(config.endpoint).host + '，正在等待结果…';
    try {
      const answer = await client.execute(config, question, continued ? history : [], operation.controller.signal);
      if (pending !== operation || operation.controller.signal.aborted) return;
      renderAnswer(answer, question, continued);
    } catch (error) {
      if (pending !== operation) return;
      byId('model-run-status').textContent = '';
      byId('model-run-error').textContent = operation.controller.signal.aborted
        ? (operation.timedOut ? '等待超时。' : '已停止等待。') + '服务端可能已开始处理并计费，请先检查服务端记录。'
        : error.message;
      byId('model-run-error').hidden = false;
    } finally {
      clearTimeout(operation.timer);
      if (pending === operation) { pending = undefined; refreshConnection(); }
    }
  }
  byId('open-model-settings').addEventListener('click', () => byId('model-settings').hidden ? openSettings() : closeSettings());
  byId('close-model-settings').addEventListener('click', closeSettings);
  byId('configure-model').addEventListener('click', () => openSettings());
  byId('model-kind').addEventListener('change', () => {
    drafts.set(editing, collect()); loadDraft(byId('model-kind').value);
    if (host?.embedded) { byId('magic-protocol-field').hidden = editing === 'image'; inlineRequest('keys'); }
  });
  byId('show-model-key').addEventListener('click', () => {
    const showing = byId('model-key').type === 'password';
    byId('model-key').type = showing ? 'text' : 'password';
    byId('show-model-key').textContent = showing ? '隐藏密钥' : '显示密钥';
    byId('show-model-key').setAttribute('aria-pressed', String(showing));
  });
  byId('model-settings-form').addEventListener('submit', event => {
    event.preventDefault();
    inlineVersion++; inlineBusy = false; inlineControls();
    try {
      const config = client.configure(collect());
      endLogin('已应用手动连接，先前的登录接入已取消。');
      connections.set(editing, config);
      drafts.set(editing, config);
      if (kind() === editing) resetRun();
      settingsError();
      byId('settings-status').textContent = (editing === 'image' ? '生图' : '文字') + '连接已应用：' + config.model + ' · ' + new URL(config.endpoint).host + '。尚未发送模型请求，返回配方即可运行。';
      refreshConnection();
    } catch (error) { settingsError(error); }
  });
  byId('clear-model-settings').addEventListener('click', () => {
    inlineVersion++; inlineBusy = false; inlineControls();
    endLogin('接入已取消；仅清除配方连接，不退出原站账号。');
    connections.delete(editing);
    drafts.delete(editing);
    loadDraft(editing);
    if (kind() === editing) resetRun();
    byId('settings-status').textContent = '当前连接与密钥已清除，其他用途的连接保留。';
    refreshConnection();
  });
  byId('run-model').addEventListener('click', () => run(byId('output').value));
  byId('stop-model').addEventListener('click', () => { if (pending) { byId('stop-model').disabled = true; pending.controller.abort(); } });
  byId('followup-form').addEventListener('submit', event => { event.preventDefault(); run(byId('followup').value.trim(), true); });
  byId('copy-answer').addEventListener('click', async () => {
    if (!latest) return;
    try {
      await navigator.clipboard.writeText(latest);
      byId('model-run-status').textContent = '最新回答已复制。';
    } catch { byId('model-run-status').textContent = '浏览器未允许复制，请选择最新回答手动复制。'; }
  });
  document.addEventListener('recipe-reset', event => { recipe = event.detail.recipe; compiled = false; resetRun(); });
  document.addEventListener('recipe-compiled', event => { recipe = event.detail.recipe; compiled = true; resetRun(); });
  byId('connect-magic-api').addEventListener('click', startLogin);
  byId('cancel-magic-connect').addEventListener('click', () => endLogin('已取消接入，不会接收先前窗口的配置。魔法 API 登录状态不受影响。'));
  window.addEventListener('message', receiveLogin);
  window.addEventListener('pagehide', () => endLogin());
})();
