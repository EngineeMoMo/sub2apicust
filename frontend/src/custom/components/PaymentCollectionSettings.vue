<template>
  <section class="card mofa-collection-settings" aria-labelledby="collection-title">
    <h3 id="collection-title">收款限制与开通联系</h3>
    <p>所有人民币通道、充值与在线订阅共用一份收款额度。按最终实付金额计算，包含手续费。</p>
    <p v-if="loading" role="status">正在读取配置…</p>
    <template v-else-if="loaded">
      <label><input v-model="policy.enabled" type="checkbox" /> 启用共享收款限额（人民币；启用后其他币种暂停在线支付）</label>
      <div class="mofa-collection-grid">
        <label>单笔实付上限（元）<input v-model.number="policy.single_max" class="input" type="number" min="0.01" step="0.01" /></label>
        <label>每日收款上限（元）<input v-model.number="policy.daily_max" class="input" type="number" min="0.01" step="0.01" /></label>
        <label>每日统计时区<input v-model.trim="policy.timezone" class="input" placeholder="Asia/Shanghai" /></label>
        <label>快捷充值金额（逗号分隔）<input v-model="quickText" class="input" placeholder="10,20,30,40" /></label>
        <label>每个用户的待付占用上限（元）<input v-model.number="policy.user_hold_max" class="input" type="number" min="0.01" step="0.01" placeholder="默认单笔上限的3倍" /></label>
        <label>无法证明期限的订单复查等待（小时）<input v-model.number="policy.unproven_release_hours" class="input" type="number" min="1" max="360" step="1" placeholder="24" /></label>
      </div>
      <p>上限50元允许支付50元；严格小于50元请设置49.99元。快捷金额为充值基数，含手续费的实付超过上限时不可选。</p>
      <p>同一用户未付款、未释放的订单合计不能超过待付占用上限，防止单个账号反复下单再取消占满全站额度；须不低于单笔上限、不超过每日上限。付款入口带有绝对期限的订单（支付宝扫码与手机网站支付）过期后约10分钟复查渠道；电脑网站支付、易支付等无法证明期限的订单，再按上面的等待时间复查。</p>
      <label><input v-model="policy.allow_custom_amount" type="checkbox" /> 允许用户输入自定义充值金额</label>
      <label>管理员开通联系方式与说明<textarea v-model="policy.contact_text" class="input" rows="3" placeholder="例如：微信 xxx。请提供套餐名称和本站账号，由管理员核实后开通。" /></label>
      <p>此说明会展示给用户与游客，请勿填写密钥。套餐编辑中可选择“联系管理员开通”。</p>
      <button type="button" class="btn btn-primary" :disabled="saving" @click="save">{{ saving ? '正在保存…' : '保存收款配置' }}</button>
      <div v-if="usage" class="mofa-collection-summary" role="status">
        <strong>今日已收 ¥{{ usage.paid.toFixed(2) }}</strong><span>待确认占用 ¥{{ usage.held.toFixed(2) }}</span><span>剩余额度 ¥{{ usage.remaining.toFixed(2) }}{{ policy.enabled ? '' : '（限额未启用）' }}</span>
      </div>
      <p>取消、失败或超时的订单在最晚可付款时间之前继续占用额度；过期后系统向渠道复查，确认未付款才释放，渠道确认已付款则同步付款。退款不恢复当日已收款额度。</p>
      <details v-if="usage?.pending?.length"><summary>检查待确认订单（最近100条）</summary>
        <div v-for="order in usage.pending" :key="order.order_id" class="mofa-collection-summary">
          <span>订单 #{{ order.order_id }} · ¥{{ order.amount }} · {{ order.status }} · 期限 {{ formatCollectionTime(order.expires_at, policy.timezone) }}</span>
          <button type="button" class="btn btn-secondary" :disabled="saving" @click="closeOrder(order.order_id)">查询并关闭渠道订单</button>
          <button type="button" class="btn btn-secondary" :disabled="saving || !canManuallyRelease(order)" :title="canManuallyRelease(order) ? '' : '等待期结束前不能人工释放'" @click="startRelease(order.order_id)">人工释放</button>
          <span v-if="!canManuallyRelease(order)">{{ formatCollectionTime(order.manual_release_after, policy.timezone) }} 后可人工释放</span>
        </div>
      </details>
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
        <button type="button" class="btn btn-secondary" :disabled="saving || !reference || !manualAmount" @click="record">登记已收款</button>
      </details>
    </template>
    <button v-else type="button" class="btn btn-secondary" @click="load">重新加载</button>
    <p v-if="message" :role="failed ? 'alert' : 'status'">{{ message }}</p>
  </section>
