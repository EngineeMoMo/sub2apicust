<template>
  <component :is="embedded ? 'div' : PublicLayout">
    <section class="mofa-public-intro">
      <h1>常见问题</h1>
      <p class="mofa-public-lead">从了解服务到接入工具，先把基础问题弄清楚。</p>
      <label for="faq-search">搜索问题或关键词</label>
      <input id="faq-search" v-model="query" class="input mofa-faq-search" type="search" placeholder="例如：充值、API Key、额度" aria-controls="faq-results">
      <p role="status">找到 {{ results.length }} 个问题</p>
    </section>
    <div id="faq-results" class="mofa-faq-results">
      <section v-for="group in groups" :key="group.name" class="mofa-faq-group">
        <h2>{{ group.name }}</h2>
        <details v-for="item in group.items" :id="item.id" :key="item.id" :open="query.trim().length > 0">
          <summary>{{ item.question }}</summary>
          <p>{{ item.answer }}</p>
        </details>
      </section>
      <div v-if="!results.length" class="mofa-public-state"><h2>没有找到相关问题</h2><p>试试“充值”“密钥”或“模型”等关键词。</p><button type="button" class="btn btn-secondary" @click="query = ''">查看全部问题</button></div>
    </div>
  </component>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import PublicLayout from '@/custom/components/PublicLayout.vue'
import { faqGroups, searchFaq } from '@/custom/guest/faq'
defineProps<{ embedded?: boolean }>()
const query = ref('')
const results = computed(() => searchFaq(query.value))
const groups = computed(() => faqGroups.map(name => ({ name, items: results.value.filter(item => item.group === name) })).filter(group => group.items.length))
</script>
