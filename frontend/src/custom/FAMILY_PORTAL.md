# 魔法家族统一首页

## 2026-10-01 最新交互

用户要求默认浅色，已保存深色仍保留；内置配方／工坊跟随主站当前主题。首页与控制台的本站产品默认进入 `/tools/recipes`、`/tools/studio`，登录后在控制台右侧正文展示，不打开新窗口。该入口复用原主站账号和认证守卫；直接访问公开 `/recipes/`、`/studio/` 仍可单独浏览，显式外部产品地址保留独立模式。

配方“模型设置”在宿主中直接下拉本人已有可用 Key、读取模型目录并选择接口格式，也能手填；点击应用只存所选连接，不运行模型。没有 Key 时引导原密钥管理，离开前提示当前材料可能丢失。用户主动运行才发送材料并可能计费，视频仍复制使用。独立静态页面没有宿主，保留原配置选择窗口，不宣称任意独立网站自动共享登录。

模块源码仍分别保留，`FamilyWorkspaceView` 和 `family-runtime/host-client.js` 是可替换的宿主接缝。当前 iframe 是可信同源内容，不能当成隔离第三方脚本的安全边界；未来独立域名需重新处理认证、精确来源、CORS与退出。

本轮预览4185显式使用 `vite --config vite.config.ts`；8080是否更新、测试与实际发布状态见 HANDOFF，不把新源码预览等同旧8080或生产已更新。下方旧默认深色／新窗口描述由本节覆盖。

## 产品范围与设计

本页是现有魔法 API 站点内的品牌官网，不是新账号系统。2026-09-30 用户否决账户摘要＋目录首页，要求恢复原品牌首屏，在下面介绍附属产品，并将登录后的产品切换放在 API 控制台顶部。保留既有狮冠 M、雾钛青、默认深色与原 BrandPanel 三条 CTA；首页 Persuade，控制台 Operate。样式只在 `custom/theme.css` 的专属作用域。

- `/family`：统一首页。原站默认 `/home` 与 `/brand` 也呈现此页；后台 `home_content` 自定义首页仍按原逻辑优先。
- `/api`：保留原 API 介绍、游客预览、套餐、接入教程与账户操作。
- 官网导航与控制台个人区都有“家族首页”。普通登录／注册不携带 `/family` 回跳，沿用原默认 API 控制台；没有修改全局登录处理。
- 匿名产品入口带站内 `/dashboard?product=<已知标识>`，先登录进入 API 控制台，顶部显示显式继续打开按钮。查询参数不是外部 URL，未知／重复参数不消费；不自动弹窗、授权或创建 Key。
- 控制台 AppHeader 装配顶部家族胶囊：API 控制台、魔法配方、魔法工坊、SynaRoute。宽屏居中，小桌面另起一行，手机横向滚动，不移除原账户／公告／移动菜单。
- 魔法 API 进入既有控制台；魔法配方／魔法工坊新窗口打开；SynaRoute 进入已有密钥页使用用户主动导入。工坊内容首版已完成，没有发布地址时禁用，不再标“仅规划”。
- 视觉图例保留工坊原玻璃瓶／狐狸 WebP，并新增成年女性 AI 参考人像，三张等宽展示；来源及准确产品边界见 `family/PRODUCT.md`、`family/SURFACE.md`，图片不是模板效果实测。

## 共用登录的准确边界

站内页面复用 `authStore` 当前会话。魔法配方没有另一套账号密码；连接模型时打开原站受保护 `/connect/recipes`，当前会话有效则无需重新输入密码，用户仍须选择已有密钥和模型、明确授权导入。首页不读取密钥、不自动授权或创建密钥、不调用模型，也不把 JWT、刷新令牌或 Key 写入产品链接。

SynaRoute 当前仅密钥导入，不等于桌面账号单点登录。尚未实现独立产品通用 OIDC／SSO 与跨产品单点退出；原站退出后配方内已授权 Key 仍存在于其当前页内存，需清除连接或刷新。会话过期、不同原站来源或浏览器隔离策略可能要求重新登录。真实账号与跨窗口闭环必须独立验收，组件模拟测试不能替代。

家族首页SynaRoute展台提供“下载客户端”与“访问官网”，未登录也可打开。链接集中在family/synarouteLinks.ts，来源为SynaRoute官方工程的站点配置和下载路由，不携带登录令牌／Key，不自动下载安装；“配置 SynaRoute”仍通过原站登录与密钥页。本轮不修改原导入逻辑或新增安装检测，浏览器焦点不能可靠证明客户端是否安装。官网下载页实时可达性与实际安装包／协议导入须另行验证。

配方页底部可在新窗口回到家族首页，保留当前页面材料。仅本机开发入口携带公开的 `api_site=<本站 origin>`，配方以既有官方／本机白名单核验后预填原站地址；恶意提示不能改写登录目标，不含任何凭据，也不会自动打开登录窗口或请求模型。

本机来源提示只能用于当前配方也位于本机 HTTP 的预览；公开 HTTPS 配方不能通过链接提示自动切换到本机登录目标。手动本机配置仍按原流程由用户选择。

## 发布配置

2026-10-01 用户确认一体化部署：默认一个主站镜像提供 `/recipes/` 和 `/studio/`，未配置外部地址时标为“本站产品”，不再禁用或自动跳到独立开发端口。Vite 定制插件从权威源码打包，Go embed 以独立静态白名单和 CSP 提供；部署／安全／本机迁移见 `deploy/FAMILY_INTEGRATED.md`。

