// [CUSTOM] 首屏骨架与语言包预加载：只改构建出的 index.html，不改上游入口或 Go 注入逻辑。
// 后端会把整页的 __CSP_NONCE_VALUE__ 替换为本次请求的 nonce（embed_on.go serveIndexHTML）。
const NONCE = '__CSP_NONCE_VALUE__'
const LOCALE_ENTRY = /[\\/]src[\\/]i18n[\\/]locales[\\/](en|zh)[\\/]index\.ts$/
// 与 router/index.ts 的公开官网白名单保持一致；仅用于选择骨架外观，不参与鉴权。
const PUBLIC_PATH = /^\/(home|brand|family|api|plans|guide|faq)?(\/|$)/
const AUTH_PATH = /^\/(login|register|email-verify|forgot-password|reset-password)(\/|$)/

// 颜色与 theme.css 的 --brand-page/--brand-border 同值；骨架挂在 #app 内，Vue 挂载即整体移除。
const SHELL_STYLE = `
body{margin:0}
.mofa-shell{min-height:100vh;min-height:100svh;background:#f5f7f7}
html.dark .mofa-shell{background:#162124}
.mofa-shell-bar{position:fixed;top:0;left:0;right:0;height:2px;background:linear-gradient(90deg,transparent,#096b68,transparent);background-size:40% 100%;background-repeat:no-repeat;animation:mofa-shell-bar 1.1s ease-in-out infinite}
html.dark .mofa-shell-bar{background-image:linear-gradient(90deg,transparent,#a1d9ce,transparent)}
.mofa-shell-public{display:none}
[data-shell=public] .mofa-shell-public{display:block}
.mofa-shell-mast{width:min(1160px,calc(100% - 64px));margin-inline:auto}
.mofa-shell-hero{width:min(1120px,calc(100% - 48px));margin-inline:auto}
.mofa-shell-mast{display:flex;align-items:center;justify-content:space-between;height:91px;border-bottom:1px solid #dce5e5}
html.dark .mofa-shell-mast{border-color:#36494d}
.mofa-shell-hero{display:grid;grid-template-columns:1.18fr 1fr;gap:52px;align-items:center;padding:100px 8px 48px}
.mofa-shell span{display:block;border-radius:8px;background:#e4edec}
html.dark .mofa-shell span{background:#22363a}
.mofa-shell-logo{width:42px;height:42px}
.mofa-shell-nav{width:min(520px,50%);height:14px}
.mofa-shell-eyebrow{width:180px;height:10px;margin-bottom:22px}
.mofa-shell-title{width:80%;height:50px;margin-bottom:14px}
.mofa-shell-title.is-short{width:52%}
.mofa-shell-text{width:70%;height:14px;margin-top:26px}
.mofa-shell-card{aspect-ratio:1.45;border-radius:24px!important}
.mofa-shell-public span{animation:mofa-shell-pulse 1.4s ease-in-out infinite}
@keyframes mofa-shell-bar{from{background-position:-40% 0}to{background-position:140% 0}}
@keyframes mofa-shell-pulse{50%{opacity:.55}}
@media (max-width:767px){.mofa-shell-mast{width:calc(100% - 40px);height:79px}.mofa-shell-nav{width:42px;height:42px}.mofa-shell-hero{width:calc(100% - 32px);grid-template-columns:1fr;padding:64px 0 36px;gap:28px}.mofa-shell-title{height:40px}}
@media (prefers-reduced-motion:reduce){.mofa-shell-bar,.mofa-shell-public span{animation:none}}
`.replace(/\n/g, '')

const SHELL_MARKUP = '<div class="mofa-shell" aria-hidden="true"><div class="mofa-shell-bar"></div>'
  + '<div class="mofa-shell-public"><div class="mofa-shell-mast"><span class="mofa-shell-logo"></span><span class="mofa-shell-nav"></span></div>'
  + '<div class="mofa-shell-hero"><div><span class="mofa-shell-eyebrow"></span><span class="mofa-shell-title"></span>'
  + '<span class="mofa-shell-title is-short"></span><span class="mofa-shell-text"></span></div><span class="mofa-shell-card"></span></div></div></div>'

/**
 * 早期脚本：与 main.ts/i18n 同口径地提前确定深色与语言，并标记骨架外观。
 * 仅读取 localStorage 与浏览器语言，不读取任何凭据。
 */
export function earlyScript(locales = {}, base = '/') {
  const files = JSON.stringify(Object.fromEntries(Object.entries(locales).map(([code, file]) => [code, base + file])))
  return `(function(){var d=document.documentElement,s=null,l=null;try{s=localStorage.getItem('theme');l=localStorage.getItem('sub2api_locale')}catch(e){}`
    + `if(s==='dark')d.classList.add('dark');var p=location.pathname;`
    + `d.setAttribute('data-shell',${AUTH_PATH}.test(p)?'auth':${PUBLIC_PATH}.test(p)?'public':'app');`
    + `var f=${files};var c=l==='en'||l==='zh'?l:((navigator.language||'').toLowerCase().indexOf('zh')===0?'zh':'en');`
    + `if(f[c]){var k=document.createElement('link');k.rel='modulepreload';k.crossOrigin='';k.href=f[c];document.head.appendChild(k)}})();`
}

export function renderAppShell(html, locales = {}, base = '/') {
  if (!html.includes('<div id="app"></div>') || html.includes('mofa-shell')) return html
  const head = `<style nonce="${NONCE}">${SHELL_STYLE}</style><script nonce="${NONCE}">${earlyScript(locales, base)}</script>`
  const charset = html.match(/<meta charset=[^>]*>/i)
  const withHead = charset
    ? html.replace(charset[0], charset[0] + head)
    : html.replace(/<head>/i, match => match + head)
  return withHead.replace('<div id="app"></div>', `<div id="app">${SHELL_MARKUP}</div>`)
}

export function localeChunks(bundle = {}) {
  const locales = {}
  for (const chunk of Object.values(bundle)) {
    const match = chunk?.type === 'chunk' && chunk.facadeModuleId ? LOCALE_ENTRY.exec(chunk.facadeModuleId) : null
    if (match) locales[match[1]] = chunk.fileName
  }
  return locales
}

export function mofaAppShell() {
  let base = '/'
  return {
    name: 'mofa-app-shell',
    configResolved(config) { base = config.base || '/' },
    transformIndexHtml: {
      order: 'post',
      handler(html, context) {
        return renderAppShell(html, localeChunks(context?.bundle), base)
      }
    }
  }
}
