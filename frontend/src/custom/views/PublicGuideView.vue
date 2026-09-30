<template>
  <component :is="embedded ? 'div' : PublicLayout" class="mofa-guide">
    <header class="mofa-guide-intro">
      <h1 v-if="!hideTitle">把魔法家族接入你的工具</h1>
      <p>从本站 API Key 到 Claude Code、Codex 和常见客户端。先选协议，再填地址；不需要把密钥交给本页。</p>
      <nav class="mofa-guide-index" aria-label="教程目录">
        <a href="#guide-prepare">接入前准备</a>
        <a href="#guide-claude">Claude Code</a>
        <a href="#guide-codex">Codex</a>
        <a href="#guide-others">其他工具</a>
        <a href="#guide-troubleshoot">排错</a>
      </nav>
    </header>

    <section id="guide-prepare" class="mofa-guide-section">
      <h2>接入前准备</h2>
      <ol class="mofa-guide-steps">
        <li>在<router-link v-if="scope === 'preview'" to="/preview/keys">API 密钥页</router-link><GuestAction v-else to="/keys" button-class="mofa-guide-inline-action">控制台的 API 密钥页</GuestAction>创建密钥，核对密钥所属分组和可用模型；登录后可点密钥旁的“使用密钥”查看针对该分组生成的配置。</li>
        <li>选择与工具匹配的协议。Claude Code 用 Anthropic Messages；Codex 用 OpenAI Responses；其他工具通常用 OpenAI Chat Completions。</li>
        <li>模型名填写你的密钥实际可用的模型 ID，不照抄下文占位符。套餐权益以<router-link :to="plansTarget">套餐页面</router-link>和账户页面为准。</li>
      </ol>
      <div class="mofa-guide-addresses" aria-label="本站接口地址">
        <div><strong>Claude Code 基础地址</strong><code>{{ rootUrl }}</code><small>客户端自行追加 /v1/messages，不要手动重复拼接。</small></div>
        <div><strong>Codex / OpenAI 基础地址</strong><code>{{ apiUrl }}</code><small>客户端分别请求 /responses 或 /chat/completions。</small></div>
      </div>
      <p class="mofa-guide-caution">本页示例只含占位符，不读取、显示或保存你的密钥。不要把密钥写进公开仓库、截图或分享链接。</p>
      <details class="mofa-guide-example">
        <summary>怎样设置命令行使用的密钥变量？</summary>
        <p>把“你的本站 API Key”替换成自己的密钥。以下设置仅对当前终端及从它启动的程序生效；桌面应用按后文的独立步骤配置。</p>
        <h3>macOS / Linux 终端</h3>
        <pre><code>export MOFAMILY_API_KEY="你的本站 API Key"</code></pre>
        <h3>Windows PowerShell</h3>
        <pre><code>$env:MOFAMILY_API_KEY = "你的本站 API Key"</code></pre>
        <p>命令中填写的密钥可能进入本机终端历史，请使用你信任的终端，不分享历史记录。运行配置命令时使用同一个终端窗口。</p>
      </details>
    </section>

    <section id="guide-claude" class="mofa-guide-section">
      <h2>Claude Code</h2>
      <p>适用于 Claude Code CLI 和本机桌面会话；选择支持 Anthropic Messages 的模型及分组。本站密钥通过 <code>ANTHROPIC_AUTH_TOKEN</code> 发送，不是 Anthropic 官方订阅登录。</p>
      <p>还没安装？按<a href="https://code.claude.com/docs/en/setup" target="_blank" rel="noopener noreferrer">官方安装步骤</a>安装 CLI，运行 <code>claude --version</code> 确认可用；桌面端从官方桌面说明进入下载。</p>
      <h3>命令行 · macOS / Linux</h3>
      <pre><code>{{ claudeUnix }}</code></pre>
      <h3>命令行 · Windows PowerShell</h3>
      <pre><code>{{ claudePowerShell }}</code></pre>
      <p>先按接入前准备设置 <code>MOFAMILY_API_KEY</code>，再在同一终端运行上述命令。进入项目目录启动后，用 <code>/model</code> 选择可用模型，发一条简短消息，并在控制台用量页核对本次请求。</p>
      <h3>Claude Code 桌面端</h3>
      <ol class="mofa-guide-steps">
        <li>新建或打开 <strong>Local</strong> 会话，在输入框附近的环境选择器进入本机环境设置。</li>
        <li>填入 <code>ANTHROPIC_BASE_URL={{ rootUrl }}</code> 和 <code>ANTHROPIC_AUTH_TOKEN=你的本站 API Key</code>，保存后重启本机会话。</li>
        <li>选择实际可用模型，发一条简短测试消息。Windows 桌面端不会读取 PowerShell profile；只在终端里 export / 设置变量可能对桌面端无效。</li>
      </ol>
      <p>云端会话与本机环境不是同一套设置；不要因为本机可用就认为云端也已接通。<a href="https://code.claude.com/docs/en/llm-gateway" target="_blank" rel="noopener noreferrer">Claude Code 网关说明</a> · <a href="https://code.claude.com/docs/en/desktop" target="_blank" rel="noopener noreferrer">桌面端说明</a></p>
    </section>

    <section id="guide-codex" class="mofa-guide-section">
      <h2>Codex CLI 与桌面端</h2>
      <p>这里的 Codex 指 <code>codex</code> CLI／Codex 桌面应用，不是 <code>openai</code> CLI。本站使用 <strong>Responses</strong> 协议；不要把 Chat Completions 地址填进 Codex 提供商。</p>
      <p>按<a href="https://learn.chatgpt.com/docs/cli" target="_blank" rel="noopener noreferrer">官方 CLI 说明</a>安装，运行 <code>codex --version</code> 确认可用。</p>
      <h3>Codex CLI</h3>
      <ol class="mofa-guide-steps">
        <li>在本机设置环境变量 <code>MOFAMILY_API_KEY</code> 为你的本站 API Key。</li>
        <li>编辑用户级 <code>~/.codex/config.toml</code>（Windows 为 <code>%USERPROFILE%\.codex\config.toml</code>），加入以下配置；若已有设置，请合并同名字段，勿覆盖整个文件。</li>
      </ol>
      <pre><code>{{ codexConfig }}</code></pre>
      <p>把 <code>填写实际可用模型ID</code> 换成密钥可调用的模型，在刚才设置密钥的终端里进入项目目录，运行 <code>codex</code>。先发一条简短消息，再核对本站用量记录。用户级提供商配置不要放到项目的 <code>.codex/config.toml</code> 中。</p>
      <h3>Codex 桌面端</h3>
      <ol class="mofa-guide-steps">
        <li>从官方桌面应用说明安装，使用本机任务。</li>
        <li>登录本站，在 API 密钥页点击“使用密钥”，选择对应系统与 <strong>API Key Mode</strong>，按界面生成的配置合并用户级 <code>config.toml</code>，并下载模型目录放到界面指定的位置。</li>
        <li>完全退出并重启桌面应用，<strong>新建任务</strong>，在模型选择器中选择配置的模型，再发简短消息核对用量。</li>
      </ol>
      <p>生成的配置包含你的密钥，请仅保存在自己的本机。若使用上面的 <code>env_key</code> 方案，必须让桌面进程能读到 <code>MOFAMILY_API_KEY</code>；仅在一个终端设置临时变量不会自动传给已经运行的桌面应用。云端任务需按云端环境单独配置。</p>
      <p><a href="https://learn.chatgpt.com/docs/config-file/config-reference" target="_blank" rel="noopener noreferrer">Codex 配置参考</a> · <a href="https://learn.chatgpt.com/docs/config-file/config-advanced" target="_blank" rel="noopener noreferrer">自定义提供商示例</a> · <a href="https://learn.chatgpt.com/docs/app" target="_blank" rel="noopener noreferrer">桌面应用说明</a></p>
    </section>

    <section id="guide-others" class="mofa-guide-section">
      <h2>其他常用工具</h2>
      <p>客户端版本会改变设置名称；以下是协议选择原则，保存前核对工具展示的最终请求地址。</p>
      <div class="mofa-guide-tools">
        <article v-for="tool in tools" :key="tool.name">
          <h3>{{ tool.name }}</h3>
          <ol class="mofa-guide-steps"><li v-for="step in tool.steps" :key="step">{{ step }}</li></ol>
          <p class="mofa-guide-tool-fields"><strong>填写：</strong>{{ tool.fields }}</p>
          <details v-if="tool.example" class="mofa-guide-example"><summary>查看配置示例</summary><pre><code>{{ tool.example }}</code></pre></details>
          <a :href="tool.docs" target="_blank" rel="noopener noreferrer">官方配置文档</a>
        </article>
      </div>
      <p class="mofa-guide-caution">Cursor 的自带 API Key 功能只覆盖其文档列出的部分模型／功能，不等同于全功能 Agent 可走本站地址。若当前版本没有自定义服务地址入口，不要强行按 OpenAI 官方地址填写。<a href="https://docs.cursor.com/settings/api-keys" target="_blank" rel="noopener noreferrer">查看 Cursor 官方限制</a>。</p>
    </section>

    <section id="guide-troubleshoot" class="mofa-guide-section">
      <h2>连接失败时先查这些</h2>
      <dl class="mofa-guide-troubleshoot">
        <div><dt>401 · 身份验证失败</dt><dd>核对密钥是否完整、是否已停用，以及客户端究竟发送了 Bearer 还是 x-api-key。不要将密钥贴给客服或公开到日志。</dd></div>
        <div><dt>403 · 无权限</dt><dd>核对分组、模型授权、套餐范围和账户状态；若使用专属账号，还需检查该账号绑定状态。具体原因以响应错误码和服务端日志为准。</dd></div>
        <div><dt>404 / 不支持的接口</dt><dd>检查是否把 <code>/v1</code> 或 <code>/messages</code> 重复拼接；Codex 应走 Responses，不要让仅支持 Chat Completions 的工具直接请求 Responses。</dd></div>
        <div><dt>503 · 暂时不可用</dt><dd>检查服务健康与可用路由，记录时间、模型和请求 ID 联系站点管理员；仅凭状态码不能判断具体故障。</dd></div>
      </dl>
      <p>仍有疑问？<router-link :to="faqTarget">查看常见问题</router-link>。涉及个人密钥、账单或用量，请登录后在控制台核对。</p>
    </section>
  </component>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useAppStore } from '@/stores'
