// 工坊直接生图：只调用用户在控制台授权的 Images API 连接。
// 不保存密钥，不自动重试；返回内容只作为图片地址使用，不当作 HTML 或脚本。
const LOCAL_HOSTS = ['localhost', '127.0.0.1', '[::1]'];
const ENDPOINT = '/images/generations';
export const IMAGE_SIZES = [
 { value: '', label: '服务默认尺寸' },
 { value: '1024x1024', label: '1024×1024 方形' },
 { value: '1024x1536', label: '1024×1536 竖幅' },
 { value: '1536x1024', label: '1536×1024 横幅' }
];
export const GENERATE_TIMEOUT_MS = 180000;
const STATUS_MESSAGES = {
 400: '请求被拒绝，请检查模型是否支持生图，或换一个尺寸。',
 401: 'API Key 无效或已过期，请重新选择密钥。',
 403: '当前密钥所在分组没有这个模型或生图权限。',
 404: '接口或模型不存在，请重新选择模型。',
 413: '提示词过长，请删减后再试。',
 429: '服务限流或额度不足，请稍后重试或检查额度。'
};
function fail(message) { throw new Error(message); }
function allowedURL(url) {
 return (url.protocol === 'https:' || (url.protocol === 'http:' && LOCAL_HOSTS.includes(url.hostname))) && !url.username && !url.password;
}
export function suggestedSize(ratio) {
 return ratio === 'portrait' ? '1024x1536' : ratio === 'wide' ? '1536x1024' : '1024x1024';
}
export function imageConnection(value) {
 const config = value && typeof value === 'object' ? value : {};
 if (config.kind !== 'image' || config.protocol !== 'images') fail('这个连接不是生图连接，请重新选择。');
 const model = typeof config.model === 'string' ? config.model.trim() : '';
 const key = typeof config.key === 'string' ? config.key.trim() : '';
 if (!model || model.length > 200) fail('请选择有效的生图模型。');
 if (key.length > 2048 || /\s/.test(key)) fail('API Key 格式有误，请重新选择密钥。');
 let url;
 try { url = new URL(String(config.base || '')); } catch { fail('生图接口地址无效，请联系站点管理员检查公开 API 地址。'); }
 if (!allowedURL(url) || url.search || url.hash) fail('生图接口需要 HTTPS 地址，本机预览可用 HTTP。');
 if (!key && !LOCAL_HOSTS.includes(url.hostname)) fail('缺少 API Key，请重新选择密钥。');
 let pathname = url.pathname.replace(/\/+$/, '');
 if (!pathname.endsWith(ENDPOINT)) pathname = (pathname || '/v1') + ENDPOINT;
 url.pathname = pathname;
 return { kind: 'image', protocol: 'images', model, key, endpoint: url.href };
}
export function imageRequest(connection, prompt, size = '') {
 const text = typeof prompt === 'string' ? prompt.trim() : '';
 if (!text) fail('提示词为空，请先填写或恢复示例材料。');
 if (text.length > 32000) fail('提示词过长，请删减到 32000 字以内。');
 if (size && !IMAGE_SIZES.some(option => option.value === size)) fail('请选择列表中的尺寸。');
 const headers = { 'Content-Type': 'application/json' };
 if (connection.key) headers.Authorization = 'Bearer ' + connection.key;
 return {
  url: connection.endpoint,
  options: { method: 'POST', headers, body: JSON.stringify({ model: connection.model, prompt: text, n: 1, ...(size ? { size } : {}) }), credentials: 'omit', redirect: 'error', referrerPolicy: 'no-referrer', cache: 'no-store' }
 };
}
export function imageResults(payload) {
 if (!payload || typeof payload !== 'object' || payload.error) fail('生图服务没有返回完成的结果，请检查模型与额度。');
 if (!Array.isArray(payload.data) || !payload.data.length) fail('接口没有返回图片，请确认所选模型支持 Images API。');
 const mime = { png: 'image/png', jpeg: 'image/jpeg', jpg: 'image/jpeg', webp: 'image/webp' }[payload.output_format] || 'image/png';
 return payload.data.slice(0, 4).map(item => {
  const revised = typeof item?.revised_prompt === 'string' ? item.revised_prompt.slice(0, 4000) : '';
  if (typeof item?.b64_json === 'string' && /^[A-Za-z0-9+/=\r\n]+$/.test(item.b64_json) && item.b64_json.length > 8) {
   return { src: 'data:' + mime + ';base64,' + item.b64_json.replace(/[\r\n]/g, ''), revised };
  }
  if (typeof item?.url === 'string') {
   if (/^data:image\/(png|jpeg|webp);base64,[A-Za-z0-9+/=]+$/.test(item.url)) return { src: item.url, revised };
   let url;
   try { url = new URL(item.url); } catch { fail('服务返回的图片地址格式不受支持。'); }
   if (!allowedURL(url)) fail('服务返回的图片地址格式不受支持。');
   return { src: url.href, revised };
  }
  fail('接口返回的图片格式不受支持。');
 });
}
export async function generateImage(connection, prompt, size, signal, transport = globalThis.fetch) {
 const plan = imageRequest(connection, prompt, size);
 let response;
 try { response = await transport(plan.url, { ...plan.options, signal }); }
 catch (error) {
  if (signal?.aborted || error?.name === 'AbortError') throw error;
  fail('无法连接生图接口，请检查网络或站点的公开 API 地址。');
 }
 if (!response.ok) fail(STATUS_MESSAGES[response.status] || '生图服务请求失败（HTTP ' + response.status + '），请稍后再试。');
 let payload;
 try { payload = await response.json(); } catch { fail('接口没有返回 JSON 结果，请确认选择的是 Images API 模型。'); }
 return imageResults(payload);
}
export function downloadName(itemId, index, src) {
 const format = /^data:image\/(png|jpeg|webp);base64,/.exec(src)?.[1];
 return format ? '魔法工坊-' + String(itemId).replace(/[^a-z0-9-]/gi, '') + '-' + (index + 1) + '.' + (format === 'jpeg' ? 'jpg' : format) : '';
}
