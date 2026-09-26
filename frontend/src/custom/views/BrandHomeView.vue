<template>
  <div class="mofa-brand-home">
    <header class="mofa-brand-mast">
      <router-link to="/brand" class="mofa-brand-lockup">
        <img :src="siteLogo" :alt="siteName" width="42" height="42">
        <span>{{ siteName }}<small>MODEL ACCESS / WORKSPACE</small></span>
      </router-link>
      <nav class="mofa-brand-nav" :aria-label="copy.home">
        <BrandThemeToggle />
        <router-link :to="authStore.isAuthenticated ? '/dashboard' : '/login'" class="btn btn-secondary">
          {{ authStore.isAuthenticated ? copy.workspace : copy.login }}
        </router-link>
      </nav>
    </header>
    <main class="mofa-brand-main">
      <BrandPanel :site-name="siteName">
        <template #actions>
          <router-link to="/dashboard" class="btn btn-primary">{{ copy.workspace }}<Icon name="arrowRight" size="sm" /></router-link>
          <router-link to="/keys" class="btn btn-secondary">{{ copy.accessAction }}</router-link>
        </template>
      </BrandPanel>
      <section class="mofa-products" :aria-label="copy.sectionTitle">
        <div class="mofa-products-heading"><h2>{{ copy.sectionTitle }}</h2><p>{{ copy.sectionNote }}</p></div>
        <div class="mofa-product-grid">
          <article class="mofa-offering">
            <span class="mofa-offering-number">01 / ACCESS</span>
            <h3>{{ copy.accessTitle }}</h3>
            <p>{{ copy.accessDescription }}</p>
            <router-link to="/keys">{{ copy.accessAction }}<Icon name="arrowRight" size="sm" /></router-link>
          </article>
          <article class="mofa-offering">
            <span class="mofa-offering-number">02 / USAGE</span>
            <h3>{{ copy.usageTitle }}</h3>
            <p>{{ copy.usageDescription }}</p>
            <router-link to="/usage">{{ copy.usageAction }}<Icon name="arrowRight" size="sm" /></router-link>
          </article>
        </div>
      </section>
    </main>
    <footer class="mofa-brand-footer"><span>© {{ new Date().getFullYear() }} {{ siteName }}</span><span>{{ copy.footer }}</span></footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useAppStore } from '@/stores'
import { useAuthStore } from '@/stores/auth'
import { sanitizeUrl } from '@/utils/url'
import Icon from '@/components/icons/Icon.vue'
import BrandPanel from '@/custom/components/BrandPanel.vue'
import BrandThemeToggle from '@/custom/components/BrandThemeToggle.vue'
import { useBrandCopy } from '@/custom/brand/copy'

const appStore = useAppStore()
const authStore = useAuthStore()
const copy = useBrandCopy()
const siteName = computed(() => appStore.siteName || 'Sub2API')
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }) || '/logo.svg')

onMounted(() => appStore.fetchPublicSettings())
</script>
