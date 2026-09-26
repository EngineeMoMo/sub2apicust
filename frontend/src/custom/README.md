# `custom/` —— 定制叠加层（换肤 + 新增/替换页面）

本目录是本 fork 的**前端定制专区**，用于降低与上游的文本冲突面。
不能保证未来零冲突或零维护：路由、store、公共类和插槽仍依赖上游契约。
规则见项目根 `CUSTOMIZATIONS.md`，逐次同步核对见 [`UPGRADE.md`](./UPGRADE.md)。

## 三种定制怎么放

### 1. 换肤（改颜色/字体）
改 [`theme.css`](./theme.css) 里的 `--color-primary-*` 变量即可，全站品牌色跟着变。
`tailwind.config.js` 的 `primary` 色阶已改为引用这些变量（`main.ts` 里已 `import './custom/theme.css'`）。

### 2. 新增页面
- 页面放 `views/`（如 `views/MyPageView.vue`）。
- 在 [`routes.ts`](./routes.ts) 里加一条路由。
- 路由已在 `src/router/index.ts` 里 `...customRoutes` 展开（404 兜底前），无需再动 router。

### 3. 整页/组件「影子替换」上游（不改上游文件）
想完全重写某个上游页面但不想每次 merge 冲突：
1. 把你的版本放到 `views/`（用**同名**，如 `views/HomeView.vue`）。
2. 在 `vite.config.ts` 的 alias 数组里，`'@'` 那条**之前**加一行（已有注释示例）：
   ```ts
   { find: /^@\/views\/HomeView\.vue$/, replacement: resolve(__dirname, 'src/custom/views/HomeView.vue') },
   ```
   上游 `src/views/HomeView.vue` 一个字不用改，构建时自动加载你的版本。

## 🔴 影子替换的代价（务必记住）
被你替换的页面仍然依赖上游的 **Pinia store / API / 路由 meta**。上游改了这些契约，
你这页得**你自己跟着改**（失效是静默的：页面还在、运行时才报错）。所以：
- 只对「真想完全掌控」的页面用影子替换，不要全站都换；
- 每次 merge 上游后，**逐个打开你替换过的页面**冒烟验证；
- 每加一个替换/新增，去 `CUSTOMIZATIONS.md` 登记。

## 接缝文件（每次 merge 都需核对）
`tailwind.config.js`（引用变量）、`vite.config.ts`（alias 数组）、
`src/main.ts`（import theme.css）、`src/router/index.ts`（展开 customRoutes）。
都打了 `[CUSTOM]` 标记，搜这个词即可定位。

雾钛青增加 `AuthLayout.vue`、`AppLayout.vue`、`AppHeader.vue` 三个最小展示接缝。
独立 `/brand` 复用原路由入口，不替换 `/`，不复制业务页；全部视觉规则放 `theme.css`。
品牌名称和站点图标仍读取公开设置，M 展示素材为原批准位图，不是新绘制矢量。
