<template>
  <div>
    <label id="alipay-desktop-wap-label" class="input-label">
      {{ t('admin.settings.payment.alipayDesktopWapQRCode') }}
    </label>
    <div class="flex items-start gap-2">
      <Toggle
        :model-value="modelValue"
        aria-labelledby="alipay-desktop-wap-label"
        aria-describedby="alipay-desktop-wap-hint"
        @update:model-value="emit('update:modelValue', $event)"
      />
      <span id="alipay-desktop-wap-hint" class="text-sm text-gray-500 dark:text-gray-400">
        {{ t('admin.settings.payment.alipayDesktopWapQRCodeHint') }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'

withDefaults(defineProps<{ modelValue?: boolean }>(), { modelValue: false })
const emit = defineEmits<{ 'update:modelValue': [value: boolean] }>()
const { t } = useI18n()
</script>