import PublicLayout from '@/custom/components/PublicLayout.vue'
import GuestAction from '@/custom/components/GuestAction.vue'

const props = withDefaults(defineProps<{ embedded?: boolean; hideTitle?: boolean; scope?: 'public' | 'console' | 'preview' }>(), { scope: 'public' })
const app = useAppStore()
const rootUrl = computed(() => (app.apiBaseUrl?.trim() || window.location.origin).replace(/\/+$/, '').replace(/\/v1$/, ''))
const apiUrl = computed(() => `${rootUrl.value}/v1`)
const plansTarget = computed(() => props.scope === 'preview' ? '/preview/plans' : props.scope === 'console' ? '/purchase?tab=subscription' : '/plans')
const faqTarget = computed(() => props.scope === 'preview' ? '/preview/faq' : props.scope === 'console' ? '/help/faq' : '/faq')
const claudeUnix = computed(() => `export ANTHROPIC_BASE_URL="${rootUrl.value}"
export ANTHROPIC_AUTH_TOKEN="$MOFAMILY_API_KEY"
claude`)
const claudePowerShell = computed(() => `$env:ANTHROPIC_BASE_URL = "${rootUrl.value}"
$env:ANTHROPIC_AUTH_TOKEN = $env:MOFAMILY_API_KEY
claude`)
const codexConfig = computed(() => `model_provider = "mofamily"
model = "填写实际可用模型ID"

[model_providers.mofamily]
name = "魔法家族"
base_url = "${apiUrl.value}"
env_key = "MOFAMILY_API_KEY"
wire_api = "responses"
supports_websockets = false`)

