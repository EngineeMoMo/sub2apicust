<template>
  <PublicLayout>
    <h1 class="sr-only">{{ app.siteName }} · 多模型 API 服务</h1>
    <BrandPanel :site-name="app.siteName">
      <template #actions>
        <RouterLink to="/plans" class="btn btn-primary">先看看订阅套餐</RouterLink>
        <RouterLink v-if="!auth.isAuthenticated" to="/preview" class="btn btn-secondary">游客预览控制台</RouterLink>
        <RouterLink to="/guide" class="btn btn-secondary">查看接入教程</RouterLink>
      </template>
    </BrandPanel>
    <section id="family-products" class="mofa-family-products" aria-labelledby="family-products-title">
      <div class="mofa-family-section-heading">
        <h2 id="family-products-title">同一个家族，不同的拿手好戏。</h2>
        <p>选择一款工具，看看它能做什么。</p>
      </div>
      <div class="mofa-family-deck">
        <div class="mofa-family-launcher" role="tablist" aria-label="预览家族产品">
          <button v-for="(product, index) in products" :id="'family-tab-' + product.id" :key="product.id" :ref="element => setTab(element, index)" type="button" role="tab" :aria-selected="selectedProduct === product.id" :aria-controls="product.id" :tabindex="selectedProduct === product.id ? 0 : -1" :class="'mofa-launch-' + product.id" @click="selectedProduct = product.id" @keydown="moveTab($event, index)">
            <FamilyProductMark :product="product.id" />
            <span><strong>{{ product.name }}</strong><small>{{ launcherLabels[product.id] }}</small></span>
            <Icon name="arrowRight" size="sm" class="mofa-launch-arrow" aria-hidden="true" />
          </button>
        </div>
        <div class="mofa-family-stage">
          <article v-for="product in products" v-show="selectedProduct === product.id" :id="product.id" :key="product.id" class="mofa-family-product" :class="'mofa-family-' + product.id" role="tabpanel" :aria-labelledby="'family-tab-' + product.id" tabindex="0">
            <div class="mofa-family-product-info">
              <div class="mofa-family-product-heading"><h3>{{ product.name }}</h3><span>{{ product.state }}</span></div>
              <p class="mofa-family-purpose">{{ product.purpose }}</p>
              <p>{{ product.description }}</p>
              <div class="mofa-family-product-action">
                <FamilyProductAction :product="product" />
                <div v-if="product.id === 'synaroute'" class="mofa-synaroute-links">
                  <a :href="SYNA_ROUTE_DOWNLOAD_URL" class="mofa-synaroute-download" target="_blank" rel="noopener noreferrer" aria-label="下载 SynaRoute 客户端（新窗口）" data-testid="synaroute-download"><Icon name="download" size="sm" aria-hidden="true" />下载客户端</a>
                  <a :href="SYNA_ROUTE_WEBSITE_URL" target="_blank" rel="noopener noreferrer" aria-label="访问 SynaRoute 官网（新窗口）" data-testid="synaroute-website">访问官网<Icon name="externalLink" size="sm" aria-hidden="true" /></a>
                </div>
              </div>
              <p class="mofa-family-account-note">{{ product.account }}</p>
            </div>
            <FamilyProductVisual :product="product.id" />
          </article>
        </div>
      </div>
    </section>
    <section class="mofa-family-session" aria-labelledby="family-session-title">
      <div><h2 id="family-session-title">一个控制台，随时出发。</h2><p>登录后进入魔法 API 控制台，从顶部打开家族工具。</p></div>
      <details><summary>登录与模型授权，如何协作？<Icon name="chevronDown" size="sm" aria-hidden="true" /></summary><p>本站控制台与配方配置选择共用当前会话。配方连接模型时，仍由你选择已有密钥并明确授权；不会自动创建密钥或产生模型费用。</p><p>SynaRoute 使用密钥导入，不是桌面账号单点登录。工坊首版是公开创作资源，不需要再注册账号；独立产品尚未实现统一退出。</p><RouterLink to="/guide">查看模型接入教程 <Icon name="arrowRight" size="sm" aria-hidden="true" /></RouterLink></details>
    </section>
  </PublicLayout>
</template>

<script setup lang="ts">
import { computed, ref, type ComponentPublicInstance } from 'vue'
import { useAppStore, useAuthStore } from '@/stores'
import PublicLayout from '@/custom/components/PublicLayout.vue'
import BrandPanel from '@/custom/components/BrandPanel.vue'
import FamilyProductVisual from '@/custom/components/FamilyProductVisual.vue'
import FamilyProductAction from '@/custom/components/FamilyProductAction.vue'
import FamilyProductMark from '@/custom/components/FamilyProductMark.vue'
import Icon from '@/components/icons/Icon.vue'
import { familyProducts, recipeDestination, studioDestination } from '@/custom/family/products'
import { SYNA_ROUTE_DOWNLOAD_URL, SYNA_ROUTE_WEBSITE_URL } from '@/custom/family/synarouteLinks'

const app = useAppStore()
const auth = useAuthStore()
const origin = window.location.origin
const recipe = recipeDestination(String(import.meta.env.VITE_MAGIC_RECIPES_URL || ''), origin, import.meta.env.DEV)
const studio = studioDestination(String(import.meta.env.VITE_MAGIC_STUDIO_URL || ''), origin, import.meta.env.DEV)
const products = computed(() => familyProducts(recipe, auth.isAdmin, studio))
const initialProduct = window.location.hash.slice(1)
const selectedProduct = ref(products.value.some(product => product.id === initialProduct) ? initialProduct : 'recipes')
const launcherLabels: Record<string, string> = { api: '接入模型', recipes: '复用方法', studio: '寻找灵感', synaroute: '本地协同' }
const tabElements: (HTMLButtonElement | undefined)[] = []

function setTab(element: Element | ComponentPublicInstance | null, index: number) {
  tabElements[index] = element instanceof HTMLButtonElement ? element : undefined
}

function moveTab(event: KeyboardEvent, index: number) {
  const lastIndex = products.value.length - 1
  const nextIndex = event.key === 'ArrowRight' ? (index + 1) % products.value.length
    : event.key === 'ArrowLeft' ? (index + lastIndex) % products.value.length
      : event.key === 'Home' ? 0 : event.key === 'End' ? lastIndex : undefined
  if (nextIndex === undefined) return
  event.preventDefault()
  selectedProduct.value = products.value[nextIndex].id
  tabElements[nextIndex]?.focus()
}
</script>
