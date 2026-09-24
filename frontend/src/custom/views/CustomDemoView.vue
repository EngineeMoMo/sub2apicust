<!-- [CUSTOM] 脚手架演示页 —— 验证「custom/ 叠加层」整条链路(路由 + 主题变量)已通。
     确认 /custom-demo 能打开、色块是品牌色后，可删除本文件和 custom/routes.ts 里对应的路由项。
     见项目根 CUSTOMIZATIONS.md。 -->
<template>
  <div class="mx-auto max-w-xl p-8">
    <h1 class="text-2xl font-bold text-gray-900 dark:text-white">Custom 叠加层就绪 ✓</h1>
    <p class="mt-2 text-sm text-gray-600 dark:text-gray-300">
      这是通过 <code>src/custom/routes.ts</code> 新增的页面，未改动上游任何路由文件。
    </p>

    <!-- 用 primary token 的色块：颜色若来自 custom/theme.css 的变量，说明换肤链路已生效 -->
    <div class="mt-6 flex items-center gap-3">
      <div class="h-10 w-10 rounded-lg bg-primary-500"></div>
      <div class="h-10 w-10 rounded-lg bg-primary-600"></div>
      <div class="h-10 w-10 rounded-lg bg-primary-500/50"></div>
      <span class="text-sm text-gray-500 dark:text-gray-400">
        改 <code>custom/theme.css</code> 的 <code>--color-primary-*</code> 这几个色块会跟着变。
      </span>
    </div>

    <button
      type="button"
      class="mt-6 rounded-lg bg-primary-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-700"
      @click="count++"
    >
      点击计数：{{ count }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const count = ref(0)
</script>
