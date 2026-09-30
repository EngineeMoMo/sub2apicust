<template>
  <PublicLayout>
    <h1 class="sr-only">{{ app.siteName }} · 多模型 API 服务</h1>
    <BrandPanel :site-name="app.siteName">
      <template #actions>
        <router-link to="/plans" class="btn btn-primary">先看看订阅套餐</router-link>
        <router-link v-if="!auth.isAuthenticated" to="/preview" class="btn btn-secondary">游客预览控制台</router-link>
        <router-link to="/faq" class="btn btn-secondary">了解如何使用</router-link>
      </template>
    </BrandPanel>
    <section class="mofa-public-section" aria-labelledby="access-title">
      <h2 id="access-title">先了解服务，再决定如何接入</h2>
      <p class="mofa-public-lead">这里提供多模型 API 接入服务。你可以先比较套餐、阅读接入说明；购买、充值和管理 API Key 时再登录。</p>
      <div class="mofa-public-columns">
        <article>
          <h3>接入你的工具</h3>
          <p>登录后创建 API Key，在支持相应接口的客户端中配置服务地址与密钥。模型和可用范围以实际套餐及控制台为准。</p>
          <GuestAction to="/keys">管理 API Key</GuestAction>
        </article>
        <article>
          <h3>按需使用与管理</h3>
          <p>在账户中查看用量、余额和订单。余额充值与订阅套餐是不同的购买方式，购买前请先阅读套餐权益。</p>
          <div class="mofa-public-actions">
            <GuestAction to="/usage">查看用量</GuestAction>
            <GuestAction v-if="app.cachedPublicSettings?.payment_enabled" to="/purchase" message="充值需要关联你的账户。请先登录，再确认金额与支付方式。">账户充值</GuestAction>
            <GuestAction to="/orders">我的订单</GuestAction>
          </div>
        </article>
      </div>
    </section>
  </PublicLayout>
</template>

<script setup lang="ts">
import { useAppStore, useAuthStore } from '@/stores'
import BrandPanel from '@/custom/components/BrandPanel.vue'
import PublicLayout from '@/custom/components/PublicLayout.vue'
import GuestAction from '@/custom/components/GuestAction.vue'
const app = useAppStore()
const auth = useAuthStore()
</script>
