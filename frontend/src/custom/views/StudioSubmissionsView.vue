<template>
  <AppLayout>
    <section class="mofa-studio-contribute">
      <header class="mofa-studio-contribute-heading">
        <div><h1>{{ reviewing ? '作品审核' : '分享你的作品与方法' }}</h1><p>好作品不只看效果。把提示词、模型和做法一起留下，让别人学得会。</p></div>
        <RouterLink to="/tools/studio" class="btn btn-secondary">返回魔法工坊</RouterLink>
      </header>
      <nav class="mofa-studio-contribute-tabs" aria-label="投稿工作区">
        <button v-if="!reviewing" :aria-current="mode === 'submit' ? 'page' : undefined" @click="mode = 'submit'">提交作品</button>
        <button v-if="!reviewing" :aria-current="mode === 'mine' ? 'page' : undefined" @click="showMine">我的投稿</button>
        <RouterLink v-if="auth.isAdmin" to="/admin/studio/submissions" :aria-current="reviewing ? 'page' : undefined">作品审核</RouterLink>
        <RouterLink v-if="reviewing" to="/tools/studio/submit">我的投稿入口</RouterLink>
      </nav>
      <p v-if="notice" role="status" class="mofa-studio-feedback">{{ notice }}</p>
      <p v-if="problem" role="alert" class="mofa-studio-feedback mofa-studio-error">{{ problem }}</p>
      <form v-if="mode === 'submit' && !reviewing" class="mofa-studio-submit-grid" @submit.prevent="submit">
        <div class="mofa-studio-panel">
          <h2>作品文件</h2><p>提交后先进入私有待审核区，只有你和管理员能查看。</p>
          <label>作品类型<select v-model="form.media" aria-describedby="studio-submission-availability" @change="clearFile"><option value="image">图片</option><option value="video" disabled>视频成片（待开放）</option></select></label>
          <p id="studio-submission-availability" class="mofa-studio-fine">当前可提交图片。视频投稿待开放，已有投稿仍可查看或撤回。</p>
          <label class="mofa-studio-file">选择{{ form.media === 'image' ? '图片' : '视频' }}<input ref="fileInput" type="file" :accept="form.media === 'image' ? 'image/png,image/jpeg,image/webp' : 'video/mp4'" @change="pickFile"><span>{{ file?.name || (form.media === 'image' ? 'PNG / JPEG / WebP · 最大12MB' : 'MP4 · 最大30MB') }}</span></label>
          <div v-if="fileURL" class="mofa-studio-media"><img v-if="form.media === 'image'" :src="fileURL" alt="待投稿图片预览"><video v-else :src="fileURL" controls playsinline preload="metadata" aria-label="待投稿视频预览" /></div>
          <p class="mofa-studio-fine">预览仅在本机；点击“提交审核”才上传。不要提交密钥、隐私信息、未经授权的作品或未成年人的性感内容。</p>
        </div>
        <div class="mofa-studio-panel">
          <h2>让作品可以被学习</h2>
          <div class="mofa-studio-field-grid">
            <label>作品标题<input v-model.trim="form.title" required minlength="2" maxlength="80"></label>
            <label>公开署名<input v-model.trim="form.author" required maxlength="40" placeholder="昵称，不必是真名"></label>
            <label>题材<select v-model="form.category" required><option value="" disabled>选择题材</option><option v-for="value in studioCategories" :key="value">{{ value }}</option></select></label>
            <label>视觉风格<select v-model="form.style" required><option value="" disabled>选择风格</option><option v-for="value in studioStyles" :key="value">{{ value }}</option></select></label>
            <label>生成模型 / 制作工具<input v-model.trim="form.model" required maxlength="80" placeholder="真实使用的模型与版本"></label>
            <label>提示词来源<select v-model="form.prompt_kind"><option value="actual">实际使用的原始提示词</option><option value="reference">反推或整理的参考模板</option></select></label>
          </div>
          <label>完整提示词<textarea v-model.trim="form.prompt" required minlength="10" maxlength="12000" rows="8" placeholder="写清主体、场景、构图、光线和视觉风格。" /></label>
          <label>参数与制作说明<textarea v-model.trim="form.notes" maxlength="1000" rows="3" placeholder="画幅、seed（如有）、后期步骤或授权来源；不要填API Key。" /></label>
          <label class="mofa-studio-check"><input v-model="rights" type="checkbox" required>我拥有素材及人物肖像的展示授权，没有侵犯他人权利；性感题材中的人物为明确成年人。</label>
          <label class="mofa-studio-check"><input v-model="consent" type="checkbox" required>我同意审核通过后公开展示作品、署名和提示词，供访问者查看与学习；这不表示自动授予素材的商用许可。</label>
          <button class="btn btn-primary" type="submit" :disabled="busy || !rights || !consent">{{ busy ? '正在上传…' : '提交审核' }}</button>
        </div>
      </form>
      <div v-else class="mofa-studio-review-grid">
        <div class="mofa-studio-panel">
          <div class="mofa-studio-list-heading"><h2>{{ reviewing ? '投稿队列' : '我的投稿' }}</h2><button class="btn btn-secondary" :disabled="busy" @click="loadList">刷新</button></div>
          <p v-if="loading" role="status">正在读取投稿…</p>
          <p v-else-if="!entries.length">目前没有投稿。{{ reviewing ? '用户提交后会出现在这里，不会自动公开。' : '提交作品后，可以在这里查看审核状态。' }}</p>
          <button v-for="entry in entries" :key="entry.id" class="mofa-studio-submission-row" :aria-pressed="selected?.id === entry.id" @click="openEntry(entry)">
            <strong>{{ entry.title }}</strong><span>{{ entry.style }} · {{ entry.media === 'video' ? '视频' : '图片' }}</span><span>{{ statusLabels[entry.status] || entry.status }}</span>
          </button>
        </div>
        <div class="mofa-studio-panel">
          <template v-if="selected">
            <h2>{{ selected.title }}</h2><p>{{ selected.author }} · {{ selected.category }} · {{ selected.style }} · {{ statusLabels[selected.status] }}</p>
            <p v-if="previewLoading" role="status">正在读取私有素材…</p>
            <div v-if="previewURL" class="mofa-studio-media"><img v-if="selected.media === 'image'" :src="previewURL" :alt="selected.title" @load="previewReady = true"><video v-else :src="previewURL" controls playsinline preload="metadata" @loadedmetadata="previewReady = true" /></div>
            <h3>{{ selected.prompt_kind === 'actual' ? '投稿人提供的原始提示词' : '反推 / 整理参考模板（非原始记录）' }}</h3>
            <pre class="mofa-studio-prompt">{{ selected.prompt }}</pre><p>模型 / 工具：{{ selected.model }}</p><p v-if="selected.notes">{{ selected.notes }}</p><p v-if="selected.reason">审核说明：{{ selected.reason }}</p>
            <template v-if="reviewing">
              <label>退回 / 下架原因<textarea v-model.trim="reason" maxlength="500" rows="3" placeholder="3–500字；请给出可修改的具体原因。" /></label>
              <label v-if="['pending', 'unpublished'].includes(selected.status)" class="mofa-studio-check"><input v-model="reviewConfirmed" type="checkbox">我已完整查看或播放素材，并检查授权、成人边界、提示词真实性和内容安全，同意公开。</label>
              <div class="mofa-studio-review-actions">
                <button v-if="['pending', 'unpublished'].includes(selected.status)" class="btn btn-primary" :disabled="busy || !previewReady || !reviewConfirmed" @click="review('approve')">审核通过并公开</button>
                <button v-if="selected.status === 'pending'" class="btn btn-secondary" :disabled="busy || reason.length < 3" @click="review('reject')">退回修改</button>
                <button v-if="selected.status === 'published'" class="btn btn-secondary" :disabled="busy || reason.length < 3" @click="review('unpublish')">下架作品</button>
              </div>
            </template>
            <template v-else-if="selected.status !== 'withdrawn'">
              <button v-if="selected.status === 'rejected' || selected.status === 'unpublished'" class="btn btn-secondary" :disabled="busy" @click="reuseEntry">修改后重新投稿</button>
              <p v-if="selected.status === 'rejected' || selected.status === 'unpublished'" class="mofa-studio-fine">重新投稿需要再次上传作品并确认授权；旧记录保留，新版本重新审核，不会直接公开。</p>
              <button v-if="!withdrawConfirm" class="btn btn-secondary" :disabled="busy" @click="withdrawConfirm = true">撤回作品</button>
              <div v-else class="mofa-studio-feedback"><p>撤回后作品不再公开，记录仍保留。确定撤回吗？</p><button class="btn btn-secondary" :disabled="busy" @click="withdrawConfirm = false">取消</button><button class="btn btn-primary" :disabled="busy" @click="withdraw">确认撤回</button></div>
            </template>
          </template>
          <p v-else>选择一份投稿，查看完整作品、提示词与审核记录。</p>
        </div>
      </div>
    </section>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import { useAuthStore } from '@/stores/auth'
