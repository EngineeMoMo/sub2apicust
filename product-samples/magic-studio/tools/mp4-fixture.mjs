// 测试用的最小 MP4 头：ftyp＋moov（mvhd 时长、tkhd 宽高）＋mdat，足够让 mp4Info 解析，不可播放。
const box = (type, body) => { const out = new Uint8Array(8 + body.length), view = new DataView(out.buffer); view.setUint32(0, out.length); out.set([...type].map(c => c.charCodeAt(0)), 4); out.set(body, 8); return out; };
const cat = (...parts) => { const out = new Uint8Array(parts.reduce((n, p) => n + p.length, 0)); let o = 0; for (const p of parts) { out.set(p, o); o += p.length; } return out; };
export function fakeMp4({ seconds = 5, width = 854, height = 480, fill = 0 } = {}) {
 const mvhd = new Uint8Array(100), mv = new DataView(mvhd.buffer); mv.setUint32(12, 1000); mv.setUint32(16, Math.round(seconds * 1000));
 const tkhd = new Uint8Array(84), tk = new DataView(tkhd.buffer); tk.setUint32(76, width * 65536); tk.setUint32(80, height * 65536);
 return cat(box('ftyp', new TextEncoder().encode('isom\0\0\0\0isom')), box('moov', cat(box('mvhd', mvhd), box('trak', box('tkhd', tkhd)))), box('mdat', new Uint8Array(64).fill(fill)));
}
