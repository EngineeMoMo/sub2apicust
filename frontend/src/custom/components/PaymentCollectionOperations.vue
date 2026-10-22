<template>
  <section class="card mofa-collection-settings" aria-labelledby="collection-operations-title">
    <h2 id="collection-operations-title" class="text-lg font-semibold">收款占用与人工处理</h2>
    <p v-if="loading" role="status">正在读取收款数据…</p>
    <template v-else-if="usage">
      <div class="mofa-collection-summary" role="status">
        <strong>今日已收 ¥{{ usage.paid.toFixed(2) }}</strong>
        <span>待确认占用 ¥{{ usage.held.toFixed(2) }}</span>
        <span>剩余额度 ¥{{ usage.remaining.toFixed(2) }}{{ policy.enabled ? '' : '（限额未启用）' }}</span>
      </div>
      <p>统计与订单时间：{{ policy.timezone }}。退款不恢复当日已收款额度。</p>
      <p>原支付通道已删除时，超期占用自动清理：{{ policy.missing_provider_auto_release ? '已开启' : '未开启' }}。限额及自动清理规则在“设置 → 支付”保存。</p>
      <p>取消、失败或超时的订单在最晚可付款时间之前继续占用额度；过期后系统向渠道复查，确认未付款才释放，渠道确认已付款则同步付款。已删除通道的超期订单按已保存的自动清理规则处理。</p>
      <div class="flex flex-wrap gap-3">
        <button type="button" class="btn btn-primary" :disabled="saving || !usage.pending?.length" @click="recheckOrders">立即批量复查</button>
        <button type="button" class="btn btn-secondary" :disabled="saving" @click="load">刷新数据</button>
      </div>
      <p>按已保存的规则复查，无需逐单操作；剩余订单由后台继续分批处理，不会跳过等待期。释放占用不会发放余额或订阅。</p>
      <h3>待确认订单（最近100条）</h3>
      <p v-if="!usage.pending?.length">暂无待确认的收款占用。</p>
      <div v-for="order in usage.pending" :key="order.order_id" class="mofa-collection-summary">
        <span>订单 #{{ order.order_id }} · ¥{{ order.amount }} · {{ order.status }} · 期限 {{ formatCollectionTime(order.expires_at, policy.timezone) }}</span>
        <button type="button" class="btn btn-secondary" :disabled="saving" @click="closeOrder(order.order_id)">查询并关闭渠道订单</button>
        <button type="button" class="btn btn-secondary" :disabled="saving || !canManuallyRelease(order)" :title="canManuallyRelease(order) ? '' : '等待期结束前不能人工释放'" @click="startRelease(order.order_id)">人工释放</button>
        <span v-if="!canManuallyRelease(order)">{{ formatCollectionTime(order.manual_release_after, policy.timezone) }} 后可人工释放</span>
      </div>
      <div v-if="releaseTarget" class="mofa-collection-summary" role="group" aria-labelledby="collection-release-title">
        <strong id="collection-release-title">人工释放订单 #{{ releaseTarget }}</strong>
        <p>仅在渠道实例已删除、长期查询失败等系统无法复查时使用。释放前先核对商户账单确认此单未收款；渠道仍能查到已付款时会拒绝并同步付款。操作会记录管理员和原因。</p>
        <label>释放原因<textarea v-model.trim="releaseReason" class="input" rows="2" maxlength="500" placeholder="例如：渠道实例已删除，已核对商户账单无此笔收款" /></label>
        <label><input v-model="releaseConfirmed" type="checkbox" /> 我已核对商户账单，确认该订单未收款</label>
        <button type="button" class="btn btn-secondary" :disabled="saving" @click="cancelRelease">取消</button>
        <button type="button" class="btn btn-primary" :disabled="saving || releaseReason.length < 3 || !releaseConfirmed" @click="releaseOrder">确认人工释放</button>
      </div>
      <details>
        <summary>登记今日线下收款</summary>
        <p>同一收款账户的线下到账也要登记；此操作只记录收款，不发放余额或订阅。金额按当前所设时区的今日入账。</p>
        <div class="mofa-collection-grid">
          <label>已收人民币金额<input v-model.number="manualAmount" class="input" type="number" min="0.01" step="0.01" /></label>
          <label>唯一收款参考号<input v-model.trim="reference" class="input" maxlength="100" placeholder="填写商户流水号，避免重复登记" /></label>
        </div>
        <button type="button" class="btn btn-secondary" :disabled="saving || !reference || !manualAmount || manualAmount <= 0" @click="record">登记已收款</button>
      </details>
    </template>
    <button v-else type="button" class="btn btn-secondary" @click="load">重新加载</button>
    <p v-if="message" :role="failed ? 'alert' : 'status'">{{ message }}</p>
    <p v-if="refreshError" role="alert">{{ refreshError }}</p>
    <TotpStepUpDialog :controller="paymentSecurity.stepUp" />
  </section>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { usePaymentSecurity } from '@/custom/usePaymentSecurity'
