<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import DedicatedPicker from '@/custom/components/DedicatedPicker.vue'
import { dedicatedAPI, dedicatedDisplayName, dedicatedErrorCode, type DedicatedBinding } from '@/custom/dedicated/api'
import { dedicatedCopy } from '@/custom/dedicated/copy'
import { dedicatedConfigMessage } from '@/custom/dedicated/config'

const { locale } = useI18n()
const copy = computed(() => dedicatedCopy[locale.value.startsWith('zh') ? 'zh' : 'en'])
const records = ref<DedicatedBinding[]>([])
const busy = ref(false)
const saving = ref(false)
const loadError = ref(false)
const error = ref('')
const success = ref('')
const page = ref(1)
const editing = ref(false)
const editID = ref<number | null>(null)
const editVersion = ref<string>()
const editRevoked = ref(false)
const reactivate = ref(false)
const selectedRecord = computed(() => records.value.find(record => record.id === editID.value))
const confirmID = ref<number | null>(null)
const deleteID = ref<number | null>(null)
const members = ref<number[]>([0])
const formElement = ref<HTMLFormElement>()
const form = reactive({ platform: 'anthropic', user_id: 0, account_id: 0, group_id: 0, label: '', expires_at: '' })
let generation = 0
function date(value: string) { return new Date(value).toLocaleString(locale.value) }
function status(record: DedicatedBinding) { return record.revoked_at ? copy.value.revokedStatus : Date.parse(record.expires_at) <= Date.now() ? copy.value.expired : copy.value.active }
function localDate(value: string) {
  const instant = new Date(value)
  return new Date(instant.getTime() - instant.getTimezoneOffset() * 60_000).toISOString().slice(0, 16)
}
async function load(target = page.value) {
  const current = ++generation
  busy.value = true
  loadError.value = false
  try {
    const result = await dedicatedAPI.list(target)
    if (current === generation) { records.value = result; page.value = target }
  } catch { if (current === generation) { loadError.value = true; records.value = [] } }
  finally { if (current === generation) busy.value = false }
}
async function edit(record?: DedicatedBinding) {
  editVersion.value = record?.updated_at
  editRevoked.value = !!record?.revoked_at
  reactivate.value = false
  editID.value = record?.id ?? null
  Object.assign(form, { platform: record?.platform || 'anthropic', user_id: record?.user_id ?? 0, account_id: record?.account_id ?? 0, group_id: record?.group_id ?? 0, label: record?.label ?? '', expires_at: record ? localDate(record.expires_at) : '' })
  members.value = record?.user_ids?.length ? [...record.user_ids] : [record?.user_id ?? 0]
  deleteID.value = null
  editing.value = true
  confirmID.value = null
  error.value = ''
  success.value = ''
  await nextTick()
  formElement.value?.querySelector<HTMLElement>('input:not(:disabled), select:not(:disabled)')?.focus()
}
function changePlatform() { form.account_id = 0; form.group_id = 0 }
function showError(failure: unknown) {
  const code = dedicatedErrorCode(failure)
  if (code === 'DEDICATED_ACCOUNT_CONFIG') {
    error.value = dedicatedConfigMessage(failure, locale.value) || copy.value.configError
    return
  }
  if (code === 'DEDICATED_ACCOUNT_STALE' || code === 'DEDICATED_ACCOUNT_RESTORE') {
    error.value = code === 'DEDICATED_ACCOUNT_STALE' ? copy.value.staleError : copy.value.restoreError
    return
  }
  error.value = code === 'DEDICATED_ACCOUNT_MODE' ? copy.value.modeError : code === 'DEDICATED_ACCOUNT_CONFIG' ? copy.value.configError : code === 'DEDICATED_ACCOUNT_CONFLICT' ? copy.value.conflict : copy.value.saveError
}
async function save() {
  if (saving.value) return
  if (editRevoked.value && !reactivate.value) { error.value = copy.value.restoreError; return }
  const expiry = Date.parse(form.expires_at)
  if (!members.value.length || members.value.some(id => !id) || new Set(members.value).size !== members.value.length || !form.account_id || !form.group_id || !form.label.trim() || !Number.isFinite(expiry) || expiry <= Date.now()) { error.value = copy.value.validInput; return }
  saving.value = true
  error.value = ''
  success.value = ''
  try {
    await dedicatedAPI.save(editID.value, { expected_updated_at: editVersion.value, reactivate: reactivate.value, user_id: members.value[0]!, user_ids: [...members.value], account_id: form.account_id, group_id: form.group_id, label: form.label.trim(), expires_at: new Date(expiry).toISOString() })
    editing.value = false
    success.value = copy.value.saved
    await load(1)
  } catch (failure) { showError(failure) }
  finally { saving.value = false }
}
async function revoke(id: number) {
  saving.value = true
  error.value = ''
  success.value = ''
  try { await dedicatedAPI.revoke(id); confirmID.value = null; success.value = copy.value.revoked; await load() }
  catch (failure) { showError(failure) }
  finally { saving.value = false }
}
onMounted(() => { void load() })
async function remove(id: number) {
  saving.value = true
  error.value = ''
  success.value = ''
  try {
    await dedicatedAPI.remove(id)
    deleteID.value = null
    if (editID.value === id) editing.value = false
    success.value = copy.value.deleted
    await load(records.value.length === 1 && page.value > 1 ? page.value - 1 : page.value)
  } catch (failure) { showError(failure) }
  finally { saving.value = false }
}
onBeforeUnmount(() => { generation++ })
</script>

