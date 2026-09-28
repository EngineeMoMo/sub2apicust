import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import ProxyAdBanner from '@/components/common/ProxyAdBanner.vue'
import header from '@/components/layout/AppHeader.vue?raw'
import home from '@/views/HomeView.vue?raw'
import usage from '@/views/KeyUsageView.vue?raw'

describe('上游推广入口清理', () => {
  it('代理广告兼容组件不渲染链接或营销内容', () => {
    const wrapper = mount(ProxyAdBanner)
    expect(wrapper.find('a').exists()).toBe(false)
    expect(wrapper.text()).toBe('')
  })
  it.each([['用户菜单', header], ['首页', home], ['用量查询', usage]])('%s不再包含上游仓库推广', (_name, source) => {
    expect(source).not.toContain('https://github.com/Wei-Shaw/sub2api')
    expect(source).not.toContain('githubUrl')
    expect(source).not.toContain("t('nav.github')")
  })
  it('保留客服配置和站点文档与版权内容', () => {
    expect(header).toContain('appStore.contactInfo')
    for (const source of [home, usage]) {
      expect(source).toContain(':href="docUrl"')
      expect(source).toContain("t('home.footer.allRightsReserved')")
    }
  })
})
