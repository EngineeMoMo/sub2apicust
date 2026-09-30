import { describe, expect, it } from 'vitest'
import authLayout from '@/components/layout/AuthLayout.vue?raw'
import appLayout from '@/components/layout/AppLayout.vue?raw'
import appHeader from '@/components/layout/AppHeader.vue?raw'
import sidebar from '@/components/layout/AppSidebar.vue?raw'
import familyHomeIcon from '@/custom/components/FamilyHomeIcon.vue?raw'
import tableLayout from '@/components/layout/TablePageLayout.vue?raw'
import main from '@/main.ts?raw'
import router from '@/router/index.ts?raw'
import paymentView from '@/views/user/PaymentView.vue?raw'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const theme = readFileSync(resolve(process.cwd(), 'src/custom/theme.css'), 'utf8')
import brandHome from '@/custom/views/BrandHomeView.vue?raw'
import apiLanding from '@/custom/views/ApiLandingView.vue?raw'
import publicLayout from '@/custom/components/PublicLayout.vue?raw'
import { customRoutes } from '@/custom/routes'

describe('雾钛青与上游布局的升级契约', () => {
  it('工坊三图在正常流等宽分列而非绝对叠放，保留置顶与减少动画支持', () => {
    expect(theme).toMatch(/\.mofa-studio-images\s*\{[^}]*display:\s*grid;[^}]*grid-template-columns:\s*repeat\(3, minmax\(0, 1fr\)\);/)
    expect(theme).toMatch(/\.mofa-family-studio\s*\{[^}]*grid-template-columns:\s*minmax\(0, \.8fr\) minmax\(0, 1\.2fr\);/)
    expect(theme).toMatch(/\.mofa-studio-card\s*\{[^}]*position:\s*relative;/)
    expect(theme).toMatch(/\.mofa-studio-card:focus-visible\s*\{[^}]*z-index:\s*30;/)
    expect(theme).toMatch(/@media \(prefers-reduced-motion: reduce\)[^\n]*\.mofa-studio-card[^\n]*transition: none;/)
  })
  it('家族首页使用独立皇冠，用户与管理员仪表盘保留原图标', () => {
    expect(sidebar).toContain("path: '/family', label: '家族首页', icon: FamilyHomeIcon")
    expect(sidebar).toContain("path: '/dashboard', label: t('nav.dashboard'), icon: DashboardIcon")
    expect(sidebar).toContain("path: '/admin/dashboard', label: t('nav.dashboard'), icon: DashboardIcon")
    expect(familyHomeIcon).toContain('stroke="currentColor"')
    expect(familyHomeIcon).toContain('aria-hidden="true"')
  })
  it('充值页在纵向flex正文中占满可用宽度，同时保留原最大宽度', () => {
    expect(paymentView).toContain('class="mx-auto max-w-4xl space-y-6"')
    expect(theme).toMatch(/\.mofa-workspace-main\s*>\s*\.mx-auto\.max-w-4xl\s*\{\s*width:\s*100%;\s*\}/)
  })
  it('认证页展示接缝不替代业务和页脚插槽', () => {
    for (const hook of ['mofa-auth', 'mofa-auth-content', 'mofa-auth-form', 'mofa-auth-card']) {
      expect(authLayout).toContain(hook)
    }
    expect(authLayout).toContain('<BrandPanel :site-name="siteName"')
    expect(authLayout).toContain('<slot />')
    expect(authLayout).toContain('<slot name="footer"')
    expect(authLayout).toContain('sanitizeUrl')
  })

  it('工作台继续装配上游页头、侧栏和内容插槽', () => {
    expect(appLayout).toContain('mofa-workspace')
    expect(appLayout).toContain('mofa-workspace-backdrop')
    expect(appLayout).toContain('<AppSidebar')
    expect(appLayout).toContain('<AppHeader')
    expect(appLayout).toContain('<slot />')
    expect(appLayout).toContain("sidebarCollapsed ? 'lg:ml-[72px]' : 'lg:ml-64'")
    expect(appHeader).toContain('mofa-workspace-header')
    expect(appHeader).toContain('mofa-topbar-context')
    expect(appLayout).toContain('<WorkspaceHeading>')
    expect(appLayout).toContain('<slot name="page-actions" />')
    expect(appHeader).toContain('toggleMobileSidebar')
    expect(appHeader).toContain('<FamilyProductSwitcher v-if="user" />')
    expect(appHeader).toContain('// [CUSTOM] 产品切换实现位于叠加层')
  })

  it('主题依赖的上游公共类仍存在，变化时提示人工复核', () => {
    for (const selector of ['sidebar-header', 'sidebar-logo', 'sidebar-link-active', 'sidebar-brand-title', 'sidebar-nav']) {
      expect(sidebar).toContain(selector)
    }
    for (const selector of ['table-page-layout', 'table-scroll-container']) {
      expect(tableLayout).toContain(selector)
    }
  })

  it('沿用原主题与根路由，公开首页复用游客布局', () => {
    expect(main.indexOf("import './custom/theme.css'")).toBeGreaterThan(main.indexOf("import './style.css'"))
    expect(router).toContain('...customRoutes')
    const brandRoute = customRoutes.find(route => route.path === '/brand')
    expect(brandRoute?.meta?.requiresAuth).toBe(false)
    expect(customRoutes.some(route => route.path === '/')).toBe(false)
    expect(publicLayout).toContain("app.fetchPublicSettings()")
    expect(brandHome).toContain('<FamilyHomeView />')
    expect(apiLanding).toContain('to="/keys"')
    expect(apiLanding).toContain('to="/usage"')
  })

  it('视觉层不使用构建 hash、强制覆盖或概念稿专用运行时', () => {
    expect(theme).not.toContain('!important')
    expect(theme).not.toContain('data-v-')
    expect(brandHome).not.toContain('window.openai')
    expect(theme).toContain('.input.input-error')
    expect(theme).toContain('html.dark .text-gradient')
  })
})
