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
      <label><input id="collection-auto-missing-provider" v-model="policy.missing_provider_auto_release" type="checkbox" /> 原支付通道已删除时，超期占用自动清理</label>
      <p>保存后生效。仅处理已取消、过期或失败且未记录付款的订单，在订单期限加上方等待时长及10分钟后自动分批清理，无需逐单点击。通道仍存在但查询失败的订单继续保留。</p>
      <p>此规则仅释放收款额度占用，不代表已核实未收款，也不会发放余额或订阅；已到账但未入账的订单仍须核对商户账单处理。默认关闭，每笔自动清理都会记录审计。</p>
      <p>上限50元允许支付50元；严格小于50元请设置49.99元。快捷金额为充值基数，含手续费的实付超过上限时不可选。</p>
      <p>同一用户未付款、未释放的订单合计不能超过待付占用上限，防止单个账号反复下单再取消占满全站额度；须不低于单笔上限、不超过每日上限。付款入口带有绝对期限的订单（支付宝扫码与手机网站支付）过期后约10分钟复查渠道；电脑网站支付、易支付等无法证明期限的订单，再按上面的等待时间复查。</p>
      <label><input v-model="policy.allow_custom_amount" type="checkbox" /> 允许用户输入自定义充值金额</label>
      <label>管理员开通联系方式与说明<textarea v-model="policy.contact_text" class="input" rows="3" placeholder="例如：微信 xxx。请提供套餐名称和本站账号，由管理员核实后开通。" /></label>
      <p>此说明会展示给用户与游客，请勿填写密钥。套餐编辑中可选择“联系管理员开通”。</p>
      <button type="button" class="btn btn-primary" :disabled="saving" @click="save">{{ saving ? '正在保存…' : '保存收款配置' }}</button>
      <RouterLink class="btn btn-secondary" to="/admin/orders/reconciliation">前往订单核账：复查占用、人工释放与登记收款</RouterLink>
    </template>
    <button v-else type="button" class="btn btn-secondary" @click="load">重新加载</button>
    <p v-if="message" :role="failed ? 'alert' : 'status'">{{ message }}</p>
    <TotpStepUpDialog :controller="paymentSecurity.stepUp" />
  </section>
</template>
<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { usePaymentSecurity } from '@/custom/usePaymentSecurity'
import TotpStepUpDialog from '@/components/auth/TotpStepUpDialog.vue'
import { extractApiErrorMessage } from '@/utils/apiError'
import type { CollectionPolicy } from '@/custom/paymentCollection'
const paymentSecurity = usePaymentSecurity()
const adminPaymentAPI = paymentSecurity.api
const policy = ref<CollectionPolicy>({ enabled: false, single_max: 50, daily_max: 1000, timezone: 'Asia/Shanghai', quick_amounts: [10,20,30,40], allow_custom_amount: true, contact_text: '' })
const quickText = ref('10,20,30,40')
const loading = ref(true), loaded = ref(false), saving = ref(false), failed = ref(false), message = ref('')
// 清空的数字输入不提交，由后端按单笔／每日上限补默认值。
const optionalNumber = (value: unknown) => typeof value === 'number' && Number.isFinite(value) ? value : undefined
async function load() {
  loading.value = true
  failed.value = false; message.value = ''
  try { const { data } = await adminPaymentAPI.getConfig(); if (data.collection) policy.value = data.collection; quickText.value = policy.value.quick_amounts.join(','); loaded.value = true }
  catch (err) { failed.value = true; message.value = extractApiErrorMessage(err, '读取收款配置失败') }
  finally { loading.value = false }
}
async function save() {
  saving.value = true; message.value = ''; failed.value = false
  try {
    const values = quickText.value.split(/[,，\s]+/).filter(Boolean).map(Number)
    if (!values.length || values.some(v => !Number.isFinite(v) || v <= 0)) throw new Error('请输入有效的快捷充值金额')
    await adminPaymentAPI.updateConfig({ collection: { ...policy.value, quick_amounts: values, user_hold_max: optionalNumber(policy.value.user_hold_max), unproven_release_hours: optionalNumber(policy.value.unproven_release_hours) } }); policy.value.quick_amounts = values
    message.value = '收款配置已保存'
  } catch (err) { message.value = paymentSecurity.errorMessage(err, err instanceof Error ? err.message : '保存失败'); failed.value = Boolean(message.value) }
  finally { saving.value = false }
}
onMounted(load)
</script>
