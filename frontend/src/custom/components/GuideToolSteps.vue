<script setup lang="ts">
defineProps<{ tool: { name: string; steps: string[]; fields: string; example: string; docs: string } }>()
</script>

<template>
  <article class="mofa-guide-tool">
    <h3>{{ tool.name }}</h3>
    <ol class="mofa-guide-steps"><li v-for="step in tool.steps" :key="step">{{ step }}</li></ol>
    <p class="mofa-guide-tool-fields"><strong>配置核对：</strong>{{ tool.fields }}</p>
    <details v-if="tool.example" class="mofa-guide-example"><summary>查看配置示例</summary><pre><code>{{ tool.example }}</code></pre></details>
    <p><strong>完成后：</strong>新建会话，选择刚配置的提供商与模型，发送“只回复 OK”。收到回复后，到本站用量页核对时间与模型；一次测试也可能计入用量。</p>
    <a :href="tool.docs" target="_blank" rel="noopener noreferrer">官方配置文档</a>
  </article>
</template>
