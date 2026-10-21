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
  function request(config, prompt, history = [], stream = false) {
    if (typeof prompt !== 'string' || !prompt.trim()) fail('请先整理提示词或填写追问。');
    const messages = [...history, { role: 'user', content: prompt }];
    if (messages.reduce((length, message) => length + message.content.length, 0) > 200000) fail('当前对话过长，请重新整理提示词开始新对话。');
    const streaming = Boolean(stream) && config.kind !== 'image';
    const body = config.kind === 'image'
      ? { model: config.model, prompt, n: 1, ...(config.size ? { size: config.size } : {}) }
      : config.protocol === 'responses'
        ? { model: config.model, input: messages, stream: streaming, store: false }
        : { model: config.model, messages, stream: streaming };
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
  // 流式回答：兼容 Chat Completions（choices[].delta）与 Responses（response.output_text.delta）事件。
  const MAX_STREAM_TEXT = 400000;
  function streamParser(config, onDelta) {
    const state = { text: '', done: false, warning: '', events: 0 };
    function add(piece) {
      if (typeof piece !== 'string' || !piece) return;
      if (state.text.length + piece.length > MAX_STREAM_TEXT) {
        piece = piece.slice(0, Math.max(0, MAX_STREAM_TEXT - state.text.length));
        state.warning = '回答过长，页面只保留前 40 万字。';
      }
      if (!piece) return;
      state.text += piece;
      onDelta?.(piece, state.text);
    }
    function event(payload) {
      if (!payload || typeof payload !== 'object') return;
      state.events++;
      if (payload.error) fail('模型服务在生成中返回错误，请检查服务状态、额度与模型权限。');
      if (config.protocol === 'responses') {
        const type = String(payload.type || '');
        if (type === 'response.output_text.delta' || type === 'response.refusal.delta') add(payload.delta);
        else if (type === 'response.completed') {
          state.done = true;
          if (!state.text) { try { add(answer(config, payload.response).text); } catch { /* 结束时再判断空回答 */ } }
        } else if (type === 'response.incomplete') { state.done = true; state.warning = '本次回答未完成，请检查服务端输出限制后再追问。'; }
        else if (type === 'response.failed' || type === 'error') fail('模型服务在生成中返回错误，请检查服务状态、额度与模型权限。');
        return;
      }
      const choice = Array.isArray(payload.choices) ? payload.choices[0] : undefined;
      const delta = choice?.delta;
      if (delta) {
        if (typeof delta.content === 'string') add(delta.content);
        else if (Array.isArray(delta.content)) add(delta.content.filter(item => item?.type === 'text').map(item => item.text).join(''));
        if (typeof delta.refusal === 'string') add(delta.refusal);
      }
      if (choice?.finish_reason) {
        state.done = true;
        if (choice.finish_reason === 'length') state.warning = '本次回答被输出限制截断，可以继续追问。';
      }
    }
    function block(raw) {
      const data = raw.split('\n').filter(line => line.startsWith('data:')).map(line => line.slice(5).replace(/^ /, '')).join('\n');
      if (!data) return;
      if (data.trim() === '[DONE]') { state.done = true; return; }
      let payload;
      try { payload = JSON.parse(data); } catch { return; }
      event(payload);
    }
    function finish() {
      if (!state.text.trim()) fail('接口没有返回文字回答，请检查模型与接口格式。');
      if (!state.done && !state.warning) state.warning = '连接在回答结束标记前断开，回答可能不完整。';
      return { text: state.text, warning: state.warning };
    }
    return { state, block, finish };
  }
  function parseEventStream(config, source, onDelta) {
    const parser = streamParser(config, onDelta);
    for (const raw of String(source).replace(/\r\n?/g, '\n').split(/\n\n+/)) parser.block(raw);
    return parser.finish();
  }
  async function readEventStream(config, body, signal, onDelta) {
    const parser = streamParser(config, onDelta);
    const reader = body.getReader();
    const decoder = new TextDecoder();
    let buffer = '';
    const drain = () => {
      let boundary;
      while ((boundary = buffer.indexOf('\n\n')) >= 0) {
        const raw = buffer.slice(0, boundary);
        buffer = buffer.slice(boundary + 2);
        parser.block(raw);
      }
    };
    try {
      for (;;) {
        if (signal?.aborted) { reader.cancel().catch(() => {}); throw new DOMException('Aborted', 'AbortError'); }
        let chunk;
        try { chunk = await reader.read(); }
        catch (error) {
          if (signal?.aborted || error?.name === 'AbortError') throw error;
          const broken = new Error('连接在回答过程中断开，已显示收到的部分。');
          broken.partial = parser.state.text;
          throw broken;
        }
        if (chunk.done) break;
        buffer += decoder.decode(chunk.value, { stream: true });
        // 只在确认不是被拆开的 CRLF 后再统一换行，避免误判事件边界。
        const tail = buffer.endsWith('\r') ? '\r' : '';
        buffer = buffer.slice(0, buffer.length - tail.length).replace(/\r\n?/g, '\n') + tail;
        drain();
      }
      buffer = (buffer + decoder.decode()).replace(/\r\n?/g, '\n');
      drain();
      if (buffer.trim()) parser.block(buffer);
    } catch (error) {
      reader.cancel().catch(() => {});
      if (error instanceof Error && !('partial' in error) && !(signal?.aborted || error.name === 'AbortError')) error.partial = parser.state.text;
      throw error;
    }
    return parser.finish();
  }
  async function execute(config, prompt, history, signal, transport = globalThis.fetch, options = {}) {
    const stream = config.kind !== 'image' && typeof options.onDelta === 'function';
    const plan = request(config, prompt, history, stream);
    let response;
    try { response = await (transport || globalThis.fetch)(plan.url, { ...plan.options, signal }); }
    catch (error) {
      if (signal?.aborted || error.name === 'AbortError') throw error;
      fail('无法连接接口。请检查地址、网络与浏览器跨域许可（CORS）；本页不会自动改用其他服务。');
    }
    if (!response.ok) {
      const messages = { 400: '请求被拒绝，请检查模型名、接口格式或图片尺寸。', 401: 'API Key 无效或已过期，请在模型设置中更换。', 403: '当前密钥无权使用该模型，请检查分组与模型权限。', 404: '接口或模型不存在，请检查地址与接口格式。', 429: '服务限流或额度不足，请检查额度后稍后重试。' };
      fail(messages[response.status] || '模型服务请求失败（HTTP ' + response.status + '），请检查服务状态。');
    }
    const type = String(response.headers?.get?.('content-type') || '');
    if (stream && /text\/event-stream/i.test(type) && typeof response.body?.getReader === 'function') {
      return { ...(await readEventStream(config, response.body, signal, options.onDelta)), messages: plan.messages, streamed: true };
    }
    if (stream && typeof response.text === 'function' && !/json/i.test(type)) {
      // 少数代理不标注事件流类型：读完整正文后再判断是事件流还是普通 JSON。
      let raw;
      try { raw = await response.text(); } catch (error) { if (signal?.aborted || error.name === 'AbortError') throw error; fail('读取接口响应失败，请检查网络后重试。'); }
      if (/^\s*(data|event|id|retry):/m.test(raw)) return { ...parseEventStream(config, raw, options.onDelta), messages: plan.messages, streamed: true };
      let payload;
      try { payload = JSON.parse(raw); } catch { fail('接口没有返回 JSON 结果，请检查是否选择了正确的兼容接口。'); }
      return { ...answer(config, payload), messages: plan.messages };
    }
    let payload;
    try { payload = await response.json(); } catch { fail('接口没有返回 JSON 结果，请检查是否选择了正确的兼容接口。'); }
    return { ...answer(config, payload), messages: plan.messages };
  }
  return { configure, request, answer, execute, parseEventStream };
});
