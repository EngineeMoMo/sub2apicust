const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { createRequire } = require('node:module');
const frontendRequire = createRequire(path.resolve(__dirname, '../../../frontend/package.json'));
const { JSDOM, VirtualConsole } = frontendRequire('jsdom');
const html = fs.readFileSync(path.join(__dirname, 'index.html'), 'utf8');
function setup(options = {}) {
  const errors = [];
  const requests = [];
  const console = new VirtualConsole();
  console.on('jsdomError', error => errors.push(error.message));
  const dom = new JSDOM(html, { url: options.url || 'http://127.0.0.1:4178/' + (options.hash || ''), runScripts: 'dangerously', virtualConsole: console,
    beforeParse(window) {
      if (options.parent) Object.defineProperty(window, 'parent', { value: options.parent });
      window.matchMedia = () => ({ matches: false });
      window.open = options.open || (() => null);
      window.fetch = async (...args) => {
        requests.push(args);
        if (!options.fetch) throw new Error('测试禁止隐式模型调用');
        return options.fetch(...args);
      };
      window.HTMLElement.prototype.scrollIntoView = () => {};
      // jsdom 窗口缺少浏览器自带的流式解码能力，测试时注入 Node 的同名实现。
      window.TextDecoder = TextDecoder;
      window.TextEncoder = TextEncoder;
      window.confirm = () => { throw new Error('不能调用浏览器原生确认框'); };
      window.HTMLDialogElement.prototype.showModal = function () { this.setAttribute('open', ''); };
      window.HTMLDialogElement.prototype.close = function (value = '') {
        if (!this.open) return;
        this.returnValue = value;
        this.removeAttribute('open');
        this.dispatchEvent(new window.Event('close'));
      };
      Object.defineProperty(window.navigator, 'clipboard', { value: { writeText: options.writeText || (async () => {}) } });
    }
  });
  return { dom, window: dom.window, document: dom.window.document, errors, requests };
}
function submit(document) { document.getElementById('recipe-form').dispatchEvent(new document.defaultView.Event('submit', { cancelable: true })); }

test('切换生图用途先清除文字旧目录，权限失败显示具体原因，不保留可应用的旧模型', async () => {
  const messages=[], parent={postMessage:message=>messages.push(message)};
  const page=setup({url:'http://127.0.0.1:8080/recipes/?embedded=1',parent});
  const {window,document}=page,nonce='a'.repeat(64);
  const send=data=>window.dispatchEvent(new window.MessageEvent('message',{source:parent,origin:'http://127.0.0.1:8080',data:{nonce,...data}}));
  const reply=async data=>{send({type:'mofa-host-response',id:messages.at(-1).id,...data});await new Promise(resolve=>setImmediate(resolve));};
  try {
    send({type:'mofa-host-init',theme:'light'});document.getElementById('open-model-settings').click();
    await reply({result:[{id:1,name:'模拟文字密钥',group:'文字组'}]});
    const key=document.getElementById('magic-key-select');key.value='1';key.dispatchEvent(new window.Event('change'));
    await reply({result:{models:['fake-text-model'],protocol:'chat'}});
    const model=document.getElementById('magic-model-select');model.value='fake-text-model';model.dispatchEvent(new window.Event('change'));
    assert.equal(document.getElementById('apply-magic-config').disabled,false);
    const kind=document.getElementById('model-kind');kind.value='image';kind.dispatchEvent(new window.Event('change'));
    assert.equal(messages.at(-1).payload.kind,'image');assert.equal(model.value,'');assert.equal(key.options.length,1);
    await reply({error:'已有有效密钥，但所属分组均未开启生图权限。'});
    assert.match(document.getElementById('magic-connect-status').textContent,/分组均未开启生图权限/);
    assert.equal(key.disabled,true);assert.equal(model.disabled,true);assert.equal(document.getElementById('apply-magic-config').disabled,true);
    document.getElementById('connect-magic-api').click();await reply({result:[{id:2,name:'模拟生图密钥',group:'生图组'}]});
    key.value='2';key.dispatchEvent(new window.Event('change'));await reply({result:{models:['fake-image-model'],protocol:'chat'}});
    assert.equal(model.options[1].value,'fake-image-model');assert.equal(model.disabled,false);
    assert.equal(page.requests.length,0);assert.deepEqual(page.errors,[]);
  } finally {window.close();}
});

