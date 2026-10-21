import { createServer } from 'node:http';
import { readFile } from 'node:fs/promises';

const image = await readFile(new URL('../../../../frontend/src/custom/assets/mofa-mark-flat.png', import.meta.url));
const wait = ms => new Promise(resolve => setTimeout(resolve, ms));
// 回答里带表格、列表和代码块，用来检查页面排版；内容是固定的模拟文字，不代表真实模型效果。
const first = [
  '## 本地模拟回答\n\n',
  '提示词已收到。以下是**排版检查用**的固定内容，不是真实模型输出。\n\n',
  '| 编号 | 任务 | 负责人 | 来源 |\n| --- | --- | --- | --- |\n',
  '| T1 | 提交首页文案初稿 | 小林 | L1 |\n',
  '| T2 | 复核价格页 | 待确认 | L2 |\n\n',
  '待确认问题：\n\n1. T2 的期限\n2. 复核范围\n\n',
  '> 可以在下方继续追问，检查多轮对话是否保留。\n\n',
  '```text\n修订示例：仅修改有问题的条目\n```'
];
const followup = ['本地模拟追问回答：', '已经携带前一轮对话。', '这是接口测试结果，没有调用真实模型。'];
const origins = new Set(['http://127.0.0.1:4178', 'http://localhost:4178']);
const server = createServer(async (request, response) => {
  if (origins.has(request.headers.origin)) response.setHeader('Access-Control-Allow-Origin', request.headers.origin);
  response.setHeader('Access-Control-Allow-Methods', 'POST, OPTIONS');
  response.setHeader('Access-Control-Allow-Headers', 'Authorization, Content-Type');
  if (request.method === 'OPTIONS') { response.writeHead(204); response.end(); return; }
  if (request.method !== 'POST') { response.writeHead(405); response.end(); return; }
  let raw = '';
  for await (const chunk of request) raw += chunk;
  try {
    const body = JSON.parse(raw);
    const messages = body.messages || body.input || [];
    const pieces = messages.length > 1 ? followup : first;
    const text = pieces.join('');
    if (body.stream && (request.url === '/v1/chat/completions' || request.url === '/v1/responses')) {
      // 请求体读完后 request.destroyed 会变为 true，要看响应连接是否被客户端关闭。
      let closed = false;
      response.on('close', () => { closed = true; });
      response.writeHead(200, { 'Content-Type': 'text/event-stream; charset=utf-8', 'Cache-Control': 'no-cache' });
      for (const piece of pieces) {
        if (closed) return;
        await wait(260);
        const event = request.url === '/v1/responses'
          ? { type: 'response.output_text.delta', delta: piece }
          : { choices: [{ delta: { content: piece } }] };
        response.write('data: ' + JSON.stringify(event) + '\n\n');
      }
      response.end(request.url === '/v1/responses'
        ? 'data: ' + JSON.stringify({ type: 'response.completed', response: { status: 'completed' } }) + '\n\n'
        : 'data: ' + JSON.stringify({ choices: [{ delta: {}, finish_reason: 'stop' }] }) + '\n\ndata: [DONE]\n\n');
      return;
    }
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
