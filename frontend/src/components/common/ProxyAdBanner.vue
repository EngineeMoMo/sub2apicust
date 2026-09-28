<script lang="ts">
import { defineComponent } from 'vue'

// [CUSTOM] 禁用上游代理购买广告，保留不渲染内容的组件兼容原有三处调用。
export default defineComponent({
  name: 'ProxyAdBanner',
  render: () => null
})
</script>