import TotpStepUpDialog from '@/components/auth/TotpStepUpDialog.vue'
import { apiClient } from '@/api/client'
import { extractApiErrorMessage } from '@/utils/apiError'
import { canManuallyRelease, formatCollectionTime, type CollectionHold, type CollectionPolicy } from '@/custom/paymentCollection'
const paymentSecurity = usePaymentSecurity()
const adminPaymentAPI = paymentSecurity.api

const policy = ref<CollectionPolicy>({ enabled: false, single_max: 50, daily_max: 1000, timezone: 'Asia/Shanghai', quick_amounts: [], allow_custom_amount: true, contact_text: '' })
const usage = ref<{ paid: number; held: number; remaining: number; pending?: CollectionHold[] } | null>(null)
const loading = ref(true), saving = ref(false), failed = ref(false), message = ref(''), refreshError = ref('')
const manualAmount = ref<number | null>(null), reference = ref('')
const releaseTarget = ref<number | null>(null), releaseReason = ref(''), releaseConfirmed = ref(false)

async function load() {
  loading.value = true; failed.value = false; message.value = ''; refreshError.value = ''
  cancelRelease()
  try {
    const [config, collection] = await Promise.all([adminPaymentAPI.getConfig(), apiClient.get('/admin/payment/collection')])
    if (config.data.collection) policy.value = config.data.collection
    usage.value = collection.data
  } catch (err) {
    usage.value = null; failed.value = true; message.value = extractApiErrorMessage(err, '读取收款数据失败')
  } finally { loading.value = false }
}

async function refreshUsage() {
  refreshError.value = ''
  try { usage.value = (await apiClient.get('/admin/payment/collection')).data }
  catch { refreshError.value = '刷新收款统计失败，请点击“刷新数据”后核对最新状态。' }
}

async function record() {
  if (saving.value || !reference.value || !manualAmount.value || manualAmount.value <= 0) return
  saving.value = true; failed.value = false; message.value = ''
  try {
    await adminPaymentAPI.recordCollection({ amount: manualAmount.value, reference: reference.value })
    reference.value = ''; manualAmount.value = null; message.value = '已登记收款，请另行分配订阅或余额'
  } catch (err) { message.value = paymentSecurity.errorMessage(err, '登记失败'); failed.value = Boolean(message.value) }
  finally { await refreshUsage(); saving.value = false }
}

async function closeOrder(id: number) {
  if (saving.value) return
  saving.value = true; failed.value = false; message.value = ''
  try { await adminPaymentAPI.closeCollection(id); message.value = '已确认渠道关单并释放收款预占' }
  catch (err) { message.value = paymentSecurity.errorMessage(err, '渠道关单未确认，额度继续保留'); failed.value = Boolean(message.value) }
  finally { await refreshUsage(); saving.value = false }
}

async function recheckOrders() {
  if (saving.value) return
  saving.value = true; failed.value = false; message.value = ''
  try {
    const { data } = await adminPaymentAPI.recheckCollection()
    message.value = `本次已释放 ${data.released} 笔收款占用，其余订单由后台继续分批复查`
  } catch (err) { message.value = paymentSecurity.errorMessage(err, '批量复查未完成，未确认的占用继续保留'); failed.value = Boolean(message.value) }
  finally { await refreshUsage(); saving.value = false }
}

function startRelease(id: number) { releaseTarget.value = id; releaseReason.value = ''; releaseConfirmed.value = false; message.value = '' }
function cancelRelease() { releaseTarget.value = null; releaseReason.value = ''; releaseConfirmed.value = false }

async function releaseOrder() {
  if (saving.value || !releaseTarget.value || releaseReason.value.length < 3 || !releaseConfirmed.value) return
  saving.value = true; failed.value = false; message.value = ''
  try {
    await adminPaymentAPI.releaseCollection(releaseTarget.value, { reason: releaseReason.value, confirmed: true })
    cancelRelease(); message.value = '已人工释放收款占用，并记录审计'
  } catch (err) { message.value = paymentSecurity.errorMessage(err, '人工释放未完成，额度继续保留'); failed.value = Boolean(message.value) }
  finally { await refreshUsage(); saving.value = false }
}

onMounted(load)
</script>
