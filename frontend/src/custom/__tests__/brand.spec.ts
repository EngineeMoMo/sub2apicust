import { afterEach, describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import BrandPanel from '@/custom/components/BrandPanel.vue'
import BrandThemeToggle from '@/custom/components/BrandThemeToggle.vue'
import { brandCopy } from '@/custom/brand/copy'

function i18n(locale = 'zh') {
  return createI18n({ legacy: false, locale, messages: {} })
}

afterEach(() => {
  document.documentElement.classList.remove('dark')
  localStorage.removeItem('theme')
})

describe('品牌展示组件', () => {
  it('保留动态站点名、批准文案和可选操作插槽', () => {
    const wrapper = mount(BrandPanel, {
      props: { siteName: '测试品牌' },
      slots: { actions: '<button>真实入口</button>' },
      global: { plugins: [i18n()] }
    })
    expect(wrapper.attributes('aria-label')).toBe('测试品牌')
    expect(wrapper.text()).toContain('让每一次连接，')
    expect(wrapper.text()).toContain('真实入口')
    expect(wrapper.find('.mofa-core-mark').attributes('style')).toContain('mofa-mark.webp')
    expect(wrapper.find('.mofa-brand-study').attributes('aria-hidden')).toBe('true')
    wrapper.unmount()
  })

  it('未传操作时不生成假按钮，英文使用独立文案', () => {
    const wrapper = mount(BrandPanel, { props: { siteName: 'Example' }, global: { plugins: [i18n('en')] } })
    expect(wrapper.find('.mofa-hero-actions').exists()).toBe(false)
    expect(wrapper.text()).toContain('Every connection,')
    expect(Object.keys(brandCopy.en).sort()).toEqual(Object.keys(brandCopy.zh).sort())
    wrapper.unmount()
  })
})

describe('品牌主题切换', () => {
  it('挂载不改变默认深色，只有点击才写入选择', async () => {
    document.documentElement.classList.add('dark')
    const wrapper = mount(BrandThemeToggle, { global: { plugins: [i18n()] } })
    expect(document.documentElement.classList.contains('dark')).toBe(true)
    expect(localStorage.getItem('theme')).toBe(null)
    expect(wrapper.attributes('aria-label')).toBe('切换浅色模式')
    await wrapper.trigger('click')
    expect(localStorage.getItem('theme')).toBe('light')
    expect(document.documentElement.classList.contains('dark')).toBe(false)
    await wrapper.trigger('click')
    expect(localStorage.getItem('theme')).toBe('dark')
    wrapper.unmount()
  })

  it('挂载不覆盖用户的浅色选择', () => {
    localStorage.setItem('theme', 'light')
    const wrapper = mount(BrandThemeToggle, { global: { plugins: [i18n('en')] } })
    expect(document.documentElement.classList.contains('dark')).toBe(false)
    expect(localStorage.getItem('theme')).toBe('light')
    expect(wrapper.attributes('aria-label')).toBe('Switch to dark mode')
    wrapper.unmount()
  })
})
