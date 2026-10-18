<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 class="text-2xl font-semibold">订单核账</h1>
          <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">集中复查收款占用、处理异常订单与登记线下收款。</p>
        </div>
        <RouterLink class="btn btn-secondary" to="/admin/orders">查看全部订单</RouterLink>
      </div>
      <PaymentCollectionOperations />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { RouterLink } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import PaymentCollectionOperations from '@/custom/components/PaymentCollectionOperations.vue'
</script>
