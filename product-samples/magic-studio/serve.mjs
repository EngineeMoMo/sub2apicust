import { createServer } from 'node:http';
import { readFile } from 'node:fs/promises';
import { extname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
const files=new Set(['index.html','theme.css','app.mjs','core.mjs','icons.mjs','catalog.mjs','new-gallery.mjs','host-client.js']);
const types={'.html':'text/html; charset=utf-8','.css':'text/css; charset=utf-8','.js':'text/javascript; charset=utf-8','.mp4':'video/mp4','.mjs':'text/javascript; charset=utf-8','.png':'image/png','.webp':'image/webp'};
export function createAppServer(){return createServer(async(request,response)=>{
 if(!['GET','HEAD'].includes(request.method)){response.writeHead(405,{Allow:'GET, HEAD'});response.end();return;}
 try{const pathname=new URL(request.url,'http://localhost').pathname,path=pathname==='/'?'index.html':pathname.slice(1);if(!files.has(path)&&!/^assets\/[a-z0-9-]+\.(png|webp|mp4)$/.test(path)){response.writeHead(404);response.end();return;}
 const data=await readFile(new URL(path==='host-client.js'?'../family-runtime/host-client.js':path,import.meta.url));response.writeHead(200,{'Content-Type':types[extname(path)]||'application/octet-stream','Cache-Control':'no-store','X-Content-Type-Options':'nosniff','Referrer-Policy':'no-referrer','Content-Security-Policy':"default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self'; connect-src 'none'; object-src 'none'; base-uri 'none'; form-action 'none'; media-src 'self'; frame-ancestors 'self'"});response.end(request.method==='HEAD'?undefined:data);
 }catch{response.writeHead(404);response.end();}
 });}
if(process.argv[1]&&fileURLToPath(import.meta.url)===resolve(process.argv[1])){const server=createAppServer();server.on('error',error=>{console.error(error.message);process.exitCode=1;});server.listen(4179,'127.0.0.1',()=>console.log('魔法工坊本地版本：http://127.0.0.1:4179'));}