test('控制台内直接下拉选择配置，不打开新窗口或自动调用模型；退出清除密钥', async () => {
  const messages = [], parent = { postMessage: message => messages.push(message) };
  const page = setup({ url: 'http://127.0.0.1:8080/recipes/?embedded=1', parent, open: () => assert.fail('不得开新窗口') });
  const { window, document } = page, nonce = 'a'.repeat(64);
  const send = data => window.dispatchEvent(new window.MessageEvent('message', { source: parent, origin: 'http://127.0.0.1:8080', data: { nonce, ...data } }));
  const answer = async result => {
    const message = messages.at(-1); send({ type: 'mofa-host-response', id: message.id, result });
    await new Promise(resolve => setImmediate(resolve));
  };
  try {
    send({ type: 'mofa-host-init', theme: 'dark' });
    assert.equal(document.documentElement.dataset.theme, 'dark');
    document.getElementById('open-model-settings').click();
    assert.equal(messages.at(-1).action, 'keys');
    await answer([{ id: 1, name: '模拟密钥', group: '模拟分组' }]);
    const key = document.getElementById('magic-key-select'); key.value = '1'; key.dispatchEvent(new window.Event('change'));
    assert.equal(messages.at(-1).action, 'models');
    await answer({ models: ['fake-model'], protocol: 'responses' });
    const model = document.getElementById('magic-model-select'); model.value = 'fake-model'; model.dispatchEvent(new window.Event('change'));
    document.getElementById('apply-magic-config').click(); assert.equal(messages.at(-1).action, 'apply');
    await answer({ kind: 'text', protocol: 'responses', base: 'https://api.test/v1', key: 'fake-key', model: 'fake-model', size: '' });
    assert.equal(document.getElementById('model-key').value, 'fake-key');
    assert.equal(page.requests.length, 0);
    send({ type: 'mofa-host-theme', theme: 'light' }); assert.equal(document.documentElement.dataset.theme, 'light');
    send({ type: 'mofa-host-session-ended' }); assert.equal(document.getElementById('model-key').value, '');
    assert.equal(window.localStorage.getItem('fake-key'), null);
    assert.deepEqual(page.errors, []);
  } finally { window.close(); }
});
test('家族本机入口仅预填允许的网站来源，无隐式登录、存储或模型请求', () => {
  const page = setup({ hash: '?api_site=' + encodeURIComponent('http://127.0.0.1:4175') });
  try {
    assert.equal(page.document.getElementById('magic-site').value, 'http://127.0.0.1:4175');
    assert.equal(page.document.getElementById('magic-family-home').href, 'http://127.0.0.1:4175/family');
    assert.equal(page.document.getElementById('magic-family-home').target, '_blank');
    assert.equal(page.document.getElementById('magic-family-home').rel, 'noopener noreferrer');
    assert.equal(page.requests.length, 0);
    assert.equal(page.window.localStorage.length, 0);
    assert.equal(page.window.sessionStorage.length, 0);
    assert.deepEqual(page.errors, []);
  } finally { page.dom.window.close(); }
});
test('恶意网站提示不能改写官方登录地址，仍可手动设置连接', () => {
  const page = setup({ hash: '?api_site=' + encodeURIComponent('https://evil.test') });
  try {
    assert.equal(page.document.getElementById('magic-site').value, 'https://ai.mofamilys.com');
    assert.equal(page.document.getElementById('magic-family-home').href, 'https://ai.mofamilys.com/family');
    assert.match(page.document.getElementById('magic-connect-status').textContent, /不受支持/);
    assert.equal(page.requests.length, 0);
    assert.deepEqual(page.errors, []);
  } finally { page.dom.window.close(); }
});
test('公开配方来源不能通过链接提示自动切换到本机登录站点', () => {
  const page = setup({ url: 'https://recipes.example.test/?api_site=' + encodeURIComponent('http://127.0.0.1:4175') });
  try {
    assert.equal(page.document.getElementById('magic-site').value, 'https://ai.mofamilys.com');
    assert.equal(page.document.getElementById('magic-family-home').href, 'https://ai.mofamilys.com/family');
    assert.match(page.document.getElementById('magic-connect-status').textContent, /不受支持/);
    assert.equal(page.requests.length, 0);
    assert.deepEqual(page.errors, []);
  } finally { page.dom.window.close(); }
});
test('创作仅两项，媒体示例整理后引导到对应生成工具', () => {
  const page = setup();
  try {
    page.document.querySelector('[data-category="creative"]').click();
    assert.equal(page.document.querySelectorAll('[data-recipe]').length, 2);
    assert.match(page.document.getElementById('assembly-note').textContent, /不调用模型/);
    assert.match(page.document.querySelector('#recipe-form [type="submit"]').textContent, /整理并生成提示词/);
    for (const [id, destination] of [['image', '支持生图'], ['video', '支持视频生成']]) {
      page.document.querySelector('[data-recipe="' + id + '"]').click();
      page.document.getElementById('load-example').click();
      submit(page.document);
      assert.ok(page.document.getElementById('result-description').textContent.includes(destination));
      assert.ok(page.document.getElementById('result-next').textContent.includes(destination));
      assert.match(page.document.getElementById('status').textContent, /未调用模型/);
      assert.ok(!page.document.getElementById('output').value.includes('{{field:'));
    }
    assert.deepEqual(page.errors, []);
  } finally { page.dom.window.close(); }
});
test('创作可选字段留空使用明确默认值，不冒充已设置参数', () => {
  for (const id of ['image', 'video']) {
    const page = setup({ hash: '#' + id });
    try {
      const recipe = JSON.parse(page.document.getElementById('recipe-data').textContent).find(item => item.id === id);
      for (const field of recipe.fields.filter(item => item.required)) input(page.document, 'input-' + field.id, field.example);
      submit(page.document);
      const output = page.document.getElementById('output').value;
      assert.ok(output.length > 0);
      for (const field of recipe.fields.filter(item => !item.required && item.emptyValue)) assert.ok(output.includes(field.emptyValue), id + ':' + field.id);
    } finally { page.dom.window.close(); }
  }
});
function input(document, id, value) {
  const control = document.getElementById(id);
  control.value = value;
  control.dispatchEvent(new document.defaultView.Event('input', { bubbles: true }));
}
function configure(page, values = {}) {
  page.document.getElementById('open-model-settings').click();
  page.document.getElementById('model-kind').value = values.kind || 'text';
  page.document.getElementById('model-kind').dispatchEvent(new page.window.Event('change'));
  for (const [id, value] of Object.entries({ base: 'https://example.test/v1', key: 'fake-test-key', name: 'test-model', protocol: 'chat', ...values })) {
    if (id !== 'kind') input(page.document, 'model-' + id, value);
  }
  page.document.getElementById('model-settings-form').dispatchEvent(new page.window.Event('submit', { cancelable: true }));
}
const settled = () => new Promise(resolve => setImmediate(resolve));
function loginMessage(page, popup, url, config, overrides = {}) {
  page.window.dispatchEvent(new page.window.MessageEvent('message', { source: popup, origin: url.origin,
    data: { type: 'mofa-recipes-connection', nonce: url.searchParams.get('nonce'), config }, ...overrides }));
}
test('登录导入精确配对、一次性应用，绝不自动发模型请求或保存令牌', () => {
  const acknowledgements = [];
  const popup = { closed: false, postMessage: (...args) => acknowledgements.push(args) };
  let url;
  const page = setup({ open: address => { url = new URL(address); return popup; } });
  try {
    page.document.getElementById('connect-magic-api').click();
    const config = { kind: 'text', protocol: 'responses', base: 'https://api.test/v1', key: 'fake-import-key', model: 'imported-model', size: '' };
    for (const overrides of [{ source: {} }, { origin: 'https://evil.test' }, { data: { type: 'mofa-recipes-connection', nonce: 'bad', config } }]) loginMessage(page, popup, url, config, overrides);
    assert.equal(page.document.getElementById('model-key').value, '');
    loginMessage(page, popup, url, config);
    assert.equal(page.document.getElementById('model-key').value, 'fake-import-key');
    assert.equal(page.document.getElementById('model-key').type, 'password');
    assert.equal(page.document.getElementById('model-name').value, 'imported-model');
    assert.deepEqual(JSON.parse(JSON.stringify(acknowledgements)), [[{ type: 'mofa-recipes-applied', nonce: url.searchParams.get('nonce') }, 'https://ai.mofamilys.com']]);
    loginMessage(page, popup, url, { ...config, key: 'replayed-key' });
    assert.equal(page.document.getElementById('model-key').value, 'fake-import-key');
    assert.equal(page.requests.length, 0);
    assert.equal(page.window.localStorage.length, 0);
    assert.equal(page.window.sessionStorage.length, 0);
    assert.equal(page.document.cookie, '');
    assert.deepEqual(page.errors, []);
  } finally { page.window.close(); }
});
test('取消或过期后拒绝回传，窗口阻止与仿冒登录有明确说明', () => {
  const popup = { closed: false, postMessage: () => assert.fail('不应确认') };
  let url;
  const page = setup({ open: address => { url = new URL(address); return popup; } });
  const config = { kind: 'text', protocol: 'chat', base: 'https://api.test/v1', key: 'fake-key', model: 'model' };
  try {
    page.document.getElementById('connect-magic-api').click();
    page.document.getElementById('cancel-magic-connect').click();
    loginMessage(page, popup, url, config);
    assert.equal(page.document.getElementById('model-key').value, '');
    page.document.getElementById('connect-magic-api').click();
    const now = page.window.Date.now();
    page.window.Date.now = () => now + 301000;
    loginMessage(page, popup, url, config);
    assert.equal(page.document.getElementById('model-key').value, '');
    page.document.getElementById('cancel-magic-connect').click();
    input(page.document, 'magic-site', 'https://evil.test');
    page.document.getElementById('connect-magic-api').click();
    assert.match(page.document.getElementById('magic-connect-status').textContent, /不接受|其他网站/);
  } finally { page.window.close(); }
  const blocked = setup();
  try {
    blocked.document.getElementById('connect-magic-api').click();
    assert.match(blocked.document.getElementById('magic-connect-status').textContent, /阻止了登录窗口/);
  } finally { blocked.window.close(); }
});
test('生图登录用途固定，切换编辑用途不会导入到错误连接', () => {
  let url;
  const popup = { closed: false, postMessage: () => {} };
  const page = setup({ open: address => { url = new URL(address); return popup; } });
  try {
    const kindControl = page.document.getElementById('model-kind');
    kindControl.value = 'image';
    kindControl.dispatchEvent(new page.window.Event('change'));
    page.document.getElementById('connect-magic-api').click();
    assert.equal(url.searchParams.get('kind'), 'image');
    kindControl.value = 'text';
    kindControl.dispatchEvent(new page.window.Event('change'));
    loginMessage(page, popup, url, { kind: 'image', protocol: 'images', base: 'https://api.test/v1', key: 'fake-image-key', model: 'image-model' });
    assert.equal(page.document.getElementById('model-key').value, '');
    kindControl.value = 'image';
    kindControl.dispatchEvent(new page.window.Event('change'));
    assert.equal(page.document.getElementById('model-key').value, 'fake-image-key');
    assert.equal(page.document.getElementById('protocol-field').hidden, true);
    assert.equal(page.requests.length, 0);
  } finally { page.window.close(); }
});
test('无效导入与手动覆盖会终止接入，清除连接不访问原站注销接口', () => {
  let url;
  const popup = { closed: false, postMessage: () => assert.fail('不应确认') };
  const page = setup({ open: address => { url = new URL(address); return popup; } });
  try {
    page.document.getElementById('connect-magic-api').click();
    loginMessage(page, popup, url, { kind: 'text', protocol: 'chat', base: 'http://unsafe.test', key: 'fake-key', model: 'test' });
    assert.match(page.document.getElementById('magic-connect-status').textContent, /无法应用/);
    assert.equal(page.document.getElementById('model-key').value, '');
    page.document.getElementById('connect-magic-api').click();
    const oldURL = url;
    configure(page);
    loginMessage(page, popup, oldURL, { kind: 'text', protocol: 'chat', base: 'https://api.test', key: 'late-key', model: 'late' });
    assert.equal(page.document.getElementById('model-key').value, 'fake-test-key');
    page.document.getElementById('clear-model-settings').click();
    assert.equal(page.document.getElementById('model-key').value, '');
    assert.equal(page.requests.length, 0);
  } finally { page.window.close(); }
});
const jsonReply = payload => ({ ok: true, json: async () => payload });
test('设置保留材料和提示词，应用不调用模型，刷新不保留密钥', () => {
  const page = setup();
  try {
    page.document.getElementById('load-example').click(); submit(page.document);
    const prompt = page.document.getElementById('output').value;
    configure(page);
    assert.match(page.document.getElementById('settings-status').textContent, /连接已应用/);
    assert.equal(page.document.getElementById('output').value, prompt);
    assert.equal(page.requests.length, 0);
    assert.equal(page.document.getElementById('model-key').type, 'password');
    page.document.getElementById('show-model-key').click();
    assert.equal(page.document.getElementById('model-key').type, 'text');
    assert.equal(page.window.localStorage.length, 0);
    assert.equal(page.window.sessionStorage.length, 0);
    const clean = setup();
    assert.equal(clean.document.getElementById('model-key').value, '');
    clean.dom.window.close();
    page.document.getElementById('clear-model-settings').click();
    assert.equal(page.document.getElementById('model-key').value, '');
    assert.match(page.document.getElementById('connection-summary').textContent, /尚未设置/);
  } finally { page.dom.window.close(); }
});
test('无连接运行定位设置，无效配置不发请求', () => {
  const page = setup();
  try {
    page.document.getElementById('load-example').click(); submit(page.document);
    page.document.getElementById('run-model').click();
    assert.equal(page.document.getElementById('model-settings').hidden, false);
    assert.equal(page.document.activeElement.id, 'model-kind');
    input(page.document, 'model-name', 'test-model');
    input(page.document, 'model-base', 'http://remote.test/v1');
    page.document.getElementById('model-settings-form').dispatchEvent(new page.window.Event('submit', { cancelable: true }));
    assert.match(page.document.getElementById('settings-error').textContent, /HTTPS/);
    assert.equal(page.document.activeElement.id, 'model-base');
    assert.equal(page.requests.length, 0);
  } finally { page.dom.window.close(); }
});
test('文字运行和追问带上历史，回答只做安全排版不解释HTML，编辑材料清除旧对话', async () => {
  const page = setup({ fetch: async () => jsonReply({ choices: [{ message: { content: '<img src=x onerror=alert(1)>模拟回答' } }] }) });
  try {
    configure(page);
    page.document.getElementById('load-example').click(); submit(page.document);
    page.document.getElementById('run-model').click();
    assert.equal(page.document.getElementById('run-model').disabled, true);
    await settled();
    assert.equal(page.document.getElementById('followup-form').hidden, false);
    assert.match(page.document.getElementById('conversation').textContent, /模拟回答/);
    assert.equal(page.document.querySelectorAll('#conversation img').length, 0);
    input(page.document, 'followup', '继续解释');
    page.document.getElementById('followup-form').dispatchEvent(new page.window.Event('submit', { cancelable: true }));
    await settled();
    const body = JSON.parse(page.requests[1][1].body);
    assert.deepEqual(body.messages.map(item => item.role), ['user', 'assistant', 'user']);
    assert.equal(body.messages[2].content, '继续解释');
    assert.equal(page.document.getElementById('followup').value, '');
    input(page.document, 'input-material', '更新材料');
    assert.equal(page.document.getElementById('model-run').hidden, true);
    assert.equal(page.document.getElementById('conversation').textContent, '');
    assert.deepEqual(page.errors, []);
  } finally { page.dom.window.close(); }
});
test('Responses 连接发送 input 并显示结果', async () => {
  const page = setup({ fetch: async () => jsonReply({ status: 'completed', output: [{ type: 'message', content: [{ type: 'output_text', text: '模拟 Responses 回答' }] }] }) });
  try {
    configure(page, { protocol: 'responses' });
    page.document.getElementById('load-example').click(); submit(page.document);
    page.document.getElementById('run-model').click(); await settled();
    assert.equal(page.requests[0][0], 'https://example.test/v1/responses');
    assert.equal(JSON.parse(page.requests[0][1].body).store, false);
    assert.match(page.document.getElementById('conversation').textContent, /模拟 Responses 回答/);
  } finally { page.dom.window.close(); }
});
test('文字与生图连接分开，生图返回图片，视频不发送文字请求', async () => {
  const page = setup({ hash: '#image', fetch: async () => jsonReply({ data: [{ b64_json: 'aGVsbG8=' }] }) });
  try {
    configure(page);
    configure(page, { kind: 'image', name: 'test-image-model', key: 'fake-image-key', size: '1024x1024' });
    page.document.getElementById('load-example').click(); submit(page.document);
    page.document.getElementById('run-model').click(); await settled();
    assert.equal(page.requests[0][0], 'https://example.test/v1/images/generations');
    assert.equal(page.requests[0][1].headers.Authorization, 'Bearer fake-image-key');
    assert.equal(page.document.querySelectorAll('#generated-images img').length, 1);
    assert.equal(page.document.getElementById('followup-form').hidden, true);
    page.document.querySelector('[data-recipe="meeting"]').click();
    page.document.getElementById('load-example').click(); submit(page.document);
    assert.match(page.document.getElementById('connection-summary').textContent, /test-model/);
    page.document.querySelector('[data-recipe="video"]').click();
    page.document.getElementById('load-example').click(); submit(page.document);
    assert.equal(page.document.getElementById('run-model').disabled, true);
    assert.match(page.document.getElementById('connection-summary').textContent, /视频接口尚未接入/);
    assert.equal(page.requests.length, 1);
  } finally { page.dom.window.close(); }
});
test('Base64图片在页内打开原图并提供下载入口，关闭和Esc恢复按钮焦点，不创建新窗口', async () => {
  const opened = [];
  const page = setup({ hash: '#image', open: (...args) => opened.push(args), fetch: async () => jsonReply({ data: [{ b64_json: 'aGVsbG8=' }] }) });
  try {
    configure(page, { kind: 'image', name: 'test-image-model' });
    page.document.getElementById('load-example').click(); submit(page.document);
    page.document.getElementById('run-model').click(); await settled();
    const button = page.document.querySelector('#generated-images button');
    assert.ok(button, '打开图片必须是页内预览按钮，不能把data地址交给新窗口');
    assert.equal(page.document.querySelectorAll('#generated-images a[target="_blank"]').length, 0);
    button.click();
    const viewer = page.document.getElementById('generated-image-viewer');
    const image = page.document.getElementById('generated-image-full');
    const close = page.document.getElementById('generated-image-close');
    const download = page.document.getElementById('generated-image-download');
    assert.equal(viewer.open, true);
    assert.equal(image.src, page.document.querySelector('#generated-images img').src);
    assert.equal(download.href, image.src);
    assert.equal(download.download, 'mofa-recipes-image-1.png');
    assert.equal(download.target, '');
    assert.equal(page.document.activeElement, close);
    close.dispatchEvent(new page.window.KeyboardEvent('keydown', { key: 'Tab', shiftKey: true, bubbles: true, cancelable: true }));
    assert.equal(page.document.activeElement, download);
    download.dispatchEvent(new page.window.KeyboardEvent('keydown', { key: 'Tab', bubbles: true, cancelable: true }));
    assert.equal(page.document.activeElement, close);
    close.click();
    assert.equal(viewer.open, false);
    assert.equal(image.getAttribute('src'), null);
    assert.equal(download.getAttribute('href'), null);
    assert.equal(page.document.activeElement, button);
    button.click(); viewer.dispatchEvent(new page.window.Event('cancel', { cancelable: true }));
    assert.equal(viewer.open, false);
    assert.equal(page.document.activeElement, button);
    assert.deepEqual(opened, []);
    assert.equal(page.requests.length, 1);
    assert.deepEqual(page.errors, []);
  } finally { page.dom.window.close(); }
});

