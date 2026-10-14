<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import { authIPBlacklistAPI, authIPError, formatAuthIPTime, type AuthIPConfig, type AuthIPEntry } from '@/custom/authIpBlacklist'

const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC'
const config = ref<AuthIPConfig | null>(null)
const form = reactive({ enabled: false, threshold: 10, minutes: 120, cooldown: 15, mode: 'timed' as AuthIPConfig['block_mode'] })
const configLoading = ref(false)
const saving = ref(false)
const configError = ref('')
const message = ref('')
const entries = ref<AuthIPEntry[]>([])
const listLoading = ref(false)
const listError = ref('')
const search = ref('')
const appliedIP = ref('')
const page = ref(1)
const pages = ref(1)
const total = ref(0)
const selected = ref<AuthIPEntry | null>(null)
const removing = ref(false)
let listRevision = 0

async function loadConfig() {
  configLoading.value = true
  config.value = null
  configError.value = ''
  try {
    const loaded = await authIPBlacklistAPI.config()
    config.value = loaded
    Object.assign(form, { enabled: loaded.enabled, threshold: loaded.threshold, minutes: loaded.window_seconds / 60, cooldown: (loaded.block_seconds || 900) / 60, mode: loaded.block_mode })
  } catch (error) { configError.value = authIPError(error).message }
  finally { configLoading.value = false }
}

async function save() {
  if (!config.value || saving.value || configLoading.value) return
  if (!Number.isInteger(form.threshold) || form.threshold < 1 || form.threshold > 1000 ||
      !Number.isInteger(form.minutes) || form.minutes < 1 || form.minutes > 1440 ||
      (form.mode === 'timed' && (!Number.isInteger(form.cooldown) || form.cooldown < 1 || form.cooldown > 1440))) {
    configError.value = '次数须为1–1000的整数，周期和封禁时长须为1–1440分钟的整数'
    return
  }
  saving.value = true
  configError.value = ''
  message.value = ''
  try {
    config.value = await authIPBlacklistAPI.save({ enabled: form.enabled, threshold: form.threshold, window_seconds: form.minutes * 60, block_seconds: form.mode === 'timed' ? form.cooldown * 60 : config.value.block_seconds, block_mode: form.mode, version: config.value.version })
    form.cooldown = config.value.block_seconds / 60
    message.value = '规则已保存，新的失败请求从新统计窗口开始计数。已有临时封禁到期恢复，永久名单须管理员移除；修改规则不会改变已有记录的期限。'
  } catch (error) {
    const detail = authIPError(error)
    if (detail.conflict) {
      await loadConfig()
      configError.value = '规则已被修改，已刷新当前规则，请重新确认后保存。'
    } else { configError.value = detail.message }
  } finally { saving.value = false }
}

async function loadEntries(target = page.value) {
  const revision = ++listRevision
  listLoading.value = true
  listError.value = ''
  try {
    const result = await authIPBlacklistAPI.list(appliedIP.value, target)
    if (revision !== listRevision) return
    if (target > result.pages && result.pages > 0) { await loadEntries(result.pages); return }
    entries.value = result.items
    page.value = result.page
    pages.value = result.pages
    total.value = result.total
  } catch (error) {
    if (revision === listRevision) { listError.value = authIPError(error).message; entries.value = [] }
  } finally { if (revision === listRevision) listLoading.value = false }
}

function filter() {
  appliedIP.value = search.value.trim()
  void loadEntries(1)
}

async function unblock() {
  if (!selected.value || removing.value) return
  const entry = selected.value
  selected.value = null
  removing.value = true
  message.value = ''
  try {
    await authIPBlacklistAPI.unblock(entry)
    await loadEntries()
    message.value = `${entry.ip} 已解封，下一次失败请求从新窗口开始计数。`
  } catch (error) {
    const detail = authIPError(error)
    if (detail.conflict) {
      await loadEntries()
      listError.value = '名单状态已变化，已刷新列表，请重新确认要移除的记录。'
    } else { listError.value = detail.message }
  } finally { removing.value = false }
}

onMounted(() => { void loadConfig(); void loadEntries(1) })
</script>

