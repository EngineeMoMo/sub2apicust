<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import type { UserSubscription } from '@/types'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTimeToMinute } from '@/utils/format'
import { platformBadgeClass, platformLabel } from '@/utils/platformColors'
import { hasPeakRate, formatPeakRateWindow, serverTimezoneLabel } from '@/utils/peak-rate'
import { effectiveSubscriptionStatus, quotaPeriods, remainingTimeLabel, subscriptionResetHint,
  type QuotaPeriod, type SubscriptionResetTimes } from '@/custom/subscriptions/timing'

const props = defineProps<{ subscription: UserSubscription; now: number; resets?: SubscriptionResetTimes; utcOffset?: string }>()
const { t, locale } = useI18n()
const router = useRouter()
const zh = computed(() => locale.value.startsWith('zh'))
const status = computed(() => effectiveSubscriptionStatus(props.subscription, props.now))
const expiresAt = computed(() => props.subscription.expires_at ? Date.parse(props.subscription.expires_at) : NaN)
const urgency = computed(() => {
  if (status.value !== 'active' || !Number.isFinite(expiresAt.value)) return 'normal'
  const remaining = expiresAt.value - props.now
  return remaining <= 3 * 86400_000 ? 'urgent' : remaining <= 7 * 86400_000 ? 'soon' : 'normal'
})
const statusLabel = computed(() => status.value === 'suspended' ? (zh.value ? '已暂停' : 'Suspended') : t(`userSubscriptions.status.${status.value}`))
const quotas = computed(() => quotaPeriods.flatMap(period => {
  const limit = props.subscription.group?.[`${period}_limit_usd`]
  if (!limit || limit <= 0) return []
  const rawUsed = props.subscription[`${period}_usage_usd`]
  const used = Number.isFinite(rawUsed) ? Math.max(0, rawUsed) : 0
  const percent = used / limit * 100
  return [{ period, limit, used, tone: percent >= 90 ? 'high' : percent >= 70 ? 'medium' : 'normal',
    hint: subscriptionResetHint(props.subscription, period, props.resets, props.now) }]
}))

function dateLabel(at: string): string {
  return Number.isFinite(Date.parse(at)) ? formatDateTimeToMinute(at, locale.value) : (zh.value ? '时间暂不可用' : 'Time unavailable')
}

function resetLabel(period: QuotaPeriod): string {
  const hint = subscriptionResetHint(props.subscription, period, props.resets, props.now)
  switch (hint.kind) {
    case 'countdown': return t('userSubscriptions.resetIn', { time: remainingTimeLabel(hint.at!, props.now, zh.value) })
    case 'pending': return zh.value ? '已到重置时间，使用后更新' : 'Reset time reached; updates on use'
    case 'ends': return t('userSubscriptions.quotaEndsIn', { time: remainingTimeLabel(hint.at!, props.now, zh.value) })
    case 'unstarted': return t('userSubscriptions.windowNotActive')
    case 'unavailable': return zh.value ? '重置时间暂未读取' : 'Reset time unavailable'
    default: return statusLabel.value
  }
}

function renew() {
  router.push({ path: '/purchase', query: { tab: 'subscription', group: String(props.subscription.group_id) } })
}
</script>

<template>
  <article class="mofa-subscription-card" :data-state="status" :data-subscription-id="subscription.id">
    <header class="mofa-subscription-header">
      <div class="mofa-subscription-identity">
        <div class="mofa-subscription-title">
          <h3>{{ subscription.group?.name || `Group #${subscription.group_id}` }}</h3>
          <span :class="['mofa-subscription-platform', platformBadgeClass(subscription.group?.platform || '')]">
            {{ platformLabel(subscription.group?.platform || '') }}
          </span>
        </div>
        <p v-if="subscription.group?.description" class="mofa-subscription-description">{{ subscription.group.description }}</p>
        <div class="mofa-subscription-rates">
          <span>{{ t('payment.planCard.rate') }}: ×{{ subscription.group?.rate_multiplier ?? 1 }}</span>
          <span v-if="hasPeakRate(subscription.group)">{{ t('payment.planCard.peakRate') }}: {{ formatPeakRateWindow(subscription.group, serverTimezoneLabel(utcOffset)) }}</span>
        </div>
      </div>
      <div class="mofa-subscription-actions">
        <span class="mofa-subscription-status" :data-state="status">{{ statusLabel }}</span>
        <button v-if="status === 'active' || status === 'expired'" type="button" class="btn btn-primary mofa-subscription-renew" @click="renew">
          {{ t('payment.renewNow') }}
          <Icon name="arrowRight" size="sm" aria-hidden="true" />
        </button>
      </div>
    </header>

    <div class="mofa-subscription-body">
      <dl class="mofa-subscription-expiry" :data-urgency="urgency">
        <dt><Icon name="clock" size="sm" aria-hidden="true" />{{ t('userSubscriptions.expires') }}</dt>
        <dd>
          <time v-if="subscription.expires_at" :datetime="subscription.expires_at">{{ dateLabel(subscription.expires_at) }}</time>
          <span v-else>{{ t('userSubscriptions.noExpiration') }}</span>
          <span v-if="status === 'active' && Number.isFinite(expiresAt)" class="mofa-subscription-remaining">
            {{ zh ? '剩余' : 'Remaining' }} {{ remainingTimeLabel(subscription.expires_at!, now, zh) }}
          </span>
        </dd>
      </dl>

      <div v-if="quotas.length" class="mofa-subscription-quotas">
        <section v-for="quota in quotas" :key="quota.period" class="mofa-subscription-quota" :data-period="quota.period">
          <div class="mofa-subscription-quota-heading">
            <h4>{{ t(`userSubscriptions.${quota.period}`) }}</h4>
            <p class="mofa-subscription-amount"><strong>${{ quota.used.toFixed(2) }}</strong><span> / ${{ quota.limit.toFixed(2) }}</span></p>
          </div>
          <progress :value="quota.used" :max="quota.limit" :data-tone="quota.tone"
            :aria-label="t(`userSubscriptions.${quota.period}`)"
            :aria-valuetext="t('userSubscriptions.usageOf', { used: `$${quota.used.toFixed(2)}`, limit: `$${quota.limit.toFixed(2)}` })" />
          <div v-if="quota.hint.kind !== 'inactive'" class="mofa-subscription-reset" :data-kind="quota.hint.kind">
            <span><Icon name="clock" size="sm" aria-hidden="true" />{{ resetLabel(quota.period) }}</span>
            <time v-if="quota.hint.at" :datetime="quota.hint.at" :title="dateLabel(quota.hint.at)">{{ dateLabel(quota.hint.at) }}</time>
          </div>
        </section>
      </div>
      <div v-else class="mofa-subscription-unlimited">
        <Icon name="check" size="md" aria-hidden="true" />
        <div><strong>{{ t('userSubscriptions.unlimited') }}</strong><p>{{ t('userSubscriptions.unlimitedDesc') }}</p></div>
      </div>
    </div>
  </article>
</template>
