<template>
  <button
    type="button"
    class="mofa-theme-toggle"
    :aria-label="isDark ? copy.light : copy.dark"
    :title="isDark ? copy.light : copy.dark"
    @click="toggleTheme"
  >
    <Icon :name="isDark ? 'sun' : 'moon'" size="md" />
  </button>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import Icon from '@/components/icons/Icon.vue'
import { useBrandCopy } from '@/custom/brand/copy'

const copy = useBrandCopy()
const isDark = ref(document.documentElement.classList.contains('dark'))

function toggleTheme() {
  isDark.value = !document.documentElement.classList.contains('dark')
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}
</script>