test('多图序号和JPEG下载类型正确，修改材料或离开页面关闭并清除原图预览', async () => {
  const page = setup({ hash: '#image', fetch: async () => jsonReply({ output_format: 'jpeg', data: [{ b64_json: 'aGVsbG8=' }, { b64_json: 'd29ybGQ=' }] }) });
  try {
    configure(page, { kind: 'image', name: 'test-image-model' });
    page.document.getElementById('load-example').click(); submit(page.document);
    page.document.getElementById('run-model').click(); await settled();
    page.document.querySelectorAll('#generated-images button')[1].click();
    assert.equal(page.document.getElementById('generated-image-title').textContent, '图片 2 · 原图预览');
    assert.equal(page.document.getElementById('generated-image-full').src, 'data:image/jpeg;base64,d29ybGQ=');
    assert.equal(page.document.getElementById('generated-image-download').download, 'mofa-recipes-image-2.jpeg');
    page.window.dispatchEvent(new page.window.Event('pagehide'));
    assert.equal(page.document.getElementById('generated-image-viewer').open, false);
    assert.equal(page.document.getElementById('generated-image-full').getAttribute('src'), null);
    page.document.querySelectorAll('#generated-images button')[0].click();
    input(page.document, 'input-scene', '已更改的场景');
    assert.equal(page.document.getElementById('generated-image-viewer').open, false);
    assert.equal(page.document.getElementById('generated-image-full').getAttribute('src'), null);
    assert.equal(page.document.getElementById('generated-image-download').getAttribute('href'), null);
    assert.equal(page.document.querySelectorAll('#generated-images button').length, 0);
    assert.equal(page.requests.length, 1);
  } finally { page.dom.window.close(); }
});

