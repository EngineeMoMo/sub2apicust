<template>
  <span class="mofa-product-mark" :class="'mofa-mark-' + product" aria-hidden="true">
    <span v-if="product === 'api'" class="mofa-product-mark-mask" :style="{ '--mofa-node-mask': 'url(' + mark + ')' }"></span>
    <svg v-else-if="product === 'recipes'" viewBox="0 0 32 32" fill="none"><path d="M9 7V4h15l4 4v18H13v-3M24 4v5h4M4 9h15l4 4v15H4V9ZM19 9v5h4M8 18h9M8 22h6" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" /></svg>
    <svg v-else-if="product === 'studio'" viewBox="0 0 32 32" fill="none"><path d="m16 3 3.5 9.5L29 16l-9.5 3.5L16 29l-3.5-9.5L3 16l9.5-3.5L16 3Z" stroke="currentColor" stroke-width="1.6" stroke-linejoin="round" /><path d="m24 4 4 4M24 28l4-4" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" /></svg>
    <svg v-else viewBox="0 0 32 32" fill="none"><path d="M4 16h8c6 0 6-9 11-9h5M23 3l5 4-5 4M12 16c6 0 6 9 11 9h5M23 21l5 4-5 4" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" /><circle cx="7" cy="16" r="3" fill="currentColor" /></svg>
  </span>
</template>

<script setup lang="ts">
import mark from '@/custom/assets/mofa-mark.webp'

defineProps<{ product: string }>()
</script>
