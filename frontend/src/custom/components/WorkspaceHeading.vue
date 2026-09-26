<template>
  <section class="mofa-workspace-heading" aria-labelledby="mofa-page-title">
    <div class="mofa-page-heading">
      <h1 id="mofa-page-title">{{ pageTitle }}</h1>
      <p v-if="pageDescription">{{ pageDescription }}</p>
    </div>
    <div v-if="$slots.actions" class="mofa-page-actions">
      <slot name="actions" />
    </div>
  </section>
</template>

<script setup lang="ts">
import { useWorkspaceHeading } from '@/custom/brand/useWorkspaceHeading'

const { pageTitle, pageDescription } = useWorkspaceHeading()
</script>
