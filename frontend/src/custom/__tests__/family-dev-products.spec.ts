import { describe, expect, it } from 'vitest'
import { familyProductEntry } from '../family/dev-products.mjs'

describe('内置子产品的 Vite 开发入口', () => {
  it.each(['recipes', 'studio'])('把 /%s/ 映射到静态首页并保留宿主参数', product => {
    expect(familyProductEntry('/' + product + '/?embedded=1', 'GET')).toEqual({ redirect: false, url: '/' + product + '/index.html?embedded=1' })
    expect(familyProductEntry('/' + product + '/', 'HEAD')).toEqual({ redirect: false, url: '/' + product + '/index.html' })
  })
  it.each(['recipes', 'studio'])('为 /%s 补齐尾斜杠，避免相对资源越出目录', product => {
    expect(familyProductEntry('/' + product + '?embedded=1')).toEqual({ redirect: true, url: '/' + product + '/?embedded=1' })
  })
  it('不改写 API、资源、宿主路由和未知路径', () => {
    for (const url of ['/api/v1/keys', '/studio/app.mjs', '/recipes/index.html', '/tools/studio', '/studio/../keys', '//studio/', '/studios/', '']) expect(familyProductEntry(url)).toBeNull()
  })
  it('不接管写入请求', () => {
    for (const method of ['POST', 'PUT', 'PATCH', 'DELETE']) expect(familyProductEntry('/studio/', method)).toBeNull()
  })
})
