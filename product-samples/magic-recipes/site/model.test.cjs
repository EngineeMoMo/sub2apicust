const test = require('node:test');
const assert = require('node:assert/strict');
const { createServer } = require('node:http');
const client = require('./model.cjs');
const connection = (values = {}) => client.configure({ kind: 'text', protocol: 'chat', base: 'https://example.test/v1/', key: 'fake-test-key', model: 'test-model', ...values });
test('保留自定义路径、完整端点，根域名补 v1', () => {
  assert.equal(connection().endpoint, 'https://example.test/v1/chat/completions');
  assert.equal(connection({ base: 'https://example.test/openai/v1/responses', protocol: 'responses' }).endpoint, 'https://example.test/openai/v1/responses');
  assert.equal(connection({ base: 'https://example.test' }).endpoint, 'https://example.test/v1/chat/completions');
  assert.equal(connection({ base: 'http://127.0.0.1:11434/v1', key: '' }).key, '');
});
test('拒绝无效地址、隐含凭证、端点冲突和不安全远程协议', () => {
  for (const base of ['javascript:alert(1)', 'http://example.test/v1', 'https://user:pass@example.test/v1', 'https://example.test/v1?key=secret', 'https://example.test/v1#secret', 'https://example.test/v1/responses']) assert.throws(() => connection({ base }));
  assert.throws(() => connection({ key: '' }), /API Key/);
  assert.throws(() => connection({ key: 'fake\nheader' }), /API Key/);
  assert.throws(() => connection({ model: '' }), /模型名/);
});
test('文字协议携带对话，密钥仅在请求头，禁止重定向与 Cookie', () => {
  const config = connection();
  const plan = client.request(config, '追问', [{ role: 'user', content: '配方' }, { role: 'assistant', content: '回答' }]);
  assert.deepEqual(JSON.parse(plan.options.body).messages.map(item => item.role), ['user', 'assistant', 'user']);
  assert.equal(plan.options.headers.Authorization, 'Bearer fake-test-key');
  assert.ok(!plan.url.includes(config.key));
  assert.ok(!plan.options.body.includes(config.key));
  assert.equal(plan.options.redirect, 'error');
  assert.equal(plan.options.credentials, 'omit');
  const responses = client.request(connection({ protocol: 'responses' }), '配方');
  assert.equal(JSON.parse(responses.options.body).store, false);
  assert.equal(JSON.parse(responses.options.body).input[0].content, '配方');
});
test('文字解析兼容标准回答、拒答与截断提示', () => {
  assert.equal(client.answer(connection(), { choices: [{ message: { content: '回答' } }] }).text, '回答');
  assert.equal(client.answer(connection(), { choices: [{ message: { refusal: '无法处理' } }] }).text, '无法处理');
  const result = client.answer(connection({ protocol: 'responses' }), { status: 'incomplete', output: [{ type: 'reasoning' }, { type: 'message', content: [{ type: 'output_text', text: '片段' }] }] });
  assert.equal(result.text, '片段');
  assert.match(result.warning, /未完成/);
  assert.throws(() => client.answer(connection(), { choices: [] }), /没有返回文字/);
});
test('生图只请求一张，支持 base64 和 HTTPS 链接，拒绝可执行链接', () => {
  const config = connection({ kind: 'image', size: '1024x1024' });
  const body = JSON.parse(client.request(config, '图像请求').options.body);
  assert.deepEqual(body, { model: 'test-model', prompt: '图像请求', n: 1, size: '1024x1024' });
  assert.equal(client.answer(config, { data: [{ b64_json: 'aGVsbG8=' }] }).images[0], 'data:image/png;base64,aGVsbG8=');
  assert.equal(client.answer(config, { data: [{ url: 'https://example.test/image.png?signature=example' }] }).images[0], 'https://example.test/image.png?signature=example');
  for (const url of ['javascript:alert(1)', 'data:image/svg+xml;base64,PHN2Zz4=', 'https://user:pass@example.test/image.png']) assert.throws(() => client.answer(config, { data: [{ url }] }));
  assert.throws(() => connection({ kind: 'image', size: '16:9' }), /尺寸/);
});
test('错误响应不显示服务返回的密钥或 HTML，不自动重试', async () => {
  let requests = 0;
  await assert.rejects(client.execute(connection(), '配方', [], undefined, async () => {
    requests++;
    return { ok: false, status: 401, json: () => { throw new Error('密钥不应读取'); } };
  }), /API Key 无效/);
  assert.equal(requests, 1);
  await assert.rejects(client.execute(connection(), '配方', [], undefined, async () => { throw new TypeError('Failed to fetch'); }), /CORS/);
  await assert.rejects(client.execute(connection(), '配方', [], undefined, async () => ({ ok: true, json: async () => { throw new Error('not json'); } })), /JSON/);
});
function eventStream(chunks, options = {}) {
  const encoder = new TextEncoder();
  let index = 0;
  const body = new ReadableStream({
    pull(controller) {
      if (index < chunks.length) controller.enqueue(encoder.encode(chunks[index++]));
      else if (options.breakAfter) controller.error(new TypeError('network'));
      else controller.close();
    }
  });
  return new Response(body, { status: 200, headers: { 'Content-Type': options.type ?? 'text/event-stream; charset=utf-8' } });
}
test('默认不流式；流式只用于文字，请求体明确 stream', () => {
  assert.equal(JSON.parse(client.request(connection(), '配方').options.body).stream, false);
  assert.equal(JSON.parse(client.request(connection(), '配方', [], true).options.body).stream, true);
  assert.equal(JSON.parse(client.request(connection({ protocol: 'responses' }), '配方', [], true).options.body).stream, true);
  assert.equal(JSON.parse(client.request(connection({ kind: 'image' }), '图像', [], true).options.body).stream, undefined);
});
test('Chat 流式逐段回调，跨块拆开的事件与 CRLF 都能拼接', async () => {
  const pieces = [];
  const chunks = ['data: {"choices":[{"delta":{"role":"assistant"}}]}\r', '\n\r\ndata: {"choices":[{"delta":{"content":"第一"}}]}\n', '\ndata: {"choices":[{"del', 'ta":{"content":"段"}}]}\n\n: keepalive\n\ndata: {"choices":[{"delta":{},"finish_reason":"stop"}]}\n\ndata: [DONE]\n\n'];
  const result = await client.execute(connection(), '配方', [], undefined, async (_url, options) => {
    assert.equal(JSON.parse(options.body).stream, true);
    return eventStream(chunks);
  }, { onDelta: (piece, text) => pieces.push([piece, text]) });
  assert.equal(result.text, '第一段');
  assert.equal(result.warning, '');
  assert.equal(result.streamed, true);
  assert.deepEqual(pieces, [['第一', '第一'], ['段', '第一段']]);
  assert.deepEqual(result.messages.map(message => message.role), ['user']);
});
test('Responses 流式解析增量、截断与失败事件', async () => {
  const config = connection({ protocol: 'responses' });
  const ok = await client.execute(config, '配方', [], undefined, async () => eventStream([
    'event: response.output_text.delta\ndata: {"type":"response.output_text.delta","delta":"你好"}\n\n',
    'event: response.incomplete\ndata: {"type":"response.incomplete","response":{"status":"incomplete"}}\n\n'
  ]), { onDelta: () => {} });
  assert.equal(ok.text, '你好');
  assert.match(ok.warning, /未完成/);
  const completed = await client.execute(config, '配方', [], undefined, async () => eventStream([
    'data: {"type":"response.completed","response":{"status":"completed","output":[{"type":"message","content":[{"type":"output_text","text":"完整回答"}]}]}}\n\n'
  ]), { onDelta: () => {} });
  assert.equal(completed.text, '完整回答');
  await assert.rejects(client.execute(config, '配方', [], undefined, async () => eventStream([
    'data: {"type":"response.output_text.delta","delta":"半"}\n\n', 'data: {"type":"response.failed","response":{"error":{"message":"server secret"}}}\n\n'
  ]), { onDelta: () => {} }), error => /生成中返回错误/.test(error.message) && error.partial === '半' && !error.message.includes('secret'));
});
test('请求流式但服务返回普通 JSON 或未标注类型时仍能解析', async () => {
  const json = await client.execute(connection(), '配方', [], undefined, async () => new Response(JSON.stringify({ choices: [{ message: { content: '普通回答' } }] }), { headers: { 'Content-Type': 'application/json' } }), { onDelta: () => {} });
  assert.equal(json.text, '普通回答');
  assert.equal(json.streamed, undefined);
  const unlabeled = await client.execute(connection(), '配方', [], undefined, async () => eventStream(['data: {"choices":[{"delta":{"content":"无类型"},"finish_reason":"stop"}]}\n\ndata: [DONE]\n\n'], { type: 'text/plain' }), { onDelta: () => {} });
  assert.equal(unlabeled.text, '无类型');
});
test('流式中断保留已收到的文字，空流和中途错误给出可读原因', async () => {
  await assert.rejects(client.execute(connection(), '配方', [], undefined, async () => eventStream(['data: {"choices":[{"delta":{"content":"已收到"}}]}\n\n'], { breakAfter: true }), { onDelta: () => {} }),
    error => /连接在回答过程中断开/.test(error.message) && error.partial === '已收到');
  await assert.rejects(client.execute(connection(), '配方', [], undefined, async () => eventStream(['data: [DONE]\n\n']), { onDelta: () => {} }), /没有返回文字/);
  await assert.rejects(client.execute(connection(), '配方', [], undefined, async () => eventStream(['data: {"error":{"message":"quota sk-secret"}}\n\n']), { onDelta: () => {} }),
    error => /生成中返回错误/.test(error.message) && !error.message.includes('sk-secret'));
  const unfinished = await client.execute(connection(), '配方', [], undefined, async () => eventStream(['data: {"choices":[{"delta":{"content":"没有结束标记"}}]}\n\n']), { onDelta: () => {} });
  assert.match(unfinished.warning, /可能不完整/);
});
test('停止等待会中止读取流，不再回调后续文字', async () => {
  const controller = new AbortController();
  const pieces = [];
  let pulls = 0;
  const body = new ReadableStream({
    pull(stream) {
      pulls++;
      if (pulls === 1) stream.enqueue(new TextEncoder().encode('data: {"choices":[{"delta":{"content":"第一段"}}]}\n\n'));
      else return new Promise(() => {});
    }
  });
  const running = client.execute(connection(), '配方', [], controller.signal, async () => new Response(body, { headers: { 'Content-Type': 'text/event-stream' } }), { onDelta: piece => { pieces.push(piece); controller.abort(); } });
  await assert.rejects(running, error => error.name === 'AbortError');
  assert.deepEqual(pieces, ['第一段']);
});
test('真实本地 HTTP 流式往返', async () => {
  const server = createServer(async (request, response) => {
    let body = '';
    for await (const chunk of request) body += chunk;
    assert.equal(JSON.parse(body).stream, true);
    response.writeHead(200, { 'Content-Type': 'text/event-stream' });
    response.write('data: {"choices":[{"delta":{"content":"本地"}}]}\n\n');
    setTimeout(() => { response.end('data: {"choices":[{"delta":{"content":"流式"},"finish_reason":"stop"}]}\n\ndata: [DONE]\n\n'); }, 20);
  });
  await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
  try {
    const pieces = [];
    const result = await client.execute(connection({ base: 'http://127.0.0.1:' + server.address().port + '/v1' }), '虚构材料', [], new AbortController().signal, undefined, { onDelta: piece => pieces.push(piece) });
    assert.equal(result.text, '本地流式');
    assert.deepEqual(pieces, ['本地', '流式']);
  } finally { await new Promise(resolve => server.close(resolve)); }
});
test('真实本地 HTTP 请求经过兼容协议往返，无付费模型调用', async () => {
  const seen = [];
  const server = createServer(async (request, response) => {
    let body = '';
    for await (const chunk of request) body += chunk;
    seen.push({ url: request.url, authorization: request.headers.authorization, body: JSON.parse(body) });
    response.writeHead(200, { 'Content-Type': 'application/json' });
    response.end(JSON.stringify({ choices: [{ message: { content: '模拟接口回答' } }] }));
  });
  await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
  try {
    const config = connection({ base: 'http://127.0.0.1:' + server.address().port + '/v1' });
    const result = await client.execute(config, '虚构材料', [], new AbortController().signal);
    assert.equal(result.text, '模拟接口回答');
    assert.equal(seen[0].url, '/v1/chat/completions');
    assert.equal(seen[0].body.messages[0].content, '虚构材料');
    assert.equal(seen[0].authorization, 'Bearer fake-test-key');
  } finally { await new Promise(resolve => server.close(resolve)); }
});