<template>
  <AppLayout>
    <div class="space-y-6">
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">IP 黑名单</h1>
        <p class="mt-2 text-sm text-gray-600 dark:text-gray-400">仅累计密码认证失败和请求格式异常，成功操作不计数。验证码校验、业务校验、已有限流和服务端故障不计数。达到次数临时封禁，到期自动恢复。</p>
        <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">正常模型调用和已登录管理员的管理操作不受影响。页面时间：{{ timezone }}；统计采用服务端滑动窗口。</p>
      </div>

      <p v-if="message" role="status" class="text-sm text-primary-700 dark:text-primary-300">{{ message }}</p>

      <form class="card space-y-4 p-5" @submit.prevent="save">
        <h2 class="text-lg font-semibold">自动拉黑规则</h2>
        <p class="text-sm text-gray-600 dark:text-gray-400">首次部署默认关闭。开启前核验真实客户端 IP；关闭自动拉黑不会解除已有封禁。</p>
        <fieldset :disabled="!config || configLoading || saving" class="space-y-4">
          <label class="flex items-center gap-2"><input v-model="form.enabled" type="checkbox" data-testid="enabled" />开启自动拉黑</label>
          <div class="grid gap-4 sm:grid-cols-3">
            <label class="space-y-1"><span>触发次数</span><input v-model.number="form.threshold" class="input w-full" type="number" min="1" max="1000" step="1" required data-testid="threshold" /></label>
            <label class="space-y-1"><span>统计周期（分钟）</span><input v-model.number="form.minutes" class="input w-full" type="number" min="1" max="1440" step="1" required data-testid="minutes" /></label>
            <label class="space-y-1"><span>封禁方式</span><select v-model="form.mode" class="input w-full" data-testid="block-mode"><option value="timed">临时封禁，到期自动解除</option><option value="permanent">永久封禁，由管理员解除</option></select></label>
            <label v-if="form.mode === 'timed'" class="space-y-1"><span>封禁时长（分钟）</span><input v-model.number="form.cooldown" class="input w-full" type="number" min="1" max="1440" step="1" required data-testid="cooldown" /></label>
          </div>
          <p class="text-sm text-gray-600 dark:text-gray-400">每次失败向前回看 {{ form.minutes }} 分钟，不按整点清零；恰好位于窗口起点的旧请求不计入。第 {{ form.threshold }} 次失败触发<span v-if="form.mode === 'timed'"> {{ form.cooldown }} 分钟封禁，到期后重新计数。</span><span v-else>永久封禁，须联系管理员移除，才可恢复敏感认证请求。</span></p>
          <button type="submit" class="btn btn-primary" :disabled="!config || saving || configLoading" data-testid="save">{{ saving ? '保存中…' : '保存规则' }}</button>
        </fieldset>
        <p v-if="configLoading" role="status">正在读取规则…</p>
        <p v-if="configError" role="alert" class="text-sm text-red-600">{{ configError }}</p>
        <button v-if="!config && !configLoading" class="btn btn-secondary" type="button" @click="loadConfig">重新读取规则</button>
      </form>

      <section class="card space-y-4 p-5">
        <h2 class="text-lg font-semibold">已封禁 IP <span class="text-sm font-normal">（{{ total }}）</span></h2>
        <form class="flex flex-wrap gap-2" @submit.prevent="filter">
          <input v-model="search" class="input min-w-0 flex-1" aria-label="精确 IP 搜索" placeholder="输入完整 IPv4 或 IPv6 地址，留空查看全部" data-testid="search" />
          <button type="submit" class="btn btn-secondary" :disabled="listLoading || removing">搜索</button>
          <button type="button" class="btn btn-secondary" :disabled="listLoading || removing" @click="loadEntries()">刷新</button>
        </form>
        <p v-if="listError" role="alert" class="text-sm text-red-600">{{ listError }}</p>
        <p v-if="listLoading" role="status">正在读取名单…</p>
        <p v-else-if="!listError && entries.length === 0" class="text-sm text-gray-500">暂无封禁记录</p>
        <div v-else-if="entries.length" class="overflow-x-auto">
          <table class="w-full text-left text-sm">
            <thead><tr class="border-b border-gray-200 dark:border-dark-600"><th class="p-3">IP</th><th class="p-3">封禁时间 / 到期时间</th><th class="p-3">触发接口</th><th class="p-3">失败次数 / 统计周期</th><th class="p-3">窗口首末请求时间</th><th class="p-3">操作</th></tr></thead>
            <tbody><tr v-for="entry in entries" :key="entry.id" class="border-b border-gray-200 dark:border-dark-600">
              <td class="whitespace-nowrap p-3 font-mono">{{ entry.ip }}</td>
              <td class="whitespace-nowrap p-3"><div>{{ formatAuthIPTime(entry.blocked_at, timezone) }}</div><div>{{ entry.blocked_until ? `到期：${formatAuthIPTime(entry.blocked_until, timezone)}` : '永久封禁：管理员移除' }}</div></td>
              <td class="max-w-xs break-all p-3 font-mono">{{ entry.trigger_path }}</td>
              <td class="whitespace-nowrap p-3">{{ entry.trigger_count }} 次 / {{ entry.window_seconds / 60 }} 分钟</td>
              <td class="whitespace-nowrap p-3"><div>{{ formatAuthIPTime(entry.first_event_at, timezone) }}</div><div>{{ formatAuthIPTime(entry.last_event_at, timezone) }}</div></td>
              <td class="p-3"><button class="btn btn-secondary" type="button" :disabled="removing || listLoading" @click="selected = entry">移除</button></td>
            </tr></tbody>
          </table>
        </div>
        <div class="flex items-center justify-end gap-3">
          <button class="btn btn-secondary" type="button" :disabled="page <= 1 || listLoading || removing" @click="loadEntries(page - 1)">上一页</button>
          <span>{{ page }} / {{ pages }}</span>
          <button class="btn btn-secondary" type="button" :disabled="page >= pages || listLoading || removing" @click="loadEntries(page + 1)">下一页</button>
        </div>
      </section>
    </div>
    <ConfirmDialog :show="selected !== null" title="移除 IP 黑名单" :message="`确认解封 ${selected?.ip || ''}？解封后重新累计可识别的认证失败。`" confirm-text="确认移除" @confirm="unblock" @cancel="selected = null" />
  </AppLayout>
</template>