test('远程图片保持原地址在页内预览，过期错误可见，不伪造跨域下载', async () => {
  const source = 'https://images.example.test/result.webp';
  const page = setup({ hash: '#image', fetch: async () => jsonReply({ data: [{ url: source }] }) });
  try {
    configure(page, { kind: 'image', name: 'test-image-model' });
    page.document.getElementById('load-example').click(); submit(page.document);
    page.document.getElementById('run-model').click(); await settled();
    const button = page.document.querySelector('#generated-images button'); button.click();
    assert.equal(page.document.getElementById('generated-image-full').src, source);
    assert.equal(page.document.getElementById('generated-image-download').hidden, true);
    assert.equal(page.document.getElementById('generated-image-download').getAttribute('href'), null);
    assert.equal(page.document.getElementById('generated-image-save-note').hidden, false);
    page.document.getElementById('generated-image-full').dispatchEvent(new page.window.Event('error'));
    assert.equal(page.document.getElementById('generated-image-error').hidden, false);
    page.document.getElementById('generated-image-close').click(); button.click();
    assert.equal(page.document.getElementById('generated-image-error').hidden, true);
    assert.equal(page.requests.length, 1);
  } finally { page.dom.window.close(); }
});

test('不支持dialog的浏览器保留图片并给出保存方法，不打开空白页', async () => {
  const page = setup({ hash: '#image', open: () => { throw new Error('不允许图片新窗口导航'); }, fetch: async () => jsonReply({ data: [{ b64_json: 'aGVsbG8=' }] }) });
  try {
    configure(page, { kind: 'image', name: 'test-image-model' });
    page.document.getElementById('load-example').click(); submit(page.document);
    page.document.getElementById('run-model').click(); await settled();
    page.document.getElementById('generated-image-viewer').showModal = undefined;
    page.document.querySelector('#generated-images button').click();
    assert.match(page.document.getElementById('model-run-status').textContent, /长按或右键保存/);
    assert.equal(page.document.querySelectorAll('#generated-images img').length, 1);
    assert.equal(page.document.getElementById('generated-image-viewer').open, false);
    assert.deepEqual(page.errors, []);
  } finally { page.dom.window.close(); }
});

