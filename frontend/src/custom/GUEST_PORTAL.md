# 游客官网维护说明

## 帮助分类与图标（2026-10-08，源码未发布）

发布前用户补充：教程准备段和API接入FAQ同时说明本站支持生图（图像生成），需要联系管理员开通权限；FAQ从16增至17题。服务说明来自用户明确要求，没有自动开启权限。

- 控制台FAQ与教程此前都用ChannelIcon，改为公共图标questionCircle／book；游客预览同步区别。上游仅AppSidebar与router锚点白名单两个接缝，样式仍只在theme.css。
- 教程分六节：接入准备、命令行工具、桌面应用、编辑器扩展、验证接入、排错反馈。原guide-claude／codex／others等锚点保留，新增分类锚点；GuideToolSteps维护通用客户端步骤。模型ID仍是占位符，地址仍来自公开设置，不读取密钥或发模型请求。
- FAQ保留四类，新增协议、桌面环境、接通确认与密钥泄露问题，共16题；分类与关键词组合，空结果可清除两项筛选。
- 官方资料核对（2026-10-08）：[Codex配置参考](https://learn.chatgpt.com/docs/config-file/config-reference)确认用户级提供商、env_key与Responses字段；[Claude桌面环境](https://code.claude.com/docs/en/desktop#local-sessions)确认本机环境编辑入口及Windows不读PowerShell profile；[Cline](https://docs.cline.bot/provider-config/openai-compatible)、[Continue](https://docs.continue.dev/reference)、[OpenCode](https://opencode.ai/docs/providers)、[Aider](https://aider.chat/docs/llms/openai-compat.html)各自配置说明；[Cherry Studio自定义服务商](https://docs.cherryai.com.cn/pre-basic/providers/zi-ding-yi-fu-wu-shang)为旧链接迁移后的入口。Cursor旧API Key链接跳转至[文档首页](https://cursor.com/docs)，仅保留版本能力核对提示，不承诺本站兼容其全部功能。
- 56项组件回归、相关ESLint、类型／生产构建通过，日志output/help-{tests,lint,build}.log。Edge使用隔离上下文及公开设置夹具检查375／1440浅深色四组合，无横向溢出／页面异常，六节目录跳转与FAQ分类通过；截图及visual.json在output/help-20261008。初次浏览器夹具误拦Vite的src/api模块，修正为仅/api/接口后通过，未改业务接口。初次构建缺预览图标联合类型，已补后通过。
- 未认证真实控制台、未实际操作第三方客户端或调用模型，未重建业务镜像／部署；生产与真实客户端验收仍待。

## 接入教程（2026-09-30，本地完成未发布）

- 公开 `/guide`、认证后 `/help/guide` 和预览 `/preview/guide` 共用PublicGuideView；控制台用ConsoleGuideView保留AppLayout，预览只加载公开设置。
- 官网／首页与FAQ新增教程入口。scope与guideTarget保证教程和FAQ在控制台、预览中右侧往返；密钥操作使用原GuestAction登录提示，不请求个人账户接口。
- 示例地址取公开apiBaseUrl；Claude Code使用根地址，Codex／OpenAI兼容客户端使用`/v1`。CLI密钥变量、桌面本机环境、模型目录与第三方配置示例均以当前官方文档和实际分组权限为准。
- 80项前端定向、国际化3项、类型／lint／构建与4175浏览器验收通过；实际客户端请求、认证后控制台数据验收及部署仍待，证据见HANDOFF。router限定教程锚点滚动接缝需随上游同步核对。

## 2026-09-30 待发布修复

- **2026-09-30 已部署本机Docker（优先于本轮未部署记录）**：用户授权本机测试。原三容器处于停止状态，保留原卷启动Postgres／Redis，先完成597905字节数据库备份与pg_restore目录校验、旧镜像标签sub2apicust:before-console-20260930，再代理预检200并构建。前端类型／国际化3测／构建及Go embed成功；仅应用容器重建，运行镜像与标签一致be2f26f98e8c3dee62d765a8b78e6cd33021dc3b6f618c76085e310180b234b8，三服务healthy，health=ok。迁移242两列和三个索引确认存在；数据库／Redis容器ID及app_data卷不变，用户2／密钥0／组3／套餐1／订单0／包号0／设置280计数保持。备份逐项比较只有Claude版本2.1.283→2.1.285与Codex版本0.158.0→0.159.2两键自动同步，09:44启动日志证实，其余278键不变。HTTP入口index-Ckx_MqIV.js及三个新页面chunk验证通过，账户与管理员接口匿名401；浏览器8080官网正常打开并留给用户测试，尚未做带数据管理员交互验收。证据output/console-deploy-20260930，备份SHA256=1D900EC14FB9915B5E912F0A84778271CEB149178B7CC526E3A9F55411EE037E。未提交推送或操作生产；迁移后的回退需兼顾数据库，不能直接跑旧单用户版本。

- 官网导航和首页主CTA的游客入口只在未登录时显示；退出登录恢复。不开放实际账户路由，也不强制禁止登录用户直接访问公开预览URL。
- 控制台侧栏FAQ指向requiresAuth的 `/help/faq`，ConsoleFaqView保留AppLayout与侧栏并内嵌PublicFaqView；隐藏内层重复h1。官网 `/faq` 和预览 `/preview/faq` 仍公开。
- guest-session测试覆盖登录态变化及FAQ不出现官网壳、只有一个main、搜索空态；本轮139前端定向回归通过。尚未部署本轮源码，不用8080旧页面验收。

**发布状态（2026-09-28）**：代码fa45b2023已推送，CI／安全扫描／GHCR全绿，包括新增Go单测与集成测试；发布镜像sha-fa45b20。先前待Go验证状态仅为历史。本机8080已运行同一业务功能，后续修正仅测试初始化与交接文档；生产由用户备份后执行./update.sh sha-fa45b20，不自动部署或下单。完整登录支付、生产首页配置仍待现场验收。

## 游客控制台预览（2026-09-28新增）

- **宽屏修正已上线本机**：工作区不得再设置1280px固定上限；页面按窗口使用对称弹性边距，文字仍保持阅读宽度，字段在手机分两列。320–2560px七档检查通过；8080镜像ea6e3a5ecc3c实测2031px左右边距均40.609px。原「无溢出」验收未覆盖留白比例，本次补上几何和截图检查，视觉仍待用户确认。

- **后续授权部署已完成**：本机8080运行ecb3961845e8，前端及Go embed编译成功，三服务健康，数据计数／设置摘要与备份前一致。浏览器实测首页→预览、密钥登录提示取消和真实套餐；下方「尚未进入Docker」为源码阶段历史。证据output/preview-20260928-175007，未操作生产或付款。

- 首页主CTA「游客预览控制台」及公共导航「游客预览」进入 `/preview`，`/preview/:section?` 展示八项菜单；个人功能为明确标注的只读说明，不使用示例余额、密钥或订单冒充数据。
- 独立GuestPreviewView不挂载AppLayout/AppHeader/AppSidebar及账户业务页面；仅读取公开设置，套餐子页才读取公开套餐。PublicPlansView与PublicFaqView的embedded模式保留侧栏且不重复公共页壳。
- 个人功能点击操作后使用GuestAction弹登录提示，可取消；真正账户路由仍受原登录守卫保护。后台模式同样拦截预览入口，不绕过管理员配置。
- 手机菜单使用内联折叠、aria-expanded与aria-controls，导航后收起并聚焦正文；继承现有深浅主题，仅theme.css新增样式。
- 本轮179项定向测试、3项国际化测试、类型检查、相关ESLint和构建通过。4175真实浏览器验证主入口、登录提示取消、真实套餐、FAQ空状态、手机菜单和桌面／手机宽度。8080仍是上一版游客官网镜像，不含本节预览；未部署本轮变更。

## 页面与配置

- `/` 仍重定向 `/home`。正常站点模式且 `home_content` 为空时显示游客官网；自定义 HTML／URL 首页保留最高优先级，`compact_home_enabled` 不再覆盖游客官网。
- `/brand` 使用同一品牌组件；`/plans` 和 `/faq` 为公开路由。后台模式仍限制匿名及非管理员访问，管理员诊断行为不变。
- 站点名、Logo、注册开关、模型广场开关沿用公开设置。主题仍默认深色，样式仅在 `theme.css`。
- FAQ 数据位于 `guest/faq.ts`，首版四组十二题，修改后运行游客测试；不要添加未经确认的退款、到账时效或可用率承诺。

## 公共接口契约

`GET /api/v1/payment/public/plans` 无需认证，沿用标准 `{code, message, data}` 响应，`data` 为 `{plans, purchase_enabled}`。

- 返回在售套餐，按现有 `sort_order` 排序；字段白名单：id、name、description、price、original_price、currency、validity_days、validity_unit、features、daily_limit_usd、weekly_limit_usd、monthly_limit_usd、supported_model_scopes。
- 不返回分组ID、订单、余额、支付渠道配置、密钥或用户信息；不直接暴露 ent 模型。权益数组解析失败时返回空数组，未知额度省略，页面不得解释为无限。
- 后台模式或订阅关闭返回404；设置或目录读取失败返回错误，不能伪装成无在售套餐。支付关闭仍允许读目录，`purchase_enabled=false`。
- 前端使用独立 Axios 客户端，不使用账户401刷新／回跳拦截器；失败在原页展示。套餐价与币种展示沿用现有卡片口径，空币种按USD展示，最终实付以后端结账为准。

## 选购与认证

- 游客点击账户功能出现可取消的登录提示；选择套餐目标为 `/purchase?tab=subscription&plan=<id>`，充值使用现有 `/purchase`。
- 密码／2FA／Passkey登录、普通注册与邮箱验证保留站内目标；已有OAuth组件继续沿用其现有回跳流程，不新增OAuth业务。
- 购买页重新读取结账套餐后按严格正整数ID匹配；未知或下架显示提示。支付恢复及已有分组续费优先，浏览、登录、选中套餐均不自动下单。
- 所有个人接口、订单创建及结账接口仍要求登录；不得把 `requiresAuth` 批量删除。

## 验证与发布

- 本轮163项定向前端测试通过，覆盖游客页面／请求隔离、路由守卫、套餐选中与续费优先级、注册回跳以及既有认证／品牌回归；类型检查、相关ESLint和前端生产构建通过。它们不等同真实支付端到端验证。
- 新增后端 `custom_public_plans_test.go` 覆盖白名单、在售过滤、排序、缺失额度、支付关闭、后台模式与订阅关闭。本机未安装Go，但2026-09-28已通过Docker内Go embed编译；后端unit测试仍未运行，未来提交后必须查本次CI，不引用旧结果。
- 2026-09-28用户授权本机Docker部署后，8080已运行新镜像`e7a582285cbd`，此前4175仅前端预览和目录404记录不再代表当前状态。真实目录返回既有1个在售套餐，字段白名单与在售数量符合数据库，4个个人接口匿名401，未添加虚构套餐。
- 浏览器实测官网与FAQ、搜索、登录提示及取消；1280桌面和390手机、深浅色检查无外层横向溢出。截图在本轮会话visualizations的 `guest-home-desktop-dark.png` 和 `guest-faq-mobile-light.png`。
- 上线前检查目标环境 `home_content` 是否遮挡官网，使用真实在售数据验证选购→登录／注册→选中套餐与确认下单。配置仍独立保存，不自动修改生产。
- 本机Docker重建必须先读 `deploy/LOCAL_DOCKER_RUNBOOK.md`；本轮已按后续授权构建并仅重建应用容器。三服务healthy，数据库／Redis容器及数据卷保留，用户／套餐／订单计数不变；280个设置仅自动版本同步键变化（日志及备份逐项比对证实）。证据和备份在output/guest-20260928-171257；未提交推送、未部署生产、未下单扣款。
- 部署后浏览器实测8080首页、FAQ搜索、真实套餐展示、购买→登录携带plan=1及390px手机宽度；完整登录后购买与支付恢复仍待验收。套餐手机截图为output/guest-20260928-171257/plans-mobile.png。
