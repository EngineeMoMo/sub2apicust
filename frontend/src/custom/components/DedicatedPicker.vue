<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { dedicatedAPI, type ChoiceKind, type DedicatedChoice } from '@/custom/dedicated/api'
import { dedicatedCopy } from '@/custom/dedicated/copy'

const props = defineProps<{ kind: ChoiceKind; modelValue: number; label: string; platform: string; disabled?: boolean; selectedLabel?: string }>()
const emit = defineEmits<{ 'update:modelValue': [value: number] }>()
const { locale } = useI18n()
const copy = computed(() => dedicatedCopy[locale.value.startsWith('zh') ? 'zh' : 'en'])
const query = ref('')
const choices = ref<DedicatedChoice[]>([])
const busy = ref(false)
const error = ref(false)
let generation = 0
const options = computed(() => props.modelValue && !choices.value.some(item => item.id === props.modelValue)
  ? [{ id: props.modelValue, label: props.selectedLabel ? `${props.selectedLabel} #${props.modelValue}` : '#' + props.modelValue }, ...choices.value] : choices.value)
async function search() {
  const current = ++generation
  busy.value = true
  error.value = false
  try {
    const result = await dedicatedAPI.choices(props.kind, query.value.trim(), props.platform)
    if (current === generation) choices.value = result
  } catch {
    if (current === generation) { error.value = true; choices.value = [] }
  } finally {
    if (current === generation) busy.value = false
  }
}
watch(() => props.platform, () => { choices.value = []; if (!props.disabled) void search() }, { immediate: true })
onBeforeUnmount(() => { generation++ })
</script>

<template>
  <fieldset class="mofa-dedicated-picker" :disabled="disabled">
    <legend>{{ label }}</legend>
    <div class="mofa-dedicated-search">
      <input v-model="query" :aria-label="label + ' — ' + copy.search" :placeholder="copy.searchPlaceholder" type="search" @keydown.enter.prevent="search" />
      <button class="btn btn-secondary" type="button" :disabled="busy || disabled" @click="search">{{ busy ? copy.loading : copy.search }}</button>
    </div>
    <select :value="modelValue" :aria-label="label" required @change="emit('update:modelValue', Number(($event.target as HTMLSelectElement).value))">
      <option :value="0">{{ copy.choose }}</option>
      <option v-for="item in options" :key="item.id" :value="item.id">{{ item.label }}</option>
    </select>
    <p v-if="error" role="alert">{{ copy.loadError }}</p>
    <p v-else-if="!busy && !choices.length && !disabled">{{ copy.noChoices }}</p>
    <p v-else-if="!disabled">{{ copy.searchHint }}</p>
  </fieldset>
</template>