import { studioAPI, studioCategories, studioStyles, statusLabels, validateStudioFile, type StudioSubmission } from '@/custom/studio/api'

const route = useRoute(), auth = useAuthStore()
const reviewing = computed(() => route.path.startsWith('/admin/'))
const mode = ref('submit'), busy = ref(false), loading = ref(false), problem = ref(''), notice = ref('')
const rights = ref(false), consent = ref(false), reviewConfirmed = ref(false), withdrawConfirm = ref(false), reason = ref('')
const file = ref<File>(), fileURL = ref(''), fileInput = ref<HTMLInputElement>()
const entries = ref<StudioSubmission[]>([]), selected = ref<StudioSubmission>()
const previewURL = ref(''), previewLoading = ref(false), previewReady = ref(false)
const form = reactive({ media: 'image' as 'image' | 'video', title: '', author: '', category: '', style: '', model: '', prompt_kind: 'actual', prompt: '', notes: '' })
let previewRequest: AbortController | undefined, listRequest: AbortController | undefined
let accountEpoch = 0
function failure(error: unknown) {
  const message = (error as { message?: unknown })?.message
  problem.value = typeof message === 'string' ? message : '操作未完成，请稍后重试。'
}
function clearFile() {
  if (fileURL.value) URL.revokeObjectURL(fileURL.value)
  fileURL.value = ''; file.value = undefined
  if (fileInput.value) fileInput.value.value = ''
}
function pickFile(event: Event) {
  const chosen = (event.target as HTMLInputElement).files?.[0]
  clearFile(); problem.value = ''
  const error = validateStudioFile(chosen, form.media)
  if (error) { problem.value = error; return }
  file.value = chosen; fileURL.value = URL.createObjectURL(chosen!)
}
function clearPreview() {
  previewRequest?.abort()
  if (previewURL.value) URL.revokeObjectURL(previewURL.value)
  previewURL.value = ''; previewReady.value = false; previewLoading.value = false
}
async function loadList() {
  listRequest?.abort(); const operation = new AbortController(); listRequest = operation
  loading.value = true; problem.value = ''
  try { const result = await studioAPI.list(reviewing.value, operation.signal); if (!operation.signal.aborted) entries.value = result }
  catch (error) { if (!operation.signal.aborted) failure(error) }
  finally { if (listRequest === operation) loading.value = false }
}
function showMine() { mode.value = 'mine'; void loadList() }
function reuseEntry() {
  if (!selected.value) return
  if (selected.value.media === 'video') { notice.value = '视频投稿待开放，原记录仍可查看或撤回。你可以先分享图片作品。'; return }
  clearFile(); clearPreview()
  const entry = selected.value
  Object.assign(form, { media: entry.media, title: entry.title, author: entry.author, category: entry.category, style: entry.style, model: entry.model, prompt_kind: entry.prompt_kind, prompt: entry.prompt, notes: entry.notes })
  rights.value = consent.value = false; problem.value = ''; notice.value = '已填入旧投稿内容。修改后重新上传素材，提交新一轮审核。'; mode.value = 'submit'
}
async function openEntry(entry: StudioSubmission) {
  clearPreview(); selected.value = entry; reason.value = ''; reviewConfirmed.value = false; withdrawConfirm.value = false
  const operation = new AbortController(); previewRequest = operation; previewLoading.value = true
  try { const blob = await studioAPI.preview(entry.id, reviewing.value, operation.signal); if (!operation.signal.aborted) previewURL.value = URL.createObjectURL(blob) }
  catch (error) { if (!operation.signal.aborted) failure(error) }
  finally { if (previewRequest === operation) previewLoading.value = false }
}
async function submit() {
  if (busy.value) return
  problem.value = ''; notice.value = ''
  if (form.media === 'video') { problem.value = '视频投稿待开放，请先选择图片作品。'; return }
  const error = validateStudioFile(file.value, form.media)
  if (error || !rights.value || !consent.value) { problem.value = error || '请确认素材权利与公开授权。'; return }
  const body = new FormData()
  Object.entries(form).forEach(([key, value]) => body.append(key, value))
  body.append('file', file.value!); body.append('rights_confirmed', 'true'); body.append('publish_consent', 'true')
  busy.value = true
  const epoch = accountEpoch
  try {
    await studioAPI.submit(body)
    if (epoch !== accountEpoch) return
    clearFile(); rights.value = false; consent.value = false
    notice.value = '已提交，状态为待审核。审核通过前不会出现在公开图库。'
    mode.value = 'mine'; await loadList()
  } catch (error) { if (epoch === accountEpoch) failure(error) }
  finally { if (epoch === accountEpoch) busy.value = false }
}
async function review(action: 'approve' | 'reject' | 'unpublish') {
  if (!auth.isAdmin || !selected.value || busy.value || (action === 'approve' && (!reviewConfirmed.value || !previewReady.value))) return
  busy.value = true; problem.value = ''
  const epoch = accountEpoch
  try {
    const updated = await studioAPI.review(selected.value, action, reason.value, reviewConfirmed.value)
    if (epoch !== accountEpoch) return
    selected.value = updated; reviewConfirmed.value = false
    notice.value = action === 'approve' ? '已通过审核并公开展示。' : action === 'reject' ? '已退回，并保留修改原因。' : '作品已下架，公开入口不再提供素材。'
    await loadList()
  } catch (error) { if (epoch === accountEpoch) failure(error) }
  finally { if (epoch === accountEpoch) busy.value = false }
}
async function withdraw() {
  if (!selected.value || busy.value || !withdrawConfirm.value) return
  busy.value = true; problem.value = ''
  const epoch = accountEpoch
  try { const updated = await studioAPI.withdraw(selected.value.id); if (epoch !== accountEpoch) return; selected.value = updated; withdrawConfirm.value = false; notice.value = '作品已撤回。'; await loadList() }
  catch (error) { if (epoch === accountEpoch) failure(error) }
  finally { if (epoch === accountEpoch) busy.value = false }
}
watch(() => route.path, () => { clearPreview(); selected.value = undefined; mode.value = 'submit'; if (reviewing.value) void loadList() })
watch(() => auth.user?.id, () => {
  accountEpoch++; clearFile(); clearPreview(); listRequest?.abort(); entries.value = []; selected.value = undefined
  Object.assign(form, { media: 'image', title: '', author: '', category: '', style: '', model: '', prompt_kind: 'actual', prompt: '', notes: '' })
  rights.value = consent.value = reviewConfirmed.value = withdrawConfirm.value = false
  busy.value = loading.value = false; problem.value = notice.value = reason.value = ''; mode.value = 'submit'
})
onMounted(() => { if (reviewing.value) void loadList() })
onUnmounted(() => { accountEpoch++; clearFile(); clearPreview(); listRequest?.abort() })
</script>
