import { mkdir, readFile, writeFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import { catalog } from './catalog.mjs';

const directory = path.dirname(fileURLToPath(import.meta.url));
const recipes = [];
for (const entry of catalog) {
  const markdown = (await readFile(path.join(directory, '..', entry.file), 'utf8')).replaceAll('\r\n', '\n');
  const promptMatch = markdown.match(/## 完整提示词\s+~~~text\n([\s\S]*?)\n~~~/);
  if (!promptMatch) throw new Error('缺少提示词：' + entry.file);
  const count = [...promptMatch[1].matchAll(/\{\{[^}]+\}\}/g)].length;
  if (count !== entry.fields.length) throw new Error('字段与模板不匹配：' + entry.file);
  let fieldIndex = 0;
  const prompt = promptMatch[1].replace(/\{\{[^}]+\}\}/g, () => '{{field:' + entry.fields[fieldIndex++].id + '}}');
  const section = heading => {
    const start = markdown.indexOf(heading);
    if (start < 0) throw new Error('缺少章节：' + heading);
    const end = markdown.indexOf('\n## ', start + heading.length);
    return markdown.slice(start + heading.length, end < 0 ? markdown.length : end).trim();
  };
  recipes.push({ ...entry, prompt, version: markdown.match(/版本：([^｜\n]+)/)[1].trim(),
    scope: section('## 用途与边界'), example: section('## 正常示例'), checks: section('## 验收标准'), repair: section('## 修订提示词') });
}
const logo = await readFile(path.join(directory, '../../../frontend/src/custom/assets/mofa-mark-flat.png'));
const style = await readFile(path.join(directory, 'theme.css'), 'utf8');
const core = await readFile(path.join(directory, 'core.cjs'), 'utf8');
const model = await readFile(path.join(directory, 'model.cjs'), 'utf8');
const connector = await readFile(path.join(directory, 'magic-connect.cjs'), 'utf8');
const modelUi = await readFile(path.join(directory, 'model-ui.js'), 'utf8');
const app = await readFile(path.join(directory, 'app.js'), 'utf8');
const host = await readFile(path.join(directory, '../../family-runtime/host-client.js'), 'utf8');
const template = await readFile(path.join(directory, 'template.html'), 'utf8');
const safeData = JSON.stringify(recipes).replaceAll('<', '\\u003c').replaceAll('\u2028', '\\u2028').replaceAll('\u2029', '\\u2029');
let html = template.replace('%%LOGO%%', 'data:image/png;base64,' + logo.toString('base64'))
  .replace('%%STYLE%%', () => style).replace('%%DATA%%', () => safeData)
  .replace('%%CORE%%', () => core).replace('%%MODEL%%', () => model)
  .replace('%%CONNECTOR%%', () => connector)
  .replace('%%MODEL_UI%%', () => host + '\n;' + modelUi).replace('%%APP%%', () => app);
const outputIndex = process.argv.indexOf('--output');
if (outputIndex >= 0 && !process.argv[outputIndex + 1]) throw new Error('缺少输出目录');
const output = outputIndex >= 0 ? path.resolve(process.argv[outputIndex + 1]) : directory;
await mkdir(output, { recursive: true });
if (process.argv.includes('--embedded')) {
  const scripts = [];
  html = html.replace('<head>', '<head>\n<meta name="mofa-api-site" content="same-origin">')
    .replace("script-src 'unsafe-inline'", "script-src 'self'")
    .replace("style-src 'unsafe-inline'", "style-src 'self'")
    .replace('img-src data:', "img-src 'self' data:")
    .replace('connect-src https:', "connect-src 'self' https:")
    .replace(/<style>[\s\S]*?<\/style>/, '<link rel="stylesheet" href="./theme.css">')
    .replace(/<script>([\s\S]*?)<\/script>/g, (_tag, script) => { scripts.push(script); return ''; })
    .replace('</body>', '<script src="./bundle.js"></script>\n</body>');
  await writeFile(path.join(output, 'theme.css'), style, 'utf8');
  await writeFile(path.join(output, 'bundle.js'), scripts.join('\n;\n'), 'utf8');
}
await writeFile(path.join(output, 'index.html'), html, 'utf8');
console.log('已生成独立页面：' + recipes.length + '份配方，' + Buffer.byteLength(html) + '字节');
