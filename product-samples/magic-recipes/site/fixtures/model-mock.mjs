import { createServer } from 'node:http';
import { readFile } from 'node:fs/promises';

const image = await readFile(new URL('../../../../frontend/src/custom/assets/mofa-mark-flat.png', import.meta.url));
const server = createServer(async (request, response) => {
  response.setHeader('Access-Control-Allow-Origin', 'http://127.0.0.1:4178');
  response.setHeader('Access-Control-Allow-Methods', 'POST, OPTIONS');
  response.setHeader('Access-Control-Allow-Headers', 'Authorization, Content-Type');
  if (request.method === 'OPTIONS') { response.writeHead(204); response.end(); return; }
  if (request.method !== 'POST') { response.writeHead(405); response.end(); return; }
  let raw = '';
  for await (const chunk of request) raw += chunk;
  try {
    const body = JSON.parse(raw);
    const messages = body.messages || body.input || [];
    const text = messages.length > 1
      ? '本地模拟追问回答：已经携带前一轮对话。这是接口测试结果，没有调用真实模型。'
      : '本地模拟模型回答：提示词已收到。你可以在下方继续追问，以检查多轮对话是否保留。此结果不代表真实模型效果。';
    const payload = request.url === '/v1/chat/completions'
      ? { choices: [{ message: { role: 'assistant', content: text } }] }
      : request.url === '/v1/responses'
        ? { status: 'completed', output: [{ type: 'message', role: 'assistant', content: [{ type: 'output_text', text }] }] }
        : request.url === '/v1/images/generations'
          ? { data: [{ b64_json: image.toString('base64') }] }
          : undefined;
    response.writeHead(payload ? 200 : 404, { 'Content-Type': 'application/json' });
    response.end(JSON.stringify(payload || { error: '模拟接口不存在' }));
  } catch { response.writeHead(400); response.end(); }
});
server.on('error', error => { console.error(error.message); process.exitCode = 1; });
server.listen(4180, '127.0.0.1', () => console.log('仅本机模拟模型接口：http://127.0.0.1:4180/v1；不调用真实模型，生图返回既有 Logo。'));
