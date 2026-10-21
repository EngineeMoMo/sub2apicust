import test from 'node:test';
import assert from 'node:assert/strict';
import { createServer } from 'node:http';
import { imageConnection, imageRequest, imageResults, generateImage, suggestedSize, downloadName, IMAGE_SIZES } from './generate.mjs';

const connection = (values = {}) => imageConnection({ kind: 'image', protocol: 'images', base: 'https://gateway.test/v1', key: 'fake-image-key', model: 'image-model', size: '', ...values });
test('只接受控制台授权的生图连接，补全 Images 端点并拒绝不安全地址', () => {
 assert.equal(connection().endpoint, 'https://gateway.test/v1/images/generations');
 assert.equal(connection({ base: 'https://gateway.test' }).endpoint, 'https://gateway.test/v1/images/generations');
 assert.equal(connection({ base: 'http://127.0.0.1:4181/v1/images/generations', key: '' }).endpoint, 'http://127.0.0.1:4181/v1/images/generations');
 for (const base of ['http://remote.test/v1', 'javascript:alert(1)', 'https://user:pass@gateway.test/v1', 'https://gateway.test/v1?key=x', 'https://gateway.test/v1#x']) assert.throws(() => connection({ base }));
 assert.throws(() => connection({ kind: 'text', protocol: 'chat' }), /不是生图连接/);
 assert.throws(() => connection({ key: 'two words' }), /API Key/);
 assert.throws(() => connection({ key: '' }), /缺少 API Key/);
 assert.throws(() => connection({ model: '' }), /生图模型/);
});
test('请求只生成一张，密钥只在请求头，不带 Cookie、不跟随重定向', () => {
 const plan = imageRequest(connection(), '  一只纸狐狸  ', '1024x1536');
 assert.deepEqual(JSON.parse(plan.options.body), { model: 'image-model', prompt: '一只纸狐狸', n: 1, size: '1024x1536' });
 assert.equal(plan.options.headers.Authorization, 'Bearer fake-image-key');
 assert.ok(!plan.url.includes('fake-image-key') && !plan.options.body.includes('fake-image-key'));
 assert.equal(plan.options.credentials, 'omit');
 assert.equal(plan.options.redirect, 'error');
 assert.equal(JSON.parse(imageRequest(connection(), '提示词').options.body).size, undefined);
 assert.throws(() => imageRequest(connection(), '   '), /提示词为空/);
 assert.throws(() => imageRequest(connection(), '提示词', '999x999'), /尺寸/);
 assert.ok(IMAGE_SIZES.every(option => option.value === '' || /^\d+x\d+$/.test(option.value)));
});
test('解析 Base64 与 HTTPS 图片，拒绝 SVG、脚本地址与凭据链接', () => {
 assert.deepEqual(imageResults({ data: [{ b64_json: 'aGVsbG8gd29ybGQ=' }] }), [{ src: 'data:image/png;base64,aGVsbG8gd29ybGQ=', revised: '' }]);
 assert.equal(imageResults({ output_format: 'webp', data: [{ b64_json: 'aGVsbG8gd29ybGQ=' }] })[0].src, 'data:image/webp;base64,aGVsbG8gd29ybGQ=');
 assert.equal(imageResults({ data: [{ url: 'https://cdn.test/a.png?sig=1', revised_prompt: '改写后的提示词' }] })[0].revised, '改写后的提示词');
 for (const url of ['javascript:alert(1)', 'data:image/svg+xml;base64,PHN2Zz4=', 'https://user:pass@cdn.test/a.png', 'ftp://cdn.test/a.png']) assert.throws(() => imageResults({ data: [{ url }] }));
 assert.throws(() => imageResults({ data: [] }), /没有返回图片/);
 assert.throws(() => imageResults({ error: { message: 'secret' } }), error => !error.message.includes('secret'));
});
test('错误状态给出可操作原因，不读取响应正文，也不重试', async () => {
 let calls = 0;
 await assert.rejects(generateImage(connection(), '提示词', '', undefined, async () => { calls++; return { ok: false, status: 403, json: () => assert.fail('不应读取') }; }), /生图权限/);
 assert.equal(calls, 1);
 await assert.rejects(generateImage(connection(), '提示词', '', undefined, async () => { throw new TypeError('Failed to fetch'); }), /无法连接生图接口/);
 const controller = new AbortController(); controller.abort();
 await assert.rejects(generateImage(connection(), '提示词', '', controller.signal, async () => { throw new DOMException('Aborted', 'AbortError'); }), error => error.name === 'AbortError');
});
test('画幅建议与下载文件名只用安全字符', () => {
 assert.equal(suggestedSize('portrait'), '1024x1536');
 assert.equal(suggestedSize('wide'), '1536x1024');
 assert.equal(suggestedSize('square'), '1024x1024');
 assert.equal(downloadName('jade-product', 0, 'data:image/jpeg;base64,AAAA'), '魔法工坊-jade-product-1.jpg');
 assert.equal(downloadName('../evil', 1, 'data:image/png;base64,AAAA'), '魔法工坊-evil-2.png');
 assert.equal(downloadName('jade-product', 0, 'https://cdn.test/a.png'), '');
});
test('真实本地 HTTP 生图往返，无付费调用', async () => {
 const seen = [];
 const server = createServer(async (request, response) => {
  let body = '';
  for await (const chunk of request) body += chunk;
  seen.push({ url: request.url, authorization: request.headers.authorization, body: JSON.parse(body) });
  response.writeHead(200, { 'Content-Type': 'application/json' });
  response.end(JSON.stringify({ data: [{ b64_json: 'aGVsbG8gd29ybGQ=' }] }));
 });
 await new Promise(resolve => server.listen(0, '127.0.0.1', resolve));
 try {
  const result = await generateImage(connection({ base: 'http://127.0.0.1:' + server.address().port + '/v1' }), '虚构提示词', '1024x1024', new AbortController().signal);
  assert.equal(result.length, 1);
  assert.equal(seen[0].url, '/v1/images/generations');
  assert.equal(seen[0].authorization, 'Bearer fake-image-key');
  assert.equal(seen[0].body.size, '1024x1024');
 } finally { await new Promise(resolve => server.close(resolve)); }
});
