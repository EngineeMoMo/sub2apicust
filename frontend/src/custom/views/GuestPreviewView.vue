<template>
  <div class="mofa-public mofa-guest-console">
    <a class="mofa-skip-link" href="#preview-main">跳到正文</a>
    <header class="mofa-guest-header">
      <router-link to="/home" class="mofa-brand-lockup">
        <img :src="siteLogo" :alt="app.siteName" width="36" height="36">
        <span>{{ app.siteName }}<small>多模型 API 服务</small></span>
      </router-link>
      <div class="mofa-guest-header-actions">
        <span class="mofa-guest-badge">游客预览</span>
        <BrandThemeToggle />
        <router-link to="/home" class="mofa-guest-home-link">返回官网</router-link>
        <GuestAction to="/dashboard" button-class="btn btn-primary">{{ auth.isAuthenticated ? '进入我的控制台' : '登录使用' }}</GuestAction>
      </div>
    </header>

    <div class="mofa-guest-shell">
      <aside class="mofa-guest-sidebar" aria-label="预览功能导航">
        <button class="btn btn-secondary mofa-guest-menu-toggle" type="button" :aria-expanded="menuOpen" aria-controls="preview-navigation" @click="menuOpen = !menuOpen">
          <Icon name="menu" size="sm" aria-hidden="true" />{{ menuOpen ? '收起菜单' : `功能菜单 · ${section.label}` }}
        </button>
        <nav id="preview-navigation" :class="{ 'is-open': menuOpen }" aria-label="游客控制台菜单">
          <p class="mofa-guest-nav-label">工作空间</p>
          <router-link v-for="item in previewSections" :key="item.id" :to="previewPath(item.id)" :aria-current="section.id === item.id ? 'page' : undefined" :class="['mofa-guest-nav-link', { 'is-active': section.id === item.id }]">
            <Icon :name="item.icon" size="sm" aria-hidden="true" /><span>{{ item.label }}</span>
          </router-link>
          <p class="mofa-guest-sidebar-note">只读预览不创建账户、不调用模型，也不会产生订单。</p>
        </nav>
      </aside>

      <main id="preview-main" ref="mainElement" class="mofa-guest-content" tabindex="-1">
        <div class="mofa-guest-notice"><Icon name="eye" size="sm" aria-hidden="true" /><p><strong>你正在浏览控制台预览。</strong>个人数据不会在这里加载；套餐与常见问题可以直接查看。</p></div>

        <PublicPlansView v-if="section.id === 'plans'" embedded />
        <PublicFaqView v-else-if="section.id === 'faq'" embedded guide-target="/preview/guide" />
        <PublicGuideView v-else-if="section.id === 'guide'" embedded scope="preview" />
        <template v-else-if="section.id === 'overview'">
          <div class="mofa-guest-heading"><h1>先看看你的工作空间</h1><p>从获取密钥到管理用量，左侧菜单带你了解每个功能。无需注册，也不展示任何人的账户数据。</p></div>
          <section class="mofa-guest-summary" aria-label="账户信息预览">
            <div v-for="label in ['账户余额', 'API 密钥', '已用额度']" :key="label"><span>{{ label }}</span><strong>登录后查看</strong><small>预览不读取账户信息</small></div>
          </section>
          <section class="mofa-guest-guide" aria-labelledby="preview-start-title">
            <h2 id="preview-start-title">从了解服务到开始使用</h2>
            <ol>
              <li><div><h3>了解套餐与接入方式</h3><p>先看在售套餐与常见问题，确认权益和使用方式是否适合你。</p></div><router-link to="/preview/plans">查看套餐<Icon name="arrowRight" size="sm" aria-hidden="true" /></router-link></li>
              <li><div><h3>在账户中管理凭证</h3><p>登录后创建 API Key，并把它配置到兼容的工具中。预览不生成密钥。</p></div><router-link to="/preview/keys">预览密钥管理<Icon name="arrowRight" size="sm" aria-hidden="true" /></router-link></li>
              <li><div><h3>跟踪自己的使用情况</h3><p>在用量、订阅和订单页面查看账户信息，不需要在预览阶段填写付款资料。</p></div><router-link to="/preview/usage">预览用量记录<Icon name="arrowRight" size="sm" aria-hidden="true" /></router-link></li>
            </ol>
          </section>
        </template>
        <template v-else>
          <div class="mofa-guest-heading"><h1>{{ section.label }}</h1><p>{{ section.description }}</p></div>
          <section class="mofa-guest-private" :aria-label="`${section.label}只读预览`">
            <p class="mofa-guest-fields-caption">登录后可查看的信息</p>
            <ul class="mofa-guest-fields"><li v-for="column in section.columns" :key="column">{{ column }}</li></ul>
            <div class="mofa-guest-locked">
              <Icon name="lock" size="lg" aria-hidden="true" />
              <h2>登录后查看你的{{ section.label === 'API 密钥' ? '密钥' : section.label.replace('我的', '') }}</h2>
              <p>这里是功能预览，不是你的账户数据。你可以继续浏览其他菜单，准备好使用时再登录。</p>
              <GuestAction :key="section.id" :to="section.target" button-class="btn btn-primary" :message="`${section.action}需要关联你的账户。登录后继续，或取消并留在预览中。`">{{ section.action }}</GuestAction>
            </div>
          </section>
          <p class="mofa-public-note">还没决定如何使用？<router-link to="/preview/plans">比较订阅套餐</router-link>，或<router-link to="/preview/faq">阅读常见问题</router-link>。</p>
        </template>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useAppStore, useAuthStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'
import Icon from '@/components/icons/Icon.vue'
import BrandThemeToggle from '@/custom/components/BrandThemeToggle.vue'
import GuestAction from '@/custom/components/GuestAction.vue'
import PublicPlansView from '@/custom/views/PublicPlansView.vue'
import PublicFaqView from '@/custom/views/PublicFaqView.vue'
import PublicGuideView from '@/custom/views/PublicGuideView.vue'
import { previewPath, previewSections } from '@/custom/guest/preview'

const app = useAppStore()
const auth = useAuthStore()
const route = useRoute()
const section = computed(() => previewSections.find(item => item.id === route.params.section) ?? previewSections[0]!)
const siteLogo = computed(() => sanitizeUrl(app.siteLogo || '', { allowRelative: true, allowDataUrl: true }) || '/logo.svg')
const menuOpen = ref(false)
const mainElement = ref<HTMLElement | null>(null)
watch(() => route.params.section, async () => {
  menuOpen.value = false
  await nextTick()
  mainElement.value?.focus()
})
onMounted(() => { void app.fetchPublicSettings().catch(() => {}) })
</script>
