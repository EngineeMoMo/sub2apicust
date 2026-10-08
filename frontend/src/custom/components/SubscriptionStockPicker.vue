<script setup lang="ts">
import { computed, ref, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminPaymentAPI } from '@/api/admin/payment'
import type { SubscriptionPlan } from '@/types/payment'

const props = defineProps<{ groupId: number | null; modelValue: number | null; inputId?: string }>()
const emit = defineEmits<{ 'update:modelValue': [number | null]; available: [boolean] }>()
const { t } = useI18n()
const plans = ref<SubscriptionPlan[]>([])
const loading = ref(true)
const failed = ref(false)
const options = computed(() => plans.value.filter(plan => plan.group_id === props.groupId))
const selected = computed(() => options.value.find(plan => plan.id === props.modelValue))
const available = computed(() => !loading.value && !failed.value && selected.value?.stock_remaining != null && selected.value.stock_remaining !== 0)
watch(available, value => emit('available', value), { immediate: true })
watch(() => props.groupId, () => { emit('update:modelValue', null); emit('available', false) })
async function load() {
  loading.value = true; failed.value = false
  try {
    const data = (await adminPaymentAPI.getPlans()).data
    if (!Array.isArray(data)) throw new Error('Invalid subscription stock response')
    plans.value = data
  }
  catch { failed.value = true; plans.value = [] }
  finally { loading.value = false }
}
function select(event: Event) {
  const value = (event.target as HTMLSelectElement).value
  emit('update:modelValue', value ? Number(value) : null)
}
function stock(plan: SubscriptionPlan): string {
  if (plan.stock_remaining == null) return t('payment.stock.loadFailed')
  return plan.stock_remaining < 0 ? t('payment.stock.unlimited') : plan.stock_remaining === 0 ? t('payment.stock.soldOut') : t('payment.stock.remaining', { count: plan.stock_remaining })
}
onMounted(load)
</script>

<template>
  <div class="mofa-stock-picker">
    <label :for="inputId || 'subscription-stock-plan'" class="input-label">{{ t('payment.stock.assignmentPlan') }}</label>
    <select :id="inputId || 'subscription-stock-plan'" class="input" :value="modelValue ?? ''" :disabled="loading || failed || !groupId" required @change="select">
      <option value="">{{ t('payment.stock.selectPlan') }}</option>
      <option v-for="plan in options" :key="plan.id" :value="plan.id" :disabled="plan.stock_remaining == null || plan.stock_remaining === 0">{{ plan.name }} · {{ stock(plan) }}</option>
    </select>
    <p class="mofa-stock-help">{{ t('payment.stock.assignmentHint') }}</p>
    <p v-if="loading" role="status">{{ t('common.loading') }}</p>
    <p v-else-if="failed" role="alert">{{ t('payment.stock.loadFailed') }}</p>
    <p v-else-if="groupId && !options.length" role="status">{{ t('payment.stock.createPlanFirst') }}</p>
    <p v-else-if="selected" role="status">{{ selected.name }} · {{ stock(selected) }}</p>
    <button type="button" class="btn btn-secondary" :disabled="loading" @click="load">{{ t('common.refresh') }}</button>
  </div>
</template>
