<template>
  <button type="button" :class="buttonClass" @click="enter"><slot /></button>
  <BaseDialog :show="show" title="登录后继续" width="narrow" @close="show = false">
    <p>{{ message }}</p>
    <template #footer>
      <button type="button" class="btn btn-secondary" @click="show = false">继续浏览</button>
      <router-link class="btn btn-primary" :to="{ path: '/login', query: { redirect: destination } }">去登录</router-link>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { safeGuestRedirect } from '@/custom/guest/navigation'

const props = withDefaults(defineProps<{ to: string; message?: string; buttonClass?: string }>(), {
  message: '这项功能需要账户信息。登录后继续，或留在这里浏览产品与套餐。',
  buttonClass: 'btn btn-secondary'
})
const auth = useAuthStore()
const router = useRouter()
const show = ref(false)
const destination = computed(() => safeGuestRedirect(props.to))
function enter() {
  if (auth.isAuthenticated) void router.push(destination.value)
  else show.value = true
}
</script>
