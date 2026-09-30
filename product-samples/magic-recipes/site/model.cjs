(function (root, factory) {
  const api = factory();
  if (typeof module === 'object' && module.exports) module.exports = api;
  else root.MofaRecipeModel = api;
})(globalThis, () => {
  'use strict';
  const paths = { chat: '/chat/completions', responses: '/responses', images: '/images/generations' };
  function fail(message, field) {
    const error = new Error(message);
    error.field = field;
    throw error;
  }
  function local(hostname) { return ['localhost', '[::1]', '127.0.0.1'].includes(hostname); }
  function safeUrl(value, field = 'base') {
    let url;
    try { url = new URL(value); } catch { fail('请填写完整的接口地址，例如 https://你的域名/v1。', field); }
    if (url.protocol !== 'https:' && !(url.protocol === 'http:' && local(url.hostname))) fail('远程接口请使用 HTTPS；本机 localhost 或 127.0.0.1 可使用 HTTP。', field);
    if (url.username || url.password || url.search || url.hash) fail('地址不能包含账号、密码、查询参数或井号片段。', field);
    return url;
  }
  function configure(values) {
    const config = {
      kind: values.kind,
      protocol: values.kind === 'image' ? 'images' : values.protocol,
      base: String(values.base || '').trim(),
      key: String(values.key || '').trim(),
      model: String(values.model || '').trim(),
      size: String(values.size || '').trim()
    };
    if (!['text', 'image'].includes(config.kind) || !paths[config.protocol] || (config.kind === 'text' && config.protocol === 'images')) fail('请选择正确的接口格式。', 'protocol');
    if (!config.model || config.model.length > 200) fail('请填写服务商实际支持的模型名（最多 200 字）。', 'model');
    if (config.key.length > 2048 || /[\r\n]/.test(config.key)) fail('API Key 格式有误，请重新粘贴。', 'key');
    const url = safeUrl(config.base);
    if (!config.key && !local(url.hostname)) fail('请填写这个接口的 API Key；免鉴权本机服务可留空。', 'key');
    let pathname = url.pathname.replace(/\/+$/, '');
    const existing = Object.values(paths).find(path => pathname.endsWith(path));
    if (existing && existing !== paths[config.protocol]) fail('接口地址与所选格式不一致，请填写根地址或对应端点。', 'base');
    if (!existing) pathname = (pathname || '/v1') + paths[config.protocol];
    url.pathname = pathname;
    config.endpoint = url.href;
    if (config.kind === 'image' && config.size && !/^(auto|\d{2,5}x\d{2,5})$/.test(config.size)) fail('尺寸请填写 auto 或宽x高，例如 1024x1024；实际支持值以服务商为准。', 'size');
    return config;
  }
  function request(config, prompt, history = []) {
    if (typeof prompt !== 'string' || !prompt.trim()) fail('请先整理提示词或填写追问。');
    const messages = [...history, { role: 'user', content: prompt }];
    if (messages.reduce((length, message) => length + message.content.length, 0) > 200000) fail('当前对话过长，请重新整理提示词开始新对话。');
    const body = config.kind === 'image'
      ? { model: config.model, prompt, n: 1, ...(config.size ? { size: config.size } : {}) }
      : config.protocol === 'responses'
        ? { model: config.model, input: messages, stream: false, store: false }
        : { model: config.model, messages, stream: false };
    const headers = { 'Content-Type': 'application/json' };
    if (config.key) headers.Authorization = 'Bearer ' + config.key;
    return { url: config.endpoint, options: { method: 'POST', headers, body: JSON.stringify(body), credentials: 'omit', redirect: 'error', referrerPolicy: 'no-referrer', cache: 'no-store' }, messages };
  }
  function imageSource(item, format) {
    if (typeof item.b64_json === 'string' && /^[A-Za-z0-9+/=\r\n]+$/.test(item.b64_json)) {
      const mime = { png: 'image/png', jpeg: 'image/jpeg', jpg: 'image/jpeg', webp: 'image/webp' }[format] || 'image/png';
      return 'data:' + mime + ';base64,' + item.b64_json.replace(/[\r\n]/g, '');
    }
    if (typeof item.url === 'string') {
      if (/^data:image\/(png|jpeg|webp);base64,[A-Za-z0-9+/=]+$/.test(item.url)) return item.url;
      const url = new URL(item.url);
      if (url.username || url.password || (url.protocol !== 'https:' && !(url.protocol === 'http:' && local(url.hostname)))) fail('服务返回的图片地址格式不受支持。');
      return url.href;
    }
    fail('接口没有返回可显示的图片，请检查模型与接口格式。');
  }
  function answer(config, payload) {
    if (!payload || typeof payload !== 'object' || payload.error || ['failed', 'cancelled', 'queued', 'in_progress'].includes(payload.status)) fail('接口没有返回完成的结果，请检查服务状态与接口格式。');
    if (config.kind === 'image') {
      if (!Array.isArray(payload.data) || !payload.data.length) fail('接口没有返回图片，请检查模型是否支持 Images API。');
      try { return { images: payload.data.map(item => imageSource(item, payload.output_format)) }; }
      catch { fail('接口返回的图片格式不受支持，请检查服务商响应。'); }
    }
    let content;
    let warning = '';
    if (config.protocol === 'responses') {
      content = (Array.isArray(payload.output) ? payload.output : []).filter(item => item?.type === 'message').flatMap(item => Array.isArray(item.content) ? item.content : []).map(item => item?.type === 'output_text' ? item.text : item?.type === 'refusal' ? item.refusal : '').filter(item => typeof item === 'string').join('\n');
      if (payload.status === 'incomplete') warning = '本次回答未完成，请检查服务端输出限制后再追问。';
    } else {
      const choice = payload.choices?.[0];
      content = choice?.message?.content || choice?.message?.refusal;
      if (Array.isArray(content)) content = content.filter(item => item.type === 'text').map(item => item.text).join('\n');
      if (choice?.finish_reason === 'length') warning = '本次回答被输出限制截断，可以继续追问。';
    }
    if (typeof content !== 'string' || !content.trim()) fail('接口没有返回文字回答，请检查模型与接口格式。');
    return { text: content, warning };
  }
  async function execute(config, prompt, history, signal, transport = globalThis.fetch) {
    const plan = request(config, prompt, history);
    let response;
    try { response = await transport(plan.url, { ...plan.options, signal }); }
    catch (error) {
      if (signal?.aborted || error.name === 'AbortError') throw error;
      fail('无法连接接口。请检查地址、网络与浏览器跨域许可（CORS）；本页不会自动改用其他服务。');
    }
    if (!response.ok) {
      const messages = { 400: '请求被拒绝，请检查模型名、接口格式或图片尺寸。', 401: 'API Key 无效或已过期，请在模型设置中更换。', 403: '当前密钥无权使用该模型，请检查分组与模型权限。', 404: '接口或模型不存在，请检查地址与接口格式。', 429: '服务限流或额度不足，请检查额度后稍后重试。' };
      fail(messages[response.status] || '模型服务请求失败（HTTP ' + response.status + '），请检查服务状态。');
    }
    let payload;
    try { payload = await response.json(); } catch { fail('接口没有返回 JSON 结果，请检查是否选择了正确的兼容接口。'); }
    return { ...answer(config, payload), messages: plan.messages };
  }
  return { configure, request, answer, execute };
});
