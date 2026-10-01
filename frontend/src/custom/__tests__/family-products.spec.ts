import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { familyLoginTarget, familyProducts, recipeDestination, studioDestination } from '@/custom/family/products'
import { safeGuestRedirect } from '@/custom/guest/navigation'

describe('家族产品地址与会话边界', () => {
  it('生产未配置地址时使用镜像内置配方，不把开发端口当上线地址', () => {
    const recipe = recipeDestination('', 'https://ai.mofamilys.com', false)
    expect(recipe).toEqual({ href: 'https://ai.mofamilys.com/recipes/', preview: false, invalid: false, integrated: true })
    expect(familyProducts(recipe, false)[1]).toMatchObject({ state: '本站产品', destination: '/tools/recipes', external: false })
  })

  it.each(['https://recipes.example.test/', '/recipes/'])('支持已配置的 HTTPS 或同源地址 %s', value => {
    const recipe = recipeDestination(value, 'https://ai.mofamilys.com', false)
    expect(recipe.href).toBe(new URL(value, 'https://ai.mofamilys.com').href)
    expect(recipe).toMatchObject({ preview: false, invalid: false })
    expect(familyProducts(recipe, false)[1].state).toBe(value === '/recipes/' ? '本站产品' : '独立产品')
  })

  it.each([
    'http://recipes.test/', 'http://127.0.0.1:4178/', '//recipes.test/',
    'https://user:password@recipes.test/', 'https://recipes.test/?key=secret',
    'https://recipes.test/#token', 'https://*.recipes.test/', 'javascript:alert(1)'
  ])('拒绝不安全或携带参数的发布地址 %s', value => {
    const recipe = recipeDestination(value, 'https://ai.mofamilys.com', false)
    expect(recipe).toEqual({ preview: false, invalid: true })
    expect(familyProducts(recipe, false)[1].state).toBe('地址配置有误')
  })

  it('仅本机开发预览添加公开 api_site 提示，不传凭据', () => {
    const recipe = recipeDestination('http://127.0.0.1:4178', 'http://127.0.0.1:4175', true)
    const destination = new URL(recipe.href!)
    expect(recipe.preview).toBe(true)
    expect(destination.origin).toBe('http://127.0.0.1:4178')
    expect([...destination.searchParams]).toEqual([['api_site', 'http://127.0.0.1:4175']])
    expect(recipeDestination('', 'https://ai.mofamilys.com', true).href).toBe('https://ai.mofamilys.com/recipes/')
    expect(recipeDestination('http://127.0.0.1:4178', 'https://ai.mofamilys.com', true).invalid).toBe(true)
  })

  it('四个产品按真实状态提供入口，不伪造桌面 SSO 或工坊上线', () => {
    const products = familyProducts({ href: 'https://recipes.test/', preview: false, invalid: false }, false)
    expect(products.map(product => product.id)).toEqual(['api', 'recipes', 'studio', 'synaroute'])
    expect(products[0].destination).toBe('/dashboard')
    expect(products[3].destination).toBe('/keys')
    expect(products[3].account).toContain('不等于已接通账号单点登录')
    expect(products[2].destination).toBeUndefined()
    expect(products[2].state).toBe('待配置发布地址')
    expect(familyProducts({ preview: false, invalid: false }, true)[0].destination).toBe('/admin/dashboard')
  })

  it('本机默认同源路径，可显式选择独立预览，远程 HTTP 仍不开放', () => {
    const recipe = recipeDestination('http://127.0.0.1:4178', 'http://127.0.0.1:8080', false)
    expect(recipe.preview).toBe(true)
    expect(recipe.href).toBe('http://127.0.0.1:4178/?api_site=http%3A%2F%2F127.0.0.1%3A8080')
    expect(recipeDestination('', 'http://127.0.0.1:8080', false)).toEqual({ href: 'http://127.0.0.1:8080/recipes/', integrated: true, preview: false, invalid: false })
    expect(recipeDestination('http://recipes.test', 'http://127.0.0.1:8080', false).invalid).toBe(true)
    expect(recipeDestination('http://127.0.0.1:4178', 'http://api.test', false).invalid).toBe(true)
  })

  it('登录先到 API 控制台，只传产品标识，不直接重定向外部来源', () => {
    const products = familyProducts({ href: 'https://recipes.test/', preview: false, invalid: false }, false)
    expect(familyLoginTarget(products[0]).query.redirect).toBe('/dashboard')
    expect(familyLoginTarget(products[3]).query.redirect).toBe('/dashboard?product=synaroute')
    const redirect = familyLoginTarget(products[1]).query.redirect
    expect(redirect).toBe('/dashboard?product=recipes')
    expect(safeGuestRedirect(redirect)).toBe(redirect)
  })

  it('工坊使用同样严格的地址校验，不传来源或凭据', () => {
    expect(studioDestination('', 'https://ai.mofamilys.com', false).href).toBe('https://ai.mofamilys.com/studio/')
    expect(studioDestination('', 'http://127.0.0.1:4175', true)).toEqual({ href: 'http://127.0.0.1:4175/studio/', integrated: true, preview: false, invalid: false })
    expect(studioDestination('http://127.0.0.1:4179', 'http://127.0.0.1:8080', false).preview).toBe(true)
    expect(studioDestination('https://studio.example.test/?key=secret', 'https://ai.mofamilys.com', false).invalid).toBe(true)
    expect(studioDestination('http://127.0.0.1:4179', 'https://ai.mofamilys.com', false).invalid).toBe(true)
  })

  it('Docker 和 CI 同时传入非敏感构建变量，不当成运行时配置', () => {
    const docker = readFileSync(resolve(process.cwd(), '../Dockerfile'), 'utf8')
    const workflow = readFileSync(resolve(process.cwd(), '../.github/workflows/custom-image.yml'), 'utf8')
    for (const suffix of ['URL', 'ORIGIN']) {
      expect(docker).toContain('ARG VITE_MAGIC_RECIPES_' + suffix + '=')
      expect(docker).toContain('VITE_MAGIC_RECIPES_' + suffix + '="${VITE_MAGIC_RECIPES_' + suffix + '}"')
      expect(workflow).toContain('VITE_MAGIC_RECIPES_' + suffix + '=${{ vars.MAGIC_RECIPES_' + suffix + ' }}')
    }
    expect(docker).not.toContain('ENV VITE_MAGIC_RECIPES_')
    expect(docker).toContain('ARG VITE_MAGIC_STUDIO_URL=')
    expect(docker).toContain('VITE_MAGIC_STUDIO_URL="${VITE_MAGIC_STUDIO_URL}"')
    expect(workflow).toContain('VITE_MAGIC_STUDIO_URL=${{ vars.MAGIC_STUDIO_URL }}')
  })
})
