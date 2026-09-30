import { createServer } from 'node:http';
import { readFile } from 'node:fs/promises';
const html = await readFile(new URL('./index.html', import.meta.url));
const server = createServer((request, response) => {
  if (!['GET', 'HEAD'].includes(request.method)) { response.writeHead(405); response.end(); return; }
  if (!['/', '/index.html'].includes(request.url.split('?')[0])) { response.writeHead(404); response.end(); return; }
  response.writeHead(200, { 'Content-Type': 'text/html; charset=utf-8', 'Cache-Control': 'no-store', 'X-Content-Type-Options': 'nosniff' });
  response.end(request.method === 'HEAD' ? undefined : html);
});
server.on('error', error => { console.error(error.message); process.exitCode = 1; });
const host = process.env.HOST || '127.0.0.1';
const port = Number(process.env.PORT || 4178);
if (!Number.isInteger(port) || port < 1 || port > 65535) throw new Error('PORT 需要是有效端口');
server.listen(port, host, () => console.log('魔法配方本地预览：http://' + host + ':' + port));
