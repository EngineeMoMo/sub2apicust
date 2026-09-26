import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore, useAuthStore } from '@/stores'
import { useAdminSettingsStore } from '@/stores/adminSettings'
import { resolveRouteMetaKeys } from '@/router/title'
import { resolveSiteBillingMode } from '@/utils/siteBillingMode'
import { intelligenceCopy } from '@/custom/intelligence/copy'

export function useWorkspaceHeading() {
  const route = useRoute()
  const { t, locale } = useI18n()
  const appStore = useAppStore()
  const authStore = useAuthStore()
  const adminSettingsStore = useAdminSettingsStore()

  const routeMetaKeys = computed(() => resolveRouteMetaKeys(route, {
    billingMode: resolveSiteBillingMode(appStore.cachedPublicSettings),
  }))

  const pageTitle = computed(() => {
    if (route.name === 'Intelligence') return intelligenceCopy[locale.value.startsWith('zh') ? 'zh' : 'en'].title
    if (route.name === 'CustomPage') {
      const id = route.params.id as string
      const publicItems = appStore.cachedPublicSettings?.custom_menu_items ?? []
      const menuItem = publicItems.find((item) => item.id === id)
        ?? (authStore.isAdmin ? adminSettingsStore.customMenuItems.find((item) => item.id === id) : undefined)
      if (menuItem?.label) return menuItem.label
    }
    const titleKey = routeMetaKeys.value.titleKey
    if (titleKey) {
      return t(titleKey)
    }
    return (route.meta.title as string) || ''
  })

  const pageDescription = computed(() => {
    if (route.name === 'Intelligence') return intelligenceCopy[locale.value.startsWith('zh') ? 'zh' : 'en'].description
    const descKey = routeMetaKeys.value.descriptionKey
    if (descKey) {
      return t(descKey)
    }
    return (route.meta.description as string) || ''
  })

  return { pageTitle, pageDescription }
}
