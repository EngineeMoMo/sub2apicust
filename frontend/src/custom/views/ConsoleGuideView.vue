<template>
  <AppLayout>
    <div class="mofa-public"><PublicGuideView embedded hide-title scope="console" /></div>
  </AppLayout>
</template>

<script setup lang="ts">
import AppLayout from '@/components/layout/AppLayout.vue'
import PublicGuideView from '@/custom/views/PublicGuideView.vue'
</script>
