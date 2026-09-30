(function (root, factory) {
  const api = factory();
  if (typeof module === 'object' && module.exports) module.exports = api;
  else root.MofaRecipeConnect = api;
})(globalThis, () => {
  'use strict';
  function portalURL(value) {
    let url;
    try { url = new URL(value); } catch { throw new Error('请填写完整的魔法 API 网站地址。'); }
    const local = ['localhost', '127.0.0.1', '[::1]'].includes(url.hostname);
    if (!(url.origin === 'https://ai.mofamilys.com' || (local && url.protocol === 'http:'))) throw new Error('请使用 https://ai.mofamilys.com，或本机魔法 API 测试站。其他网站请使用手动配置。');
    if (url.username || url.password || url.search || url.hash || url.pathname !== '/') throw new Error('网站地址仅填写域名与端口，不包含路径、密码或参数。');
    return url;
  }
  function begin(portal, origin, crypto) {
    const url = portalURL(portal);
    const page = new URL(origin);
    const local = ['localhost', '127.0.0.1', '[::1]'].includes(page.hostname);
    if (page.origin !== origin || !(page.protocol === 'https:' || (page.protocol === 'http:' && local))) throw new Error('请通过 HTTPS 网站或本机 HTTP 预览打开配方页，不支持 file 地址登录接入。');
    const bytes = crypto.getRandomValues(new Uint8Array(32));
    const nonce = Array.from(bytes, byte => byte.toString(16).padStart(2, '0')).join('');
    return { portal: url.origin, origin, nonce };
  }
  function loginURL(connection, kind) {
    if (!['text', 'image'].includes(kind)) throw new Error('请选择文字或生图用途。');
    const url = new URL('/connect/recipes', connection.portal);
    url.searchParams.set('origin', connection.origin);
    url.searchParams.set('nonce', connection.nonce);
    url.searchParams.set('kind', kind);
    return url.href;
  }
  function accepts(event, connection, now = Date.now()) {
    return Boolean(connection && now < connection.expires && event.source === connection.popup
      && event.origin === connection.portal && event.data?.type === 'mofa-recipes-connection'
      && event.data.nonce === connection.nonce && event.data.config?.kind === connection.kind);
  }
  return { portalURL, begin, loginURL, accepts };
});