</template>
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { adminPaymentAPI } from '@/api/admin/payment'
import { apiClient } from '@/api/client'
import { extractApiErrorMessage } from '@/utils/apiError'
import { canManuallyRelease, formatCollectionTime, type CollectionHold, type CollectionPolicy } from '@/custom/paymentCollection'
const policy = ref<CollectionPolicy>({ enabled: false, single_max: 50, daily_max: 1000, timezone: 'Asia/Shanghai', quick_amounts: [10,20,30,40], allow_custom_amount: true, contact_text: '' })
const quickText = ref('10,20,30,40')
const loading = ref(true), loaded = ref(false), saving = ref(false), failed = ref(false), message = ref('')
const manualAmount = ref<number | null>(null), reference = ref('')
const releaseTarget = ref<number | null>(null), releaseReason = ref(''), releaseConfirmed = ref(false)
const usage = ref<{paid:number;held:number;remaining:number;pending?:CollectionHold[]} | null>(null)
// 清空的数字输入不提交，由后端按单笔／每日上限补默认值。
const optionalNumber = (value: unknown) => typeof value === 'number' && Number.isFinite(value) ? value : undefined
async function refreshUsage() { usage.value = (await apiClient.get('/admin/payment/collection')).data }
async function load() {
  loading.value = true
  try { const { data } = await adminPaymentAPI.getConfig(); if (data.collection) policy.value = data.collection; quickText.value = policy.value.quick_amounts.join(','); loaded.value = true; await refreshUsage() }
  catch (err) { failed.value = true; message.value = extractApiErrorMessage(err, '读取收款配置失败') }
  finally { loading.value = false }
}
async function save() {
  saving.value = true; message.value = ''; failed.value = false
  try {
    const values = quickText.value.split(/[,，\s]+/).filter(Boolean).map(Number)
    if (!values.length || values.some(v => !Number.isFinite(v) || v <= 0)) throw new Error('请输入有效的快捷充值金额')
    await adminPaymentAPI.updateConfig({ collection: { ...policy.value, quick_amounts: values, user_hold_max: optionalNumber(policy.value.user_hold_max), unproven_release_hours: optionalNumber(policy.value.unproven_release_hours) } }); policy.value.quick_amounts = values
    await refreshUsage(); message.value = '收款配置已保存'
  } catch (err) { failed.value = true; message.value = extractApiErrorMessage(err, err instanceof Error ? err.message : '保存失败') }
  finally { saving.value = false }
}
async function record() {
  saving.value = true; failed.value = false; message.value = ''
  try { await apiClient.post('/admin/payment/collection/manual', { amount: manualAmount.value, reference: reference.value }); await refreshUsage(); reference.value = ''; manualAmount.value = null; message.value = '已登记收款，请另行分配订阅或余额' }
  catch (err) { failed.value = true; message.value = extractApiErrorMessage(err, '登记失败') }
  finally { saving.value = false }
}
async function closeOrder(id: number) {
  saving.value = true; failed.value = false; message.value = ''
  try { await apiClient.post(`/admin/payment/collection/${id}/close`); message.value = '已确认渠道关单并释放收款预占' }
  catch (err) { failed.value = true; message.value = extractApiErrorMessage(err, '渠道关单未确认，额度继续保留') }
  finally { try { await refreshUsage() } catch { /* 保留主要操作结果，下次加载刷新统计。 */ } saving.value = false }
}
function startRelease(id: number) { releaseTarget.value = id; releaseReason.value = ''; releaseConfirmed.value = false; message.value = '' }
function cancelRelease() { releaseTarget.value = null; releaseReason.value = ''; releaseConfirmed.value = false }
async function releaseOrder() {
  if (!releaseTarget.value || releaseReason.value.length < 3 || !releaseConfirmed.value) return
  saving.value = true; failed.value = false; message.value = ''
  try { await apiClient.post(`/admin/payment/collection/${releaseTarget.value}/release`, { reason: releaseReason.value, confirmed: true }); cancelRelease(); message.value = '已人工释放收款占用，并记录审计' }
  catch (err) { failed.value = true; message.value = extractApiErrorMessage(err, '人工释放未完成，额度继续保留') }
  finally { try { await refreshUsage() } catch { /* 保留主要操作结果，下次加载刷新统计。 */ } saving.value = false }
}
onMounted(load)
</script>
