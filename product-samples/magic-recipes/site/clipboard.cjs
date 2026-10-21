// 复制到剪贴板：先用异步剪贴板接口，被拒（页面未聚焦、非安全上下文、部分内置浏览器）时与主站 useClipboard 一样改用 execCommand。
// 临时文本框放进已打开的模态框，避免模态打开时页面其余部分被设为 inert 导致无法选中；结束后把焦点还给原来的元素。
const RecipeClipboard = (() => {
  async function write(text, win = globalThis) {
    const value = String(text ?? '');
    try {
      if (win.navigator?.clipboard?.writeText) { await win.navigator.clipboard.writeText(value); return true; }
    } catch {}
    const doc = win.document;
    if (!doc?.body) return false;
    const active = doc.activeElement;
    const area = doc.createElement('textarea');
    area.value = value; area.readOnly = true; area.setAttribute('aria-hidden', 'true');
    area.style.cssText = 'position:fixed;left:0;top:0;width:1px;height:1px;opacity:0;pointer-events:none';
    (doc.querySelector('dialog[open]') || doc.body).append(area);
    try {
      area.focus({ preventScroll: true }); area.select(); area.setSelectionRange(0, value.length);
      return doc.execCommand('copy') === true;
    } catch { return false; }
    finally {
      area.remove();
      if (active?.isConnected && typeof active.focus === 'function') active.focus({ preventScroll: true });
    }
  }
  return { write };
})();
if (typeof module !== 'undefined' && module.exports) module.exports = RecipeClipboard;
else globalThis.MofaClipboard = RecipeClipboard;
