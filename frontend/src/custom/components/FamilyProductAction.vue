<template>
  <button v-if="!product.destination" type="button" disabled :title="product.state">发布后开放</button>
  <RouterLink v-else-if="!auth.isAuthenticated" :to="familyLoginTarget(product)" class="btn btn-secondary">登录后{{ product.action }}<Icon name="arrowRight" size="sm" aria-hidden="true" /></RouterLink>
  <a v-else-if="product.external" :href="product.destination" target="_blank" rel="noopener noreferrer" class="btn btn-primary" :aria-label="product.action + '（新窗口）'">{{ product.action }}<Icon name="externalLink" size="sm" aria-hidden="true" /></a>
  <RouterLink v-else :to="product.destination" class="btn btn-primary">{{ product.action }}<Icon name="arrowRight" size="sm" aria-hidden="true" /></RouterLink>
</template>

<script setup lang="ts">
import { useAuthStore } from '@/stores'
import Icon from '@/components/icons/Icon.vue'
import { familyLoginTarget, type FamilyProduct } from '@/custom/family/products'

defineProps<{ product: FamilyProduct }>()
const auth = useAuthStore()
</script>