const tools = computed(() => [
  { name: 'Cherry Studio', steps: ['打开设置 → 模型服务 → 添加自定义服务，选择 OpenAI 类型。', '填写本站密钥与地址，启用服务，添加实际可用的模型 ID。', '新建助手并选择该模型，发送简短消息测试。'], fields: `服务地址先填 ${rootUrl.value}；客户端追加 /v1/chat/completions，保存前核对最终路径。`, example: '', docs: 'https://docs.cherry-ai.com/cherry-studio-wen-dang/en-us/pre-basic/providers/zi-ding-yi-fu-wu-shang' },
  { name: 'Cline', steps: ['在编辑器中打开 Cline 设置，将 API Provider 选为 OpenAI Compatible。', '填写 Base URL、本站 API Key 和实际模型 ID，保存。', '新建任务发送简短消息，确认调用正常后再处理项目。'], fields: `Base URL：${apiUrl.value}；模型需支持 Cline 操作所需的能力。`, example: '', docs: 'https://github.com/cline/cline/blob/main/docs/provider-config/openai-compatible.mdx' },
  { name: 'Continue', steps: ['打开 Continue 的配置入口，编辑本机 config.yaml。', '按示例添加模型，将模型与密钥占位符替换成自己的值；已有配置请合并 models 列表。', '保存后在聊天模型选择器选择魔法家族，发送简短消息。'], fields: `apiBase：${apiUrl.value}；本机配置包含密钥，不提交到仓库。`, example: `name: 魔法家族
version: 1.0.0
schema: v1
models:
  - name: 魔法家族
    provider: openai
    model: 填写实际可用模型ID
    apiBase: ${apiUrl.value}
    apiKey: 你的本站 API Key
    roles:
      - chat
      - edit`, docs: 'https://docs.continue.dev/reference' },
  { name: 'OpenCode', steps: ['运行 /connect → Other，将提供商 ID 填为 mofamily，并输入本站密钥。', '按当前 v1 文档将示例合并到 opencode.json，替换模型 ID。', '运行 /models，选择魔法家族的模型，发送简短消息测试。'], fields: `baseURL：${apiUrl.value}；示例使用 Chat Completions。Responses 模型按官方文档改用 @ai-sdk/openai，v2 配置请查对应版本文档。`, example: JSON.stringify({ $schema: 'https://opencode.ai/config.json', provider: { mofamily: { npm: '@ai-sdk/openai-compatible', name: '魔法家族', options: { baseURL: apiUrl.value }, models: { '填写实际可用模型ID': { name: '魔法家族模型' } } } } }, null, 2), docs: 'https://opencode.ai/docs/providers' },
  { name: 'Aider', steps: ['按接入前准备设置 MOFAMILY_API_KEY，进入项目目录。', '设置 OpenAI 兼容地址与密钥变量，再运行下面的模型启动命令。', '替换模型 ID，先发送简短消息核对调用。'], fields: '以下为 macOS / Linux 终端示例；PowerShell 对应使用 $env:OPENAI_API_BASE 和 $env:OPENAI_API_KEY。', example: `export OPENAI_API_BASE="${apiUrl.value}"
export OPENAI_API_KEY="$MOFAMILY_API_KEY"
aider --model openai/填写实际可用模型ID`, docs: 'https://aider.chat/docs/llms/openai-compat.html' }
])

onMounted(() => { void app.fetchPublicSettings().catch(() => {}) })
</script>