配方内置版本自动使用当前站点登录配置选择与家族首页，不受 `api_site` 查询参数影响。主站登录有效即可复用会话，但选择密钥及模型仍需主动授权，工坊无需第二套登录。以下变量只在选择外部独立托管时需要，空值使用实际内置产品，不要求创建变量。

本机 Docker 仍可显式配置独立 HTTP 配方地址，只有两端均为回环 HTTP 时接受并标为“本机预览”；该兼容选项不是默认部署方案，远程 HTTP 仍拒绝。旧独立端口说明见 `deploy/FAMILY_LOCAL_DOCKER.md`，当前默认使用一体化文档。

| 构建变量 | 用途 | 值的形式 |
| --- | --- | --- |
| `VITE_MAGIC_RECIPES_URL` | 产品打开地址 | 已部署的 HTTPS 完整 URL 或同源 `/recipes/`，不得含凭据、查询参数或片段 |
| `VITE_MAGIC_RECIPES_ORIGIN` | 配置回传接收来源许可 | 独立配方的精确 origin（协议＋域名＋端口），不得含路径或通配；同源可留空 |
| `VITE_MAGIC_STUDIO_URL` | 工坊打开地址 | 与配方一样的安全地址校验，不携带 api_site 或任何账号数据 |

Dockerfile 的前端构建 stage 已新增三个对应 ARG，`custom-image.yml` 从仓库 Actions Variables 的 `MAGIC_RECIPES_URL`／`MAGIC_RECIPES_ORIGIN`／`MAGIC_STUDIO_URL` 传入。源码接线不等于已设置仓库变量、运行 CI 或上线；这些是公开前端配置，绝不能填密钥。变量在构建时写入前端，不能给已构建容器加运行时环境变量冒充生效。2026-10-01本机8080已改为镜像内置两个产品，不依赖4178／4179；旧独立服务仅为兼容预览或回退，不作为当前默认方案。

### 在哪里配置（GitHub 镜像构建）

默认一体化部署无需配置这些地址。只有选择独立托管时才按以下步骤操作；不是魔法 API 管理后台的设置项，也不是运行容器的 `.env`：

1. 确认确实要覆盖内置产品，再独立发布静态页并核对实际 HTTPS 地址。配方独立发布说明见 `product-samples/magic-recipes/site/README.md`，工坊说明见 `product-samples/magic-studio/README.md`。
2. 打开 `EngineeMoMo/sub2apicust` 仓库 → **Settings → Secrets and variables → Actions → Variables → New repository variable**。创建下面三个 **Repository Variables**，不是 Secrets，也不需要在变量名前加 `VITE_`。[GitHub 官方操作说明](https://docs.github.com/en/actions/how-tos/write-workflows/choose-what-workflows-do/use-variables#creating-configuration-variables-for-a-repository)。

   | 仓库变量名 | 示例值（仅示意，须换成实际已部署地址） |
   | --- | --- |
   | `MAGIC_RECIPES_URL` | `https://recipes.example.com/` |
   | `MAGIC_RECIPES_ORIGIN` | `https://recipes.example.com`（无路径；同源配方可留空） |
   | `MAGIC_STUDIO_URL` | `https://studio.example.com/` |

3. 独立来源的配方还需在 API 部署配置里添加精确 CORS origin，保留既有许可项；不能用 `*` 代替。仅配置工坊公开浏览地址不需要模型授权 CORS。
4. 打开仓库 **Actions → Build custom image (GHCR) → Run workflow**，选择需要发布的分支并重新构建。只保存变量或重启旧容器不会生效；手动重建同一提交可能沿用同一 SHA 标签，应确认新的构建成功并拉取新镜像。
5. 用户备份后更新 API 镜像，再检查首页与登录后控制台中的配方／工坊入口。未提交的本地配图修改不会被远端构建包含；真实登录与配方模型授权仍需单独验收。

默认镜像在构建时直接生成配方与工坊，不使用本机已生成的 `site/index.html`。仅选择独立来源覆盖时还须保留既有来源并配置 API CORS，构建来源许可与 API CORS 是两层检查。内置路径由 Go 静态处理，不交给主站 SPA 兜底。不要放松认证、窗口安全头或生产来源限制。

## 验收与维护

- 单元／组件：`family-products.spec.ts`、`family-home.spec.ts`、`family-switcher.spec.ts`、`family-header.spec.ts`、`guest-session.spec.ts`、`upgrade-contract.spec.ts`；路由守卫 `feature-access.spec.ts`，配方 `ui.test.cjs`。
- 检查匿名、用户、管理员与退出状态，空配置使用内置路径、错误显式地址禁用，原 API 介绍及后台模式限制，桌面／手机与双主题。
- 用户自行真实登录后核验产品切换、配方已有 Key／模型选择与显式授权；不替用户生成收费请求或开启桌面导入。
- 上游接缝为 `router/index.ts` 的公开模式等待、`AppSidebar.vue` 的个人区菜单、`AppHeader.vue` 的组件及 import、Dockerfile 的三个构建 ARG／RUN 与 workflow 的三个 Variables 接线，均有 `[CUSTOM]` 标记。同步时保留原守卫、原菜单、账户操作和构建逻辑，不能覆盖其他定制。
- 准确测试结果、当前8080镜像、备份和未验项见根 `HANDOFF.md` 最新状态；本机更新不等于生产发布，没有提交推送。
