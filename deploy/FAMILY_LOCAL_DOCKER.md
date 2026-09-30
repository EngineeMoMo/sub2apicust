# 魔法家族与配方的本机 Docker 测试

本文件只用于当前 Windows 本机，不是生产部署方案。主站品牌首页在 `http://127.0.0.1:8080/home`（`/family` 同页），配方使用独立静态容器 `http://127.0.0.1:4178/`；两者仅绑定本机回环接口，复用原 PostgreSQL、Redis 与应用数据卷，不创建新账号体系。工坊入口显式连接已有 `http://127.0.0.1:4179/` Node 预览，不是本轮 Docker 服务，需该独立预览仍在运行。

## 准备与构建

先读 `LOCAL_DOCKER_RUNBOOK.md`，确认本机 Docker context、代理与原数据卷，保留旧镜像及经过 `pg_restore --list` 校验的数据库备份。不要删除卷、执行 `down -v` 或将本机代理带到生产。

在仓库根生成配方单文件：

```powershell
node product-samples/magic-recipes/site/build.mjs
```

在 deploy 目录使用原 gitignored 本机配置，加本仓不含凭据的局部示例叠加层：

```powershell
docker compose -p sub2apicust-local --env-file .env -f docker-compose.override.yml -f docker-compose.family-local.yml.example build sub2api recipes
```

构建过程必须按 runbook 给实际进程设置已验证代理。`recipes` 的 Dockerfile 只复制生成的 `index.html` 和静态服务，不含测试、源码材料或真实密钥。Node 服务默认仍监听本机；容器通过 HOST／PORT 显式使用内部监听地址，Compose 对外只发布回环4178，非 root、只读文件系统、无新权限且具健康检查。

叠加层注入公开的 `VITE_MAGIC_RECIPES_URL`／`VITE_MAGIC_RECIPES_ORIGIN`／`VITE_MAGIC_STUDIO_URL`，只在本机 HTTP 主站显式允许本机 HTTP 独立产品，并标明“本机预览”；空配置的构建不会自动开放4178／4179。远程 HTTP 仍拒绝，生产继续使用已部署 HTTPS 或同源路径。URL 不含 JWT、密码或 API Key，配方 `api_site` 只有本机主站 origin，工坊无来源提示。

`CORS_ALLOWED_ORIGINS` 仅允许127.0.0.1／localhost的4178。本次现场读取原环境及配置没有 CORS 来源，预检403，故新增这两个精确来源；其他环境应保留原有来源后再追加，不可直接覆盖已有白名单。未修改生产配置或使用通配来源。

## 启动与验收

4178 如已有本线程启动的 Node 预览，先停止该进程；不杀未知占用者、不刷新用户填写中的页面。构建成功后只更新应用并启动配方：

```powershell
docker compose -p sub2apicust-local --env-file .env -f docker-compose.override.yml -f docker-compose.family-local.yml.example up -d --no-deps --no-build --wait --wait-timeout 120 sub2api recipes
```

- 对比运行容器与镜像标签 ID，确认不是旧镜像；原数据库／Redis ID及app_data卷保持，业务计数和设置摘要与备份比较。
- 主站 `/health`、`/family`、原介绍 `/api` 及配方 GET／HEAD正常，非法静态路径404，非GET／HEAD方法405。配方没有额外 API 或后台。
- 4178来源模型端点的 OPTIONS应204且返回精确来源／Authorization／Content-Type，未知来源应403；此项不调用生成模型，也不替代真实 Key 与权限验证。
- 首页保留原品牌首屏，下方产品介绍。普通登录进入 API 控制台，子产品匿名入口登录后仍先进入控制台并显示继续打开按钮，顶部胶囊切换其他产品。不会自动弹窗或授权。
- 有效原站会话进入 `/connect/recipes` 不用再次输入密码，选择已有密钥与模型并授权才回传。不自动创建 Key，不共享登录令牌，不保证独立产品的通用SSO或单点退出。
- 用户自行登录、验配置导入，检查提示词后如需实际调用模型再自行运行。生产域名、真实服务权限／计费、SynaRoute账号SSO及工坊Docker部署仍不在本次部署范围。

后续本机重建继续使用上述叠加层，否则原配置单独构建会丢失本机产品地址与新增来源许可。仅首页变化时可以只 build／up `sub2api`，不重建配方、数据库、Redis。准确镜像、备份和验证结果见根 `HANDOFF.md`，初版证据在Git忽略的 `output/family-docker-20260930/`，用户纠正后的新版在 `output/family-layout-docker-20260930/`，旧备份保留。镜像回退与数据库恢复是不同操作，不能为了回退首页自动恢复数据库或回滚其他未发布业务修复。
