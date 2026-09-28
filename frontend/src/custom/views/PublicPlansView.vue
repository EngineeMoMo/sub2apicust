<template>
  <component :is="embedded ? 'div' : PublicLayout">
    <section class="mofa-public-intro">
      <h1>找到适合你的订阅套餐</h1>
      <p class="mofa-public-lead">先比较价格、有效期与权益。浏览无需登录，决定购买时再登录账户。</p>
      <p>这里展示后台当前在售套餐。标价不是最终支付报价，币种换算、手续费及实付金额以结账页为准。</p>
    </section>
    <div v-if="loading" class="mofa-public-state" role="status">正在读取在售套餐…</div>
    <div v-else-if="error" class="mofa-public-state" role="alert">
      <h2>{{ unavailable ? '订阅套餐暂未开放' : '暂时无法读取套餐' }}</h2>
      <p>{{ unavailable ? '你仍可浏览产品介绍和常见问题。' : '请稍后重试；无需为了查看套餐而登录。' }}</p>
      <button type="button" class="btn btn-secondary" @click="load">重新加载</button>
    </div>
    <template v-else>
      <p v-if="!catalog.purchase_enabled && catalog.plans.length" class="mofa-public-state" role="status">在线支付暂未开放，当前仅供浏览，暂不能购买。</p>
      <div v-if="!catalog.plans.length" class="mofa-public-state"><h2>暂无在售套餐</h2><p>套餐上架后会在这里展示。你可以先查看常见问题，了解接入方式。</p></div>
      <div v-else class="mofa-plan-grid">
        <article v-for="plan in catalog.plans" :key="plan.id" class="mofa-public-plan">
          <h2>{{ plan.name }}</h2>
          <p>{{ plan.description }}</p>
          <p class="mofa-plan-price">{{ currencySymbol(plan.currency || 'USD') }}{{ plan.price }} <small>{{ plan.currency || 'USD' }}</small></p>
          <p v-if="plan.original_price && plan.original_price > plan.price"><s>{{ currencySymbol(plan.currency || 'USD') }}{{ plan.original_price }}</s></p>
          <p>有效期：{{ planValiditySuffix(plan, t) }}</p>
          <dl class="mofa-plan-limits">
            <template v-for="limit in limits" :key="limit.key">
              <template v-if="plan[limit.key] != null"><dt>{{ limit.label }}</dt><dd>{{ plan[limit.key] }} USD</dd></template>
            </template>
          </dl>
          <p v-if="limits.every(limit => plan[limit.key] == null)" class="mofa-public-muted">未提供额度说明，请在购买前确认套餐权益。</p>
          <p v-if="plan.supported_model_scopes.length">模型范围：{{ plan.supported_model_scopes.join('、') }}</p>
          <ul v-if="plan.features.length"><li v-for="(feature, index) in plan.features" :key="index">{{ feature }}</li></ul>
          <GuestAction v-if="catalog.purchase_enabled" :to="`/purchase?tab=subscription&plan=${plan.id}`" button-class="btn btn-primary" message="登录后会返回此套餐的购买页面。请确认权益和实付金额后再下单。">选择此套餐</GuestAction>
          <button v-else type="button" class="btn btn-secondary" disabled>暂不可购买</button>
        </article>
      </div>
    </template>
    <p class="mofa-public-note">不确定订阅和充值有什么区别？<router-link :to="embedded ? '/preview/faq' : '/faq'">查看常见问题</router-link></p>
  </component>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import axios from 'axios'
import PublicLayout from '@/custom/components/PublicLayout.vue'
import GuestAction from '@/custom/components/GuestAction.vue'
import { fetchPublicPlans, type PublicCatalog } from '@/custom/guest/api'
import { currencySymbol } from '@/components/payment/currency'
import { planValiditySuffix } from '@/components/payment/validity'

const { t } = useI18n()
defineProps<{ embedded?: boolean }>()
const loading = ref(true)
const error = ref(false)
const unavailable = ref(false)
const catalog = ref<PublicCatalog>({ plans: [], purchase_enabled: false })
const limits = [
  { key: 'daily_limit_usd', label: '日额度' },
  { key: 'weekly_limit_usd', label: '周额度' },
  { key: 'monthly_limit_usd', label: '月额度' }
] as const
async function load() {
  loading.value = true
  error.value = false
  try { catalog.value = await fetchPublicPlans() }
  catch (cause) {
    error.value = true
    unavailable.value = axios.isAxiosError(cause) && cause.response?.status === 404
  } finally { loading.value = false }
}
onMounted(load)
</script>
