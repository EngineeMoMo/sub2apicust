// [CUSTOM] 新增页面的路由集中在这里，见项目根 CUSTOMIZATIONS.md。
// router/index.ts 只在 404 兜底前 `...customRoutes` 展开一次 —— 那是唯一接缝。
// 往这里加你自己的页面即可（component 用懒加载，指向 custom/views/ 下的 .vue）。
import type { RouteRecordRaw } from 'vue-router'

export const customRoutes: RouteRecordRaw[] = [
  // —— 脚手架演示，确认整条链路(路由+主题)通了。验证后可直接删掉这一项。
  {
    path: '/custom-demo',
    name: 'CustomDemo',
    component: () => import('@/custom/views/CustomDemoView.vue'),
    meta: {
      title: 'Custom Demo'
    }
  }
  // —— 你的新页面示例：
  // {
  //   path: '/my-page',
  //   name: 'MyPage',
  //   component: () => import('@/custom/views/MyPageView.vue'),
  //   meta: { title: '我的页面', requiresAuth: true }
  // }
]