<template>
  <AppLayout>
    <template #page-actions><button class="btn btn-secondary" type="button" :disabled="saving || busy" @click="editing = false; error = ''; load()">{{ copy.refresh }}</button><button class="btn btn-primary" type="button" :disabled="saving" @click="edit()">{{ copy.create }}</button></template>
    <div class="mofa-dedicated">
      <details class="mofa-dedicated-setup" :open="!records.length || editing">
        <summary>{{ copy.setupTitle }}</summary><p>{{ copy.setup }}</p><p>{{ copy.billing }}</p>
        <nav :aria-label="copy.setupLinks"><RouterLink to="/admin/groups">{{ copy.groups }}</RouterLink><RouterLink to="/admin/users">{{ copy.users }}</RouterLink><RouterLink to="/admin/accounts">{{ copy.accounts }}</RouterLink></nav>
      </details>
      <p v-if="error" class="mofa-dedicated-feedback" role="alert">{{ error }}</p>
      <p v-if="success" class="mofa-dedicated-feedback" role="status">{{ success }}</p>
      <form v-if="editing" ref="formElement" class="mofa-dedicated-form" @submit.prevent="save">
        <h2>{{ editID ? copy.edit + ' #' + editID : copy.create }}</h2>
        <p v-if="editID" class="mofa-dedicated-note">{{ copy.editHint }}</p>
        <fieldset :disabled="saving">
          <label class="mofa-dedicated-field"><span>{{ copy.platform }}</span><select v-model="form.platform" @change="changePlatform"><option value="anthropic">Claude</option><option value="openai">ChatGPT / Codex</option></select></label>
          <p>{{ copy.membersHint }}</p>
          <div class="mofa-dedicated-form-grid">
            <div v-for="(_, index) in members" :key="index">
              <DedicatedPicker v-model="members[index]!" kind="users" :label="copy.user + ' ' + (index + 1)" :platform="form.platform" :selected-label="selectedRecord?.users?.find(user => user.id === members[index])?.name || (members[index] === selectedRecord?.user_id ? selectedRecord?.user_name : undefined)" />
              <button v-if="members.length > 1" type="button" class="btn btn-secondary" :aria-label="copy.removeUser + ' ' + (index + 1)" @click="members.splice(index, 1)">{{ copy.removeUser }}</button>
            </div>
          </div>
          <button type="button" class="btn btn-secondary" @click="members.push(0)">{{ copy.addUser }}</button>
          <div class="mofa-dedicated-form-grid">
            <DedicatedPicker v-model="form.account_id" kind="accounts" :label="copy.account" :platform="form.platform" :selected-label="form.account_id === selectedRecord?.account_id ? selectedRecord?.account_name : undefined" />
            <DedicatedPicker v-model="form.group_id" kind="groups" :label="copy.group" :platform="form.platform" :selected-label="form.group_id === selectedRecord?.group_id ? selectedRecord?.group_name : undefined" />
          </div>
          <div class="mofa-dedicated-form-grid">
            <label class="mofa-dedicated-field"><span>{{ copy.label }}</span><input v-model="form.label" type="text" maxlength="80" required aria-describedby="dedicated-label-hint" /><small id="dedicated-label-hint">{{ copy.labelHint }}</small></label>
            <label class="mofa-dedicated-field"><span>{{ copy.until }}</span><input v-model="form.expires_at" type="datetime-local" required /></label>
          </div>
          <label v-if="editRevoked" class="mofa-dedicated-restore"><input v-model="reactivate" type="checkbox" required /> {{ copy.restoreConfirm }}</label>
          <div class="mofa-dedicated-actions"><button class="btn btn-primary" type="submit">{{ saving ? copy.saving : copy.save }}</button><button class="btn btn-secondary" type="button" @click="editing = false">{{ copy.cancel }}</button></div>
        </fieldset>
      </form>
      <p v-if="busy" role="status">{{ copy.loading }}</p>
      <div v-else-if="loadError" class="mofa-dedicated-empty" role="alert"><p>{{ copy.loadError }}</p><button class="btn btn-secondary" @click="load()">{{ copy.refresh }}</button></div>
      <p v-else-if="!records.length" class="mofa-dedicated-empty">{{ copy.adminEmpty }}</p>
      <div v-else class="mofa-dedicated-table-wrap" tabindex="0" role="region" :aria-label="copy.adminTitle">
        <table class="mofa-dedicated-table">
          <caption class="sr-only">{{ copy.adminTitle }}</caption>
          <thead><tr><th>{{ copy.label }}</th><th>{{ copy.user }}</th><th>{{ copy.account }} / {{ copy.group }}</th><th>{{ copy.expires }}</th><th><span class="sr-only">{{ copy.edit }}</span></th></tr></thead>
          <tbody><tr v-for="record in records" :key="record.id">
            <th scope="row">{{ record.label }}<small>#{{ record.id }} · {{ status(record) }}</small><small class="mofa-dedicated-health">{{ copy.configHealth[record.config_status || 'unknown'] || copy.configHealth.unknown }}</small></th>
            <td><template v-if="record.users?.length"><div v-for="user in record.users" :key="user.id">{{ dedicatedDisplayName(user.name, user.id) }}</div></template><template v-else>{{ dedicatedDisplayName(record.user_name, record.user_id) }}</template></td><td>{{ dedicatedDisplayName(record.account_name, record.account_id) }} / {{ dedicatedDisplayName(record.group_name, record.group_id) }}</td><td>{{ date(record.expires_at) }}</td>
            <td><div class="mofa-dedicated-actions"><button class="btn btn-secondary" :disabled="saving" @click="edit(record)">{{ copy.edit }}</button><button v-if="!record.revoked_at" class="btn btn-secondary" :disabled="saving" @click="confirmID = record.id">{{ copy.revoke }}</button></div>
              <div v-if="confirmID === record.id" class="mofa-dedicated-confirm"><p>{{ copy.revokeQuestion }}</p><button class="btn btn-danger" :disabled="saving" @click="revoke(record.id)">{{ copy.confirmRevoke }}</button><button class="btn btn-secondary" :disabled="saving" @click="confirmID = null">{{ copy.cancel }}</button></div>
              <button v-if="record.revoked_at" class="btn btn-secondary" :disabled="saving" @click="deleteID = record.id">{{ copy.delete }}</button>
              <div v-if="deleteID === record.id" class="mofa-dedicated-confirm"><p>{{ copy.deleteQuestion }}</p><button class="btn btn-danger" :disabled="saving" @click="remove(record.id)">{{ copy.confirmDelete }}</button><button class="btn btn-secondary" :disabled="saving" @click="deleteID = null">{{ copy.cancel }}</button></div>
            </td>
          </tr></tbody>
        </table>
      </div>
      <nav v-if="!busy && (page > 1 || records.length === 50)" class="mofa-dedicated-pagination" :aria-label="copy.page"><button class="btn btn-secondary" :disabled="page === 1 || saving" @click="load(page - 1)">{{ copy.previous }}</button><span>{{ copy.page }} {{ page }}</span><button class="btn btn-secondary" :disabled="records.length < 50 || saving" @click="load(page + 1)">{{ copy.next }}</button></nav>
    </div>
  </AppLayout>
</template>
