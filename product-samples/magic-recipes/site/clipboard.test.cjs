const test = require('node:test');
const assert = require('node:assert/strict');
const path = require('node:path');
const { createRequire } = require('node:module');
const { JSDOM } = createRequire(path.resolve(__dirname, '../../../frontend/package.json'))('jsdom');
const clipboard = require('./clipboard.cjs');
function page(writeText) {
  const dom = new JSDOM('<body><button id="trigger">复制</button><dialog id="modal"><button id="inside">复制</button></dialog></body>', { url: 'https://example.test/' });
  const win = dom.window;
  if (writeText) Object.defineProperty(win.navigator, 'clipboard', { value: { writeText } });
  return { dom, win, doc: win.document };
}
test('剪贴板接口可用时直接写入，不触发 execCommand', async () => {
  const written = [];
  const { dom, win, doc } = page(async text => { written.push(text); });
  try {
    let fallback = 0; doc.execCommand = () => { fallback++; return true; };
    assert.equal(await clipboard.write('第一段', win), true);
    assert.deepEqual(written, ['第一段']); assert.equal(fallback, 0);
  } finally { dom.window.close(); }
});
test('接口被拒时用 execCommand 复制完整文本，模态内放临时框，结束后清理并还原焦点', async () => {
  const { dom, win, doc } = page(async () => { throw new Error('NotAllowedError'); });
  try {
    const calls = [];
    doc.execCommand = command => { const area = doc.activeElement; calls.push({ command, text: area.value, selected: area.selectionEnd - area.selectionStart, inDialog: Boolean(area.closest('dialog')) }); return true; };
    doc.getElementById('trigger').focus();
    assert.equal(await clipboard.write('多行\n文本', win), true);
    assert.deepEqual(calls[0], { command: 'copy', text: '多行\n文本', selected: 5, inDialog: false });
    assert.equal(doc.activeElement.id, 'trigger'); assert.equal(doc.querySelectorAll('textarea').length, 0);
    doc.getElementById('modal').setAttribute('open', ''); doc.getElementById('inside').focus();
    assert.equal(await clipboard.write('模态内', win), true);
    assert.equal(calls[1].inDialog, true); assert.equal(doc.activeElement.id, 'inside');
  } finally { dom.window.close(); }
});
test('没有剪贴板接口且 execCommand 失败或不存在时返回 false，交给页面的手动复制提示', async () => {
  const { dom, win, doc } = page();
  try {
    assert.equal(await clipboard.write('无接口', win), false);
    doc.execCommand = () => false;
    assert.equal(await clipboard.write('被拒', win), false);
    assert.equal(doc.querySelectorAll('textarea').length, 0);
  } finally { dom.window.close(); }
});
