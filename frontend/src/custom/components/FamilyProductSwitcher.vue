<template>
  <div class="mofa-product-switcher">
    <nav class="mofa-product-capsule" aria-label="魔法家族产品切换">
      <template v-for="product in products" :key="product.id">
        <button v-if="!product.destination" type="button" disabled :title="product.state" :class="'mofa-switch-' + product.id" :aria-label="product.name + '：' + product.state"><FamilyProductMark :product="product.id" />{{ product.name }}</button>
        <a v-else-if="product.external" :href="product.destination" target="_blank" rel="noopener noreferrer" :class="'mofa-switch-' + product.id" :aria-label="product.name + '（新窗口）'" :title="product.state + ' · ' + product.account"><FamilyProductMark :product="product.id" />{{ product.name }}<Icon name="externalLink" size="xs" aria-hidden="true" /></a>
        <RouterLink v-else :to="product.destination" :aria-current="activeProduct === product.id ? 'page' : undefined" :class="['mofa-switch-' + product.id, { 'mofa-product-current': activeProduct === product.id }]" :aria-label="product.id === 'synaroute' ? 'SynaRoute（配置 API 密钥）' : undefined" :title="product.account"><FamilyProductMark :product="product.id" />{{ product.id === 'api' ? 'API 控制台' : product.name }}</RouterLink>
      </template>
    </nav>
    <aside v-if="intent" class="mofa-product-intent" aria-label="继续打开产品">
      <div><strong>继续使用{{ intent.name }}</strong><p>已进入 API 控制台。{{ intent.external ? '独立部署的产品将在新窗口打开。' : intent.id === 'synaroute' ? '进入密钥页，主动配置桌面工具。' : '点击即可在当前控制台工作区打开，登录会保留。' }}</p></div>
      <FamilyProductAction :product="intent" />
      <button type="button" class="mofa-product-dismiss" aria-label="留在 API 控制台" @click="dismissIntent"><Icon name="x" size="sm" aria-hidden="true" /></button>
    </aside>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores'
import Icon from '@/components/icons/Icon.vue'
import FamilyProductAction from '@/custom/components/FamilyProductAction.vue'
import FamilyProductMark from '@/custom/components/FamilyProductMark.vue'
import { familyProducts, recipeDestination, studioDestination } from '@/custom/family/products'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const origin = window.location.origin
const recipe = recipeDestination(String(import.meta.env.VITE_MAGIC_RECIPES_URL || ''), origin, import.meta.env.DEV)
const studio = studioDestination(String(import.meta.env.VITE_MAGIC_STUDIO_URL || ''), origin, import.meta.env.DEV)
const products = computed(() => familyProducts(recipe, auth.isAdmin, studio))
const activeProduct = computed(() => {
  if (route.path === '/tools/recipes') return 'recipes'
  if (route.path === '/tools/studio' || route.path.startsWith('/tools/studio/')) return 'studio'
  if (route.path === '/keys' && route.query.product === 'synaroute') return 'synaroute'
  return 'api'
})
const intent = computed(() => auth.isAuthenticated && ['/dashboard', '/admin/dashboard'].includes(route.path) && typeof route.query.product === 'string'
  ? products.value.find(product => product.id !== 'api' && product.id === route.query.product && product.destination)
  : undefined)

function dismissIntent() {
  const query = { ...route.query }
  delete query.product
  void router.replace({ path: route.path, query, hash: route.hash })
}
</script>
