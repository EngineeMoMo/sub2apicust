import createDOMPurify from 'dompurify'
import { describe, expect, it } from 'vitest'

describe('DOMPurify 安全回归', () => {
  it.each([
    ['$', '{securityProbe}'],
    ['{', '{securityProbe}}'],
    ['<', '%securityProbe%>']
  ])('清除影子根中跨节点拼接的模板表达式：%s%s', (prefix, suffix) => {
    const purifier = createDOMPurify(window)
    const host = document.createElement('div')
    const shadow = host.attachShadow({ mode: 'open' })
    shadow.append(
      document.createTextNode('保留内容：'),
      document.createTextNode(prefix),
      document.createElement('object'),
      document.createTextNode(suffix)
    )

    purifier.sanitize(host, { IN_PLACE: true, SAFE_FOR_TEMPLATES: true })

    expect(host.shadowRoot).toBe(shadow)
    expect(shadow.textContent).toContain('保留内容：')
    expect(shadow.textContent).not.toContain(prefix + suffix)
  })

  it.each(['parent', 'sibling'])('钩子移除%s后仍清除脱离节点的事件属性', (target) => {
    const purifier = createDOMPurify(window)
    const container = document.createElement('section')
    container.innerHTML = '<figure><img id="remove-trigger"><img id="detached-risk" onerror="securityProbe()"></figure><p>保留内容</p>'
    const riskyImage = container.querySelector('#detached-risk')
    if (!riskyImage) throw new Error('缺少攻击载荷节点')
    purifier.addHook('afterSanitizeAttributes', (element) => {
      if (element.id !== 'remove-trigger') return
      if (target === 'parent') element.parentElement?.remove()
      else element.nextElementSibling?.remove()
    })

    purifier.sanitize(container, { IN_PLACE: true })

    expect(container.contains(riskyImage)).toBe(false)
    expect(riskyImage.getAttribute('onerror')).toBeNull()
    expect(container.textContent).toContain('保留内容')
  })

  it('保留常规富文本并删除脚本和事件属性', () => {
    const purifier = createDOMPurify(window)
    const sanitized = purifier.sanitize('<p><strong>公告</strong><img src="/logo.png" onerror="securityProbe()"></p><script>securityProbe()</script>')
    const container = document.createElement('div')
    container.innerHTML = sanitized

    expect(container.querySelector('strong')?.textContent).toBe('公告')
    expect(container.querySelector('img')?.getAttribute('src')).toBe('/logo.png')
    expect(container.querySelector('[onerror], script')).toBeNull()
  })
})
