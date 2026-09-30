<template>
  <div class="mofa-product-visual">
    <div v-if="product === 'api'" class="mofa-api-diagram" aria-hidden="true">
      <div class="mofa-api-orbit"><span class="mofa-api-node-mark" :style="{ '--mofa-node-mask': 'url(' + mark + ')' }"></span><span class="mofa-api-orbit-label">魔法 API</span></div>
      <span class="mofa-api-orbit-tag mofa-orbit-tools">你的 AI 工具</span>
      <span class="mofa-api-orbit-tag mofa-orbit-models">多模型接入</span>
      <span class="mofa-api-orbit-tag mofa-orbit-keys">API 密钥</span>
      <span class="mofa-api-orbit-tag mofa-orbit-usage">用量管理</span>
    </div>
    <figure v-else-if="product === 'recipes'" class="mofa-recipe-preview">
      <div class="mofa-recipe-sheet" aria-hidden="true">
        <div class="mofa-recipe-sheet-tags"><span>工作</span><span>大学</span><span>创作</span></div>
        <strong>大学论文提纲</strong>
        <p>先理清研究问题，<br>再搭起论证结构。</p>
        <dl><div><dt>输入</dt><dd>研究目标、已有资料</dd></div><div><dt>边界</dt><dd>结构要求、引用规范</dd></div></dl>
        <div class="mofa-recipe-sheet-action">整理并生成提示词<Icon name="arrowRight" size="sm" /></div>
      </div>
      <div class="mofa-recipe-output" aria-hidden="true"><Icon name="document" size="sm" /><span>可复制，也可连接模型</span><Icon name="check" size="sm" /></div>
      <figcaption>配方结构示意 · 非模型输出</figcaption>
    </figure>
    <figure v-else-if="product === 'studio'" class="mofa-studio-preview">
      <div class="mofa-studio-images">
        <img :src="studioBottle" alt="玻璃瓶原创图例" width="320" height="400" loading="lazy">
        <img :src="studioFox" alt="纸艺狐狸原创图例" width="320" height="400" loading="lazy">
      </div>
      <figcaption>原创视觉示例 · 非模板实测</figcaption>
    </figure>
    <figure v-else class="mofa-route-preview">
      <div class="mofa-route-map" aria-hidden="true"><span class="mofa-route-source"><FamilyProductMark product="synaroute" />SynaRoute</span><svg viewBox="0 0 360 110" fill="none"><path d="M180 0v40M45 98V55q0-15 15-15h240q15 0 15 15v43M180 40v58" stroke="currentColor" stroke-width="1.5" /><circle cx="180" cy="40" r="5" fill="currentColor" /></svg><div class="mofa-route-targets"><span><Icon name="terminal" size="md" />本地工具</span><span><Icon name="key" size="md" />API 配置</span><span><Icon name="arrowsUpDown" size="md" />模型路由</span></div></div>
      <figcaption>配置关系示意 · 不代表账号单点登录</figcaption>
    </figure>
  </div>
</template>

<script setup lang="ts">
import mark from '@/custom/assets/mofa-mark.webp'
import studioBottle from '@/custom/assets/studio-jade-bottle.webp'
import studioFox from '@/custom/assets/studio-paper-fox.webp'
import Icon from '@/components/icons/Icon.vue'
import FamilyProductMark from '@/custom/components/FamilyProductMark.vue'

defineProps<{ product: string }>()
</script>
