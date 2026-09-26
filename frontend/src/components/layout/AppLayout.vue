<template>
  <!-- [CUSTOM] 品牌布局样式钩子；原菜单、页头、内容插槽与响应式行为保留。 -->
  <div class="mofa-workspace min-h-screen bg-gray-50 dark:bg-dark-950" :class="{ 'mofa-workspace--collapsed': sidebarCollapsed }">
    <!-- Background Decoration -->
    <div class="mofa-workspace-backdrop pointer-events-none fixed inset-0 bg-mesh-gradient"></div>

    <!-- Sidebar -->
    <AppSidebar />

    <!-- Main Content Area -->
    <div
      class="mofa-workspace-shell relative min-h-screen transition-all duration-300"
      :class="[sidebarCollapsed ? 'lg:ml-[72px]' : 'lg:ml-64']"
    >
      <!-- Header -->
      <AppHeader />

      <!-- Main Content -->
      <main class="mofa-workspace-main p-4 md:p-6 lg:p-8">
        <!-- [CUSTOM] 正文标题与可选操作插槽，不复制业务页面逻辑。 -->
        <WorkspaceHeading>
          <template v-if="$slots['page-actions']" #actions><slot name="page-actions" /></template>
        </WorkspaceHeading>
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import '@/styles/onboarding.css'
import { computed, onMounted } from 'vue'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import { useOnboardingTour } from '@/composables/useOnboardingTour'
import { useOnboardingStore } from '@/stores/onboarding'
import AppSidebar from './AppSidebar.vue'
import AppHeader from './AppHeader.vue'
// [CUSTOM] 品牌展示组件独立维护。
import WorkspaceHeading from '@/custom/components/WorkspaceHeading.vue'

const appStore = useAppStore()
const authStore = useAuthStore()
const sidebarCollapsed = computed(() => appStore.sidebarCollapsed)
const isAdmin = computed(() => authStore.user?.role === 'admin')

const { replayTour } = useOnboardingTour({
  storageKey: isAdmin.value ? 'admin_guide' : 'user_guide',
  autoStart: true
})

const onboardingStore = useOnboardingStore()

onMounted(() => {
  onboardingStore.setReplayCallback(replayTour)
})

defineExpose({ replayTour })
</script>
