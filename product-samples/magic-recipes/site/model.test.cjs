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