test('停止等待可取消，切换配方后迟到结果不污染新配方', async () => {
  const cancelled = setup({ fetch: (_url, options) => new Promise((_resolve, reject) => options.signal.addEventListener('abort', () => reject(new Error('aborted')), { once: true })) });
  try {
    configure(cancelled);
    cancelled.document.getElementById('load-example').click(); submit(cancelled.document);
    cancelled.document.getElementById('run-model').click();
    cancelled.document.getElementById('stop-model').click(); await settled();
    assert.match(cancelled.document.getElementById('model-run-error').textContent, /已停止等待/);
    assert.equal(cancelled.document.getElementById('run-model').disabled, false);
  } finally { cancelled.dom.window.close(); }
  let finish;
  const page = setup({ fetch: () => new Promise(resolve => { finish = resolve; }) });
  try {
    configure(page);
    page.document.getElementById('load-example').click(); submit(page.document);
    page.document.getElementById('run-model').click();
    page.document.querySelector('[data-recipe="paper"]').click();
    assert.equal(page.requests[0][1].signal.aborted, true);
    finish(jsonReply({ choices: [{ message: { content: '旧配方回答' } }] }));
    await settled();
    assert.equal(page.document.getElementById('model-answer').hidden, true);
    assert.equal(page.document.getElementById('conversation').textContent, '');
  } finally { page.dom.window.close(); }
});
test('接口拒绝显示可恢复错误，失败追问保留输入与已成功对话', async () => {
  let calls = 0;
  const page = setup({ fetch: async () => ++calls === 1 ? jsonReply({ choices: [{ message: { content: '第一轮回答' } }] }) : { ok: false, status: 401 } });
  try {
    configure(page);
    page.document.getElementById('load-example').click(); submit(page.document);
    page.document.getElementById('run-model').click(); await settled();
    input(page.document, 'followup', '待发送问题');
    page.document.getElementById('followup-form').dispatchEvent(new page.window.Event('submit', { cancelable: true }));
    await settled();
    assert.match(page.document.getElementById('model-run-error').textContent, /API Key 无效/);
    assert.equal(page.document.getElementById('followup').value, '待发送问题');
    assert.match(page.document.getElementById('conversation').textContent, /第一轮回答/);
    assert.equal(calls, 2);
  } finally { page.dom.window.close(); }
});

