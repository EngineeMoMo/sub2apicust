<template>
  <div class="mofa-brand-home mofa-public">
    <a class="mofa-skip-link" href="#public-main">跳到正文</a>
    <header class="mofa-brand-mast">
      <router-link to="/home" class="mofa-brand-lockup">
        <img :src="siteLogo" :alt="app.siteName" width="42" height="42">
        <span>{{ app.siteName }}<small>多模型 API 服务</small></span>
      </router-link>
      <nav class="mofa-brand-nav" aria-label="官网导航">
        <router-link to="/home">产品介绍</router-link>
        <router-link to="/plans">订阅套餐</router-link>
        <router-link to="/faq">常见问题</router-link>
        <router-link v-if="!auth.isAuthenticated" to="/preview">游客预览</router-link>
        <router-link v-if="app.cachedPublicSettings?.model_plaza_enabled && (auth.isAuthenticated || !app.cachedPublicSettings?.model_plaza_require_auth)" to="/model-plaza">模型广场</router-link>
        <BrandThemeToggle />
        <router-link v-if="auth.isAuthenticated" :to="auth.isAdmin ? '/admin/dashboard' : '/dashboard'" class="btn btn-primary">进入控制台</router-link>
        <template v-else>
          <router-link to="/login" class="btn btn-secondary">登录</router-link>
          <router-link v-if="app.cachedPublicSettings?.registration_enabled" to="/register" class="btn btn-primary">注册</router-link>
        </template>
      </nav>
    </header>
    <main id="public-main" class="mofa-public-main" tabindex="-1"><slot /></main>
    <footer class="mofa-brand-footer">
      <span>© {{ new Date().getFullYear() }} {{ app.siteName }}</span>
      <router-link to="/faq">使用前有疑问？查看常见问题</router-link>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useAppStore, useAuthStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'
import BrandThemeToggle from '@/custom/components/BrandThemeToggle.vue'

const app = useAppStore()
const auth = useAuthStore()
const siteLogo = computed(() => sanitizeUrl(app.siteLogo || '', { allowRelative: true, allowDataUrl: true }) || '/logo.svg')
onMounted(() => { void app.fetchPublicSettings().catch(() => {}) })
</script>
