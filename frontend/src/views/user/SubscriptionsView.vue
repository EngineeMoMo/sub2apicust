<template>
  <AppLayout>
    <!-- [CUSTOM] 订阅展示复用原列表与续费路由；时间、卡片样式在 custom 层维护。 -->
    <div class="mofa-subscriptions">
      <div v-if="loading" class="flex justify-center py-12" role="status" :aria-label="t('common.loading')">
        <div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
      </div>

      <div v-else-if="loadFailed" class="card p-12 text-center">
        <p role="alert">{{ t('userSubscriptions.failedToLoad') }}</p>
        <button type="button" class="btn btn-primary mt-4" @click="loadSubscriptions">{{ zh ? '重试' : 'Retry' }}</button>
      </div>

      <div v-else-if="subscriptions.length === 0" class="card p-12 text-center">
        <div class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-gray-100 dark:bg-dark-700">
          <Icon name="creditCard" size="xl" class="text-gray-400" />
        </div>
        <h3 class="mb-2 text-lg font-semibold text-gray-900 dark:text-white">{{ t('userSubscriptions.noActiveSubscriptions') }}</h3>
        <p class="text-gray-500 dark:text-dark-400">{{ t('userSubscriptions.noActiveSubscriptionsDesc') }}</p>
      </div>

      <template v-else>
        <!-- [CUSTOM] 库存失败不伪装不限量；保留已有权益展示，禁止发起续订。 -->
        <div v-if="stockError" class="mofa-subscription-notice" role="status">
          <p>{{ t('payment.stock.loadFailed') }}</p>
          <button type="button" class="btn btn-secondary" @click="loadSubscriptions">{{ zh ? '重试' : 'Retry' }}</button>
        </div>
        <div v-if="timingUnavailable" class="mofa-subscription-notice" role="status">
          <p>{{ zh ? '重置时间暂未读取，订阅和用量仍正常显示。' : 'Reset times could not be loaded. Subscriptions and usage are still shown.' }}</p>
          <button type="button" class="btn btn-secondary" @click="loadSubscriptions">{{ zh ? '重试' : 'Retry' }}</button>
        </div>
        <div class="mofa-subscriptions-grid">
          <SubscriptionCard v-for="subscription in subscriptions" :key="subscription.id"
            :subscription="subscription" :now="now" :resets="resetTimes[subscription.id]"
            :plans="stockPlans.filter(plan => plan.group_id === subscription.group_id)" :stock-loading="stockLoading" :stock-error="stockError"
            :utc-offset="appStore.cachedPublicSettings?.server_utc_offset" />
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import subscriptionsAPI from '@/api/subscriptions'
// [CUSTOM] 使用真实商品库存；不从已有订阅数量推算剩余。
import { paymentAPI } from '@/api/payment'
import type { SubscriptionPlan } from '@/types/payment'
import type { UserSubscription } from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
// [CUSTOM] 后端权威 resets_at 与只更新显示的共享时钟，不改额度或续期。
import SubscriptionCard from '@/custom/components/SubscriptionCard.vue'
import { fetchSubscriptionResetTimes, useSubscriptionClock, type SubscriptionResetTimes } from '@/custom/subscriptions/timing'

const { t, locale } = useI18n()
const zh = computed(() => locale.value.startsWith('zh'))
const appStore = useAppStore()
const subscriptions = ref<UserSubscription[]>([])
const resetTimes = ref<Record<number, SubscriptionResetTimes>>({})
const loading = ref(true)
const loadFailed = ref(false)
const timingUnavailable = ref(false)
const stockPlans = ref<SubscriptionPlan[]>([])
const stockLoading = ref(true)
const stockError = ref(false)
const now = useSubscriptionClock()
let loadVersion = 0

async function loadSubscriptions() {
  const version = ++loadVersion
  loading.value = true
  loadFailed.value = false
  timingUnavailable.value = false
  resetTimes.value = {}
  // [CUSTOM] 库存请求独立，旧响应与卸载后的响应不得覆盖当前页面。
  stockLoading.value = true
  stockError.value = false
  void paymentAPI.getPlans().then(response => {
    if (!Array.isArray(response.data)) throw new Error('Invalid subscription stock response')
    if (version === loadVersion) stockPlans.value = response.data
  }).catch(() => {
    if (version === loadVersion) { stockPlans.value = []; stockError.value = true }
  }).finally(() => { if (version === loadVersion) stockLoading.value = false })
  // [CUSTOM] 辅助时间请求不阻塞列表；重试或卸载后丢弃旧请求结果。
  void fetchSubscriptionResetTimes().then(times => {
    if (version === loadVersion) resetTimes.value = times
  }, () => {
    if (version === loadVersion) timingUnavailable.value = true
  })
  try {
    const list = await subscriptionsAPI.getMySubscriptions()
    if (version === loadVersion) subscriptions.value = list
  } catch {
    if (version === loadVersion) {
      loadFailed.value = true
      appStore.showError(t('userSubscriptions.failedToLoad'))
    }
  } finally {
    if (version === loadVersion) loading.value = false
  }
}

onMounted(loadSubscriptions)
onUnmounted(() => { loadVersion++ })
</script>
