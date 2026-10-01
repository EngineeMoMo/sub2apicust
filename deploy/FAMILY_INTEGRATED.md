# 魔法家族一体化部署

2026-10-01 用户授权配方与工坊随主站一起部署，并更新本机 8080。本方案覆盖旧文档的“必须另行托管”：一个应用镜像同时提供主站、`/recipes/`、`/studio/`，数据库与 Redis 保持原样，无需额外配方／工坊服务或新账号。

## 构建与产品边界

- Vite 定制插件在开发及生产构建前从 `product-samples` 权威源码生成 `frontend/public/recipes`、`frontend/public/studio`；生成目录有标记且走 gitignore，不提交或复制本机已有生成产物到 Docker。
- 配方外置 JS／CSS，保留惰性的 JSON 数据块，不需要放宽全站 CSP 或启用执行脚本的 `unsafe-inline`。工坊只复制运行模块、WebP、mp4 和品牌 PNG，不发布原始生成 PNG、来源记录、教程 Markdown、测试或服务端脚本。
- Go 的 embed 前端服务在主 SPA 兜底前处理两个命名空间，GET／HEAD 可用、其他方法 405、未知文件 404；无目录列举或路径穿越，JS 模块 MIME 明确，产品资源 `no-store`，原 API 路由和主站设置注入不变。镜像构建运行 `go test -tags embed -run '^TestFamily' -v ./internal/web`，只执行本轮新增路由测试，不宣称全量后端测试通过；测试源码不会复制到最终运行镜像。
- 本机首次构建额外尝试完整 web 测试，旧静态文件用例因 `/logo.png` 不存在失败：当前已提交的 `frontend/public` 只有 `logo.svg`，旧用例未被本轮修改。本轮不恢复旧品牌图或修改无关用例；失败日志保留为 `output/family-integrated-20261001/docker-build.log`，限定回归的重新构建另存日志。
- 主站产品入口默认进入受原认证守卫保护的 `/tools/recipes`、`/tools/studio`，在控制台正文加载 `/recipes/?embedded=1`、`/studio/?embedded=1`。子产品实时跟随主站主题，无偏好默认浅色；直接打开公开静态路径仍支持独立模式。
- 配方在宿主内直接读取本人已有 Key 和模型目录、选择并应用；独立静态模式仍使用 `/connect/recipes` 选择窗口。只回传所选连接，不把 JWT／刷新令牌写入消息或URL，不自动生成请求。宿主退出时清除内嵌连接及未完成请求，不等于撤销 API Key或跨产品全局单点退出。
- 子产品安全头仅允许同源嵌入（SAMEORIGIN、`frame-ancestors 'self'`），它们是可信模块，不是沙箱。工坊额外打包 mp4 与两个运行模块，`media-src 'self'`、MIME和 Range206 定向回归；主站全局认证／安全头保持。
- 工坊视频是署名的 CC BY 授权原片节选，有实际播放，不是本站模型生成或模板实测。工坊不新增生成模型、自动安装 Skills 或定时任务；付费模型效果与计费仍由用户主动实测。

## 默认部署，不必配置产品地址

新版本不设置 `MAGIC_RECIPES_URL`／`MAGIC_RECIPES_ORIGIN`／`MAGIC_STUDIO_URL` 时，首页和控制台默认打开同源工作区 `/tools/recipes`／`/tools/studio`；对应静态产品随镜像内置。同源模式不需要额外跨域来源。使用同一个协议、域名与端口；不要在 8080 登录后改用 4184 或 localhost 的另一个来源来检验共享登录。

保留独立产品源码与显式外部覆盖选项：确实要另行托管时，再配置 Actions Variables 的安全 HTTPS 地址、精确配方 origin 和 API CORS，然后重建镜像。错误的显式配置仍禁用，不静默回退。不能给旧容器增加环境变量冒充构建变量已生效。

生产仍由用户备份后更新成功 CI／GHCR 对应的固定镜像标签；本轮没有生产发布或自动提交推送。若前面先使用旧功能 `sha-25071d4`，它不包含本次一体化代码；应等待本次另行提交和构建成功后的新标签。

## 本机 8080 升级

工坊投稿审核版无需新服务或迁移，但必须保留原app_data卷，并备份 `pricing.data_dir/studio-submissions` 全目录（默认 `/app/data/studio-submissions`）；Postgres备份不含用户上传文件。前后端同版本部署后，投稿入口 `/tools/studio/submit`、管理员 `/admin/studio/submissions`、公开目录 `/api/v1/studio/gallery` 才形成闭环。反代上传上限须容纳30MiB＋表单；不要公开映射私有目录。当前仅单应用实例文件存储，多副本需先改存储层，详见工坊COMMUNITY.md。隔离测试镜像不等于8080已经升级。

先按 `LOCAL_DOCKER_RUNBOOK.md` 核对原项目、代理、数据库备份与回退镜像。当前本机曾使用旧 `docker-compose.family-local.yml.example` 传入精确 CORS；迁移时保留该叠加层的已有许可，最后加一体化层覆盖旧构建地址，只 build／up 原应用服务，旧配方容器不重建、不删除。

```powershell
docker compose -p sub2apicust-local --env-file .env -f docker-compose.override.yml -f docker-compose.family-local.yml.example -f docker-compose.family-integrated.yml.example build sub2api
docker compose -p sub2apicust-local --env-file .env -f docker-compose.override.yml -f docker-compose.family-local.yml.example -f docker-compose.family-integrated.yml.example up -d --no-deps --no-build --wait --wait-timeout 120 sub2api
```

以上在 deploy 目录执行，构建进程需要 runbook 的临时代理。其他环境不必增加旧独立叠加层；不要覆盖其原有 CORS。无需 `down`、删卷、重建数据库／Redis或给配方增加独立端口。

更新后验证镜像与容器 ID 一致、原数据库／Redis ID 与 app_data 卷保持、业务计数和设置摘要保持；检查 `/health`、`/family`、`/recipes/`、`/studio/`、相对静态资源与 GET／HEAD／404／405／CSP，个人接口仍匿名 401。真实账号授权、模型权限及计费不由静态页面或模拟测试代替。准确备份／镜像／未验项见 HANDOFF。