test('全部11份配方虚构示例均能生成且字段不串位', () => {
  const page = setup();
  try {
    const recipes = JSON.parse(page.document.getElementById('recipe-data').textContent);
    for (const recipe of recipes) {
      page.document.querySelector('[data-recipe="' + recipe.id + '"]').click();
      page.document.getElementById('load-example').click();
      submit(page.document);
      const result = page.document.getElementById('output').value;
      assert.equal(page.document.getElementById('result').hidden, false);
      assert.ok(!result.includes('{{field:'));
      for (const field of recipe.fields) assert.ok(result.includes(field.example.trim()), recipe.id + ':' + field.id);
    }
    assert.deepEqual(page.errors, []);
  } finally { page.dom.window.close(); }
});
test('必填错误可定位，编辑后旧提示词立即失效', () => {
  const page = setup();
  try {
    submit(page.document);
    assert.equal(page.document.activeElement.id, 'input-material');
    assert.equal(page.document.getElementById('input-material').getAttribute('aria-invalid'), 'true');
    page.document.getElementById('load-example').click();
    submit(page.document);
    input(page.document, 'input-material', '更新后的材料');
    assert.equal(page.document.getElementById('result').hidden, true);
    assert.equal(page.document.getElementById('output').value, '');
  } finally { page.dom.window.close(); }
});
test('大学筛选只有三项，搜索无结果可恢复', () => {
  const page = setup();
  try {
    page.document.querySelector('[data-category="university"]').click();
    assert.equal(page.document.querySelectorAll('[data-recipe]').length, 3);
    input(page.document, 'search', '小学');
    assert.equal(page.document.getElementById('empty').hidden, false);
    page.document.getElementById('reset-search').click();
    assert.equal(page.document.querySelectorAll('[data-recipe]').length, 11);
  } finally { page.dom.window.close(); }
});
test('切换配方保留本页草稿，重新打开不持久化', () => {
  const page = setup();
  try {
    input(page.document, 'input-material', '仅本页的材料');
    page.document.querySelector('[data-recipe="paper"]').click();
    page.document.querySelector('[data-recipe="meeting"]').click();
    assert.equal(page.document.getElementById('input-material').value, '仅本页的材料');
    assert.equal(page.window.localStorage.length, 0);
    assert.equal(page.window.sessionStorage.length, 0);
    const clean = setup();
    assert.equal(clean.document.getElementById('input-material').value, '');
    clean.dom.window.close();
  } finally { page.dom.window.close(); }
});
test('用户材料不作为HTML执行，保留原始文本', () => {
  const page = setup();
  try {
    const originalImages = page.document.querySelectorAll('img').length;
    const payload = '</textarea><img src=x onerror="alert(1)">{{field:date}}';
    input(page.document, 'input-material', payload);
    submit(page.document);
    assert.ok(page.document.getElementById('output').value.includes(payload));
    assert.equal(page.document.querySelectorAll('img').length, originalImages);
    assert.equal(page.document.querySelector('img[onerror]'), null);
    assert.deepEqual(page.errors, []);
  } finally { page.dom.window.close(); }
});
test('取消页面确认框不覆盖或清空材料，并恢复触发按钮焦点', async () => {
  const page = setup();
  try {
    input(page.document, 'input-material', '需要保留');
    page.document.getElementById('load-example').click();
    assert.equal(page.document.getElementById('confirm-dialog').open, true);
    assert.equal(page.document.activeElement.id, 'confirm-cancel');
    page.document.getElementById('confirm-cancel').click();
    await new Promise(resolve => setImmediate(resolve));
    assert.equal(page.document.getElementById('confirm-dialog').open, false);
    assert.equal(page.document.activeElement.id, 'load-example');
    page.document.getElementById('clear').click();
    page.document.getElementById('confirm-close').click();
    await new Promise(resolve => setImmediate(resolve));
    assert.equal(page.document.getElementById('input-material').value, '需要保留');
    assert.equal(page.document.activeElement.id, 'clear');
    assert.deepEqual(page.errors, []);
  } finally { page.dom.window.close(); }
});
test('确认替换与确认清空执行各自操作并使旧结果失效', async () => {
  const page = setup();
  try {
    input(page.document, 'input-material', '原材料');
    submit(page.document);
    page.document.getElementById('load-example').click();
    assert.match(page.document.getElementById('confirm-title').textContent, /替换/);
    assert.equal(page.document.getElementById('input-material').value, '原材料');
    page.document.getElementById('confirm-accept').click();
    await new Promise(resolve => setImmediate(resolve));
    assert.match(page.document.getElementById('input-material').value, /小林/);
    assert.equal(page.document.getElementById('result').hidden, true);
    submit(page.document);
    page.document.getElementById('clear').click();
    assert.equal(page.document.getElementById('confirm-dialog').dataset.destructive, 'true');
    page.document.getElementById('confirm-accept').click();
    await new Promise(resolve => setImmediate(resolve));
    assert.equal(page.document.getElementById('input-material').value, '');
    assert.equal(page.document.getElementById('output').value, '');
    assert.equal(page.document.getElementById('example-notice').hidden, true);
  } finally { page.dom.window.close(); }
});
test('Escape取消语义保持原材料，再次打开不沿用旧的确认结果', async () => {
  const page = setup();
  try {
    input(page.document, 'input-material', '保留内容');
    page.document.getElementById('clear').click();
    page.document.getElementById('confirm-dialog').dispatchEvent(new page.window.Event('cancel', { cancelable: true }));
    await new Promise(resolve => setImmediate(resolve));
    assert.equal(page.document.getElementById('input-material').value, '保留内容');
    assert.equal(page.document.getElementById('confirm-dialog').open, false);
    page.document.getElementById('load-example').click();
    assert.equal(page.document.getElementById('confirm-dialog').returnValue, '');
    assert.equal(page.document.getElementById('confirm-dialog').dataset.destructive, 'false');
    page.document.getElementById('confirm-cancel').click();
    await new Promise(resolve => setImmediate(resolve));
    assert.equal(page.document.getElementById('input-material').value, '保留内容');
  } finally { page.dom.window.close(); }
});
test('确认期间切换配方只取消，不修改另一份草稿', async () => {
  const page = setup();
  try {
    input(page.document, 'input-material', '会议原材料');
    page.document.getElementById('clear').click();
    page.document.querySelector('[data-recipe="paper"]').click();
    await new Promise(resolve => setImmediate(resolve));
    assert.equal(page.document.getElementById('confirm-dialog').open, false);
    assert.equal(page.document.getElementById('input-material').value, '');
    page.document.querySelector('[data-recipe="meeting"]').click();
    assert.equal(page.document.getElementById('input-material').value, '会议原材料');
  } finally { page.dom.window.close(); }
});
test('Tab与Shift+Tab在弹框按钮间循环', () => {
  const page = setup();
  try {
    input(page.document, 'input-material', '保留内容');
    page.document.getElementById('clear').click();
    page.document.getElementById('confirm-accept').focus();
    page.document.activeElement.dispatchEvent(new page.window.KeyboardEvent('keydown', { key: 'Tab', bubbles: true, cancelable: true }));
    assert.equal(page.document.activeElement.id, 'confirm-close');
    page.document.activeElement.dispatchEvent(new page.window.KeyboardEvent('keydown', { key: 'Tab', shiftKey: true, bubbles: true, cancelable: true }));
    assert.equal(page.document.activeElement.id, 'confirm-accept');
  } finally { page.dom.window.close(); }
});
test('不支持页面对话框时保留材料，不回退原生confirm', async () => {
  const page = setup();
  try {
    input(page.document, 'input-material', '保留内容');
    page.document.getElementById('confirm-dialog').showModal = undefined;
    page.document.getElementById('clear').click();
    await new Promise(resolve => setImmediate(resolve));
    assert.equal(page.document.getElementById('input-material').value, '保留内容');
    assert.match(page.document.getElementById('status').textContent, /材料未改动/);
    assert.deepEqual(page.errors, []);
  } finally { page.dom.window.close(); }
});
test('复制成功传递完整文本，失败提供选中回退', async () => {
  let copied;
  const page = setup({ writeText: async text => { copied = text; } });
  try {
    page.document.getElementById('load-example').click(); submit(page.document);
    page.document.getElementById('copy').click(); await new Promise(resolve => setImmediate(resolve));
    assert.equal(copied, page.document.getElementById('output').value);
    assert.match(page.document.getElementById('status').textContent, /已复制/);
  } finally { page.dom.window.close(); }
  const failed = setup({ writeText: async () => { throw new Error('denied'); } });
  try {
    failed.document.getElementById('load-example').click(); submit(failed.document);
    failed.document.getElementById('copy').click(); await new Promise(resolve => setImmediate(resolve));
    assert.match(failed.document.getElementById('status').textContent, /未允许自动复制/);
    const output = failed.document.getElementById('output');
    assert.equal(output.selectionEnd, output.value.length);
    assert.equal(failed.document.getElementById('copy').disabled, false);
  } finally { failed.dom.window.close(); }
});
test('剪贴板接口被拒时改用 execCommand 复制成功，按钮保持焦点、不留临时文本框', async () => {
  const page = setup({ writeText: async () => { throw new Error('denied'); } });
  const copied = [];
  try {
    page.window.document.execCommand = command => { copied.push({ command, text: page.document.activeElement?.value }); return command === 'copy'; };
    page.document.getElementById('load-example').click(); submit(page.document);
    const button = page.document.getElementById('copy');
    button.focus(); button.click(); await new Promise(resolve => setImmediate(resolve));
    assert.equal(copied.length, 1); assert.equal(copied[0].command, 'copy');
    assert.equal(copied[0].text, page.document.getElementById('output').value);
    assert.match(page.document.getElementById('status').textContent, /已复制/);
    assert.equal(page.document.activeElement, button); assert.equal(button.disabled, false); assert.equal(button.hasAttribute('aria-busy'), false);
    assert.equal(page.document.querySelectorAll('body > textarea').length, 0);
    assert.equal([...page.document.querySelectorAll('#recipe-steps li')].find(li => li.dataset.step === 'use').dataset.state, 'done');
    assert.deepEqual(page.errors, []);
  } finally { page.dom.window.close(); }
});
test('深浅色、方法视图和配方直达可用', () => {
  const page = setup({ hash: '#lab' });
  try {
    assert.equal(page.document.getElementById('recipe-title').textContent, '大学编程实验复盘');
    page.document.getElementById('theme').click();
    assert.equal(page.document.documentElement.dataset.theme, 'dark');
    page.document.getElementById('guide-view').click();
    assert.equal(page.document.getElementById('editor').hidden, true);
    page.document.getElementById('back-to-edit').click();
    assert.equal(page.document.getElementById('editor').hidden, false);
  } finally { page.dom.window.close(); }
});
test('教程有完整四步和模型边界，默认不遮挡配方', () => {
  const page = setup();
  try {
    assert.equal(page.document.getElementById('usage-tutorial').open, false);
    assert.equal(page.document.querySelectorAll('.tutorial-steps > li').length, 4);
    assert.match(page.document.getElementById('tutorial-summary').textContent, /整理不调用模型/);
    assert.match(page.document.querySelector('.tutorial-content').textContent, /新建对话/);
    assert.match(page.document.querySelector('.tutorial-content').textContent, /怎样检查 AI 的输出/);
    assert.match(page.document.querySelector('.tutorial-faq').textContent, /刷新或关闭页面会丢失输入/);
  } finally { page.dom.window.close(); }
});
test('打开和关闭教程不改变当前配方或草稿，返回后聚焦填写位置', () => {
  const page = setup({ hash: '#paper' });
  try {
    input(page.document, 'input-material', '论文原材料');
    page.document.getElementById('open-tutorial').click();
    assert.equal(page.document.getElementById('usage-tutorial').open, true);
    assert.equal(page.document.getElementById('open-tutorial').getAttribute('aria-expanded'), 'true');
    page.document.getElementById('tutorial-start').click();
    assert.equal(page.document.getElementById('usage-tutorial').open, false);
    assert.equal(page.document.getElementById('open-tutorial').getAttribute('aria-expanded'), 'false');
    assert.equal(page.document.getElementById('recipe-title').textContent, '大学论文精读');
    assert.equal(page.document.getElementById('input-material').value, '论文原材料');
    assert.equal(page.document.activeElement.id, 'input-course');
    assert.equal(page.window.location.hash, '#paper');
  } finally { page.dom.window.close(); }
});
test('生成结果可进入教程再返回，提示词和材料均保留', () => {
  const page = setup();
  try {
    page.document.getElementById('load-example').click();
    submit(page.document);
    const output = page.document.getElementById('output').value;
    page.document.getElementById('result-tutorial').click();
    assert.equal(page.document.getElementById('usage-tutorial').open, true);
    page.document.getElementById('tutorial-start').click();
    assert.equal(page.document.getElementById('output').value, output);
    assert.equal(page.document.getElementById('result').hidden, false);
    assert.match(page.document.querySelector('.result-next').textContent, /粘贴完整提示词并发送/);
    page.document.getElementById('usage-tutorial').open = true;
    page.document.getElementById('usage-tutorial').dispatchEvent(new page.window.Event('toggle'));
    assert.equal(page.document.getElementById('result-tutorial').getAttribute('aria-expanded'), 'true');
    assert.deepEqual(page.errors, []);
  } finally { page.dom.window.close(); }
});
test('初始页面和提示词整理不发送网络请求，也不保存密钥或材料', () => {
  const page = setup();
  try {
    assert.equal(page.document.querySelectorAll('script[src], link[href], iframe').length, 0);
    for (const image of page.document.querySelectorAll('img[src]')) assert.ok(image.src.startsWith('data:'));
    assert.equal(page.document.getElementById('generated-image-full').getAttribute('src'), null);
    page.document.getElementById('load-example').click(); submit(page.document);
    assert.equal(page.requests.length, 0);
    assert.equal(page.window.localStorage.length, 0);
    assert.equal(page.window.sessionStorage.length, 0);
    assert.match(page.document.querySelector('[http-equiv="Content-Security-Policy"]').content, /connect-src https:/);
    // 浏览器不接受 CSP 里的 IPv6 字面地址（会报错并忽略），策略里只保留 localhost 与 127.0.0.1。
    assert.doesNotMatch(page.document.querySelector('[http-equiv="Content-Security-Policy"]').content, /\[::1\]/);
  } finally { page.dom.window.close(); }
});
function sseReply(chunks, options = {}) {
  let index = 0;
  const encoder = new TextEncoder();
  return {
    ok: true,
    headers: { get: name => name.toLowerCase() === 'content-type' ? 'text/event-stream' : null },
    body: { getReader: () => ({
      read: () => {
        if (index < chunks.length) return Promise.resolve({ done: false, value: encoder.encode(chunks[index++]) });
        if (options.hold) return new Promise((_resolve, reject) => options.signal.addEventListener('abort', () => reject(new DOMException('Aborted', 'AbortError')), { once: true }));
        return Promise.resolve({ done: true });
      },
      cancel: () => Promise.resolve()
    }) }
  };
}
async function waitFor(check, label) {
  for (let attempt = 0; attempt < 80; attempt++) {
    if (check()) return;
    await new Promise(resolve => setTimeout(resolve, 10));
  }
  assert.fail('等待超时：' + label);
}
test('流式回答边到边显示，结束后表格排版，可看原文与单条复制，追问带上完整历史', async () => {
  const copied = [];
  let calls = 0;
  const page = setup({ writeText: async text => { copied.push(text); }, fetch: async (_url, options) => {
    calls++;
    assert.equal(JSON.parse(options.body).stream, true);
    return calls === 1
      ? sseReply(['data: {"choices":[{"delta":{"content":"| 编号 | 任务 |\\n"}}]}\n\n', 'data: {"choices":[{"delta":{"content":"| --- | --- |\\n| T1 | 提交初稿 |"}}]}\n\n', 'data: {"choices":[{"delta":{},"finish_reason":"stop"}]}\n\ndata: [DONE]\n\n'])
      : sseReply(['data: {"choices":[{"delta":{"content":"追问的回答"},"finish_reason":"stop"}]}\n\ndata: [DONE]\n\n']);
  } });
  try {
    configure(page);
    page.document.getElementById('load-example').click(); submit(page.document);
    page.document.getElementById('run-model').click();
    await waitFor(() => page.document.querySelector('#conversation .md table') && !page.document.querySelector('#conversation .streaming'), '表格渲染');
    const entry = page.document.querySelector('#conversation .conversation-entry');
    assert.equal(entry.querySelector('td:nth-child(2)').textContent, '提交初稿');
    assert.equal(entry.querySelector('.entry-state').hidden, true);
    assert.equal(page.document.getElementById('model-waiting').hidden, true);
    const [rawToggle, copy] = entry.querySelectorAll('.entry-actions button');
    rawToggle.click();
    assert.match(entry.querySelector('pre').textContent, /\| T1 \| 提交初稿 \|/);
    assert.equal(entry.querySelector('table'), null);
    rawToggle.click();
    assert.ok(entry.querySelector('table'));
    copy.click(); await settled();
    assert.equal(copied.at(-1), '| 编号 | 任务 |\n| --- | --- |\n| T1 | 提交初稿 |');
    assert.equal(page.document.querySelector('#recipe-steps [data-step="use"]').dataset.state, 'done');
    assert.equal(page.document.getElementById('export-conversation').hidden, false);
    input(page.document, 'followup', '继续说明');
    page.document.getElementById('followup-form').dispatchEvent(new page.window.Event('submit', { cancelable: true }));
    await waitFor(() => /追问的回答/.test(page.document.getElementById('conversation').textContent) && !page.document.querySelector('#conversation .streaming'), '追问回答');
    const body = JSON.parse(page.requests[1][1].body);
    assert.deepEqual(body.messages.map(message => message.role), ['user', 'assistant', 'user']);
    assert.equal(body.messages[1].content, '| 编号 | 任务 |\n| --- | --- |\n| T1 | 提交初稿 |');
    assert.equal(page.document.querySelectorAll('#conversation .conversation-entry').length, 3);
    assert.equal(page.document.querySelectorAll('#conversation .from-user').length, 1);
    assert.deepEqual(page.errors, []);
  } finally { page.dom.window.close(); }
});
test('流式中途停止保留已收到部分并标为未完成，不带入后续追问', async () => {
  let calls = 0;
  const page = setup({ fetch: async (_url, options) => {
    calls++;
    if (calls === 2) return sseReply(['data: {"choices":[{"delta":{"content":"半截"}}]}\n\n'], { hold: true, signal: options.signal });
    return sseReply(['data: {"choices":[{"delta":{"content":"完整回答' + calls + '"},"finish_reason":"stop"}]}\n\ndata: [DONE]\n\n']);
  } });
  try {
    configure(page);
    page.document.getElementById('load-example').click(); submit(page.document);
    page.document.getElementById('run-model').click();
    await waitFor(() => /完整回答1/.test(page.document.getElementById('conversation').textContent) && !page.document.querySelector('#conversation .streaming'), '第一轮');
    input(page.document, 'followup', '请补充');
    page.document.getElementById('followup-form').dispatchEvent(new page.window.Event('submit', { cancelable: true }));
    await waitFor(() => page.document.querySelector('#conversation .streaming'), '开始接收');
    page.document.getElementById('stop-model').click();
    await waitFor(() => !page.document.getElementById('model-run-error').hidden, '停止提示');
    const partial = page.document.querySelector('#conversation .incomplete');
    assert.ok(partial);
    assert.match(partial.textContent, /半截/);
    assert.equal(partial.querySelector('.entry-state').textContent, '未完成');
    assert.match(page.document.getElementById('model-run-error').textContent, /已停止等待.*未完成/);
    assert.equal(page.document.getElementById('followup').value, '请补充');
    assert.equal(page.document.getElementById('run-model').disabled, false);
    input(page.document, 'followup', '再试一次');
    page.document.getElementById('followup-form').dispatchEvent(new page.window.Event('submit', { cancelable: true }));
    await waitFor(() => /完整回答3/.test(page.document.getElementById('conversation').textContent), '第三轮');
    const body = JSON.parse(page.requests[2][1].body);
    assert.deepEqual(body.messages.map(message => message.content), [page.document.getElementById('output').value, '完整回答1', '再试一次']);
  } finally { page.dom.window.close(); }
});
test('关闭边生成边显示后按普通请求运行，生图不显示该开关', async () => {
  const page = setup({ fetch: async () => jsonReply({ choices: [{ message: { content: '普通回答' } }] }) });
  try {
    configure(page);
    page.document.getElementById('load-example').click(); submit(page.document);
    assert.equal(page.document.getElementById('stream-toggle').hidden, false);
    page.document.getElementById('stream-answer').checked = false;
    page.document.getElementById('run-model').click(); await settled();
    assert.equal(JSON.parse(page.requests[0][1].body).stream, false);
    assert.match(page.document.getElementById('conversation').textContent, /普通回答/);
    page.document.querySelector('[data-recipe="image"]').click();
    page.document.getElementById('load-example').click(); submit(page.document);
    assert.equal(page.document.getElementById('stream-toggle').hidden, true);
    assert.equal(page.document.getElementById('open-model-settings').dataset.connected, 'true');
  } finally { page.dom.window.close(); }
});
test('使用进度随填写、整理和复制推进，修改材料后回到整理步骤', async () => {
  const page = setup();
  try {
    const states = () => [...page.document.querySelectorAll('#recipe-steps li')].map(item => item.dataset.state);
    assert.deepEqual(states(), ['current', 'todo', 'todo']);
    page.document.getElementById('load-example').click();
    assert.deepEqual(states(), ['done', 'current', 'todo']);
    submit(page.document);
    assert.deepEqual(states(), ['done', 'done', 'current']);
    assert.match(page.document.getElementById('output-meta').textContent, /共 \d+ 字/);
    page.document.getElementById('copy').click(); await settled();
    assert.deepEqual(states(), ['done', 'done', 'done']);
    assert.equal(page.document.getElementById('copy').textContent, '已复制 ✓');
    input(page.document, 'input-material', '换一份材料');
    assert.deepEqual(states(), ['done', 'current', 'todo']);
    assert.equal(page.document.querySelector('#recipe-steps [aria-current="step"]').dataset.step, 'compile');
  } finally { page.dom.window.close(); }
});
test('工作配方新增代码审查、周报和来信回复，方法页按表格与代码块排版', () => {
  const page = setup({ hash: '#review' });
  try {
    page.document.querySelector('[data-category="work"]').click();
    assert.deepEqual([...page.document.querySelectorAll('[data-recipe]')].map(item => item.dataset.recipe), ['meeting', 'requirements', 'triage', 'review', 'weekly', 'reply']);
    assert.equal(page.document.getElementById('recipe-title').textContent, '代码改动审查');
    page.document.getElementById('guide-view').click();
    assert.ok(page.document.querySelector('#example table'));
    assert.ok(page.document.querySelector('#example .md-code pre'));
    assert.ok(page.document.querySelector('#repair .md-code-copy'));
    assert.equal(page.document.querySelectorAll('#guide script, #guide img').length, 0);
    assert.equal(page.document.getElementById('workspace').dataset.tone, 'work');
    page.document.querySelector('[data-category="all"]').click();
    page.document.querySelector('[data-recipe="paper"]').click();
    assert.equal(page.document.getElementById('workspace').dataset.tone, 'university');
    assert.equal(page.document.getElementById('hero-count').textContent, '11');
    assert.deepEqual(page.errors, []);
  } finally { page.dom.window.close(); }
});
test('导出对话为本地 Markdown 文件，包含提示词与回答', async () => {
  const page = setup({ fetch: async () => jsonReply({ choices: [{ message: { content: '导出用回答' } }] }) });
  try {
    let exported;
    page.window.URL.createObjectURL = blob => { exported = blob; return 'blob:test'; };
    page.window.URL.revokeObjectURL = () => {};
    configure(page);
    page.document.getElementById('load-example').click(); submit(page.document);
    page.document.getElementById('run-model').click(); await settled();
    assert.equal(page.document.getElementById('export-conversation').hidden, false);
    page.document.getElementById('export-conversation').click();
    const text = await new Promise(resolve => { const reader = new page.window.FileReader(); reader.onload = () => resolve(reader.result); reader.readAsText(exported); });
    assert.match(text, /## 提示词/);
    assert.match(text, /## 模型回答\n\n导出用回答/);
    assert.ok(text.includes(page.document.getElementById('output').value));
    assert.equal(page.requests.length, 1);
  } finally { page.dom.window.close(); }
});
test('按 / 聚焦搜索，Ctrl+Enter 整理当前配方', () => {
  const page = setup();
  try {
    page.document.body.dispatchEvent(new page.window.KeyboardEvent('keydown', { key: '/', bubbles: true, cancelable: true }));
    assert.equal(page.document.activeElement.id, 'search');
    page.document.getElementById('load-example').click();
    page.document.getElementById('input-material').dispatchEvent(new page.window.KeyboardEvent('keydown', { key: 'Enter', ctrlKey: true, bubbles: true, cancelable: true }));
    assert.equal(page.document.getElementById('result').hidden, false);
  } finally { page.dom.window.close(); }
});
