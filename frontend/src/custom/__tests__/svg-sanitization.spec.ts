import { describe, expect, it } from 'vitest'
import { sanitizeSvg } from '@/utils/sanitize'

// [CUSTOM] 净化库升级后，管理员上传的 SVG 图标保留绘图，移除可执行内容。
describe('SVG 图标净化', () => {
  it('保留图标路径与颜色，移除脚本、事件和 foreignObject', () => {
    const clean = sanitizeSvg('<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" onload="alert(1)"><path fill="#096B68" d="M0 0h24v24z"/><script>alert(1)</script><foreignObject><iframe src="https://example.com"/></foreignObject></svg>')
    const doc = new DOMParser().parseFromString(clean, 'image/svg+xml')
    expect(doc.querySelector('path')?.getAttribute('fill')).toBe('#096B68')
    expect(doc.querySelector('path')?.getAttribute('d')).toBe('M0 0h24v24z')
    expect(doc.querySelector('svg')?.getAttribute('viewBox')).toBe('0 0 24 24')
    expect(doc.querySelector('script, foreignObject, iframe, [onload]')).toBeNull()
  })

  it('移除链接中的 javascript 协议并允许空图标', () => {
    const clean = sanitizeSvg('<svg xmlns="http://www.w3.org/2000/svg"><a href="javascript:alert(1)"><path d="M0 0h1"/></a></svg>')
    expect(clean).not.toContain('javascript:')
    expect(clean).toContain('<path')
    expect(sanitizeSvg('')).toBe('')
  })
})
