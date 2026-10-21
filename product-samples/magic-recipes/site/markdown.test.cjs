const test = require('node:test');
const assert = require('node:assert/strict');
const path = require('node:path');
const { createRequire } = require('node:module');
const markdown = require('./markdown.cjs');
const frontendRequire = createRequire(path.resolve(__dirname, '../../../frontend/package.json'));
const { JSDOM } = frontendRequire('jsdom');
function rendered(source, options) {
  const dom = new JSDOM('<!doctype html><div id="root"></div>');
  const root = dom.window.document.getElementById('root');
  root.append(markdown.render(source, dom.window.document, options));
  return { root, window: dom.window };
}
test('常见回答结构：标题、段落换行、列表、引用与分隔线', () => {
  const blocks = markdown.parse('## 结论\n第一行\n第二行\n\n- 甲\n- 乙\n\n> 提示\n\n---');
  assert.deepEqual(blocks.map(block => block.type), ['heading', 'paragraph', 'list', 'quote', 'rule']);
  const { root, window } = rendered('## 结论\n第一行\n第二行\n\n1. 甲\n2. 乙\n   - 子项');
  try {
    assert.equal(root.querySelector('h5').textContent, '结论');
    assert.equal(root.querySelectorAll('p br').length, 1);
    assert.equal(root.querySelectorAll('ol > li').length, 2);
    assert.equal(root.querySelector('ol li ul li').textContent, '子项');
  } finally { window.close(); }
});
test('会议待办这类表格渲染为表格，转义竖线与对齐保留', () => {
  const { root, window } = rendered('| 编号 | 任务 | 来源 |\n| --- | :---: | ---: |\n| T1 | 提交 \\| 初稿 | L1 |\n| T2 | 复核 |');
  try {
    assert.equal(root.querySelectorAll('th').length, 3);
    assert.equal(root.querySelectorAll('tbody tr').length, 2);
    assert.equal(root.querySelector('tbody td:nth-child(2)').textContent, '提交 | 初稿');
    assert.equal(root.querySelector('th:nth-child(2)').className, 'align-center');
    assert.equal(root.querySelectorAll('tbody tr:nth-child(2) td').length, 3);
    assert.equal(root.querySelector('.md-table').getAttribute('role'), 'region');
  } finally { window.close(); }
});
test('代码块原样保留并可复制，行内代码、粗体、斜体与删除线', () => {
  const copied = [];
  const { root, window } = rendered('```python\nprint("<b>")\n```\n\n用 `pip` 安装 **重要** *提示* ~~旧~~', { onCopy: text => copied.push(text) });
  try {
    assert.equal(root.querySelector('pre code').textContent, 'print("<b>")');
    assert.equal(root.querySelector('.md-code-head span').textContent, 'python');
    root.querySelector('.md-code-copy').click();
    assert.deepEqual(copied, ['print("<b>")']);
    assert.equal(root.querySelector('p code').textContent, 'pip');
    assert.equal(root.querySelector('strong').textContent, '重要');
    assert.equal(root.querySelector('em').textContent, '提示');
    assert.equal(root.querySelector('del').textContent, '旧');
  } finally { window.close(); }
});
test('原始HTML与脚本只作为文本，不产生元素或事件', () => {
  const { root, window } = rendered('<img src=x onerror=alert(1)>\n\n<script>alert(1)</script>\n\n**<a href="javascript:alert(1)">x</a>**');
  try {
    assert.equal(root.querySelectorAll('img,script,[onerror]').length, 0);
    assert.equal(root.querySelectorAll('a').length, 0);
    assert.match(root.textContent, /<img src=x onerror=alert\(1\)>/);
    assert.match(root.textContent, /<script>/);
  } finally { window.close(); }
});
test('链接只允许 http(s)/mailto，图片不加载远程地址', () => {
  const { root, window } = rendered('[文档](https://example.test/doc) [坏](javascript:alert(1)) [本地](/admin) ![图](https://example.test/a.png) 详见 https://example.test/x。');
  try {
    const links = [...root.querySelectorAll('a')];
    assert.deepEqual(links.map(link => link.getAttribute('href')), ['https://example.test/doc', 'https://example.test/a.png', 'https://example.test/x']);
    assert.ok(links.every(link => link.target === '_blank' && link.rel === 'noopener noreferrer'));
    assert.equal(root.querySelectorAll('img').length, 0);
    assert.match(root.textContent, /\[坏\]\(javascript:alert\(1\)\)/);
    assert.match(root.textContent, /\[本地\]\(\/admin\)/);
    assert.match(links[1].textContent, /^图片：图$/);
    assert.equal(markdown.safeHref('https://user:pass@example.test/'), '');
  } finally { window.close(); }
});
test('任务清单与松散列表', () => {
  const { root, window } = rendered('- [x] 已完成\n- [ ] 待处理\n\n1. 第一步\n\n   说明\n2. 第二步');
  try {
    assert.equal(root.querySelectorAll('.md-task').length, 2);
    assert.equal(root.querySelectorAll('.md-check.done').length, 1);
    assert.match(root.querySelector('.md-task').textContent, /已完成：已完成/);
    assert.equal(root.querySelectorAll('ol > li > p').length, 3);
  } finally { window.close(); }
});
test('未闭合代码块、深层嵌套与超长行保持有界并可读', () => {
  assert.equal(markdown.parse('```\n仍在输出').at(-1).text, '仍在输出');
  const deep = Array.from({ length: 40 }, (_, depth) => ' '.repeat(depth * 2) + '- 第' + depth + '层').join('\n');
  const started = Date.now();
  const { root, window } = rendered(deep + '\n\n' + '*'.repeat(20000) + '\n\n' + '['.repeat(5000));
  try {
    assert.ok(Date.now() - started < 2000);
    assert.ok(root.querySelectorAll('ul').length <= 7);
    assert.match(root.textContent, /第39层/);
  } finally { window.close(); }
});
