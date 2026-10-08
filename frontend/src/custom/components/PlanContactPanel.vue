<template>
  <section class="card mofa-collection-settings" aria-label="联系管理员开通">
    <h3>{{ name }} · 联系管理员开通</h3>
    <p>请提供套餐名称和本站账号，由管理员确认后分配订阅。联系咨询不占用库存。</p>
    <p class="mofa-contact-text">{{ contact || '请通过本站客服渠道联系管理员，确认套餐权益和开通方式。' }}</p>
    <button type="button" class="btn btn-secondary" @click="copy">复制开通信息</button>
    <p v-if="message" role="status">{{ message }}</p>
  </section>
</template>
<script setup lang="ts">
import { ref } from 'vue'
const props = defineProps<{name:string;contact?:string}>()
const message = ref('')
async function copy() {
  try { await navigator.clipboard.writeText(`套餐：${props.name}\n${props.contact || '请联系本站管理员开通'}`); message.value = '已复制，请补充本站账号后联系管理员' }
  catch { message.value = '复制失败，请手动复制上方信息' }
}
</script>
