<template>
  <section class="mofa-brand-panel" :aria-label="siteName">
    <div class="mofa-brand-copy">
      <p class="mofa-eyebrow">MODEL ACCESS / YOUR WORKSPACE</p>
      <h2 class="mofa-hero-title">{{ copy.headline }}<br><span>{{ copy.highlight }}</span></h2>
      <p class="mofa-hero-description">{{ copy.introduction }}</p>
      <div v-if="$slots.actions" class="mofa-hero-actions"><slot name="actions" /></div>
    </div>
    <div class="mofa-brand-study" aria-hidden="true">
      <div class="mofa-study-heading"><span>{{ siteName }}</span><span>M — 01</span></div>
      <span class="mofa-core-mark" :style="{ '--mofa-logo-mask': 'url(' + markUrl + ')' }"></span>
      <div class="mofa-study-footer"><span>{{ copy.access }}</span><span>＋</span><span>{{ copy.control }}</span></div>
    </div>
    <p class="mofa-brand-caption">{{ copy.identity }}</p>
  </section>
</template>

<script setup lang="ts">
import markUrl from '@/custom/assets/mofa-mark.webp'
import { useBrandCopy } from '@/custom/brand/copy'

defineProps<{ siteName: string }>()
const copy = useBrandCopy()
</script>
